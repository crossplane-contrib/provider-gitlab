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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	sharedGroupsV1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/shared/groups/v1alpha1"
)

// DeployTokenParameters define the desired state of a Gitlab deploy token
// https://docs.gitlab.com/ee/api/deploy_tokens.html
type DeployTokenParameters struct {
	sharedGroupsV1alpha1.DeployTokenParameters `json:",inline"`

	// GroupIDRef is a reference to a group to retrieve its groupId
	// +optional
	// +immutable
	GroupIDRef *xpv1.Reference `json:"groupIdRef,omitempty"`

	// GroupIDSelector selects reference to a group to retrieve its groupId.
	// +optional
	GroupIDSelector *xpv1.Selector `json:"groupIdSelector,omitempty"`
}

// A DeployTokenSpec defines the desired state of a Gitlab Group.
type DeployTokenSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       DeployTokenParameters `json:"forProvider"`
}

// A DeployTokenStatus represents the observed state of a Gitlab Group.
type DeployTokenStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          sharedGroupsV1alpha1.DeployTokenObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A DeployToken is a managed resource that represents a Gitlab deploy token
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gitlab}
type DeployToken struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeployTokenSpec   `json:"spec"`
	Status DeployTokenStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DeployTokenList contains a list of Group items
type DeployTokenList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DeployToken `json:"items"`
}
