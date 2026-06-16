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
	"strings"

	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/instance/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
)

// ApplicationClient defines the GitLab Applications service operations used by the controller.
type ApplicationClient interface {
	CreateApplication(opt *gitlab.CreateApplicationOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Application, *gitlab.Response, error)
	ListApplications(opt *gitlab.ListApplicationsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.Application, *gitlab.Response, error)
	DeleteApplication(application int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewApplicationClient returns a new GitLab Applications client.
func NewApplicationClient(cfg common.Config) ApplicationClient {
	git := common.NewClient(cfg)
	return git.Applications
}

// GenerateApplicationObservation creates ApplicationObservation from a gitlab.Application.
func GenerateApplicationObservation(a *gitlab.Application) v1alpha1.ApplicationObservation {
	if a == nil {
		return v1alpha1.ApplicationObservation{}
	}
	obs := v1alpha1.ApplicationObservation{
		ID:            a.ID,
		Name:          a.ApplicationName,
		ApplicationID: a.ApplicationID,
		CallbackURL:   a.CallbackURL,
		Confidential:  a.Confidential,
		Scopes:        a.Scopes,
	}
	return obs
}

// GenerateCreateApplicationOptions produces CreateApplicationOptions from ApplicationParameters.
func GenerateCreateApplicationOptions(p *v1alpha1.ApplicationParameters) *gitlab.CreateApplicationOptions {
	options := &gitlab.CreateApplicationOptions{
		Name:         &p.Name,
		RedirectURI:  &p.RedirectURI,
		Confidential: p.Confidential,
	}

	if len(p.Scopes) > 0 {
		options.Scopes = ptr.To(strings.Join(p.Scopes, " "))
	}

	return options
}

// IsApplicationUpToDate checks whether the ApplicationParameters are in sync with a gitlab.Application.
// Since GitLab has no update endpoint for applications, this compares observable fields only.
func IsApplicationUpToDate(p *v1alpha1.ApplicationParameters, a *gitlab.Application) bool {
	if p == nil {
		return true
	}
	return a != nil &&
		p.Name == a.ApplicationName &&
		p.RedirectURI == a.CallbackURL &&
		clients.IsComparableEqualToComparablePtr(p.Confidential, a.Confidential) &&
		clients.AreStringSlicesEqual(p.Scopes, a.Scopes)
}

// FindApplicationByID returns the application with the given numeric ID from the list,
// or nil if not found.
func FindApplicationByID(apps []*gitlab.Application, id int64) *gitlab.Application {
	for _, a := range apps {
		if a != nil && a.ID == id {
			return a
		}
	}
	return nil
}
