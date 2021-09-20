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

package projects

import (
	"strings"

	"github.com/xanzy/go-gitlab"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
)

const (
	errMemberNotFound = "404 Project Member Not Found"
)

// MemberClient defines Gitlab Member service operations
type MemberClient interface {
	GetProjectMember(pid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error)
	AddProjectMember(pid interface{}, opt *gitlab.AddProjectMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error)
	EditProjectMember(pid interface{}, user int, opt *gitlab.EditProjectMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error)
	DeleteProjectMember(pid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewMemberClient returns a new Gitlab Project Member service
func NewMemberClient(cfg clients.Config) MemberClient {
	git := clients.NewClient(cfg)
	return git.ProjectMembers
}

// IsErrorMemberNotFound helper function to test for errMemberNotFound error.
func IsErrorMemberNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errMemberNotFound)
}

// GenerateMemberObservation is used to produce v1alpha1.MemberObservation from
// gitlab.Member.
func GenerateMemberObservation(projectMember *gitlab.ProjectMember) v1alpha1.MemberObservation { // nolint:gocyclo
	if projectMember == nil {
		return v1alpha1.MemberObservation{}
	}

	o := v1alpha1.MemberObservation{
		Username:  projectMember.Username,
		Email:     projectMember.Email,
		Name:      projectMember.Name,
		State:     projectMember.State,
		AvatarURL: projectMember.AvatarURL,
		WebURL:    projectMember.WebURL,
	}

	if o.CreatedAt == nil && projectMember.CreatedAt != nil {
		o.CreatedAt = &metav1.Time{Time: *projectMember.CreatedAt}
	}

	return o
}

// GenerateAddMemberOptions generates project member add options
func GenerateAddMemberOptions(p *v1alpha1.MemberParameters) *gitlab.AddProjectMemberOptions {
	projectMember := &gitlab.AddProjectMemberOptions{
		UserID:      &p.UserID,
		AccessLevel: accessLevelValueV1alpha1ToGitlab(&p.AccessLevel),
	}
	if p.ExpiresAt != nil {
		projectMember.ExpiresAt = p.ExpiresAt
	}
	return projectMember
}

// GenerateEditMemberOptions generates project member edit options
func GenerateEditMemberOptions(p *v1alpha1.MemberParameters) *gitlab.EditProjectMemberOptions {
	projectMember := &gitlab.EditProjectMemberOptions{
		AccessLevel: accessLevelValueV1alpha1ToGitlab(&p.AccessLevel),
	}
	if p.ExpiresAt != nil {
		projectMember.ExpiresAt = p.ExpiresAt
	}
	return projectMember
}

// accessLevelValueV1alpha1ToGitlab converts *v1alpha1.AccessLevelValue to *gitlab.AccessLevelValue
func accessLevelValueV1alpha1ToGitlab(from *v1alpha1.AccessLevelValue) *gitlab.AccessLevelValue {
	return (*gitlab.AccessLevelValue)(from)
}
