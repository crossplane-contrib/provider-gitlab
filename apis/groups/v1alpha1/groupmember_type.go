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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// GroupMemberSAMLIdentity represents the SAML Identity link for the group member.
//
// GitLab API docs: https://docs.gitlab.com/ce/api/members.html#list-all-members-of-a-group-or-project
// Gitlab MR for API change: https://gitlab.com/gitlab-org/gitlab/-/merge_requests/20357
// Gitlab MR for API Doc change: https://gitlab.com/gitlab-org/gitlab/-/merge_requests/25652
type GroupMemberSAMLIdentity struct {
	ExternUID      string `json:"externUID"`
	Provider       string `json:"provider"`
	SAMLProviderID int    `json:"samlProviderID"`
}

// A GroupMemberParameters defines the desired state of a Gitlab GroupMember.
type GroupMemberParameters struct {

	// The ID of the group owned by the authenticated user.
	// +optional
	// +immutable
	GroupID *int `json:"groupId,omitempty"`

	// GroupIDRef is a reference to a group to retrieve its groupId
	// +optional
	// +immutable
	GroupIDRef *xpv1.Reference `json:"groupIdRef,omitempty"`

	// GroupIDSelector selects reference to a group to retrieve its groupId.
	// +optional
	GroupIDSelector *xpv1.Selector `json:"groupIdSelector,omitempty"`

	// The user ID of the member.
	// +immutable
	UserID int `json:"userID"`

	// A valid access level.
	// +immutable
	AccessLevel AccessLevelValue `json:"accessLevel"`

	// A date string in the format YEAR-MONTH-DAY.
	// +optional
	ExpiresAt *string `json:"expiresAt,omitempty"`
}

// GroupMemberObservation represents a group member.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/groups.html#list-group-members
type GroupMemberObservation struct {
	Username          string                   `json:"username,omitempty"`
	Name              string                   `json:"name,omitempty"`
	State             string                   `json:"state,omitempty"`
	AvatarURL         string                   `json:"avatarURL,omitempty"`
	WebURL            string                   `json:"webURL,omitempty"`
	GroupSAMLIdentity *GroupMemberSAMLIdentity `json:"groupSamlIdentity,omitempty"`
}

// A GroupMemberSpec defines the desired state of a Gitlab GroupMember.
type GroupMemberSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       GroupMemberParameters `json:"forProvider"`
}

// A GroupMemberStatus represents the observed state of a Gitlab GroupMember.
type GroupMemberStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          GroupMemberObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A GroupMember is a managed resource that represents a Gitlab GroupMember
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Group ID",type="integer",JSONPath=".spec.forProvider.groupId"
// +kubebuilder:printcolumn:name="Username",type="string",JSONPath=".status.atProvider.username"
// +kubebuilder:printcolumn:name="Acceess Level",type="integer",JSONPath=".spec.forProvider.accessLevel"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gitlab}
type GroupMember struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GroupMemberSpec   `json:"spec"`
	Status GroupMemberStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GroupMemberList contains a list of GroupMember items
type GroupMemberList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GroupMember `json:"items"`
}
