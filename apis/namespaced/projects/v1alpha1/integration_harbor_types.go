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

	"github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
)

// IntegrationHarborParameters defines the desired state of a GitLab Project Harbor Integration.
type IntegrationHarborParameters struct {
	// ProjectID is the ID of the project.
	// +optional
	// +immutable
	ProjectID *int64 `json:"projectId,omitempty"`

	// ProjectIDRef is a reference to a project to retrieve its projectId.
	// +optional
	// +immutable
	ProjectIDRef *v2.NamespacedReference `json:"projectIdRef,omitempty"`

	// ProjectIDSelector selects reference to a project to retrieve its projectId.
	// +optional
	ProjectIDSelector *v2.NamespacedSelector `json:"projectIdSelector,omitempty"`

	// URL is the base URL of the Harbor instance which is being linked to this GitLab project. For example, https://demo.goharbor.io.
	// +required
	URL string `json:"url"`

	// ProjectName is the name of the Harbor project. Must follow Harbor's requirements.
	// +required
	ProjectName string `json:"projectName"`

	// Username is the Harbor user used to authenticate against the API.
	// +required
	Username string `json:"username"`

	// PasswordSecretRef is a reference to a secret that contains the password of the Harbor user.
	// WARNING: This field is NOT reconciled as the GitLab API does not return it as it is a write-only field.
	// +required
	PasswordSecretRef v2.LocalSecretKeySelector `json:"passwordSecretRef"`

	// UseInheritedSettings indicates whether to inherit the default Harbor settings from the parent group.
	// +optional
	UseInheritedSettings *bool `json:"useInheritedSettings,omitempty"`
}

// IntegrationHarborObservation represents the observed state of a GitLab Project Harbor Integration.
type IntegrationHarborObservation struct {
	v1alpha1.CommonIntegrationObservation `json:",inline"`

	// URL is the base URL of the Harbor instance.
	URL string `json:"url,omitempty"`
	// ProjectName is the name of the Harbor project.
	ProjectName string `json:"projectName,omitempty"`
	// Username is the Harbor user used to authenticate against the API.
	Username string `json:"username,omitempty"`
	// UseInheritedSettings indicates whether the default Harbor settings from the parent group are inherited.
	UseInheritedSettings bool `json:"useInheritedSettings,omitempty"`
}

// A IntegrationHarborSpec defines the desired state of a GitLab Project Harbor Integration.
type IntegrationHarborSpec struct {
	v2.ManagedResourceSpec `json:",inline"`
	// ForProvider represents the desired state of the harbor integration.
	ForProvider IntegrationHarborParameters `json:"forProvider"`
}

// A IntegrationHarborStatus represents the observed state of a GitLab Project Harbor Integration.
type IntegrationHarborStatus struct {
	v2.ManagedResourceStatus `json:",inline"`
	// AtProvider represents the observed state of the harbor integration.
	AtProvider IntegrationHarborObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A IntegrationHarbor is a managed resource that represents a GitLab Project Harbor Integration.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="PROJECT",type="string",JSONPath=".spec.forProvider.projectId"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}
type IntegrationHarbor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              IntegrationHarborSpec   `json:"spec"`
	Status            IntegrationHarborStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IntegrationHarborList contains a list of IntegrationHarbor items.
type IntegrationHarborList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IntegrationHarbor `json:"items"`
}
