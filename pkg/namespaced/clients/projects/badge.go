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

	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
)

// ProjectBadgeClient defines Gitlab Project service operations
type BadgeClient interface {
	ListProjectBadges(gid any, opt *gitlab.ListProjectBadgesOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectBadge, *gitlab.Response, error)
	GetProjectBadge(gid any, badge int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectBadge, *gitlab.Response, error)
	AddProjectBadge(gid any, opt *gitlab.AddProjectBadgeOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectBadge, *gitlab.Response, error)
	EditProjectBadge(gid any, badge int, opt *gitlab.EditProjectBadgeOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectBadge, *gitlab.Response, error)
	DeleteProjectBadge(gid any, badge int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
	PreviewProjectBadge(gid any, opt *gitlab.ProjectBadgePreviewOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectBadge, *gitlab.Response, error)
}

// IsErrorProjectBadgeNotFound helper function to test for errProjectBadgeNotFound error.
func IsErrorProjectBadgeNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errProjectNotFound)
}

// NewBadgeClient returns a new Gitlab ProjectBadge service
func NewBadgeClient(cfg common.Config) BadgeClient {
	git := common.NewClient(cfg)
	return git.ProjectBadges
}

// GenerateAddProjectBadgeOptions generates project creation options from v1alpha1 parameters
func GenerateAddProjectBadgeOptions(p *v1alpha1.BadgeParameters) *gitlab.AddProjectBadgeOptions {
	badge := &gitlab.AddProjectBadgeOptions{
		Name:     p.Name,
		ImageURL: &p.ImageURL,
		LinkURL:  &p.LinkURL,
	}
	return badge
}

// GenerateEditProjectBadgeOptions generates project edit options from v1alpha1 parameters
func GenerateEditProjectBadgeOptions(p *v1alpha1.BadgeParameters) *gitlab.EditProjectBadgeOptions {
	badge := &gitlab.EditProjectBadgeOptions{
		Name:     p.Name,
		ImageURL: &p.ImageURL,
		LinkURL:  &p.LinkURL,
	}
	return badge
}

// GenerateBadgeObservation generates v1alpha1 observation from Gitlab ProjectBadge
func GenerateBadgeObservation(b *gitlab.ProjectBadge) v1alpha1.BadgeObservation {
	return v1alpha1.BadgeObservation{
		ID:               b.ID,
		LinkURL:          b.LinkURL,
		RenderedLinkURL:  b.RenderedLinkURL,
		ImageURL:         b.ImageURL,
		RenderedImageURL: b.RenderedImageURL,
		Name:             b.Name,
	}
}

// IsBadgeUpToDate checks whether the observed Gitlab ProjectBadge is up to date
// compared to the desired v1alpha1 parameters
func IsBadgeUpToDate(spec *v1alpha1.BadgeParameters, observed *gitlab.ProjectBadge) bool {
	if spec == nil {
		return true
	}
	if observed == nil {
		return false
	}

	checks := []bool{
		clients.IsComparableEqualToComparablePtr(spec.Name, observed.Name),
		cmp.Equal(spec.ImageURL, observed.ImageURL),
		cmp.Equal(spec.LinkURL, observed.LinkURL),
	}

	for _, check := range checks {
		if !check {
			return false
		}
	}
	return true
}
