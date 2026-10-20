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

// CustomIssueTrackerClient defines GitLab Custom Issue Tracker integration operations for a project.
type CustomIssueTrackerClient interface {
	GetCustomIssueTrackerService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.CustomIssueTrackerService, *gitlab.Response, error)
	SetCustomIssueTrackerService(pid any, opt *gitlab.SetCustomIssueTrackerServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.CustomIssueTrackerService, *gitlab.Response, error)
	DeleteCustomIssueTrackerService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewCustomIssueTrackerClient returns a new GitLab Services client for project Custom Issue Tracker operations.
func NewCustomIssueTrackerClient(cfg common.Config) CustomIssueTrackerClient {
	git := common.NewClient(cfg)
	return git.Services
}

// GenerateSetCustomIssueTrackerServiceOptions produces SetCustomIssueTrackerServiceOptions from IntegrationCustomIssueTrackerParameters.
func GenerateSetCustomIssueTrackerServiceOptions(in *v1alpha1.IntegrationCustomIssueTrackerParameters) *gitlab.SetCustomIssueTrackerServiceOptions {
	return &gitlab.SetCustomIssueTrackerServiceOptions{
		ProjectURL:  &in.ProjectURL,
		IssuesURL:   &in.IssuesURL,
		NewIssueURL: &in.NewIssueURL,
	}
}

// GenerateIntegrationCustomIssueTrackerObservation converts gitlab.CustomIssueTrackerService to IntegrationCustomIssueTrackerObservation.
func GenerateIntegrationCustomIssueTrackerObservation(observation *gitlab.CustomIssueTrackerService) v1alpha1.IntegrationCustomIssueTrackerObservation {
	if observation == nil || observation.Properties == nil {
		return v1alpha1.IntegrationCustomIssueTrackerObservation{}
	}

	return v1alpha1.IntegrationCustomIssueTrackerObservation{
		CommonIntegrationObservation: common.GenerateCommonIntegrationObservation(&observation.Service),
		ProjectURL:                   observation.Properties.ProjectURL,
		IssuesURL:                    observation.Properties.IssuesURL,
		NewIssueURL:                  observation.Properties.NewIssueURL,
	}
}

// IsIntegrationCustomIssueTrackerUpToDate returns true if the desired spec matches the observed Custom Issue Tracker service.
func IsIntegrationCustomIssueTrackerUpToDate(spec *v1alpha1.IntegrationCustomIssueTrackerParameters, observation *gitlab.CustomIssueTrackerService) bool {
	if observation == nil || observation.Properties == nil {
		return false
	}

	return spec.ProjectURL == observation.Properties.ProjectURL &&
		spec.IssuesURL == observation.Properties.IssuesURL &&
		spec.NewIssueURL == observation.Properties.NewIssueURL
}
