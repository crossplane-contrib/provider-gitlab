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

package serviceaccounts

import (
	"context"
	"strconv"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/errors"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/statemetrics"
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects"
)

const (
	errNotServiceAccount = "managed resource is not a Gitlab service account custom resource"
	errGetFailed         = "cannot get Gitlab service account"
	errCreateFailed      = "cannot create Gitlab service account"
	errUpdateFailed      = "cannot update Gitlab service account"
	errDeleteFailed      = "cannot delete Gitlab service account"
	errIDNotInt          = "specified ID is not an integer"
	errMissingProjectID  = "missing Spec.ForProvider.ProjectID"
)

// SetupServiceAccount adds a controller that reconciles GitLab project Service Accounts.
func SetupServiceAccount(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.ServiceAccountGroupKind)

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnector(&connector{kube: mgr.GetClient(), newGitlabClientFn: projects.NewServiceAccountClient}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}

	if o.Features.Enabled(feature.EnableBetaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.ServiceAccountGroupVersionKind),
		reconcilerOpts...)

	if err := mgr.Add(statemetrics.NewMRStateRecorder(
		mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &v1alpha1.ServiceAccountList{}, o.MetricOptions.PollStateMetricInterval)); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.ServiceAccount{}).
		Complete(r)
}

// SetupServiceAccountGated adds a controller with CRD gate support.
func SetupServiceAccountGated(mgr ctrl.Manager, o controller.Options) error {
	o.Gate.Register(func() {
		if err := SetupServiceAccount(mgr, o); err != nil {
			mgr.GetLogger().Error(err, "unable to setup reconciler", "gvk", v1alpha1.ServiceAccountGroupVersionKind.String())
		}
	}, v1alpha1.ServiceAccountGroupVersionKind)
	return nil
}

// connector is responsible for producing an ExternalClient for Gitlab Service Accounts
type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg common.Config) projects.ServiceAccountClient
}

// Connect creates a new Gitlab client for the given managed resource.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.ServiceAccount)
	if !ok {
		return nil, errors.New(errNotServiceAccount)
	}
	cfg, err := common.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

// external represents the external client for Gitlab Service Accounts
type external struct {
	kube   client.Client
	client projects.ServiceAccountClient
}

// Observe checks if the Gitlab Service Account external resource exists and whether
// if it is up to date.
func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.ServiceAccount)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotServiceAccount)
	}

	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	serviceAccountID, err := strconv.ParseInt(externalName, 10, 64)
	if err != nil {
		return managed.ExternalObservation{}, errors.New(errIDNotInt)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(errMissingProjectID)
	}

	serviceAccount, res, err := projects.GetProjectServiceAccount(
		e.client,
		*cr.Spec.ForProvider.ProjectID,
		serviceAccountID,
		gitlab.WithContext(ctx),
	)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}

	// There is no get-by-id endpoint for project service accounts, so a missing
	// account surfaces as a nil result rather than a 404.
	if serviceAccount == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	cr.Status.AtProvider = projects.GenerateServiceAccountObservation(serviceAccount)
	cr.Status.SetConditions(v2.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        projects.IsServiceAccountUpToDate(&cr.Spec.ForProvider, serviceAccount),
		ResourceLateInitialized: false,
	}, nil
}

// Create creates the external resource for Gitlab project Service Account.
func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ServiceAccount)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotServiceAccount)
	}

	cr.Status.SetConditions(v2.Creating())
	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalCreation{}, errors.New(errMissingProjectID)
	}

	serviceAccount, _, err := e.client.CreateProjectServiceAccount(
		*cr.Spec.ForProvider.ProjectID,
		projects.GenerateServiceAccountCreateOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx))
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	meta.SetExternalName(cr, strconv.FormatInt(serviceAccount.ID, 10))
	cr.Status.AtProvider = projects.GenerateServiceAccountObservation(serviceAccount)

	return managed.ExternalCreation{}, nil
}

// Update updates the external resource to match the desired state.
func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.ServiceAccount)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotServiceAccount)
	}

	cr.Status.SetConditions(v2.Creating())
	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalUpdate{}, errors.New(errMissingProjectID)
	}

	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalUpdate{}, errors.New(errCreateFailed)
	}

	serviceAccountID, err := strconv.ParseInt(externalName, 10, 64)
	if err != nil {
		return managed.ExternalUpdate{}, errors.New(errIDNotInt)
	}

	serviceAccount, _, err := e.client.UpdateProjectServiceAccount(
		*cr.Spec.ForProvider.ProjectID,
		serviceAccountID,
		projects.GenerateUpdateServiceAccountOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx))
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}

	cr.Status.AtProvider = projects.GenerateServiceAccountObservation(serviceAccount)

	return managed.ExternalUpdate{}, nil
}

// Delete removes the service account resource.
// WARNING: deleting a service account user may also delete resources owned by it.
func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.ServiceAccount)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotServiceAccount)
	}

	externalName := meta.GetExternalName(mg)
	if externalName == "" {
		return managed.ExternalDelete{}, nil
	}
	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalDelete{}, errors.New(errMissingProjectID)
	}

	serviceAccountID, err := strconv.ParseInt(externalName, 10, 64)
	if err != nil {
		return managed.ExternalDelete{}, errors.New(errIDNotInt)
	}

	_, err = e.client.DeleteProjectServiceAccount(
		*cr.Spec.ForProvider.ProjectID,
		serviceAccountID,
		&gitlab.DeleteProjectServiceAccountOptions{},
		gitlab.WithContext(ctx))
	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteFailed)
	}

	return managed.ExternalDelete{}, nil
}

func (e *external) Disconnect(ctx context.Context) error {
	// Disconnect is not implemented as it is a new method required by the SDK
	return nil
}
