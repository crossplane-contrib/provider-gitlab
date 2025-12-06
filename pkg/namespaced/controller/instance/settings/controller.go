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
package settings

import (
	"context"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/errors"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/statemetrics"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/instance/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/instance"
)

const (
	errNotSettings     = "managed resource is not a Gitlab settings custom resource"
	errGetFailed       = "cannot get Gitlab settings"
	errCreateFailed    = "cannot create Gitlab settings"
	errUpdateFailed    = "cannot update Gitlab settings"
	errDeleteFailed    = "cannot delete Gitlab settings"
	errSettingsMissing = "Settings are missing"
)

// SetupApplicationSettings adds a controller that reconciles GitLab Instance Settings.
func SetupApplicationSettings(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.ApplicationSettingsGroupKind)

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabClientFn: instance.NewApplicationSettingsClient}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}

	if o.Features.Enabled(feature.EnableBetaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.ApplicationSettingsGroupVersionKind),
		reconcilerOpts...)

	if err := mgr.Add(statemetrics.NewMRStateRecorder(
		mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &v1alpha1.ApplicationSettingsList{}, o.MetricOptions.PollStateMetricInterval)); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.ApplicationSettings{}).
		Complete(r)
}

// SetupApplicationSettingsGated adds a controller with CRD gate support.
func SetupApplicationSettingsGated(mgr ctrl.Manager, o controller.Options) error {
	o.Gate.Register(func() {
		if err := SetupApplicationSettings(mgr, o); err != nil {
			mgr.GetLogger().Error(err, "unable to setup reconciler", "gvk", v1alpha1.ApplicationSettingsGroupVersionKind.String())
		}
	}, v1alpha1.ApplicationSettingsGroupVersionKind)
	return nil
}

// connector is responsible for producing an ExternalClient for Gitlab Instance Settings
type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg common.Config) instance.ApplicationSettingsClient
}

// Connect creates a new Gitlab client for the given managed resource.
// It fetches the provider configuration and uses it to instantiate the client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.ApplicationSettings)
	if !ok {
		return nil, errors.New(errNotSettings)
	}
	cfg, err := common.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

// external represents the external client for Gitlab Instance Settings.
type external struct {
	kube   client.Client
	client instance.ApplicationSettingsClient
}

// Observe checks if the Gitlab Instance Settings external resource exists and
// if it is up to date.
func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.ApplicationSettings)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotSettings)
	}

	settings, res, err := e.client.GetSettings(gitlab.WithContext(ctx))
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}

	// Deleting: only need to determine external resource still exists.
	if !cr.ObjectMeta.DeletionTimestamp.IsZero() {
		return managed.ExternalObservation{ResourceExists: true}, nil
	}

	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        instance.IsApplicationSettingsUpToDate(&cr.Spec.ForProvider, settings),
		ResourceLateInitialized: false,
	}, nil
}

// Create creates the external resource for Gitlab Instance Settings.
// Upon creation, Gitlab Instance Settings cannot be uniquely identified,
// so we simply call UpdateSettings with the desired parameters.
func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ApplicationSettings)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotSettings)
	}

	cr.Status.SetConditions(xpv1.Creating())
	_, _, err := e.client.UpdateSettings(
		instance.GenerateUpdateApplicationSettingsOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx))
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	return managed.ExternalCreation{}, nil
}

// Update updates the external resource to match the desired state.
// As Gitlab Instance Settings cannot be uniquely identified, we simply call
// UpdateSettings with the desired parameters.
func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.ApplicationSettings)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotSettings)
	}

	cr.Status.SetConditions(xpv1.Creating())
	_, _, err := e.client.UpdateSettings(
		instance.GenerateUpdateApplicationSettingsOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx))
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}

	return managed.ExternalUpdate{}, nil
}

// Delete removes the application settings resource. As Gitlab instance settings
// cannot be deleted, we simply return success.
func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	_, ok := mg.(*v1alpha1.ApplicationSettings)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotSettings)
	}

	return managed.ExternalDelete{}, nil
}

func (e *external) Disconnect(ctx context.Context) error {
	// Disconnect is not implemented as it is a new method required by the SDK
	return nil
}
