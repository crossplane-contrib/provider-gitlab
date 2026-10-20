/*
Copyright 2026 The Crossplane Authors.

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

package protectedtags

import (
	"context"

	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects"
)

const (
	errNotProtectedTag  = "managed resource is not a GitLab protected tag custom resource"
	errProjectIDMissing = "ProjectID is missing"
	errGetFailed        = "cannot get GitLab protected tag"
	errCreateFailed     = "cannot create GitLab protected tag"
	errUnprotectFailed  = "cannot unprotect GitLab protected tag for update"
	errReprotectFailed  = "cannot protect GitLab protected tag again after update"
	errDeleteFailed     = "cannot delete GitLab protected tag"
)

// SetupProtectedTag adds a controller that reconciles ProtectedTags.
func SetupProtectedTag(mgr ctrl.Manager, o common.ControllerOptions) error {
	name := managed.ControllerName(v1alpha1.ProtectedTagGroupKind)

	r := common.NewReconciler(mgr, o, name,
		resource.ManagedKind(v1alpha1.ProtectedTagGroupVersionKind),
		managed.WithExternalConnector(&connector{kube: mgr.GetClient(), newGitlabClientFn: projects.NewProtectedTagClient}),
		managed.WithInitializers(),
	)

	if err := common.AddMRStateRecorder(mgr, o, &v1alpha1.ProtectedTagList{}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.ProtectedTag{}).
		Complete(r)
}

// SetupProtectedTagGated adds a controller with CRD gate support.
func SetupProtectedTagGated(mgr ctrl.Manager, o common.ControllerOptions) error {
	o.Gate.Register(func() {
		if err := SetupProtectedTag(mgr, o); err != nil {
			mgr.GetLogger().Error(err, "unable to setup reconciler", "gvk", v1alpha1.ProtectedTagGroupVersionKind.String())
		}
	}, v1alpha1.ProtectedTagGroupVersionKind)
	return nil
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg common.Config) projects.ProtectedTagClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.ProtectedTag)
	if !ok {
		return nil, errors.New(errNotProtectedTag)
	}
	cfg, err := common.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newGitlabClientFn(*cfg)}, nil
}

type external struct {
	client projects.ProtectedTagClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.ProtectedTag)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotProtectedTag)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(errProjectIDMissing)
	}

	pt, res, err := e.client.GetProtectedTag(*cr.Spec.ForProvider.ProjectID, cr.Spec.ForProvider.TagName, gitlab.WithContext(ctx))
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}

	cr.Status.AtProvider = projects.GenerateProtectedTagObservation(pt)
	cr.Status.SetConditions(v2.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: projects.IsProtectedTagUpToDate(&cr.Spec.ForProvider, pt),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ProtectedTag)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotProtectedTag)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalCreation{}, errors.New(errProjectIDMissing)
	}

	cr.Status.SetConditions(v2.Creating())

	_, _, err := e.client.ProtectRepositoryTags(*cr.Spec.ForProvider.ProjectID, projects.GenerateProtectRepositoryTagsOptions(&cr.Spec.ForProvider), gitlab.WithContext(ctx))
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.ProtectedTag)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotProtectedTag)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalUpdate{}, errors.New(errProjectIDMissing)
	}

	// The GitLab REST API has no update endpoint for protected tags, so the
	// tag stays unprotected between these two calls.
	if _, err := e.client.UnprotectRepositoryTags(*cr.Spec.ForProvider.ProjectID, cr.Spec.ForProvider.TagName, gitlab.WithContext(ctx)); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUnprotectFailed)
	}

	_, _, err := e.client.ProtectRepositoryTags(*cr.Spec.ForProvider.ProjectID, projects.GenerateProtectRepositoryTagsOptions(&cr.Spec.ForProvider), gitlab.WithContext(ctx))
	return managed.ExternalUpdate{}, errors.Wrap(err, errReprotectFailed)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.ProtectedTag)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotProtectedTag)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalDelete{}, errors.New(errProjectIDMissing)
	}

	cr.Status.SetConditions(v2.Deleting())

	_, err := e.client.UnprotectRepositoryTags(*cr.Spec.ForProvider.ProjectID, cr.Spec.ForProvider.TagName, gitlab.WithContext(ctx))
	return managed.ExternalDelete{}, errors.Wrap(err, errDeleteFailed)
}

func (e *external) Disconnect(ctx context.Context) error {
	return nil
}
