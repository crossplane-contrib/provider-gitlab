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

	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
)

// AccessTokenClient defines Gitlab Group service operations
type AccessTokenClient interface {
	GetGroupAccessToken(pid interface{}, id int64, options ...gitlab.RequestOptionFunc) (*gitlab.GroupAccessToken, *gitlab.Response, error)
	CreateGroupAccessToken(pid interface{}, opt *gitlab.CreateGroupAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupAccessToken, *gitlab.Response, error)
	RevokeGroupAccessToken(pid interface{}, id int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
	RotateGroupAccessToken(gid any, id int64, opt *gitlab.RotateGroupAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupAccessToken, *gitlab.Response, error)
	RotateSelf(opt *gitlab.RotatePersonalAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error)
	GetSelf(options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error)
	RevokeSelf(options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// IsErrorGroupAccessTokenNotFound helper function to test for errGroupAccessTokenNotFound error.
func IsErrorGroupAccessTokenNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errGroupNotFound)
}

// NewAccessTokenClient returns a new Gitlab GroupAccessToken service
// NewAccessTokenClient returns a new Gitlab GroupAccessToken service
func NewAccessTokenClient(cfg common.Config) AccessTokenClient {
	git := common.NewClient(cfg)
	return &accessTokenClient{
		GroupAccessTokensServiceInterface:    git.GroupAccessTokens,
		PersonalAccessTokensServiceInterface: git.PersonalAccessTokens,
	}
}

// accessTokenClient composes group and personal access token services.
type accessTokenClient struct {
	gitlab.GroupAccessTokensServiceInterface
	gitlab.PersonalAccessTokensServiceInterface
}

func (c *accessTokenClient) RotateSelf(opt *gitlab.RotatePersonalAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return c.PersonalAccessTokensServiceInterface.RotatePersonalAccessTokenSelf(opt, options...)
}

// GetSelf returns the token used to authenticate the request (self-inform),
// via GET /personal_access_tokens/self. In self-managed mode the group access
// token authenticates as its own bot user, so this returns the very token the
// resource manages.
func (c *accessTokenClient) GetSelf(options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return c.PersonalAccessTokensServiceInterface.GetSinglePersonalAccessToken(options...)
}

// RevokeSelf revokes the token used to authenticate the request, via
// DELETE /personal_access_tokens/self.
func (c *accessTokenClient) RevokeSelf(options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.PersonalAccessTokensServiceInterface.RevokePersonalAccessTokenSelf(options...)
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
		accesstoken.ExpiresAt = common.GenerateRenewalExpiration(*p.RenewalPeriodDays)
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
		ExpiresAt:   common.TimeToMetaTime((*time.Time)(at.ExpiresAt)),
		Active:      at.Active,
		CreatedAt:   common.TimeToMetaTime(at.CreatedAt),
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
		accesstoken.ExpiresAt = common.GenerateRenewalExpiration(*p.RenewalPeriodDays)
	}

	return accesstoken
}

// GenerateRotateSelfOptions generates self-rotation options from access token parameters.
func GenerateRotateSelfOptions(p *v1alpha1.AccessTokenParameters) *gitlab.RotatePersonalAccessTokenOptions {
	opt := &gitlab.RotatePersonalAccessTokenOptions{}

	if p.ExpiresAt != nil {
		opt.ExpiresAt = (*gitlab.ISOTime)(&p.ExpiresAt.Time)
	} else if p.RenewalPeriodDays != nil {
		opt.ExpiresAt = common.GenerateRenewalExpiration(*p.RenewalPeriodDays)
	}

	return opt
}

// GenerateGroupAccessTokenObservationFromPAT builds the access-token observation
// from a gitlab.PersonalAccessToken, as returned by the self-inform endpoint in
// self-managed mode. A group access token is backed by a bot-user personal
// access token, so the self endpoints return a PersonalAccessToken rather than a
// GroupAccessToken. AccessLevel is not exposed on the self endpoint and is left
// zero.
func GenerateGroupAccessTokenObservationFromPAT(at *gitlab.PersonalAccessToken) v1alpha1.AccessTokenObservation {
	if at == nil {
		return v1alpha1.AccessTokenObservation{}
	}

	return v1alpha1.AccessTokenObservation{
		ID:          at.ID,
		Name:        at.Name,
		Description: at.Description,
		UserID:      at.UserID,
		Scopes:      at.Scopes,
		ExpiresAt:   common.TimeToMetaTime((*time.Time)(at.ExpiresAt)),
		Active:      at.Active,
		CreatedAt:   common.TimeToMetaTime(at.CreatedAt),
		Revoked:     at.Revoked,
	}
}

// ShouldRotateAccessTokenFromPAT returns true when the self-managed token must
// be rotated, evaluated against a gitlab.PersonalAccessToken from the self-inform
// endpoint.
func ShouldRotateAccessTokenFromPAT(p *v1alpha1.AccessTokenParameters, a *gitlab.PersonalAccessToken) bool {
	if a == nil {
		return true
	}

	var desiredExpiresAt *time.Time
	if p.ExpiresAt != nil {
		desiredExpiresAt = &p.ExpiresAt.Time
	}

	var createdAt *time.Time
	if p.RenewalPeriodDays != nil {
		createdAt = a.CreatedAt
	}

	return common.ShouldRotateToken(a.Active, createdAt, a.ExpiresAt, desiredExpiresAt, p.RenewBeforeDays, time.Now().UTC())
}

// ShouldRotateAccessToken returns true when the token must be rotated.
func ShouldRotateAccessToken(p *v1alpha1.AccessTokenParameters, a *gitlab.GroupAccessToken) bool {
	if a == nil {
		return true
	}

	var desiredExpiresAt *time.Time
	if p.ExpiresAt != nil {
		desiredExpiresAt = &p.ExpiresAt.Time
	}

	var createdAt *time.Time
	if p.RenewalPeriodDays != nil {
		createdAt = a.CreatedAt
	}

	return common.ShouldRotateToken(a.Active, createdAt, a.ExpiresAt, desiredExpiresAt, p.RenewBeforeDays, time.Now().UTC())
}
