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

func TestGenerateSetYouTrackServiceOptions(t *testing.T) {
	got := GenerateSetYouTrackServiceOptions(&projectsv1alpha1.IntegrationYouTrackParameters{
		ProjectID:  ptr.To[int64](123),
		ProjectURL: "https://youtrack.example.com/project",
		IssuesURL:  "https://youtrack.example.com/issues/:id",
	})
	want := &gitlab.SetYouTrackServiceOptions{
		ProjectURL: ptr.To("https://youtrack.example.com/project"),
		IssuesURL:  ptr.To("https://youtrack.example.com/issues/:id"),
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("GenerateSetYouTrackServiceOptions(): -want, +got:\n%s", diff)
	}
}

func TestGenerateIntegrationYouTrackObservation(t *testing.T) {
	cases := map[string]struct {
		service *gitlab.YouTrackService
		want    projectsv1alpha1.IntegrationYouTrackObservation
	}{
		"FullObservation": {
			service: &gitlab.YouTrackService{
				Service: gitlab.Service{
					ID:     7,
					Title:  "JetBrains YouTrack",
					Slug:   "youtrack",
					Active: true,
				},
				Properties: &gitlab.YouTrackServiceProperties{
					ProjectURL: "https://youtrack.example.com/project",
					IssuesURL:  "https://youtrack.example.com/issues/:id",
				},
			},
			want: projectsv1alpha1.IntegrationYouTrackObservation{
				CommonIntegrationObservation: v1alpha1.CommonIntegrationObservation{
					ID:                             ptr.To[int64](7),
					Title:                          ptr.To("JetBrains YouTrack"),
					Slug:                           ptr.To("youtrack"),
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
				ProjectURL: "https://youtrack.example.com/project",
				IssuesURL:  "https://youtrack.example.com/issues/:id",
			},
		},
		"NilService": {
			service: nil,
			want:    projectsv1alpha1.IntegrationYouTrackObservation{},
		},
		"NilProperties": {
			service: &gitlab.YouTrackService{Service: gitlab.Service{ID: 7}},
			want:    projectsv1alpha1.IntegrationYouTrackObservation{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateIntegrationYouTrackObservation(tc.service)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateIntegrationYouTrackObservation(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsIntegrationYouTrackUpToDate(t *testing.T) {
	spec := &projectsv1alpha1.IntegrationYouTrackParameters{
		ProjectURL: "https://youtrack.example.com/project",
		IssuesURL:  "https://youtrack.example.com/issues/:id",
	}

	cases := map[string]struct {
		observation *gitlab.YouTrackService
		want        bool
	}{
		"UpToDate": {
			observation: &gitlab.YouTrackService{
				Properties: &gitlab.YouTrackServiceProperties{
					ProjectURL: "https://youtrack.example.com/project",
					IssuesURL:  "https://youtrack.example.com/issues/:id",
				},
			},
			want: true,
		},
		"OutOfDateProjectURL": {
			observation: &gitlab.YouTrackService{
				Properties: &gitlab.YouTrackServiceProperties{
					ProjectURL: "https://other.example.com/project",
					IssuesURL:  "https://youtrack.example.com/issues/:id",
				},
			},
			want: false,
		},
		"OutOfDateIssuesURL": {
			observation: &gitlab.YouTrackService{
				Properties: &gitlab.YouTrackServiceProperties{
					ProjectURL: "https://youtrack.example.com/project",
					IssuesURL:  "https://other.example.com/issues/:id",
				},
			},
			want: false,
		},
		"NilProperties": {
			observation: &gitlab.YouTrackService{},
			want:        false,
		},
		"NilObservation": {
			observation: nil,
			want:        false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsIntegrationYouTrackUpToDate(spec, tc.observation)
			if got != tc.want {
				t.Errorf("IsIntegrationYouTrackUpToDate(): got %v, want %v", got, tc.want)
			}
		})
	}
}
