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

package groups

import (
	"context"
	"strconv"
	"strings"
	"time"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/feature"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/statemetrics"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
	secretstoreapi "github.com/crossplane-contrib/provider-gitlab/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups"
	"github.com/crossplane-contrib/provider-gitlab/pkg/features"
)

const (
	errNotGroup          = "managed resource is not a Gitlab Group custom resource"
	errIDNotInt          = "specified ID is not an integer"
	errGetFailed         = "cannot get Gitlab Group"
	errCreateFailed      = "cannot create Gitlab Group"
	errUpdateFailed      = "cannot update Gitlab Group"
	errShareFailed       = "cannot share Gitlab Group with: %v"
	errUnshareFailed     = "cannot unshare Gitlab Group from: %v"
	errDeleteFailed      = "cannot delete Gitlab Group"
	errMissingGroupID    = "missing group ID for group to share with"
	errSWGMissingGroupID = "FOllowing SharedWithGroup is missing GroupID: %v"
	errLateInitialize    = "Error during LateInitialization: "
)

// SetupGroup adds a controller that reconciles Groups.
func SetupGroup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.GroupKubernetesGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), secretstoreapi.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabClientFn: groups.NewGroupClient}),
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
		resource.ManagedKind(v1alpha1.GroupKubernetesGroupVersionKind),
		reconcilerOpts...)

	if err := mgr.Add(statemetrics.NewMRStateRecorder(
		mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &v1alpha1.GroupList{}, o.MetricOptions.PollStateMetricInterval)); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Group{}).
		Complete(r)
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg clients.Config) groups.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Group)
	if !ok {
		return nil, errors.New(errNotGroup)
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
	cr, ok := mg.(*v1alpha1.Group)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotGroup)
	}

	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	groupID, err := strconv.Atoi(externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.New(errIDNotInt)
	}

	//nolint:staticcheck // Keeping this for backward compatibility during deprecation
	cr.Spec.ForProvider.EmailsEnabled = lateInitializeEmailsEnabled(cr.Spec.ForProvider.EmailsEnabled, cr.Spec.ForProvider.EmailsDisabled)

	grp, res, err := e.client.GetGroup(groupID, nil)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}

	current := cr.Spec.ForProvider.DeepCopy()

	err = lateInitialize(&cr.Spec.ForProvider, grp)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}
	isResourceLateInitialized := !cmp.Equal(current, &cr.Spec.ForProvider)

	cr.Status.AtProvider = groups.GenerateObservation(grp)
	cr.Status.SetConditions(xpv1.Available())
	isUpToDate, err := isGroupUpToDate(&cr.Spec.ForProvider, grp)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isUpToDate,
		ResourceLateInitialized: isResourceLateInitialized,
		ConnectionDetails:       managed.ConnectionDetails{"runnersToken": []byte(grp.RunnersToken)},
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Group)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotGroup)
	}

	grp, _, err := e.client.CreateGroup(
		groups.GenerateCreateGroupOptions(cr.Name, &cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	meta.SetExternalName(cr, strconv.Itoa(grp.ID))
	return managed.ExternalCreation{}, nil
}

//nolint:gocyclo
func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Group)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotGroup)
	}
	grp, _, err := e.client.UpdateGroup(
		meta.GetExternalName(cr),
		groups.GenerateEditGroupOptions(cr.Name, &cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}

	if len(cr.Spec.ForProvider.SharedWithGroups) > 0 {
		for _, sh := range cr.Spec.ForProvider.SharedWithGroups {
			if sh.GroupID == nil {
				return managed.ExternalUpdate{}, errors.New(errMissingGroupID)
			}
			if notShared(*sh.GroupID, grp) {
				opt := gitlab.ShareGroupWithGroupOptions{
					GroupID:     sh.GroupID,
					GroupAccess: (*gitlab.AccessLevelValue)(&sh.GroupAccessLevel),
				}
				if sh.ExpiresAt != nil {
					opt.ExpiresAt = (*gitlab.ISOTime)(&sh.ExpiresAt.Time)
				}
				_, _, err = e.client.ShareGroupWithGroup(grp.ID, &opt)
				if err != nil {
					return managed.ExternalUpdate{}, errors.Wrapf(err, errShareFailed, *sh.GroupID)
				}
			}
		}
	}

	if len(grp.SharedWithGroups) > 0 {
		for _, sh := range grp.SharedWithGroups {
			isNotUnshared, err := notUnshared(sh.GroupID, cr.Spec.ForProvider.SharedWithGroups)
			if err != nil {
				return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
			}
			if isNotUnshared {
				_, err = e.client.UnshareGroupFromGroup(grp.ID, sh.GroupID)
				if err != nil {
					return managed.ExternalUpdate{}, errors.Wrapf(err, errUnshareFailed, sh.GroupID)
				}
			}
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.Group)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotGroup)
	}

	_, err := e.client.DeleteGroup(meta.GetExternalName(cr), &gitlab.DeleteGroupOptions{}, gitlab.WithContext(ctx))
	// if the group is for some reason already marked for deletion, we ignore the error and continue to delete the group permanently
	if err != nil && !strings.Contains(err.Error(), "Group has been already marked for deletion") {
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteFailed)
	}

	// permanent deletion is only available on subgroups; when executed against top-level groups the backend will return an error
	isSubGroup := cr.Status.AtProvider.FullPath != nil && *cr.Status.AtProvider.FullPath != cr.Spec.ForProvider.Path
	if cr.Spec.ForProvider.PermanentlyRemove != nil && *cr.Spec.ForProvider.PermanentlyRemove && isSubGroup {
		_, err = e.client.DeleteGroup(meta.GetExternalName(cr), &gitlab.DeleteGroupOptions{
			PermanentlyRemove: cr.Spec.ForProvider.PermanentlyRemove,
			FullPath:          cr.Spec.ForProvider.FullPathToRemove,
		}, gitlab.WithContext(ctx))
	}
	return managed.ExternalDelete{}, errors.Wrap(err, errDeleteFailed)
}

func (e *external) Disconnect(ctx context.Context) error {
	// Disconnect is not implemented as it is a new method required by the SDK
	return nil
}

// isGroupUpToDate checks whether there is a change in any of the modifiable fields.
func isGroupUpToDate(p *v1alpha1.GroupParameters, g *gitlab.Group) (bool, error) { //nolint:gocyclo
	if p.Name != nil && !cmp.Equal(*p.Name, g.Name) {
		return false, nil
	}
	if !cmp.Equal(p.Path, g.Path) {
		return false, nil
	}
	if !cmp.Equal(p.Description, clients.StringToPtr(g.Description)) {
		return false, nil
	}
	if !clients.IsBoolEqualToBoolPtr(p.MembershipLock, g.MembershipLock) {
		return false, nil
	}
	if (p.Visibility != nil) && (!cmp.Equal(string(*p.Visibility), string(g.Visibility))) {
		return false, nil
	}
	if (p.ProjectCreationLevel != nil) && (!cmp.Equal(string(*p.ProjectCreationLevel), string(g.ProjectCreationLevel))) {
		return false, nil
	}
	if (p.SubGroupCreationLevel != nil) && (!cmp.Equal(string(*p.SubGroupCreationLevel), string(g.SubGroupCreationLevel))) {
		return false, nil
	}
	if !clients.IsBoolEqualToBoolPtr(p.ShareWithGroupLock, g.ShareWithGroupLock) {
		return false, nil
	}
	if !clients.IsBoolEqualToBoolPtr(p.RequireTwoFactorAuth, g.RequireTwoFactorAuth) {
		return false, nil
	}
	if !clients.IsIntEqualToIntPtr(p.TwoFactorGracePeriod, g.TwoFactorGracePeriod) {
		return false, nil
	}
	if !clients.IsBoolEqualToBoolPtr(p.AutoDevopsEnabled, g.AutoDevopsEnabled) {
		return false, nil
	}
	if !clients.IsBoolEqualToBoolPtr(p.EmailsEnabled, g.EmailsEnabled) {
		return false, nil
	}
	if !clients.IsBoolEqualToBoolPtr(p.MentionsDisabled, g.MentionsDisabled) {
		return false, nil
	}
	if !clients.IsBoolEqualToBoolPtr(p.LFSEnabled, g.LFSEnabled) {
		return false, nil
	}
	if !clients.IsBoolEqualToBoolPtr(p.RequestAccessEnabled, g.RequestAccessEnabled) {
		return false, nil
	}
	if !clients.IsIntEqualToIntPtr(p.ParentID, g.ParentID) {
		return false, nil
	}
	if !clients.IsIntEqualToIntPtr(p.SharedRunnersMinutesLimit, g.SharedRunnersMinutesLimit) {
		return false, nil
	}
	if !clients.IsIntEqualToIntPtr(p.ExtraSharedRunnersMinutesLimit, g.ExtraSharedRunnersMinutesLimit) {
		return false, nil
	}
	if ok, err := isSharedWithGroupsUpToDate(p, g); err != nil || !ok {
		return false, err
	}
	return true, nil
}

func isSharedWithGroupsUpToDate(cr *v1alpha1.GroupParameters, in *gitlab.Group) (bool, error) {
	if len(cr.SharedWithGroups) != len(in.SharedWithGroups) {
		return false, nil
	}

	inIDs := make(map[int]any)
	for _, v := range in.SharedWithGroups {
		inIDs[v.GroupID] = nil
	}

	crIDs := make(map[int]any)
	for _, v := range cr.SharedWithGroups {
		if v.GroupID == nil {
			return false, errors.Errorf(errSWGMissingGroupID, v)
		}
		crIDs[*v.GroupID] = nil
	}

	for ID := range inIDs {
		_, ok := crIDs[ID]
		if !ok {
			return false, nil
		}
	}

	for ID := range crIDs {
		_, ok := inIDs[ID]
		if !ok {
			return false, nil
		}
	}

	return true, nil
}

// lateInitialize fills the empty fields in the group spec with the
// values seen in gitlab.Group.
func lateInitialize(in *v1alpha1.GroupParameters, group *gitlab.Group) error { //nolint:gocyclo
	if group == nil {
		return nil
	}
	if in.Path == "" && group.Path != "" {
		in.Path = group.Path
	}

	in.Description = clients.LateInitializeStringPtr(in.Description, group.Description)
	in.Visibility = lateInitializeVisibilityValue(in.Visibility, group.Visibility)
	in.ProjectCreationLevel = lateInitializeProjectCreationLevelValue(in.ProjectCreationLevel, group.ProjectCreationLevel)
	in.SubGroupCreationLevel = lateInitializeSubGroupCreationLevelValue(in.SubGroupCreationLevel, group.SubGroupCreationLevel)

	if len(group.SharedWithGroups) > 0 && len(in.SharedWithGroups) > 0 {
		if err := lateInitializeSharedWithGroups(in, group); err != nil {
			return err
		}
	}
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
	if in.EmailsEnabled == nil {
		in.EmailsEnabled = &group.EmailsEnabled
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
	return nil
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

func lateInitializeEmailsEnabled(in *bool, from *bool) *bool {
	if in == nil && from != nil {
		value := !(*from)
		return &value
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

func lateInitializeSharedWithGroups(cr *v1alpha1.GroupParameters, in *gitlab.Group) error {
	inMap := map[int]struct {
		GroupID          int             `json:"group_id"`
		GroupName        string          `json:"group_name"`
		GroupFullPath    string          `json:"group_full_path"`
		GroupAccessLevel int             `json:"group_access_level"`
		ExpiresAt        *gitlab.ISOTime `json:"expires_at"`
		MemberRoleID     int             `json:"member_role_id"`
	}{}

	for _, inswg := range in.SharedWithGroups {
		inMap[inswg.GroupID] = inswg
	}

	for i := range cr.SharedWithGroups {
		if cr.SharedWithGroups[i].GroupID == nil {
			return errors.Errorf(errSWGMissingGroupID, cr.SharedWithGroups[i])
		}
		inswg, ok := inMap[*cr.SharedWithGroups[i].GroupID]
		if ok {
			if cr.SharedWithGroups[i].ExpiresAt == nil && inswg.ExpiresAt != nil {
				cr.SharedWithGroups[i].ExpiresAt = &metav1.Time{Time: time.Time(*inswg.ExpiresAt)}
			}
		}
	}
	return nil
}

func notUnshared(groupID int, sh []v1alpha1.SharedWithGroups) (bool, error) {
	for _, cr := range sh {
		if cr.GroupID == nil {
			return false, errors.Errorf(errSWGMissingGroupID, cr)
		}
		if groupID == *cr.GroupID {
			return false, nil
		}
	}
	return true, nil
}

func notShared(groupID int, grp *gitlab.Group) bool {
	for _, in := range grp.SharedWithGroups {
		if in.GroupID == groupID {
			return false
		}
	}
	return true
}
