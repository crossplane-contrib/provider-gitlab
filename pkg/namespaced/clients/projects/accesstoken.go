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

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
)

// AccessTokenClient defines Gitlab Project service operations
type AccessTokenClient interface {
	GetProjectAccessToken(pid interface{}, id int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectAccessToken, *gitlab.Response, error)
	CreateProjectAccessToken(pid interface{}, opt *gitlab.CreateProjectAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectAccessToken, *gitlab.Response, error)
	RevokeProjectAccessToken(pid interface{}, id int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// IsErrorProjectAccessTokenNotFound helper function to test for errProjectAccessTokenNotFound error.
func IsErrorProjectAccessTokenNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errProjectNotFound)
}

// NewAccessTokenClient returns a new Gitlab ProjectAccessToken service
func NewAccessTokenClient(cfg common.Config) AccessTokenClient {
	git := common.NewClient(cfg)
	return git.ProjectAccessTokens
}

// GenerateCreateProjectAccessTokenOptions generates project creation options
func GenerateCreateProjectAccessTokenOptions(name string, p *v1alpha1.AccessTokenParameters) *gitlab.CreateProjectAccessTokenOptions {
	accesstoken := &gitlab.CreateProjectAccessTokenOptions{
		Name:   &p.Name,
		Scopes: &p.Scopes,
	}

	if p.ExpiresAt != nil {
		accesstoken.ExpiresAt = (*gitlab.ISOTime)(&p.ExpiresAt.Time)
	}

	if p.AccessLevel != nil {
		accesstoken.AccessLevel = (*gitlab.AccessLevelValue)(p.AccessLevel)
	}

	return accesstoken
}
