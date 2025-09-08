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

// VariableType indicates the type of the GitLab CI variable.
type VariableType string

// VariableParameters define the desired state of a Gitlab CI Variable
// https://docs.gitlab.com/ee/api/project_level_variables.html
type VariableParameters struct {
	sharedProjectsV1alpha1.VariableParameters `json:",inline"`
	// ProjectIDRef is a reference to a project to retrieve its projectId.
	// +optional
	// +immutable
	ProjectIDRef *xpv1.NamespacedReference `json:"projectIdRef,omitempty"`

	// ProjectIDSelector selects reference to a project to retrieve its projectId.
	// +optional
	ProjectIDSelector *xpv1.NamespacedSelector `json:"projectIdSelector,omitempty"`

	// ValueSecretRef is used to obtain the value from a secret. This will set Masked and Raw to true if they
	// have not been set implicitly. Mutually exclusive with Value.
	// +optional
	// +nullable
	ValueSecretRef *xpv1.LocalSecretKeySelector `json:"valueSecretRef,omitempty"`
}

// A VariableSpec defines the desired state of a Gitlab Project CI
// Variable.
type VariableSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`
	ForProvider              VariableParameters `json:"forProvider"`
}

// A VariableStatus represents the observed state of a Gitlab Project CI
// Variable.
type VariableStatus struct {
	xpv1.ResourceStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// A Variable is a managed resource that represents a Gitlab CI variable.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}
type Variable struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VariableSpec   `json:"spec"`
	Status VariableStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VariableList contains a list of Variable items.
type VariableList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Variable `json:"items"`
}
