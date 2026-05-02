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
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
)

// AccessTokenClient defines Gitlab Group service operations
type AccessTokenClient interface {
	GetGroupAccessToken(pid interface{}, id int64, options ...gitlab.RequestOptionFunc) (*gitlab.GroupAccessToken, *gitlab.Response, error)
	CreateGroupAccessToken(pid interface{}, opt *gitlab.CreateGroupAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupAccessToken, *gitlab.Response, error)
	RevokeGroupAccessToken(pid interface{}, id int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
	RotateGroupAccessToken(gid any, id int64, opt *gitlab.RotateGroupAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupAccessToken, *gitlab.Response, error)
}

// IsErrorGroupAccessTokenNotFound helper function to test for errGroupAccessTokenNotFound error.
func IsErrorGroupAccessTokenNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errGroupNotFound)
}

// NewAccessTokenClient returns a new Gitlab GroupAccessToken service
func NewAccessTokenClient(cfg common.Config) AccessTokenClient {
	git := common.NewClient(cfg)
	return git.GroupAccessTokens
}

// GenerateCreateGroupAccessTokenOptions generates project creation options
func GenerateCreateGroupAccessTokenOptions(name string, p *v1alpha1.AccessTokenParameters) *gitlab.CreateGroupAccessTokenOptions {
	accesstoken := &gitlab.CreateGroupAccessTokenOptions{
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

// GenerateGroupAccessTokenObservation generates group access token observation from gitlab.GroupAccessToken
func GenerateGroupAccessTokenObservation(at *gitlab.GroupAccessToken) v1alpha1.AccessTokenObservation {
	if at == nil {
		return v1alpha1.AccessTokenObservation{}
	}

	return v1alpha1.AccessTokenObservation{
		ID:          at.ID,
		Name:        at.Name,
		Description: at.Description,
		UserID:      at.UserID,
		Scopes:      at.Scopes,
		ExpiresAt:   clients.TimeToMetaTime((*time.Time)(at.ExpiresAt)),
		Active:      at.Active,
		CreatedAt:   clients.TimeToMetaTime(at.CreatedAt),
		Revoked:     at.Revoked,
		AccessLevel: int64(at.AccessLevel),
	}
}

// GenerateRotateGroupAccessTokenOptions generates group access token rotation options
func GenerateRotateGroupAccessTokenOptions(p *v1alpha1.AccessTokenParameters) *gitlab.RotateGroupAccessTokenOptions {
	accesstoken := &gitlab.RotateGroupAccessTokenOptions{}

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
func ShouldRotateAccessToken(p *v1alpha1.AccessTokenParameters, a *gitlab.GroupAccessToken) bool {
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
