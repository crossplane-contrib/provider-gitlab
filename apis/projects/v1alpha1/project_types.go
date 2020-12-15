/*
Copyright 2020 The Crossplane Authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

// AccessControlValue represents an access control value within GitLab,
// used for managing access to certain project features.
//
// GitLab API docs: https://docs.gitlab.com/ce/api/projects.html
type AccessControlValue string

// List of available access control values.
//
// GitLab API docs: https://docs.gitlab.com/ce/api/projects.html
const (
	DisabledAccessControl AccessControlValue = "disabled"
	EnabledAccessControl  AccessControlValue = "enabled"
	PrivateAccessControl  AccessControlValue = "private"
	PublicAccessControl   AccessControlValue = "public"
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

// MergeMethodValue represents a project merge type within GitLab.
//
// GitLab API docs: https://docs.gitlab.com/ce/api/projects.html#project-merge-method
type MergeMethodValue string

// List of available merge type
//
// GitLab API docs: https://docs.gitlab.com/ce/api/projects.html#project-merge-method
const (
	NoFastForwardMerge MergeMethodValue = "merge"
	FastForwardMerge   MergeMethodValue = "ff"
	RebaseMerge        MergeMethodValue = "rebase_merge"
)

// UserIdentity represents a user identity.
type UserIdentity struct {
	Provider  string `json:"provider"`
	ExternUID string `json:"externUid"`
}

// User represents a GitLab user.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html
type User struct {
	ID                        int                `json:"id"`
	Username                  string             `json:"username"`
	Email                     string             `json:"email"`
	Name                      string             `json:"name"`
	State                     string             `json:"state"`
	WebURL                    string             `json:"webUrl"`
	CreatedAt                 *metav1.Time       `json:"createdAt"`
	Bio                       string             `json:"bio"`
	Location                  string             `json:"location"`
	PublicEmail               string             `json:"publicEmail"`
	Skype                     string             `json:"skype"`
	Linkedin                  string             `json:"linkedin"`
	Twitter                   string             `json:"twitter"`
	WebsiteURL                string             `json:"websiteUrl"`
	Organization              string             `json:"organization"`
	ExternUID                 string             `json:"externUid"`
	Provider                  string             `json:"provider"`
	ThemeID                   int                `json:"themeId"`
	LastActivityOn            *metav1.Time       `json:"lastActivityOn"`
	ColorSchemeID             int                `json:"colorSchemeId"`
	IsAdmin                   bool               `json:"isAdmin"`
	AvatarURL                 string             `json:"avatarUrl"`
	CanCreateGroup            bool               `json:"canCreateGroup"`
	CanCreateProject          bool               `json:"canCreateProject"`
	ProjectsLimit             int                `json:"projectsLimit"`
	CurrentSignInAt           *metav1.Time       `json:"currentSignInAt"`
	LastSignInAt              *metav1.Time       `json:"lastSignInAt"`
	ConfirmedAt               *metav1.Time       `json:"confirmedAt"`
	TwoFactorEnabled          bool               `json:"twoFactorEnabled"`
	Identities                []*UserIdentity    `json:"identities"`
	External                  bool               `json:"external"`
	PrivateProfile            bool               `json:"privateProfile"`
	SharedRunnersMinutesLimit int                `json:"sharedRunnersMinutesLimit"`
	CustomAttributes          []*CustomAttribute `json:"customAttributes"`
}

// ProjectParameters define the desired state of a Gitlab Project
type ProjectParameters struct {
	// +optional
	Path *string `json:"path,omitempty"`

	// +immutable
	// +optional
	NamespaceID *int `json:"namespaceId,omitempty"`

	// +optional
	DefaultBranch *string `json:"defaultBranch,omitempty"`

	// +optional
	Description *string `json:"description,omitempty"`

	// +optional
	IssuesAccessLevel *AccessControlValue `json:"issuesAccessLevel,omitempty"`

	// +optional
	RepositoryAccessLevel *AccessControlValue `json:"repositoryAccessLevel,omitempty"`

	// +optional
	MergeRequestsAccessLevel *AccessControlValue `json:"mergeRequestsAccessLevel,omitempty"`

	// +optional
	ForkingAccessLevel *AccessControlValue `json:"forkingAccessLevel,omitempty"`

	// +optional
	BuildsAccessLevel *AccessControlValue `json:"buildsAccessLevel,omitempty"`

	// +optional
	WikiAccessLevel *AccessControlValue `json:"wikiAccessLevel,omitempty"`

	// +optional
	SnippetsAccessLevel *AccessControlValue `json:"snippetsAccessLevel,omitempty"`

	// +optional
	PagesAccessLevel *AccessControlValue `json:"pagesAccessLevel,omitempty"`

	// +optional
	EmailsDisabled *bool `json:"emailsDisabled,omitempty"`

	// +optional
	ResolveOutdatedDiffDiscussions *bool `json:"resolveOutdatedDiffDiscussions,omitempty"`

	// +optional
	ContainerRegistryEnabled *bool `json:"containerRegistryEnabled,omitempty"`

	// +optional
	SharedRunnersEnabled *bool `json:"sharedRunnersEnabled,omitempty"`

	// +optional
	Visibility *VisibilityValue `json:"visibility,omitempty"`

	// +optional
	ImportURL *string `json:"importUrl,omitempty"`

	// +optional
	PublicBuilds *bool `json:"publicBuilds,omitempty"`

	// +optional
	OnlyAllowMergeIfPipelineSucceeds *bool `json:"onlyAllowMergeIfPipelineSucceeds,omitempty"`

	// +optional
	OnlyAllowMergeIfAllDiscussionsAreResolved *bool `json:"onlyAllowMergeIfAllDiscussionsAreResolved,omitempty"`

	// +optional
	MergeMethod *MergeMethodValue `json:"mergeMethod,omitempty"`

	// +optional
	RemoveSourceBranchAfterMerge *bool `json:"removeSourceBranchAfterMerge,omitempty"`

	// +optional
	LFSEnabled *bool `json:"lfsEnabled,omitempty"`

	// +optional
	RequestAccessEnabled *bool `json:"requestAccessEnabled,omitempty"`

	// +optional
	TagList []string `json:"tagList,omitempty"`

	// +immutable
	// +optional
	PrintingMergeRequestLinkEnabled *bool `json:"printingMergeRequestLinkEnabled,omitempty"`

	// +optional
	BuildGitStrategy *string `json:"buildGitStrategy,omitempty"`

	// +optional
	BuildTimeout *int `json:"buildTimeout,omitempty"`

	// +optional
	AutoCancelPendingPipelines *string `json:"autoCancelPendingPipelines,omitempty"`

	// +optional
	BuildCoverageRegex *string `json:"buildCoverageRegex,omitempty"`

	// +optional
	CIConfigPath *string `json:"ciConfigPath,omitempty"`

	// CIDefaultGitDepth can't be provided during project creation
	// but in can be changed afterwards with the EditProject API call
	// +optional
	CIDefaultGitDepth *int `json:"ciDefaultGitDepth,omitempty"`

	// +optional
	AutoDevopsEnabled *bool `json:"autoDevopsEnabled,omitempty"`

	// +optional
	AutoDevopsDeployStrategy *string `json:"autoDevopsDeployStrategy,omitempty"`

	// +optional
	ApprovalsBeforeMerge *int `json:"approvalsBeforeMerge,omitempty"`

	// +optional
	ExternalAuthorizationClassificationLabel *string `json:"externalAuthorizationClassificationLabel,omitempty"`

	// +optional
	Mirror *bool `json:"mirror,omitempty"`

	// MirrorUserID can't be provided during project creation
	// but in can be changed afterwards with the EditProject API call
	// +optional
	MirrorUserID *int `json:"mirrorUserId,omitempty"`

	// +optional
	MirrorTriggerBuilds *bool `json:"mirrorTriggerBuilds,omitempty"`

	// OnlyMirrorProtectedBranches can't be provided during project creation
	// but in can be changed afterwards with the EditProject API call
	// +optional
	OnlyMirrorProtectedBranches *bool `json:"onlyMirrorProtectedBranches,omitempty"`

	// MirrorOverwritesDivergedBranches can't be provided during project creation
	// but in can be changed afterwards with the EditProject API call
	// +optional
	MirrorOverwritesDivergedBranches *bool `json:"mirrorOverwritesDivergedBranches,omitempty"`

	// +immutable
	// +optional
	InitializeWithReadme *bool `json:"initializeWithReadme,omitempty"`

	// +immutable
	// +optional
	TemplateName *string `json:"templateName,omitempty"`

	// +immutable
	// +optional
	TemplateProjectID *int `json:"templateProjectId,omitempty"`

	// +immutable
	// +optional
	UseCustomTemplate *bool `json:"useCustomTemplate,omitempty"`

	// +immutable
	// +optional
	GroupWithProjectTemplatesID *int `json:"groupWithProjectTemplatesId,omitempty"`

	// +optional
	PackagesEnabled *bool `json:"packagesEnabled,omitempty"`

	// +optional
	ServiceDeskEnabled *bool `json:"serviceDeskEnabled,omitempty"`

	// +optional
	AutocloseReferencedIssues *bool `json:"autocloseReferencedIssues,omitempty"`
}

// ProjectNamespace represents a project namespace.
type ProjectNamespace struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Kind      string `json:"kind"`
	FullPath  string `json:"fullPath"`
	AvatarURL string `json:"avatarUrl"`
	WebURL    string `json:"webUrl"`
}

// Permissions represents permissions.
type Permissions struct {
	ProjectAccess *ProjectAccess `json:"projectAccess,omitempty"`
	GroupAccess   *GroupAccess   `json:"groupAccess,omitempty"`
}

// AccessLevelValue represents a permission level within GitLab.
//
// GitLab API docs: https://docs.gitlab.com/ce/permissions/permissions.html
type AccessLevelValue int

// NotificationLevelValue represents a notification level.
type NotificationLevelValue int

// ProjectAccess represents project access.
type ProjectAccess struct {
	AccessLevel       AccessLevelValue       `json:"accessLevel"`
	NotificationLevel NotificationLevelValue `json:"notificationLevel"`
}

// GroupAccess represents group access.
type GroupAccess struct {
	AccessLevel       AccessLevelValue       `json:"accessLevel"`
	NotificationLevel NotificationLevelValue `json:"notificationLevel"`
}

// ForkParent represents the parent project when this is a fork.
type ForkParent struct {
	HTTPURLToRepo     string `json:"httpUrlToRepo"`
	ID                int    `json:"id"`
	Name              string `json:"name"`
	NameWithNamespace string `json:"nameWithNamespace"`
	Path              string `json:"path"`
	PathWithNamespace string `json:"pathWithNamespace"`
	WebURL            string `json:"webUrl"`
}

// StorageStatistics represents a statistics record for a group or project.
type StorageStatistics struct {
	StorageSize      int64 `json:"storageSize"`
	RepositorySize   int64 `json:"repositorySize"`
	LfsObjectsSize   int64 `json:"lfsObjectsSize"`
	JobArtifactsSize int64 `json:"jobArtifactsSize"`
}

// ProjectStatistics represents a statistics record for a project.
type ProjectStatistics struct {
	StorageStatistics `json:",inline"`
	CommitCount       int `json:"commitCount"`
}

// Links represents a project web links for self, issues, mergeRequests,
// repoBranches, labels, events, members.
type Links struct {
	Self          string `json:"self"`
	Issues        string `json:"issues"`
	MergeRequests string `json:"mergeRequests"`
	RepoBranches  string `json:"repoBranches"`
	Labels        string `json:"labels"`
	Events        string `json:"events"`
	Members       string `json:"members"`
}

// CustomAttribute struct is used to unmarshal response to api calls.
//
// GitLab API docs: https://docs.gitlab.com/ce/api/custom_attributes.html
type CustomAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// SharedWithGroups struct used in gitlab project
type SharedWithGroups struct {
	GroupID          int    `json:"groupId,omitempty"`
	GroupName        string `json:"groupName,omitempty"`
	GroupAccessLevel int    `json:"groupAccessLevel,omitempty"`
}

// ProjectObservation is the observed state of a Project.
type ProjectObservation struct {
	ID                   int                `json:"id,omitempty"`
	Public               bool               `json:"public,omitempty"`
	SSHURLToRepo         string             `json:"sshUrlToRepo,omitempty"`
	HTTPURLToRepo        string             `json:"httpUrlToRepo,omitempty"`
	WebURL               string             `json:"webUrl,omitempty"`
	ReadmeURL            string             `json:"readmeUrl,omitempty"`
	Owner                *User              `json:"owner,omitempty"`
	PathWithNamespace    string             `json:"pathWithNamespace,omitempty"`
	IssuesEnabled        bool               `json:"issuesEnabled,omitempty"`
	OpenIssuesCount      int                `json:"openIssuesCount,omitempty"`
	MergeRequestsEnabled bool               `json:"mergeRequestsEnabled,omitempty"`
	JobsEnabled          bool               `json:"jobsEnabled,omitempty"`
	WikiEnabled          bool               `json:"wikiEnabled,omitempty"`
	SnippetsEnabled      bool               `json:"snippetsEnabled,omitempty"`
	CreatedAt            *metav1.Time       `json:"createdAt,omitempty"`
	LastActivityAt       *metav1.Time       `json:"lastActivityAt,omitempty"`
	CreatorID            int                `json:"creatorId,omitempty"`
	Namespace            *ProjectNamespace  `json:"namespace,omitempty"`
	ImportStatus         string             `json:"importStatus,omitempty"`
	ImportError          string             `json:"importError,omitempty"`
	Permissions          *Permissions       `json:"permissions,omitempty"`
	MarkedForDeletionAt  *metav1.Time       `json:"markedForDeletionAt,omitempty"`
	Archived             bool               `json:"archived,omitempty"`
	ForksCount           int                `json:"forksCount,omitempty"`
	StarCount            int                `json:"starCount,omitempty"`
	RunnersToken         string             `json:"runnersToken,omitempty"`
	ForkedFromProject    *ForkParent        `json:"forkedFromProject,omitempty"`
	SharedWithGroups     []SharedWithGroups `json:"sharedWithGroups,omitempty"`
	Statistics           *ProjectStatistics `json:"statistics,omitempty"`
	Links                *Links             `json:"links,omitempty"`
	CustomAttributes     []CustomAttribute  `json:"customAttributes,omitempty"`
	ComplianceFrameworks []string           `json:"complianceFrameworks,omitempty"`
}

// A ProjectSpec defines the desired state of a Gitlab Project.
type ProjectSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  ProjectParameters `json:"forProvider"`
}

// A ProjectStatus represents the observed state of a Gitlab Project.
type ProjectStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     ProjectObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Project is a managed resource that represents an AWS Elastic Kubernetes
// Service Project.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gitlab}
type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectSpec   `json:"spec"`
	Status ProjectStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProjectList contains a list of Project items
type ProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Project `json:"items"`
}

// Bool is a helper routine that allocates a new bool value
// to store v and returns a pointer to it.
func Bool(v bool) *bool {
	p := new(bool)
	*p = v
	return p
}

// Int is a helper routine that allocates a new int32 value
// to store v and returns a pointer to it, but unlike Int32
// its argument value is an int.
func Int(v int) *int {
	p := new(int)
	*p = v
	return p
}

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string {
	p := new(string)
	*p = v
	return p
}

// StringSlice is a helper routine that allocates a new []string value
// to store v and returns a pointer to it.
func StringSlice(v []string) *[]string {
	p := new([]string)
	*p = v
	return p
}
