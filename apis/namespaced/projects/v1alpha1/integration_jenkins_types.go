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

// IntegrationJenkinsParameters defines the desired state of a GitLab Project Jenkins Integration.
type IntegrationJenkinsParameters struct {
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

	// URL of the Jenkins server.
	// +required
	URL string `json:"url"`

	// ProjectName is the name of the Jenkins project.
	// +required
	ProjectName string `json:"projectName"`

	// EnableSSLVerification enables SSL verification. GitLab defaults to true.
	// +optional
	EnableSSLVerification *bool `json:"enableSslVerification,omitempty"`

	// Username of the Jenkins server.
	// +optional
	Username *string `json:"username,omitempty"`

	// PasswordSecretRef is a reference to a secret that contains the password of the Jenkins server.
	// WARNING: This field is NOT reconciled as the GitLab API does not return it as it is a write-only field.
	// +optional
	PasswordSecretRef *v2.LocalSecretKeySelector `json:"passwordSecretRef,omitempty"`

	// PushEvents triggers a Jenkins build on push events.
	// +optional
	PushEvents *bool `json:"pushEvents,omitempty"`

	// MergeRequestsEvents triggers a Jenkins build on merge request events.
	// +optional
	MergeRequestsEvents *bool `json:"mergeRequestsEvents,omitempty"`

	// TagPushEvents triggers a Jenkins build on tag push events.
	// +optional
	TagPushEvents *bool `json:"tagPushEvents,omitempty"`
}

// IntegrationJenkinsObservation represents the observed state of a GitLab Project Jenkins Integration.
type IntegrationJenkinsObservation struct {
	v1alpha1.CommonIntegrationObservation `json:",inline"`

	// URL of the Jenkins server.
	URL string `json:"url,omitempty"`
	// ProjectName is the name of the Jenkins project.
	ProjectName string `json:"projectName,omitempty"`
	// EnableSSLVerification indicates whether SSL verification is enabled.
	EnableSSLVerification bool `json:"enableSslVerification,omitempty"`
	// Username of the Jenkins server.
	Username string `json:"username,omitempty"`
}

// A IntegrationJenkinsSpec defines the desired state of a GitLab Project Jenkins Integration.
type IntegrationJenkinsSpec struct {
	v2.ManagedResourceSpec `json:",inline"`
	// ForProvider represents the desired state of the Jenkins integration.
	ForProvider IntegrationJenkinsParameters `json:"forProvider"`
}

// A IntegrationJenkinsStatus represents the observed state of a GitLab Project Jenkins Integration.
type IntegrationJenkinsStatus struct {
	v2.ManagedResourceStatus `json:",inline"`
	// AtProvider represents the observed state of the Jenkins integration.
	AtProvider IntegrationJenkinsObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A IntegrationJenkins is a managed resource that represents a GitLab Project Jenkins Integration.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="PROJECT",type="string",JSONPath=".spec.forProvider.projectId"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}
type IntegrationJenkins struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              IntegrationJenkinsSpec   `json:"spec"`
	Status            IntegrationJenkinsStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IntegrationJenkinsList contains a list of IntegrationJenkins items.
type IntegrationJenkinsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IntegrationJenkins `json:"items"`
}

// IntegrationJenkins type metadata
var (
	IntegrationJenkinsKind             = reflect.TypeOf(IntegrationJenkins{}).Name()
	IntegrationJenkinsGroupKind        = schema.GroupKind{Group: Group, Kind: IntegrationJenkinsKind}.String()
	IntegrationJenkinsKindAPIVersion   = IntegrationJenkinsKind + "." + SchemeGroupVersion.String()
	IntegrationJenkinsGroupVersionKind = SchemeGroupVersion.WithKind(IntegrationJenkinsKind)
)
