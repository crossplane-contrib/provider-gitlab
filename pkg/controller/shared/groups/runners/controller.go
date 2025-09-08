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

package runners

import (
	"context"
	"slices"
	"strconv"

	"github.com/crossplane/crossplane-runtime/v2/apis/common"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiCluster "github.com/crossplane-contrib/provider-gitlab/apis/cluster/groups/v1alpha1"
	apiNamespaced "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	sharedGroupsV1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/shared/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	runners "github.com/crossplane-contrib/provider-gitlab/pkg/clients/runners"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/users"
)

const (
	ErrNotRunner               = "managed resource is not a Runner custom resource"
	ErrIDNotInt                = "specified ID is not an integer"
	ErrGetFailed               = "cannot get Gitlab Runner"
	ErrCreateFailed            = "cannot create Gitlab Runner"
	ErrUpdateFailed            = "cannot update Gitlab Runner"
	ErrDeleteFailed            = "cannot delete Gitlab Runner"
	ErrRunnertNotFound         = "cannot find Gitlab Runner"
	ErrMissingGroupID          = "missing Spec.ForProvider.GroupID"
	ErrMissingExternalName     = "external name annotation not found"
	ErrMissingConnectionSecret = "writeConnectionSecretToRef must be specified to receive the runner token"
)

type External struct {
	Client           runners.RunnerClient
	UserRunnerClient users.RunnerClient
	Kube             client.Client
}

type options struct {
	externalName        string
	parameters          sharedGroupsV1alpha1.RunnerParameters
	setConditions       func(c ...common.Condition)
	atProvider          *sharedGroupsV1alpha1.RunnerObservation
	hasConnectionSecret bool
	mg                  resource.Managed
}

func (e *External) extractOptions(mg resource.Managed) (*options, error) {
	switch cr := mg.(type) {
	case *apiCluster.Runner:
		return &options{
			externalName:        meta.GetExternalName(cr),
			parameters:          cr.Spec.ForProvider.RunnerParameters,
			setConditions:       cr.Status.SetConditions,
			atProvider:          &cr.Status.AtProvider,
			hasConnectionSecret: cr.Spec.WriteConnectionSecretToReference != nil,
			mg:                  mg,
		}, nil
	case *apiNamespaced.Runner:
		return &options{
			externalName:        meta.GetExternalName(cr),
			parameters:          cr.Spec.ForProvider.RunnerParameters,
			setConditions:       cr.Status.SetConditions,
			atProvider:          &cr.Status.AtProvider,
			hasConnectionSecret: cr.Spec.WriteConnectionSecretToReference != nil,
			mg:                  mg,
		}, nil
	default:
		return nil, errors.New(ErrNotRunner)
	}
}

func (e *External) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	o, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if o.parameters.GroupID == nil {
		return managed.ExternalObservation{}, errors.New(ErrMissingGroupID)
	}

	if o.externalName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	runnerID, err := strconv.Atoi(o.externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.New(ErrIDNotInt)
	}

	runner, res, err := e.Client.GetRunnerDetails(runnerID, nil)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, ErrGetFailed)
	}

	// we need to make sure the token expiration time is preserved as it is not returned by the API
	tokenExpiresAt := o.atProvider.TokenExpiresAt
	*o.atProvider = runners.GenerateGroupRunnerObservation(runner)
	o.atProvider.TokenExpiresAt = tokenExpiresAt

	o.setConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isRunnerUpToDate(&o.parameters, runner),
		ResourceLateInitialized: false,
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

	// Validate that connection details will be published
	if !o.hasConnectionSecret {
		return managed.ExternalCreation{}, errors.New(ErrMissingConnectionSecret)
	}

	runner, _, err := e.UserRunnerClient.CreateUserRunner(
		users.GenerateGroupRunnerOptions(&o.parameters),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, ErrCreateFailed)
	}

	meta.SetExternalName(o.mg, strconv.Itoa(runner.ID))

	if runner.TokenExpiresAt != nil {
		t := metav1.NewTime(*runner.TokenExpiresAt)
		o.atProvider.TokenExpiresAt = &t
	} else {
		o.atProvider.TokenExpiresAt = nil
	}

	o.setConditions(xpv1.Creating())

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{
			"token": []byte(runner.Token),
		},
	}, nil
}

func (e *External) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	o, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	if o.parameters.GroupID == nil {
		return managed.ExternalUpdate{}, errors.New(ErrMissingGroupID)
	}

	if o.externalName == "" {
		return managed.ExternalUpdate{}, errors.New(ErrMissingExternalName)
	}

	runnerID, err := strconv.Atoi(o.externalName)
	if err != nil {
		return managed.ExternalUpdate{}, errors.New(ErrIDNotInt)
	}

	_, _, err = e.Client.UpdateRunnerDetails(
		runnerID,
		runners.GenerateEditRunnerOptions(&o.parameters.CommonRunnerParameters),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, ErrUpdateFailed)
	}

	return managed.ExternalUpdate{}, nil
}

func (e *External) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	o, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalDelete{}, err
	}

	groupID := o.parameters.GroupID

	if groupID == nil {
		return managed.ExternalDelete{}, errors.New(ErrMissingGroupID)
	}

	if o.externalName == "" {
		return managed.ExternalDelete{}, errors.New(ErrMissingExternalName)
	}

	runnerID, err := strconv.Atoi(o.externalName)
	if err != nil {
		return managed.ExternalDelete{}, errors.New(ErrIDNotInt)
	}

	_, err = e.Client.DeleteRegisteredRunnerByID(
		runnerID,
		gitlab.WithContext(ctx),
	)

	return managed.ExternalDelete{}, errors.Wrap(err, ErrDeleteFailed)
}

func isRunnerUpToDate(p *sharedGroupsV1alpha1.RunnerParameters, r *gitlab.RunnerDetails) bool { //nolint:gocyclo
	if p.Description != nil && *p.Description != r.Description {
		return false
	}
	if p.Paused != nil && *p.Paused != r.Paused {
		return false
	}
	if p.Locked != nil && *p.Locked != r.Locked {
		return false
	}
	if p.RunUntagged != nil && *p.RunUntagged != r.RunUntagged {
		return false
	}
	if p.TagList != nil && !slices.Equal(*p.TagList, r.TagList) {
		return false
	}
	if p.AccessLevel != nil && *p.AccessLevel != r.AccessLevel {
		return false
	}
	if p.MaximumTimeout != nil && *p.MaximumTimeout != r.MaximumTimeout {
		return false
	}
	if p.MaintenanceNote != nil && *p.MaintenanceNote != r.MaintenanceNote {
		return false
	}

	return true
}

func (e *External) Disconnect(ctx context.Context) error {
	return nil
}
