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

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

	// ProjectIDRef is a reference to a project to retrieve its projectId
	// +optional
	// +immutable
	ProjectIDRef *xpv1.Reference `json:"projectIdRef,omitempty"`

	// ProjectIDSelector selects reference to a project to retrieve its projectId.
	// +optional
	ProjectIDSelector *xpv1.Selector `json:"projectIdSelector,omitempty"`

	// The number of required approvals for this rule.
	ApprovalsRequired *int `json:"approvalsRequired,omitempty"`

	// The name of the approval rule
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
	// +optional
	// +immutable
	RuleType *RuleType `json:"ruleType,omitempty"`

	// The IDs of users as approvers. If used with usernames, adds both lists of users.
	// +optional
	UserIDs *[]int `json:"userIds,omitempty"`

	// The IDs of users as approvers. If used with usernames, adds both lists of users.
	// +optional
	Usernames *[]string `json:"usernames,omitempty"`
}

// ApprovalRuleObservation represents a project member.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/projects.html#list-project-team-members
type ApprovalRuleObservation struct {
	// Username  string       `json:"username,omitempty"`
	// Email     string       `json:"email,omitempty"`
	// Name      string       `json:"name,omitempty"`
	// State     string       `json:"state,omitempty"`
	// CreatedAt *metav1.Time `json:"createdAt,omitempty"`
	// WebURL    string       `json:"webURL,omitempty"`
	// AvatarURL string       `json:"avatarURL,omitempty"`
}

// A ApprovalRuleSpec defines the desired state of a Gitlab Project Member.
type ApprovalRuleSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ApprovalRuleParameters `json:"forProvider"`
}

// A ApprovalRuleStatus represents the observed state of a Gitlab Project Member.
type ApprovalRuleStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ApprovalRuleObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A ApprovalRule is a managed resource that represents a Gitlab Project ApprovalRule
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Project ID",type="integer",JSONPath=".spec.forProvider.projectId"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gitlab}
type ApprovalRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApprovalRuleSpec   `json:"spec"`
	Status ApprovalRuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApprovalRuleList contains a list of Approval Rules items
type ApprovalRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApprovalRule `json:"items"`
}
