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
)

// A MemberParameters defines the desired state of a Gitlab Project Member.
type MemberParameters struct {

	// The ID of the project owned by the authenticated user.
	// +optional
	// +immutable
	ProjectID *int `json:"projectId,omitempty"`

	// The user ID of the member.
	// +kubebuilder:example=123
	// +optional
	UserID *int `json:"userID,omitempty"`

	// The username of the member.
	// +kubebuilder:example="john.doe"
	// +optional
	UserName *string `json:"userName,omitempty"`

	// A valid access level.
	// +kubebuilder:example=30
	// +immutable
	AccessLevel AccessLevelValue `json:"accessLevel"`

	// A date string in the format YEAR-MONTH-DAY.
	// +kubebuilder:example="2024-12-31"
	// +optional
	ExpiresAt *string `json:"expiresAt,omitempty"`
}

// MemberObservation represents a project member.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/projects.html#list-project-team-members
type MemberObservation struct {
	Username  string       `json:"username,omitempty"`
	Email     string       `json:"email,omitempty"`
	Name      string       `json:"name,omitempty"`
	State     string       `json:"state,omitempty"`
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`
	WebURL    string       `json:"webURL,omitempty"`
	AvatarURL string       `json:"avatarURL,omitempty"`
}
