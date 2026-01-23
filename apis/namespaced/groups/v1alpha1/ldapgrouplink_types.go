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

// LdapGroupLinkParameters define the desired state of a Gitlab Group Ldap Link
// https://docs.gitlab.com/api/group_ldap_links/
type LdapGroupLinkParameters struct {
	// GroupID is the ID of the group to create the deploy token in.
	// +optional
	// +immutable
	GroupID *int `json:"groupId,omitempty"`

	// GroupIDRef is a reference to a group to retrieve its groupId
	// +optional
	// +immutable
	GroupIDRef *xpv1.NamespacedReference `json:"groupIdRef,omitempty"`

	// GroupIDSelector selects reference to a group to retrieve its groupId.
	// +optional
	GroupIDSelector *xpv1.NamespacedSelector `json:"groupIdSelector,omitempty"`

	// GroupAccess is the defined role for members of the LDAP group
	// +immutable
	GroupAccess AccessLevelValue `json:"groupAccess"`

	// LdapProvider provider ID for the LDAP group link
	// +immutable
	LdapProvider string `json:"ldapProvider"`

	// The CN of an LDAP group
	// +immutable
	CN string `json:"cn"`
}

// LdapGroupLinkObservation represents a Group Ldap Link.
type LdapGroupLinkObservation struct {
	CN string `json:"cn,omitempty"`
}

// A LdapGroupLinkSpec defines the desired state of a Gitlab Ldap group sync.
type LdapGroupLinkSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`
	ForProvider              LdapGroupLinkParameters `json:"forProvider"`
}

// A LdapGroupLinkStatus represents the observed state of a Gitlab Ldap group sync.
type LdapGroupLinkStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          LdapGroupLinkObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A LdapGroupLink is a managed resource that represents a Gitlab Ldap group sync connection
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".status.atProvider.name"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}
type LdapGroupLink struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LdapGroupLinkSpec   `json:"spec"`
	Status LdapGroupLinkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LdapGroupLinkList contains a list of group items
type LdapGroupLinkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LdapGroupLink `json:"items"`
}
