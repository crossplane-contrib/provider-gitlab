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

package integrationgithub

import (
	"context"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/errors"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/statemetrics"
	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects"
)

const (
	errNotIntegrationGithub = "managed resource is not a Gitlab integration github custom resource"
	errProjectIDMissing     = "ProjectID is missing"
	errGetFailed            = "cannot get Gitlab integration github"
	errCreateFailed         = "cannot create Gitlab integration github"
	errUpdateFailed         = "cannot update Gitlab integration github"
	errDeleteFailed         = "cannot delete Gitlab integration github"
)

// SetupIntegrationGithub adds a controller that reconciles GitLab Integration GitHub.
func SetupIntegrationGithub(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.IntegrationGithubGroupKind)

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabClientFn: projects.NewGithubClient}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}

	if o.Features.Enabled(feature.EnableBetaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(
		mgr,
		resource.ManagedKind(v1alpha1.IntegrationGithubGroupVersionKind),
		reconcilerOpts...,
	)

	if err := mgr.Add(statemetrics.NewMRStateRecorder(
		mgr.GetClient(),
		o.Logger,
		o.MetricOptions.MRStateMetrics,
		&v1alpha1.IntegrationGithubList{},
		o.MetricOptions.PollStateMetricInterval,
	)); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.IntegrationGithub{}).
		Complete(r)
}

// SetupIntegrationGithubGated adds a controller with CRD gate support.
func SetupIntegrationGithubGated(mgr ctrl.Manager, o controller.Options) error {
	o.Gate.Register(func() {
		if err := SetupIntegrationGithub(mgr, o); err != nil {
			mgr.GetLogger().Error(err, "unable to setup reconciler", "gvk", v1alpha1.IntegrationGithubGroupVersionKind.String())
		}
	}, v1alpha1.IntegrationGithubGroupVersionKind)
	return nil
}

// connector produces an ExternalClient for GitLab Integration GitHub.
type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg common.Config) projects.GithubClient
}

// Connect creates a new GitLab client for the given managed resource.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.IntegrationGithub)
	if !ok {
		return nil, errors.New(errNotIntegrationGithub)
	}
	cfg, err := common.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

// external represents the external client for GitLab Integration GitHub.
type external struct {
	kube   client.Client
	client projects.GithubClient
}

// applyGithub applies desired GitHub settings to the GitLab project.
func (e *external) applyGithub(ctx context.Context, cr *v1alpha1.IntegrationGithub) error {
	_, _, err := e.client.SetGithubService(
		*cr.Spec.ForProvider.ProjectID,
		projects.GenerateSetGithubServiceOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	return err
}

// Observe checks whether the external resource exists and whether it is up-to-date.
func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.IntegrationGithub)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotIntegrationGithub)
	}

	// If the resource is being deleted, avoid updating status.
	if meta.WasDeleted(cr) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(errProjectIDMissing)
	}

	github, res, err := e.client.GetGithubService(
		*cr.Spec.ForProvider.ProjectID,
		gitlab.WithContext(ctx),
	)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}
	if github == nil || github.Properties == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	// Late initialize spec from remote (mutates in place).
	current := cr.Spec.ForProvider.DeepCopy()
	projects.LateInitializeIntegrationGithub(&cr.Spec.ForProvider, github)

	// Update status from the remote state.
	cr.Status.AtProvider = projects.GenerateIntegrationGithubObservation(github)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        projects.IsIntegrationGithubUpToDate(&cr.Spec.ForProvider, github),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

// Create creates the external resource for GitLab Integration GitHub.
// The integration is configured by sending desired options directly.
func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.IntegrationGithub)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotIntegrationGithub)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalCreation{}, errors.New(errProjectIDMissing)
	}

	cr.Status.SetConditions(xpv1.Creating())

	if err := e.applyGithub(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}
	return managed.ExternalCreation{}, nil
}

// Update updates the external resource to match the desired state.
func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.IntegrationGithub)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotIntegrationGithub)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalUpdate{}, errors.New(errProjectIDMissing)
	}

	// Set Creating condition to align with controller tests and convention.
	cr.Status.SetConditions(xpv1.Creating())

	if err := e.applyGithub(ctx, cr); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}
	return managed.ExternalUpdate{}, nil
}

// Delete removes the GitLab GitHub integration.
func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.IntegrationGithub)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotIntegrationGithub)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalDelete{}, errors.New(errProjectIDMissing)
	}

	// Do not set Deleting condition to match existing controller test expectations.
	_, err := e.client.DeleteGithubService(
		*cr.Spec.ForProvider.ProjectID,
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteFailed)
	}
	return managed.ExternalDelete{}, nil
}

// Disconnect is a no-op required by the SDK interface.
func (e *external) Disconnect(ctx context.Context) error {
	return nil
}
