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

package integrationharbor

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
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects"
)

const (
	errNotIntegrationHarbor = "managed resource is not a Gitlab integration harbor custom resource"
	errProjectIDMissing     = "ProjectID is missing"
	errPasswordMissing      = "cannot resolve Harbor password from secret reference"
	errGetFailed            = "cannot get Gitlab integration harbor"
	errCreateFailed         = "cannot create Gitlab integration harbor"
	errUpdateFailed         = "cannot update Gitlab integration harbor"
	errDeleteFailed         = "cannot delete Gitlab integration harbor"
)

// SetupIntegrationHarbor adds a controller that reconciles GitLab Project Harbor integrations.
func SetupIntegrationHarbor(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.IntegrationHarborGroupKind)

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnector(&connector{kube: mgr.GetClient(), newGitlabClientFn: projects.NewHarborClient}),
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
		resource.ManagedKind(v1alpha1.IntegrationHarborGroupVersionKind),
		reconcilerOpts...,
	)

	if err := mgr.Add(statemetrics.NewMRStateRecorder(
		mgr.GetClient(),
		o.Logger,
		o.MetricOptions.MRStateMetrics,
		&v1alpha1.IntegrationHarborList{},
		o.MetricOptions.PollStateMetricInterval,
	)); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.IntegrationHarbor{}).
		Complete(r)
}

// SetupIntegrationHarborGated adds a controller with CRD gate support for SafeStart.
func SetupIntegrationHarborGated(mgr ctrl.Manager, o controller.Options) error {
	o.Gate.Register(func() {
		if err := SetupIntegrationHarbor(mgr, o); err != nil {
			mgr.GetLogger().Error(err, "unable to setup reconciler", "gvk", v1alpha1.IntegrationHarborGroupVersionKind.String())
		}
	}, v1alpha1.IntegrationHarborGroupVersionKind)
	return nil
}

// connector produces an ExternalClient for GitLab Project Harbor integrations.
type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg common.Config) projects.HarborClient
}

// Connect creates a new GitLab client for the given managed resource.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.IntegrationHarbor)
	if !ok {
		return nil, errors.New(errNotIntegrationHarbor)
	}
	cfg, err := common.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

// external represents the external client for GitLab Project Harbor integrations.
type external struct {
	kube   client.Client
	client projects.HarborClient
}

// resolvePassword reads the Harbor password from the referenced secret.
func (e *external) resolvePassword(ctx context.Context, cr *v1alpha1.IntegrationHarbor) (string, error) {
	password, err := common.GetTokenValueFromLocalSecret(ctx, e.kube, cr, &cr.Spec.ForProvider.PasswordSecretRef)
	if err != nil {
		return "", errors.Wrap(err, errPasswordMissing)
	}
	if password == nil {
		return "", nil
	}
	return *password, nil
}

// applyHarbor applies the desired Harbor settings to the GitLab project.
func (e *external) applyHarbor(ctx context.Context, cr *v1alpha1.IntegrationHarbor) error {
	password, err := e.resolvePassword(ctx, cr)
	if err != nil {
		return err
	}
	_, _, err = e.client.SetHarborService(
		*cr.Spec.ForProvider.ProjectID,
		projects.GenerateSetHarborServiceOptions(&cr.Spec.ForProvider, password),
		gitlab.WithContext(ctx),
	)
	return err
}

// Observe checks whether the external resource exists and whether it is up-to-date.
func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.IntegrationHarbor)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotIntegrationHarbor)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(errProjectIDMissing)
	}

	harbor, res, err := e.client.GetHarborService(
		*cr.Spec.ForProvider.ProjectID,
		gitlab.WithContext(ctx),
	)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}
	if harbor == nil || !harbor.Active {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	current := cr.Spec.ForProvider.DeepCopy()
	projects.LateInitializeIntegrationHarbor(&cr.Spec.ForProvider, harbor)

	cr.Status.AtProvider = projects.GenerateIntegrationHarborObservation(harbor)
	cr.Status.SetConditions(v2.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        projects.IsIntegrationHarborUpToDate(&cr.Spec.ForProvider, harbor),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

// Create creates the external resource for the GitLab Project Harbor integration.
func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.IntegrationHarbor)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotIntegrationHarbor)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalCreation{}, errors.New(errProjectIDMissing)
	}

	cr.Status.SetConditions(v2.Creating())

	if err := e.applyHarbor(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}
	return managed.ExternalCreation{}, nil
}

// Update updates the external resource to match the desired state.
func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.IntegrationHarbor)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotIntegrationHarbor)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalUpdate{}, errors.New(errProjectIDMissing)
	}

	if err := e.applyHarbor(ctx, cr); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}
	return managed.ExternalUpdate{}, nil
}

// Delete removes the GitLab Harbor integration from the project.
func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.IntegrationHarbor)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotIntegrationHarbor)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalDelete{}, errors.New(errProjectIDMissing)
	}

	_, err := e.client.DeleteHarborService(
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
