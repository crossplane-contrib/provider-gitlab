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

	"github.com/xanzy/go-gitlab"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
)

const (
	errSubGroupNotFound = "404 SubGroup Not Found"
)

// SubGroupClient defines Gitlab Group service operations
type SubGroupClient interface {
	GetGroup(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error)
	CreateGroup(opt *gitlab.CreateGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error)
	UpdateGroup(gid interface{}, opt *gitlab.UpdateGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error)
	DeleteGroup(gid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewSubGroupClient returns a new Gitlab Group service
func NewSubGroupClient(cfg clients.Config) Client {
	git := clients.NewClient(cfg)
	return git.Groups
}

// IsErrorSubGroupNotFound helper function to test for errSubGroupNotFound error.
func IsErrorSubGroupNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errSubGroupNotFound)
}

// GenerateSubGroupObservation is used to produce v1alpha1.SubGroupObservation from
// gitlab.Group.
func GenerateSubGroupObservation(grp *gitlab.Group) v1alpha1.SubGroupObservation { // nolint:gocyclo
	if grp == nil {
		return v1alpha1.SubGroupObservation{}
	}
	group := v1alpha1.SubGroupObservation{
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

// GenerateCreateSubGroupOptions generates subgroup creation options
func GenerateCreateSubGroupOptions(name string, p *v1alpha1.SubGroupParameters) *gitlab.CreateGroupOptions {
	group := &gitlab.CreateGroupOptions{
		Name:                           &name,
		Path:                           &p.Path,
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

// GenerateEditSubGroupOptions generates subgroup edit options
func GenerateEditSubGroupOptions(name string, p *v1alpha1.SubGroupParameters) *gitlab.UpdateGroupOptions {
	group := &gitlab.UpdateGroupOptions{
		Name:                           &name,
		Path:                           &p.Path,
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
