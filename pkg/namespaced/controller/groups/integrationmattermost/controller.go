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

	"github.com/crossplane/crossplane-runtime/v2/pkg/errors"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
	"k8s.io/utils/ptr"
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
	errWebHookRefMissing        = "webhookSecretRef is required unless useInheritedSettings is true"
	errWebHookMissing           = "cannot resolve Mattermost webhook from secret reference"
	errWebHookEmpty             = "Mattermost webhook from secret reference is empty"
	errGetFailed                = "cannot get Gitlab group integration mattermost"
	errCreateFailed             = "cannot create Gitlab group integration mattermost"
	errUpdateFailed             = "cannot update Gitlab group integration mattermost"
	errDeleteFailed             = "cannot delete Gitlab group integration mattermost"
)

// SetupIntegrationMattermost adds a controller that reconciles GitLab Group Mattermost integrations.
func SetupIntegrationMattermost(mgr ctrl.Manager, o common.ControllerOptions) error {
	name := managed.ControllerName(v1alpha1.IntegrationMattermostGroupKind)

	r := common.NewReconciler(mgr, o, name,
		resource.ManagedKind(v1alpha1.IntegrationMattermostGroupVersionKind),
		managed.WithExternalConnector(&connector{kube: mgr.GetClient(), newGitlabClientFn: groups.NewMattermostClient}),
		managed.WithInitializers(),
	)

	if err := common.AddMRStateRecorder(mgr, o, &v1alpha1.IntegrationMattermostList{}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.IntegrationMattermost{}).
		Complete(r)
}

// SetupIntegrationMattermostGated adds a controller with CRD gate support for SafeStart.
func SetupIntegrationMattermostGated(mgr ctrl.Manager, o common.ControllerOptions) error {
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
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

// external represents the external client for GitLab Group Mattermost integrations.
type external struct {
	kube   client.Client
	client groups.MattermostClient
}

// resolveWebHook reads the Mattermost webhook from the referenced secret.
func (e *external) resolveWebHook(ctx context.Context, cr *v1alpha1.IntegrationMattermost) (string, error) {
	if cr.Spec.ForProvider.WebHookSecretRef == nil {
		return "", errors.New(errWebHookRefMissing)
	}
	webhook, err := common.GetTokenValueFromLocalSecret(ctx, e.kube, cr, cr.Spec.ForProvider.WebHookSecretRef)
	if err != nil {
		return "", errors.Wrap(err, errWebHookMissing)
	}
	// GitLab rejects an empty webhook with a bare 400 Bad Request.
	if webhook == nil || *webhook == "" {
		return "", errors.New(errWebHookEmpty)
	}
	return *webhook, nil
}

// exists reports whether the observed integration is managed by cr.
//
// GitLab propagates the settings of the closest ancestor with custom settings
// (or else of the instance) to every group without custom settings. Such an
// inherited integration exists when cr asks to inherit, even if inactive, so
// that it is not created over and over. It cannot be removed, so it is
// considered gone while cr is being deleted.
func exists(cr *v1alpha1.IntegrationMattermost, mattermost *groups.GroupMattermostIntegration) bool {
	switch {
	case mattermost == nil:
		return false
	case mattermost.Inherited && meta.WasDeleted(cr):
		return false
	case mattermost.Inherited && ptr.Deref(cr.Spec.ForProvider.UseInheritedSettings, false):
		return true
	default:
		return mattermost.Active
	}
}

// applyMattermost applies the desired Mattermost settings to the GitLab group.
func (e *external) applyMattermost(ctx context.Context, cr *v1alpha1.IntegrationMattermost) error {
	// GitLab ignores the webhook while inheriting, so the secret is not needed.
	var webhook string
	if !ptr.Deref(cr.Spec.ForProvider.UseInheritedSettings, false) {
		var err error
		if webhook, err = e.resolveWebHook(ctx, cr); err != nil {
			return err
		}
	}
	_, _, err := e.client.SetGroupMattermostIntegration(
		*cr.Spec.ForProvider.GroupID,
		groups.GenerateGroupMattermostIntegrationOptions(&cr.Spec.ForProvider, webhook),
		gitlab.WithContext(ctx),
	)
	return err
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

	// Late initialize useInheritedSettings from an inherited integration, so that a group
	// inheriting its settings is not detached from them by a spec that leaves it unset.
	lateInitialized := false
	if mattermost != nil && mattermost.Inherited && cr.Spec.ForProvider.UseInheritedSettings == nil && !meta.WasDeleted(cr) {
		cr.Spec.ForProvider.UseInheritedSettings = ptr.To(true)
		lateInitialized = true
	}

	if !exists(cr, mattermost) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	cr.Status.AtProvider = groups.GenerateIntegrationMattermostObservation(mattermost)
	cr.Status.SetConditions(v2.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: groups.IsIntegrationMattermostUpToDate(&cr.Spec.ForProvider, mattermost),

		ResourceLateInitialized: lateInitialized,
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

	if err := e.applyMattermost(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}
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

	if err := e.applyMattermost(ctx, cr); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
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
