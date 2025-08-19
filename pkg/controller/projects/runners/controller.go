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
	"slices"
	"strconv"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/feature"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/statemetrics"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	secretstoreapi "github.com/crossplane-contrib/provider-gitlab/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	runners "github.com/crossplane-contrib/provider-gitlab/pkg/clients/runners"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/users"
	"github.com/crossplane-contrib/provider-gitlab/pkg/features"
)

const (
	errNotRunner               = "managed resource is not a Runner custom resource"
	errIDNotInt                = "specified ID is not an integer"
	errGetFailed               = "cannot get Gitlab Runner"
	errCreateFailed            = "cannot create Gitlab Runner"
	errUpdateFailed            = "cannot update Gitlab Runner"
	errDeleteFailed            = "cannot delete Gitlab Runner"
	errRunnertNotFound         = "cannot find Gitlab Runner"
	errMissingProjectID        = "missing Spec.ForProvider.ProjectID"
	errMissingExternalName     = "external name annotation not found"
	errMissingConnectionSecret = "writeConnectionSecretToRef or publishConnectionDetailsTo must be specified to receive the runner token"
)

// SetupRunner adds a controller that reconciles samlgrouplinks.
func SetupRunner(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.RunnerGroupKind)
	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}

	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), secretstoreapi.StoreConfigGroupVersionKind))
	}

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
		managed.WithConnectionPublishers(cps...),
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

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg clients.Config) runners.RunnerClient
	newRunnerClientFn func(cfg clients.Config) users.RunnerClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Runner)
	if !ok {
		return nil, errors.New(errNotRunner)
	}

	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg), userRunnerClient: c.newRunnerClientFn(*cfg)}, nil
}

type external struct {
	kube             client.Client
	client           runners.RunnerClient
	userRunnerClient users.RunnerClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Runner)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRunner)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(errMissingProjectID)
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
	cr.Status.AtProvider = runners.GenerateProjectRunnerObservation(runner)
	cr.Status.AtProvider.CommonRunnerObservation.TokenExpiresAt = tokenExpiresAt
	cr.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isRunnerUpToDate(&cr.Spec.ForProvider, runner),
		ResourceLateInitialized: false,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Runner)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRunner)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalCreation{}, errors.New(errMissingProjectID)
	}

	// Validate that connection details will be published
	if cr.Spec.WriteConnectionSecretToReference == nil && cr.Spec.PublishConnectionDetailsTo == nil {
		return managed.ExternalCreation{}, errors.New(errMissingConnectionSecret)
	}

	runner, _, err := e.userRunnerClient.CreateUserRunner(
		users.GenerateProjectRunnerOptions(&cr.Spec.ForProvider),
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

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Runner)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRunner)
	}
	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalUpdate{}, errors.New(errMissingProjectID)
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

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.Runner)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotRunner)
	}
	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalDelete{}, errors.New(errMissingProjectID)
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

func isRunnerUpToDate(p *v1alpha1.RunnerParameters, r *gitlab.RunnerDetails) bool { //nolint:gocyclo
	if p.Description != nil && *p.Description != r.Description {
		return false
	}
	if p.Paused != nil && *p.Paused != r.Paused {
		return false
	}
	if p.Locked != nil && *p.Locked != r.Locked {
		return false
	}
	if p.RunUntagged != nil && *p.RunUntagged != r.RunUntagged {
		return false
	}
	if p.TagList != nil && !slices.Equal(*p.TagList, r.TagList) {
		return false
	}
	if p.AccessLevel != nil && *p.AccessLevel != r.AccessLevel {
		return false
	}
	if p.MaximumTimeout != nil && *p.MaximumTimeout != r.MaximumTimeout {
		return false
	}
	if p.MaintenanceNote != nil && *p.MaintenanceNote != r.MaintenanceNote {
		return false
	}

	return true
}
