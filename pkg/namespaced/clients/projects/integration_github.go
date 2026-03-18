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

// GithubClient defines GitLab GitHub integration operations.
type GithubClient interface {
	GetGithubService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.GithubService, *gitlab.Response, error)
	SetGithubService(pid any, opt *gitlab.SetGithubServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GithubService, *gitlab.Response, error)
	DeleteGithubService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewGithubClient returns a new GitLab Services client.
func NewGithubClient(cfg common.Config) GithubClient {
	git := common.NewClient(cfg)
	return git.Services
}

// GenerateSetGithubServiceOptions produces SetGithubServiceOptions from IntegrationGithubParameters.
func GenerateSetGithubServiceOptions(in *v1alpha1.IntegrationGithubParameters) *gitlab.SetGithubServiceOptions {
	if in == nil {
		return &gitlab.SetGithubServiceOptions{}
	}

	opts := gitlab.SetGithubServiceOptions{
		RepositoryURL: in.RepositoryURL,
		StaticContext: in.StaticContext,
	}

	// Token is write-only in GitLab and typically required when initially setting the service.
	// Only include it when provided to avoid unintentionally clearing remote configuration.
	if in.Token != "" {
		opts.Token = &in.Token
	}

	return &opts
}

// GenerateIntegrationGithubObservation converts gitlab.GithubService to IntegrationGithubObservation.
func GenerateIntegrationGithubObservation(observation *gitlab.GithubService) v1alpha1.IntegrationGithubObservation {
	if observation == nil || observation.Properties == nil {
		return v1alpha1.IntegrationGithubObservation{}
	}

	commonObservation := common.GenerateCommonIntegrationObservation(&observation.Service)

	return v1alpha1.IntegrationGithubObservation{
		CommonIntegrationObservation: commonObservation,
		RepositoryURL:                observation.Properties.RepositoryURL,
		StaticContext:                observation.Properties.StaticContext,
	}
}

// IsIntegrationGithubUpToDate returns true if spec matches the observed GitLab GitHub service.
//
// Note: Token is intentionally excluded from comparison because GitLab does not return it (write-only).
func IsIntegrationGithubUpToDate(spec *v1alpha1.IntegrationGithubParameters, observation *gitlab.GithubService) bool {
	if observation == nil || observation.Properties == nil {
		return false
	}

	return clients.IsComparableEqualToComparablePtr(spec.RepositoryURL, observation.Properties.RepositoryURL) &&
		clients.IsComparableEqualToComparablePtr(spec.StaticContext, observation.Properties.StaticContext)
}

// LateInitializeIntegrationGithub fills nil spec fields using values from the remote GitHub service.
// It mutates the spec in place and does NOT touch write-only fields like Token.
func LateInitializeIntegrationGithub(in *v1alpha1.IntegrationGithubParameters, svc *gitlab.GithubService) {
	if in == nil || svc == nil || svc.Properties == nil {
		return
	}

	in.RepositoryURL = clients.LateInitializeStringPtr(in.RepositoryURL, svc.Properties.RepositoryURL)
	in.StaticContext = clients.LateInitializeFromValue(in.StaticContext, svc.Properties.StaticContext)

	// Token is write-only; do NOT late-initialize it from observation.
}
