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
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"

	commonv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
)

// ServiceAccountClient defines Gitlab project service account service operations.
type ServiceAccountClient interface {
	ListProjectServiceAccounts(pid any, opt *gitlab.ListProjectServiceAccountsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectServiceAccount, *gitlab.Response, error)
	CreateProjectServiceAccount(pid any, opt *gitlab.CreateProjectServiceAccountOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectServiceAccount, *gitlab.Response, error)
	UpdateProjectServiceAccount(pid any, serviceAccount int64, opt *gitlab.UpdateProjectServiceAccountOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectServiceAccount, *gitlab.Response, error)
	DeleteProjectServiceAccount(pid any, serviceAccount int64, opt *gitlab.DeleteProjectServiceAccountOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewServiceAccountClient returns a new Gitlab project service account service.
func NewServiceAccountClient(cfg common.Config) ServiceAccountClient {
	git := common.NewClient(cfg)
	return git.Projects
}

// GetProjectServiceAccount paginates a project's service accounts and returns the
// one matching serviceAccountID. There is no get-by-id endpoint, so a missing
// service account returns a nil result and the last response for not-found detection.
func GetProjectServiceAccount(c ServiceAccountClient, pid any, serviceAccountID int64, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectServiceAccount, *gitlab.Response, error) {
	opt := &gitlab.ListProjectServiceAccountsOptions{
		ListOptions: gitlab.ListOptions{PerPage: 100, Page: 1},
	}

	var lastRes *gitlab.Response
	for {
		accounts, res, err := c.ListProjectServiceAccounts(pid, opt, options...)
		if err != nil {
			return nil, res, err
		}
		lastRes = res

		for _, a := range accounts {
			if a.ID == serviceAccountID {
				return a, res, nil
			}
		}

		if res == nil || res.NextPage == 0 {
			break
		}
		opt.Page = res.NextPage
	}

	return nil, lastRes, nil
}

// GenerateServiceAccountObservation is used to produce ServiceAccountObservation from a gitlab.ProjectServiceAccount.
func GenerateServiceAccountObservation(a *gitlab.ProjectServiceAccount) v1alpha1.ServiceAccountObservation {
	if a == nil {
		return v1alpha1.ServiceAccountObservation{}
	}
	return v1alpha1.ServiceAccountObservation{
		CommonServiceAccountObservation: commonv1alpha1.CommonServiceAccountObservation{
			ID:       a.ID,
			Name:     a.Name,
			Username: a.Username,
			Email:    a.Email,
		},
	}
}

// GenerateServiceAccountCreateOptions produces CreateProjectServiceAccountOptions from ServiceAccountParameters.
func GenerateServiceAccountCreateOptions(p *v1alpha1.ServiceAccountParameters) *gitlab.CreateProjectServiceAccountOptions {
	return &gitlab.CreateProjectServiceAccountOptions{
		Name:     p.Name,
		Username: p.Username,
		Email:    p.Email,
	}
}

// GenerateUpdateServiceAccountOptions produces UpdateProjectServiceAccountOptions from ServiceAccountParameters.
func GenerateUpdateServiceAccountOptions(p *v1alpha1.ServiceAccountParameters) *gitlab.UpdateProjectServiceAccountOptions {
	return &gitlab.UpdateProjectServiceAccountOptions{
		Name:     p.Name,
		Username: p.Username,
		Email:    p.Email,
	}
}

// IsServiceAccountUpToDate checks whether the ServiceAccountParameters is in sync
// with the observed gitlab.ProjectServiceAccount.
func IsServiceAccountUpToDate(p *v1alpha1.ServiceAccountParameters, a *gitlab.ProjectServiceAccount) bool {
	if p == nil {
		return true
	}
	if a == nil {
		return false
	}

	return clients.IsComparableEqualToComparablePtr(p.Name, a.Name) &&
		clients.IsComparableEqualToComparablePtr(p.Username, a.Username) &&
		clients.IsComparableEqualToComparablePtr(p.Email, a.Email)
}
