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

// AccessTokenParameters define the desired state of a Gitlab access token
// https://docs.gitlab.com/ee/api/access_tokens.html
type AccessTokenParameters struct {
	// ProjectID is the ID of the project to create the access token in.
	// +optional
	// +immutable
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1.Project
	ProjectID *string `json:"projectId,omitempty"`

	// ProjectIDRef is a reference to a project to retrieve its projectId
	// +optional
	// +immutable
	ProjectIDRef *xpv1.Reference `json:"projectIdRef,omitempty"`

	// ProjectIDSelector selects reference to a project to retrieve its projectId.
	// +optional
	ProjectIDSelector *xpv1.Selector `json:"projectIdSelector,omitempty"`

	// Expiration date of the access token. The date cannot be set later than the maximum allowable lifetime of an access token.
	// If not set, the maximum allowable lifetime of a personal access token is 365 days.
	// Expected in ISO 8601 format (2019-03-15T08:00:00Z)
	// +immutable
	ExpiresAt *metav1.Time `json:"expiresAt,omitempty"`

	// Access level for the project. Default is 40.
	// Valid values are 10 (Guest), 20 (Reporter), 30 (Developer), 40 (Maintainer), and 50 (Owner).
	// +optional
	// +immutable
	AccessLevel *AccessLevelValue `json:"accessLevel,omitempty"`

	// Scopes indicates the access token scopes.
	// Must be at least one of read_repository, read_registry, write_registry,
	// read_package_registry, or write_package_registry.
	// +immutable
	Scopes []string `json:"scopes"`

	// Name of the project access token
	// +required
	Name string `json:"name"`
}

// AccessTokenObservation represents a access token.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/project_access_tokens.html
type AccessTokenObservation struct {
	TokenID *int `json:"id,omitempty"`
}

// A AccessTokenSpec defines the desired state of a Gitlab Project.
type AccessTokenSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       AccessTokenParameters `json:"forProvider"`
}

// A AccessTokenStatus represents the observed state of a Gitlab Project.
type AccessTokenStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          AccessTokenObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A AccessToken is a managed resource that represents a Gitlab project access token
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gitlab}
type AccessToken struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AccessTokenSpec   `json:"spec"`
	Status AccessTokenStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AccessTokenList contains a list of Project items
type AccessTokenList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AccessToken `json:"items"`
}
