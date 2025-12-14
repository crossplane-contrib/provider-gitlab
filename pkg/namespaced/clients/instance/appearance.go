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
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/instance/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
)

// ApplicationSettingsClient defines Gitlab Application Settings service operations
type AppearanceClient interface {
	GetAppearance(options ...gitlab.RequestOptionFunc) (*gitlab.Appearance, *gitlab.Response, error)
	ChangeAppearance(opt *gitlab.ChangeAppearanceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Appearance, *gitlab.Response, error)
}

// NewAppearanceClient returns a new Gitlab Appearance service
func NewAppearanceClient(cfg common.Config) AppearanceClient {
	git := common.NewClient(cfg)
	return git.Appearance
}

// GenerateUpdateAppearanceOptions generates Gitlab ChangeAppearanceOptions from AppearanceParameters
func GenerateUpdateAppearanceOptions(params *v1alpha1.AppearanceParameters) *gitlab.ChangeAppearanceOptions {
	if params == nil {
		return &gitlab.ChangeAppearanceOptions{}
	}
	return &gitlab.ChangeAppearanceOptions{
		Title:                       params.Title,
		Description:                 params.Description,
		PWAName:                     params.PWAName,
		PWAShortName:                params.PWAShortName,
		PWADescription:              params.PWADescription,
		PWAIcon:                     params.PWAIcon,
		Logo:                        params.Logo,
		HeaderLogo:                  params.HeaderLogo,
		Favicon:                     params.Favicon,
		MemberGuidelines:            params.MemberGuidelines,
		NewProjectGuidelines:        params.NewProjectGuidelines,
		ProfileImageGuidelines:      params.ProfileImageGuidelines,
		HeaderMessage:               params.HeaderMessage,
		FooterMessage:               params.FooterMessage,
		MessageBackgroundColor:      params.MessageBackgroundColor,
		MessageFontColor:            params.MessageFontColor,
		EmailHeaderAndFooterEnabled: params.EmailHeaderAndFooterEnabled,
	}
}

// GenerateAppearanceObservation generates an observation from a Gitlab Appearance
func GenerateAppearanceObservation(observed *gitlab.Appearance) v1alpha1.AppearanceObservation {
	if observed == nil {
		return v1alpha1.AppearanceObservation{}
	}
	return v1alpha1.AppearanceObservation{
		Title:                       observed.Title,
		Description:                 observed.Description,
		PWAName:                     observed.PWAName,
		PWAShortName:                observed.PWAShortName,
		PWADescription:              observed.PWADescription,
		PWAIcon:                     observed.PWAIcon,
		Logo:                        observed.Logo,
		HeaderLogo:                  observed.HeaderLogo,
		Favicon:                     observed.Favicon,
		MemberGuidelines:            observed.MemberGuidelines,
		NewProjectGuidelines:        observed.NewProjectGuidelines,
		ProfileImageGuidelines:      observed.ProfileImageGuidelines,
		HeaderMessage:               observed.HeaderMessage,
		FooterMessage:               observed.FooterMessage,
		MessageBackgroundColor:      observed.MessageBackgroundColor,
		MessageFontColor:            observed.MessageFontColor,
		EmailHeaderAndFooterEnabled: observed.EmailHeaderAndFooterEnabled,
	}
}

// IsAppearanceUpToDate checks whether the Appearance parameters are up to date
func IsAppearanceUpToDate(spec *v1alpha1.AppearanceParameters, observed *gitlab.Appearance) bool {
	if spec == nil {
		return true
	}
	if observed == nil {
		return false
	}

	// Use a compact list to keep cyclomatic complexity low
	checks := []bool{
		clients.IsComparableEqualToComparablePtr(spec.Title, observed.Title),
		clients.IsComparableEqualToComparablePtr(spec.Description, observed.Description),
		clients.IsComparableEqualToComparablePtr(spec.PWAName, observed.PWAName),
		clients.IsComparableEqualToComparablePtr(spec.PWAShortName, observed.PWAShortName),
		clients.IsComparableEqualToComparablePtr(spec.PWADescription, observed.PWADescription),
		clients.IsComparableEqualToComparablePtr(spec.PWAIcon, observed.PWAIcon),
		clients.IsComparableEqualToComparablePtr(spec.Logo, observed.Logo),
		clients.IsComparableEqualToComparablePtr(spec.HeaderLogo, observed.HeaderLogo),
		clients.IsComparableEqualToComparablePtr(spec.Favicon, observed.Favicon),
		clients.IsComparableEqualToComparablePtr(spec.MemberGuidelines, observed.MemberGuidelines),
		clients.IsComparableEqualToComparablePtr(spec.NewProjectGuidelines, observed.NewProjectGuidelines),
		clients.IsComparableEqualToComparablePtr(spec.ProfileImageGuidelines, observed.ProfileImageGuidelines),
		clients.IsComparableEqualToComparablePtr(spec.HeaderMessage, observed.HeaderMessage),
		clients.IsComparableEqualToComparablePtr(spec.FooterMessage, observed.FooterMessage),
		clients.IsComparableEqualToComparablePtr(spec.MessageBackgroundColor, observed.MessageBackgroundColor),
		clients.IsComparableEqualToComparablePtr(spec.MessageFontColor, observed.MessageFontColor),
		clients.IsComparableEqualToComparablePtr(spec.EmailHeaderAndFooterEnabled, observed.EmailHeaderAndFooterEnabled),
	}

	for _, ok := range checks {
		if !ok {
			return false
		}
	}
	return true
}
