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

// BadgeParameters define the desired state of a Gitlab project badge
// https://docs.gitlab.com/ee/api/badges.html
type BadgeParameters struct {
	// ProjectID is the ID of the project to create the deploy token in.
	// +optional
	// +immutable
	ProjectID *int `json:"projectId,omitempty"`

	// ProjectIDRef is a reference to a project to retrieve its projectId
	// +optional
	// +immutable
	ProjectIDRef *xpv1.NamespacedReference `json:"projectIdRef,omitempty"`

	// ProjectIDSelector selects reference to a project to retrieve its projectId.
	// +optional
	ProjectIDSelector *xpv1.NamespacedSelector `json:"projectIdSelector,omitempty"`

	// ID is the ID of an existing badge to import and manage.
	// If set, the controller will adopt the existing badge instead of creating a new one.
	// If the badge with this ID does not exist, resource creation will fail.
	// +optional
	// +immutable
	ID *int `json:"id,omitempty"`
	// LinkURL is the onclick redirect URL of the badge.
	// Supports gitlab format templating using variables like %{project_name}
	// +required
	LinkURL string `json:"linkURL"`
	// ImageURL is the display image URL of the badge.
	// Supports gitlab format templating using variables like %{project_name}
	// +required
	ImageURL string `json:"imageURL"`
	// Name is the display text of the badge. It is recommended to keep it short.
	// +optional
	Name *string `json:"name,omitempty"`
}

// BadgeObservation represents a project badge.
//
// GitLab API docs:
// https://docs.gitlab.com/api/project_badges/
type BadgeObservation struct {
	// ID of the badge
	ID int `json:"id,omitempty"`
	// LinkURL is the onclick redirect URL of the badge.
	LinkURL string `json:"linkURL,omitempty"`
	// RenderedLinkURL is the rendered onclick redirect URL of the badge.
	RenderedLinkURL string `json:"renderedLinkURL,omitempty"`
	// ImageURL is the display image URL of the badge.
	ImageURL string `json:"imageURL,omitempty"`
	// RenderedImageURL is the rendered display image URL of the badge.
	RenderedImageURL string `json:"renderedImageURL,omitempty"`
	// Name is the display text of the badge.
	Name string `json:"name,omitempty"`
}

// A BadgeSpec defines the desired state of a Gitlab Project Badge.
type BadgeSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`
	ForProvider              BadgeParameters `json:"forProvider"`
}

// A BadgeStatus represents the observed state of a Gitlab Project Badge.
type BadgeStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          BadgeObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Badge is a managed resource that represents a Gitlab project badge
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}
type Badge struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BadgeSpec   `json:"spec"`
	Status BadgeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BadgeList contains a list of Badge items
type BadgeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Badge `json:"items"`
}
