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

// IntegrationYouTrackParameters defines the desired state of a GitLab Project YouTrack Integration.
type IntegrationYouTrackParameters struct {
	// ProjectID is the ID of the project.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="projectId is immutable"
	ProjectID *int64 `json:"projectId,omitempty"`

	// ProjectIDRef is a reference to a project to retrieve its projectId.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="projectIdRef is immutable"
	ProjectIDRef *v2.NamespacedReference `json:"projectIdRef,omitempty"`

	// ProjectIDSelector selects reference to a project to retrieve its projectId.
	// +optional
	ProjectIDSelector *v2.NamespacedSelector `json:"projectIdSelector,omitempty"`

	// ProjectURL is the URL of the project in YouTrack.
	// +required
	// +kubebuilder:validation:Pattern=`^https?://\S+$`
	ProjectURL string `json:"projectUrl"`

	// IssuesURL is the URL to view an issue in YouTrack. GitLab replaces :id with the issue number.
	// +required
	// +kubebuilder:validation:Pattern=`^https?://\S+$`
	IssuesURL string `json:"issuesUrl"`
}

// IntegrationYouTrackObservation represents the observed state of a GitLab Project YouTrack Integration.
type IntegrationYouTrackObservation struct {
	v1alpha1.CommonIntegrationObservation `json:",inline"`

	// ProjectURL is the URL of the project in YouTrack.
	ProjectURL string `json:"projectUrl,omitempty"`
	// IssuesURL is the URL to view an issue in YouTrack.
	IssuesURL string `json:"issuesUrl,omitempty"`
}

// A IntegrationYouTrackSpec defines the desired state of a GitLab Project YouTrack Integration.
type IntegrationYouTrackSpec struct {
	v2.ManagedResourceSpec `json:",inline"`
	// ForProvider represents the desired state of the YouTrack integration.
	ForProvider IntegrationYouTrackParameters `json:"forProvider"`
}

// A IntegrationYouTrackStatus represents the observed state of a GitLab Project YouTrack Integration.
type IntegrationYouTrackStatus struct {
	v2.ManagedResourceStatus `json:",inline"`
	// AtProvider represents the observed state of the YouTrack integration.
	AtProvider IntegrationYouTrackObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A IntegrationYouTrack is a managed resource that represents a GitLab Project YouTrack Integration.
// GitLab allows only one issue tracker integration, such as Custom Issue Tracker, Redmine, or YouTrack, to be active per project.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="PROJECT",type="string",JSONPath=".spec.forProvider.projectId"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}
type IntegrationYouTrack struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              IntegrationYouTrackSpec   `json:"spec"`
	Status            IntegrationYouTrackStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IntegrationYouTrackList contains a list of IntegrationYouTrack items.
type IntegrationYouTrackList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IntegrationYouTrack `json:"items"`
}

// IntegrationYouTrack type metadata
var (
	IntegrationYouTrackKind             = reflect.TypeOf(IntegrationYouTrack{}).Name()
	IntegrationYouTrackGroupKind        = schema.GroupKind{Group: Group, Kind: IntegrationYouTrackKind}.String()
	IntegrationYouTrackKindAPIVersion   = IntegrationYouTrackKind + "." + SchemeGroupVersion.String()
	IntegrationYouTrackGroupVersionKind = SchemeGroupVersion.WithKind(IntegrationYouTrackKind)
)
