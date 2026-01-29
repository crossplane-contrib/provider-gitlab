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

package license

import (
	"context"
	"strconv"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/statemetrics"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/instance/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/instance"
)

const (
	errNotLicense              = "managed resource is not a License custom resource"
	errIDNotInt                = "specified ID is not an integer"
	errGetFailed               = "cannot get Gitlab License"
	errCreateFailed            = "cannot create Gitlab License"
	errUpdateFailed            = "cannot update Gitlab License"
	errDeleteFailed            = "cannot delete Gitlab License"
	errMissingExternalName     = "external name annotation not found"
	errMissingConnectionSecret = "writeConnectionSecretToRef must be specified to receive the license key"
	errMissingLicenseKey       = "license key must be provided via spec, secret reference or endpoint configuration"

	// ConnectionDetails keys
	keyLicense = "license"
)

// SetupLicense adds a controller that reconciles instance licenses.
func SetupLicense(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.LicenseGroupKind)

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{
			kube:              mgr.GetClient(),
			newGitlabClientFn: instance.NewLicenseClient,
		}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}
	if o.Features.Enabled(feature.EnableBetaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.LicenseGroupVersionKind),
		reconcilerOpts...)

	if err := mgr.Add(statemetrics.NewMRStateRecorder(
		mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &v1alpha1.LicenseList{}, o.MetricOptions.PollStateMetricInterval)); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.License{}).
		Complete(r)
}

// SetupLicenseGated adds a controller with CRD gate support.
func SetupLicenseGated(mgr ctrl.Manager, o controller.Options) error {
	o.Gate.Register(func() {
		if err := SetupLicense(mgr, o); err != nil {
			mgr.GetLogger().Error(err, "unable to setup reconciler", "gvk", v1alpha1.LicenseGroupVersionKind.String())
		}
	}, v1alpha1.LicenseGroupVersionKind)
	return nil
}

// connector is responsible for producing an ExternalClient for Licenses
type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg common.Config) instance.LicenseClient
}

// Connect establishes a connection to the external resource
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.License)
	if !ok {
		return nil, errors.New(errNotLicense)
	}

	cfg, err := common.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

// external is the external client used to manage Gitlab Licenses.
type external struct {
	kube   client.Client
	client instance.LicenseClient
}

// observeDeletion checks if the license still exists when the resource is being deleted.
// It attempts to delete the license - if the delete returns 404, the license is already gone.
// This is necessary because GetLicense() returns the "current" license which may
// still show our license due to eventual consistency after deletion.
func (e *external) observeDeletion(ctx context.Context, licenseID int64) (managed.ExternalObservation, error) {
	res, err := e.client.DeleteLicense(licenseID, gitlab.WithContext(ctx))
	if err != nil && clients.IsResponseNotFound(res) {
		// License no longer exists - we can proceed with finalization
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errDeleteFailed)
	}
	// Return ResourceExists: false so the reconciler can proceed to unpublish and remove finalizer
	return managed.ExternalObservation{ResourceExists: false}, nil
}

// Observe observes the external resource
func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.License)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotLicense)
	}

	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	licenseID, err := strconv.ParseInt(externalName, 10, 64)
	if err != nil {
		return managed.ExternalObservation{}, errors.New(errIDNotInt)
	}

	// Handle deletion case separately
	if !cr.ObjectMeta.DeletionTimestamp.IsZero() {
		return e.observeDeletion(ctx, licenseID)
	}

	// Retrieve license from GitLab API
	license, res, err := e.client.GetLicense(gitlab.WithContext(ctx))
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}

	// GetLicense returns the current active license, not a specific one by ID.
	// If the returned license ID doesn't match our external name, our license doesn't exist.
	if license.ID != licenseID {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	match, err := e.checkLicenseKeyMatch(ctx, mg, cr)
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	if !match {
		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: false,
		}, nil
	}

	cr.Status.AtProvider = instance.GenerateLicenseObservation(license)
	cr.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        instance.IsLicenseUpToDate(&cr.Spec.ForProvider, license),
		ResourceLateInitialized: false,
	}, nil
}

// Create creates the external resource with the desired state
func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.License)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotLicense)
	}

	// Validate that connection details will be published
	if cr.Spec.WriteConnectionSecretToReference == nil {
		return managed.ExternalCreation{}, errors.New(errMissingConnectionSecret)
	}

	// Build connection details from secrets, contains the license key, and possibly endpoint info
	connectionDetails, err := e.getLicenseFromSecrets(mg, ctx, &cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errMissingLicenseKey)
	}

	// Ensure license key is present
	licenseKey, ok := connectionDetails[keyLicense]
	if !ok || len(licenseKey) == 0 {
		return managed.ExternalCreation{}, errors.New(errMissingLicenseKey)
	}

	// Call Gitlab API to add license
	license, _, err := e.client.AddLicense(
		instance.GenerateAddLicenseOptions(string(licenseKey)),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	cr.SetConditions(xpv1.Creating())
	meta.SetExternalName(cr, strconv.FormatInt(license.ID, 10))

	return managed.ExternalCreation{
		ConnectionDetails: connectionDetails,
	}, nil
}

// Update handles updates to the external resource.
// There is no way to update an existing license in the Gitlab API,
// so we will re-apply the license using the AddLicense endpoint.
func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.License)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotLicense)
	}

	// Validate that connection details will be published
	if cr.Spec.WriteConnectionSecretToReference == nil {
		return managed.ExternalUpdate{}, errors.New(errMissingConnectionSecret)
	}

	// Retrieve connection details and confirm the retrieved license key matches the one
	// saved in the connection details during creation
	connectionDetails, err := e.getLicenseFromSecrets(mg, ctx, &cr.Spec.ForProvider)
	if isErrorFetchingLicenseFromEndpoint(err) {
		// If we fail to fetch the license from the endpoint during update, we skip the update
		// as we cannot re-apply the license without it
		return managed.ExternalUpdate{}, nil
	}
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errMissingLicenseKey)
	}

	// Ensure license key is present
	licenseKey := connectionDetails[keyLicense]
	if len(licenseKey) == 0 {
		return managed.ExternalUpdate{}, errors.New(errMissingLicenseKey)
	}

	// Retrieve saved license key from connection secret
	existingLicenseKey, err := common.GetTokenValueFromLocalSecret(ctx, e.kube, mg, &xpv1.LocalSecretKeySelector{
		LocalSecretReference: *cr.Spec.WriteConnectionSecretToReference,
		Key:                  keyLicense,
	})
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "cannot get existing license key from connection secret")
	}

	// Return early if the license key has not changed
	if existingLicenseKey != nil && *existingLicenseKey == string(licenseKey) {
		return managed.ExternalUpdate{}, nil
	}

	// Update the license
	license, _, err := e.client.AddLicense(
		instance.GenerateAddLicenseOptions(string(licenseKey)),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}

	meta.SetExternalName(cr, strconv.FormatInt(license.ID, 10))

	return managed.ExternalUpdate{
		ConnectionDetails: connectionDetails,
	}, nil
}

// Delete removes the external resource
func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.License)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotLicense)
	}

	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalDelete{}, errors.New(errMissingExternalName)
	}

	licenseID, err := strconv.ParseInt(externalName, 10, 64)
	if err != nil {
		return managed.ExternalDelete{}, errors.New(errIDNotInt)
	}

	res, err := e.client.DeleteLicense(
		licenseID,
		gitlab.WithContext(ctx),
	)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalDelete{}, nil
		}
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteFailed)
	}

	return managed.ExternalDelete{}, nil
}

func (e *external) checkLicenseKeyMatch(ctx context.Context, mg resource.Managed, cr *v1alpha1.License) (bool, error) {
	// Retrieve connection details and confirm the retrieved license key matches the one
	// saved in the connection details during creation
	connectionDetails, err := e.getLicenseFromSecrets(mg, ctx, &cr.Spec.ForProvider)
	if err != nil {
		if !isErrorFetchingLicenseFromEndpoint(err) {
			return false, errors.Wrap(err, errMissingLicenseKey)
		}
		// If we fail to fetch the license from the endpoint during observe, we skip updating
		// the connection details as we cannot confirm the license key without it
		return true, nil
	}

	return e.verifyLicenseKey(ctx, mg, connectionDetails)
}

// verifyLicenseKey checks if the license key in the connection details matches the one in the secret
func (e *external) verifyLicenseKey(ctx context.Context, mg resource.Managed, connectionDetails managed.ConnectionDetails) (bool, error) {
	cr, ok := mg.(*v1alpha1.License)
	if !ok {
		return false, errors.New(errNotLicense)
	}

	if cr.Spec.WriteConnectionSecretToReference == nil {
		return false, errors.New(errMissingConnectionSecret)
	}

	// Ensure license key is present
	retrievedLicenseKey, ok := connectionDetails[keyLicense]
	if !ok || len(retrievedLicenseKey) == 0 {
		return false, errors.New(errMissingLicenseKey)
	}

	// Retrieve the current license key from the connection secret
	existingLicenseKey, err := common.GetTokenValueFromLocalSecret(ctx, e.kube, mg, &xpv1.LocalSecretKeySelector{
		LocalSecretReference: *cr.Spec.WriteConnectionSecretToReference,
		Key:                  keyLicense,
	})
	if err != nil {
		return false, errors.Wrap(err, "cannot get existing license key from connection secret")
	}

	// Return true if license keys match
	if existingLicenseKey != nil && *existingLicenseKey == string(retrievedLicenseKey) {
		return true, nil
	}
	return false, nil
}

func (e *external) Disconnect(ctx context.Context) error {
	// Disconnect is not implemented as it is a new method required by the SDK
	return nil
}
