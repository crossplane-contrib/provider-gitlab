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

package common

import (
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
)

// GenerateCommonIntegrationObservation generates a CommonIntegrationObservation from a gitlab.Service
func GenerateCommonIntegrationObservation(integration *gitlab.Service) v1alpha1.CommonIntegrationObservation {
	return v1alpha1.CommonIntegrationObservation{
		ID:                             integration.ID,
		Title:                          integration.Title,
		Slug:                           integration.Slug,
		CreatedAt:                      clients.TimeToMetaTime(integration.CreatedAt),
		UpdatedAt:                      clients.TimeToMetaTime(integration.UpdatedAt),
		Active:                         integration.Active,
		AlertEvents:                    integration.AlertEvents,
		CommitEvents:                   integration.CommitEvents,
		ConfidentialIssuesEvents:       integration.ConfidentialIssuesEvents,
		ConfidentialNoteEvents:         integration.ConfidentialNoteEvents,
		DeploymentEvents:               integration.DeploymentEvents,
		GroupConfidentialMentionEvents: integration.GroupConfidentialMentionEvents,
		GroupMentionEvents:             integration.GroupMentionEvents,
		IncidentEvents:                 integration.IncidentEvents,
		IssuesEvents:                   integration.IssuesEvents,
		JobEvents:                      integration.JobEvents,
		MergeRequestsEvents:            integration.MergeRequestsEvents,
		NoteEvents:                     integration.NoteEvents,
		PipelineEvents:                 integration.PipelineEvents,
		PushEvents:                     integration.PushEvents,
		TagPushEvents:                  integration.TagPushEvents,
		VulnerabilityEvents:            integration.VulnerabilityEvents,
		WikiPageEvents:                 integration.WikiPageEvents,
		CommentOnEventEnabled:          integration.CommentOnEventEnabled,
		Inherited:                      integration.Inherited,
	}
}
