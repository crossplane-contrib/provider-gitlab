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
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceAccountAccessTokenParameters define the desired state of a GitLab project
// service account personal access token.
// https://docs.gitlab.com/api/service_accounts/
// +kubebuilder:validation:XValidation:rule="(has(self.expiresAt) ? 1 : 0) + (has(self.renewalPeriodDays) ? 1 : 0) == 1",message="exactly one of expiresAt or renewalPeriodDays must be set"
// +kubebuilder:validation:XValidation:rule="!has(self.renewBeforeDays) || has(self.renewalPeriodDays)",message="renewBeforeDays requires renewalPeriodDays"
// +kubebuilder:validation:XValidation:rule="!has(self.renewBeforeDays) || !has(self.renewalPeriodDays) || self.renewalPeriodDays > self.renewBeforeDays",message="renewalPeriodDays must be greater than renewBeforeDays"
type ServiceAccountAccessTokenParameters struct {
	// ProjectID is the ID of the project that owns the service account.
	// +optional
	// +immutable
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="projectId is immutable"
	ProjectID *int64 `json:"projectId,omitempty"`

	// ProjectIDRef is a reference to a project to retrieve its projectId.
	// +optional
	// +immutable
	ProjectIDRef *v2.NamespacedReference `json:"projectIdRef,omitempty"`

	// ProjectIDSelector selects reference to a project to retrieve its projectId.
	// +optional
	ProjectIDSelector *v2.NamespacedSelector `json:"projectIdSelector,omitempty"`

	// ServiceAccountID is the ID of the service account user the personal
	// access token belongs to.
	// +optional
	// +immutable
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="serviceAccountId is immutable"
	ServiceAccountID *int64 `json:"serviceAccountId,omitempty"`

	// ServiceAccountIDRef is a reference to a ServiceAccount to retrieve its serviceAccountId.
	// +optional
	// +immutable
	ServiceAccountIDRef *v2.NamespacedReference `json:"serviceAccountIdRef,omitempty"`

	// ServiceAccountIDSelector selects a reference to a ServiceAccount to retrieve its serviceAccountId.
	// +optional
	ServiceAccountIDSelector *v2.NamespacedSelector `json:"serviceAccountIdSelector,omitempty"`

	// ExpiresAt is the expiration date of the access token in ISO 8601 format (2019-03-15T08:00:00Z).
	// The date cannot be set later than the maximum allowable lifetime of an access token.
	// Since GitLab 16.0, tokens must have an expiration date.
	// Mutually exclusive with RenewalPeriodDays.
	// +optional
	ExpiresAt *metav1.Time `json:"expiresAt,omitempty"`

	// RenewalPeriodDays is the number of days each token generation should live.
	// The provider rotates it shortly before expiry, or once inactive, setting the new
	// expiry to RenewalPeriodDays days from the rotation time.
	// Mutually exclusive with ExpiresAt.
	// +kubebuilder:validation:Minimum=1
	// +optional
	RenewalPeriodDays *int `json:"renewalPeriodDays,omitempty"`

	// RenewBeforeDays overrides the default 2/3-lifetime renewal threshold.
	// When set, the provider rotates the token when fewer than RenewBeforeDays
	// remain before expiry. Only valid with RenewalPeriodDays.
	// Must be less than RenewalPeriodDays.
	// +kubebuilder:validation:Minimum=1
	// +optional
	RenewBeforeDays *int `json:"renewBeforeDays,omitempty"`

	// Scopes indicates the access token scopes.
	// Must be at least one of api, read_api, read_repository, write_repository,
	// read_registry, write_registry, read_package_registry, write_package_registry,
	// or self_rotate.
	// +immutable
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="scopes is immutable"
	Scopes []string `json:"scopes"`

	// Name of the service account access token
	// +required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="name is immutable"
	Name string `json:"name"`

	// Description of the service account access token
	// WARNING: this field is only reconciled on expiration / revokation of the token
	// +optional
	Description *string `json:"description,omitempty"`
}

// ServiceAccountAccessTokenObservation represents a service account personal access token.
//
// GitLab API docs:
// https://docs.gitlab.com/api/service_accounts/
type ServiceAccountAccessTokenObservation struct {
	ID          int64        `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	UserID      int64        `json:"userId"`
	Scopes      []string     `json:"scopes"`
	ExpiresAt   *metav1.Time `json:"expiresAt,omitempty"`
	Active      bool         `json:"active"`
	CreatedAt   *metav1.Time `json:"createdAt"`
	LastUsedAt  *metav1.Time `json:"lastUsedAt,omitempty"`
	Revoked     bool         `json:"revoked"`
}

// A ServiceAccountAccessTokenSpec defines the desired state of a GitLab service account access token.
type ServiceAccountAccessTokenSpec struct {
	v2.ManagedResourceSpec `json:",inline"`
	ForProvider            ServiceAccountAccessTokenParameters `json:"forProvider"`
}

// A ServiceAccountAccessTokenStatus represents the observed state of a GitLab service account access token.
type ServiceAccountAccessTokenStatus struct {
	v2.ManagedResourceStatus `json:",inline"`
	AtProvider               ServiceAccountAccessTokenObservation `json:"atProvider,omitempty"`

	// RenewAt is the computed time at which the provider will renew
	// the token. Only populated for renewal-managed tokens.
	// +optional
	RenewAt *metav1.Time `json:"renewAt,omitempty"`
}

// +kubebuilder:object:root=true

// A ServiceAccountAccessToken is a managed resource that represents a GitLab project
// service account personal access token.
// This is only available with at least a Premium license.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="SELF",type="string",JSONPath=".status.conditions[?(@.type=='SelfManaged')].status"
// +kubebuilder:printcolumn:name="RENEW-AT",type="string",JSONPath=".status.renewAt"
// +kubebuilder:printcolumn:name="EXPIRES-AT",type="string",JSONPath=".status.atProvider.expiresAt"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab},shortName=psaat
type ServiceAccountAccessToken struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceAccountAccessTokenSpec   `json:"spec"`
	Status ServiceAccountAccessTokenStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServiceAccountAccessTokenList contains a list of ServiceAccountAccessToken items
type ServiceAccountAccessTokenList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceAccountAccessToken `json:"items"`
}
