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

func ptrBool(v bool) *bool       { return &v }
func ptrInt64(v int64) *int64    { return &v }
func ptrString(v string) *string { return &v }

// GenerateCommonIntegrationObservation generates a CommonIntegrationObservation from a gitlab.Service
func GenerateCommonIntegrationObservation(integration *gitlab.Service) v1alpha1.CommonIntegrationObservation {
	return v1alpha1.CommonIntegrationObservation{
		ID:                             ptrInt64(integration.ID),
		Title:                          ptrString(integration.Title),
		Slug:                           ptrString(integration.Slug),
		CreatedAt:                      clients.TimeToMetaTime(integration.CreatedAt),
		UpdatedAt:                      clients.TimeToMetaTime(integration.UpdatedAt),
		Active:                         ptrBool(integration.Active),
		AlertEvents:                    ptrBool(integration.AlertEvents),
		CommitEvents:                   ptrBool(integration.CommitEvents),
		ConfidentialIssuesEvents:       ptrBool(integration.ConfidentialIssuesEvents),
		ConfidentialNoteEvents:         ptrBool(integration.ConfidentialNoteEvents),
		DeploymentEvents:               ptrBool(integration.DeploymentEvents),
		GroupConfidentialMentionEvents: ptrBool(integration.GroupConfidentialMentionEvents),
		GroupMentionEvents:             ptrBool(integration.GroupMentionEvents),
		IncidentEvents:                 ptrBool(integration.IncidentEvents),
		IssuesEvents:                   ptrBool(integration.IssuesEvents),
		JobEvents:                      ptrBool(integration.JobEvents),
		MergeRequestsEvents:            ptrBool(integration.MergeRequestsEvents),
		NoteEvents:                     ptrBool(integration.NoteEvents),
		PipelineEvents:                 ptrBool(integration.PipelineEvents),
		PushEvents:                     ptrBool(integration.PushEvents),
		TagPushEvents:                  ptrBool(integration.TagPushEvents),
		VulnerabilityEvents:            ptrBool(integration.VulnerabilityEvents),
		WikiPageEvents:                 ptrBool(integration.WikiPageEvents),
		CommentOnEventEnabled:          ptrBool(integration.CommentOnEventEnabled),
		Inherited:                      ptrBool(integration.Inherited),
	}
}
