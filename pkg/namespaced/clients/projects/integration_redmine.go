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

// RedmineClient defines GitLab Redmine integration operations for a project.
type RedmineClient interface {
	GetRedmineService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.RedmineService, *gitlab.Response, error)
	SetRedmineService(pid any, opt *gitlab.SetRedmineServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.RedmineService, *gitlab.Response, error)
	DeleteRedmineService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewRedmineClient returns a new GitLab Services client for project Redmine operations.
func NewRedmineClient(cfg common.Config) RedmineClient {
	git := common.NewClient(cfg)
	return git.Services
}

// GenerateSetRedmineServiceOptions produces SetRedmineServiceOptions from IntegrationRedmineParameters.
func GenerateSetRedmineServiceOptions(in *v1alpha1.IntegrationRedmineParameters) *gitlab.SetRedmineServiceOptions {
	return &gitlab.SetRedmineServiceOptions{
		ProjectURL:  &in.ProjectURL,
		IssuesURL:   &in.IssuesURL,
		NewIssueURL: &in.NewIssueURL,
	}
}

// GenerateIntegrationRedmineObservation converts gitlab.RedmineService to IntegrationRedmineObservation.
func GenerateIntegrationRedmineObservation(observation *gitlab.RedmineService) v1alpha1.IntegrationRedmineObservation {
	if observation == nil || observation.Properties == nil {
		return v1alpha1.IntegrationRedmineObservation{}
	}

	return v1alpha1.IntegrationRedmineObservation{
		CommonIntegrationObservation: common.GenerateCommonIntegrationObservation(&observation.Service),
		ProjectURL:                   observation.Properties.ProjectURL,
		IssuesURL:                    observation.Properties.IssuesURL,
		NewIssueURL:                  observation.Properties.NewIssueURL,
	}
}

// IsIntegrationRedmineUpToDate returns true if the desired spec matches the observed Redmine service.
func IsIntegrationRedmineUpToDate(spec *v1alpha1.IntegrationRedmineParameters, observation *gitlab.RedmineService) bool {
	if observation == nil || observation.Properties == nil {
		return false
	}

	return spec.ProjectURL == observation.Properties.ProjectURL &&
		spec.IssuesURL == observation.Properties.IssuesURL &&
		spec.NewIssueURL == observation.Properties.NewIssueURL
}
