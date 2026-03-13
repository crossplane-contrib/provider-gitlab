/*
Copyright 2021 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
*/

package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	// +cluster-scope:delete=1
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EnvironmentAccessLevelParameters configures who can deploy to the environment.
type EnvironmentAccessLevelParameters struct {
	// AccessLevel is a numeric GitLab access level (30=Developer, 40=Maintainer, 60=Admin).
	// +optional
	AccessLevel *int `json:"accessLevel,omitempty"`

	// UserID is a GitLab user ID allowed to deploy.
	// +optional
	UserID *int64 `json:"userId,omitempty"`

	// GroupID is a GitLab group ID allowed to deploy.
	// +optional
	GroupID *int64 `json:"groupId,omitempty"`

	// GroupInheritanceType controls how inherited group memberships are treated.
	// 0 => direct only, 1 => include inherited.
	// +optional
	GroupInheritanceType *int64 `json:"groupInheritanceType,omitempty"`
}

// EnvironmentApprovalRuleParameters defines who can approve and how many approvals are required.
type EnvironmentApprovalRuleParameters struct {
	// AccessLevel is a numeric GitLab access level (30=Developer, 40=Maintainer, 60=Admin).
	// +optional
	AccessLevel *int `json:"accessLevel,omitempty"`

	// UserID is a GitLab user ID allowed to approve.
	// +optional
	UserID *int64 `json:"userId,omitempty"`

	// GroupID is a GitLab group ID allowed to approve.
	// +optional
	GroupID *int64 `json:"groupId,omitempty"`

	// RequiredApprovals required for this rule.
	// +optional
	RequiredApprovals *int64 `json:"requiredApprovals,omitempty"`

	// GroupInheritanceType controls how inherited group memberships are treated.
	// 0 => direct only, 1 => include inherited.
	// +optional
	GroupInheritanceType *int64 `json:"groupInheritanceType,omitempty"`
}

// EnvironmentAccessLevelObservation reflects a deploy access level entry returned by GitLab.
type EnvironmentAccessLevelObservation struct {
	ID                     *int64  `json:"id,omitempty"`
	AccessLevel            *int    `json:"accessLevel,omitempty"`
	AccessLevelDescription *string `json:"accessLevelDescription,omitempty"`
	UserID                 *int64  `json:"userId,omitempty"`
	GroupID                *int64  `json:"groupId,omitempty"`
	GroupInheritanceType   *int64  `json:"groupInheritanceType,omitempty"`
}

// EnvironmentApprovalRuleObservation reflects an approval rule entry returned by GitLab.
type EnvironmentApprovalRuleObservation struct {
	ID                     *int64  `json:"id,omitempty"`
	UserID                 *int64  `json:"userId,omitempty"`
	GroupID                *int64  `json:"groupId,omitempty"`
	AccessLevel            *int    `json:"accessLevel,omitempty"`
	AccessLevelDescription *string `json:"accessLevelDescription,omitempty"`
	RequiredApprovals      *int64  `json:"requiredApprovals,omitempty"`
	GroupInheritanceType   *int64  `json:"groupInheritanceType,omitempty"`
}

// ProtectedEnvironmentParameters define the desired state of a GitLab Protected Environment.
type ProtectedEnvironmentParameters struct {
	// Name is the environment name (e.g. "production", "staging").
	// +kubebuilder:validation:Required
	Name *string `json:"name"`

	// ProjectID is the ID or path of the GitLab project.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1.Project
	// +optional
	ProjectID *string `json:"projectId,omitempty"`

	// ProjectIDRef is a reference to a Project to populate projectId.
	// +optional
	ProjectIDRef *xpv1.NamespacedReference `json:"projectIdRef,omitempty"`

	// ProjectIDSelector selects a reference to a Project to populate projectId.
	// +optional
	ProjectIDSelector *xpv1.NamespacedSelector `json:"projectIdSelector,omitempty"`

	// DeployAccessLevels configures who can deploy to this environment.
	// +optional
	DeployAccessLevels *[]EnvironmentAccessLevelParameters `json:"deployAccessLevels,omitempty"`

	// ApprovalRules configures who can approve and how many approvals are required.
	// +optional
	ApprovalRules *[]EnvironmentApprovalRuleParameters `json:"approvalRules,omitempty"`

	// RequiredApprovalCount is the unified (legacy) approval count.
	// +optional
	RequiredApprovalCount *int64 `json:"requiredApprovalCount,omitempty"`
}

// ProtectedEnvironmentObservation reflects the observed state from GitLab.
type ProtectedEnvironmentObservation struct {
	Name                  *string `json:"name,omitempty"`
	RequiredApprovalCount *int64  `json:"requiredApprovalCount,omitempty"`

	DeployAccessLevels []EnvironmentAccessLevelObservation  `json:"deployAccessLevels,omitempty"`
	ApprovalRules      []EnvironmentApprovalRuleObservation `json:"approvalRules,omitempty"`
}

type ProtectedEnvironmentSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`
	ForProvider              ProtectedEnvironmentParameters `json:"forProvider"`
}

type ProtectedEnvironmentStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ProtectedEnvironmentObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// ProtectedEnvironment is a managed resource that represents a GitLab Protected Environment.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="NAME",type="string",JSONPath=".spec.forProvider.name"
// +kubebuilder:printcolumn:name="PROJECT",type="string",JSONPath=".spec.forProvider.projectId"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}
type ProtectedEnvironment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProtectedEnvironmentSpec   `json:"spec"`
	Status ProtectedEnvironmentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProtectedEnvironmentList contains a list of ProtectedEnvironment items.
type ProtectedEnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProtectedEnvironment `json:"items"`
}
