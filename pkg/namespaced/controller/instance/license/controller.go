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

	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
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
func SetupLicense(mgr ctrl.Manager, o common.ControllerOptions) error {
	name := managed.ControllerName(v1alpha1.LicenseGroupKind)

	r := common.NewReconciler(mgr, o, name,
		resource.ManagedKind(v1alpha1.LicenseGroupVersionKind),
		managed.WithExternalConnector(&connector{
			kube:              mgr.GetClient(),
			newGitlabClientFn: instance.NewLicenseClient,
		}),
		managed.WithInitializers(),
	)

	if err := common.AddMRStateRecorder(mgr, o, &v1alpha1.LicenseList{}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.License{}).
		Complete(r)
}

// SetupLicenseGated adds a controller with CRD gate support.
func SetupLicenseGated(mgr ctrl.Manager, o common.ControllerOptions) error {
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

// observeDeletionByName parses the external name and observes the deletion.
// An empty external name means nothing was ever adopted or created.
func (e *external) observeDeletionByName(ctx context.Context, externalName string) (managed.ExternalObservation, error) {
	if externalName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	licenseID, err := strconv.ParseInt(externalName, 10, 64)
	if err != nil {
		return managed.ExternalObservation{}, errors.New(errIDNotInt)
	}
	return e.observeDeletion(ctx, licenseID)
}

// licenseKeyMatches reports whether the license key is unchanged. The key is
// only needed to detect drift of a license we manage, so it is skipped when
// observing only.
func (e *external) licenseKeyMatches(ctx context.Context, mg resource.Managed, cr *v1alpha1.License, observeOnly bool) (bool, error) {
	if observeOnly {
		return true, nil
	}
	return e.checkLicenseKeyMatch(ctx, mg, cr)
}

// isUnadoptable reports whether there is nothing to observe yet: a managed
// license without an external name has not been created.
func isUnadoptable(externalName string, observeOnly bool) bool {
	return externalName == "" && !observeOnly
}

// getCurrentLicense returns the active license, or nil if there is none.
func (e *external) getCurrentLicense(ctx context.Context) (*gitlab.License, error) {
	license, res, err := e.client.GetLicense(gitlab.WithContext(ctx))
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return nil, nil
		}
		return nil, errors.Wrap(err, errGetFailed)
	}
	return license, nil
}

// adoptLicense matches the current active license against the external name.
// GetLicense returns the current license, not a specific one by ID, so a
// different ID means ours does not exist. Without an external name
// (observe-only), the current license is adopted and the name set.
func adoptLicense(cr *v1alpha1.License, license *gitlab.License, externalName string) (lateInitialized, exists bool, err error) {
	if externalName == "" {
		meta.SetExternalName(cr, strconv.FormatInt(license.ID, 10))
		return true, true, nil
	}
	id, err := strconv.ParseInt(externalName, 10, 64)
	if err != nil {
		return false, false, errors.New(errIDNotInt)
	}
	return false, license.ID == id, nil
}

// Observe observes the external resource
func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.License)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotLicense)
	}

	externalName := meta.GetExternalName(cr)
	observeOnly := !canWriteLicense(cr)
	if isUnadoptable(externalName, observeOnly) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	// Handle deletion case separately
	if !cr.ObjectMeta.DeletionTimestamp.IsZero() {
		return e.observeDeletionByName(ctx, externalName)
	}

	// Retrieve license from GitLab API
	license, err := e.getCurrentLicense(ctx)
	if err != nil || license == nil {
		return managed.ExternalObservation{}, err
	}

	lateInitialized, exists, err := adoptLicense(cr, license, externalName)
	if err != nil || !exists {
		return managed.ExternalObservation{}, err
	}

	// The license key is only needed to detect drift of a license we manage.
	match, err := e.licenseKeyMatches(ctx, mg, cr, observeOnly)
	if err != nil || !match {
		return managed.ExternalObservation{ResourceExists: err == nil}, err
	}

	cr.Status.AtProvider = instance.GenerateLicenseObservation(license)
	cr.SetConditions(v2.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        instance.IsLicenseUpToDate(&cr.Spec.ForProvider, license),
		ResourceLateInitialized: lateInitialized,
	}, nil
}

// canWriteLicense reports whether the management policies allow creating or
// updating the license, i.e. whether a license source is required.
func canWriteLicense(cr *v1alpha1.License) bool {
	policies := cr.GetManagementPolicies()
	if len(policies) == 0 {
		return true
	}
	for _, p := range policies {
		if p == v2.ManagementActionAll || p == v2.ManagementActionCreate || p == v2.ManagementActionUpdate {
			return true
		}
	}
	return false
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

	cr.SetConditions(v2.Creating())
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
	existingLicenseKey, err := common.GetTokenValueFromLocalSecret(ctx, e.kube, mg, &v2.LocalSecretKeySelector{
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
	existingLicenseKey, err := common.GetTokenValueFromLocalSecret(ctx, e.kube, mg, &v2.LocalSecretKeySelector{
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
