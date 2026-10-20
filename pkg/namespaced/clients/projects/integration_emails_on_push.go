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
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
)

// EmailsOnPushClient defines GitLab Emails on Push integration operations for a project.
type EmailsOnPushClient interface {
	GetEmailsOnPushService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.EmailsOnPushService, *gitlab.Response, error)
	SetEmailsOnPushService(pid any, opt *gitlab.SetEmailsOnPushServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.EmailsOnPushService, *gitlab.Response, error)
	DeleteEmailsOnPushService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewEmailsOnPushClient returns a new GitLab Services client for project Emails on Push operations.
func NewEmailsOnPushClient(cfg common.Config) EmailsOnPushClient {
	git := common.NewClient(cfg)
	return git.Services
}

// GenerateSetEmailsOnPushServiceOptions produces SetEmailsOnPushServiceOptions from IntegrationEmailsOnPushParameters.
func GenerateSetEmailsOnPushServiceOptions(in *v1alpha1.IntegrationEmailsOnPushParameters) *gitlab.SetEmailsOnPushServiceOptions {
	if in == nil {
		return &gitlab.SetEmailsOnPushServiceOptions{}
	}

	return &gitlab.SetEmailsOnPushServiceOptions{
		Recipients:             &in.Recipients,
		DisableDiffs:           in.DisableDiffs,
		SendFromCommitterEmail: in.SendFromCommitterEmail,
		PushEvents:             in.PushEvents,
		TagPushEvents:          in.TagPushEvents,
		BranchesToBeNotified:   in.BranchesToBeNotified,
	}
}

// GenerateIntegrationEmailsOnPushObservation converts gitlab.EmailsOnPushService to IntegrationEmailsOnPushObservation.
func GenerateIntegrationEmailsOnPushObservation(observation *gitlab.EmailsOnPushService) v1alpha1.IntegrationEmailsOnPushObservation {
	if observation == nil || observation.Properties == nil {
		return v1alpha1.IntegrationEmailsOnPushObservation{}
	}

	return v1alpha1.IntegrationEmailsOnPushObservation{
		CommonIntegrationObservation: common.GenerateCommonIntegrationObservation(&observation.Service),
		Recipients:                   observation.Properties.Recipients,
		DisableDiffs:                 observation.Properties.DisableDiffs,
		SendFromCommitterEmail:       observation.Properties.SendFromCommitterEmail,
		BranchesToBeNotified:         observation.Properties.BranchesToBeNotified,
	}
}

// IsIntegrationEmailsOnPushUpToDate returns true if the desired spec matches the observed Emails on Push service.
func IsIntegrationEmailsOnPushUpToDate(spec *v1alpha1.IntegrationEmailsOnPushParameters, observation *gitlab.EmailsOnPushService) bool {
	if observation == nil || observation.Properties == nil {
		return false
	}

	return spec.Recipients == observation.Properties.Recipients &&
		clients.IsComparableEqualToComparablePtr(spec.DisableDiffs, observation.Properties.DisableDiffs) &&
		clients.IsComparableEqualToComparablePtr(spec.SendFromCommitterEmail, observation.Properties.SendFromCommitterEmail) &&
		clients.IsComparableEqualToComparablePtr(spec.BranchesToBeNotified, observation.Properties.BranchesToBeNotified) &&
		clients.IsComparableEqualToComparablePtr(spec.PushEvents, observation.PushEvents) &&
		clients.IsComparableEqualToComparablePtr(spec.TagPushEvents, observation.TagPushEvents)
}

// LateInitializeIntegrationEmailsOnPush fills nil spec fields using values from the remote Emails on Push service.
func LateInitializeIntegrationEmailsOnPush(in *v1alpha1.IntegrationEmailsOnPushParameters, svc *gitlab.EmailsOnPushService) {
	if in == nil || svc == nil || svc.Properties == nil {
		return
	}

	in.DisableDiffs = clients.LateInitializeFromValue(in.DisableDiffs, svc.Properties.DisableDiffs)
	in.SendFromCommitterEmail = clients.LateInitializeFromValue(in.SendFromCommitterEmail, svc.Properties.SendFromCommitterEmail)
	in.BranchesToBeNotified = clients.LateInitializeStringPtr(in.BranchesToBeNotified, svc.Properties.BranchesToBeNotified)
	in.PushEvents = clients.LateInitializeFromValue(in.PushEvents, svc.PushEvents)
	in.TagPushEvents = clients.LateInitializeFromValue(in.TagPushEvents, svc.TagPushEvents)
}
