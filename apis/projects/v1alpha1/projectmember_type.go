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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// A ProjectMemberParameters defines the desired state of a Gitlab ProjectMember.
type ProjectMemberParameters struct {

	// The ID of the project owned by the authenticated user.
	// +immutable
	ProjectID int `json:"projectID"`

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

// ProjectMemberObservation represents a project member.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/projects.html#list-project-team-members
type ProjectMemberObservation struct {
	Username  string       `json:"username,omitempty"`
	Email     string       `json:"email,omitempty"`
	Name      string       `json:"name,omitempty"`
	State     string       `json:"state,omitempty"`
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`
	WebURL    string       `json:"webURL,omitempty"`
	AvatarURL string       `json:"avatarURL,omitempty"`
}

// A ProjectMemberSpec defines the desired state of a Gitlab ProjectMember.
type ProjectMemberSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ProjectMemberParameters `json:"forProvider"`
}

// A ProjectMemberStatus represents the observed state of a Gitlab ProjectMember.
type ProjectMemberStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ProjectMemberObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A ProjectMember is a managed resource that represents a Gitlab ProjectMember
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Project ID",type="integer",JSONPath=".spec.forProvider.projectID"
// +kubebuilder:printcolumn:name="Username",type="string",JSONPath=".status.atProvider.username"
// +kubebuilder:printcolumn:name="Acceess Level",type="integer",JSONPath=".spec.forProvider.accessLevel"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gitlab}
type ProjectMember struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectMemberSpec   `json:"spec"`
	Status ProjectMemberStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProjectMemberList contains a list of ProjectMember items
type ProjectMemberList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProjectMember `json:"items"`
}
