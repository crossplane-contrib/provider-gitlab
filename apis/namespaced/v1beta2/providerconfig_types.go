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
// to a single namespace, so they are not honoured here; use a
// ClusterProviderConfig if such sources are required.
//
// Fs, Env and SecretRef.Namespace are ignored, but they must stay declared:
// v1beta2 is the storage version and conversion from the deprecated v1beta1 is
// strategy None, so a field v1beta1 accepts and this version does not know
// makes the round trip through storage lossy. The API server performs exactly
// that round trip when it merges managed fields, which breaks server-side
// apply on v1beta1 as soon as a second field manager (the provider itself,
// writing its finalizer on the storage version) owns part of the object. That
// is the normal state for a ProviderConfig rendered by a Composition, since
// Crossplane applies composed resources server-side.
type ProviderCredentials struct {
	// Source of the provider credentials.
	// +kubebuilder:validation:Enum=None;Secret
	Source v2.CredentialsSource `json:"source"`

	// Method of authentification can be BasicAuth, JobToken, OAuthToken or PersonalAccessToken (default)
	// +optional
	Method auth.AuthType `json:"method"`

	// A SecretRef is a reference to a secret key that contains the credentials
	// that must be used to connect to the provider. The secret is always read
	// from this ProviderConfig's own namespace; secretRef.namespace is ignored.
	// +optional
	SecretRef *IgnoredNamespaceSecretKeySelector `json:"secretRef,omitempty"`

	// Fs is ignored. The credentials source enum does not accept Filesystem, so
	// nothing reads this; it is declared only to keep conversion from v1beta1
	// lossless.
	// +optional
	Fs *v2.FsSelector `json:"fs,omitempty"`

	// Env is ignored. The credentials source enum does not accept Environment,
	// so nothing reads this; it is declared only to keep conversion from
	// v1beta1 lossless.
	// +optional
	Env *v2.EnvSelector `json:"env,omitempty"`
}

// An IgnoredNamespaceSecretKeySelector references a credentials secret key. It
// is shaped like the cluster scoped v2.SecretKeySelector, but Namespace is
// never honoured: a namespaced ProviderConfig always reads its secret from its
// own namespace. The field is declared only to keep conversion from v1beta1
// lossless.
//
// Use Local to obtain the selector the provider resolves. Do not read
// Namespace.
type IgnoredNamespaceSecretKeySelector struct {
	// Name of the secret.
	Name string `json:"name"`

	// Namespace is ignored. The secret is read from the ProviderConfig's own
	// namespace.
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// The key to select.
	Key string `json:"key"`
}

// Local drops the ignored namespace and returns the selector the provider
// resolves, so no caller can accidentally read Namespace. It returns nil for a
// nil receiver, matching the optional field.
func (s *IgnoredNamespaceSecretKeySelector) Local() *v2.LocalSecretKeySelector {
	if s == nil {
		return nil
	}
	return &v2.LocalSecretKeySelector{
		LocalSecretReference: v2.LocalSecretReference{Name: s.Name},
		Key:                  s.Key,
	}
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
