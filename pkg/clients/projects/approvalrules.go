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
	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"gitlab.com/gitlab-org/api/client-go"
)

// MemberClient defines Gitlab Member service operations
type ApprovalRulesClient interface {
	GetProjectApprovalRule(pid any, ruleID int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectApprovalRule, *gitlab.Response, error)
	CreateProjectApprovalRule(pid any, opt *gitlab.CreateProjectLevelRuleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectApprovalRule, *gitlab.Response, error)
	UpdateProjectApprovalRule(pid any, approvalRule int, opt *gitlab.UpdateProjectLevelRuleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectApprovalRule, *gitlab.Response, error)
	DeleteProjectApprovalRule(pid any, approvalRule int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewMemberClient returns a new Gitlab Project Member service
func NewApprovalRulesClient(cfg clients.Config) ApprovalRulesClient {
	git := clients.NewClient(cfg)
	return git.Projects
}

// GenerateMemberObservation is used to produce v1alpha1.MemberObservation from
// gitlab.Member.
// func GenerateApprovalRulesObservation(projectMember *gitlab.ProjectMember) v1alpha1.ApprovalRuleObservation {
// 	if projectMember == nil {
// 		return v1alpha1.ApprovalRuleObservation{}
// 	}
//
// 	o := v1alpha1.ApprovalRuleObservation{
// 		Username:  projectMember.Username,
// 		Email:     projectMember.Email,
// 		Name:      projectMember.Name,
// 		State:     projectMember.State,
// 		AvatarURL: projectMember.AvatarURL,
// 		WebURL:    projectMember.WebURL,
// 	}
//
// 	if o.CreatedAt == nil && projectMember.CreatedAt != nil {
// 		o.CreatedAt = &metav1.Time{Time: *projectMember.CreatedAt}
// 	}
//
// 	return o
// }

// GenerateAddMemberOptions generates project member add options
func GenerateCreateApprovalRulesOptions(p *v1alpha1.ApprovalRuleParameters) *gitlab.CreateProjectLevelRuleOptions {
	approvalRulesOptions := &gitlab.CreateProjectLevelRuleOptions{
		Name:                          p.Name,
		ApprovalsRequired:             p.ApprovalsRequired,
		RuleType:                      (*string)(p.RuleType),
		AppliesToAllProtectedBranches: p.AppliesToAllProtectedBranches,
		UserIDs:                       p.UserIDs,
		GroupIDs:                      p.GroupIDs,
		ProtectedBranchIDs:            p.ProtectedBranchIDs,
		Usernames:                     p.Usernames,
	}

	return approvalRulesOptions
}

// GenerateEditMemberOptions generates project member edit options
func GenerateUpdateApprovalRulesOptions(p *v1alpha1.ApprovalRuleParameters) *gitlab.UpdateProjectLevelRuleOptions {
	approvalRulesOptions := &gitlab.UpdateProjectLevelRuleOptions{
		Name:                          p.Name,
		ApprovalsRequired:             p.ApprovalsRequired,
		AppliesToAllProtectedBranches: p.AppliesToAllProtectedBranches,
		UserIDs:                       p.UserIDs,
		GroupIDs:                      p.GroupIDs,
		ProtectedBranchIDs:            p.ProtectedBranchIDs,
		Usernames:                     p.Usernames,
	}

	return approvalRulesOptions
}
