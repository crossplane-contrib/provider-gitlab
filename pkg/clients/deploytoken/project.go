/*
Copyright 2020 The Crossplane Authors.

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

package deploytoken

import (
	"strings"

	"github.com/xanzy/go-gitlab"

	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
)

const (
	errProjectNotFound = "404 Project Not Found"
)

// ProjectDeployTokenClient defines Gitlab Project service operations
type ProjectDeployTokenClient interface {
	ListProjectDeployTokens(pid interface{}, opt *gitlab.ListProjectDeployTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.DeployToken, *gitlab.Response, error)
	CreateProjectDeployToken(pid interface{}, opt *gitlab.CreateProjectDeployTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error)
	DeleteProjectDeployToken(pid interface{}, deployToken int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// IsErrorProjectNotFound helper function to test for errProjectNotFound error.
func IsErrorProjectNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errProjectNotFound)
}

// NewProjectDeployTokenClient returns a new Gitlab ProjectDeployToken service
func NewProjectDeployTokenClient(cfg clients.Config) ProjectDeployTokenClient {
	git := clients.NewClient(cfg)
	return git.DeployTokens
}
