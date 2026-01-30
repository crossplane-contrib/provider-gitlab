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

package common

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
)

// TestGenerateCommonIntegrationObservation tests the conversion from
// GitLab Service to CommonIntegrationObservation
func TestGenerateCommonIntegrationObservation(t *testing.T) {
	// Test data
	testID := 123
	testTitle := "Test Integration"
	testSlug := "test-integration"
	testCreatedAt := time.Now()
	testUpdatedAt := time.Now().Add(time.Hour)
	testActive := true
	testAlertEvents := true
	testCommitEvents := true
	testConfidentialIssuesEvents := true
	testConfidentialNoteEvents := true
	testDeploymentEvents := true
	testGroupConfidentialMentionEvents := true
	testGroupMentionEvents := true
	testIncidentEvents := true
	testIssuesEvents := true
	testJobEvents := true
	testMergeRequestsEvents := true
	testNoteEvents := true
	testPipelineEvents := true
	testPushEvents := true
	testTagPushEvents := true
	testVulnerabilityEvents := true
	testWikiPageEvents := true
	testCommentOnEventEnabled := true
	testInherited := true

	type args struct {
		integration *gitlab.Service
	}
	cases := map[string]struct {
		args args
		want v1alpha1.CommonIntegrationObservation
	}{
		"AllFieldsSet": {
			args: args{
				integration: &gitlab.Service{
					ID:                             testID,
					Title:                          testTitle,
					Slug:                           testSlug,
					CreatedAt:                      &testCreatedAt,
					UpdatedAt:                      &testUpdatedAt,
					Active:                         testActive,
					AlertEvents:                    testAlertEvents,
					CommitEvents:                   testCommitEvents,
					ConfidentialIssuesEvents:       testConfidentialIssuesEvents,
					ConfidentialNoteEvents:         testConfidentialNoteEvents,
					DeploymentEvents:               testDeploymentEvents,
					GroupConfidentialMentionEvents: testGroupConfidentialMentionEvents,
					GroupMentionEvents:             testGroupMentionEvents,
					IncidentEvents:                 testIncidentEvents,
					IssuesEvents:                   testIssuesEvents,
					JobEvents:                      testJobEvents,
					MergeRequestsEvents:            testMergeRequestsEvents,
					NoteEvents:                     testNoteEvents,
					PipelineEvents:                 testPipelineEvents,
					PushEvents:                     testPushEvents,
					TagPushEvents:                  testTagPushEvents,
					VulnerabilityEvents:            testVulnerabilityEvents,
					WikiPageEvents:                 testWikiPageEvents,
					CommentOnEventEnabled:          testCommentOnEventEnabled,
					Inherited:                      testInherited,
				},
			},
			want: v1alpha1.CommonIntegrationObservation{
				ID:                             testID,
				Title:                          testTitle,
				Slug:                           testSlug,
				CreatedAt:                      clients.TimeToMetaTime(&testCreatedAt),
				UpdatedAt:                      clients.TimeToMetaTime(&testUpdatedAt),
				Active:                         testActive,
				AlertEvents:                    testAlertEvents,
				CommitEvents:                   testCommitEvents,
				ConfidentialIssuesEvents:       testConfidentialIssuesEvents,
				ConfidentialNoteEvents:         testConfidentialNoteEvents,
				DeploymentEvents:               testDeploymentEvents,
				GroupConfidentialMentionEvents: testGroupConfidentialMentionEvents,
				GroupMentionEvents:             testGroupMentionEvents,
				IncidentEvents:                 testIncidentEvents,
				IssuesEvents:                   testIssuesEvents,
				JobEvents:                      testJobEvents,
				MergeRequestsEvents:            testMergeRequestsEvents,
				NoteEvents:                     testNoteEvents,
				PipelineEvents:                 testPipelineEvents,
				PushEvents:                     testPushEvents,
				TagPushEvents:                  testTagPushEvents,
				VulnerabilityEvents:            testVulnerabilityEvents,
				WikiPageEvents:                 testWikiPageEvents,
				CommentOnEventEnabled:          testCommentOnEventEnabled,
				Inherited:                      testInherited,
			},
		},
		"MinimalFields": {
			args: args{
				integration: &gitlab.Service{
					ID:    testID,
					Title: testTitle,
				},
			},
			want: v1alpha1.CommonIntegrationObservation{
				ID:    testID,
				Title: testTitle,
			},
		},
		"NilTimestamps": {
			args: args{
				integration: &gitlab.Service{
					ID:        testID,
					Title:     testTitle,
					CreatedAt: nil,
					UpdatedAt: nil,
				},
			},
			want: v1alpha1.CommonIntegrationObservation{
				ID:    testID,
				Title: testTitle,
			},
		},
		"OnlyBooleanEventsFalse": {
			args: args{
				integration: &gitlab.Service{
					ID:                       testID,
					Title:                    testTitle,
					Active:                   false,
					AlertEvents:              false,
					CommitEvents:             false,
					ConfidentialIssuesEvents: false,
					ConfidentialNoteEvents:   false,
					DeploymentEvents:         false,
					IssuesEvents:             false,
					JobEvents:                false,
					MergeRequestsEvents:      false,
					NoteEvents:               false,
					PipelineEvents:           false,
					PushEvents:               false,
					TagPushEvents:            false,
					VulnerabilityEvents:      false,
					WikiPageEvents:           false,
					CommentOnEventEnabled:    false,
					Inherited:                false,
				},
			},
			want: v1alpha1.CommonIntegrationObservation{
				ID:    testID,
				Title: testTitle,
			},
		},
		"MixedEventStates": {
			args: args{
				integration: &gitlab.Service{
					ID:                       testID,
					Title:                    testTitle,
					Active:                   true,
					AlertEvents:              false,
					CommitEvents:             true,
					ConfidentialIssuesEvents: false,
					PushEvents:               true,
					MergeRequestsEvents:      true,
					IssuesEvents:             false,
				},
			},
			want: v1alpha1.CommonIntegrationObservation{
				ID:                       testID,
				Title:                    testTitle,
				Active:                   true,
				AlertEvents:              false,
				CommitEvents:             true,
				ConfidentialIssuesEvents: false,
				PushEvents:               true,
				MergeRequestsEvents:      true,
				IssuesEvents:             false,
			},
		},
		"EmptySlugAndTitle": {
			args: args{
				integration: &gitlab.Service{
					ID:    testID,
					Title: "",
					Slug:  "",
				},
			},
			want: v1alpha1.CommonIntegrationObservation{
				ID:    testID,
				Title: "",
				Slug:  "",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateCommonIntegrationObservation(tc.args.integration)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateCommonIntegrationObservation(): -want, +got:\n%s", diff)
			}
		})
	}
}
