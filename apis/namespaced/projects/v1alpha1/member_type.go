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
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	sharedProjectsV1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/shared/projects/v1alpha1"
)

// A MemberParameters defines the desired state of a Gitlab Project Member.
type MemberParameters struct {
	sharedProjectsV1alpha1.MemberParameters `json:",inline"`

	// ProjectIDRef is a reference to a project to retrieve its projectId
	// +optional
	// +immutable
	ProjectIDRef *xpv1.NamespacedReference `json:"projectIdRef,omitempty"`

	// ProjectIDSelector selects reference to a project to retrieve its projectId.
	// +optional
	ProjectIDSelector *xpv1.NamespacedSelector `json:"projectIdSelector,omitempty"`
}

// A MemberSpec defines the desired state of a Gitlab Project Member.
type MemberSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`
	ForProvider              MemberParameters `json:"forProvider"`
}

// A MemberStatus represents the observed state of a Gitlab Project Member.
type MemberStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          sharedProjectsV1alpha1.MemberObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Member is a managed resource that represents a Gitlab Project Member
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Project ID",type="integer",JSONPath=".spec.forProvider.projectId"
// +kubebuilder:printcolumn:name="Username",type="string",JSONPath=".status.atProvider.username"
// +kubebuilder:printcolumn:name="Acceess Level",type="integer",JSONPath=".spec.forProvider.accessLevel"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}
type Member struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MemberSpec   `json:"spec"`
	Status MemberStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MemberList contains a list of Member items
type MemberList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Member `json:"items"`
}
