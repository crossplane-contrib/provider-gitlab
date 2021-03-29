/*
Copyright 2020 The Crossplane Authors.

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
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
)

const (
	errGroupNotFound = "404 Group Not Found"
)

// Client defines Gitlab Group service operations
type Client interface {
	GetGroup(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error)
	CreateGroup(opt *gitlab.CreateGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error)
	UpdateGroup(gid interface{}, opt *gitlab.UpdateGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error)
	DeleteGroup(gid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewGroupClient returns a new Gitlab Group service
func NewGroupClient(cfg clients.Config) Client {
	git := clients.NewClient(cfg)
	return git.Groups
}

// IsErrorGroupNotFound helper function to test for errGroupNotFound error.
func IsErrorGroupNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errGroupNotFound)
}

// VisibilityValueV1alpha1ToGitlab converts *v1alpha1.VisibilityValue to *gitlab.VisibilityValue
func VisibilityValueV1alpha1ToGitlab(from *v1alpha1.VisibilityValue) *gitlab.VisibilityValue {
	return (*gitlab.VisibilityValue)(from)
}

// ProjectCreationLevelValueV1alpha1ToGitlab converts *v1alpha1.ProjectCreationLevelValue to *gitlab.ProjectCreationLevelValue
func ProjectCreationLevelValueV1alpha1ToGitlab(from *v1alpha1.ProjectCreationLevelValue) *gitlab.ProjectCreationLevelValue {
	return (*gitlab.ProjectCreationLevelValue)(from)
}

// SubGroupCreationLevelValueV1alpha1ToGitlab converts *v1alpha1.SubGroupCreationLevelValue to *gitlab.SubGroupCreationLevelValue
func SubGroupCreationLevelValueV1alpha1ToGitlab(from *v1alpha1.SubGroupCreationLevelValue) *gitlab.SubGroupCreationLevelValue {
	return (*gitlab.SubGroupCreationLevelValue)(from)
}

// LateInitializeVisibilityValue returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeVisibilityValue(in *v1alpha1.VisibilityValue, from gitlab.VisibilityValue) *v1alpha1.VisibilityValue {
	if in == nil && from != "" {
		return (*v1alpha1.VisibilityValue)(&from)
	}
	return in
}

// LateInitializeSubGroupCreationLevelValue returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeSubGroupCreationLevelValue(in *v1alpha1.SubGroupCreationLevelValue, from gitlab.SubGroupCreationLevelValue) *v1alpha1.SubGroupCreationLevelValue {
	if in == nil && from != "" {
		return (*v1alpha1.SubGroupCreationLevelValue)(&from)
	}
	return in
}

// LateInitializeProjectCreationLevelValue returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeProjectCreationLevelValue(in *v1alpha1.ProjectCreationLevelValue, from gitlab.ProjectCreationLevelValue) *v1alpha1.ProjectCreationLevelValue {
	if in == nil && from != "" {
		return (*v1alpha1.ProjectCreationLevelValue)(&from)
	}
	return in
}

// LateInitialize fills the empty fields in the group spec with the
// values seen in gitlab.Group.
func LateInitialize(in *v1alpha1.GroupParameters, group *gitlab.Group) { // nolint:gocyclo
	if group == nil {
		return
	}

	in.Path = clients.LateInitializeStringPtr(in.Path, group.Path)
	in.Description = clients.LateInitializeStringPtr(in.Description, group.Description)
	in.Visibility = LateInitializeVisibilityValue(in.Visibility, group.Visibility)
	in.ProjectCreationLevel = LateInitializeProjectCreationLevelValue(in.ProjectCreationLevel, group.ProjectCreationLevel)
	in.SubGroupCreationLevel = LateInitializeSubGroupCreationLevelValue(in.SubGroupCreationLevel, group.SubGroupCreationLevel)

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

// GenerateObservation is used to produce v1alpha1.GroupGitLabObservation from
// gitlab.Group.
func GenerateObservation(grp *gitlab.Group) v1alpha1.GroupObservation { // nolint:gocyclo
	if grp == nil {
		return v1alpha1.GroupObservation{}
	}
	group := v1alpha1.GroupObservation{
		ID:           grp.ID,
		AvatarURL:    grp.AvatarURL,
		WebURL:       grp.WebURL,
		FullName:     grp.FullName,
		FullPath:     grp.FullPath,
		RunnersToken: grp.RunnersToken,
		LDAPCN:       grp.LDAPCN,
	}

	if grp.CreatedAt != nil {
		group.CreatedAt = &metav1.Time{Time: *grp.CreatedAt}
	}

	if grp.MarkedForDeletionOn != nil {
		group.MarkedForDeletionOn = &metav1.Time{Time: time.Time(*grp.MarkedForDeletionOn)}
	}

	if grp.Statistics != nil {
		group.Statistics = &v1alpha1.StorageStatistics{
			StorageSize:      grp.Statistics.StorageSize,
			RepositorySize:   grp.Statistics.RepositorySize,
			LfsObjectsSize:   grp.Statistics.LfsObjectsSize,
			JobArtifactsSize: grp.Statistics.JobArtifactsSize,
		}
	}

	if len(group.CustomAttributes) == 0 && len(grp.CustomAttributes) > 0 {
		group.CustomAttributes = make([]v1alpha1.CustomAttribute, len(grp.CustomAttributes))
		for i, c := range grp.CustomAttributes {
			group.CustomAttributes[i].Key = c.Key
			group.CustomAttributes[i].Value = c.Value
		}
	}

	if len(group.SharedWithGroups) == 0 && len(grp.SharedWithGroups) > 0 {
		group.SharedWithGroups = make([]v1alpha1.SharedWithGroups, len(grp.SharedWithGroups))
		for i, c := range grp.SharedWithGroups {
			group.SharedWithGroups[i].GroupID = c.GroupID
			group.SharedWithGroups[i].GroupName = c.GroupName
			group.SharedWithGroups[i].GroupFullPath = c.GroupFullPath
			group.SharedWithGroups[i].GroupAccessLevel = c.GroupAccessLevel
			group.SharedWithGroups[i].ExpiresAt = &metav1.Time{Time: time.Time(*c.ExpiresAt)}
		}
	}

	if len(group.LDAPGroupLinks) == 0 && len(grp.LDAPGroupLinks) > 0 {
		group.LDAPGroupLinks = make([]v1alpha1.LDAPGroupLink, len(grp.LDAPGroupLinks))
		for i, c := range grp.LDAPGroupLinks {
			group.LDAPGroupLinks[i].CN = c.CN
			group.LDAPGroupLinks[i].GroupAccess = v1alpha1.AccessLevelValue(c.GroupAccess)
			group.LDAPGroupLinks[i].Provider = c.Provider
		}
	}

	return group
}

// GenerateCreateGroupOptions generates group creation options
func GenerateCreateGroupOptions(name string, p *v1alpha1.GroupParameters) *gitlab.CreateGroupOptions {
	group := &gitlab.CreateGroupOptions{
		Name:                           &name,
		Path:                           p.Path,
		Description:                    p.Description,
		MembershipLock:                 p.MembershipLock,
		Visibility:                     VisibilityValueV1alpha1ToGitlab(p.Visibility),
		ShareWithGroupLock:             p.ShareWithGroupLock,
		RequireTwoFactorAuth:           p.RequireTwoFactorAuth,
		TwoFactorGracePeriod:           p.TwoFactorGracePeriod,
		ProjectCreationLevel:           ProjectCreationLevelValueV1alpha1ToGitlab(p.ProjectCreationLevel),
		AutoDevopsEnabled:              p.AutoDevopsEnabled,
		SubGroupCreationLevel:          SubGroupCreationLevelValueV1alpha1ToGitlab(p.SubGroupCreationLevel),
		EmailsDisabled:                 p.EmailsDisabled,
		MentionsDisabled:               p.MentionsDisabled,
		LFSEnabled:                     p.LFSEnabled,
		RequestAccessEnabled:           p.RequestAccessEnabled,
		ParentID:                       p.ParentID,
		SharedRunnersMinutesLimit:      p.SharedRunnersMinutesLimit,
		ExtraSharedRunnersMinutesLimit: p.ExtraSharedRunnersMinutesLimit,
	}

	return group
}

// GenerateEditGroupOptions generates group edit options
func GenerateEditGroupOptions(name string, p *v1alpha1.GroupParameters) *gitlab.UpdateGroupOptions {
	group := &gitlab.UpdateGroupOptions{
		Name:                           &name,
		Path:                           p.Path,
		Description:                    p.Description,
		MembershipLock:                 p.MembershipLock,
		Visibility:                     VisibilityValueV1alpha1ToGitlab(p.Visibility),
		ShareWithGroupLock:             p.ShareWithGroupLock,
		RequireTwoFactorAuth:           p.RequireTwoFactorAuth,
		TwoFactorGracePeriod:           p.TwoFactorGracePeriod,
		ProjectCreationLevel:           ProjectCreationLevelValueV1alpha1ToGitlab(p.ProjectCreationLevel),
		AutoDevopsEnabled:              p.AutoDevopsEnabled,
		SubGroupCreationLevel:          SubGroupCreationLevelValueV1alpha1ToGitlab(p.SubGroupCreationLevel),
		EmailsDisabled:                 p.EmailsDisabled,
		MentionsDisabled:               p.MentionsDisabled,
		LFSEnabled:                     p.LFSEnabled,
		RequestAccessEnabled:           p.RequestAccessEnabled,
		ParentID:                       p.ParentID,
		SharedRunnersMinutesLimit:      p.SharedRunnersMinutesLimit,
		ExtraSharedRunnersMinutesLimit: p.ExtraSharedRunnersMinutesLimit,
	}
	return group
}

// IsGroupUpToDate checks whether there is a change in any of the modifiable fields.
func IsGroupUpToDate(p *v1alpha1.GroupParameters, g *gitlab.Group) bool { // nolint:gocyclo
	if !cmp.Equal(p.Path, clients.StringToPtr(g.Path)) {
		return false
	}
	if !cmp.Equal(p.Description, clients.StringToPtr(g.Description)) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.MembershipLock, g.MembershipLock) {
		return false
	}
	if p.Visibility != nil {
		if !cmp.Equal(string(*p.Visibility), string(g.Visibility)) {
			return false
		}
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
