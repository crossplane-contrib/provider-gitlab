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

package badges

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
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects"
)

const (
	errNotBadge         = "managed resource is not a Gitlab badge custom resource"
	errGetFailed        = "cannot get Gitlab badge"
	errCreateFailed     = "cannot create Gitlab badge"
	errDeleteFailed     = "cannot delete Gitlab badge"
	errProjectIDMissing = "ProjectID is missing"
	errWrongIDSet       = "ID must be set to reference existing badge if not empty"
)

// SetupBadge adds a controller that reconciles ProjectBadges.
func SetupBadge(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.BadgeGroupKind)

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabClientFn: projects.NewBadgeClient}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}

	if o.Features.Enabled(feature.EnableBetaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.BadgeGroupVersionKind),
		reconcilerOpts...)

	if err := mgr.Add(statemetrics.NewMRStateRecorder(
		mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &v1alpha1.BadgeList{}, o.MetricOptions.PollStateMetricInterval)); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Badge{}).
		Complete(r)
}

// SetupBadgeGated adds a controller with CRD gate support.
func SetupBadgeGated(mgr ctrl.Manager, o controller.Options) error {
	o.Gate.Register(func() {
		if err := SetupBadge(mgr, o); err != nil {
			mgr.GetLogger().Error(err, "unable to setup reconciler", "gvk", v1alpha1.BadgeGroupVersionKind.String())
		}
	}, v1alpha1.BadgeGroupVersionKind)
	return nil
}

// connector is responsible for producing an ExternalClient for ProjectBadges
type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg common.Config) projects.BadgeClient
}

// Connect establishes a connection to the external system.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Badge)
	if !ok {
		return nil, errors.New(errNotBadge)
	}
	cfg, err := common.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

// external is an external client for ProjectBadges
type external struct {
	kube   client.Client
	client projects.BadgeClient
}

// Observe retrieves the external resource (badge) and its current state.
func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Badge)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotBadge)
	}
	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(errProjectIDMissing)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	badgeID, err := strconv.Atoi(meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalObservation{}, errors.New(errNotBadge)
	}

	badge, res, err := e.client.GetProjectBadge(*cr.Spec.ForProvider.ProjectID, badgeID, gitlab.WithContext(ctx))
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	lateInitializeBadge(&cr.Spec.ForProvider, badge)

	cr.Status.AtProvider = projects.GenerateBadgeObservation(badge)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        projects.IsBadgeUpToDate(&cr.Spec.ForProvider, badge),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

// Create creates the external resource (badge).
func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Badge)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotBadge)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalCreation{}, errors.Wrap(errors.New("ProjectId must be set directly or via reference"), errCreateFailed)
	}

	cr.Status.SetConditions(xpv1.Creating())

	// if ID is already set, check if it does exist, else create a new one
	if cr.Spec.ForProvider.ID != nil {
		id := *cr.Spec.ForProvider.ID
		badge, res, err := e.client.GetProjectBadge(*cr.Spec.ForProvider.ProjectID, id)
		if err != nil || clients.IsResponseNotFound(res) {
			return managed.ExternalCreation{}, errors.Wrap(err, errWrongIDSet)
		}
		// found it, set the external name and return
		meta.SetExternalName(cr, strconv.Itoa(badge.ID))
		return managed.ExternalCreation{}, nil
	}

	badge, _, err := e.client.AddProjectBadge(
		*cr.Spec.ForProvider.ProjectID,
		projects.GenerateAddProjectBadgeOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	meta.SetExternalName(cr, strconv.Itoa(badge.ID))

	return managed.ExternalCreation{}, nil
}

// Update updates the external resource (badge) to match the managed resource.
func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Badge)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotBadge)
	}

	badgeID, err := strconv.Atoi(meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalUpdate{}, errors.New(errNotBadge)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalUpdate{}, errors.New(errProjectIDMissing)
	}

	_, _, err = e.client.EditProjectBadge(
		*cr.Spec.ForProvider.ProjectID,
		badgeID,
		projects.GenerateEditProjectBadgeOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errCreateFailed)
	}
	return managed.ExternalUpdate{}, nil
}

// Delete deletes the external resource (badge).
func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.Badge)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotBadge)
	}

	badgeID, err := strconv.Atoi(meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalDelete{}, errors.New(errNotBadge)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalDelete{}, errors.New(errProjectIDMissing)
	}

	_, deleteError := e.client.DeleteProjectBadge(
		*cr.Spec.ForProvider.ProjectID,
		badgeID,
		gitlab.WithContext(ctx),
	)

	return managed.ExternalDelete{}, errors.Wrap(deleteError, errDeleteFailed)
}

func (e *external) Disconnect(ctx context.Context) error {
	// Disconnect is not implemented as it is a new method required by the SDK
	return nil
}

// lateInitializeBadge fills the empty fields in the badge spec with the
// values seen in gitlab badge.
func lateInitializeBadge(in *v1alpha1.BadgeParameters, badge *gitlab.ProjectBadge) {
	if badge == nil {
		return
	}

	if in.Name == nil {
		in.Name = &badge.Name
	}
}
