/*
Copyright 2026 The Crossplane Authors.

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
	"testing"

	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	projectsv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
)

func TestGenerateSetCustomIssueTrackerServiceOptions(t *testing.T) {
	got := GenerateSetCustomIssueTrackerServiceOptions(&projectsv1alpha1.IntegrationCustomIssueTrackerParameters{
		ProjectID:   ptr.To[int64](123),
		ProjectURL:  "https://tracker.example.com/project",
		IssuesURL:   "https://tracker.example.com/issues/:id",
		NewIssueURL: "https://tracker.example.com/issues/new",
	})
	want := &gitlab.SetCustomIssueTrackerServiceOptions{
		ProjectURL:  ptr.To("https://tracker.example.com/project"),
		IssuesURL:   ptr.To("https://tracker.example.com/issues/:id"),
		NewIssueURL: ptr.To("https://tracker.example.com/issues/new"),
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("GenerateSetCustomIssueTrackerServiceOptions(): -want, +got:\n%s", diff)
	}
}

func TestGenerateIntegrationCustomIssueTrackerObservation(t *testing.T) {
	cases := map[string]struct {
		service *gitlab.CustomIssueTrackerService
		want    projectsv1alpha1.IntegrationCustomIssueTrackerObservation
	}{
		"FullObservation": {
			service: &gitlab.CustomIssueTrackerService{
				Service: gitlab.Service{
					ID:     7,
					Title:  "Custom issue tracker",
					Slug:   "custom-issue-tracker",
					Active: true,
				},
				Properties: &gitlab.CustomIssueTrackerServiceProperties{
					ProjectURL:  "https://tracker.example.com/project",
					IssuesURL:   "https://tracker.example.com/issues/:id",
					NewIssueURL: "https://tracker.example.com/issues/new",
				},
			},
			want: projectsv1alpha1.IntegrationCustomIssueTrackerObservation{
				CommonIntegrationObservation: v1alpha1.CommonIntegrationObservation{
					ID:                             ptr.To[int64](7),
					Title:                          ptr.To("Custom issue tracker"),
					Slug:                           ptr.To("custom-issue-tracker"),
					Active:                         ptr.To(true),
					AlertEvents:                    ptr.To(false),
					CommitEvents:                   ptr.To(false),
					ConfidentialIssuesEvents:       ptr.To(false),
					ConfidentialNoteEvents:         ptr.To(false),
					DeploymentEvents:               ptr.To(false),
					GroupConfidentialMentionEvents: ptr.To(false),
					GroupMentionEvents:             ptr.To(false),
					IncidentEvents:                 ptr.To(false),
					IssuesEvents:                   ptr.To(false),
					JobEvents:                      ptr.To(false),
					MergeRequestsEvents:            ptr.To(false),
					NoteEvents:                     ptr.To(false),
					PipelineEvents:                 ptr.To(false),
					PushEvents:                     ptr.To(false),
					TagPushEvents:                  ptr.To(false),
					VulnerabilityEvents:            ptr.To(false),
					WikiPageEvents:                 ptr.To(false),
					CommentOnEventEnabled:          ptr.To(false),
					Inherited:                      ptr.To(false),
				},
				ProjectURL:  "https://tracker.example.com/project",
				IssuesURL:   "https://tracker.example.com/issues/:id",
				NewIssueURL: "https://tracker.example.com/issues/new",
			},
		},
		"NilService": {
			service: nil,
			want:    projectsv1alpha1.IntegrationCustomIssueTrackerObservation{},
		},
		"NilProperties": {
			service: &gitlab.CustomIssueTrackerService{Service: gitlab.Service{ID: 7}},
			want:    projectsv1alpha1.IntegrationCustomIssueTrackerObservation{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateIntegrationCustomIssueTrackerObservation(tc.service)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateIntegrationCustomIssueTrackerObservation(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsIntegrationCustomIssueTrackerUpToDate(t *testing.T) {
	spec := &projectsv1alpha1.IntegrationCustomIssueTrackerParameters{
		ProjectURL:  "https://tracker.example.com/project",
		IssuesURL:   "https://tracker.example.com/issues/:id",
		NewIssueURL: "https://tracker.example.com/issues/new",
	}

	cases := map[string]struct {
		observation *gitlab.CustomIssueTrackerService
		want        bool
	}{
		"UpToDate": {
			observation: &gitlab.CustomIssueTrackerService{
				Properties: &gitlab.CustomIssueTrackerServiceProperties{
					ProjectURL:  "https://tracker.example.com/project",
					IssuesURL:   "https://tracker.example.com/issues/:id",
					NewIssueURL: "https://tracker.example.com/issues/new",
				},
			},
			want: true,
		},
		"OutOfDateProjectURL": {
			observation: &gitlab.CustomIssueTrackerService{
				Properties: &gitlab.CustomIssueTrackerServiceProperties{
					ProjectURL:  "https://other.example.com/project",
					IssuesURL:   "https://tracker.example.com/issues/:id",
					NewIssueURL: "https://tracker.example.com/issues/new",
				},
			},
			want: false,
		},
		"OutOfDateIssuesURL": {
			observation: &gitlab.CustomIssueTrackerService{
				Properties: &gitlab.CustomIssueTrackerServiceProperties{
					ProjectURL:  "https://tracker.example.com/project",
					IssuesURL:   "https://other.example.com/issues/:id",
					NewIssueURL: "https://tracker.example.com/issues/new",
				},
			},
			want: false,
		},
		"OutOfDateNewIssueURL": {
			observation: &gitlab.CustomIssueTrackerService{
				Properties: &gitlab.CustomIssueTrackerServiceProperties{
					ProjectURL:  "https://tracker.example.com/project",
					IssuesURL:   "https://tracker.example.com/issues/:id",
					NewIssueURL: "https://other.example.com/issues/new",
				},
			},
			want: false,
		},
		"NilProperties": {
			observation: &gitlab.CustomIssueTrackerService{},
			want:        false,
		},
		"NilObservation": {
			observation: nil,
			want:        false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsIntegrationCustomIssueTrackerUpToDate(spec, tc.observation)
			if got != tc.want {
				t.Errorf("IsIntegrationCustomIssueTrackerUpToDate(): got %v, want %v", got, tc.want)
			}
		})
	}
}
