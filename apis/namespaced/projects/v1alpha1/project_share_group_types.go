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
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	// +cluster-scope:delete=1
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProjectShareGroupParameters are the fields the user sets in YAML.
type ProjectShareGroupParameters struct {
	// AccessLevel is the permission level (e.g., 30 for Developer, 40 for Maintainer)
	// +kubebuilder:validation:Enum=10;20;30;40;50
	AccessLevel int `json:"accessLevel"`

	// ExpiresAt is optional. Format: "2024-09-26"
	// +optional
	// +immutable
	ExpiresAt *string `json:"expiresAt,omitempty"`

	// --- REFS CONFIGURATION ---
	// The fields below allow the user to say "Use the ID from that other Project/Group"

	// GroupID is the ID of the group to share with.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1.Group
	// +crossplane:generate:reference:refFieldName=GroupRef
	// +crossplane:generate:reference:selectorFieldName=GroupSelector
	// +optional
	// +immutable
	GroupID *string `json:"groupId,omitempty"`

	// GroupRef allows referencing a Group resource by name.
	// +optional
	// +immutable
	GroupRef *xpv1.NamespacedReference `json:"groupRef,omitempty"`

	// GroupSelector allows selecting a Group by labels.
	// +optional
	GroupSelector *xpv1.NamespacedSelector `json:"groupSelector,omitempty"`

	// ProjectID is the ID of the project being shared.
	// +crossplane:generate:reference:type=Project
	// +crossplane:generate:reference:refFieldName=ProjectRef
	// +crossplane:generate:reference:selectorFieldName=ProjectSelector
	// +optional
	// +immutable
	ProjectID *string `json:"projectId,omitempty"`

	// ProjectRef allows referencing a Project resource by name.
	// +optional
	// +immutable
	ProjectRef *xpv1.NamespacedReference `json:"projectRef,omitempty"`

	// ProjectSelector allows selecting a Project by labels.
	// +optional
	ProjectSelector *xpv1.NamespacedSelector `json:"projectSelector,omitempty"`
}

// ProjectShareGroupObservation are the fields we read back from GitLab (if any).
type ProjectShareGroupObservation struct {
	// We might not need much here, but usually we store the ID.
	ID string `json:"id,omitempty"`
}

// ProjectShareGroupSpec defines the desired state.
type ProjectShareGroupSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`
	ForProvider              ProjectShareGroupParameters `json:"forProvider"`
}

// ProjectShareGroupStatus defines the observed state.
type ProjectShareGroupStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ProjectShareGroupObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}

// ProjectShareGroup is the Schema for the ProjectShareGroups API
type ProjectShareGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectShareGroupSpec   `json:"spec,omitempty"`
	Status ProjectShareGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProjectShareGroupList contains a list of ProjectShareGroup
type ProjectShareGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProjectShareGroup `json:"items"`
}
