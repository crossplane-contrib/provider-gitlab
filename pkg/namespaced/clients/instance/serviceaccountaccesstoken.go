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

package instance

import (
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/instance/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
)

// ServiceAccountAccessTokenClient defines GitLab instance service account personal
// access token operations.
//
// Instance service accounts have no dedicated token endpoints; their tokens are
// managed through the personal access tokens API. Owner mode therefore relies on
// instance-admin endpoints (create a token for an arbitrary user, and address
// tokens by ID), while self-managed mode uses the generic self endpoints.
type ServiceAccountAccessTokenClient interface {
	// Owner-authenticated endpoints (the ProviderConfig is an instance admin).
	GetUser(user int64, opt *gitlab.GetUserOptions, options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error)
	CreatePersonalAccessToken(user int64, opt *gitlab.CreatePersonalAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error)
	GetSinglePersonalAccessTokenByID(token int64, options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error)
	RotatePersonalAccessTokenByID(token int64, opt *gitlab.RotatePersonalAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error)
	RevokePersonalAccessTokenByID(token int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

	// Self-authenticated endpoints (the ProviderConfig authenticates with the
	// very token this resource manages). Used in self-managed mode.
	GetServiceAccountSelf(options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error)
	RotateServiceAccountSelf(opt *gitlab.RotatePersonalAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error)
	RevokeServiceAccountSelf(options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewServiceAccountAccessTokenClient returns a new GitLab instance service account access token service.
func NewServiceAccountAccessTokenClient(cfg common.Config) ServiceAccountAccessTokenClient {
	git := common.NewClient(cfg)
	return &serviceAccountAccessTokenClient{
		UsersServiceInterface:                git.Users,
		PersonalAccessTokensServiceInterface: git.PersonalAccessTokens,
	}
}

// serviceAccountAccessTokenClient composes the users and personal access token services.
type serviceAccountAccessTokenClient struct {
	gitlab.UsersServiceInterface
	gitlab.PersonalAccessTokensServiceInterface
}

// GetServiceAccountSelf returns the token used to authenticate the request
// (self-inform), via GET /personal_access_tokens/self.
func (c *serviceAccountAccessTokenClient) GetServiceAccountSelf(options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return c.PersonalAccessTokensServiceInterface.GetSinglePersonalAccessToken(options...)
}

// RotateServiceAccountSelf rotates the token used to authenticate the request,
// via POST /personal_access_tokens/self/rotate.
func (c *serviceAccountAccessTokenClient) RotateServiceAccountSelf(opt *gitlab.RotatePersonalAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return c.PersonalAccessTokensServiceInterface.RotatePersonalAccessTokenSelf(opt, options...)
}

// RevokeServiceAccountSelf revokes the token used to authenticate the request,
// via DELETE /personal_access_tokens/self.
func (c *serviceAccountAccessTokenClient) RevokeServiceAccountSelf(options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.PersonalAccessTokensServiceInterface.RevokePersonalAccessTokenSelf(options...)
}

// GenerateCreateServiceAccountAccessTokenOptions generates creation options for the
// admin user-token endpoint.
func GenerateCreateServiceAccountAccessTokenOptions(p *v1alpha1.ServiceAccountAccessTokenParameters) *gitlab.CreatePersonalAccessTokenOptions {
	accesstoken := &gitlab.CreatePersonalAccessTokenOptions{
		Name:   &p.Name,
		Scopes: &p.Scopes,
	}

	if p.ExpiresAt != nil {
		accesstoken.ExpiresAt = (*gitlab.ISOTime)(&p.ExpiresAt.Time)
	} else if p.RenewalPeriodDays != nil {
		accesstoken.ExpiresAt = common.GenerateRenewalExpiration(*p.RenewalPeriodDays)
	}

	if p.Description != nil {
		accesstoken.Description = ptr.To(*p.Description)
	}

	return accesstoken
}

// GenerateServiceAccountAccessTokenObservation generates an observation from a gitlab.PersonalAccessToken.
func GenerateServiceAccountAccessTokenObservation(at *gitlab.PersonalAccessToken) v1alpha1.ServiceAccountAccessTokenObservation {
	if at == nil {
		return v1alpha1.ServiceAccountAccessTokenObservation{}
	}

	return v1alpha1.ServiceAccountAccessTokenObservation{
		ID:          at.ID,
		Name:        at.Name,
		Description: at.Description,
		UserID:      at.UserID,
		Scopes:      at.Scopes,
		ExpiresAt:   common.TimeToMetaTime((*time.Time)(at.ExpiresAt)),
		Active:      at.Active,
		CreatedAt:   common.TimeToMetaTime(at.CreatedAt),
		LastUsedAt:  common.TimeToMetaTime(at.LastUsedAt),
		Revoked:     at.Revoked,
	}
}

// GenerateRotateServiceAccountAccessTokenOptions generates rotation options. The same
// option type is used for both the owner (by-id) and self rotate endpoints.
func GenerateRotateServiceAccountAccessTokenOptions(p *v1alpha1.ServiceAccountAccessTokenParameters) *gitlab.RotatePersonalAccessTokenOptions {
	accesstoken := &gitlab.RotatePersonalAccessTokenOptions{}

	if p.ExpiresAt != nil {
		accesstoken.ExpiresAt = (*gitlab.ISOTime)(&p.ExpiresAt.Time)
	} else if p.RenewalPeriodDays != nil {
		accesstoken.ExpiresAt = common.GenerateRenewalExpiration(*p.RenewalPeriodDays)
	}

	return accesstoken
}

// ShouldRotateServiceAccountAccessToken returns true when the token must be rotated.
func ShouldRotateServiceAccountAccessToken(p *v1alpha1.ServiceAccountAccessTokenParameters, a *gitlab.PersonalAccessToken) bool {
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
