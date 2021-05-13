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

	"github.com/xanzy/go-gitlab"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
)

const (
	errGroupMemberNotFound = "404 Group Member Not Found"
)

// GroupMemberClient defines Gitlab GroupMember service operations
type GroupMemberClient interface {
	GetGroupMember(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error)
	AddGroupMember(gid interface{}, opt *gitlab.AddGroupMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error)
	EditGroupMember(gid interface{}, user int, opt *gitlab.EditGroupMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error)
	RemoveGroupMember(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewGroupMemberClient returns a new Gitlab Group Member service
func NewGroupMemberClient(cfg clients.Config) GroupMemberClient {
	git := clients.NewClient(cfg)
	return git.GroupMembers
}

// IsErrorGroupMemberNotFound helper function to test for errGroupMemberNotFound error.
func IsErrorGroupMemberNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errGroupMemberNotFound)
}

// GenerateGroupMemberObservation is used to produce v1alpha1.GroupMemberObservation from
// gitlab.GroupMember.
func GenerateGroupMemberObservation(groupMember *gitlab.GroupMember) v1alpha1.GroupMemberObservation { // nolint:gocyclo
	if groupMember == nil {
		return v1alpha1.GroupMemberObservation{}
	}

	o := v1alpha1.GroupMemberObservation{
		Username:          groupMember.Username,
		Name:              groupMember.Name,
		State:             groupMember.State,
		AvatarURL:         groupMember.AvatarURL,
		WebURL:            groupMember.WebURL,
		GroupSAMLIdentity: groupMemberSAMLIdentityGitlabToV1alpha1(groupMember.GroupSAMLIdentity),
	}

	return o
}

// GenerateAddGroupMemberOptions generates group member add options
func GenerateAddGroupMemberOptions(p *v1alpha1.GroupMemberParameters) *gitlab.AddGroupMemberOptions {
	groupMember := &gitlab.AddGroupMemberOptions{
		UserID:      &p.UserID,
		AccessLevel: accessLevelValueV1alpha1ToGitlab(&p.AccessLevel),
	}
	if p.ExpiresAt != nil {
		groupMember.ExpiresAt = p.ExpiresAt
	}
	return groupMember
}

// GenerateEditGroupMemberOptions generates group member edit options
func GenerateEditGroupMemberOptions(p *v1alpha1.GroupMemberParameters) *gitlab.EditGroupMemberOptions {
	groupMember := &gitlab.EditGroupMemberOptions{
		AccessLevel: accessLevelValueV1alpha1ToGitlab(&p.AccessLevel),
	}
	if p.ExpiresAt != nil {
		groupMember.ExpiresAt = p.ExpiresAt
	}
	return groupMember
}

// accessLevelValueV1alpha1ToGitlab converts *v1alpha1.AccessLevelValue to *gitlab.AccessLevelValue
func accessLevelValueV1alpha1ToGitlab(from *v1alpha1.AccessLevelValue) *gitlab.AccessLevelValue {
	return (*gitlab.AccessLevelValue)(from)
}

// groupMemberSAMLIdentityGitlabToV1alpha1 converts *gitlab.GroupMemberSAMLIdentity to *v1alpha1.GroupMemberSAMLIdentity
func groupMemberSAMLIdentityGitlabToV1alpha1(from *gitlab.GroupMemberSAMLIdentity) *v1alpha1.GroupMemberSAMLIdentity {
	return (*v1alpha1.GroupMemberSAMLIdentity)(from)
}
