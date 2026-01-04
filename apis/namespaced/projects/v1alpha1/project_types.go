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
	ExternUID string `json:"externUID"`
}

// User represents a GitLab user.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html
type User struct {
	ID                        int64              `json:"ID,omitempty"`
	Username                  string             `json:"username,omitempty"`
	Email                     string             `json:"email,omitempty"`
	Name                      string             `json:"name,omitempty"`
	State                     string             `json:"state,omitempty"`
	WebURL                    string             `json:"webURL,omitempty"`
	CreatedAt                 *metav1.Time       `json:"createdAt,omitempty"`
	Bio                       string             `json:"bio,omitempty"`
	Location                  string             `json:"location,omitempty"`
	PublicEmail               string             `json:"publicEmail,omitempty"`
	Skype                     string             `json:"skype,omitempty"`
	Linkedin                  string             `json:"linkedin,omitempty"`
	Twitter                   string             `json:"twitter,omitempty"`
	WebsiteURL                string             `json:"websiteURL,omitempty"`
	Organization              string             `json:"organization,omitempty"`
	ExternUID                 string             `json:"externUID,omitempty"`
	Provider                  string             `json:"provider,omitempty"`
	ThemeID                   int64              `json:"themeID,omitempty"`
	LastActivityOn            *metav1.Time       `json:"lastActivityOn,omitempty"`
	ColorSchemeID             int64              `json:"colorSchemeID,omitempty"`
	IsAdmin                   bool               `json:"isAdmin,omitempty"`
	AvatarURL                 string             `json:"avatarURL,omitempty"`
	CanCreateGroup            bool               `json:"canCreateGroup,omitempty"`
	CanCreateProject          bool               `json:"canCreateProject,omitempty"`
	ProjectsLimit             int64              `json:"projectsLimit,omitempty"`
	CurrentSignInAt           *metav1.Time       `json:"currentSignInAt,omitempty"`
	LastSignInAt              *metav1.Time       `json:"lastSignInAt,omitempty"`
	ConfirmedAt               *metav1.Time       `json:"confirmedAt,omitempty"`
	TwoFactorEnabled          bool               `json:"twoFactorEnabled,omitempty"`
	Identities                []*UserIdentity    `json:"identities,omitempty"`
	External                  bool               `json:"external,omitempty"`
	PrivateProfile            bool               `json:"privateProfile,omitempty"`
	SharedRunnersMinutesLimit int64              `json:"sharedRunnersMinutesLimit,omitempty"`
	CustomAttributes          []*CustomAttribute `json:"customAttributes,omitempty"`
}

// ContainerExpirationPolicy represents the container expiration policy.
type ContainerExpirationPolicy struct {
	Cadence         string       `json:"cadence"`
	KeepN           int64        `json:"keepN"`
	OlderThan       string       `json:"olderThan"`
	NameRegexDelete string       `json:"nameRegexDelete"`
	NameRegexKeep   string       `json:"nameRegexKeep"`
	Enabled         bool         `json:"enabled"`
	NextRunAt       *metav1.Time `json:"nextRunAt"`
}

// ProjectLicense represent the license for a project.
type ProjectLicense struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	Nickname  string `json:"nickname"`
	HTMLURL   string `json:"HTMLURL"`
	SourceURL string `json:"sourceURL"`
}

// ContainerExpirationPolicyAttributes represents the available container
// expiration policy attributes.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#create-project
type ContainerExpirationPolicyAttributes struct {
	Cadence         *string `json:"cadence,omitempty"`
	KeepN           *int64  `json:"keepN,omitempty"`
	OlderThan       *string `json:"olderThan,omitempty"`
	NameRegexDelete *string `json:"nameRegexDelete,omitempty"`
	NameRegexKeep   *string `json:"nameRegexKeep,omitempty"`
	Enabled         *bool   `json:"enabled,omitempty"`

	// Deprecated members
	NameRegex *string `url:"name_regex,omitempty" json:"nameRegex,omitempty"`
}

// ProjectParameters define the desired state of a Gitlab Project
type ProjectParameters struct {
	// Set whether or not merge requests can be merged with skipped jobs.
	// +optional
	AllowMergeOnSkippedPipeline *bool `json:"allowMergeOnSkippedPipeline,omitempty"`

	// How many approvers should approve merge request by default.More actions
	// To configure approval rules, see Merge request approvals API.
	// Deprecated: use Approvals Resource instead
	// +optional
	ApprovalsBeforeMerge *int64 `json:"approvalsBeforeMerge,omitempty"`

	// Auto-cancel pending pipelines. This isn’t a boolean, but enabled/disabled.
	// +optional
	AutoCancelPendingPipelines *string `json:"autoCancelPendingPipelines,omitempty"`

	// Auto Deploy strategy (continuous, manual or timedIncremental).
	// +optional
	AutoDevopsDeployStrategy *string `json:"autoDevopsDeployStrategy,omitempty"`

	// Enable Auto DevOps for this project.
	// +optional
	AutoDevopsEnabled *bool `json:"autoDevopsEnabled,omitempty"`

	// Set whether auto-closing referenced issues on default branch.
	// +optional
	AutocloseReferencedIssues *bool `json:"autocloseReferencedIssues,omitempty"`

	// Test coverage parsing.
	// +optional
	BuildCoverageRegex *string `json:"buildCoverageRegex,omitempty"`

	// The Git strategy. Defaults to fetch.
	// +optional
	BuildGitStrategy *string `json:"buildGitStrategy,omitempty"`

	// The maximum amount of time, in seconds, that a job can run.
	// +optional
	BuildTimeout *int64 `json:"buildTimeout,omitempty"`

	// One of disabled, private, or enabled.
	// +optional
	BuildsAccessLevel *AccessControlValue `json:"buildsAccessLevel,omitempty"`

	// The path to CI configuration file.
	// +optional
	CIConfigPath *string `json:"ciConfigPath,omitempty"`

	// Default number of revisions for shallow cloning.
	// +optional
	CIDefaultGitDepth *int64 `json:"ciDefaultGitDepth,omitempty"`

	// When a new deployment job starts, skip older deployment jobs that are still pending
	// +optional
	CIForwardDeploymentEnabled *bool `json:"ciForwardDeploymentEnabled,omitempty"`

	// Update the image cleanup policy for this project. Accepts: cadence (string), keepN (integer), olderThan (string),
	// nameRegex (string), nameRegexDelete (string), nameRegexKeep (string), enabled (boolean).
	// +optional
	ContainerExpirationPolicyAttributes *ContainerExpirationPolicyAttributes `json:"containerExpirationPolicyAttributes,omitempty"`

	// Enable container registry for this project.
	//
	// Deprecated: Use ContainerRegistryAccessLevel.
	//
	// +optional
	ContainerRegistryEnabled *bool `json:"containerRegistryEnabled,omitempty"`

	// Set visibility of container registry.
	// +optional
	ContainerRegistryAccessLevel *AccessControlValue `json:"containerRegistryAccessLevel,omitempty"`

	// The default branch name. Requires initializeWithReadme to be true.
	// +optional
	DefaultBranch *string `json:"defaultBranch,omitempty"`

	// Short project description.
	// +optional
	Description *string `json:"description,omitempty"`

	// Name is the human-readable name of the project.
	// If set, it overrides metadata.name.
	// +kubebuilder:validation:MaxLength:=255
	// +optional
	Name *string `json:"name,omitempty"`

	// Disable email notifications.
	// +optional
	EmailsDisabled *bool `json:"emailsDisabled,omitempty"`

	// The classification label for the project.
	// +optional
	ExternalAuthorizationClassificationLabel *string `json:"externalAuthorizationClassificationLabel,omitempty"`

	// One of disabled, private, or enabled.
	// +optional
	ForkingAccessLevel *AccessControlValue `json:"forkingAccessLevel,omitempty"`

	// For group-level custom templates, specifies ID of group from which all the custom project templates are sourced.
	// Leave empty for instance-level templates. Requires useCustomTemplate to be true.
	// +optional
	// +immutable
	GroupWithProjectTemplatesID *int64 `json:"groupWithProjectTemplatesId,omitempty"`

	// URL to import repository from.
	// +optional
	ImportURL *string `json:"importUrl,omitempty"`

	// false by default.
	// +optional
	// +immutable
	InitializeWithReadme *bool `json:"initializeWithReadme,omitempty"`

	// One of disabled, private, or enabled.
	// +optional
	IssuesAccessLevel *AccessControlValue `json:"issuesAccessLevel,omitempty"`

	// Default description for Issues. Description is parsed with GitLab Flavored Markdown.
	// See Templates for issues and merge requests.
	// +optional
	IssuesTemplate *string `json:"issuesTemplate,omitempty"`

	// Enable LFS.
	// +optional
	LFSEnabled *bool `json:"lfsEnabled,omitempty"`

	// Set the merge method used.
	// +optional
	MergeMethod *MergeMethodValue `json:"mergeMethod,omitempty"`

	// One of disabled, private, or enabled.
	// +optional
	MergeRequestsAccessLevel *AccessControlValue `json:"mergeRequestsAccessLevel,omitempty"`

	// Default description for Merge Requests. Description is parsed with GitLab Flavored Markdown.
	// See Templates for issues and merge requests.
	// +optional
	MergeRequestsTemplate *string `json:"mergeRequestsTemplate,omitempty"`

	// Enables pull mirroring in a project.
	// +optional
	Mirror *bool `json:"mirror,omitempty"`

	// Pull mirror overwrites diverged branches.
	// +optional
	MirrorOverwritesDivergedBranches *bool `json:"mirrorOverwritesDivergedBranches,omitempty"`

	// Pull mirroring triggers builds.
	// +optional
	MirrorTriggerBuilds *bool `json:"mirrorTriggerBuilds,omitempty"`

	// User responsible for all the activity surrounding a pull mirror event. (admins only)
	// +optional
	MirrorUserID *int64 `json:"mirrorUserId,omitempty"`

	// Namespace for the new project (defaults to the current user’s namespace).
	// +optional
	NamespaceID *int64 `json:"namespaceId,omitempty"`

	// NamespaceIDRef is a reference to a project to retrieve its namespaceId
	// +optional
	// +immutable
	NamespaceIDRef *xpv1.NamespacedReference `json:"namespaceIdRef,omitempty"`

	// NamespaceIDSelector selects reference to a project to retrieve its namespaceId.
	// +optional
	NamespaceIDSelector *xpv1.NamespacedSelector `json:"namespaceIdSelector,omitempty"`

	// Set whether merge requests can only be merged when all the discussions are resolved.
	// +optional
	OnlyAllowMergeIfAllDiscussionsAreResolved *bool `json:"onlyAllowMergeIfAllDiscussionsAreResolved,omitempty"`

	// Set whether merge requests can only be merged with successful jobs.
	// +optional
	OnlyAllowMergeIfPipelineSucceeds *bool `json:"onlyAllowMergeIfPipelineSucceeds,omitempty"`

	// Only mirror protected branches.
	// +optional
	OnlyMirrorProtectedBranches *bool `json:"onlyMirrorProtectedBranches,omitempty"`

	// One of disabled, private, or enabled.
	// +optional
	OperationsAccessLevel *AccessControlValue `json:"operationsAccessLevel,omitempty"`

	// Enable or disable packages repository feature.
	// +optional
	PackagesEnabled *bool `json:"packagesEnabled,omitempty"`

	// One of disabled, private, enabled, or public.
	// +optional
	PagesAccessLevel *AccessControlValue `json:"pagesAccessLevel,omitempty"`

	// Repository name for new project.
	// Generated based on name if not provided (generated as lowercase with dashes).
	// +optional
	Path *string `json:"path,omitempty"`

	// Show link to create/view merge request when pushing from the command line.
	// +optional
	// +immutable
	PrintingMergeRequestLinkEnabled *bool `json:"printingMergeRequestLinkEnabled,omitempty"`

	// Deprecated: Use PublicJobs instead. This field will be removed in a future version.
	// If true, jobs can be viewed by non-project members.
	// +optional
	PublicBuilds *bool `json:"publicBuilds,omitempty"`

	// If true, jobs can be viewed by non-project members.
	// This replaces the deprecated PublicBuilds field.
	// If both PublicBuilds and PublicJobs are set, PublicJobs takes precedence.
	// +optional
	PublicJobs *bool `json:"publicJobs,omitempty"`

	// PushRule specifies are pre-receive Git hooks you can enable in a
	// user-friendly interface.
	// Push rules give you more control over what can and can’t be pushed to
	// your repository.
	// +optional
	PushRules *PushRules `json:"pushRules,omitempty"`

	// Enable Delete source branch option by default for all new merge requests.
	// +optional
	RemoveSourceBranchAfterMerge *bool `json:"removeSourceBranchAfterMerge,omitempty"`

	// One of disabled, private, or enabled.
	// +optional
	RepositoryAccessLevel *AccessControlValue `json:"repositoryAccessLevel,omitempty"`

	// Allow users to request member access.
	// +optional
	RequestAccessEnabled *bool `json:"requestAccessEnabled,omitempty"`

	// Automatically resolve merge request diffs discussions on lines changed with a push.
	// +optional
	ResolveOutdatedDiffDiscussions *bool `json:"resolveOutdatedDiffDiscussions,omitempty"`

	// Enable or disable Service Desk feature.
	// +optional
	ServiceDeskEnabled *bool `json:"serviceDeskEnabled,omitempty"`

	// Enable shared runners for this project.
	// +optional
	SharedRunnersEnabled *bool `json:"sharedRunnersEnabled,omitempty"`

	// One of disabled, private, or enabled.
	// +optional
	SnippetsAccessLevel *AccessControlValue `json:"snippetsAccessLevel,omitempty"`

	// The commit message used to apply merge request suggestions.
	// +optional
	SuggestionCommitMessage *string `json:"suggestionCommitMessage,omitempty"`

	// The list of tags for a project; put array of tags,
	// that should be finally assigned to a project.
	//
	// Deprecated: Use topics instead.
	//
	// +optional
	TagList []string `json:"tagList,omitempty"`

	// The list of topics for the project;
	// +optional
	Topics []string `json:"topics,omitempty"`

	// When used without useCustomTemplate, name of a built-in project template.
	// When used with useCustomTemplate, name of a custom project template.
	// +optional
	// +immutable
	TemplateName *string `json:"templateName,omitempty"`

	// When used with useCustomTemplate, project ID of a custom project template.
	// This is preferable to using templateName since templateName may be ambiguous.
	// +optional
	// +immutable
	TemplateProjectID *int64 `json:"templateProjectId,omitempty"`

	// Use either custom instance or group (with groupWithProjectTemplatesId) project template.
	// +optional
	// +immutable
	UseCustomTemplate *bool `json:"useCustomTemplate,omitempty"`

	// See project visibility level.
	// +optional
	Visibility *VisibilityValue `json:"visibility,omitempty"`

	// One of disabled, private, or enabled.
	// +optional
	WikiAccessLevel *AccessControlValue `json:"wikiAccessLevel,omitempty"`

	// Force the immediate deletion of the project when removed. In GitLab Premium and Ultimate a project is by default
	// just marked for deletion and removed permanently after seven days. Defaults to false.
	// +optional
	PermanentlyRemove *bool `json:"permanentlyRemove,omitempty"`

	// RemoveFinalizerOnPendingDeletion specifies wether the finalizer of this
	// object should be removed in case the Kubernetes object and
	// the external Gitlab project are marked for pending deletion.
	RemoveFinalizerOnPendingDeletion *bool `json:"removeFinalizerOnPendingDeletion,omitempty"`
}

type PushRules struct {
	// All commit author emails must match this regular expression.
	AuthorEmailRegex *string `json:"authorEmailRegex,omitempty"`

	// All branch names must match this regular expression.
	BranchNameRegex *string `json:"branchNameRegex,omitempty"`

	// Users can only push commits to this repository if the committer email is
	// one of their own verified emails.
	CommitCommitterCheck *bool `json:"commitCommitterCheck,omitempty"`

	// Users can only push commits to this repository if the commit author name
	// is consistent with their GitLab account name.
	CommitCommitterNameCheck *bool `json:"commitCommitterNameCheck,omitempty"`

	// No commit message is allowed to match this regular expression.
	CommitMessageNegativeRegex *string `json:"commitMessageNegativeRegex,omitempty"`

	// All commit messages must match this regular expression.
	CommitMessageRegex *string `json:"commitMessageRegex,omitempty"`

	// Deny deleting a tag.
	DenyDeleteTag *bool `json:"denyDeleteTag,omitempty"`

	// All committed filenames must not match this regular expression.
	FileNameRegex *string `json:"fileNameRegex,omitempty"`

	// Maximum file size (MB).
	MaxFileSize *int64 `json:"maxFileSize,omitempty"`

	// Restrict commits by author (email) to existing GitLab users.
	MemberCheck *bool `json:"memberCheck,omitempty"`

	// GitLab rejects any files that are likely to contain secrets.
	PreventSecrets *bool `json:"preventSecrets,omitempty"`

	// Reject commit when it’s not signed.
	RejectUnsignedCommits *bool `json:"rejectUnsignedCommits,omitempty"`

	// Reject commit when it’s not DCO certified.
	RejectNonDCOCommits *bool `json:"rejectNonDcoCommits,omitempty"`
}

// ProjectNamespace represents a project namespace.
type ProjectNamespace struct {
	ID        int64  `json:"ID"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Kind      string `json:"kind"`
	FullPath  string `json:"fullPath"`
	AvatarURL string `json:"avatarURL"`
	WebURL    string `json:"webURL"`
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
	HTTPURLToRepo     string `json:"HTTPURLToRepo"`
	ID                int64  `json:"ID"`
	Name              string `json:"name"`
	NameWithNamespace string `json:"nameWithNamespace"`
	Path              string `json:"path"`
	PathWithNamespace string `json:"pathWithNamespace"`
	WebURL            string `json:"webURL"`
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
	CommitCount       int64 `json:"commitCount"`
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
	GroupID          int64  `json:"groupID,omitempty"`
	GroupName        string `json:"groupName,omitempty"`
	GroupAccessLevel int64  `json:"groupAccessLevel,omitempty"`
}

// ProjectObservation is the observed state of a Project.
type ProjectObservation struct {
	ID                        int64                      `json:"id,omitempty"`
	Archived                  bool                       `json:"archived,omitempty"`
	AvatarURL                 string                     `json:"avatarUrl,omitempty"`
	ComplianceFrameworks      []string                   `json:"complianceFrameworks,omitempty"`
	ContainerExpirationPolicy *ContainerExpirationPolicy `json:"containerExpirationPolicy,omitempty"`
	CreatedAt                 *metav1.Time               `json:"createdAt,omitempty"`
	CreatorID                 int64                      `json:"creatorId,omitempty"`
	CustomAttributes          []CustomAttribute          `json:"customAttributes,omitempty"`
	EmptyRepo                 bool                       `json:"emptyRepo,omitempty"`
	ForkedFromProject         *ForkParent                `json:"forkedFromProject,omitempty"`
	ForksCount                int64                      `json:"forksCount,omitempty"`
	HTTPURLToRepo             string                     `json:"httpUrlToRepo,omitempty"`
	ImportError               string                     `json:"importError,omitempty"`
	ImportStatus              string                     `json:"importStatus,omitempty"`
	IssuesEnabled             bool                       `json:"issuesEnabled,omitempty"`
	JobsEnabled               bool                       `json:"jobsEnabled,omitempty"`
	IssuesAccessLevel         AccessControlValue         `json:"issuesAccessLevel,omitempty"`
	BuildsAccessLevel         AccessControlValue         `json:"buildsAccessLevel,omitempty"`
	LastActivityAt            *metav1.Time               `json:"lastActivityAt,omitempty"`
	License                   *ProjectLicense            `json:"license,omitempty"`
	LicenseURL                string                     `json:"licenseUrl,omitempty"`
	Links                     *Links                     `json:"links,omitempty"`
	MarkedForDeletionAt       *metav1.Time               `json:"markedForDeletionAt,omitempty"`
	MergeRequestsEnabled      bool                       `json:"mergeRequestsEnabled,omitempty"`
	MarkedForDeletionOn       *metav1.Time               `json:"markedForDeletionOn,omitempty"`
	MergeRequestsAccessLevel  AccessControlValue         `json:"mergeRequestsAccessLevel,omitempty"`
	NameWithNamespace         string                     `json:"nameWithNamespace,omitempty"`
	Namespace                 *ProjectNamespace          `json:"namespace,omitempty"`
	OpenIssuesCount           int64                      `json:"openIssuesCount,omitempty"`
	Owner                     *User                      `json:"owner,omitempty"`
	PathWithNamespace         string                     `json:"pathWithNamespace,omitempty"`
	Permissions               *Permissions               `json:"permissions,omitempty"`
	Public                    bool                       `json:"public,omitempty"`
	ReadmeURL                 string                     `json:"readmeUrl,omitempty"`
	SSHURLToRepo              string                     `json:"sshUrlToRepo,omitempty"`
	ServiceDeskAddress        string                     `json:"serviceDeskAddress,omitempty"`
	SharedWithGroups          []SharedWithGroups         `json:"sharedWithGroups,omitempty"`
	SnippetsEnabled           bool                       `json:"snippetsEnabled,omitempty"`
	SnippetsAccessLevel       AccessControlValue         `json:"snippetsAccessLevel,omitempty"`
	StarCount                 int64                      `json:"starCount,omitempty"`
	Statistics                *ProjectStatistics         `json:"statistics,omitempty"`
	WebURL                    string                     `json:"webUrl,omitempty"`
	WikiEnabled               bool                       `json:"wikiEnabled,omitempty"`
	WikiAccessLevel           AccessControlValue         `json:"wikiAccessLevel,omitempty"`
}

// A ProjectSpec defines the desired state of a Gitlab Project.
type ProjectSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`
	ForProvider              ProjectParameters `json:"forProvider"`
}

// A ProjectStatus represents the observed state of a Gitlab Project.
type ProjectStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ProjectObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Project is a managed resource that represents a Gitlab Project
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="PATH WITH NAMESPACE",type="string",JSONPath=".status.atProvider.pathWithNamespace"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}
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
