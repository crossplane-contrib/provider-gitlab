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

	"github.com/xanzy/go-gitlab"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	secretstoreapi "github.com/crossplane-contrib/provider-gitlab/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
	"github.com/crossplane-contrib/provider-gitlab/pkg/features"
)

const (
	errNotHook          = "managed resource is not a Gitlab project hook custom resource"
	errProjectIDMissing = "ProjectID is missing"
	errGetFailed        = "cannot get Gitlab project hook"
	errKubeUpdateFailed = "cannot update Gitlab project hook custom resource"
	errCreateFailed     = "cannot create Gitlab project hook"
	errUpdateFailed     = "cannot update Gitlab project hook"
	errDeleteFailed     = "cannot delete Gitlab project hook"
)

// SetupHook adds a controller that reconciles Hooks.
func SetupHook(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.HookKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), secretstoreapi.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabClientFn: projects.NewHookClient}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.HookGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Hook{}).
		Complete(r)
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg clients.Config) projects.HookClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Hook)
	if !ok {
		return nil, errors.New(errNotHook)
	}
	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

type external struct {
	kube   client.Client
	client projects.HookClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Hook)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotHook)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	hookid, err := strconv.Atoi(meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalObservation{}, errors.New(errNotHook)
	}
	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(errProjectIDMissing)
	}

	projecthook, res, err := e.client.GetProjectHook(*cr.Spec.ForProvider.ProjectID, hookid)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(projects.IsErrorHookNotFound, err), errGetFailed)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	projects.LateInitializeHook(&cr.Spec.ForProvider, projecthook)

	cr.Status.AtProvider = projects.GenerateHookObservation(projecthook)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        projects.IsHookUpToDate(&cr.Spec.ForProvider, projecthook),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Hook)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotHook)
	}

	cr.Status.SetConditions(xpv1.Creating())
	hook, _, err := e.client.AddProjectHook(*cr.Spec.ForProvider.ProjectID, projects.GenerateCreateHookOptions(&cr.Spec.ForProvider), gitlab.WithContext(ctx))
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}
	err = e.updateExternalName(ctx, cr, hook)
	return managed.ExternalCreation{}, errors.Wrap(err, errKubeUpdateFailed)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Hook)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotHook)
	}

	hookid, err := strconv.Atoi(meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalUpdate{}, errors.New(errNotHook)
	}
	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalUpdate{}, errors.New(errProjectIDMissing)
	}

	_, _, err = e.client.EditProjectHook(*cr.Spec.ForProvider.ProjectID, hookid, projects.GenerateEditHookOptions(&cr.Spec.ForProvider), gitlab.WithContext(ctx))
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Hook)
	if !ok {
		return errors.New(errNotHook)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	if cr.Spec.ForProvider.ProjectID == nil {
		return errors.New(errProjectIDMissing)
	}
	_, err := e.client.DeleteProjectHook(*cr.Spec.ForProvider.ProjectID, cr.Status.AtProvider.ID, gitlab.WithContext(ctx))
	return errors.Wrap(err, errDeleteFailed)
}

func (e *external) updateExternalName(ctx context.Context, cr *v1alpha1.Hook, projecthook *gitlab.ProjectHook) error {
	meta.SetExternalName(cr, strconv.Itoa(projecthook.ID))
	return e.kube.Update(ctx, cr)
}
