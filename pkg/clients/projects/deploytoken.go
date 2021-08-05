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
	"strings"

	"github.com/xanzy/go-gitlab"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
)

// DeployTokenClient defines Gitlab Project service operations
type DeployTokenClient interface {
	ListProjectDeployTokens(pid interface{}, opt *gitlab.ListProjectDeployTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.DeployToken, *gitlab.Response, error)
	CreateProjectDeployToken(pid interface{}, opt *gitlab.CreateProjectDeployTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error)
	DeleteProjectDeployToken(pid interface{}, deployToken int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// IsErrorProjectDeployTokenNotFound helper function to test for errProjectDeployTokenNotFound error.
func IsErrorProjectDeployTokenNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errProjectNotFound)
}

// NewDeployTokenClient returns a new Gitlab ProjectDeployToken service
func NewDeployTokenClient(cfg clients.Config) DeployTokenClient {
	git := clients.NewClient(cfg)
	return git.DeployTokens
}

// GenerateCreateProjectDeployTokenOptions generates project creation options
func GenerateCreateProjectDeployTokenOptions(name string, p *v1alpha1.DeployTokenParameters) *gitlab.CreateProjectDeployTokenOptions {
	deploytoken := &gitlab.CreateProjectDeployTokenOptions{
		Name:   &name,
		Scopes: p.Scopes,
	}

	if p.ExpiresAt != nil {
		deploytoken.ExpiresAt = &p.ExpiresAt.Time
	}

	if p.Username != nil {
		deploytoken.Username = p.Username
	}

	return deploytoken
}
