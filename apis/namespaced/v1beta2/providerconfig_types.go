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

package v1beta2

import (
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	auth "github.com/crossplane-contrib/provider-gitlab/pkg/common/auth"
)

// A ProviderConfigSpec defines the desired state of a namespaced ProviderConfig.
type ProviderConfigSpec struct {
	// Base URL of the Gitlab Service
	BaseURL string `json:"baseURL,omitempty"`

	// Credentials required to authenticate to this provider.
	Credentials ProviderCredentials `json:"credentials"`

	// InsecureSkipVerify ignores self signed TLS certificates when connecting
	// to Gitlab.
	InsecureSkipVerify *bool `json:"insecureSkipVerify,omitempty"`
}

// ProviderCredentials required to authenticate. As the ProviderConfig is
// namespaced, credentials come from a Secret in the same namespace. Filesystem
// and environment sources reference provider-global state and cannot be scoped
// to a single namespace, so they are not offered here; use a
// ClusterProviderConfig if such sources are required.
type ProviderCredentials struct {
	// Source of the provider credentials.
	// +kubebuilder:validation:Enum=None;Secret
	Source v2.CredentialsSource `json:"source"`

	// Method of authentification can be BasicAuth, JobToken, OAuthToken or PersonalAccessToken (default)
	// +optional
	Method auth.AuthType `json:"method"`

	// A SecretRef is a reference to a secret key, in the same namespace as this
	// ProviderConfig, that contains the credentials that must be used to connect
	// to the provider.
	// +optional
	SecretRef *v2.LocalSecretKeySelector `json:"secretRef,omitempty"`
}

// A ProviderConfigStatus represents the status of a ProviderConfig.
type ProviderConfigStatus struct {
	v2.ProviderConfigStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// A ProviderConfig configures how gitlab controller should connect to Gitlab API.
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="SECRET-NAME",type="string",JSONPath=".spec.credentials.secretRef.name",priority=1
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,provider,gitlab}
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
type ProviderConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProviderConfigSpec   `json:"spec"`
	Status ProviderConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProviderConfigList contains a list of ProviderConfig
type ProviderConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProviderConfig `json:"items"`
}
