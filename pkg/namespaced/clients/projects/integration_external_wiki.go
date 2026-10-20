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
)

// ExternalWikiClient defines GitLab External Wiki integration operations for a project.
type ExternalWikiClient interface {
	GetExternalWikiService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.ExternalWikiService, *gitlab.Response, error)
	SetExternalWikiService(pid any, opt *gitlab.SetExternalWikiServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ExternalWikiService, *gitlab.Response, error)
	DeleteExternalWikiService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewExternalWikiClient returns a new GitLab Services client for project External Wiki operations.
func NewExternalWikiClient(cfg common.Config) ExternalWikiClient {
	git := common.NewClient(cfg)
	return git.Services
}

// GenerateSetExternalWikiServiceOptions produces SetExternalWikiServiceOptions from IntegrationExternalWikiParameters.
func GenerateSetExternalWikiServiceOptions(in *v1alpha1.IntegrationExternalWikiParameters) *gitlab.SetExternalWikiServiceOptions {
	return &gitlab.SetExternalWikiServiceOptions{
		ExternalWikiURL: &in.ExternalWikiURL,
	}
}

// GenerateIntegrationExternalWikiObservation converts gitlab.ExternalWikiService to IntegrationExternalWikiObservation.
func GenerateIntegrationExternalWikiObservation(observation *gitlab.ExternalWikiService) v1alpha1.IntegrationExternalWikiObservation {
	if observation == nil || observation.Properties == nil {
		return v1alpha1.IntegrationExternalWikiObservation{}
	}

	return v1alpha1.IntegrationExternalWikiObservation{
		CommonIntegrationObservation: common.GenerateCommonIntegrationObservation(&observation.Service),
		ExternalWikiURL:              observation.Properties.ExternalWikiURL,
	}
}

// IsIntegrationExternalWikiUpToDate returns true if the desired spec matches the observed External Wiki service.
func IsIntegrationExternalWikiUpToDate(spec *v1alpha1.IntegrationExternalWikiParameters, observation *gitlab.ExternalWikiService) bool {
	if observation == nil || observation.Properties == nil {
		return false
	}

	return spec.ExternalWikiURL == observation.Properties.ExternalWikiURL
}
