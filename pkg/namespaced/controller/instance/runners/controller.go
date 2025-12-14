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

package runners

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/instance/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
	runners "github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/runners"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/users"
)

const (
	errNotRunner               = "managed resource is not a Runner custom resource"
	errIDNotInt                = "specified ID is not an integer"
	errGetFailed               = "cannot get Gitlab Runner"
	errCreateFailed            = "cannot create Gitlab Runner"
	errUpdateFailed            = "cannot update Gitlab Runner"
	errDeleteFailed            = "cannot delete Gitlab Runner"
	errMissingExternalName     = "external name annotation not found"
	errMissingConnectionSecret = "writeConnectionSecretToRef or publishConnectionDetailsTo must be specified to receive the runner token"
)

// SetupRunner adds a controller that reconciles instance runners.
func SetupRunner(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.RunnerGroupKind)

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{
			kube:              mgr.GetClient(),
			newGitlabClientFn: runners.NewRunnerClient,
			newRunnerClientFn: users.NewRunnerClient,
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
		resource.ManagedKind(v1alpha1.RunnerGroupVersionKind),
		reconcilerOpts...)

	if err := mgr.Add(statemetrics.NewMRStateRecorder(
		mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &v1alpha1.RunnerList{}, o.MetricOptions.PollStateMetricInterval)); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Runner{}).
		Complete(r)
}

// SetupRunnerGated adds a controller with CRD gate support.
func SetupRunnerGated(mgr ctrl.Manager, o controller.Options) error {
	o.Gate.Register(func() {
		if err := SetupRunner(mgr, o); err != nil {
			mgr.GetLogger().Error(err, "unable to setup reconciler", "gvk", v1alpha1.RunnerGroupVersionKind.String())
		}
	}, v1alpha1.RunnerGroupVersionKind)
	return nil
}

// connector is responsible for producing an ExternalClient for Runners
type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg common.Config) runners.RunnerClient
	newRunnerClientFn func(cfg common.Config) users.RunnerClient
}

// Connect establishes a connection to the external resource
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Runner)
	if !ok {
		return nil, errors.New(errNotRunner)
	}

	cfg, err := common.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg), userRunnerClient: c.newRunnerClientFn(*cfg)}, nil
}

// external is the external client used to manage Gitlab Runners.
type external struct {
	kube             client.Client
	client           runners.RunnerClient
	userRunnerClient users.RunnerClient
}

// Update updates the external resource to match the desired state
func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Runner)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRunner)
	}

	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	runnerID, err := strconv.Atoi(externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.New(errIDNotInt)
	}

	runner, res, err := e.client.GetRunnerDetails(runnerID, nil)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}

	// we need to make sure the token expiration time is preserved as it is not returned by the API
	tokenExpiresAt := cr.Status.AtProvider.CommonRunnerObservation.TokenExpiresAt
	cr.Status.AtProvider = runners.GenerateInstanceRunnerObservation(runner)
	cr.Status.AtProvider.CommonRunnerObservation.TokenExpiresAt = tokenExpiresAt
	cr.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        runners.IsRunnerUpToDate(&cr.Spec.ForProvider.CommonRunnerParameters, runner),
		ResourceLateInitialized: false,
	}, nil
}

// Create creates the external resource with the desired state
func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Runner)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRunner)
	}

	// Validate that connection details will be published
	if cr.Spec.WriteConnectionSecretToReference == nil {
		return managed.ExternalCreation{}, errors.New(errMissingConnectionSecret)
	}

	runner, _, err := e.userRunnerClient.CreateUserRunner(
		users.GenerateInstanceRunnerOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	meta.SetExternalName(cr, strconv.Itoa(runner.ID))

	if runner.TokenExpiresAt != nil {
		t := metav1.NewTime(*runner.TokenExpiresAt)
		cr.Status.AtProvider.CommonRunnerObservation.TokenExpiresAt = &t
	} else {
		cr.Status.AtProvider.CommonRunnerObservation.TokenExpiresAt = nil
	}

	cr.SetConditions(xpv1.Creating())

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{
			"token": []byte(runner.Token),
		},
	}, nil
}

// Update handles updates to the external resource.
func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Runner)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRunner)
	}

	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalUpdate{}, errors.New(errMissingExternalName)
	}

	runnerID, err := strconv.Atoi(externalName)
	if err != nil {
		return managed.ExternalUpdate{}, errors.New(errIDNotInt)
	}

	_, _, err = e.client.UpdateRunnerDetails(
		runnerID,
		runners.GenerateEditRunnerOptions(&cr.Spec.ForProvider.CommonRunnerParameters),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}

	return managed.ExternalUpdate{}, nil
}

// Delete removes the external resource
func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.Runner)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotRunner)
	}

	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalDelete{}, errors.New(errMissingExternalName)
	}

	runnerID, err := strconv.Atoi(externalName)
	if err != nil {
		return managed.ExternalDelete{}, errors.New(errIDNotInt)
	}

	_, err = e.client.DeleteRegisteredRunnerByID(
		runnerID,
		gitlab.WithContext(ctx),
	)

	return managed.ExternalDelete{}, errors.Wrap(err, errDeleteFailed)
}

func (e *external) Disconnect(ctx context.Context) error {
	// Disconnect is not implemented as it is a new method required by the SDK
	return nil
}
