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

// +kubebuilder:object:generate=true

package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	// +cluster-scope:delete=1
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
)

// ServiceAccountParameters defines the desired state of Gitlab Group ServiceAccount
type ServiceAccountParameters struct {
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

	commonv1alpha1.CommonServiceAccountParameters `json:",inline"`
}

// ServiceAccountObservation represents the observed state of the Gitlab Group ServiceAccount
type ServiceAccountObservation struct {
	commonv1alpha1.CommonServiceAccountObservation `json:",inline"`
}

// A ServiceAccountSpec defines the desired state of a GitLab group service account.
type ServiceAccountSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`
	// Defines the desired state of the ServiceAccount.
	ForProvider ServiceAccountParameters `json:"forProvider"`
}

// A ServiceAccountStatus represents the observed state of the GitLab group service account.
type ServiceAccountStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	// Represents the observed state of the ServiceAccount.
	AtProvider ServiceAccountObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A ServiceAccount is a managed resource that represents a Gitlab group ServiceAccount.
// This is only available with at least a Premium license.
//
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}
type ServiceAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceAccountSpec   `json:"spec"`
	Status ServiceAccountStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServiceAccountList contains a list of ServiceAccount items.
type ServiceAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceAccount `json:"items"`
}
