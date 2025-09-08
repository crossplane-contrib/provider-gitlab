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

	"github.com/crossplane/crossplane-runtime/v2/apis/common"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiCluster "github.com/crossplane-contrib/provider-gitlab/apis/cluster/groups/v1alpha1"
	apiNamespaced "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	sharedGroupsV1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/shared/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups"
)

const (
	ErrNotGroup          = "managed resource is not a Gitlab Group custom resource"
	ErrIDNotInt          = "specified ID is not an integer"
	ErrGetFailed         = "cannot get Gitlab Group"
	ErrCreateFailed      = "cannot create Gitlab Group"
	ErrUpdateFailed      = "cannot update Gitlab Group"
	ErrShareFailed       = "cannot share Gitlab Group with: %v"
	ErrUnshareFailed     = "cannot unshare Gitlab Group from: %v"
	ErrDeleteFailed      = "cannot delete Gitlab Group"
	ErrMissingGroupID    = "missing group ID for group to share with"
	ErrSWGMissingGroupID = "FOllowing SharedWithGroup is missing GroupID: %v"
	ErrLateInitialize    = "Error during LateInitialization: "
)

type External struct {
	Client groups.Client
	Kube   client.Client
}

type options struct {
	externalName     string
	parameters       *sharedGroupsV1alpha1.GroupParameters
	atProvider       *sharedGroupsV1alpha1.GroupObservation
	setConditions    func(c ...common.Condition)
	sharedWithGroups []*sharedGroupsV1alpha1.SharedWithGroups
	mg               resource.Managed
}

func getSharedWithGroupsFromCluster(in []apiCluster.SharedWithGroups) []*sharedGroupsV1alpha1.SharedWithGroups {
	out := make([]*sharedGroupsV1alpha1.SharedWithGroups, len(in))
	for i := range in {
		out[i] = &in[i].SharedWithGroups
	}
	return out
}

func getSharedWithGroupsFromNamespaced(in []apiNamespaced.SharedWithGroups) []*sharedGroupsV1alpha1.SharedWithGroups {
	out := make([]*sharedGroupsV1alpha1.SharedWithGroups, len(in))
	for i := range in {
		out[i] = &in[i].SharedWithGroups
	}
	return out
}

func deepCopySharedWithGroups(in []*sharedGroupsV1alpha1.SharedWithGroups) []*sharedGroupsV1alpha1.SharedWithGroups {
	out := make([]*sharedGroupsV1alpha1.SharedWithGroups, len(in))
	for i, swg := range in {
		out[i] = swg.DeepCopy()
	}
	return out
}

func (e *External) extractOptions(mg resource.Managed) (*options, error) {
	switch cr := mg.(type) {
	case *apiCluster.Group:
		return &options{
			externalName:     meta.GetExternalName(cr),
			parameters:       &cr.Spec.ForProvider.GroupParameters,
			atProvider:       &cr.Status.AtProvider,
			setConditions:    cr.Status.SetConditions,
			sharedWithGroups: getSharedWithGroupsFromCluster(cr.Spec.ForProvider.SharedWithGroups),
			mg:               mg,
		}, nil
	case *apiNamespaced.Group:
		return &options{
			externalName:     meta.GetExternalName(cr),
			parameters:       &cr.Spec.ForProvider.GroupParameters,
			atProvider:       &cr.Status.AtProvider,
			setConditions:    cr.Status.SetConditions,
			sharedWithGroups: getSharedWithGroupsFromNamespaced(cr.Spec.ForProvider.SharedWithGroups),
			mg:               mg,
		}, nil
	default:
		return nil, errors.New(ErrNotGroup)
	}
}

//nolint:gocyclo
func (e *External) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	o, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if o.externalName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	groupID, err := strconv.Atoi(o.externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.New(ErrIDNotInt)
	}

	o.parameters.EmailsEnabled = lateInitializeEmailsEnabled(o.parameters.EmailsEnabled, o.parameters.EmailsDisabled) //nolint:staticcheck

	grp, res, err := e.Client.GetGroup(groupID, nil)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, ErrGetFailed)
	}

	// Check if the group is in a pending deletion state and either remove the
	// finalizer if specified or keep tracking it.
	//
	// Mark the resource as unavailable if the group is in a deletion state but
	// managed resource is not.
	if grp.MarkedForDeletionOn != nil {
		if meta.WasDeleted(o.mg) {
			if ptr.Deref(o.parameters.RemoveFinalizerOnPendingDeletion, false) {
				return managed.ExternalObservation{}, nil
			}
			o.setConditions(xpv1.Deleting().WithMessage("Group is in pending deletion state"))
		} else {
			o.setConditions(xpv1.Unavailable().WithMessage("Group is in pending deletion state but this managed resource is not"))
		}
	} else {
		o.setConditions(xpv1.Available())
	}

	currentParameters := o.parameters.DeepCopy()
	currentSharedWithGroups := deepCopySharedWithGroups(o.sharedWithGroups)

	if err := lateInitialize(o.parameters, grp, o.sharedWithGroups); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, ErrGetFailed)
	}

	isResourceLateInitialized := !cmp.Equal(currentParameters, o.parameters) || !cmp.Equal(currentSharedWithGroups, o.sharedWithGroups)

	*o.atProvider = groups.GenerateObservation(grp)
	isUpToDate := isGroupUpToDate(o.parameters, grp)

	// Check SharedWithGroups separately since it's not in Common
	if isUpToDate {
		isUpToDate, err = isSharedWithGroupsUpToDate(o.sharedWithGroups, grp)
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, ErrGetFailed)
		}
	}

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isUpToDate,
		ResourceLateInitialized: isResourceLateInitialized,
		ConnectionDetails:       managed.ConnectionDetails{"runnersToken": []byte(grp.RunnersToken)},
	}, nil
}

func (e *External) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	o, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	grp, _, err := e.Client.CreateGroup(
		groups.GenerateCreateGroupOptions(o.mg.GetName(), o.parameters),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, ErrCreateFailed)
	}

	meta.SetExternalName(o.mg, strconv.Itoa(grp.ID))
	return managed.ExternalCreation{}, nil
}

//nolint:gocyclo
func (e *External) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	o, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	grp, _, err := e.Client.UpdateGroup(
		o.externalName,
		groups.GenerateEditGroupOptions(o.mg.GetName(), o.parameters),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, ErrUpdateFailed)
	}

	if len(o.sharedWithGroups) > 0 {
		err := iterateSharedWithGroups(o.sharedWithGroups, func(groupID *int, groupAccessLevel int, expiresAt *metav1.Time) error {
			if groupID == nil {
				return errors.New(ErrMissingGroupID)
			}
			if notShared(*groupID, grp) {
				opt := gitlab.ShareGroupWithGroupOptions{
					GroupID:     groupID,
					GroupAccess: (*gitlab.AccessLevelValue)(&groupAccessLevel),
				}
				if expiresAt != nil {
					opt.ExpiresAt = (*gitlab.ISOTime)(&expiresAt.Time)
				}
				_, _, err = e.Client.ShareGroupWithGroup(grp.ID, &opt)
				if err != nil {
					return errors.Wrapf(err, ErrShareFailed, *groupID)
				}
			}
			return nil
		})
		if err != nil {
			return managed.ExternalUpdate{}, err
		}
	}

	if len(grp.SharedWithGroups) > 0 {
		for _, sh := range grp.SharedWithGroups {
			isNotUnshared, err := notUnshared(sh.GroupID, o.sharedWithGroups)
			if err != nil {
				return managed.ExternalUpdate{}, errors.Wrap(err, ErrUpdateFailed)
			}
			if isNotUnshared {
				_, err = e.Client.UnshareGroupFromGroup(grp.ID, sh.GroupID)
				if err != nil {
					return managed.ExternalUpdate{}, errors.Wrapf(err, ErrUnshareFailed, sh.GroupID)
				}
			}
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (e *External) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	o, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalDelete{}, err
	}

	_, err = e.Client.DeleteGroup(o.externalName, &gitlab.DeleteGroupOptions{}, gitlab.WithContext(ctx))
	// if the group is for some reason already marked for deletion, we ignore the error and continue to delete the group permanently
	if err != nil && !strings.Contains(err.Error(), "Group has been already marked for deletion") {
		return managed.ExternalDelete{}, errors.Wrap(err, ErrDeleteFailed)
	}

	// permanent deletion is only available on subgroups; when executed against top-level groups the backend will return an error
	isSubGroup := o.atProvider.FullPath != nil && *o.atProvider.FullPath != o.parameters.Path
	if o.parameters.PermanentlyRemove != nil && *o.parameters.PermanentlyRemove && isSubGroup {
		_, err = e.Client.DeleteGroup(o.externalName, &gitlab.DeleteGroupOptions{
			PermanentlyRemove: o.parameters.PermanentlyRemove,
			FullPath:          o.parameters.FullPathToRemove,
		}, gitlab.WithContext(ctx))
	}
	return managed.ExternalDelete{}, errors.Wrap(err, ErrDeleteFailed)
}

// isGroupUpToDate checks whether there is a change in any of the modifiable fields.
func isGroupUpToDate(p *sharedGroupsV1alpha1.GroupParameters, g *gitlab.Group) bool { //nolint:gocyclo
	if p.Name != nil && !cmp.Equal(*p.Name, g.Name) {
		return false
	}
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
	if (p.ProjectCreationLevel != nil) && (!cmp.Equal(string(*p.ProjectCreationLevel), string(g.ProjectCreationLevel))) {
		return false
	}
	if (p.SubGroupCreationLevel != nil) && (!cmp.Equal(string(*p.SubGroupCreationLevel), string(g.SubGroupCreationLevel))) {
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
	if !clients.IsBoolEqualToBoolPtr(p.EmailsEnabled, g.EmailsEnabled) {
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

func isSharedWithGroupsUpToDate(cr []*sharedGroupsV1alpha1.SharedWithGroups, in *gitlab.Group) (bool, error) {
	if len(cr) != len(in.SharedWithGroups) {
		return false, nil
	}

	inIDs := make(map[int]any)
	for _, v := range in.SharedWithGroups {
		inIDs[v.GroupID] = nil
	}

	crIDs := make(map[int]any)
	for _, v := range cr {
		if v.GroupID == nil {
			return false, errors.Errorf(ErrSWGMissingGroupID, v)
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
func lateInitialize(in *sharedGroupsV1alpha1.GroupParameters, group *gitlab.Group, sharedWithGroups []*sharedGroupsV1alpha1.SharedWithGroups) error { //nolint:gocyclo
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
	// Only late-initialize runner limits if no shared groups configuration,
	// otherwise we would trigger an update of that group (even if minutes is 0).
	// That update is not supported.
	if len(group.SharedWithGroups) == 0 && len(sharedWithGroups) == 0 {
		if in.SharedRunnersMinutesLimit == nil {
			in.SharedRunnersMinutesLimit = &group.SharedRunnersMinutesLimit
		}
		if in.ExtraSharedRunnersMinutesLimit == nil {
			in.ExtraSharedRunnersMinutesLimit = &group.ExtraSharedRunnersMinutesLimit
		}
	} else {
		inMap := map[int]gitlab.SharedWithGroup{}
		for _, inswg := range group.SharedWithGroups {
			inMap[inswg.GroupID] = inswg
		}

		for i := range sharedWithGroups {
			if sharedWithGroups[i].GroupID == nil {
				return errors.Errorf(ErrSWGMissingGroupID, sharedWithGroups[i])
			}
			inswg, ok := inMap[*sharedWithGroups[i].GroupID]
			if ok {
				if sharedWithGroups[i].ExpiresAt == nil && inswg.ExpiresAt != nil {
					sharedWithGroups[i].ExpiresAt = &metav1.Time{Time: time.Time(*inswg.ExpiresAt)}
				}
			}
		}
	}

	return nil
}

// lateInitializeVisibilityValue returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func lateInitializeVisibilityValue(in *sharedGroupsV1alpha1.VisibilityValue, from gitlab.VisibilityValue) *sharedGroupsV1alpha1.VisibilityValue {
	if in == nil && from != "" {
		return (*sharedGroupsV1alpha1.VisibilityValue)(&from)
	}
	return in
}

// lateInitializeSubGroupCreationLevelValue returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func lateInitializeSubGroupCreationLevelValue(in *sharedGroupsV1alpha1.SubGroupCreationLevelValue, from gitlab.SubGroupCreationLevelValue) *sharedGroupsV1alpha1.SubGroupCreationLevelValue {
	if in == nil && from != "" {
		return (*sharedGroupsV1alpha1.SubGroupCreationLevelValue)(&from)
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
func lateInitializeProjectCreationLevelValue(in *sharedGroupsV1alpha1.ProjectCreationLevelValue, from gitlab.ProjectCreationLevelValue) *sharedGroupsV1alpha1.ProjectCreationLevelValue {
	if in == nil && from != "" {
		return (*sharedGroupsV1alpha1.ProjectCreationLevelValue)(&from)
	}
	return in
}

func notUnshared(groupID int, sh []*sharedGroupsV1alpha1.SharedWithGroups) (bool, error) {
	for _, cr := range sh {
		if cr.GroupID == nil {
			return false, errors.Errorf(ErrSWGMissingGroupID, cr)
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

// Helper functions to work with SharedWithGroups
func iterateSharedWithGroups(swg []*sharedGroupsV1alpha1.SharedWithGroups, fn func(groupID *int, groupAccessLevel int, expiresAt *metav1.Time) error) error {
	for _, sh := range swg {
		if err := fn(sh.GroupID, sh.GroupAccessLevel, sh.ExpiresAt); err != nil {
			return err
		}
	}
	return nil
}

func (e *External) Disconnect(ctx context.Context) error {
	return nil
}
