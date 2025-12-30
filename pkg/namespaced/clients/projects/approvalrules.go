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
	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
)

// ApprovalRulesClient Gitlab Member service operations
type ApprovalRulesClient interface {
	GetProjectApprovalRule(pid any, ruleID int64, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectApprovalRule, *gitlab.Response, error)
	CreateProjectApprovalRule(pid any, opt *gitlab.CreateProjectLevelRuleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectApprovalRule, *gitlab.Response, error)
	UpdateProjectApprovalRule(pid any, approvalRule int64, opt *gitlab.UpdateProjectLevelRuleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectApprovalRule, *gitlab.Response, error)
	DeleteProjectApprovalRule(pid any, approvalRule int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewApprovalRulesClient returns a new Gitlab Project Member service
func NewApprovalRulesClient(cfg common.Config) ApprovalRulesClient {
	git := common.NewClient(cfg)
	return git.Projects
}

// GenerateCreateApprovalRulesOptions generates project member add options
func GenerateCreateApprovalRulesOptions(p *v1alpha1.ApprovalRuleParameters) *gitlab.CreateProjectLevelRuleOptions {
	approvalRulesOptions := &gitlab.CreateProjectLevelRuleOptions{
		Name:                          p.Name,
		RuleType:                      (*string)(p.RuleType),
		AppliesToAllProtectedBranches: p.AppliesToAllProtectedBranches,
		Usernames:                     p.Usernames,
	}
	if p.ApprovalsRequired != nil {
		val := p.ApprovalsRequired
		approvalRulesOptions.ApprovalsRequired = val
	}
	if p.UserIDs != nil {
		vals := make([]int64, len(*p.UserIDs))
		copy(vals, *p.UserIDs)
		approvalRulesOptions.UserIDs = &vals
	}
	if p.GroupIDs != nil {
		vals := make([]int64, len(*p.GroupIDs))
		copy(vals, *p.GroupIDs)
		approvalRulesOptions.GroupIDs = &vals
	}
	if p.ProtectedBranchIDs != nil {
		vals := make([]int64, len(*p.ProtectedBranchIDs))
		copy(vals, *p.ProtectedBranchIDs)
		approvalRulesOptions.ProtectedBranchIDs = &vals
	}

	return approvalRulesOptions
}

// GenerateUpdateApprovalRulesOptions generates project member edit options
func GenerateUpdateApprovalRulesOptions(p *v1alpha1.ApprovalRuleParameters) *gitlab.UpdateProjectLevelRuleOptions {
	approvalRulesOptions := &gitlab.UpdateProjectLevelRuleOptions{
		Name:                          p.Name,
		AppliesToAllProtectedBranches: p.AppliesToAllProtectedBranches,
		Usernames:                     p.Usernames,
	}
	if p.ApprovalsRequired != nil {
		val := p.ApprovalsRequired
		approvalRulesOptions.ApprovalsRequired = val
	}
	if p.UserIDs != nil {
		vals := make([]int64, len(*p.UserIDs))
		copy(vals, *p.UserIDs)
		approvalRulesOptions.UserIDs = &vals
	}
	if p.GroupIDs != nil {
		vals := make([]int64, len(*p.GroupIDs))
		copy(vals, *p.GroupIDs)
		approvalRulesOptions.GroupIDs = &vals
	}
	if p.ProtectedBranchIDs != nil {
		vals := make([]int64, len(*p.ProtectedBranchIDs))
		copy(vals, *p.ProtectedBranchIDs)
		approvalRulesOptions.ProtectedBranchIDs = &vals
	}

	return approvalRulesOptions
}

// IsApprovalRuleUpToDate checks whether there is a change in any of the modifiable fields.
func IsApprovalRuleUpToDate(p *v1alpha1.ApprovalRuleParameters, g *gitlab.ProjectApprovalRule) bool {
	if !cmp.Equal(p.Name, clients.StringToPtr(g.Name)) {
		return false
	}

	if !clients.IsBoolEqualToBoolPtr(p.AppliesToAllProtectedBranches, g.AppliesToAllProtectedBranches) {
		return false
	}

	if !clients.IsInt64EqualToInt64Ptr(p.ApprovalsRequired, g.ApprovalsRequired) {
		return false
	}

	if !clients.IsStringEqualToStringPtr((*string)(p.RuleType), g.RuleType) {
		return false
	}

	if !isGroupIDsUpToDate(p, g) {
		return false
	}

	if !isProtectedBranchesIDsUpToDate(p, g) {
		return false
	}

	if !isUserIDsUpToDate(p, g) {
		return false
	}

	if !isUsernamesUpToDate(p, g) {
		return false
	}

	return true
}

func isGroupIDsUpToDate(cr *v1alpha1.ApprovalRuleParameters, in *gitlab.ProjectApprovalRule) bool {
	if cr.GroupIDs == nil {
		return len(in.Groups) == 0
	}

	if len(*cr.GroupIDs) != len(in.Groups) {
		return false
	}

	inIDs := make(map[int64]any)
	for _, v := range in.Groups {
		inIDs[v.ID] = nil
	}

	crIDs := make(map[int64]any)
	for _, v := range *cr.GroupIDs {
		crIDs[v] = nil
	}

	for ID := range inIDs {
		_, ok := crIDs[ID]
		if !ok {
			return false
		}
	}

	for ID := range crIDs {
		_, ok := inIDs[ID]
		if !ok {
			return false
		}
	}

	return true
}

func isProtectedBranchesIDsUpToDate(cr *v1alpha1.ApprovalRuleParameters, in *gitlab.ProjectApprovalRule) bool {
	if cr.ProtectedBranchIDs == nil {
		return len(in.ProtectedBranches) == 0
	}

	if len(*cr.ProtectedBranchIDs) != len(in.ProtectedBranches) {
		return false
	}

	inIDs := make(map[int64]any)
	for _, v := range in.ProtectedBranches {
		inIDs[v.ID] = nil
	}

	crIDs := make(map[int64]any)
	for _, v := range *cr.ProtectedBranchIDs {
		crIDs[v] = nil
	}

	for ID := range inIDs {
		_, ok := crIDs[ID]
		if !ok {
			return false
		}
	}

	for ID := range crIDs {
		_, ok := inIDs[ID]
		if !ok {
			return false
		}
	}

	return true
}

func isUserIDsUpToDate(cr *v1alpha1.ApprovalRuleParameters, in *gitlab.ProjectApprovalRule) bool {
	if cr.UserIDs == nil {
		return len(in.Users) == 0
	}

	if len(*cr.UserIDs) != len(in.Users) {
		return false
	}

	inIDs := make(map[int64]any)
	for _, v := range in.Users {
		inIDs[v.ID] = nil
	}

	crIDs := make(map[int64]any)
	for _, v := range *cr.UserIDs {
		crIDs[v] = nil
	}

	for ID := range inIDs {
		_, ok := crIDs[ID]
		if !ok {
			return false
		}
	}

	for ID := range crIDs {
		_, ok := inIDs[ID]
		if !ok {
			return false
		}
	}

	return true
}

func isUsernamesUpToDate(cr *v1alpha1.ApprovalRuleParameters, in *gitlab.ProjectApprovalRule) bool {
	if cr.Usernames == nil {
		return len(in.Users) == 0
	}

	if len(*cr.Usernames) != len(in.Users) {
		return false
	}

	inIDs := make(map[string]any)
	for _, v := range in.Users {
		inIDs[v.Username] = nil
	}

	crIDs := make(map[string]any)
	for _, v := range *cr.Usernames {
		crIDs[v] = nil
	}

	for ID := range inIDs {
		_, ok := crIDs[ID]
		if !ok {
			return false
		}
	}

	for ID := range crIDs {
		_, ok := inIDs[ID]
		if !ok {
			return false
		}
	}

	return true
}
