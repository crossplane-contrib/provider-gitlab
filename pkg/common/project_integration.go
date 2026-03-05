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
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
)

// GenerateCommonIntegrationObservation generates a CommonIntegrationObservation from a gitlab.Service
func GenerateCommonIntegrationObservation(integration *gitlab.Service) v1alpha1.CommonIntegrationObservation {
	return v1alpha1.CommonIntegrationObservation{
		ID:                             ptr.To(integration.ID),
		Title:                          ptr.To(integration.Title),
		Slug:                           ptr.To(integration.Slug),
		CreatedAt:                      clients.TimeToMetaTime(integration.CreatedAt),
		UpdatedAt:                      clients.TimeToMetaTime(integration.UpdatedAt),
		Active:                         ptr.To(integration.Active),
		AlertEvents:                    ptr.To(integration.AlertEvents),
		CommitEvents:                   ptr.To(integration.CommitEvents),
		ConfidentialIssuesEvents:       ptr.To(integration.ConfidentialIssuesEvents),
		ConfidentialNoteEvents:         ptr.To(integration.ConfidentialNoteEvents),
		DeploymentEvents:               ptr.To(integration.DeploymentEvents),
		GroupConfidentialMentionEvents: ptr.To(integration.GroupConfidentialMentionEvents),
		GroupMentionEvents:             ptr.To(integration.GroupMentionEvents),
		IncidentEvents:                 ptr.To(integration.IncidentEvents),
		IssuesEvents:                   ptr.To(integration.IssuesEvents),
		JobEvents:                      ptr.To(integration.JobEvents),
		MergeRequestsEvents:            ptr.To(integration.MergeRequestsEvents),
		NoteEvents:                     ptr.To(integration.NoteEvents),
		PipelineEvents:                 ptr.To(integration.PipelineEvents),
		PushEvents:                     ptr.To(integration.PushEvents),
		TagPushEvents:                  ptr.To(integration.TagPushEvents),
		VulnerabilityEvents:            ptr.To(integration.VulnerabilityEvents),
		WikiPageEvents:                 ptr.To(integration.WikiPageEvents),
		CommentOnEventEnabled:          ptr.To(integration.CommentOnEventEnabled),
		Inherited:                      ptr.To(integration.Inherited),
	}
}
