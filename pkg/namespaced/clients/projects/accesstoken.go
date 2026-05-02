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
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
)

// AccessTokenClient defines Gitlab Project service operations
type AccessTokenClient interface {
	GetProjectAccessToken(pid interface{}, id int64, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectAccessToken, *gitlab.Response, error)
	CreateProjectAccessToken(pid interface{}, opt *gitlab.CreateProjectAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectAccessToken, *gitlab.Response, error)
	RevokeProjectAccessToken(pid interface{}, id int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
	RotateProjectAccessToken(pid interface{}, id int64, opt *gitlab.RotateProjectAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectAccessToken, *gitlab.Response, error)
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
	} else if p.RenewalPeriodDays != nil {
		accesstoken.ExpiresAt = generateRenewalExpiration(*p.RenewalPeriodDays)
	}

	if p.AccessLevel != nil {
		accesstoken.AccessLevel = (*gitlab.AccessLevelValue)(p.AccessLevel)
	}

	if p.Description != nil {
		accesstoken.Description = ptr.To(*p.Description)
	}

	return accesstoken
}

// GenerateProjectAccessTokenObservation generates project access token observation from gitlab.ProjectAccessToken.
func GenerateProjectAccessTokenObservation(at *gitlab.ProjectAccessToken) v1alpha1.AccessTokenObservation {
	if at == nil {
		return v1alpha1.AccessTokenObservation{}
	}

	return v1alpha1.AccessTokenObservation{
		TokenID: ptr.To(at.ID),
	}
}

// GenerateRotateProjectAccessTokenOptions generates project access token rotation options
func GenerateRotateProjectAccessTokenOptions(p *v1alpha1.AccessTokenParameters) *gitlab.RotateProjectAccessTokenOptions {
	accesstoken := &gitlab.RotateProjectAccessTokenOptions{}

	if p.ExpiresAt != nil {
		accesstoken.ExpiresAt = (*gitlab.ISOTime)(&p.ExpiresAt.Time)
	} else if p.RenewalPeriodDays != nil {
		accesstoken.ExpiresAt = generateRenewalExpiration(*p.RenewalPeriodDays)
	}

	return accesstoken
}

func generateRenewalExpiration(days int) *gitlab.ISOTime {
	return (*gitlab.ISOTime)(ptr.To(time.Now().UTC().AddDate(0, 0, days)))
}

// ShouldRotateAccessToken returns true when the token must be rotated:
// the token is inactive, or ExpiresAt is set and the actual expiry does not match.
func ShouldRotateAccessToken(p *v1alpha1.AccessTokenParameters, a *gitlab.ProjectAccessToken) bool {
	if a == nil || !a.Active {
		return true
	}

	if p != nil && p.ExpiresAt != nil {
		if a.ExpiresAt == nil {
			return true
		}
		return !sameDay(p.ExpiresAt.Time, time.Time(*a.ExpiresAt))
	}

	return false
}

func sameDay(a, b time.Time) bool {
	a = a.UTC()
	b = b.UTC()
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}
