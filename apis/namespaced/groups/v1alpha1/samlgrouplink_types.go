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

	sharedGroupsV1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/shared/groups/v1alpha1"
)

// SamlGroupLinkParameters define the desired state of a Gitlab Group Saml Link
// https://docs.gitlab.com/ee/api/groups.html#saml-group-links
type SamlGroupLinkParameters struct {
	sharedGroupsV1alpha1.SamlGroupLinkParameters `json:",inline"`

	// GroupIDRef is a reference to a group to retrieve its groupId
	// +optional
	// +immutable
	GroupIDRef *xpv1.NamespacedReference `json:"groupIdRef,omitempty"`

	// GroupIDSelector selects reference to a group to retrieve its groupId.
	// +optional
	GroupIDSelector *xpv1.NamespacedSelector `json:"groupIdSelector,omitempty"`
}

// A SamlGroupLinkSpec defines the desired state of a Gitlab SAML group sync.
type SamlGroupLinkSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`
	ForProvider              SamlGroupLinkParameters `json:"forProvider"`
}

// A SamlGroupLinkStatus represents the observed state of a Gitlab SAML group sync.
type SamlGroupLinkStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          sharedGroupsV1alpha1.SamlGroupLinkObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A SamlGroupLink is a managed resource that represents a Gitlab saml group sync connection
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".status.atProvider.name"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}
type SamlGroupLink struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SamlGroupLinkSpec   `json:"spec"`
	Status SamlGroupLinkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SamlGroupLinkList contains a list of group items
type SamlGroupLinkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SamlGroupLink `json:"items"`
}
