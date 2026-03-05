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

	"github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
)

// IntegrationMattermostParameters defines the desired state of a GitLab Project Mattermost Integration.
type IntegrationMattermostParameters struct {
	// ProjectID is the ID of the project.
	// +optional
	// +immutable
	ProjectID *int64 `json:"projectId,omitempty"`

	// ProjectIDRef is a reference to a project to retrieve its projectId
	// +optional
	// +immutable
	ProjectIDRef *xpv1.NamespacedReference `json:"projectIdRef,omitempty"`

	// ProjectIDSelector selects reference to a project to retrieve its projectId.
	// +optional
	ProjectIDSelector *xpv1.NamespacedSelector `json:"projectIdSelector,omitempty"`

	// Mattermost notifications webhook (for example, http://mattermost.example.com/hooks/...).
	// WARNING: This field is NOT reconciled as the GitLab API does not return it as it is a write-only field.
	// +required
	WebHook string `json:"webhook,omitempty"`

	// Mattermost notifications username.
	// +optional
	Username *string `json:"username,omitempty"`

	// Default channel to use if no other channel is configured.
	// +optional
	Channel *string `json:"channel,omitempty"`

	// Send notifications for broken pipelines.
	// +optional
	NotifyOnlyBrokenPipelines *bool `json:"notifyOnlyBrokenPipelines,omitempty"`

	// Branches to send notifications for. Valid options are all, default, protected, and default_and_protected. The default value is default.
	// +kubebuilder:validation:Enum:=all;default;protected;default_and_protected
	// +optional
	BranchesToBeNotified *string `json:"branchesToBeNotified,omitempty"`

	// 	Enable notifications for push events.
	// +optional
	PushEvents *bool `json:"pushEvents,omitempty"`

	// Enable notifications for issue events.
	// +optional
	IssuesEvents *bool `json:"issuesEvents,omitempty"`

	// Enable notifications for confidential issue events.
	// +optional
	ConfidentialIssuesEvents *bool `json:"confidentialIssuesEvents,omitempty"`

	// Enable notifications for merge request events.
	// +optional
	MergeRequestsEvents *bool `json:"mergeRequestsEvents,omitempty"`

	// Enable notifications for tag push events.
	// +optional
	TagPushEvents *bool `json:"tagPushEvents,omitempty"`

	// Enable notifications for note events.
	// +optional
	NoteEvents *bool `json:"noteEvents,omitempty"`

	// Enable notifications for confidential note events.
	// +optional
	ConfidentialNoteChannel *string `json:"confidentialNoteChannel,omitempty"`

	// Enable notifications for pipeline events.
	// +optional
	PipelineEvents *bool `json:"pipelineEvents,omitempty"`

	// Enable notifications for wiki page events.
	// +optional
	WikiPageEvents *bool `json:"wikiPageEvents,omitempty"`

	// The name of the channel to receive notifications for push events.
	// +optional
	PushChannel *string `json:"pushChannel,omitempty"`

	// The name of the channel to receive notifications for issue events.
	// +optional
	IssueChannel *string `json:"issueChannel,omitempty"`

	// The name of the channel to receive notifications for confidential issue events.
	// +optional
	ConfidentialIssueChannel *string `json:"confidentialIssueChannel,omitempty"`

	// The name of the channel to receive notifications for merge request events.
	// +optional
	MergeRequestChannel *string `json:"mergeRequestChannel,omitempty"`

	// The name of the channel to receive notifications for note events.
	// +optional
	NoteChannel *string `json:"noteChannel,omitempty"`

	// The name of the channel to receive notifications for confidential note events.
	// +optional
	ConfidentialNoteEvents *bool `json:"confidentialNoteEvents,omitempty"`

	// The name of the channel to receive notifications for tag push events.
	// +optional
	TagPushChannel *string `json:"tagPushChannel,omitempty"`

	// The name of the channel to receive notifications for pipeline events.
	// +optional
	PipelineChannel *string `json:"pipelineChannel,omitempty"`

	// The name of the channel to receive notifications for wiki page events.
	// +optional
	WikiPageChannel *string `json:"wikiPageChannel,omitempty"`
}

// IntegrationMattermostObservation represents the observed state of a GitLab Project Mattermost Integration.
type IntegrationMattermostObservation struct {
	v1alpha1.CommonIntegrationObservation `json:",inline"`
	// Mattermost notifications webhook (for example, http://mattermost.example.com/hooks/...).
	// This field is not returned by GitLab API for security reasons. So it will always be empty in the observation.
	WebHook string `json:"webhook"`
	// Mattermost notifications username.
	Username string `json:"username"`
	// Default channel to use if no other channel is configured.
	Channel string `json:"channel"`
	// Send notifications for broken pipelines.
	NotifyOnlyBrokenPipelines bool `json:"notifyOnlyBrokenPipelines"`
	// Branches to send notifications for. Valid options are all, default, protected, and default_and_protected. The default value is default.
	BranchesToBeNotified string `json:"branchesToBeNotified"`
	// Channel to use for confidential issues events.
	ConfidentialIssueChannel string `json:"confidentialIssueChannel"`
	// Channel to use for confidential notes events.
	ConfidentialNoteChannel string `json:"confidentialNoteChannel"`
	// Channel to use for issue events.
	IssueChannel string `json:"issueChannel"`
	// Channel to use for merge request events.
	MergeRequestChannel string `json:"mergeRequestChannel"`
	// Channel to use for note events.
	NoteChannel string `json:"noteChannel"`
	// Channel to use for tag push events.
	TagPushChannel string `json:"tagPushChannel"`
	// Channel to use for pipeline events.
	PipelineChannel string `json:"pipelineChannel"`
	// Channel to use for push events.
	PushChannel string `json:"pushChannel"`
	// Channel to use for vulnerability events.
	VulnerabilityChannel string `json:"vulnerabilityChannel"`
	// Channel to use for wiki page events.
	WikiPageChannel string `json:"wikiPageChannel"`
}

// A IntegrationMattermostSpec defines the desired state of a GitLab Project Mattermost Integration.
type IntegrationMattermostSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`
	// ForProvider represents the desired state of the mattermost integration
	ForProvider IntegrationMattermostParameters `json:"forProvider"`
}

// A IntegrationMattermostStatus represents the observed state of a GitLab Project Mattermost Integration.
type IntegrationMattermostStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	// AtProvider represents the observed state of the mattermost integration
	AtProvider IntegrationMattermostObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A IntegrationMattermost is a managed resource that represents a GitLab Project Mattermost Integration
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="BRANCH",type="string",JSONPath=".spec.forProvider.branchName"
// +kubebuilder:printcolumn:name="PROJECT",type="string",JSONPath=".spec.forProvider.projectId"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}
type IntegrationMattermost struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              IntegrationMattermostSpec   `json:"spec"`
	Status            IntegrationMattermostStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IntegrationMattermostList contains a list of IntegrationMattermost items
type IntegrationMattermostList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IntegrationMattermost `json:"items"`
}
