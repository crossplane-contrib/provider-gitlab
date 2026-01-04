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
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	projectsv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
)

func ptrBool(v bool) *bool       { return &v }
func ptrInt64(v int64) *int64    { return &v }
func ptrString(v string) *string { return &v }

var (
	// Common test variables for Mattermost integration
	testProjectID                 int64 = 123
	testWebHook                         = "https://mattermost.example.com/hooks/webhook123"
	testUsername                        = "gitlab-bot"
	testChannel                         = "general"
	testNotifyOnlyBrokenPipelines       = true
	testBranchesToBeNotified            = "all"
	testPushEvents                      = true
	testIssuesEvents                    = true
	testConfidentialIssuesEvents        = false
	testMergeRequestsEvents             = true
	testTagPushEvents                   = true
	testNoteEvents                      = true
	testConfidentialNoteEvents          = false
	testPipelineEvents                  = true
	testWikiPageEvents                  = false
	testPushChannel                     = "ci-cd"
	testIssueChannel                    = "issues"
	testConfidentialIssueChannel        = "confidential-issues"
	testMergeRequestChannel             = "merge-requests"
	testNoteChannel                     = "comments"
	testConfidentialNoteChannel         = "confidential-comments"
	testTagPushChannel                  = "releases"
	testPipelineChannel                 = "pipelines"
	testWikiPageChannel                 = "wiki"
	testVulnerabilityChannel            = "security"
	testID                        int64 = 456
	testTitle                           = "Mattermost"
	testSlug                            = "mattermost"
	testActive                          = true
	testAlertEvents                     = false
	testCommitEvents                    = true
	testGroupConfidentialMention        = false
	testGroupMentionEvents              = false
	testIncidentEvents                  = false
	testJobEvents                       = true
	testDeploymentEvents                = false
	testVulnerabilityEvents             = false
	testCommentOnEventEnabled           = false
	testInherited                       = false
	testCreatedAt                       = time.Now()
	testUpdatedAt                       = time.Now()
)

// TestGenerateSetMattermostServiceOptions tests the conversion from
// IntegrationMattermostParameters to GitLab SetMattermostServiceOptions
func TestGenerateSetMattermostServiceOptions(t *testing.T) {
	type args struct {
		parameters *projectsv1alpha1.IntegrationMattermostParameters
	}
	cases := map[string]struct {
		args args
		want *gitlab.SetMattermostServiceOptions
	}{
		"AllFieldsSet": {
			args: args{
				parameters: &projectsv1alpha1.IntegrationMattermostParameters{
					ProjectID:                 &testProjectID,
					WebHook:                   testWebHook,
					Username:                  &testUsername,
					Channel:                   &testChannel,
					NotifyOnlyBrokenPipelines: &testNotifyOnlyBrokenPipelines,
					BranchesToBeNotified:      &testBranchesToBeNotified,
					PushEvents:                &testPushEvents,
					IssuesEvents:              &testIssuesEvents,
					ConfidentialIssuesEvents:  &testConfidentialIssuesEvents,
					MergeRequestsEvents:       &testMergeRequestsEvents,
					TagPushEvents:             &testTagPushEvents,
					NoteEvents:                &testNoteEvents,
					ConfidentialNoteEvents:    &testConfidentialNoteEvents,
					PipelineEvents:            &testPipelineEvents,
					WikiPageEvents:            &testWikiPageEvents,
					PushChannel:               &testPushChannel,
					IssueChannel:              &testIssueChannel,
					ConfidentialIssueChannel:  &testConfidentialIssueChannel,
					MergeRequestChannel:       &testMergeRequestChannel,
					NoteChannel:               &testNoteChannel,
					ConfidentialNoteChannel:   &testConfidentialNoteChannel,
					TagPushChannel:            &testTagPushChannel,
					PipelineChannel:           &testPipelineChannel,
					WikiPageChannel:           &testWikiPageChannel,
				},
			},
			want: &gitlab.SetMattermostServiceOptions{
				WebHook:                   &testWebHook,
				Username:                  &testUsername,
				Channel:                   &testChannel,
				NotifyOnlyBrokenPipelines: &testNotifyOnlyBrokenPipelines,
				BranchesToBeNotified:      &testBranchesToBeNotified,
				PushEvents:                &testPushEvents,
				IssuesEvents:              &testIssuesEvents,
				ConfidentialIssuesEvents:  &testConfidentialIssuesEvents,
				MergeRequestsEvents:       &testMergeRequestsEvents,
				TagPushEvents:             &testTagPushEvents,
				NoteEvents:                &testNoteEvents,
				ConfidentialNoteEvents:    &testConfidentialNoteEvents,
				PipelineEvents:            &testPipelineEvents,
				WikiPageEvents:            &testWikiPageEvents,
				PushChannel:               &testPushChannel,
				IssueChannel:              &testIssueChannel,
				ConfidentialIssueChannel:  &testConfidentialIssueChannel,
				MergeRequestChannel:       &testMergeRequestChannel,
				NoteChannel:               &testNoteChannel,
				ConfidentialNoteChannel:   &testConfidentialNoteChannel,
				TagPushChannel:            &testTagPushChannel,
				PipelineChannel:           &testPipelineChannel,
				WikiPageChannel:           &testWikiPageChannel,
			},
		},
		"OnlyRequiredFields": {
			args: args{
				parameters: &projectsv1alpha1.IntegrationMattermostParameters{
					ProjectID: &testProjectID,
					WebHook:   testWebHook,
				},
			},
			want: &gitlab.SetMattermostServiceOptions{
				WebHook: &testWebHook,
			},
		},
		"NilInput": {
			args: args{
				parameters: nil,
			},
			want: &gitlab.SetMattermostServiceOptions{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateSetMattermostServiceOptions(tc.args.parameters)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateSetMattermostServiceOptions(): -want, +got:\n%s", diff)
			}
		})
	}
}

// TestGenerateIntegrationMattermostObservation tests the conversion from
// GitLab MattermostService to IntegrationMattermostObservation
func TestGenerateIntegrationMattermostObservation(t *testing.T) {
	type args struct {
		service *gitlab.MattermostService
	}
	cases := map[string]struct {
		args args
		want projectsv1alpha1.IntegrationMattermostObservation
	}{
		"FullObservation": {
			args: args{
				service: &gitlab.MattermostService{
					Service: gitlab.Service{
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
						GroupConfidentialMentionEvents: testGroupConfidentialMention,
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
					Properties: &gitlab.MattermostServiceProperties{
						WebHook:                   testWebHook,
						Username:                  testUsername,
						Channel:                   testChannel,
						NotifyOnlyBrokenPipelines: gitlab.BoolValue(testNotifyOnlyBrokenPipelines),
						BranchesToBeNotified:      testBranchesToBeNotified,
						PushChannel:               testPushChannel,
						IssueChannel:              testIssueChannel,
						ConfidentialIssueChannel:  testConfidentialIssueChannel,
						MergeRequestChannel:       testMergeRequestChannel,
						NoteChannel:               testNoteChannel,
						ConfidentialNoteChannel:   testConfidentialNoteChannel,
						TagPushChannel:            testTagPushChannel,
						PipelineChannel:           testPipelineChannel,
						WikiPageChannel:           testWikiPageChannel,
						VulnerabilityChannel:      testVulnerabilityChannel,
					},
				},
			},
			want: projectsv1alpha1.IntegrationMattermostObservation{
				CommonIntegrationObservation: v1alpha1.CommonIntegrationObservation{
					ID:                             ptrInt64(testID),
					Title:                          ptrString(testTitle),
					Slug:                           ptrString(testSlug),
					CreatedAt:                      clients.TimeToMetaTime(&testCreatedAt),
					UpdatedAt:                      clients.TimeToMetaTime(&testUpdatedAt),
					Active:                         ptrBool(true),
					AlertEvents:                    ptrBool(false),
					CommitEvents:                   ptrBool(true),
					ConfidentialIssuesEvents:       ptrBool(false),
					ConfidentialNoteEvents:         ptrBool(false),
					DeploymentEvents:               ptrBool(false),
					GroupConfidentialMentionEvents: ptrBool(false),
					GroupMentionEvents:             ptrBool(false),
					IncidentEvents:                 ptrBool(false),
					IssuesEvents:                   ptrBool(true),
					JobEvents:                      ptrBool(true),
					MergeRequestsEvents:            ptrBool(true),
					NoteEvents:                     ptrBool(true),
					PipelineEvents:                 ptrBool(true),
					PushEvents:                     ptrBool(true),
					TagPushEvents:                  ptrBool(true),
					VulnerabilityEvents:            ptrBool(false),
					WikiPageEvents:                 ptrBool(false),
					CommentOnEventEnabled:          ptrBool(false),
					Inherited:                      ptrBool(false),
				},
				WebHook:                   testWebHook,
				Username:                  testUsername,
				Channel:                   testChannel,
				NotifyOnlyBrokenPipelines: testNotifyOnlyBrokenPipelines,
				BranchesToBeNotified:      testBranchesToBeNotified,
				PushChannel:               testPushChannel,
				IssueChannel:              testIssueChannel,
				ConfidentialIssueChannel:  testConfidentialIssueChannel,
				MergeRequestChannel:       testMergeRequestChannel,
				NoteChannel:               testNoteChannel,
				ConfidentialNoteChannel:   testConfidentialNoteChannel,
				TagPushChannel:            testTagPushChannel,
				PipelineChannel:           testPipelineChannel,
				WikiPageChannel:           testWikiPageChannel,
				VulnerabilityChannel:      testVulnerabilityChannel,
			},
		},
		"MinimalObservation": {
			args: args{
				service: &gitlab.MattermostService{
					Service: gitlab.Service{
						ID:    testID,
						Title: testTitle,
					},
					Properties: &gitlab.MattermostServiceProperties{
						WebHook: testWebHook,
					},
				},
			},
			want: projectsv1alpha1.IntegrationMattermostObservation{
				CommonIntegrationObservation: v1alpha1.CommonIntegrationObservation{
					ID:                             ptrInt64(testID),
					Title:                          ptrString(testTitle),
					Slug:                           ptrString(""),
					CreatedAt:                      nil,
					UpdatedAt:                      nil,
					Active:                         ptrBool(false),
					AlertEvents:                    ptrBool(false),
					CommitEvents:                   ptrBool(false),
					ConfidentialIssuesEvents:       ptrBool(false),
					ConfidentialNoteEvents:         ptrBool(false),
					DeploymentEvents:               ptrBool(false),
					GroupConfidentialMentionEvents: ptrBool(false),
					GroupMentionEvents:             ptrBool(false),
					IncidentEvents:                 ptrBool(false),
					IssuesEvents:                   ptrBool(false),
					JobEvents:                      ptrBool(false),
					MergeRequestsEvents:            ptrBool(false),
					NoteEvents:                     ptrBool(false),
					PipelineEvents:                 ptrBool(false),
					PushEvents:                     ptrBool(false),
					TagPushEvents:                  ptrBool(false),
					VulnerabilityEvents:            ptrBool(false),
					WikiPageEvents:                 ptrBool(false),
					CommentOnEventEnabled:          ptrBool(false),
					Inherited:                      ptrBool(false),
				},
				WebHook: testWebHook,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateIntegrationMattermostObservation(tc.args.service)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateIntegrationMattermostObservation(): -want, +got:\n%s", diff)
			}
		})
	}
}

// TestIsIntegrationMattermostUpToDate tests whether the spec matches the observation
func TestIsIntegrationMattermostUpToDate(t *testing.T) {
	type args struct {
		spec        *projectsv1alpha1.IntegrationMattermostParameters
		observation *gitlab.MattermostService
	}
	cases := map[string]struct {
		args args
		want bool
	}{
		"UpToDate": {
			args: args{
				spec: &projectsv1alpha1.IntegrationMattermostParameters{
					ProjectID:                 &testProjectID,
					WebHook:                   testWebHook,
					Username:                  &testUsername,
					Channel:                   &testChannel,
					NotifyOnlyBrokenPipelines: &testNotifyOnlyBrokenPipelines,
					BranchesToBeNotified:      &testBranchesToBeNotified,
					PushEvents:                &testPushEvents,
					IssuesEvents:              &testIssuesEvents,
					ConfidentialIssuesEvents:  &testConfidentialIssuesEvents,
					MergeRequestsEvents:       &testMergeRequestsEvents,
					TagPushEvents:             &testTagPushEvents,
					NoteEvents:                &testNoteEvents,
					ConfidentialNoteEvents:    &testConfidentialNoteEvents,
					PipelineEvents:            &testPipelineEvents,
					WikiPageEvents:            &testWikiPageEvents,
					PushChannel:               &testPushChannel,
					IssueChannel:              &testIssueChannel,
					ConfidentialIssueChannel:  &testConfidentialIssueChannel,
					MergeRequestChannel:       &testMergeRequestChannel,
					NoteChannel:               &testNoteChannel,
					ConfidentialNoteChannel:   &testConfidentialNoteChannel,
					TagPushChannel:            &testTagPushChannel,
					PipelineChannel:           &testPipelineChannel,
					WikiPageChannel:           &testWikiPageChannel,
				},
				observation: &gitlab.MattermostService{
					Service: gitlab.Service{
						PushEvents:               testPushEvents,
						IssuesEvents:             testIssuesEvents,
						ConfidentialIssuesEvents: testConfidentialIssuesEvents,
						MergeRequestsEvents:      testMergeRequestsEvents,
						TagPushEvents:            testTagPushEvents,
						NoteEvents:               testNoteEvents,
						ConfidentialNoteEvents:   testConfidentialNoteEvents,
						PipelineEvents:           testPipelineEvents,
						WikiPageEvents:           testWikiPageEvents,
					},
					Properties: &gitlab.MattermostServiceProperties{
						WebHook:                   testWebHook,
						Username:                  testUsername,
						Channel:                   testChannel,
						NotifyOnlyBrokenPipelines: gitlab.BoolValue(testNotifyOnlyBrokenPipelines),
						BranchesToBeNotified:      testBranchesToBeNotified,
						PushChannel:               testPushChannel,
						IssueChannel:              testIssueChannel,
						ConfidentialIssueChannel:  testConfidentialIssueChannel,
						MergeRequestChannel:       testMergeRequestChannel,
						NoteChannel:               testNoteChannel,
						ConfidentialNoteChannel:   testConfidentialNoteChannel,
						TagPushChannel:            testTagPushChannel,
						PipelineChannel:           testPipelineChannel,
						WikiPageChannel:           testWikiPageChannel,
					},
				},
			},
			want: true,
		},
		"DifferentWebHookIsIgnored": {
			// GitLab API doesn't return webhook URLs for security reasons,
			// so different webhooks should NOT cause the resource to be out-of-date
			args: args{
				spec: &projectsv1alpha1.IntegrationMattermostParameters{
					WebHook: "https://mattermost.example.com/hooks/different",
				},
				observation: &gitlab.MattermostService{
					Properties: &gitlab.MattermostServiceProperties{
						WebHook: testWebHook, // This will always be empty from GitLab API
					},
				},
			},
			want: true, // Changed from false: WebHook field is intentionally ignored
		},
		"OutOfDateUsername": {
			args: args{
				spec: &projectsv1alpha1.IntegrationMattermostParameters{
					WebHook:  testWebHook,
					Username: func() *string { s := "different-user"; return &s }(),
				},
				observation: &gitlab.MattermostService{
					Properties: &gitlab.MattermostServiceProperties{
						WebHook:  testWebHook,
						Username: testUsername,
					},
				},
			},
			want: false,
		},
		"OutOfDateChannel": {
			args: args{
				spec: &projectsv1alpha1.IntegrationMattermostParameters{
					WebHook: testWebHook,
					Channel: func() *string { s := "different-channel"; return &s }(),
				},
				observation: &gitlab.MattermostService{
					Properties: &gitlab.MattermostServiceProperties{
						WebHook: testWebHook,
						Channel: testChannel,
					},
				},
			},
			want: false,
		},
		"OutOfDateNotifyOnlyBrokenPipelines": {
			args: args{
				spec: &projectsv1alpha1.IntegrationMattermostParameters{
					WebHook:                   testWebHook,
					NotifyOnlyBrokenPipelines: func() *bool { b := false; return &b }(),
				},
				observation: &gitlab.MattermostService{
					Properties: &gitlab.MattermostServiceProperties{
						WebHook:                   testWebHook,
						NotifyOnlyBrokenPipelines: gitlab.BoolValue(true),
					},
				},
			},
			want: false,
		},
		"OutOfDatePushEvents": {
			args: args{
				spec: &projectsv1alpha1.IntegrationMattermostParameters{
					WebHook:    testWebHook,
					PushEvents: func() *bool { b := false; return &b }(),
				},
				observation: &gitlab.MattermostService{
					Service: gitlab.Service{
						PushEvents: true,
					},
					Properties: &gitlab.MattermostServiceProperties{
						WebHook: testWebHook,
					},
				},
			},
			want: false,
		},
		"NilSpecFields": {
			args: args{
				spec: &projectsv1alpha1.IntegrationMattermostParameters{
					WebHook: testWebHook,
				},
				observation: &gitlab.MattermostService{
					Properties: &gitlab.MattermostServiceProperties{
						WebHook: testWebHook,
					},
				},
			},
			want: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsIntegrationMattermostUpToDate(tc.args.spec, tc.args.observation)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("IsIntegrationMattermostUpToDate(): -want, +got:\n%s", diff)
			}
		})
	}
}
