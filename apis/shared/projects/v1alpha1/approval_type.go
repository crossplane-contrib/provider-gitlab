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

package v1alpha1

type RuleType string

const (
	RuleTypeAnyApprover    RuleType = "any_approver"
	RuleTypeRegular        RuleType = "regular"
	RuleTypeReportApprover RuleType = "report_approver"
)

// A ApprovalRuleParameters defines the desired state of a Gitlab Project Member.
type ApprovalRuleParameters struct {

	// The ID of the project owned by the authenticated user.
	// +optional
	// +immutable
	ProjectID *int `json:"projectId,omitempty"`

	// The number of required approvals for this rule.
	// +kubebuilder:example=2
	ApprovalsRequired *int `json:"approvalsRequired,omitempty"`

	// The name of the approval rule
	// +kubebuilder:example="Security Team Review"
	Name *string `json:"name,omitempty"`

	// If true, applies the rule to all protected branches and ignores the protected_branch_ids attribute.
	// +optional
	AppliesToAllProtectedBranches *bool `json:"appliesToAllProtectedBranches,omitempty"`

	// The IDs of groups as approvers.
	// +optional
	GroupIDs *[]int `json:"groupIds,omitempty"`

	// The IDs of protected branches to scope the rule by.
	// +optional
	ProtectedBranchIDs *[]int `json:"protectedBranchIds,omitempty"`

	// The rule type. Supported values include any_approver, regular, and report_approver
	// +kubebuilder:example="any_approver"
	// +optional
	// +immutable
	RuleType *RuleType `json:"ruleType,omitempty"`

	// The IDs of users as approvers. If used with usernames, adds both lists of users.
	// +optional
	UserIDs *[]int `json:"userIds,omitempty"`

	// The usernames of approvers. If used with user_ids, adds both lists of users.
	// +kubebuilder:example={john.doe,jane.smith}
	// +optional
	Usernames *[]string `json:"usernames,omitempty"`
}

// ApprovalRuleObservation represents a project member.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/projects.html#list-project-team-members
type ApprovalRuleObservation struct{}
