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

// AccessTokenParameters define the desired state of a Gitlab access token
// https://docs.gitlab.com/ee/api/access_tokens.html
type AccessTokenParameters struct {
	// GroupID is the ID of the group to create the deploy token in.
	// +optional
	// +immutable
	GroupID *int `json:"groupId,omitempty"`

	// GroupIDRef is a reference to a group to retrieve its groupId
	// +optional
	// +immutable
	GroupIDRef *xpv1.NamespacedReference `json:"groupIdRef,omitempty"`

	// GroupIDSelector selects reference to a group to retrieve its groupId.
	// +optional
	GroupIDSelector *xpv1.NamespacedSelector `json:"groupIdSelector,omitempty"`

	// Expiration date of the access token. The date cannot be set later than the maximum allowable lifetime of an access token.
	// If not set, the maximum allowable lifetime of a group access token is configured to the maximum allowable lifetime limit.
	// Expected in ISO 8601 format (2019-03-15T08:00:00Z)
	// +immutable
	ExpiresAt *metav1.Time `json:"expiresAt"`

	// Access level for the group. Default is 40.
	// Valid values are 10 (Guest), 20 (Reporter), 30 (Developer), 40 (Maintainer), and 50 (Owner).
	// +optional
	// +immutable
	AccessLevel *AccessLevelValue `json:"accessLevel,omitempty"`

	// Scopes indicates the access token scopes.
	// Must be at least one of read_repository, read_registry, write_registry,
	// read_package_registry, or write_package_registry.
	// +immutable
	Scopes []string `json:"scopes"`

	// Name of the group access token
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

// A AccessTokenSpec defines the desired state of a Gitlab group.
type AccessTokenSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`
	ForProvider              AccessTokenParameters `json:"forProvider"`
}

// A AccessTokenStatus represents the observed state of a Gitlab group.
type AccessTokenStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          AccessTokenObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A AccessToken is a managed resource that represents a Gitlab group access token
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}
type AccessToken struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AccessTokenSpec   `json:"spec"`
	Status AccessTokenStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AccessTokenList contains a list of group items
type AccessTokenList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AccessToken `json:"items"`
}
