/*
Copyright 2021 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package accesstokens

import (
	"context"
	"strconv"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/apis/common"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiCluster "github.com/crossplane-contrib/provider-gitlab/apis/cluster/projects/v1alpha1"
	apiNamespaced "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	sharedProjectsV1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/shared/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
)

const (
	ErrNotAccessToken      = "managed resource is not a Gitlab accesstoken custom resource"
	ErrExternalNameNotInt  = "custom resource external name is not an integer"
	ErrFailedParseID       = "cannot parse Access Token ID to int"
	ErrGetFailed           = "cannot get Gitlab accesstoken"
	ErrCreateFailed        = "cannot create Gitlab accesstoken"
	ErrDeleteFailed        = "cannot delete Gitlab accesstoken"
	ErrAccessTokenNotFound = "cannot find Gitlab accesstoken"
	ErrMissingProjectID    = "missing Spec.ForProvider.ProjectID"
)

type External struct {
	Client projects.AccessTokenClient
}

type options struct {
	externalName  string
	parameters    *sharedProjectsV1alpha1.AccessTokenParameters
	atProvider    *sharedProjectsV1alpha1.AccessTokenObservation
	setConditions func(c ...common.Condition)
	mg            resource.Managed
}

func (e *External) extractOptions(mg resource.Managed) (*options, error) {
	switch cr := mg.(type) {
	case *apiCluster.AccessToken:
		return &options{
			externalName:  meta.GetExternalName(cr),
			parameters:    &cr.Spec.ForProvider.AccessTokenParameters,
			atProvider:    &cr.Status.AtProvider,
			setConditions: cr.SetConditions,
			mg:            mg,
		}, nil
	case *apiNamespaced.AccessToken:
		return &options{
			externalName:  meta.GetExternalName(cr),
			parameters:    &cr.Spec.ForProvider.AccessTokenParameters,
			atProvider:    &cr.Status.AtProvider,
			setConditions: cr.SetConditions,
			mg:            mg,
		}, nil
	default:
		return nil, errors.New(ErrNotAccessToken)
	}
}

func (e *External) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if opts.externalName == "" {
		return managed.ExternalObservation{}, nil
	}

	accessTokenID, err := strconv.Atoi(opts.externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, ErrFailedParseID)
	}

	if opts.parameters.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(ErrMissingProjectID)
	}

	at, res, err := e.Client.GetProjectAccessToken(*opts.parameters.ProjectID, accessTokenID)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, ErrAccessTokenNotFound)
	}

	if at.Revoked {
		return managed.ExternalObservation{}, nil
	}

	current := opts.parameters.DeepCopy()
	lateInitializeProjectAccessToken(opts.parameters, at)

	opts.setConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        true,
		ResourceLateInitialized: !cmp.Equal(current, opts.parameters),
	}, nil
}

func (e *External) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	if opts.parameters.ProjectID == nil {
		return managed.ExternalCreation{}, errors.New(ErrMissingProjectID)
	}

	at, _, err := e.Client.CreateProjectAccessToken(
		*opts.parameters.ProjectID,
		projects.GenerateCreateProjectAccessTokenOptions(opts.mg.GetName(), opts.parameters),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, ErrCreateFailed)
	}

	meta.SetExternalName(opts.mg, strconv.Itoa(at.ID))
	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{
			"token": []byte(at.Token),
		},
	}, nil
}

func (e *External) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// it's not possible to update a ProjectAccessToken
	return managed.ExternalUpdate{}, nil
}

func (e *External) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalDelete{}, err
	}

	accessTokenID, err := strconv.Atoi(opts.externalName)
	if err != nil {
		return managed.ExternalDelete{}, errors.New(ErrExternalNameNotInt)
	}

	if opts.parameters.ProjectID == nil {
		return managed.ExternalDelete{}, errors.New(ErrMissingProjectID)
	}

	_, err = e.Client.RevokeProjectAccessToken(
		*opts.parameters.ProjectID,
		accessTokenID,
		gitlab.WithContext(ctx),
	)

	return managed.ExternalDelete{}, errors.Wrap(err, ErrDeleteFailed)
}

// lateInitializeProjectAccessToken fills the empty fields in the access token spec with the
// values seen in gitlab access token.
func lateInitializeProjectAccessToken(in *sharedProjectsV1alpha1.AccessTokenParameters, accessToken *gitlab.ProjectAccessToken) {
	if accessToken == nil {
		return
	}

	if in.AccessLevel == nil {
		in.AccessLevel = (*sharedProjectsV1alpha1.AccessLevelValue)(&accessToken.AccessLevel)
	}

	if in.ExpiresAt == nil && accessToken.ExpiresAt != nil {
		in.ExpiresAt = &metav1.Time{Time: time.Time(*accessToken.ExpiresAt)}
	}
}

func (e *External) Disconnect(ctx context.Context) error {
	return nil
}
