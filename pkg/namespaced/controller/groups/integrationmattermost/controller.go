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

package integrationmattermost

import (
	"context"

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

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/groups"
)

const (
	errNotIntegrationMattermost = "managed resource is not a Gitlab group integration mattermost custom resource"
	errGroupIDMissing           = "GroupID is missing"
	errWebHookMissing           = "cannot resolve Mattermost webhook from secret reference"
	errGetFailed                = "cannot get Gitlab group integration mattermost"
	errCreateFailed             = "cannot create Gitlab group integration mattermost"
	errUpdateFailed             = "cannot update Gitlab group integration mattermost"
	errDeleteFailed             = "cannot delete Gitlab group integration mattermost"

	// webHookSecretHashName identifies the webhook in its secret hash annotation.
	webHookSecretHashName = "webhook"
)

// SetupIntegrationMattermost adds a controller that reconciles GitLab Group Mattermost integrations.
func SetupIntegrationMattermost(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.IntegrationMattermostGroupKind)

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnector(&connector{kube: mgr.GetClient(), newGitlabClientFn: groups.NewMattermostClient}),
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
		resource.ManagedKind(v1alpha1.IntegrationMattermostGroupVersionKind),
		reconcilerOpts...,
	)

	if err := mgr.Add(statemetrics.NewMRStateRecorder(
		mgr.GetClient(),
		o.Logger,
		o.MetricOptions.MRStateMetrics,
		&v1alpha1.IntegrationMattermostList{},
		o.MetricOptions.PollStateMetricInterval,
	)); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.IntegrationMattermost{}).
		Complete(r)
}

// SetupIntegrationMattermostGated adds a controller with CRD gate support for SafeStart.
func SetupIntegrationMattermostGated(mgr ctrl.Manager, o controller.Options) error {
	o.Gate.Register(func() {
		if err := SetupIntegrationMattermost(mgr, o); err != nil {
			mgr.GetLogger().Error(err, "unable to setup reconciler", "gvk", v1alpha1.IntegrationMattermostGroupVersionKind.String())
		}
	}, v1alpha1.IntegrationMattermostGroupVersionKind)
	return nil
}

// connector produces an ExternalClient for GitLab Group Mattermost integrations.
type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg common.Config) groups.MattermostClient
}

// Connect creates a new GitLab client for the given managed resource.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.IntegrationMattermost)
	if !ok {
		return nil, errors.New(errNotIntegrationMattermost)
	}
	cfg, err := common.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	hasher, err := common.NewSecretHasher(ctx, c.kube, cr, cfg)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg), hasher: hasher}, nil
}

// external represents the external client for GitLab Group Mattermost integrations.
type external struct {
	kube   client.Client
	client groups.MattermostClient
	hasher *common.SecretHasher
}

// resolveWebHook reads the Mattermost webhook from the referenced secret.
func (e *external) resolveWebHook(ctx context.Context, cr *v1alpha1.IntegrationMattermost) (string, error) {
	webhook, err := common.GetTokenValueFromLocalSecret(ctx, e.kube, cr, &cr.Spec.ForProvider.WebHookSecretRef)
	if err != nil {
		return "", errors.Wrap(err, errWebHookMissing)
	}
	if webhook == nil {
		return "", nil
	}
	return *webhook, nil
}

// applyMattermost applies the desired Mattermost settings to the GitLab group.
// It returns the applied webhook so that its hash can be recorded.
func (e *external) applyMattermost(ctx context.Context, cr *v1alpha1.IntegrationMattermost) (string, error) {
	webhook, err := e.resolveWebHook(ctx, cr)
	if err != nil {
		return "", err
	}
	_, _, err = e.client.SetGroupMattermostIntegration(
		*cr.Spec.ForProvider.GroupID,
		groups.GenerateGroupMattermostIntegrationOptions(&cr.Spec.ForProvider, webhook),
		gitlab.WithContext(ctx),
	)
	return webhook, err
}

// isWebHookUpToDate reports whether the webhook currently held in the secret is
// the one last applied to GitLab, as GitLab never returns the webhook. The check
// needs a secret hash key, and is skipped while deleting so that a removed
// secret cannot block deletion.
func (e *external) isWebHookUpToDate(ctx context.Context, cr *v1alpha1.IntegrationMattermost) (bool, error) {
	if !e.hasher.Enabled() || meta.WasDeleted(cr) {
		return true, nil
	}
	webhook, err := e.resolveWebHook(ctx, cr)
	if err != nil {
		return false, err
	}
	return e.hasher.IsUpToDate(cr, webHookSecretHashName, webhook), nil
}

// Observe checks whether the external resource exists and whether it is up-to-date.
func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.IntegrationMattermost)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotIntegrationMattermost)
	}

	if cr.Spec.ForProvider.GroupID == nil {
		return managed.ExternalObservation{}, errors.New(errGroupIDMissing)
	}

	mattermost, res, err := e.client.GetGroupMattermostIntegration(
		*cr.Spec.ForProvider.GroupID,
		gitlab.WithContext(ctx),
	)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}
	if mattermost == nil || !mattermost.Active {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	cr.Status.AtProvider = groups.GenerateIntegrationMattermostObservation(mattermost)
	cr.Status.SetConditions(v2.Available())

	upToDate := groups.IsIntegrationMattermostUpToDate(&cr.Spec.ForProvider, mattermost)
	if upToDate {
		if upToDate, err = e.isWebHookUpToDate(ctx, cr); err != nil {
			return managed.ExternalObservation{}, err
		}
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

// Create creates the external resource for the GitLab Group Mattermost integration.
func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.IntegrationMattermost)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotIntegrationMattermost)
	}

	if cr.Spec.ForProvider.GroupID == nil {
		return managed.ExternalCreation{}, errors.New(errGroupIDMissing)
	}

	cr.Status.SetConditions(v2.Creating())

	webhook, err := e.applyMattermost(ctx, cr)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}
	// The managed reconciler persists annotations set during Create.
	e.hasher.Set(cr, webHookSecretHashName, webhook)
	return managed.ExternalCreation{}, nil
}

// Update updates the external resource to match the desired state.
func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.IntegrationMattermost)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotIntegrationMattermost)
	}

	if cr.Spec.ForProvider.GroupID == nil {
		return managed.ExternalUpdate{}, errors.New(errGroupIDMissing)
	}

	webhook, err := e.applyMattermost(ctx, cr)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}
	// The managed reconciler does not persist annotations set during Update.
	if err := e.hasher.Persist(ctx, e.kube, cr, webHookSecretHashName, webhook); err != nil {
		return managed.ExternalUpdate{}, err
	}
	return managed.ExternalUpdate{}, nil
}

// Delete removes the GitLab Mattermost integration from the group.
func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.IntegrationMattermost)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotIntegrationMattermost)
	}

	if cr.Spec.ForProvider.GroupID == nil {
		return managed.ExternalDelete{}, errors.New(errGroupIDMissing)
	}

	_, err := e.client.DeleteGroupMattermostIntegration(
		*cr.Spec.ForProvider.GroupID,
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
