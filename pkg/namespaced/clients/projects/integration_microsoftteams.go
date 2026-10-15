/*
Copyright 2026 The Crossplane Authors.

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

package projects

import (
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
)

// MicrosoftTeamsClient defines GitLab Microsoft Teams integration operations for a project.
type MicrosoftTeamsClient interface {
	GetMicrosoftTeamsService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.MicrosoftTeamsService, *gitlab.Response, error)
	SetMicrosoftTeamsService(pid any, opt *gitlab.SetMicrosoftTeamsServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.MicrosoftTeamsService, *gitlab.Response, error)
	DeleteMicrosoftTeamsService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewMicrosoftTeamsClient returns a new GitLab Services client for project Microsoft Teams operations.
func NewMicrosoftTeamsClient(cfg common.Config) MicrosoftTeamsClient {
	git := common.NewClient(cfg)
	return git.Services
}

// GenerateSetMicrosoftTeamsServiceOptions produces SetMicrosoftTeamsServiceOptions from IntegrationMicrosoftTeamsParameters.
// The webhook must be resolved from the referenced secret before calling this function.
// GitLab rejects every set request without a webhook, so it is always sent.
func GenerateSetMicrosoftTeamsServiceOptions(in *v1alpha1.IntegrationMicrosoftTeamsParameters, webhook string) *gitlab.SetMicrosoftTeamsServiceOptions {
	if in == nil {
		return &gitlab.SetMicrosoftTeamsServiceOptions{}
	}

	return &gitlab.SetMicrosoftTeamsServiceOptions{
		WebHook:                   &webhook,
		NotifyOnlyBrokenPipelines: in.NotifyOnlyBrokenPipelines,
		BranchesToBeNotified:      in.BranchesToBeNotified,
		PushEvents:                in.PushEvents,
		IssuesEvents:              in.IssuesEvents,
		ConfidentialIssuesEvents:  in.ConfidentialIssuesEvents,
		MergeRequestsEvents:       in.MergeRequestsEvents,
		TagPushEvents:             in.TagPushEvents,
		NoteEvents:                in.NoteEvents,
		ConfidentialNoteEvents:    in.ConfidentialNoteEvents,
		PipelineEvents:            in.PipelineEvents,
		WikiPageEvents:            in.WikiPageEvents,
	}
}

// GenerateIntegrationMicrosoftTeamsObservation converts gitlab.MicrosoftTeamsService to IntegrationMicrosoftTeamsObservation.
func GenerateIntegrationMicrosoftTeamsObservation(observation *gitlab.MicrosoftTeamsService) v1alpha1.IntegrationMicrosoftTeamsObservation {
	if observation == nil || observation.Properties == nil {
		return v1alpha1.IntegrationMicrosoftTeamsObservation{}
	}

	return v1alpha1.IntegrationMicrosoftTeamsObservation{
		CommonIntegrationObservation: common.GenerateCommonIntegrationObservation(&observation.Service),
		NotifyOnlyBrokenPipelines:    bool(observation.Properties.NotifyOnlyBrokenPipelines),
		BranchesToBeNotified:         observation.Properties.BranchesToBeNotified,
	}
}

// IsIntegrationMicrosoftTeamsUpToDate returns true if the desired spec matches the observed Microsoft Teams service.
// Event toggles are compared against the top-level service fields because GitLab does not return them in properties.
//
// Note: WebHook is intentionally excluded from comparison because GitLab does not return it (write-only).
func IsIntegrationMicrosoftTeamsUpToDate(spec *v1alpha1.IntegrationMicrosoftTeamsParameters, observation *gitlab.MicrosoftTeamsService) bool { //nolint:gocyclo
	if observation == nil || observation.Properties == nil {
		return false
	}

	return clients.IsComparableEqualToComparablePtr(spec.NotifyOnlyBrokenPipelines, bool(observation.Properties.NotifyOnlyBrokenPipelines)) &&
		clients.IsComparableEqualToComparablePtr(spec.BranchesToBeNotified, observation.Properties.BranchesToBeNotified) &&
		clients.IsComparableEqualToComparablePtr(spec.PushEvents, observation.PushEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.IssuesEvents, observation.IssuesEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.ConfidentialIssuesEvents, observation.ConfidentialIssuesEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.MergeRequestsEvents, observation.MergeRequestsEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.TagPushEvents, observation.TagPushEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.NoteEvents, observation.NoteEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.ConfidentialNoteEvents, observation.ConfidentialNoteEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.PipelineEvents, observation.PipelineEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.WikiPageEvents, observation.WikiPageEvents)
}

// LateInitializeIntegrationMicrosoftTeams fills nil spec fields using values from the remote Microsoft Teams service.
// It mutates the spec in place and does NOT touch write-only fields like the webhook.
// Bool fields are skipped: GitLab may return null, which client-go decodes as false.
func LateInitializeIntegrationMicrosoftTeams(in *v1alpha1.IntegrationMicrosoftTeamsParameters, svc *gitlab.MicrosoftTeamsService) {
	if in == nil || svc == nil || svc.Properties == nil {
		return
	}

	in.BranchesToBeNotified = clients.LateInitializeStringPtr(in.BranchesToBeNotified, svc.Properties.BranchesToBeNotified)
}
