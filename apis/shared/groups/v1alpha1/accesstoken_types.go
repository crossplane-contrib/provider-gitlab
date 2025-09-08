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

// AccessTokenParameters define the desired state of a Gitlab access token
// https://docs.gitlab.com/ee/api/access_tokens.html
type AccessTokenParameters struct {
	// GroupID is the ID of the group to create the deploy token in.
	// +kubebuilder:example=10
	// +optional
	// +immutable
	GroupID *int `json:"groupId,omitempty"`

	// Expiration date of the access token. The date cannot be set later than the maximum allowable lifetime of an access token.
	// If not set, the maximum allowable lifetime of a personal access token is 365 days.
	// Expected in ISO 8601 format (2019-03-15T08:00:00Z)
	// +immutable
	ExpiresAt *metav1.Time `json:"expiresAt,omitempty"`

	// Access level for the group. Default is 40.
	// Valid values are 10 (Guest), 20 (Reporter), 30 (Developer), 40 (Maintainer), and 50 (Owner).
	// +kubebuilder:example=40
	// +optional
	// +immutable
	AccessLevel *AccessLevelValue `json:"accessLevel,omitempty"`

	// Scopes indicates the access token scopes.
	// Must be at least one of read_repository, read_registry, write_registry,
	// read_package_registry, or write_package_registry.
	// +kubebuilder:example={read_repository,read_registry,write_registry,read_package_registry,write_package_registry}
	// +immutable
	Scopes []string `json:"scopes"`

	// Name of the group access token
	// +kubebuilder:example="my-access-token"
	// +required
	Name string `json:"name"`
}

// AccessTokenObservation represents a access token.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_access_tokens.html
type AccessTokenObservation struct {
	TokenID *int `json:"id,omitempty"`
}
