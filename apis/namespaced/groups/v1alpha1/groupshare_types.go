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
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GroupShareParameters define the desired state of a GitLab group share: the
// shared group gives the members of the invited group access at the given
// level. https://docs.gitlab.com/api/groups/#share-groups-with-groups
// The two group IDs identify the share, so they cannot change once set.
// +kubebuilder:validation:XValidation:rule="!has(oldSelf.sharedGroupId) || (has(self.sharedGroupId) && self.sharedGroupId == oldSelf.sharedGroupId)",message="sharedGroupId cannot be changed or removed once set"
// +kubebuilder:validation:XValidation:rule="!has(oldSelf.invitedGroupId) || (has(self.invitedGroupId) && self.invitedGroupId == oldSelf.invitedGroupId)",message="invitedGroupId cannot be changed or removed once set"
type GroupShareParameters struct {
	// SharedGroupID is the ID of the group that is shared.
	// +optional
	// +immutable
	SharedGroupID *int64 `json:"sharedGroupId,omitempty"`

	// SharedGroupIDRef references the Group that is shared.
	// +optional
	// +immutable
	SharedGroupIDRef *v2.NamespacedReference `json:"sharedGroupIdRef,omitempty"`

	// SharedGroupIDSelector selects the Group that is shared.
	// +optional
	SharedGroupIDSelector *v2.NamespacedSelector `json:"sharedGroupIdSelector,omitempty"`

	// InvitedGroupID is the ID of the group whose members get access to the
	// shared group.
	// +optional
	// +immutable
	InvitedGroupID *int64 `json:"invitedGroupId,omitempty"`

	// InvitedGroupIDRef references the Group whose members get access.
	// +optional
	// +immutable
	InvitedGroupIDRef *v2.NamespacedReference `json:"invitedGroupIdRef,omitempty"`

	// InvitedGroupIDSelector selects the Group whose members get access.
	// +optional
	InvitedGroupIDSelector *v2.NamespacedSelector `json:"invitedGroupIdSelector,omitempty"`

	// GroupAccessLevel is the role the members of the invited group get in the
	// shared group: 5 Minimal Access (GitLab Premium and Ultimate), 10 Guest,
	// 15 Planner, 20 Reporter, 25 Security Manager (when enabled),
	// 30 Developer, 40 Maintainer, 50 Owner. Custom roles are not supported.
	// https://docs.gitlab.com/api/members/#roles
	// +kubebuilder:validation:Enum=5;10;15;20;25;30;40;50
	GroupAccessLevel int64 `json:"groupAccessLevel"`

	// ExpiresAt is the date the share expires, in the format 2026-12-31.
	// Copied from GitLab when not set.
	// +optional
	// +kubebuilder:validation:Pattern=`^\d{4}-\d{2}-\d{2}$`
	ExpiresAt *string `json:"expiresAt,omitempty"`
}

// GroupShareObservation is the observed state of a GitLab group share.
type GroupShareObservation struct {
	// InvitedGroupFullPath is the full path of the invited group.
	// +optional
	InvitedGroupFullPath string `json:"invitedGroupFullPath,omitempty"`

	// GroupAccessLevel is the role GitLab reports for the share.
	// +optional
	GroupAccessLevel int64 `json:"groupAccessLevel,omitempty"`

	// ExpiresAt is the expiry date GitLab reports for the share.
	// +optional
	ExpiresAt *string `json:"expiresAt,omitempty"`
}

// A GroupShareSpec defines the desired state of a GitLab group share.
type GroupShareSpec struct {
	v2.ManagedResourceSpec `json:",inline"`
	ForProvider            GroupShareParameters `json:"forProvider"`
}

// A GroupShareStatus represents the observed state of a GitLab group share.
type GroupShareStatus struct {
	v2.ManagedResourceStatus `json:",inline"`
	AtProvider               GroupShareObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A GroupShare shares a GitLab group with another group: the members of the
// invited group get access to the shared group. Use it instead of a Group's
// sharedWithGroups to manage shares one by one, for example with references
// in Compositions. Set manageSharedWithGroups to false on a Group whose
// shares are managed by GroupShares; otherwise that Group removes them.
// The provider's credentials must be able to read both groups, because GitLab
// only reports shares with groups the caller can see. GitLab cannot change a
// share, so a new access level or expiry date replaces it: the invited
// group's members briefly lose access, and keep it lost if creating the new
// share fails, until a later reconcile succeeds.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="SHARED-GROUP",type="integer",JSONPath=".spec.forProvider.sharedGroupId"
// +kubebuilder:printcolumn:name="INVITED-GROUP",type="integer",JSONPath=".spec.forProvider.invitedGroupId"
// +kubebuilder:printcolumn:name="ACCESS",type="integer",JSONPath=".spec.forProvider.groupAccessLevel"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}
type GroupShare struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GroupShareSpec   `json:"spec"`
	Status GroupShareStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GroupShareList contains a list of GroupShare items.
type GroupShareList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GroupShare `json:"items"`
}
