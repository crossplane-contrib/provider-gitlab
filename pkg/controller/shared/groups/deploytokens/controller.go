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
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiCluster "github.com/crossplane-contrib/provider-gitlab/apis/cluster/groups/v1alpha1"
	apiNamespaced "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	sharedGroupsV1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/shared/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups"
)

const (
	ErrNotDeployToken = "managed resource is not a Gitlab deploytoken custom resource"
	ErrGetFailed      = "cannot get Gitlab deploytoken"
	ErrCreateFailed   = "cannot create Gitlab deploytoken"
	ErrDeleteFailed   = "cannot delete Gitlab deploytoken"
	ErrIDNotInt       = "ID is not integer value"
	ErrGroupIDMissing = "GroupID is missing"
)

type External struct {
	Client groups.DeployTokenClient
	Kube   client.Client
}

type options struct {
	externalName  string
	parameters    *sharedGroupsV1alpha1.DeployTokenParameters
	setConditions func(c ...common.Condition)
	mg            resource.Managed
}

func (e *External) extractOptions(mg resource.Managed) (*options, error) {
	switch cr := mg.(type) {
	case *apiCluster.DeployToken:
		return &options{
			externalName:  meta.GetExternalName(cr),
			parameters:    &cr.Spec.ForProvider.DeployTokenParameters,
			setConditions: cr.Status.SetConditions,
			mg:            mg,
		}, nil
	case *apiNamespaced.DeployToken:
		return &options{
			externalName:  meta.GetExternalName(cr),
			parameters:    &cr.Spec.ForProvider.DeployTokenParameters,
			setConditions: cr.Status.SetConditions,
			mg:            mg,
		}, nil
	default:
		return nil, errors.New(ErrNotDeployToken)
	}
}

func (e *External) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	o, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if o.externalName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	id, err := strconv.Atoi(o.externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.New(ErrIDNotInt)
	}

	if o.parameters.GroupID == nil {
		return managed.ExternalObservation{}, errors.New(ErrGroupIDMissing)
	}

	dt, res, err := e.Client.GetGroupDeployToken(*o.parameters.GroupID, id)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, ErrGetFailed)
	}

	current := o.parameters.DeepCopy()
	lateInitializeGroupDeployToken(o.parameters, dt)

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
		return managed.ExternalCreation{}, errors.Wrap(errors.New("GroupId must be set directly or via reference"), ErrCreateFailed)
	}

	dt, _, err := e.Client.CreateGroupDeployToken(
		*o.parameters.GroupID,
		groups.GenerateCreateGroupDeployTokenOptions(o.mg.GetName(), o.parameters),
		gitlab.WithContext(ctx),
	)

	connectionDetails := managed.ConnectionDetails{}
	connectionDetails["token"] = []byte(dt.Token)

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, ErrCreateFailed)
	}

	meta.SetExternalName(o.mg, strconv.Itoa(dt.ID))
	return managed.ExternalCreation{ConnectionDetails: connectionDetails}, nil
}

func (e *External) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// it's not possible to update a GroupDeployToken
	return managed.ExternalUpdate{}, nil
}

func (e *External) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	o, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalDelete{}, err
	}

	deployTokenID, err := strconv.Atoi(meta.GetExternalName(o.mg))
	if err != nil {
		return managed.ExternalDelete{}, errors.New(ErrNotDeployToken)
	}

	if o.parameters.GroupID == nil {
		return managed.ExternalDelete{}, errors.New(ErrGroupIDMissing)
	}

	_, deleteError := e.Client.DeleteGroupDeployToken(
		*o.parameters.GroupID,
		deployTokenID,
		gitlab.WithContext(ctx),
	)

	return managed.ExternalDelete{}, errors.Wrap(deleteError, ErrDeleteFailed)
}

// lateInitializeGroupDeployToken fills the empty fields in the deploy token spec with the
// values seen in gitlab deploy token.
func lateInitializeGroupDeployToken(in *sharedGroupsV1alpha1.DeployTokenParameters, deployToken *gitlab.DeployToken) {
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
