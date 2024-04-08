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

// AccessTokenClient defines Gitlab Group service operations
type AccessTokenClient interface {
	GetGroupAccessToken(pid interface{}, id int, options ...gitlab.RequestOptionFunc) (*gitlab.GroupAccessToken, *gitlab.Response, error)
	CreateGroupAccessToken(pid interface{}, opt *gitlab.CreateGroupAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupAccessToken, *gitlab.Response, error)
	RevokeGroupAccessToken(pid interface{}, id int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// IsErrorGroupAccessTokenNotFound helper function to test for errGroupAccessTokenNotFound error.
func IsErrorGroupAccessTokenNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errGroupNotFound)
}

// NewAccessTokenClient returns a new Gitlab GroupAccessToken service
func NewAccessTokenClient(cfg clients.Config) AccessTokenClient {
	git := clients.NewClient(cfg)
	return git.GroupAccessTokens
}

// GenerateCreateGroupAccessTokenOptions generates project creation options
func GenerateCreateGroupAccessTokenOptions(name string, p *v1alpha1.AccessTokenParameters) *gitlab.CreateGroupAccessTokenOptions {
	accesstoken := &gitlab.CreateGroupAccessTokenOptions{
		Name:   &name,
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
