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

// TelegramClient defines GitLab Telegram integration operations for a project.
type TelegramClient interface {
	GetTelegramService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.TelegramService, *gitlab.Response, error)
	SetTelegramService(pid any, opt *gitlab.SetTelegramServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.TelegramService, *gitlab.Response, error)
	DeleteTelegramService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewTelegramClient returns a new GitLab Services client for project Telegram operations.
func NewTelegramClient(cfg common.Config) TelegramClient {
	git := common.NewClient(cfg)
	return git.Services
}

// GenerateSetTelegramServiceOptions produces SetTelegramServiceOptions from IntegrationTelegramParameters.
// The token must be resolved from the referenced secret before calling this function.
// GitLab rejects every set request without a token, so it is always sent.
func GenerateSetTelegramServiceOptions(in *v1alpha1.IntegrationTelegramParameters, token string) *gitlab.SetTelegramServiceOptions {
	if in == nil {
		return &gitlab.SetTelegramServiceOptions{}
	}

	return &gitlab.SetTelegramServiceOptions{
		Token:                     &token,
		Room:                      &in.Room,
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

// GenerateIntegrationTelegramObservation converts gitlab.TelegramService to IntegrationTelegramObservation.
func GenerateIntegrationTelegramObservation(observation *gitlab.TelegramService) v1alpha1.IntegrationTelegramObservation {
	if observation == nil || observation.Properties == nil {
		return v1alpha1.IntegrationTelegramObservation{}
	}

	return v1alpha1.IntegrationTelegramObservation{
		CommonIntegrationObservation: common.GenerateCommonIntegrationObservation(&observation.Service),
		Room:                         observation.Properties.Room,
		NotifyOnlyBrokenPipelines:    observation.Properties.NotifyOnlyBrokenPipelines,
		BranchesToBeNotified:         observation.Properties.BranchesToBeNotified,
	}
}

// IsIntegrationTelegramUpToDate returns true if the desired spec matches the observed Telegram service.
//
// Note: Token is intentionally excluded from comparison because GitLab does not return it (write-only).
func IsIntegrationTelegramUpToDate(spec *v1alpha1.IntegrationTelegramParameters, observation *gitlab.TelegramService) bool { //nolint:gocyclo
	if observation == nil || observation.Properties == nil {
		return false
	}

	return spec.Room == observation.Properties.Room &&
		clients.IsComparableEqualToComparablePtr(spec.NotifyOnlyBrokenPipelines, observation.Properties.NotifyOnlyBrokenPipelines) &&
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

// LateInitializeIntegrationTelegram fills nil spec fields using values from the remote Telegram service.
// It mutates the spec in place and does NOT touch write-only fields like the token.
func LateInitializeIntegrationTelegram(in *v1alpha1.IntegrationTelegramParameters, svc *gitlab.TelegramService) {
	if in == nil || svc == nil || svc.Properties == nil {
		return
	}

	in.NotifyOnlyBrokenPipelines = clients.LateInitializeFromValue(in.NotifyOnlyBrokenPipelines, svc.Properties.NotifyOnlyBrokenPipelines)
	in.BranchesToBeNotified = clients.LateInitializeStringPtr(in.BranchesToBeNotified, svc.Properties.BranchesToBeNotified)
	in.PushEvents = clients.LateInitializeFromValue(in.PushEvents, svc.PushEvents)
	in.IssuesEvents = clients.LateInitializeFromValue(in.IssuesEvents, svc.IssuesEvents)
	in.ConfidentialIssuesEvents = clients.LateInitializeFromValue(in.ConfidentialIssuesEvents, svc.ConfidentialIssuesEvents)
	in.MergeRequestsEvents = clients.LateInitializeFromValue(in.MergeRequestsEvents, svc.MergeRequestsEvents)
	in.TagPushEvents = clients.LateInitializeFromValue(in.TagPushEvents, svc.TagPushEvents)
	in.NoteEvents = clients.LateInitializeFromValue(in.NoteEvents, svc.NoteEvents)
	in.ConfidentialNoteEvents = clients.LateInitializeFromValue(in.ConfidentialNoteEvents, svc.ConfidentialNoteEvents)
	in.PipelineEvents = clients.LateInitializeFromValue(in.PipelineEvents, svc.PipelineEvents)
	in.WikiPageEvents = clients.LateInitializeFromValue(in.WikiPageEvents, svc.WikiPageEvents)
}
