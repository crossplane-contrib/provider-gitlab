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
	"strings"

	"github.com/xanzy/go-gitlab"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
)

// DeployTokenClient defines Gitlab Group service operations
type DeployTokenClient interface {
	ListGroupDeployTokens(gid interface{}, opt *gitlab.ListGroupDeployTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.DeployToken, *gitlab.Response, error)
	CreateGroupDeployToken(gid interface{}, opt *gitlab.CreateGroupDeployTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error)
	DeleteGroupDeployToken(gid interface{}, deployToken int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// IsErrorGroupDeployTokenNotFound helper function to test for errGroupDeployTokenNotFound error.
func IsErrorGroupDeployTokenNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errGroupNotFound)
}

// NewDeployTokenClient returns a new Gitlab GroupDeployToken service
func NewDeployTokenClient(cfg clients.Config) DeployTokenClient {
	git := clients.NewClient(cfg)
	return git.DeployTokens
}

// GenerateCreateGroupDeployTokenOptions generates group creation options
func GenerateCreateGroupDeployTokenOptions(name string, p *v1alpha1.DeployTokenParameters) *gitlab.CreateGroupDeployTokenOptions {
	deploytoken := &gitlab.CreateGroupDeployTokenOptions{
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
