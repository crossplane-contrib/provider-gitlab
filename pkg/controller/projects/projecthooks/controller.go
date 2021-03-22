/*
Copyright 2020 The Crossplane Authors.

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

package projecthooks

import (
	"context"
	"strconv"

	"github.com/xanzy/go-gitlab"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
)

const (
	errNotProjectHook   = "managed resource is not a Gitlab projecthook custom resource"
	errGetFailed        = "cannot get Gitlab projecthook"
	errKubeUpdateFailed = "cannot update Gitlab projecthook custom resource"
	errCreateFailed     = "cannot create Gitlab projecthook"
	errUpdateFailed     = "cannot update Gitlab projecthook"
	errDeleteFailed     = "cannot delete Gitlab projecthook"
)

// SetupProjectHook adds a controller that reconciles ProjectHooks.
func SetupProjectHook(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.ProjectHookKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.ProjectHook{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.ProjectHookGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabClientFn: projects.NewProjectHookClient}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg clients.Config) projects.ProjectHookClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.ProjectHook)
	if !ok {
		return nil, errors.New(errNotProjectHook)
	}
	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

type external struct {
	kube   client.Client
	client projects.ProjectHookClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.ProjectHook)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotProjectHook)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	hookid, err := strconv.Atoi(meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalObservation{}, errors.New(errNotProjectHook)
	}

	projecthook, _, err := e.client.GetProjectHook(*cr.Spec.ForProvider.ProjectID, hookid)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(projects.IsErrorProjectHookNotFound, err), errGetFailed)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	projects.LateInitializeProjectHook(&cr.Spec.ForProvider, projecthook)

	cr.Status.AtProvider = projects.GenerateProjectHookObservation(projecthook)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        projects.IsProjectHookUpToDate(&cr.Spec.ForProvider, projecthook),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ProjectHook)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotProjectHook)
	}

	cr.Status.SetConditions(xpv1.Creating())
	projecthook, _, err := e.client.AddProjectHook(*cr.Spec.ForProvider.ProjectID, projects.GenerateCreateProjectHookOptions(&cr.Spec.ForProvider), gitlab.WithContext(ctx))
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}
	err = e.updateExternalName(cr, projecthook)
	return managed.ExternalCreation{}, errors.Wrap(err, errKubeUpdateFailed)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.ProjectHook)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotProjectHook)
	}

	hookid, err := strconv.Atoi(meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalUpdate{}, errors.New(errNotProjectHook)
	}

	_, _, err = e.client.EditProjectHook(*cr.Spec.ForProvider.ProjectID, hookid, projects.GenerateEditProjectHookOptions(&cr.Spec.ForProvider), gitlab.WithContext(ctx))
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.ProjectHook)
	if !ok {
		return errors.New(errNotProjectHook)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.DeleteProjectHook(*cr.Spec.ForProvider.ProjectID, cr.Status.AtProvider.ID, gitlab.WithContext(ctx))
	return errors.Wrap(err, errDeleteFailed)
}

func (e *external) updateExternalName(cr *v1alpha1.ProjectHook, projecthook *gitlab.ProjectHook) error {
	meta.SetExternalName(cr, strconv.Itoa(projecthook.ID))
	return e.kube.Update(context.Background(), cr)
}
