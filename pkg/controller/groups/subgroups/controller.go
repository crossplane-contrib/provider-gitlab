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

package subgroups

import (
	"context"
	"strconv"

	"github.com/xanzy/go-gitlab"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups"
)

const (
	errNotSubGroup  = "managed resource is not a Gitlab SubGroup custom resource"
	errGetFailed    = "cannot get Gitlab SubGroup"
	errCreateFailed = "cannot create Gitlab SubGroup"
	errUpdateFailed = "cannot update Gitlab SubGroup"
	errDeleteFailed = "cannot delete Gitlab SubGroup"
)

// SetupSubGroup adds a controller that reconciles Groups.
func SetupSubGroup(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.SubGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.SubGroup{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.SubGroupKubernetesGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabClientFn: groups.NewSubGroupClient}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg clients.Config) groups.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.SubGroup)
	if !ok {
		return nil, errors.New(errNotSubGroup)
	}
	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

type external struct {
	kube   client.Client
	client groups.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.SubGroup)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotSubGroup)
	}

	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	subGroupID, err := strconv.Atoi(externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.New(errNotSubGroup)
	}

	grp, _, err := e.client.GetGroup(subGroupID, nil)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(groups.IsErrorGroupNotFound, err), errGetFailed)
	}

	current := cr.Spec.ForProvider.DeepCopy()

	lateInitialize(&cr.Spec.ForProvider, grp)
	resourceLateInitialized := !cmp.Equal(current, &cr.Spec.ForProvider)

	cr.Status.AtProvider = groups.GenerateSubGroupObservation(grp)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isSubGroupUpToDate(&cr.Spec.ForProvider, grp),
		ResourceLateInitialized: resourceLateInitialized,
		ConnectionDetails:       managed.ConnectionDetails{"runnersToken": []byte(grp.RunnersToken)},
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.SubGroup)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotSubGroup)
	}

	grp, _, err := e.client.CreateGroup(
		groups.GenerateCreateSubGroupOptions(cr.Name, &cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	meta.SetExternalName(cr, strconv.Itoa(grp.ID))
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.SubGroup)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotSubGroup)
	}
	_, _, err := e.client.UpdateGroup(
		meta.GetExternalName(cr),
		groups.GenerateEditSubGroupOptions(cr.Name, &cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.SubGroup)
	if !ok {
		return errors.New(errNotSubGroup)
	}

	_, err := e.client.DeleteGroup(meta.GetExternalName(cr), gitlab.WithContext(ctx))
	return errors.Wrap(err, errDeleteFailed)
}

// isSubGroupUpToDate checks whether there is a change in any of the modifiable fields.
func isSubGroupUpToDate(p *v1alpha1.SubGroupParameters, g *gitlab.Group) bool { // nolint:gocyclo
	if !cmp.Equal(p.Path, g.Path) {
		return false
	}
	if !cmp.Equal(p.Description, clients.StringToPtr(g.Description)) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.MembershipLock, g.MembershipLock) {
		return false
	}
	if (p.Visibility != nil) && (!cmp.Equal(string(*p.Visibility), string(g.Visibility))) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.ShareWithGroupLock, g.ShareWithGroupLock) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.RequireTwoFactorAuth, g.RequireTwoFactorAuth) {
		return false
	}
	if !clients.IsIntEqualToIntPtr(p.TwoFactorGracePeriod, g.TwoFactorGracePeriod) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.AutoDevopsEnabled, g.AutoDevopsEnabled) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.EmailsDisabled, g.EmailsDisabled) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.MentionsDisabled, g.MentionsDisabled) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.LFSEnabled, g.LFSEnabled) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.RequestAccessEnabled, g.RequestAccessEnabled) {
		return false
	}
	if !clients.IsIntEqualToIntPtr(p.ParentID, g.ParentID) {
		return false
	}
	if !clients.IsIntEqualToIntPtr(p.SharedRunnersMinutesLimit, g.SharedRunnersMinutesLimit) {
		return false
	}
	if !clients.IsIntEqualToIntPtr(p.ExtraSharedRunnersMinutesLimit, g.ExtraSharedRunnersMinutesLimit) {
		return false
	}
	return true
}

// lateInitialize fills the empty fields in the subgroup spec with the
// values seen in gitlab.Group.
func lateInitialize(in *v1alpha1.SubGroupParameters, group *gitlab.Group) { // nolint:gocyclo
	if group == nil {
		return
	}

	if in.Path == "" && group.Path != "" {
		in.Path = group.Path
	}

	in.Description = clients.LateInitializeStringPtr(in.Description, group.Description)
	in.Visibility = lateInitializeVisibilityValue(in.Visibility, group.Visibility)
	in.ProjectCreationLevel = lateInitializeProjectCreationLevelValue(in.ProjectCreationLevel, group.ProjectCreationLevel)
	in.SubGroupCreationLevel = lateInitializeSubGroupCreationLevelValue(in.SubGroupCreationLevel, group.SubGroupCreationLevel)

	if in.MembershipLock == nil {
		in.MembershipLock = &group.MembershipLock
	}

	if in.ShareWithGroupLock == nil {
		in.ShareWithGroupLock = &group.ShareWithGroupLock
	}

	if in.RequireTwoFactorAuth == nil {
		in.RequireTwoFactorAuth = &group.RequireTwoFactorAuth
	}

	if in.TwoFactorGracePeriod == nil {
		in.TwoFactorGracePeriod = &group.TwoFactorGracePeriod
	}

	if in.AutoDevopsEnabled == nil {
		in.AutoDevopsEnabled = &group.AutoDevopsEnabled
	}

	if in.EmailsDisabled == nil {
		in.EmailsDisabled = &group.EmailsDisabled
	}
	if in.MentionsDisabled == nil {
		in.MentionsDisabled = &group.MentionsDisabled
	}
	if in.LFSEnabled == nil {
		in.LFSEnabled = &group.LFSEnabled
	}
	if in.RequestAccessEnabled == nil {
		in.RequestAccessEnabled = &group.RequestAccessEnabled
	}
	if in.ParentID == nil {
		in.ParentID = &group.ParentID
	}
	if in.SharedRunnersMinutesLimit == nil {
		in.SharedRunnersMinutesLimit = &group.SharedRunnersMinutesLimit
	}
	if in.ExtraSharedRunnersMinutesLimit == nil {
		in.ExtraSharedRunnersMinutesLimit = &group.ExtraSharedRunnersMinutesLimit
	}
}

// lateInitializeVisibilityValue returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func lateInitializeVisibilityValue(in *v1alpha1.VisibilityValue, from gitlab.VisibilityValue) *v1alpha1.VisibilityValue {
	if in == nil && from != "" {
		return (*v1alpha1.VisibilityValue)(&from)
	}
	return in
}

// lateInitializeSubGroupCreationLevelValue returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func lateInitializeSubGroupCreationLevelValue(in *v1alpha1.SubGroupCreationLevelValue, from gitlab.SubGroupCreationLevelValue) *v1alpha1.SubGroupCreationLevelValue {
	if in == nil && from != "" {
		return (*v1alpha1.SubGroupCreationLevelValue)(&from)
	}
	return in
}

// lateInitializeProjectCreationLevelValue returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func lateInitializeProjectCreationLevelValue(in *v1alpha1.ProjectCreationLevelValue, from gitlab.ProjectCreationLevelValue) *v1alpha1.ProjectCreationLevelValue {
	if in == nil && from != "" {
		return (*v1alpha1.ProjectCreationLevelValue)(&from)
	}
	return in
}
