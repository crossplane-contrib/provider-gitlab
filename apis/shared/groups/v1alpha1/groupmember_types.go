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

// MemberSAMLIdentity represents the SAML Identity link for the group member.
// GitLab API docs: https://docs.gitlab.com/ce/api/members.html#list-all-members-of-a-group-or-project
// Gitlab MR for API change: https://gitlab.com/gitlab-org/gitlab/-/merge_requests/20357
// Gitlab MR for API Doc change: https://gitlab.com/gitlab-org/gitlab/-/merge_requests/25652
type MemberSAMLIdentity struct {
	ExternUID      string `json:"externUID"`
	Provider       string `json:"provider"`
	SAMLProviderID int    `json:"samlProviderID"`
}

// A MemberParameters defines the desired state of a Gitlab Group Member.
type MemberParameters struct {

	// The ID of the group owned by the authenticated user.
	// +kubebuilder:example=10
	// +optional
	// +immutable
	GroupID *int `json:"groupId,omitempty"`

	// The user ID of the member.
	// +kubebuilder:example=90
	// +optional
	UserID *int `json:"userID,omitempty"`

	// The userName of the member.
	// +kubebuilder:example=maxmustermann
	// +optional
	UserName *string `json:"userName,omitempty"`

	// A valid access level.
	// +kubebuilder:example=10
	// +immutable
	AccessLevel AccessLevelValue `json:"accessLevel"`

	// A date string in the format YEAR-MONTH-DAY.
	// +optional
	ExpiresAt *string `json:"expiresAt,omitempty"`
}

// MemberObservation represents a group member.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/groups.html#list-group-members
type MemberObservation struct {
	Username          string              `json:"username,omitempty"`
	Name              string              `json:"name,omitempty"`
	State             string              `json:"state,omitempty"`
	AvatarURL         string              `json:"avatarURL,omitempty"`
	WebURL            string              `json:"webURL,omitempty"`
	GroupSAMLIdentity *MemberSAMLIdentity `json:"groupSamlIdentity,omitempty"`
}
