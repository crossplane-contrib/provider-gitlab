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

package integrationyoutrack

import (
	"context"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/errors"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/statemetrics"
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects"
)

const (
	errNotIntegrationYouTrack = "managed resource is not a Gitlab integration youtrack custom resource"
	errProjectIDMissing       = "ProjectID is missing"
	errGetFailed              = "cannot get Gitlab integration youtrack"
	errCreateFailed           = "cannot create Gitlab integration youtrack"
	errUpdateFailed           = "cannot update Gitlab integration youtrack"
	errDeleteFailed           = "cannot delete Gitlab integration youtrack"
)

// SetupIntegrationYouTrack adds a controller that reconciles GitLab Project YouTrack integrations.
func SetupIntegrationYouTrack(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.IntegrationYouTrackGroupKind)

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnector(&connector{kube: mgr.GetClient(), newGitlabClientFn: projects.NewYouTrackClient}),
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
		resource.ManagedKind(v1alpha1.IntegrationYouTrackGroupVersionKind),
		reconcilerOpts...,
	)

	if err := mgr.Add(statemetrics.NewMRStateRecorder(
		mgr.GetClient(),
		o.Logger,
		o.MetricOptions.MRStateMetrics,
		&v1alpha1.IntegrationYouTrackList{},
		o.MetricOptions.PollStateMetricInterval,
	)); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.IntegrationYouTrack{}).
		Complete(r)
}

// SetupIntegrationYouTrackGated adds a controller with CRD gate support for SafeStart.
func SetupIntegrationYouTrackGated(mgr ctrl.Manager, o controller.Options) error {
	o.Gate.Register(func() {
		if err := SetupIntegrationYouTrack(mgr, o); err != nil {
			mgr.GetLogger().Error(err, "unable to setup reconciler", "gvk", v1alpha1.IntegrationYouTrackGroupVersionKind.String())
		}
	}, v1alpha1.IntegrationYouTrackGroupVersionKind)
	return nil
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg common.Config) projects.YouTrackClient
}

// Connect creates a new GitLab client for the given managed resource.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.IntegrationYouTrack)
	if !ok {
		return nil, errors.New(errNotIntegrationYouTrack)
	}
	cfg, err := common.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newGitlabClientFn(*cfg)}, nil
}

type external struct {
	client projects.YouTrackClient
}

func (e *external) apply(ctx context.Context, cr *v1alpha1.IntegrationYouTrack) error {
	_, _, err := e.client.SetYouTrackService(
		*cr.Spec.ForProvider.ProjectID,
		projects.GenerateSetYouTrackServiceOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	return err
}

// Observe checks whether the external resource exists and whether it is up-to-date.
func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.IntegrationYouTrack)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotIntegrationYouTrack)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(errProjectIDMissing)
	}

	tracker, res, err := e.client.GetYouTrackService(
		*cr.Spec.ForProvider.ProjectID,
		gitlab.WithContext(ctx),
	)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}
	if tracker == nil || !tracker.Active {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	cr.Status.AtProvider = projects.GenerateIntegrationYouTrackObservation(tracker)
	cr.Status.SetConditions(v2.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: projects.IsIntegrationYouTrackUpToDate(&cr.Spec.ForProvider, tracker),
	}, nil
}

// Create enables the YouTrack integration on the GitLab project.
func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.IntegrationYouTrack)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotIntegrationYouTrack)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalCreation{}, errors.New(errProjectIDMissing)
	}

	cr.Status.SetConditions(v2.Creating())

	if err := e.apply(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}
	return managed.ExternalCreation{}, nil
}

// Update updates the YouTrack integration to match the desired state.
func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.IntegrationYouTrack)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotIntegrationYouTrack)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalUpdate{}, errors.New(errProjectIDMissing)
	}

	if err := e.apply(ctx, cr); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}
	return managed.ExternalUpdate{}, nil
}

// Delete disables the YouTrack integration on the GitLab project.
func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.IntegrationYouTrack)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotIntegrationYouTrack)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalDelete{}, errors.New(errProjectIDMissing)
	}

	_, err := e.client.DeleteYouTrackService(
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
