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

func TestGenerateSetRedmineServiceOptions(t *testing.T) {
	got := GenerateSetRedmineServiceOptions(&projectsv1alpha1.IntegrationRedmineParameters{
		ProjectID:   ptr.To[int64](123),
		ProjectURL:  "https://redmine.example.com/project",
		IssuesURL:   "https://redmine.example.com/issues/:id",
		NewIssueURL: "https://redmine.example.com/issues/new",
	})
	want := &gitlab.SetRedmineServiceOptions{
		ProjectURL:  ptr.To("https://redmine.example.com/project"),
		IssuesURL:   ptr.To("https://redmine.example.com/issues/:id"),
		NewIssueURL: ptr.To("https://redmine.example.com/issues/new"),
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("GenerateSetRedmineServiceOptions(): -want, +got:\n%s", diff)
	}
}

func TestGenerateIntegrationRedmineObservation(t *testing.T) {
	cases := map[string]struct {
		service *gitlab.RedmineService
		want    projectsv1alpha1.IntegrationRedmineObservation
	}{
		"FullObservation": {
			service: &gitlab.RedmineService{
				Service: gitlab.Service{
					ID:     7,
					Title:  "Redmine",
					Slug:   "redmine",
					Active: true,
				},
				Properties: &gitlab.RedmineServiceProperties{
					ProjectURL:  "https://redmine.example.com/project",
					IssuesURL:   "https://redmine.example.com/issues/:id",
					NewIssueURL: "https://redmine.example.com/issues/new",
				},
			},
			want: projectsv1alpha1.IntegrationRedmineObservation{
				CommonIntegrationObservation: v1alpha1.CommonIntegrationObservation{
					ID:                             ptr.To[int64](7),
					Title:                          ptr.To("Redmine"),
					Slug:                           ptr.To("redmine"),
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
				ProjectURL:  "https://redmine.example.com/project",
				IssuesURL:   "https://redmine.example.com/issues/:id",
				NewIssueURL: "https://redmine.example.com/issues/new",
			},
		},
		"NilService": {
			service: nil,
			want:    projectsv1alpha1.IntegrationRedmineObservation{},
		},
		"NilProperties": {
			service: &gitlab.RedmineService{Service: gitlab.Service{ID: 7}},
			want:    projectsv1alpha1.IntegrationRedmineObservation{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateIntegrationRedmineObservation(tc.service)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateIntegrationRedmineObservation(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsIntegrationRedmineUpToDate(t *testing.T) {
	spec := &projectsv1alpha1.IntegrationRedmineParameters{
		ProjectURL:  "https://redmine.example.com/project",
		IssuesURL:   "https://redmine.example.com/issues/:id",
		NewIssueURL: "https://redmine.example.com/issues/new",
	}

	cases := map[string]struct {
		observation *gitlab.RedmineService
		want        bool
	}{
		"UpToDate": {
			observation: &gitlab.RedmineService{
				Properties: &gitlab.RedmineServiceProperties{
					ProjectURL:  "https://redmine.example.com/project",
					IssuesURL:   "https://redmine.example.com/issues/:id",
					NewIssueURL: "https://redmine.example.com/issues/new",
				},
			},
			want: true,
		},
		"OutOfDateProjectURL": {
			observation: &gitlab.RedmineService{
				Properties: &gitlab.RedmineServiceProperties{
					ProjectURL:  "https://other.example.com/project",
					IssuesURL:   "https://redmine.example.com/issues/:id",
					NewIssueURL: "https://redmine.example.com/issues/new",
				},
			},
			want: false,
		},
		"OutOfDateIssuesURL": {
			observation: &gitlab.RedmineService{
				Properties: &gitlab.RedmineServiceProperties{
					ProjectURL:  "https://redmine.example.com/project",
					IssuesURL:   "https://other.example.com/issues/:id",
					NewIssueURL: "https://redmine.example.com/issues/new",
				},
			},
			want: false,
		},
		"OutOfDateNewIssueURL": {
			observation: &gitlab.RedmineService{
				Properties: &gitlab.RedmineServiceProperties{
					ProjectURL:  "https://redmine.example.com/project",
					IssuesURL:   "https://redmine.example.com/issues/:id",
					NewIssueURL: "https://other.example.com/issues/new",
				},
			},
			want: false,
		},
		"NilProperties": {
			observation: &gitlab.RedmineService{},
			want:        false,
		},
		"NilObservation": {
			observation: nil,
			want:        false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsIntegrationRedmineUpToDate(spec, tc.observation)
			if got != tc.want {
				t.Errorf("IsIntegrationRedmineUpToDate(): got %v, want %v", got, tc.want)
			}
		})
	}
}
