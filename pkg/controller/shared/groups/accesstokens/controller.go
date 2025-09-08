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
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiCluster "github.com/crossplane-contrib/provider-gitlab/apis/cluster/groups/v1alpha1"
	apiNamespaced "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	sharedGroupsV1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/shared/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups"
)

const (
	ErrNotAccessToken       = "managed resource is not a Gitlab accesstoken custom resource"
	ErrExternalNameNotInt   = "custom resource external name is not an integer"
	ErrFailedParseID        = "cannot parse Access Token ID to int"
	ErrGetFailed            = "cannot get Gitlab accesstoken"
	ErrCreateFailed         = "cannot create Gitlab accesstoken"
	ErrDeleteFailed         = "cannot delete Gitlab accesstoken"
	ErrAccessTokentNotFound = "cannot find Gitlab accesstoken"
	ErrMissingGroupID       = "missing Spec.ForProvider.GroupID"
)

type External struct {
	Client groups.AccessTokenClient
	Kube   client.Client
}

type options struct {
	externalName  string
	parameters    *sharedGroupsV1alpha1.AccessTokenParameters
	setConditions func(c ...common.Condition)
	mg            resource.Managed
}

func (e *External) extractOptions(mg resource.Managed) (*options, error) {
	switch cr := mg.(type) {
	case *apiCluster.AccessToken:
		return &options{
			externalName:  meta.GetExternalName(cr),
			parameters:    &cr.Spec.ForProvider.AccessTokenParameters,
			setConditions: cr.Status.SetConditions,
			mg:            mg,
		}, nil
	case *apiNamespaced.AccessToken:
		return &options{
			externalName:  meta.GetExternalName(cr),
			parameters:    &cr.Spec.ForProvider.AccessTokenParameters,
			setConditions: cr.Status.SetConditions,
			mg:            mg,
		}, nil
	default:
		return nil, errors.New(ErrNotAccessToken)
	}
}

func (e *External) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	o, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if o.externalName == "" {
		return managed.ExternalObservation{}, nil
	}

	accessTokenID, err := strconv.Atoi(o.externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, ErrFailedParseID)
	}

	if o.parameters.GroupID == nil {
		return managed.ExternalObservation{}, errors.New(ErrMissingGroupID)
	}

	at, res, err := e.Client.GetGroupAccessToken(*o.parameters.GroupID, accessTokenID)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, ErrAccessTokentNotFound)
	}

	if at.Revoked {
		return managed.ExternalObservation{}, nil
	}

	current := o.parameters.DeepCopy()
	lateInitializeGroupAccessToken(o.parameters, at)

	o.setConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        true,
		ResourceLateInitialized: !cmp.Equal(current, o.parameters),
	}, nil
}

func (e *External) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	o, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	if o.parameters.GroupID == nil {
		return managed.ExternalCreation{}, errors.New(ErrMissingGroupID)
	}

	at, _, err := e.Client.CreateGroupAccessToken(
		*o.parameters.GroupID,
		groups.GenerateCreateGroupAccessTokenOptions(o.mg.GetName(), o.parameters),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, ErrCreateFailed)
	}

	meta.SetExternalName(o.mg, strconv.Itoa(at.ID))
	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{
			"token": []byte(at.Token),
		},
	}, nil
}

func (e *External) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (e *External) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	o, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalDelete{}, err
	}

	accessTokenID, err := strconv.Atoi(meta.GetExternalName(o.mg))
	if err != nil {
		return managed.ExternalDelete{}, errors.New(ErrExternalNameNotInt)
	}

	if o.parameters.GroupID == nil {
		return managed.ExternalDelete{}, errors.New(ErrMissingGroupID)
	}
	_, err = e.Client.RevokeGroupAccessToken(
		*o.parameters.GroupID,
		accessTokenID,
		gitlab.WithContext(ctx),
	)

	if err != nil && !groups.IsErrorGroupAccessTokenNotFound(err) {
		return managed.ExternalDelete{}, errors.Wrap(err, ErrDeleteFailed)
	}

	return managed.ExternalDelete{}, nil
}

// lateInitializeGroupAccessToken fills the empty fields in the access token spec with the
// values seen in gitlab access token.
func lateInitializeGroupAccessToken(in *sharedGroupsV1alpha1.AccessTokenParameters, accessToken *gitlab.GroupAccessToken) {
	if accessToken == nil {
		return
	}

	if in.AccessLevel == nil {
		in.AccessLevel = (*sharedGroupsV1alpha1.AccessLevelValue)(&accessToken.AccessLevel)
	}

	if in.ExpiresAt == nil && accessToken.ExpiresAt != nil {
		in.ExpiresAt = &metav1.Time{Time: time.Time(*accessToken.ExpiresAt)}
	}
}

func (e *External) Disconnect(ctx context.Context) error {
	return nil
}
