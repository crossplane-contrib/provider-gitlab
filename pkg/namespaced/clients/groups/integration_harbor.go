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
)

// HarborClient defines GitLab Harbor integration operations for a group.
type HarborClient interface {
	GetGroupHarborSettings(gid any, options ...gitlab.RequestOptionFunc) (*gitlab.HarborIntegration, *gitlab.Response, error)
	SetUpGroupHarbor(gid any, opt *gitlab.SetUpHarborOptions, options ...gitlab.RequestOptionFunc) (*gitlab.HarborIntegration, *gitlab.Response, error)
	DisableGroupHarbor(gid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewHarborClient returns a new GitLab Integrations client for group Harbor operations.
func NewHarborClient(cfg common.Config) HarborClient {
	git := common.NewClient(cfg)
	return git.Integrations
}

// GenerateSetUpHarborOptions produces SetUpHarborOptions from IntegrationHarborParameters.
// The password value must be resolved from the referenced secret prior to calling this function
// and passed in through `password`. An empty password is omitted from the options so that
// existing remote state is not unintentionally cleared.
func GenerateSetUpHarborOptions(in *v1alpha1.IntegrationHarborParameters, password string) *gitlab.SetUpHarborOptions {
	if in == nil {
		return &gitlab.SetUpHarborOptions{}
	}

	opts := gitlab.SetUpHarborOptions{
		URL:                  &in.URL,
		ProjectName:          &in.ProjectName,
		Username:             &in.Username,
		UseInheritedSettings: in.UseInheritedSettings,
	}

	if password != "" {
		opts.Password = &password
	}

	return &opts
}

// GenerateIntegrationHarborObservation converts gitlab.HarborIntegration to IntegrationHarborObservation.
func GenerateIntegrationHarborObservation(observation *gitlab.HarborIntegration) v1alpha1.IntegrationHarborObservation {
	if observation == nil {
		return v1alpha1.IntegrationHarborObservation{}
	}

	return v1alpha1.IntegrationHarborObservation{
		CommonIntegrationObservation: common.GenerateCommonIntegrationObservation(&observation.Integration),
	}
}

// IsIntegrationHarborUpToDate returns true if the remote Harbor integration matches the desired spec.
func IsIntegrationHarborUpToDate(spec *v1alpha1.IntegrationHarborParameters, observation *gitlab.HarborIntegration) bool {
	if spec == nil || observation == nil {
		return false
	}

	if !observation.Active {
		return false
	}

	return spec.URL == observation.Properties.URL &&
		spec.ProjectName == observation.Properties.ProjectName &&
		spec.Username == observation.Properties.Username
}
