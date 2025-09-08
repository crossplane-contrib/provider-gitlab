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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// AccessLevelValue represents a permission level within GitLab.
//
// GitLab API docs: https://docs.gitlab.com/ce/permissions/permissions.html
type AccessLevelValue int

// List of available access levels
//
// GitLab API docs: https://docs.gitlab.com/ce/permissions/permissions.html
const (
	NoPermissions            AccessLevelValue = 0
	MinimalAccessPermissions AccessLevelValue = 5
	GuestPermissions         AccessLevelValue = 10
	ReporterPermissions      AccessLevelValue = 20
	DeveloperPermissions     AccessLevelValue = 30
	MaintainerPermissions    AccessLevelValue = 40
	OwnerPermissions         AccessLevelValue = 50

	// These are deprecated and should be removed in a future version
	MasterPermissions AccessLevelValue = 40
	OwnerPermission   AccessLevelValue = 50
)

// VisibilityValue represents a visibility level within GitLab.
//
// GitLab API docs: https://docs.gitlab.com/ce/api/
type VisibilityValue string

// List of available visibility levels.
//
// GitLab API docs: https://docs.gitlab.com/ce/api/
const (
	PrivateVisibility  VisibilityValue = "private"
	InternalVisibility VisibilityValue = "internal"
	PublicVisibility   VisibilityValue = "public"
)

// ProjectCreationLevelValue represents a project creation level within GitLab.
//
// GitLab API docs: https://docs.gitlab.com/ce/api/
type ProjectCreationLevelValue string

// List of available project creation levels.
//
// GitLab API docs: https://docs.gitlab.com/ce/api/
const (
	NoOneProjectCreation      ProjectCreationLevelValue = "noone"
	MaintainerProjectCreation ProjectCreationLevelValue = "maintainer"
	DeveloperProjectCreation  ProjectCreationLevelValue = "developer"
)

// SubGroupCreationLevelValue represents a sub group creation level within GitLab.
//
// GitLab API docs: https://docs.gitlab.com/ce/api/
type SubGroupCreationLevelValue string

// List of available sub group creation levels.
//
// GitLab API docs: https://docs.gitlab.com/ce/api/
const (
	OwnerSubGroupCreationLevelValue      SubGroupCreationLevelValue = "owner"
	MaintainerSubGroupCreationLevelValue SubGroupCreationLevelValue = "maintainer"
)

// SharedWithGroups represents a GitLab Shared with groups.
// At least one of the fields [GroupID, GroupIDRef, GroupIDSelector] must be set.
type SharedWithGroups struct {
	// The ID of the group to share with.
	// +kubebuilder:example=10
	// +optional
	GroupID *int `json:"groupId,omitempty"`

	// The role (access_level) to grant the group
	// https://docs.gitlab.com/ee/api/members.html#roles
	// +kubebuilder:example=30
	// +required
	// +immutable
	GroupAccessLevel int `json:"groupAccessLevel"`

	// Share expiration date in ISO 8601 format: 2016-09-26
	// +optional
	// +immutable
	ExpiresAt *metav1.Time `json:"expiresAt,omitempty"`
}

// GroupParameters define the desired state of a Gitlab Project
type GroupParameters struct {
	// The path of the group.
	// +kubebuilder:example="example-group-path"
	// +immutable
	Path string `json:"path"`

	// The group’s description.
	// +kubebuilder:example="example group description"
	// +optional
	Description *string `json:"description,omitempty"`

	// Name is the human-readable name of the group.
	// If set, it overrides metadata.name.
	// +kubebuilder:validation:MaxLength:=255
	// +kubebuilder:example="Example Group"
	// +optional
	Name *string `json:"name,omitempty"`

	// Prevent adding new members to project membership within this group.
	// +optional
	MembershipLock *bool `json:"membershipLock,omitempty"`

	// The group’s visibility. Can be private, internal, or public.
	// +kubebuilder:example="internal"
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
	//
	// Deprecated: Use emailsEnabled instead.
	// +optional
	EmailsDisabled *bool `json:"emailsDisabled,omitempty"`

	// Enable email notifications.
	// +optional
	EmailsEnabled *bool `json:"emailsEnabled,omitempty"`

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
	// +kubebuilder:example=9
	// +optional
	ParentID *int `json:"parentId,omitempty"`

	// Pipeline minutes quota for this group (included in plan).
	// Can be nil (default; inherit system default), 0 (unlimited) or > 0.
	// +optional
	SharedRunnersMinutesLimit *int `json:"sharedRunnersMinutesLimit,omitempty"`

	// Extra pipeline minutes quota for this group (purchased in addition to the minutes included in the plan).
	// +optional
	ExtraSharedRunnersMinutesLimit *int `json:"extraSharedRunnersMinutesLimit,omitempty"`

	// Force the immediate deletion of the group when removed. In GitLab Premium and Ultimate a group is by default
	// just marked for deletion and removed permanently after seven days. Defaults to false.
	// +optional
	PermanentlyRemove *bool `json:"permanentlyRemove,omitempty"`

	// RemoveFinalizerOnPendingDeletion specifies wether the finalizer of this
	// object should be removed in case the Kubernetes object and
	// the external Gitlab group are marked for pending deletion.
	RemoveFinalizerOnPendingDeletion *bool `json:"removeFinalizerOnPendingDeletion,omitempty"`

	// Full path of group to delete permanently. Only required if PermanentlyRemove is set to true.
	// GitLab Premium and Ultimate only.
	// +optional
	FullPathToRemove *string `json:"fullPathToRemove,omitempty"`
}

// LDAPGroupLink represents a GitLab LDAP group link.
//
// GitLab API docs: https://docs.gitlab.com/ce/api/groups.html#ldap-group-links
type LDAPGroupLink struct {
	CN          string           `json:"cn"`
	GroupAccess AccessLevelValue `json:"groupAccess"`
	Provider    string           `json:"provider"`
}

// CustomAttribute struct is used to unmarshal response to api calls.
//
// GitLab API docs: https://docs.gitlab.com/ce/api/custom_attributes.html
type CustomAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// StorageStatistics represents a statistics record for a group or project.
type StorageStatistics struct {
	StorageSize      int64 `json:"storageSize"`
	RepositorySize   int64 `json:"repositorySize"`
	LfsObjectsSize   int64 `json:"lfsObjectsSize"`
	JobArtifactsSize int64 `json:"jobArtifactsSize"`
}

// GroupObservation is the observed state of a Group.
type GroupObservation struct {
	ID                  *int                          `json:"id,omitempty"`
	AvatarURL           *string                       `json:"avatarUrl,omitempty"`
	WebURL              *string                       `json:"webUrl,omitempty"`
	FullName            *string                       `json:"fullName,omitempty"`
	FullPath            *string                       `json:"fullPath,omitempty"`
	Statistics          *StorageStatistics            `json:"statistics,omitempty"`
	CustomAttributes    []CustomAttribute             `json:"customAttributes,omitempty"`
	LDAPCN              *string                       `json:"ldapCn,omitempty"`
	LDAPAccess          *AccessLevelValue             `json:"ldapAccess,omitempty"`
	LDAPGroupLinks      []LDAPGroupLink               `json:"ldapGroupLinks,omitempty"`
	MarkedForDeletionOn *metav1.Time                  `json:"markedForDeletionOn,omitempty"`
	CreatedAt           *metav1.Time                  `json:"createdAt,omitempty"`
	SharedWithGroups    []SharedWithGroupsObservation `json:"sharedWithGroups,omitempty"`
}

// SharedWithGroupsObservation is the observed state of a SharedWithGroups.
type SharedWithGroupsObservation struct {
	GroupID          *int         `json:"groupId,omitempty"`
	GroupName        *string      `json:"groupName,omitempty"`
	GroupFullPath    *string      `json:"groupFullPath,omitempty"`
	GroupAccessLevel *int         `json:"groupAccessLevel,omitempty"`
	ExpiresAt        *metav1.Time `json:"expiresAt,omitempty"`
}
