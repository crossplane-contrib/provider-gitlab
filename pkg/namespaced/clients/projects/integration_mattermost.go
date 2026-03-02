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

package projects

import (
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
)

// MattermostClient defines GitLab Mattermost integration operations.
type MattermostClient interface {
	GetMattermostService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.MattermostService, *gitlab.Response, error)
	SetMattermostService(pid any, opt *gitlab.SetMattermostServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.MattermostService, *gitlab.Response, error)
	DeleteMattermostService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewMattermostClient returns a new GitLab Services client.
func NewMattermostClient(cfg common.Config) MattermostClient {
	git := common.NewClient(cfg)
	return git.Services
}

// GenerateSetMattermostServiceOptions produces SetMattermostServiceOptions from IntegrationMattermostParameters.
func GenerateSetMattermostServiceOptions(in *v1alpha1.IntegrationMattermostParameters) *gitlab.SetMattermostServiceOptions {
	if in == nil {
		return &gitlab.SetMattermostServiceOptions{}
	}

	opts := gitlab.SetMattermostServiceOptions{
		Username:                  in.Username,
		Channel:                   in.Channel,
		NotifyOnlyBrokenPipelines: in.NotifyOnlyBrokenPipelines,
		BranchesToBeNotified:      in.BranchesToBeNotified,

		PushEvents:               in.PushEvents,
		IssuesEvents:             in.IssuesEvents,
		ConfidentialIssuesEvents: in.ConfidentialIssuesEvents,
		MergeRequestsEvents:      in.MergeRequestsEvents,
		TagPushEvents:            in.TagPushEvents,
		NoteEvents:               in.NoteEvents,
		ConfidentialNoteEvents:   in.ConfidentialNoteEvents,
		PipelineEvents:           in.PipelineEvents,
		WikiPageEvents:           in.WikiPageEvents,

		PushChannel:              in.PushChannel,
		IssueChannel:             in.IssueChannel,
		ConfidentialIssueChannel: in.ConfidentialIssueChannel,
		MergeRequestChannel:      in.MergeRequestChannel,
		NoteChannel:              in.NoteChannel,
		ConfidentialNoteChannel:  in.ConfidentialNoteChannel,
		TagPushChannel:           in.TagPushChannel,
		PipelineChannel:          in.PipelineChannel,
		WikiPageChannel:          in.WikiPageChannel,
	}

	// WebHook is write-only in GitLab and typically required when initially setting the service.
	// Only include it when provided to avoid unintentionally clearing remote configuration.
	if in.WebHook != "" {
		opts.WebHook = &in.WebHook
	}

	return &opts
}

// GenerateIntegrationMattermostObservation converts gitlab.MattermostService to IntegrationMattermostObservation.
func GenerateIntegrationMattermostObservation(observation *gitlab.MattermostService) v1alpha1.IntegrationMattermostObservation {
	if observation == nil || observation.Properties == nil {
		return v1alpha1.IntegrationMattermostObservation{}
	}

	commonObservation := common.GenerateCommonIntegrationObservation(&observation.Service)

	return v1alpha1.IntegrationMattermostObservation{
		CommonIntegrationObservation: commonObservation,

		// WebHook is typically write-only in GitLab and not returned; leave as observed.
		WebHook:                   observation.Properties.WebHook,
		Username:                  observation.Properties.Username,
		Channel:                   observation.Properties.Channel,
		NotifyOnlyBrokenPipelines: bool(observation.Properties.NotifyOnlyBrokenPipelines),

		BranchesToBeNotified:     observation.Properties.BranchesToBeNotified,
		ConfidentialIssueChannel: observation.Properties.ConfidentialIssueChannel,
		ConfidentialNoteChannel:  observation.Properties.ConfidentialNoteChannel,
		IssueChannel:             observation.Properties.IssueChannel,
		MergeRequestChannel:      observation.Properties.MergeRequestChannel,
		NoteChannel:              observation.Properties.NoteChannel,
		TagPushChannel:           observation.Properties.TagPushChannel,
		PipelineChannel:          observation.Properties.PipelineChannel,
		PushChannel:              observation.Properties.PushChannel,
		VulnerabilityChannel:     observation.Properties.VulnerabilityChannel,
		WikiPageChannel:          observation.Properties.WikiPageChannel,
	}
}

// IsIntegrationMattermostUpToDate returns true if spec matches the observed GitLab Mattermost service.
//
// Note: WebHook is intentionally excluded from comparison because GitLab does not return it (write-only).
func IsIntegrationMattermostUpToDate(spec *v1alpha1.IntegrationMattermostParameters, observation *gitlab.MattermostService) bool { //nolint:gocyclo
	if observation == nil || observation.Properties == nil {
		return false
	}

	return clients.IsComparableEqualToComparablePtr(spec.Username, observation.Properties.Username) &&
		clients.IsComparableEqualToComparablePtr(spec.Channel, observation.Properties.Channel) &&
		clients.IsComparableEqualToComparablePtr(spec.NotifyOnlyBrokenPipelines, bool(observation.Properties.NotifyOnlyBrokenPipelines)) &&
		clients.IsComparableEqualToComparablePtr(spec.BranchesToBeNotified, observation.Properties.BranchesToBeNotified) &&
		clients.IsComparableEqualToComparablePtr(spec.PushEvents, observation.PushEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.IssuesEvents, observation.IssuesEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.ConfidentialIssuesEvents, observation.ConfidentialIssuesEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.MergeRequestsEvents, observation.MergeRequestsEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.TagPushEvents, observation.TagPushEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.NoteEvents, observation.NoteEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.ConfidentialNoteEvents, observation.ConfidentialNoteEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.PipelineEvents, observation.PipelineEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.WikiPageEvents, observation.WikiPageEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.PushChannel, observation.Properties.PushChannel) &&
		clients.IsComparableEqualToComparablePtr(spec.IssueChannel, observation.Properties.IssueChannel) &&
		clients.IsComparableEqualToComparablePtr(spec.ConfidentialIssueChannel, observation.Properties.ConfidentialIssueChannel) &&
		clients.IsComparableEqualToComparablePtr(spec.MergeRequestChannel, observation.Properties.MergeRequestChannel) &&
		clients.IsComparableEqualToComparablePtr(spec.NoteChannel, observation.Properties.NoteChannel) &&
		clients.IsComparableEqualToComparablePtr(spec.ConfidentialNoteChannel, observation.Properties.ConfidentialNoteChannel) &&
		clients.IsComparableEqualToComparablePtr(spec.TagPushChannel, observation.Properties.TagPushChannel) &&
		clients.IsComparableEqualToComparablePtr(spec.PipelineChannel, observation.Properties.PipelineChannel) &&
		clients.IsComparableEqualToComparablePtr(spec.WikiPageChannel, observation.Properties.WikiPageChannel)
}

// LateInitializeIntegrationMattermost fills nil spec fields using values from the remote Mattermost service.
// It mutates the spec in place and does NOT touch write-only fields like WebHook.
func LateInitializeIntegrationMattermost(in *v1alpha1.IntegrationMattermostParameters, svc *gitlab.MattermostService) {
	if in == nil || svc == nil || svc.Properties == nil {
		return
	}

	// Strings (only if remote value is non-empty).
	in.Username = clients.LateInitializeStringPtr(in.Username, svc.Properties.Username)
	in.Channel = clients.LateInitializeStringPtr(in.Channel, svc.Properties.Channel)
	in.BranchesToBeNotified = clients.LateInitializeStringPtr(in.BranchesToBeNotified, svc.Properties.BranchesToBeNotified)

	in.PushChannel = clients.LateInitializeStringPtr(in.PushChannel, svc.Properties.PushChannel)
	in.IssueChannel = clients.LateInitializeStringPtr(in.IssueChannel, svc.Properties.IssueChannel)
	in.ConfidentialIssueChannel = clients.LateInitializeStringPtr(in.ConfidentialIssueChannel, svc.Properties.ConfidentialIssueChannel)
	in.MergeRequestChannel = clients.LateInitializeStringPtr(in.MergeRequestChannel, svc.Properties.MergeRequestChannel)
	in.NoteChannel = clients.LateInitializeStringPtr(in.NoteChannel, svc.Properties.NoteChannel)
	in.ConfidentialNoteChannel = clients.LateInitializeStringPtr(in.ConfidentialNoteChannel, svc.Properties.ConfidentialNoteChannel)
	in.TagPushChannel = clients.LateInitializeStringPtr(in.TagPushChannel, svc.Properties.TagPushChannel)
	in.PipelineChannel = clients.LateInitializeStringPtr(in.PipelineChannel, svc.Properties.PipelineChannel)
	in.WikiPageChannel = clients.LateInitializeStringPtr(in.WikiPageChannel, svc.Properties.WikiPageChannel)

	// Booleans (initialize regardless of zero-ness when spec is nil).
	in.NotifyOnlyBrokenPipelines = clients.LateInitializeFromValue(in.NotifyOnlyBrokenPipelines, bool(svc.Properties.NotifyOnlyBrokenPipelines))

	// Event toggles are top-level booleans on MattermostService.
	in.PushEvents = clients.LateInitializeFromValue(in.PushEvents, svc.PushEvents)
	in.IssuesEvents = clients.LateInitializeFromValue(in.IssuesEvents, svc.IssuesEvents)
	in.ConfidentialIssuesEvents = clients.LateInitializeFromValue(in.ConfidentialIssuesEvents, svc.ConfidentialIssuesEvents)
	in.MergeRequestsEvents = clients.LateInitializeFromValue(in.MergeRequestsEvents, svc.MergeRequestsEvents)
	in.TagPushEvents = clients.LateInitializeFromValue(in.TagPushEvents, svc.TagPushEvents)
	in.NoteEvents = clients.LateInitializeFromValue(in.NoteEvents, svc.NoteEvents)
	in.ConfidentialNoteEvents = clients.LateInitializeFromValue(in.ConfidentialNoteEvents, svc.ConfidentialNoteEvents)
	in.PipelineEvents = clients.LateInitializeFromValue(in.PipelineEvents, svc.PipelineEvents)
	in.WikiPageEvents = clients.LateInitializeFromValue(in.WikiPageEvents, svc.WikiPageEvents)

	// WebHook is write-only; do NOT late-initialize it from observation.
}
