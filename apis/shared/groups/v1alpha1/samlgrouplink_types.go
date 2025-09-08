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

// SamlGroupLinkParameters define the desired state of a Gitlab Group Saml Link
// https://docs.gitlab.com/ee/api/groups.html#saml-group-links
type SamlGroupLinkParameters struct {
	// GroupID is the ID of the group to create the saml group link for.
	// +kubebuilder:example=10
	// +optional
	// +immutable
	GroupID *int `json:"groupId,omitempty"`

	// name is the name of the saml group to attach to the gitlab group
	// +kubebuilder:example="my-saml-group"
	// +immutable
	Name *string `json:"name"`

	// accessLevel is the defined role for members of the SAML group
	// +kubebuilder:example=10
	// +immutable
	AccessLevel AccessLevelValue `json:"accessLevel"`

	// memberRoleID is the defined member role assigned to members of the group
	// +optional
	// +immutable
	MemberRoleID *int `json:"memberRoleId,omitempty"`
}

// SamlGroupLinkObservation represents a Group Saml Link.
type SamlGroupLinkObservation struct {
	Name string `json:"name,omitempty"`
}
