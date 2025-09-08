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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	sharedGroupsV1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/shared/groups/v1alpha1"
)

// GroupParameters define the desired state of a Gitlab Project
type GroupParameters struct {
	sharedGroupsV1alpha1.GroupParameters `json:",inline"`

	// ParentIDRef is a reference to a group to retrieve its parentId
	// +optional
	// +immutable
	ParentIDRef *xpv1.Reference `json:"parentIdRef,omitempty"`

	// ParentIDSelector selects reference to a group to retrieve its parentId.
	// +optional
	ParentIDSelector *xpv1.Selector `json:"parentIdSelector,omitempty"`

	// SharedWithGroups create links for sharing a group with another group.
	// +optional
	SharedWithGroups []SharedWithGroups `json:"sharedWithGroups,omitempty"`
}

// SharedWithGroups represents a GitLab Shared with groups.
// At least one of the fields [GroupID, GroupIDRef, GroupIDSelector] must be set.
type SharedWithGroups struct {
	sharedGroupsV1alpha1.SharedWithGroups `json:",inline"`

	// GroupIDRef is a reference to a group to retrieve its ID.
	GroupIDRef *xpv1.Reference `json:"groupIdRef,omitempty"`

	// GroupIDSelector selects reference to a group to retrieve its ID.
	GroupIDSelector *xpv1.Selector `json:"groupIdSelector,omitempty"`
}

// A GroupSpec defines the desired state of a Gitlab Group.
type GroupSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       GroupParameters `json:"forProvider"`
}

// A GroupStatus represents the observed state of a Gitlab Group.
type GroupStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          sharedGroupsV1alpha1.GroupObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Group is a managed resource that represents a Gitlab Group
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".status.atProvider.ID"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gitlab}
type Group struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GroupSpec   `json:"spec"`
	Status GroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GroupList contains a list of Group items
type GroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Group `json:"items"`
}
