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
	gitlab "gitlab.com/gitlab-org/api/client-go"

	commonv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
)

// ServiceAccountClient defines Gitlab Service Account service operations
type ServiceAccountClient interface {
	CreateServiceAccount(gid any, opt *gitlab.CreateServiceAccountOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupServiceAccount, *gitlab.Response, error)
	UpdateServiceAccount(gid any, serviceAccount int, opt *gitlab.UpdateServiceAccountOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupServiceAccount, *gitlab.Response, error)
	DeleteServiceAccount(gid any, serviceAccount int, opt *gitlab.DeleteServiceAccountOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewServiceAccountClient returns a new Gitlab Service Account service
func NewServiceAccountClient(cfg common.Config) ServiceAccountClient {
	git := common.NewClient(cfg)
	return git.Groups
}

// GenerateServiceAccountObservation is used to produce ServiceAccountObservation from Gitlab User
func GenerateServiceAccountObservation(u *gitlab.GroupServiceAccount) v1alpha1.ServiceAccountObservation {
	return v1alpha1.ServiceAccountObservation{
		CommonServiceAccountObservation: commonv1alpha1.CommonServiceAccountObservation{
			ID:       u.ID,
			Name:     u.Name,
			Username: u.UserName,
			Email:    u.Email,
		},
	}
}

// GenerateServiceAccountObservationFromUser is used to produce ServiceAccountObservation from Gitlab User
func GenerateServiceAccountObservationFromUser(u *gitlab.User) v1alpha1.ServiceAccountObservation {
	return v1alpha1.ServiceAccountObservation{
		CommonServiceAccountObservation: commonv1alpha1.CommonServiceAccountObservation{
			ID:       u.ID,
			Name:     u.Name,
			Username: u.Username,
			Email:    u.Email,
		},
	}
}

// GenerateUpdateServiceAccountOptions is used to produce ModifyUserOptions from ServiceAccountParameters
func GenerateUpdateServiceAccountOptions(p *v1alpha1.ServiceAccountParameters) *gitlab.UpdateServiceAccountOptions {
	return &gitlab.UpdateServiceAccountOptions{
		Name:     p.Name,
		Username: p.Username,
	}
}

// GenerateServiceAccountCreateOptions is used to produce CreateServiceAccountUserOptions from ServiceAccountParameters
func GenerateServiceAccountCreateOptions(p *v1alpha1.ServiceAccountParameters) *gitlab.CreateServiceAccountOptions {
	return &gitlab.CreateServiceAccountOptions{
		Name:     p.Name,
		Username: p.Username,
		Email:    p.Email,
	}
}

// IsServiceAccountUpToDate checks whether the ServiceAccountParameters is in sync with Gitlab User
// returns true if the parameters are nil or all fields are in sync
func IsServiceAccountUpToDate(p *v1alpha1.ServiceAccountParameters, u *gitlab.User) bool {
	if p == nil {
		return true
	}
	if u == nil {
		return false
	}

	if !clients.IsComparableEqualToComparablePtr(p.Name, u.Name) {
		return false
	}

	if !clients.IsComparableEqualToComparablePtr(p.Username, u.Username) {
		return false
	}

	if !clients.IsComparableEqualToComparablePtr(p.Email, u.Email) {
		return false
	}

	return true
}
