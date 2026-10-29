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
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
)

// ServiceAccountParameters defines the desired state of a Gitlab Project ServiceAccount
// +kubebuilder:validation:XValidation:rule="!(has(self.avatar) && has(self.avatarSecretRef))",message="avatar and avatarSecretRef are mutually exclusive"
type ServiceAccountParameters struct {
	// ProjectID is the ID of the project to create the service account in.
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

	commonv1alpha1.CommonServiceAccountParameters `json:",inline"`

	// AvatarSecretRef selects a Secret key that holds the avatar image as raw
	// bytes, as `kubectl create secret generic <name> --from-file=avatar=<image>`
	// stores it. The rules of avatar apply. The Secret is read on every poll, so
	// a changed image is uploaded on the next poll. Mutually exclusive with avatar.
	// +optional
	AvatarSecretRef *v2.LocalSecretKeySelector `json:"avatarSecretRef,omitempty"`
}

// ServiceAccountObservation represents the observed state of the Gitlab Project ServiceAccount
type ServiceAccountObservation struct {
	commonv1alpha1.CommonServiceAccountObservation `json:",inline"`
}

// A ServiceAccountSpec defines the desired state of a GitLab project service account.
type ServiceAccountSpec struct {
	v2.ManagedResourceSpec `json:",inline"`
	// Defines the desired state of the ServiceAccount.
	ForProvider ServiceAccountParameters `json:"forProvider"`
}

// A ServiceAccountStatus represents the observed state of the GitLab project service account.
type ServiceAccountStatus struct {
	v2.ManagedResourceStatus `json:",inline"`
	// Represents the observed state of the ServiceAccount.
	AtProvider ServiceAccountObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A ServiceAccount is a managed resource that represents a Gitlab project ServiceAccount.
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
