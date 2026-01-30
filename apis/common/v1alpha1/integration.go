/*
Copyright 2022 The Crossplane Authors.

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

// CommonIntegrationObservation represents a GitLab group or project integration observation.
//
// GitLab API docs:
// https://docs.gitlab.com/api/group_integrations/
// https://docs.gitlab.com/api/project_integrations/
type CommonIntegrationObservation struct {
	ID                             int          `json:"id"`
	Title                          string       `json:"title"`
	Slug                           string       `json:"slug"`
	CreatedAt                      *metav1.Time `json:"createdAt,omitempty"`
	UpdatedAt                      *metav1.Time `json:"updatedAt,omitempty"`
	Active                         bool         `json:"active"`
	AlertEvents                    bool         `json:"alertEvents"`
	CommitEvents                   bool         `json:"commitEvents"`
	ConfidentialIssuesEvents       bool         `json:"confidentialIssuesEvents"`
	ConfidentialNoteEvents         bool         `json:"confidentialNoteEvents"`
	DeploymentEvents               bool         `json:"deploymentEvents"`
	GroupConfidentialMentionEvents bool         `json:"groupConfidentialMentionEvents"`
	GroupMentionEvents             bool         `json:"groupMentionEvents"`
	IncidentEvents                 bool         `json:"incidentEvents"`
	IssuesEvents                   bool         `json:"issuesEvents"`
	JobEvents                      bool         `json:"jobEvents"`
	MergeRequestsEvents            bool         `json:"mergeRequestsEvents"`
	NoteEvents                     bool         `json:"noteEvents"`
	PipelineEvents                 bool         `json:"pipelineEvents"`
	PushEvents                     bool         `json:"pushEvents"`
	TagPushEvents                  bool         `json:"tagPushEvents"`
	VulnerabilityEvents            bool         `json:"vulnerabilityEvents"`
	WikiPageEvents                 bool         `json:"wikiPageEvents"`
	CommentOnEventEnabled          bool         `json:"commentOnEventEnabled"`
	Inherited                      bool         `json:"inherited"`
}

type CommonIntegrationParameters struct {
	// Send notifications for broken pipelines.
	// +optional
	NotifyOnlyBrokenPipelines *bool `json:"notify_only_broken_pipelines,omitempty"`

	// Branches to send notifications for. Valid options are all, default, protected, and default_and_protected. The default value is default.
	// +kubebuilder:validation:Enum:=all;default;protected;default_and_protected
	// +optional
	BranchesToBeNotified *string `json:"branches_to_be_notified,omitempty"`

	// 	Enable notifications for push events.
	// +optional
	PushEvents *bool `json:"push_events,omitempty"`

	// Enable notifications for issue events.
	// +optional
	IssuesEvents *bool `json:"issues_events,omitempty"`

	// Enable notifications for confidential issue events.
	// +optional
	ConfidentialIssuesEvents *bool `json:"confidential_issues_events,omitempty"`

	// Enable notifications for merge request events.
	// +optional
	MergeRequestsEvents *bool `json:"merge_requests_events,omitempty"`

	// Enable notifications for tag push events.
	// +optional
	TagPushEvents *bool `json:"tag_push_events,omitempty"`

	// Enable notifications for note events.
	// +optional
	NoteEvents *bool `json:"note_events,omitempty"`

	// Enable notifications for confidential note events.
	// +optional
	ConfidentialNoteChannel *string `json:"confidential_note_channel,omitempty"`

	// Enable notifications for pipeline events.
	// +optional
	PipelineEvents *bool `json:"pipeline_events,omitempty"`

	// Enable notifications for wiki page events.
	// +optional
	WikiPageEvents *bool `json:"wiki_page_events,omitempty"`
}
