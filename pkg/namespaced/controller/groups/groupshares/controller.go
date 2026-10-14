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

// Package groupshares reconciles GitLab group shares: one group giving the
// members of another group access.
package groupshares

import (
	"context"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/statemetrics"
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
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
	errNotGroupShare   = "managed resource is not a GroupShare"
	errMissingGroupIDs = "sharedGroupId and invitedGroupId must be set"
	errGetGroup        = "cannot get the shared GitLab group"
	errShare           = "cannot share the GitLab group"
	errUnshare         = "cannot unshare the GitLab group"
	errExpiresAt       = "cannot parse expiresAt"
)

// SetupGroupShare adds a controller that reconciles GroupShares.
func SetupGroupShare(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.GroupShareGroupKind)

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnector(&connector{kube: mgr.GetClient(), newGitlabClientFn: groups.NewGroupClient}),
		// The share is identified by its two groups; there is no external name.
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}

	if o.Features.Enabled(feature.EnableBetaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.GroupShareGroupVersionKind),
		reconcilerOpts...)

	if err := mgr.Add(statemetrics.NewMRStateRecorder(
		mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &v1alpha1.GroupShareList{}, o.MetricOptions.PollStateMetricInterval)); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.GroupShare{}).
		Complete(r)
}

// SetupGroupShareGated adds a controller with CRD gate support.
func SetupGroupShareGated(mgr ctrl.Manager, o controller.Options) error {
	o.Gate.Register(func() {
		if err := SetupGroupShare(mgr, o); err != nil {
			mgr.GetLogger().Error(err, "unable to setup reconciler", "gvk", v1alpha1.GroupShareGroupVersionKind.String())
		}
	}, v1alpha1.GroupShareGroupVersionKind)
	return nil
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg common.Config) groups.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.GroupShare)
	if !ok {
		return nil, errors.New(errNotGroupShare)
	}
	cfg, err := common.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newGitlabClientFn(*cfg)}, nil
}

type external struct {
	client groups.Client
}

// groupIDs returns the shared and the invited group of a share.
func groupIDs(cr *v1alpha1.GroupShare) (shared, invited int64, err error) {
	p := cr.Spec.ForProvider
	if p.SharedGroupID == nil || p.InvitedGroupID == nil {
		return 0, 0, errors.New(errMissingGroupIDs)
	}
	return *p.SharedGroupID, *p.InvitedGroupID, nil
}

// date formats a GitLab date the way the spec holds it.
func date(t *gitlab.ISOTime) *string {
	if t == nil {
		return nil
	}
	return new(time.Time(*t).Format(time.DateOnly))
}

// findShare returns the share of grp with the invited group, if any.
func findShare(grp *gitlab.Group, invited int64) *gitlab.SharedWithGroup {
	for i := range grp.SharedWithGroups {
		if grp.SharedWithGroups[i].GroupID == invited {
			return &grp.SharedWithGroups[i]
		}
	}
	return nil
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.GroupShare)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotGroupShare)
	}
	shared, invited, err := groupIDs(cr)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	// Without with_projects=false GitLab lists the group's projects in every
	// poll. client-go deprecates the option for v5 of the API, GitLab 19 needs it.
	grp, res, err := e.client.GetGroup(shared, &gitlab.GetGroupOptions{WithProjects: new(false)}, gitlab.WithContext(ctx)) //nolint:staticcheck
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetGroup)
	}

	found := findShare(grp, invited)
	if found == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	expires := date(found.ExpiresAt)
	cr.Status.AtProvider = v1alpha1.GroupShareObservation{
		InvitedGroupFullPath: found.GroupFullPath,
		GroupAccessLevel:     found.GroupAccessLevel,
		ExpiresAt:            expires,
	}
	cr.SetConditions(v2.Available())

	p := &cr.Spec.ForProvider
	lateInitialized := false
	if p.ExpiresAt == nil && expires != nil {
		p.ExpiresAt = expires
		lateInitialized = true
	}

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        p.GroupAccessLevel == found.GroupAccessLevel && ptr.Deref(p.ExpiresAt, "") == ptr.Deref(expires, ""),
		ResourceLateInitialized: lateInitialized,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.GroupShare)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotGroupShare)
	}
	shared, invited, err := groupIDs(cr)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	opt := &gitlab.ShareGroupWithGroupOptions{
		GroupID:     new(invited),
		GroupAccess: new(gitlab.AccessLevelValue(cr.Spec.ForProvider.GroupAccessLevel)),
	}
	if cr.Spec.ForProvider.ExpiresAt != nil {
		t, err := time.Parse(time.DateOnly, *cr.Spec.ForProvider.ExpiresAt)
		if err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, errExpiresAt)
		}
		opt.ExpiresAt = new(gitlab.ISOTime(t))
	}

	_, _, err = e.client.ShareGroupWithGroup(shared, opt, gitlab.WithContext(ctx))
	return managed.ExternalCreation{}, errors.Wrap(err, errShare)
}

// Update replaces the share, because GitLab cannot change one. The members of
// the invited group lose access between the two requests, and until a later
// reconcile if creating the new share fails.
func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	if _, err := e.Delete(ctx, mg); err != nil {
		return managed.ExternalUpdate{}, err
	}
	_, err := e.Create(ctx, mg)
	return managed.ExternalUpdate{}, err
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.GroupShare)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotGroupShare)
	}
	shared, invited, err := groupIDs(cr)
	if err != nil {
		return managed.ExternalDelete{}, err
	}

	res, err := e.client.UnshareGroupFromGroup(shared, invited, gitlab.WithContext(ctx))
	if err != nil && !clients.IsResponseNotFound(res) {
		return managed.ExternalDelete{}, errors.Wrap(err, errUnshare)
	}
	return managed.ExternalDelete{}, nil
}

func (e *external) Disconnect(ctx context.Context) error {
	return nil
}
