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
	"slices"
	"strings"

	gitlab "gitlab.com/gitlab-org/api/client-go/v3"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
)

// PipelinesEmailClient defines GitLab pipeline status emails integration operations for a project.
type PipelinesEmailClient interface {
	GetPipelinesEmailService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.PipelinesEmailService, *gitlab.Response, error)
	SetPipelinesEmailService(pid any, opt *gitlab.SetPipelinesEmailServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelinesEmailService, *gitlab.Response, error)
	DeletePipelinesEmailService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewPipelinesEmailClient returns a new GitLab Services client for project pipeline status emails operations.
func NewPipelinesEmailClient(cfg common.Config) PipelinesEmailClient {
	git := common.NewClient(cfg)
	return git.Services
}

// parseRecipients splits the recipients string the way GitLab does, on commas, spaces and newlines.
func parseRecipients(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\r' || r == '\n'
	})
}

// GenerateSetPipelinesEmailServiceOptions produces SetPipelinesEmailServiceOptions from IntegrationPipelinesEmailParameters.
func GenerateSetPipelinesEmailServiceOptions(in *v1alpha1.IntegrationPipelinesEmailParameters) *gitlab.SetPipelinesEmailServiceOptions {
	recipients := strings.Join(in.Recipients, ",")
	return &gitlab.SetPipelinesEmailServiceOptions{
		Recipients:                &recipients,
		NotifyOnlyBrokenPipelines: in.NotifyOnlyBrokenPipelines,
		BranchesToBeNotified:      in.BranchesToBeNotified,
		PipelineEvents:            in.PipelineEvents,
	}
}

// GenerateIntegrationPipelinesEmailObservation converts gitlab.PipelinesEmailService to IntegrationPipelinesEmailObservation.
func GenerateIntegrationPipelinesEmailObservation(observation *gitlab.PipelinesEmailService) v1alpha1.IntegrationPipelinesEmailObservation {
	if observation == nil || observation.Properties == nil {
		return v1alpha1.IntegrationPipelinesEmailObservation{}
	}

	return v1alpha1.IntegrationPipelinesEmailObservation{
		CommonIntegrationObservation: common.GenerateCommonIntegrationObservation(&observation.Service),
		Recipients:                   parseRecipients(observation.Properties.Recipients),
		NotifyOnlyBrokenPipelines:    bool(observation.Properties.NotifyOnlyBrokenPipelines),
		BranchesToBeNotified:         observation.Properties.BranchesToBeNotified,
	}
}

// IsIntegrationPipelinesEmailUpToDate returns true if the desired spec matches the observed pipeline status emails service.
// Recipients compare as a set because GitLab treats them as one.
func IsIntegrationPipelinesEmailUpToDate(spec *v1alpha1.IntegrationPipelinesEmailParameters, observation *gitlab.PipelinesEmailService) bool {
	if observation == nil || observation.Properties == nil {
		return false
	}

	return slices.Equal(slices.Sorted(slices.Values(spec.Recipients)), slices.Sorted(slices.Values(parseRecipients(observation.Properties.Recipients)))) &&
		clients.IsComparableEqualToComparablePtr(spec.NotifyOnlyBrokenPipelines, bool(observation.Properties.NotifyOnlyBrokenPipelines)) &&
		clients.IsComparableEqualToComparablePtr(spec.BranchesToBeNotified, observation.Properties.BranchesToBeNotified) &&
		clients.IsComparableEqualToComparablePtr(spec.PipelineEvents, observation.PipelineEvents)
}

// LateInitializeIntegrationPipelinesEmail fills nil spec fields using values from the remote pipeline status emails service.
func LateInitializeIntegrationPipelinesEmail(in *v1alpha1.IntegrationPipelinesEmailParameters, svc *gitlab.PipelinesEmailService) {
	if svc == nil || svc.Properties == nil {
		return
	}

	in.NotifyOnlyBrokenPipelines = clients.LateInitializeFromValue(in.NotifyOnlyBrokenPipelines, bool(svc.Properties.NotifyOnlyBrokenPipelines))
	in.BranchesToBeNotified = clients.LateInitializeStringPtr(in.BranchesToBeNotified, svc.Properties.BranchesToBeNotified)
	in.PipelineEvents = clients.LateInitializeFromValue(in.PipelineEvents, svc.PipelineEvents)
}
