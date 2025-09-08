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

package hooks

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
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiCluster "github.com/crossplane-contrib/provider-gitlab/apis/cluster/projects/v1alpha1"
	apiNamespaced "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	sharedProjectsV1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/shared/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
)

const (
	ErrNotHook          = "managed resource is not a Gitlab project hook custom resource"
	ErrProjectIDMissing = "ProjectID is missing"
	ErrGetFailed        = "cannot get Gitlab project hook"
	ErrKubeUpdateFailed = "cannot update Gitlab project hook custom resource"
	ErrCreateFailed     = "cannot create Gitlab project hook"
	ErrUpdateFailed     = "cannot update Gitlab project hook"
	ErrDeleteFailed     = "cannot delete Gitlab project hook"
	ErrSecretRefInvalid = "invalid token reference"
)

type External struct {
	Client projects.HookClient
	Kube   client.Client
}

type options struct {
	externalName   string
	parameters     *sharedProjectsV1alpha1.HookParameters
	atProvider     *sharedProjectsV1alpha1.HookObservation
	setConditions  func(c ...common.Condition)
	mg             resource.Managed
	tokenSecretRef *xpv1.SecretKeySelector
}

func (e *External) extractOptions(mg resource.Managed) (*options, error) {
	switch cr := mg.(type) {
	case *apiCluster.Hook:
		var tokenSecretRef *xpv1.SecretKeySelector
		if cr.Spec.ForProvider.Token != nil {
			tokenSecretRef = cr.Spec.ForProvider.Token.SecretRef
		}
		return &options{
			externalName:   meta.GetExternalName(cr),
			parameters:     &cr.Spec.ForProvider.HookParameters,
			atProvider:     &cr.Status.AtProvider,
			setConditions:  cr.SetConditions,
			mg:             mg,
			tokenSecretRef: tokenSecretRef,
		}, nil
	case *apiNamespaced.Hook:
		var tokenSecretRef *xpv1.SecretKeySelector
		if cr.Spec.ForProvider.Token != nil {
			tokenSecretRef = cr.Spec.ForProvider.Token.SecretRef.ToSecretKeySelector(cr.GetNamespace())
		}
		return &options{
			externalName:   meta.GetExternalName(cr),
			parameters:     &cr.Spec.ForProvider.HookParameters,
			atProvider:     &cr.Status.AtProvider,
			setConditions:  cr.SetConditions,
			mg:             mg,
			tokenSecretRef: tokenSecretRef,
		}, nil
	default:
		return nil, errors.New(ErrNotHook)
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

	hookid, err := strconv.Atoi(opts.externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.New(ErrNotHook)
	}

	if opts.parameters.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(ErrProjectIDMissing)
	}

	projecthook, res, err := e.Client.GetProjectHook(*opts.parameters.ProjectID, hookid)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(projects.IsErrorHookNotFound, err), ErrGetFailed)
	}

	current := opts.parameters.DeepCopy()
	projects.LateInitializeHook(opts.parameters, projecthook)

	*opts.atProvider = projects.GenerateHookObservation(projecthook)
	opts.setConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        projects.IsHookUpToDate(opts.parameters, projecthook),
		ResourceLateInitialized: !cmp.Equal(current, opts.parameters),
	}, nil
}

func (e *External) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	opts.setConditions(xpv1.Creating())
	hookOptions, err := projects.GenerateCreateHookOptions(opts.parameters, opts.tokenSecretRef, e.Kube, ctx)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, ErrSecretRefInvalid)
	}

	hook, _, err := e.Client.AddProjectHook(*opts.parameters.ProjectID, hookOptions, gitlab.WithContext(ctx))
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, ErrCreateFailed)
	}

	err = e.updateExternalName(ctx, opts.mg, hook)
	return managed.ExternalCreation{}, errors.Wrap(err, ErrKubeUpdateFailed)
}

func (e *External) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	hookid, err := strconv.Atoi(opts.externalName)
	if err != nil {
		return managed.ExternalUpdate{}, errors.New(ErrNotHook)
	}

	if opts.parameters.ProjectID == nil {
		return managed.ExternalUpdate{}, errors.New(ErrProjectIDMissing)
	}

	editHookOptions, err := projects.GenerateEditHookOptions(opts.parameters, opts.tokenSecretRef, e.Kube, ctx)
	if err != nil {
		return managed.ExternalUpdate{}, errors.New(ErrSecretRefInvalid)
	}

	_, _, err = e.Client.EditProjectHook(*opts.parameters.ProjectID, hookid, editHookOptions, gitlab.WithContext(ctx))
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, ErrUpdateFailed)
	}

	return managed.ExternalUpdate{}, nil
}

func (e *External) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalDelete{}, err
	}

	opts.setConditions(xpv1.Deleting())

	if opts.parameters.ProjectID == nil {
		return managed.ExternalDelete{}, errors.New(ErrProjectIDMissing)
	}

	_, err = e.Client.DeleteProjectHook(*opts.parameters.ProjectID, opts.atProvider.ID, gitlab.WithContext(ctx))
	return managed.ExternalDelete{}, errors.Wrap(err, ErrDeleteFailed)
}

func (e *External) updateExternalName(ctx context.Context, mg resource.Managed, projecthook *gitlab.ProjectHook) error {
	meta.SetExternalName(mg, strconv.Itoa(projecthook.ID))
	return e.Kube.Update(ctx, mg)
}

func (e *External) Disconnect(ctx context.Context) error {
	return nil
}
