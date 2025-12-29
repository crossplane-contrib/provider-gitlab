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
	"strings"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
)

const (
	errGroupNotFound = "404 Group Not Found"
)

// Client defines Gitlab Group service operations
type Client interface {
	GetGroup(gid interface{}, opt *gitlab.GetGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error)
	CreateGroup(opt *gitlab.CreateGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error)
	UpdateGroup(gid interface{}, opt *gitlab.UpdateGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error)
	DeleteGroup(gid interface{}, opt *gitlab.DeleteGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
	ShareGroupWithGroup(gid interface{}, opt *gitlab.ShareGroupWithGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error)
	UnshareGroupFromGroup(gid interface{}, groupID int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewGroupClient returns a new Gitlab Group service
func NewGroupClient(cfg common.Config) Client {
	git := common.NewClient(cfg)
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

// GenerateObservation is used to produce v1alpha1.GroupGitLabObservation from
// gitlab.Group.
func GenerateObservation(grp *gitlab.Group) v1alpha1.GroupObservation { //nolint:gocyclo
	if grp == nil {
		return v1alpha1.GroupObservation{}
	}
	group := v1alpha1.GroupObservation{
		ID:        func() *int { i := int(grp.ID); return &i }(),
		AvatarURL: &grp.AvatarURL,
		WebURL:    &grp.WebURL,
		FullName:  &grp.FullName,
		FullPath:  &grp.FullPath,
		LDAPCN:    &grp.LDAPCN,
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
			LfsObjectsSize:   grp.Statistics.LFSObjectsSize,
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

	if len(grp.LDAPGroupLinks) > 0 {
		group.LDAPGroupLinks = make([]v1alpha1.LDAPGroupLink, len(grp.LDAPGroupLinks))
		for i, c := range grp.LDAPGroupLinks {
			group.LDAPGroupLinks[i].CN = c.CN
			group.LDAPGroupLinks[i].GroupAccess = v1alpha1.AccessLevelValue(c.GroupAccess)
			group.LDAPGroupLinks[i].Provider = c.Provider
		}
	}

	if len(grp.SharedWithGroups) > 0 {
		arr := make([]v1alpha1.SharedWithGroupsObservation, 0)
		for _, v := range grp.SharedWithGroups {
			groupID := int(v.GroupID)
			groupAccessLevel := int(v.GroupAccessLevel)
			sg := v1alpha1.SharedWithGroupsObservation{
				GroupID:          &groupID,
				GroupName:        &v.GroupName,
				GroupFullPath:    &v.GroupFullPath,
				GroupAccessLevel: &groupAccessLevel,
			}
			if v.ExpiresAt != nil {
				sg.ExpiresAt = &metav1.Time{Time: time.Time(*v.ExpiresAt)}
			}
			arr = append(arr, sg)
		}
		group.SharedWithGroups = arr
	}
	return group
}

// GenerateCreateGroupOptions generates group creation options
func GenerateCreateGroupOptions(name string, p *v1alpha1.GroupParameters) *gitlab.CreateGroupOptions {
	// Name field overrides resource name
	if p.Name != nil {
		name = *p.Name
	}

	group := &gitlab.CreateGroupOptions{
		Name:                  &name,
		Path:                  &p.Path,
		Description:           p.Description,
		MembershipLock:        p.MembershipLock,
		Visibility:            VisibilityValueV1alpha1ToGitlab(p.Visibility),
		ShareWithGroupLock:    p.ShareWithGroupLock,
		RequireTwoFactorAuth:  p.RequireTwoFactorAuth,
		ProjectCreationLevel:  ProjectCreationLevelValueV1alpha1ToGitlab(p.ProjectCreationLevel),
		AutoDevopsEnabled:     p.AutoDevopsEnabled,
		SubGroupCreationLevel: SubGroupCreationLevelValueV1alpha1ToGitlab(p.SubGroupCreationLevel),
		MentionsDisabled:      p.MentionsDisabled,
		EmailsEnabled:         p.EmailsEnabled,
		LFSEnabled:            p.LFSEnabled,
		RequestAccessEnabled:  p.RequestAccessEnabled,
	}
	if p.TwoFactorGracePeriod != nil {
		val := int64(*p.TwoFactorGracePeriod)
		group.TwoFactorGracePeriod = &val
	}
	if p.ParentID != nil {
		val := int64(*p.ParentID)
		group.ParentID = &val
	}
	if p.SharedRunnersMinutesLimit != nil {
		val := int64(*p.SharedRunnersMinutesLimit)
		group.SharedRunnersMinutesLimit = &val
	}
	if p.ExtraSharedRunnersMinutesLimit != nil {
		val := int64(*p.ExtraSharedRunnersMinutesLimit)
		group.ExtraSharedRunnersMinutesLimit = &val
	}

	return group
}

// GenerateEditGroupOptions generates group edit options
func GenerateEditGroupOptions(name string, p *v1alpha1.GroupParameters) *gitlab.UpdateGroupOptions {
	// Name field overrides resource name
	if p.Name != nil {
		name = *p.Name
	}

	group := &gitlab.UpdateGroupOptions{
		Name:                  &name,
		Path:                  &p.Path,
		Description:           p.Description,
		MembershipLock:        p.MembershipLock,
		Visibility:            VisibilityValueV1alpha1ToGitlab(p.Visibility),
		ShareWithGroupLock:    p.ShareWithGroupLock,
		RequireTwoFactorAuth:  p.RequireTwoFactorAuth,
		ProjectCreationLevel:  ProjectCreationLevelValueV1alpha1ToGitlab(p.ProjectCreationLevel),
		AutoDevopsEnabled:     p.AutoDevopsEnabled,
		SubGroupCreationLevel: SubGroupCreationLevelValueV1alpha1ToGitlab(p.SubGroupCreationLevel),
		EmailsEnabled:         p.EmailsEnabled,
		MentionsDisabled:      p.MentionsDisabled,
		LFSEnabled:            p.LFSEnabled,
		RequestAccessEnabled:  p.RequestAccessEnabled,
	}
	if p.TwoFactorGracePeriod != nil {
		val := int64(*p.TwoFactorGracePeriod)
		group.TwoFactorGracePeriod = &val
	}
	if p.SharedRunnersMinutesLimit != nil {
		val := int64(*p.SharedRunnersMinutesLimit)
		group.SharedRunnersMinutesLimit = &val
	}
	if p.ExtraSharedRunnersMinutesLimit != nil {
		val := int64(*p.ExtraSharedRunnersMinutesLimit)
		group.ExtraSharedRunnersMinutesLimit = &val
	}
	return group
}
