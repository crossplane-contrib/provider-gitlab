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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// GroupProjectParameters define the desired state of a Gitlab Project
type GroupProjectParameters struct {
	// Set whether or not merge requests can be merged with skipped jobs.
	// +optional
	AllowMergeOnSkippedPipeline *bool `json:"allowMergeOnSkippedPipeline,omitempty"`

	// How many approvers should approve merge request by default.
	// To configure approval rules, see Merge request approvals API.
	// +optional
	ApprovalsBeforeMerge *int `json:"approvalsBeforeMerge,omitempty"`

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
	BuildTimeout *int `json:"buildTimeout,omitempty"`

	// One of disabled, private, or enabled.
	// +optional
	BuildsAccessLevel *AccessControlValue `json:"buildsAccessLevel,omitempty"`

	// The path to CI configuration file.
	// +optional
	CIConfigPath *string `json:"ciConfigPath,omitempty"`

	// Default number of revisions for shallow cloning.
	// +optional
	CIDefaultGitDepth *int `json:"ciDefaultGitDepth,omitempty"`

	// When a new deployment job starts, skip older deployment jobs that are still pending
	// +optional
	CIForwardDeploymentEnabled *bool `json:"ciForwardDeploymentEnabled,omitempty"`

	// Update the image cleanup policy for this project. Accepts: cadence (string), keepN (integer), olderThan (string),
	// nameRegex (string), nameRegexDelete (string), nameRegexKeep (string), enabled (boolean).
	// +optional
	ContainerExpirationPolicyAttributes *ContainerExpirationPolicyAttributes `json:"containerExpirationPolicyAttributes,omitempty"`

	// Enable container registry for this project.
	// +optional
	ContainerRegistryEnabled *bool `json:"containerRegistryEnabled,omitempty"`

	// The default branch name. Requires initializeWithReadme to be true.
	// +optional
	DefaultBranch *string `json:"defaultBranch,omitempty"`

	// Short project description.
	// +optional
	Description *string `json:"description,omitempty"`

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
	GroupWithProjectTemplatesID *int `json:"groupWithProjectTemplatesId,omitempty"`

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
	MirrorUserID *int `json:"mirrorUserId,omitempty"`

	// Namespace for the new project (defaults to the current user’s namespace).
	// +optional
	NamespaceID *int `json:"namespaceId,omitempty"`

	// NamespaceIDRef is a reference to a project to retrieve its namespaceId
	// +optional
	// +immutable
	NamespaceIDRef *xpv1.Reference `json:"namespaceIdRef,omitempty"`

	// NamespaceIDSelector selects reference to a project to retrieve its namespaceId.
	// +optional
	NamespaceIDSelector *xpv1.Selector `json:"namespaceIdSelector,omitempty"`

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

	// If true, jobs can be viewed by non-project members.
	// +optional
	PublicBuilds *bool `json:"publicBuilds,omitempty"`

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
	// that should be finally assigned to a project. Use topics instead.
	// +optional
	TagList []string `json:"tagList,omitempty"`

	// When used without useCustomTemplate, name of a built-in project template.
	// When used with useCustomTemplate, name of a custom project template.
	// +optional
	// +immutable
	TemplateName *string `json:"templateName,omitempty"`

	// When used with useCustomTemplate, project ID of a custom project template.
	// This is preferable to using templateName since templateName may be ambiguous.
	// +optional
	// +immutable
	TemplateProjectID *int `json:"templateProjectId,omitempty"`

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
}

// GroupProjectObservation is the observed state of a Project.
type GroupProjectObservation struct {
	ID                        int                        `json:"id,omitempty"`
	Archived                  bool                       `json:"archived,omitempty"`
	AvatarURL                 string                     `json:"avatarUrl,omitempty"`
	ComplianceFrameworks      []string                   `json:"complianceFrameworks,omitempty"`
	ContainerExpirationPolicy *ContainerExpirationPolicy `json:"containerExpirationPolicy,omitempty"`
	CreatedAt                 *metav1.Time               `json:"createdAt,omitempty"`
	CreatorID                 int                        `json:"creatorId,omitempty"`
	CustomAttributes          []CustomAttribute          `json:"customAttributes,omitempty"`
	EmptyRepo                 bool                       `json:"emptyRepo,omitempty"`
	ForkedFromProject         *ForkParent                `json:"forkedFromProject,omitempty"`
	ForksCount                int                        `json:"forksCount,omitempty"`
	HTTPURLToRepo             string                     `json:"httpUrlToRepo,omitempty"`
	ImportError               string                     `json:"importError,omitempty"`
	ImportStatus              string                     `json:"importStatus,omitempty"`
	IssuesEnabled             bool                       `json:"issuesEnabled,omitempty"`
	JobsEnabled               bool                       `json:"jobsEnabled,omitempty"`
	LastActivityAt            *metav1.Time               `json:"lastActivityAt,omitempty"`
	License                   *ProjectLicense            `json:"license,omitempty"`
	LicenseURL                string                     `json:"licenseUrl,omitempty"`
	Links                     *Links                     `json:"links,omitempty"`
	MarkedForDeletionAt       *metav1.Time               `json:"markedForDeletionAt,omitempty"`
	MergeRequestsEnabled      bool                       `json:"mergeRequestsEnabled,omitempty"`
	NameWithNamespace         string                     `json:"nameWithNamespace,omitempty"`
	Namespace                 *ProjectNamespace          `json:"namespace,omitempty"`
	OpenIssuesCount           int                        `json:"openIssuesCount,omitempty"`
	Owner                     *User                      `json:"owner,omitempty"`
	PathWithNamespace         string                     `json:"pathWithNamespace,omitempty"`
	Permissions               *Permissions               `json:"permissions,omitempty"`
	Public                    bool                       `json:"public,omitempty"`
	ReadmeURL                 string                     `json:"readmeUrl,omitempty"`
	RunnersToken              string                     `json:"runnersToken,omitempty"`
	SSHURLToRepo              string                     `json:"sshUrlToRepo,omitempty"`
	ServiceDeskAddress        string                     `json:"serviceDeskAddress,omitempty"`
	SharedWithGroups          []SharedWithGroups         `json:"sharedWithGroups,omitempty"`
	SnippetsEnabled           bool                       `json:"snippetsEnabled,omitempty"`
	StarCount                 int                        `json:"starCount,omitempty"`
	Statistics                *ProjectStatistics         `json:"statistics,omitempty"`
	WebURL                    string                     `json:"webUrl,omitempty"`
	WikiEnabled               bool                       `json:"wikiEnabled,omitempty"`
}

// A GroupProjectSpec defines the desired state of a Gitlab Project.
type GroupProjectSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       GroupProjectParameters `json:"forProvider"`
}

// A GroupProjectStatus represents the observed state of a Gitlab Project.
type GroupProjectStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          GroupProjectObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A GroupProject is a managed resource that represents a Gitlab Project
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="PATH WITH NAMESPACE",type="string",JSONPath=".status.atProvider.pathWithNamespace"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gitlab}
type GroupProject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GroupProjectSpec   `json:"spec"`
	Status GroupProjectStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GroupProjectList contains a list of GroupProject items
type GroupProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GroupProject `json:"items"`
}
