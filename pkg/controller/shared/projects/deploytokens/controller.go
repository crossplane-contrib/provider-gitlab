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

package deploytokens

import (
	"context"
	"strconv"

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
	ErrNotDeployToken   = "managed resource is not a Gitlab deploytoken custom resource"
	ErrIDNotInt         = "ID is not an integer"
	ErrGetFailed        = "cannot get Gitlab deploytoken"
	ErrCreateFailed     = "cannot create Gitlab deploytoken"
	ErrDeleteFailed     = "cannot delete Gitlab deploytoken"
	ErrProjectIDMissing = "projectID missing"
)

type External struct {
	Client projects.DeployTokenClient
}

type options struct {
	externalName  string
	parameters    *sharedProjectsV1alpha1.DeployTokenParameters
	atProvider    *sharedProjectsV1alpha1.DeployTokenObservation
	setConditions func(c ...common.Condition)
	mg            resource.Managed
}

func (e *External) extractOptions(mg resource.Managed) (*options, error) {
	switch cr := mg.(type) {
	case *apiCluster.DeployToken:
		return &options{
			externalName:  meta.GetExternalName(cr),
			parameters:    &cr.Spec.ForProvider.DeployTokenParameters,
			atProvider:    &cr.Status.AtProvider,
			setConditions: cr.SetConditions,
			mg:            mg,
		}, nil
	case *apiNamespaced.DeployToken:
		return &options{
			externalName:  meta.GetExternalName(cr),
			parameters:    &cr.Spec.ForProvider.DeployTokenParameters,
			atProvider:    &cr.Status.AtProvider,
			setConditions: cr.SetConditions,
			mg:            mg,
		}, nil
	default:
		return nil, errors.New(ErrNotDeployToken)
	}
}

func (e *External) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if opts.externalName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	id, err := strconv.Atoi(opts.externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.New(ErrIDNotInt)
	}

	if opts.parameters.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(ErrProjectIDMissing)
	}

	dt, res, err := e.Client.GetProjectDeployToken(*opts.parameters.ProjectID, id)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, ErrGetFailed)
	}

	current := opts.parameters.DeepCopy()
	lateInitializeProjectDeployToken(opts.parameters, dt)

	*opts.atProvider = sharedProjectsV1alpha1.DeployTokenObservation{}
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
		return managed.ExternalCreation{}, errors.New(ErrProjectIDMissing)
	}

	dt, _, err := e.Client.CreateProjectDeployToken(
		*opts.parameters.ProjectID,
		projects.GenerateCreateProjectDeployTokenOptions(opts.mg.GetName(), opts.parameters),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, ErrCreateFailed)
	}

	connectionDetails := managed.ConnectionDetails{}
	connectionDetails["token"] = []byte(dt.Token)

	meta.SetExternalName(opts.mg, strconv.Itoa(dt.ID))
	return managed.ExternalCreation{ConnectionDetails: connectionDetails}, nil
}

func (e *External) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// it's not possible to update a ProjectDeployToken
	return managed.ExternalUpdate{}, nil
}

func (e *External) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalDelete{}, err
	}

	deployTokenID, err := strconv.Atoi(opts.externalName)
	if err != nil {
		return managed.ExternalDelete{}, errors.New(ErrNotDeployToken)
	}

	if opts.parameters.ProjectID == nil {
		return managed.ExternalDelete{}, errors.New(ErrProjectIDMissing)
	}

	_, deleteError := e.Client.DeleteProjectDeployToken(
		*opts.parameters.ProjectID,
		deployTokenID,
		gitlab.WithContext(ctx),
	)

	return managed.ExternalDelete{}, errors.Wrap(deleteError, ErrDeleteFailed)
}

// lateInitializeProjectDeployToken fills the empty fields in the deploy token spec with the
// values seen in gitlab deploy token.
func lateInitializeProjectDeployToken(in *sharedProjectsV1alpha1.DeployTokenParameters, deployToken *gitlab.DeployToken) {
	if deployToken == nil {
		return
	}

	if in.Username == nil {
		in.Username = &deployToken.Username
	}

	if in.ExpiresAt == nil && deployToken.ExpiresAt != nil {
		in.ExpiresAt = &metav1.Time{Time: *deployToken.ExpiresAt}
	}
}

func (e *External) Disconnect(ctx context.Context) error {
	return nil
}
