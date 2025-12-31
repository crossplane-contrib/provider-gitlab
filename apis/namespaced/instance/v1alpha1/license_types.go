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

// LicenseParameters define the desired state of a GitLab License
//
// GitLab API docs:
// https://docs.gitlab.com/api/license/
type LicenseParameters struct {
	// License is the license key to be applied to the GitLab instance.
	// This field can be set to a valid license string.
	// +optional
	License *string `json:"license,omitempty"`

	// LicenseSecretRef references a secret key selector that contains the license key.
	// This allows for secure storage and retrieval of the license key from a Kubernetes secret.
	// +optional
	LicenseSecretRef *xpv1.LocalSecretKeySelector `json:"licenseSecretRef,omitempty"`

	// LicenseEndpointURL is the URL of the license endpoint.
	// This can be used to fetch the license from a remote service.
	// +optional
	LicenseEndpointURL *string `json:"licenseEndpointURL,omitempty"`

	// LicenseEndpointURLSecretRef references a secret key selector that contains the license endpoint URL.
	// This allows for secure storage and retrieval of the URL from a Kubernetes secret.
	// +optional
	LicenseEndpointURLSecretRef *xpv1.LocalSecretKeySelector `json:"licenseEndpointURLSecretRef,omitempty"`

	// LicenseEndpointUsername is the username for authenticating with the license endpoint.
	// This will be combined with the password in the Authorization Basic header.
	// +optional
	LicenseEndpointUsername *string `json:"licenseEndpointUsername,omitempty"`

	// LicenseEndpointUsernameSecretRef references a secret key selector that contains the username for the license endpoint.
	// This allows for secure storage and retrieval of the username from a Kubernetes secret.
	// This will be combined with the password in the Authorization Basic header.
	// +optional
	LicenseEndpointUsernameSecretRef *xpv1.LocalSecretKeySelector `json:"licenseEndpointUsernameSecretRef,omitempty"`

	// LicenseEndpointPassword is the password for authenticating with the license endpoint.
	// This will be combined with the username in the Authorization Basic header.
	// +optional
	LicenseEndpointPassword *string `json:"licenseEndpointPassword,omitempty"`

	// LicenseEndpointPasswordSecretRef references a secret key selector that contains the password for the license endpoint.
	// This allows for secure storage and retrieval of the password from a Kubernetes secret.
	// This will be combined with the username in the Authorization Basic header.
	// +optional
	LicenseEndpointPasswordSecretRef *xpv1.LocalSecretKeySelector `json:"licenseEndpointPasswordSecretRef,omitempty"`

	// LicenseEndpointToken is the token for authenticating with the license endpoint.
	// This will be used as a Authorization Bearer token in the Authorization header.
	// +optional
	LicenseEndpointToken *string `json:"licenseEndpointToken,omitempty"`

	// LicenseEndpointTokenSecretRef references a secret key selector that contains the token for the license endpoint.
	// This allows for secure storage and retrieval of the token from a Kubernetes secret.
	// This will be used as a Authorization Bearer token in the Authorization header.
	// +optional
	LicenseEndpointTokenSecretRef *xpv1.LocalSecretKeySelector `json:"licenseEndpointTokenSecretRef,omitempty"`
}

// LicenseObservation represents the observed state of an instance License.
//
// GitLab API docs: https://docs.gitlab.com/api/license/
type LicenseObservation struct {
	// ID is the unique identifier of the license.
	ID int `json:"id"`
	// Plan is the name of the license plan.
	Plan string `json:"plan"`
	// CreatedAt is the timestamp when the license was created.
	CreatedAt *metav1.Time `json:"createdAt"`
	// StartsAt is the timestamp when the license becomes valid.
	StartsAt *metav1.Time `json:"startsAt"`
	// ExpiresAt is the timestamp when the license expires.
	ExpiresAt *metav1.Time `json:"expiresAt"`
	// HistoricalMax represents the maximum number of users once the license has expired.
	HistoricalMax int `json:"historicalMax"`
	// MaximumUserCount is the maximum number of users allowed by the license.
	MaximumUserCount int `json:"maximumUserCount"`
	// Expired indicates whether the license has expired.
	// When true, the license is no longer valid and the resource status will be defined as not ready.
	Expired bool `json:"expired"`
	// Overage is the difference between the number of billable users and the licensed number of users.
	Overage int `json:"overage"`
	// UserLimit is the user limit defined by the license.
	UserLimit int `json:"userLimit"`
	// ActiveUsers is the current number (billable) of active users.
	ActiveUsers int `json:"activeUsers"`
	// Licensee contains information about the licensee.
	Licensee Licensee `json:"licensee"`
	// AddOns contains information about any add-ons included with the license.
	AddOns AddOns `json:"addOns"`
}

// Licensee contains information about the licensee.
type Licensee struct {
	// Name is the name of the licensee.
	Name string `json:"name"`
	// Company is the company of the licensee.
	Company string `json:"company"`
	// Email is the email of the licensee.
	Email string `json:"email"`
}

// Add on codes that may occur in legacy licenses that don't have a plan yet.
// https://gitlab.com/gitlab-org/gitlab/-/blob/master/ee/app/models/license.rb
type AddOns struct {
	// GitLabAuditorUser indicates if the GitLab Auditor User add-on is included.
	GitLabAuditorUser int `json:"gitlabAuditorUser"`
	// GitLabDeployBoard indicates if the GitLab Deploy Board add-on is included.
	GitLabDeployBoard int `json:"gitlabDeployBoard"`
	// GitLabFileLocks indicates if the GitLab File Locks add-on is included.
	GitLabFileLocks int `json:"gitlabFileLocks"`
	// GitLabGeo indicates if the GitLab Geo add-on is included.
	GitLabGeo int `json:"gitlabGeo"`
	// GitLabServiceDesk indicates if the GitLab Service Desk add-on is included.
	GitLabServiceDesk int `json:"gitlabServiceDesk"`
}

// LicenseSpec defines the desired state of a instance License.
// This includes the configuration parameters for creating and managing
// a GitLab License linked to a specific instance.
type LicenseSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`
	// ForProvider contains the desired state of the License
	ForProvider LicenseParameters `json:"forProvider"`
}

// LicenseStatus represents the observed state of a instance License.
// This includes the current status and properties of the license as
// reported by the GitLab API.
type LicenseStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	// AtProvider contains the observed state of the License
	AtProvider LicenseObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A License is a managed resource that represents a GitLab License linked to a instance.
//
// WARNING: This resource should not be used multiple times for the same GitLab instance.
//
// IMPORTANT: If you choose to use a LicenseSecretRef OR a LicenseEndpointURL / LicenseEndpointURLSecretRef,
// You must configure the writeConnectionSecretToRef or publishConnectionDetailsTo to receive the license key.
//
// Example usage:
//
//	spec:
//	  writeConnectionSecretToRef:
//	    name: my-license-key
//	    namespace: default
//
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}
type License struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LicenseSpec   `json:"spec"`
	Status LicenseStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LicenseList contains a list of instance License resources.
type LicenseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []License `json:"items"`
}
