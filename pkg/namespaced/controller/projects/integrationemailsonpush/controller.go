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

package integrationemailsonpush

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
	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects"
)

const (
	errNotIntegrationEmailsOnPush = "managed resource is not a Gitlab integration emails on push custom resource"
	errProjectIDMissing           = "ProjectID is missing"
	errGetFailed                  = "cannot get Gitlab integration emails on push"
	errCreateFailed               = "cannot create Gitlab integration emails on push"
	errUpdateFailed               = "cannot update Gitlab integration emails on push"
	errDeleteFailed               = "cannot delete Gitlab integration emails on push"
)

// SetupIntegrationEmailsOnPush adds a controller that reconciles GitLab Project Emails on Push integrations.
func SetupIntegrationEmailsOnPush(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.IntegrationEmailsOnPushGroupKind)

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnector(&connector{kube: mgr.GetClient(), newGitlabClientFn: projects.NewEmailsOnPushClient}),
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
		resource.ManagedKind(v1alpha1.IntegrationEmailsOnPushGroupVersionKind),
		reconcilerOpts...,
	)

	if err := mgr.Add(statemetrics.NewMRStateRecorder(
		mgr.GetClient(),
		o.Logger,
		o.MetricOptions.MRStateMetrics,
		&v1alpha1.IntegrationEmailsOnPushList{},
		o.MetricOptions.PollStateMetricInterval,
	)); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.IntegrationEmailsOnPush{}).
		Complete(r)
}

// SetupIntegrationEmailsOnPushGated adds a controller with CRD gate support for SafeStart.
func SetupIntegrationEmailsOnPushGated(mgr ctrl.Manager, o controller.Options) error {
	o.Gate.Register(func() {
		if err := SetupIntegrationEmailsOnPush(mgr, o); err != nil {
			mgr.GetLogger().Error(err, "unable to setup reconciler", "gvk", v1alpha1.IntegrationEmailsOnPushGroupVersionKind.String())
		}
	}, v1alpha1.IntegrationEmailsOnPushGroupVersionKind)
	return nil
}

// connector produces an ExternalClient for GitLab Project Emails on Push integrations.
type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg common.Config) projects.EmailsOnPushClient
}

// Connect creates a new GitLab client for the given managed resource.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.IntegrationEmailsOnPush)
	if !ok {
		return nil, errors.New(errNotIntegrationEmailsOnPush)
	}
	cfg, err := common.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newGitlabClientFn(*cfg)}, nil
}

// external represents the external client for GitLab Project Emails on Push integrations.
type external struct {
	client projects.EmailsOnPushClient
}

// applyEmailsOnPush applies the desired Emails on Push settings to the GitLab project.
func (e *external) applyEmailsOnPush(ctx context.Context, cr *v1alpha1.IntegrationEmailsOnPush) error {
	_, _, err := e.client.SetEmailsOnPushService(
		*cr.Spec.ForProvider.ProjectID,
		projects.GenerateSetEmailsOnPushServiceOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	return err
}

// Observe checks whether the external resource exists and whether it is up-to-date.
func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.IntegrationEmailsOnPush)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotIntegrationEmailsOnPush)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(errProjectIDMissing)
	}

	svc, res, err := e.client.GetEmailsOnPushService(
		*cr.Spec.ForProvider.ProjectID,
		gitlab.WithContext(ctx),
	)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}
	// GitLab keeps a disabled integration record, so an inactive one counts as absent.
	if svc == nil || !svc.Active || svc.Properties == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	current := cr.Spec.ForProvider.DeepCopy()
	projects.LateInitializeIntegrationEmailsOnPush(&cr.Spec.ForProvider, svc)

	cr.Status.AtProvider = projects.GenerateIntegrationEmailsOnPushObservation(svc)
	cr.Status.SetConditions(v2.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        projects.IsIntegrationEmailsOnPushUpToDate(&cr.Spec.ForProvider, svc),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

// Create creates the external resource for the GitLab Project Emails on Push integration.
func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.IntegrationEmailsOnPush)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotIntegrationEmailsOnPush)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalCreation{}, errors.New(errProjectIDMissing)
	}

	cr.Status.SetConditions(v2.Creating())

	if err := e.applyEmailsOnPush(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}
	return managed.ExternalCreation{}, nil
}

// Update updates the external resource to match the desired state.
func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.IntegrationEmailsOnPush)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotIntegrationEmailsOnPush)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalUpdate{}, errors.New(errProjectIDMissing)
	}

	if err := e.applyEmailsOnPush(ctx, cr); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}
	return managed.ExternalUpdate{}, nil
}

// Delete removes the GitLab Emails on Push integration from the project.
func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.IntegrationEmailsOnPush)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotIntegrationEmailsOnPush)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalDelete{}, errors.New(errProjectIDMissing)
	}

	_, err := e.client.DeleteEmailsOnPushService(
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
