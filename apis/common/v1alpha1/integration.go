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
	ID                             *int64       `json:"id,omitempty"`
	Title                          *string      `json:"title,omitempty"`
	Slug                           *string      `json:"slug,omitempty"`
	CreatedAt                      *metav1.Time `json:"createdAt,omitempty"`
	UpdatedAt                      *metav1.Time `json:"updatedAt,omitempty"`
	Active                         *bool        `json:"active,omitempty"`
	AlertEvents                    *bool        `json:"alertEvents,omitempty"`
	CommitEvents                   *bool        `json:"commitEvents,omitempty"`
	ConfidentialIssuesEvents       *bool        `json:"confidentialIssuesEvents,omitempty"`
	ConfidentialNoteEvents         *bool        `json:"confidentialNoteEvents,omitempty"`
	DeploymentEvents               *bool        `json:"deploymentEvents,omitempty"`
	GroupConfidentialMentionEvents *bool        `json:"groupConfidentialMentionEvents,omitempty"`
	GroupMentionEvents             *bool        `json:"groupMentionEvents,omitempty"`
	IncidentEvents                 *bool        `json:"incidentEvents,omitempty"`
	IssuesEvents                   *bool        `json:"issuesEvents,omitempty"`
	JobEvents                      *bool        `json:"jobEvents,omitempty"`
	MergeRequestsEvents            *bool        `json:"mergeRequestsEvents,omitempty"`
	NoteEvents                     *bool        `json:"noteEvents,omitempty"`
	PipelineEvents                 *bool        `json:"pipelineEvents,omitempty"`
	PushEvents                     *bool        `json:"pushEvents,omitempty"`
	TagPushEvents                  *bool        `json:"tagPushEvents,omitempty"`
	VulnerabilityEvents            *bool        `json:"vulnerabilityEvents,omitempty"`
	WikiPageEvents                 *bool        `json:"wikiPageEvents,omitempty"`
	CommentOnEventEnabled          *bool        `json:"commentOnEventEnabled,omitempty"`
	Inherited                      *bool        `json:"inherited,omitempty"`
}

type CommonIntegrationParameters struct {
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
}
