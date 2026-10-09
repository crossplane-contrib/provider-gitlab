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

package groups

import (
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
)

// MattermostClient defines GitLab Mattermost integration operations for a group.
type MattermostClient interface {
	GetGroupMattermostIntegration(gid any, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMattermostIntegration, *gitlab.Response, error)
	SetGroupMattermostIntegration(gid any, opt *gitlab.GroupMattermostIntegrationOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMattermostIntegration, *gitlab.Response, error)
	DeleteGroupMattermostIntegration(gid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewMattermostClient returns a new GitLab Integrations client for group Mattermost operations.
func NewMattermostClient(cfg common.Config) MattermostClient {
	git := common.NewClient(cfg)
	return git.Integrations
}

// GenerateGroupMattermostIntegrationOptions produces GroupMattermostIntegrationOptions from IntegrationMattermostParameters.
// The webhook value must be resolved from the referenced secret prior to calling this function
// and passed in through `webhook`. An empty webhook is omitted from the options so that
// existing remote state is not unintentionally cleared.
func GenerateGroupMattermostIntegrationOptions(in *v1alpha1.IntegrationMattermostParameters, webhook string) *gitlab.GroupMattermostIntegrationOptions {
	if in == nil {
		return &gitlab.GroupMattermostIntegrationOptions{}
	}

	opts := gitlab.GroupMattermostIntegrationOptions{
		Username:                   in.Username,
		Channel:                    in.Channel,
		NotifyOnlyBrokenPipelines:  in.NotifyOnlyBrokenPipelines,
		BranchesToBeNotified:       in.BranchesToBeNotified,
		LabelsToBeNotified:         in.LabelsToBeNotified,
		LabelsToBeNotifiedBehavior: in.LabelsToBeNotifiedBehavior,
		UseInheritedSettings:       in.UseInheritedSettings,

		PushEvents:               in.PushEvents,
		IssuesEvents:             in.IssuesEvents,
		ConfidentialIssuesEvents: in.ConfidentialIssuesEvents,
		MergeRequestsEvents:      in.MergeRequestsEvents,
		TagPushEvents:            in.TagPushEvents,
		NoteEvents:               in.NoteEvents,
		ConfidentialNoteEvents:   in.ConfidentialNoteEvents,
		PipelineEvents:           in.PipelineEvents,
		WikiPageEvents:           in.WikiPageEvents,
		DeploymentEvents:         in.DeploymentEvents,
		AlertEvents:              in.AlertEvents,
		VulnerabilityEvents:      in.VulnerabilityEvents,

		PushChannel:              in.PushChannel,
		IssueChannel:             in.IssueChannel,
		ConfidentialIssueChannel: in.ConfidentialIssueChannel,
		MergeRequestChannel:      in.MergeRequestChannel,
		NoteChannel:              in.NoteChannel,
		ConfidentialNoteChannel:  in.ConfidentialNoteChannel,
		TagPushChannel:           in.TagPushChannel,
		PipelineChannel:          in.PipelineChannel,
		WikiPageChannel:          in.WikiPageChannel,
		DeploymentChannel:        in.DeploymentChannel,
		AlertChannel:             in.AlertChannel,
		VulnerabilityChannel:     in.VulnerabilityChannel,
	}

	if webhook != "" {
		opts.WebHook = &webhook
	}

	return &opts
}

// GenerateIntegrationMattermostObservation converts gitlab.GroupMattermostIntegration to IntegrationMattermostObservation.
// The webhook is intentionally not surfaced as it is a secret.
func GenerateIntegrationMattermostObservation(observation *gitlab.GroupMattermostIntegration) v1alpha1.IntegrationMattermostObservation {
	if observation == nil {
		return v1alpha1.IntegrationMattermostObservation{}
	}

	out := v1alpha1.IntegrationMattermostObservation{
		CommonIntegrationObservation: common.GenerateCommonIntegrationObservation(&observation.Integration),
	}

	if p := observation.Properties; p != nil {
		out.Username = p.Username
		out.Channel = p.Channel
		out.PushChannel = p.PushChannel
		out.IssueChannel = p.IssueChannel
		out.ConfidentialIssueChannel = p.ConfidentialIssueChannel
		out.MergeRequestChannel = p.MergeRequestChannel
		out.NoteChannel = p.NoteChannel
		out.ConfidentialNoteChannel = p.ConfidentialNoteChannel
		out.TagPushChannel = p.TagPushChannel
		out.PipelineChannel = p.PipelineChannel
		out.WikiPageChannel = p.WikiPageChannel
		out.DeploymentChannel = p.DeploymentChannel
		out.AlertChannel = p.AlertChannel
		out.VulnerabilityChannel = p.VulnerabilityChannel
	}

	return out
}

// IsIntegrationMattermostUpToDate returns true if the remote Mattermost integration matches the desired spec.
//
// Note: WebHook is excluded because GitLab does not return it (write-only). NotifyOnlyBrokenPipelines,
// BranchesToBeNotified, LabelsToBeNotified and LabelsToBeNotifiedBehavior are excluded because the
// GitLab client does not decode them from the API response properties.
func IsIntegrationMattermostUpToDate(spec *v1alpha1.IntegrationMattermostParameters, observation *gitlab.GroupMattermostIntegration) bool { //nolint:gocyclo
	if spec == nil || observation == nil || observation.Properties == nil {
		return false
	}

	if !observation.Active {
		return false
	}

	p := observation.Properties
	return clients.IsComparableEqualToComparablePtr(spec.Username, p.Username) &&
		clients.IsComparableEqualToComparablePtr(spec.Channel, p.Channel) &&
		clients.IsComparableEqualToComparablePtr(spec.PushEvents, observation.PushEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.IssuesEvents, observation.IssuesEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.ConfidentialIssuesEvents, observation.ConfidentialIssuesEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.MergeRequestsEvents, observation.MergeRequestsEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.TagPushEvents, observation.TagPushEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.NoteEvents, observation.NoteEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.ConfidentialNoteEvents, observation.ConfidentialNoteEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.PipelineEvents, observation.PipelineEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.WikiPageEvents, observation.WikiPageEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.DeploymentEvents, observation.DeploymentEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.AlertEvents, observation.AlertEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.VulnerabilityEvents, observation.VulnerabilityEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.PushChannel, p.PushChannel) &&
		clients.IsComparableEqualToComparablePtr(spec.IssueChannel, p.IssueChannel) &&
		clients.IsComparableEqualToComparablePtr(spec.ConfidentialIssueChannel, p.ConfidentialIssueChannel) &&
		clients.IsComparableEqualToComparablePtr(spec.MergeRequestChannel, p.MergeRequestChannel) &&
		clients.IsComparableEqualToComparablePtr(spec.NoteChannel, p.NoteChannel) &&
		clients.IsComparableEqualToComparablePtr(spec.ConfidentialNoteChannel, p.ConfidentialNoteChannel) &&
		clients.IsComparableEqualToComparablePtr(spec.TagPushChannel, p.TagPushChannel) &&
		clients.IsComparableEqualToComparablePtr(spec.PipelineChannel, p.PipelineChannel) &&
		clients.IsComparableEqualToComparablePtr(spec.WikiPageChannel, p.WikiPageChannel) &&
		clients.IsComparableEqualToComparablePtr(spec.DeploymentChannel, p.DeploymentChannel) &&
		clients.IsComparableEqualToComparablePtr(spec.AlertChannel, p.AlertChannel) &&
		clients.IsComparableEqualToComparablePtr(spec.VulnerabilityChannel, p.VulnerabilityChannel)
}
