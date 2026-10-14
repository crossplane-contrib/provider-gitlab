/*
Copyright 2026 The Crossplane Authors.

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
	"reflect"

	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
)

// IntegrationExternalWikiParameters defines the desired state of a GitLab Project External Wiki Integration.
type IntegrationExternalWikiParameters struct {
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

	// ExternalWikiURL is the URL of the external wiki.
	// +required
	// +kubebuilder:validation:Pattern=`^https?://.+`
	ExternalWikiURL string `json:"externalWikiUrl"`
}

// IntegrationExternalWikiObservation represents the observed state of a GitLab Project External Wiki Integration.
type IntegrationExternalWikiObservation struct {
	v1alpha1.CommonIntegrationObservation `json:",inline"`

	// ExternalWikiURL is the URL of the external wiki.
	ExternalWikiURL string `json:"externalWikiUrl,omitempty"`
}

// A IntegrationExternalWikiSpec defines the desired state of a GitLab Project External Wiki Integration.
type IntegrationExternalWikiSpec struct {
	v2.ManagedResourceSpec `json:",inline"`
	// ForProvider represents the desired state of the external wiki integration.
	ForProvider IntegrationExternalWikiParameters `json:"forProvider"`
}

// A IntegrationExternalWikiStatus represents the observed state of a GitLab Project External Wiki Integration.
type IntegrationExternalWikiStatus struct {
	v2.ManagedResourceStatus `json:",inline"`
	// AtProvider represents the observed state of the external wiki integration.
	AtProvider IntegrationExternalWikiObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A IntegrationExternalWiki is a managed resource that represents a GitLab Project External Wiki Integration.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="PROJECT",type="string",JSONPath=".spec.forProvider.projectId"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}
type IntegrationExternalWiki struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              IntegrationExternalWikiSpec   `json:"spec"`
	Status            IntegrationExternalWikiStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IntegrationExternalWikiList contains a list of IntegrationExternalWiki items.
type IntegrationExternalWikiList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IntegrationExternalWiki `json:"items"`
}

// IntegrationExternalWiki type metadata
var (
	IntegrationExternalWikiKind             = reflect.TypeOf(IntegrationExternalWiki{}).Name()
	IntegrationExternalWikiGroupKind        = schema.GroupKind{Group: Group, Kind: IntegrationExternalWikiKind}.String()
	IntegrationExternalWikiKindAPIVersion   = IntegrationExternalWikiKind + "." + SchemeGroupVersion.String()
	IntegrationExternalWikiGroupVersionKind = SchemeGroupVersion.WithKind(IntegrationExternalWikiKind)
)
