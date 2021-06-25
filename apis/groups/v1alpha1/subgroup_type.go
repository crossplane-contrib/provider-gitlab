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
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SubGroupParameters define the desired state of a Gitlab Group
type SubGroupParameters struct {
	// The path of the group.
	// +immutable
	Path string `json:"path"`

	// The group’s description.
	// +optional
	Description *string `json:"description,omitempty"`

	// Prevent adding new members to project membership within this group.
	// +optional
	MembershipLock *bool `json:"membershipLock,omitempty"`

	// The group’s visibility. Can be private, internal, or public.
	// +optional
	Visibility *VisibilityValue `json:"visibility,omitempty"`

	// Prevent sharing a project with another group within this group.
	// +optional
	ShareWithGroupLock *bool `json:"shareWithGroupLock,omitempty"`

	// Require all users in this group to setup Two-factor authentication.
	// +optional
	RequireTwoFactorAuth *bool `json:"requireTwoFactorAuthentication,omitempty"`

	// Time before Two-factor authentication is enforced (in hours).
	// +optional
	TwoFactorGracePeriod *int `json:"twoFactorGracePeriod,omitempty"`

	// developers can create projects in the group.
	// Can be noone (No one), maintainer (Maintainers), or developer (Developers + Maintainers).
	// +optional
	ProjectCreationLevel *ProjectCreationLevelValue `json:"projectCreationLevel,omitempty"`

	// Default to Auto DevOps pipeline for all projects within this group.
	// +optional
	AutoDevopsEnabled *bool `json:"autoDevopsEnabled,omitempty"`

	// Allowed to create subgroups. Can be owner (Owners), or maintainer (Maintainers).
	// +optional
	SubGroupCreationLevel *SubGroupCreationLevelValue `json:"subgroupCreationLevel,omitempty"`

	// Disable email notifications.
	// +optional
	EmailsDisabled *bool `json:"emailsDisabled,omitempty"`

	// Disable the capability of a group from getting mentioned.
	// +optional
	MentionsDisabled *bool `json:"mentionsDisabled,omitempty"`

	// Enable/disable Large File Storage (LFS) for the projects in this group.
	// +optional
	LFSEnabled *bool `json:"lfsEnabled,omitempty"`

	// Allow users to request member access.
	// +optional
	RequestAccessEnabled *bool `json:"requestAccessEnabled,omitempty"`

	// The parent group ID for creating nested group.
	// +optional
	ParentID *int `json:"parentId,omitempty"`

	// ParentIDRef is a reference to a group to retrieve its parentId
	// +optional
	// +immutable
	ParentIDRef *xpv1.Reference `json:"parentIdRef,omitempty"`

	// ParentIDSelector selects reference to a group to retrieve its parentId.
	// +optional
	ParentIDSelector *xpv1.Selector `json:"parentIdSelector,omitempty"`

	// Pipeline minutes quota for this group (included in plan).
	// Can be nil (default; inherit system default), 0 (unlimited) or > 0.
	// +optional
	SharedRunnersMinutesLimit *int `json:"sharedRunnersMinutesLimit,omitempty"`

	// Extra pipeline minutes quota for this group (purchased in addition to the minutes included in the plan).
	// +optional
	ExtraSharedRunnersMinutesLimit *int `json:"extraSharedRunnersMinutesLimit,omitempty"`
}

// SubGroupObservation is the observed state of a Group.
type SubGroupObservation struct {
	ID                  int                `json:"id,omitempty"`
	AvatarURL           string             `json:"avatarUrl,omitempty"`
	WebURL              string             `json:"webUrl,omitempty"`
	FullName            string             `json:"fullName,omitempty"`
	FullPath            string             `json:"fullPath,omitempty"`
	Statistics          *StorageStatistics `json:"statistics,omitempty"`
	CustomAttributes    []CustomAttribute  `json:"customAttributes,omitempty"`
	RunnersToken        string             `json:"runnersToken,omitempty"`
	SharedWithGroups    []SharedWithGroups `json:"sharedWithGroups,omitempty"`
	LDAPCN              string             `json:"ldapCn,omitempty"`
	LDAPAccess          AccessLevelValue   `json:"ldapAccess,omitempty"`
	LDAPGroupLinks      []LDAPGroupLink    `json:"ldapGroupLinks,omitempty"`
	MarkedForDeletionOn *metav1.Time       `json:"markedForDeletionOn,omitempty"`
	CreatedAt           *metav1.Time       `json:"createdAt,omitempty"`
}

// A SubGroupSpec defines the desired state of a Gitlab Group.
type SubGroupSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       SubGroupParameters `json:"forProvider"`
}

// A SubGroupStatus represents the observed state of a Gitlab Group.
type SubGroupStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          SubGroupObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A SubGroup is a managed resource that represents a Gitlab Group
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".status.atProvider.ID"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gitlab}
type SubGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubGroupSpec   `json:"spec"`
	Status SubGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SubGroupList contains a list of SubGroup items
type SubGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SubGroup `json:"items"`
}
