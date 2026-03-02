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

package integrationmattermost

import (
	"context"
	"net/http"
	"testing"
	"time"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	commonv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects/fake"
)

var (
	unexpectedItem resource.Managed
	errBoom        = errors.New("boom")

	// Test data
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
	testCreatedAt                       = time.Now()
	testUpdatedAt                       = time.Now()
)

type args struct {
	mattermostClient projects.MattermostClient
	kube             client.Client
	cr               resource.Managed
}

type mattermostModifier func(*v1alpha1.IntegrationMattermost)

func withProjectID(id int64) mattermostModifier {
	return func(r *v1alpha1.IntegrationMattermost) {
		r.Spec.ForProvider.ProjectID = &id
	}
}

func withWebHook(webhook string) mattermostModifier {
	return func(r *v1alpha1.IntegrationMattermost) {
		r.Spec.ForProvider.WebHook = webhook
	}
}

func withUsername(username string) mattermostModifier {
	return func(r *v1alpha1.IntegrationMattermost) {
		r.Spec.ForProvider.Username = &username
	}
}

func withChannel(channel string) mattermostModifier {
	return func(r *v1alpha1.IntegrationMattermost) {
		r.Spec.ForProvider.Channel = &channel
	}
}

func withConditions(c ...xpv1.Condition) mattermostModifier {
	return func(cr *v1alpha1.IntegrationMattermost) {
		cr.Status.ConditionedStatus.Conditions = c
	}
}

func withStatus(s v1alpha1.IntegrationMattermostObservation) mattermostModifier {
	return func(r *v1alpha1.IntegrationMattermost) {
		r.Status.AtProvider = s
	}
}

func withDeletionTimestamp(t time.Time) mattermostModifier {
	return func(r *v1alpha1.IntegrationMattermost) {
		r.ObjectMeta.DeletionTimestamp = &metav1.Time{Time: t}
	}
}

// withForProvider allows customizing ForProvider fields in-place.
func withForProvider(mut func(*v1alpha1.IntegrationMattermostParameters)) mattermostModifier {
	return func(r *v1alpha1.IntegrationMattermost) {
		mut(&r.Spec.ForProvider)
	}
}

func integrationMattermost(m ...mattermostModifier) *v1alpha1.IntegrationMattermost {
	cr := &v1alpha1.IntegrationMattermost{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestConnect(t *testing.T) {
	type want struct {
		cr     resource.Managed
		result managed.ExternalClient
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"InvalidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errNotIntegrationMattermost),
			},
		},
		"ProviderConfigRefNotGivenError": {
			args: args{
				cr:   integrationMattermost(),
				kube: &test.MockClient{MockGet: test.NewMockGetFn(nil)},
			},
			want: want{
				cr:  integrationMattermost(),
				err: errors.New("providerConfigRef is not given"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &connector{kube: tc.kube, newGitlabClientFn: nil}
			o, err := c.Connect(context.Background(), tc.args.cr)

			if (tc.want.err == nil) != (err == nil) {
				t.Errorf("Connect(...): -want error, +got error: \nwant: %v\ngot: %v", tc.want.err, err)
			}
			if tc.want.err != nil && err != nil && tc.want.err.Error() != err.Error() {
				t.Errorf("Connect(...): -want error, +got error: \nwant: %v\ngot: %v", tc.want.err, err)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("Connect(...): -want result, +got result:\n%s", diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     resource.Managed
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"InvalidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errNotIntegrationMattermost),
			},
		},
		"NotFound": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockGetMattermostService: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.MattermostService, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errBoom
					},
				},
				cr: integrationMattermost(
					withProjectID(testProjectID),
					withWebHook(testWebHook),
				),
			},
			want: want{
				cr: integrationMattermost(
					withProjectID(testProjectID),
					withWebHook(testWebHook),
				),
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"GetFailed": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockGetMattermostService: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.MattermostService, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 500}}, errBoom
					},
				},
				cr: integrationMattermost(
					withProjectID(testProjectID),
					withWebHook(testWebHook),
				),
			},
			want: want{
				cr: integrationMattermost(
					withProjectID(testProjectID),
					withWebHook(testWebHook),
				),
				result: managed.ExternalObservation{},
				err:    errors.Wrap(errBoom, errGetFailed),
			},
		},
		"SuccessUpToDate": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockGetMattermostService: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.MattermostService, *gitlab.Response, error) {
						return &gitlab.MattermostService{
							Service: gitlab.Service{
								ID:                       testID,
								Title:                    testTitle,
								Slug:                     testSlug,
								CreatedAt:                &testCreatedAt,
								UpdatedAt:                &testUpdatedAt,
								Active:                   testActive,
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
								VulnerabilityChannel:      testVulnerabilityChannel,
							},
						}, &gitlab.Response{}, nil
					},
				},
				cr: integrationMattermost(
					withProjectID(testProjectID),
					withWebHook(testWebHook),
					withUsername(testUsername),
					withChannel(testChannel),
				),
			},
			want: want{
				cr: integrationMattermost(
					withProjectID(testProjectID),
					withWebHook(testWebHook),
					withUsername(testUsername),
					withChannel(testChannel),
					withForProvider(func(p *v1alpha1.IntegrationMattermostParameters) {
						p.NotifyOnlyBrokenPipelines = ptr.To(true)
						p.BranchesToBeNotified = ptr.To(testBranchesToBeNotified)
						p.PushEvents = ptr.To(true)
						p.IssuesEvents = ptr.To(true)
						p.ConfidentialIssuesEvents = ptr.To(false)
						p.MergeRequestsEvents = ptr.To(true)
						p.TagPushEvents = ptr.To(true)
						p.NoteEvents = ptr.To(true)
						p.ConfidentialNoteEvents = ptr.To(false)
						p.PipelineEvents = ptr.To(true)
						p.WikiPageEvents = ptr.To(false)

						p.PushChannel = ptr.To(testPushChannel)
						p.IssueChannel = ptr.To(testIssueChannel)
						p.ConfidentialIssueChannel = ptr.To(testConfidentialIssueChannel)
						p.MergeRequestChannel = ptr.To(testMergeRequestChannel)
						p.NoteChannel = ptr.To(testNoteChannel)
						p.ConfidentialNoteChannel = ptr.To(testConfidentialNoteChannel)
						p.TagPushChannel = ptr.To(testTagPushChannel)
						p.PipelineChannel = ptr.To(testPipelineChannel)
						p.WikiPageChannel = ptr.To(testWikiPageChannel)
					}),
					withConditions(xpv1.Available()),
					withStatus(v1alpha1.IntegrationMattermostObservation{
						CommonIntegrationObservation: commonv1alpha1.CommonIntegrationObservation{
							ID:                             ptr.To(testID),
							Title:                          ptr.To(testTitle),
							Slug:                           ptr.To(testSlug),
							CreatedAt:                      &metav1.Time{Time: testCreatedAt},
							UpdatedAt:                      &metav1.Time{Time: testUpdatedAt},
							Active:                         ptr.To(true),
							AlertEvents:                    ptr.To(false),
							CommitEvents:                   ptr.To(false),
							ConfidentialIssuesEvents:       ptr.To(false),
							ConfidentialNoteEvents:         ptr.To(false),
							DeploymentEvents:               ptr.To(false),
							GroupConfidentialMentionEvents: ptr.To(false),
							GroupMentionEvents:             ptr.To(false),
							IncidentEvents:                 ptr.To(false),
							IssuesEvents:                   ptr.To(true),
							JobEvents:                      ptr.To(false),
							MergeRequestsEvents:            ptr.To(true),
							NoteEvents:                     ptr.To(true),
							PipelineEvents:                 ptr.To(true),
							PushEvents:                     ptr.To(true),
							TagPushEvents:                  ptr.To(true),
							VulnerabilityEvents:            ptr.To(false),
							WikiPageEvents:                 ptr.To(false),
							CommentOnEventEnabled:          ptr.To(false),
							Inherited:                      ptr.To(false),
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
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
				},
			},
		},
		"SuccessNeedsUpdate": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockGetMattermostService: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.MattermostService, *gitlab.Response, error) {
						return &gitlab.MattermostService{
							Service: gitlab.Service{
								ID:         testID,
								Title:      testTitle,
								PushEvents: false,
							},
							Properties: &gitlab.MattermostServiceProperties{
								WebHook:  testWebHook,
								Username: "different-user",
							},
						}, &gitlab.Response{}, nil
					},
				},
				cr: integrationMattermost(
					withProjectID(testProjectID),
					withWebHook(testWebHook),
					withUsername(testUsername),
				),
			},
			want: want{
				cr: integrationMattermost(
					withProjectID(testProjectID),
					withWebHook(testWebHook),
					withUsername(testUsername),
					withForProvider(func(p *v1alpha1.IntegrationMattermostParameters) {
						p.NotifyOnlyBrokenPipelines = ptr.To(false)
						p.PushEvents = ptr.To(false)
						p.IssuesEvents = ptr.To(false)
						p.ConfidentialIssuesEvents = ptr.To(false)
						p.MergeRequestsEvents = ptr.To(false)
						p.TagPushEvents = ptr.To(false)
						p.NoteEvents = ptr.To(false)
						p.ConfidentialNoteEvents = ptr.To(false)
						p.PipelineEvents = ptr.To(false)
						p.WikiPageEvents = ptr.To(false)
					}),
					withConditions(xpv1.Available()),
					withStatus(v1alpha1.IntegrationMattermostObservation{
						CommonIntegrationObservation: commonv1alpha1.CommonIntegrationObservation{
							ID:                             ptr.To(testID),
							Title:                          ptr.To(testTitle),
							Slug:                           ptr.To(""),
							Active:                         ptr.To(false),
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
						WebHook:  testWebHook,
						Username: "different-user",
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: true,
				},
			},
		},
		"DeletingResource": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockGetMattermostService: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.MattermostService, *gitlab.Response, error) {
						return &gitlab.MattermostService{
							Service: gitlab.Service{
								ID:    testID,
								Title: testTitle,
							},
							Properties: &gitlab.MattermostServiceProperties{
								WebHook: testWebHook,
							},
						}, &gitlab.Response{}, nil
					},
				},
				cr: integrationMattermost(
					withProjectID(testProjectID),
					withWebHook(testWebHook),
					withDeletionTimestamp(time.Now()),
				),
			},
			want: want{
				cr: integrationMattermost(
					withProjectID(testProjectID),
					withWebHook(testWebHook),
					withDeletionTimestamp(time.Now()),
				),
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"LateInitialization": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockGetMattermostService: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.MattermostService, *gitlab.Response, error) {
						return &gitlab.MattermostService{
							Service: gitlab.Service{
								ID:         testID,
								Title:      testTitle,
								PushEvents: testPushEvents,
							},
							Properties: &gitlab.MattermostServiceProperties{
								WebHook:  testWebHook,
								Username: testUsername,
								Channel:  testChannel,
							},
						}, &gitlab.Response{}, nil
					},
				},
				cr: integrationMattermost(
					withProjectID(testProjectID),
					withWebHook(testWebHook),
				),
			},
			want: want{
				cr: integrationMattermost(
					withProjectID(testProjectID),
					withWebHook(testWebHook),
					withForProvider(func(p *v1alpha1.IntegrationMattermostParameters) {
						p.Username = ptr.To(testUsername)
						p.Channel = ptr.To(testChannel)

						p.NotifyOnlyBrokenPipelines = ptr.To(false)
						p.PushEvents = ptr.To(true)
						p.IssuesEvents = ptr.To(false)
						p.ConfidentialIssuesEvents = ptr.To(false)
						p.MergeRequestsEvents = ptr.To(false)
						p.TagPushEvents = ptr.To(false)
						p.NoteEvents = ptr.To(false)
						p.ConfidentialNoteEvents = ptr.To(false)
						p.PipelineEvents = ptr.To(false)
						p.WikiPageEvents = ptr.To(false)
					}),
					withConditions(xpv1.Available()),
					withStatus(v1alpha1.IntegrationMattermostObservation{
						CommonIntegrationObservation: commonv1alpha1.CommonIntegrationObservation{
							ID:                             ptr.To(testID),
							Title:                          ptr.To(testTitle),
							Slug:                           ptr.To(""),
							Active:                         ptr.To(false),
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
							PushEvents:                     ptr.To(true),
							TagPushEvents:                  ptr.To(false),
							VulnerabilityEvents:            ptr.To(false),
							WikiPageEvents:                 ptr.To(false),
							CommentOnEventEnabled:          ptr.To(false),
							Inherited:                      ptr.To(false),
						},
						WebHook:  testWebHook,
						Username: testUsername,
						Channel:  testChannel,
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.mattermostClient}
			o, err := e.Observe(context.Background(), tc.args.cr)

			if (tc.want.err == nil) != (err == nil) {
				t.Errorf("Observe(...): -want error, +got error: \nwant: %v\ngot: %v", tc.want.err, err)
			}
			if tc.want.err != nil && err != nil && tc.want.err.Error() != err.Error() {
				t.Errorf("Observe(...): -want error, +got error: \nwant: %v\ngot: %v", tc.want.err, err)
			}

			opts := []cmp.Option{test.EquateConditions()}
			if cr, ok := tc.args.cr.(*v1alpha1.IntegrationMattermost); ok && cr != nil && !cr.ObjectMeta.DeletionTimestamp.IsZero() {
				opts = append(opts, cmp.FilterPath(func(p cmp.Path) bool {
					return p.String() == "ObjectMeta.DeletionTimestamp"
				}, cmp.Ignore()))
			}

			if diff := cmp.Diff(tc.want.cr, tc.args.cr, opts...); diff != "" {
				t.Errorf("Observe(...): -want CR, +got CR:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("Observe(...): -want result, +got result:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type want struct {
		cr     resource.Managed
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"InvalidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errNotIntegrationMattermost),
			},
		},
		"SuccessfulCreate": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockSetMattermostService: func(pid any, opt *gitlab.SetMattermostServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.MattermostService, *gitlab.Response, error) {
						return &gitlab.MattermostService{
							Service: gitlab.Service{
								ID:    testID,
								Title: testTitle,
							},
							Properties: &gitlab.MattermostServiceProperties{
								WebHook: testWebHook,
							},
						}, &gitlab.Response{}, nil
					},
				},
				cr: integrationMattermost(
					withProjectID(testProjectID),
					withWebHook(testWebHook),
					withUsername(testUsername),
				),
			},
			want: want{
				cr: integrationMattermost(
					withProjectID(testProjectID),
					withWebHook(testWebHook),
					withUsername(testUsername),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalCreation{},
			},
		},
		"CreateFailed": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockSetMattermostService: func(pid any, opt *gitlab.SetMattermostServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.MattermostService, *gitlab.Response, error) {
						return nil, &gitlab.Response{}, errBoom
					},
				},
				cr: integrationMattermost(
					withProjectID(testProjectID),
					withWebHook(testWebHook),
				),
			},
			want: want{
				cr: integrationMattermost(
					withProjectID(testProjectID),
					withWebHook(testWebHook),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalCreation{},
				err:    errors.Wrap(errBoom, errCreateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.mattermostClient}
			o, err := e.Create(context.Background(), tc.args.cr)

			if (tc.want.err == nil) != (err == nil) {
				t.Errorf("Create(...): -want error, +got error: \nwant: %v\ngot: %v", tc.want.err, err)
			}
			if tc.want.err != nil && err != nil && tc.want.err.Error() != err.Error() {
				t.Errorf("Create(...): -want error, +got error: \nwant: %v\ngot: %v", tc.want.err, err)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("Create(...): -want CR, +got CR:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("Create(...): -want result, +got result:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type want struct {
		cr     resource.Managed
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"InvalidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errNotIntegrationMattermost),
			},
		},
		"SuccessfulUpdate": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockSetMattermostService: func(pid any, opt *gitlab.SetMattermostServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.MattermostService, *gitlab.Response, error) {
						return &gitlab.MattermostService{
							Service: gitlab.Service{
								ID:    testID,
								Title: testTitle,
							},
							Properties: &gitlab.MattermostServiceProperties{
								WebHook: testWebHook,
							},
						}, &gitlab.Response{}, nil
					},
				},
				cr: integrationMattermost(
					withProjectID(testProjectID),
					withWebHook(testWebHook),
					withUsername(testUsername),
				),
			},
			want: want{
				cr: integrationMattermost(
					withProjectID(testProjectID),
					withWebHook(testWebHook),
					withUsername(testUsername),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalUpdate{},
			},
		},
		"UpdateFailed": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockSetMattermostService: func(pid any, opt *gitlab.SetMattermostServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.MattermostService, *gitlab.Response, error) {
						return nil, &gitlab.Response{}, errBoom
					},
				},
				cr: integrationMattermost(
					withProjectID(testProjectID),
					withWebHook(testWebHook),
				),
			},
			want: want{
				cr: integrationMattermost(
					withProjectID(testProjectID),
					withWebHook(testWebHook),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalUpdate{},
				err:    errors.Wrap(errBoom, errUpdateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.mattermostClient}
			o, err := e.Update(context.Background(), tc.args.cr)

			if (tc.want.err == nil) != (err == nil) {
				t.Errorf("Update(...): -want error, +got error: \nwant: %v\ngot: %v", tc.want.err, err)
			}
			if tc.want.err != nil && err != nil && tc.want.err.Error() != err.Error() {
				t.Errorf("Update(...): -want error, +got error: \nwant: %v\ngot: %v", tc.want.err, err)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("Update(...): -want CR, +got CR:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("Update(...): -want result, +got result:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type want struct {
		cr     resource.Managed
		result managed.ExternalDelete
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"InvalidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errNotIntegrationMattermost),
			},
		},
		"SuccessfulDelete": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockDeleteMattermostService: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: integrationMattermost(
					withProjectID(testProjectID),
					withWebHook(testWebHook),
				),
			},
			want: want{
				cr: integrationMattermost(
					withProjectID(testProjectID),
					withWebHook(testWebHook),
				),
				result: managed.ExternalDelete{},
			},
		},
		"DeleteFailed": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockDeleteMattermostService: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: integrationMattermost(
					withProjectID(testProjectID),
					withWebHook(testWebHook),
				),
			},
			want: want{
				cr: integrationMattermost(
					withProjectID(testProjectID),
					withWebHook(testWebHook),
				),
				result: managed.ExternalDelete{},
				err:    errors.Wrap(errBoom, errDeleteFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.mattermostClient}
			o, err := e.Delete(context.Background(), tc.args.cr)

			if (tc.want.err == nil) != (err == nil) {
				t.Errorf("Delete(...): -want error, +got error: \nwant: %v\ngot: %v", tc.want.err, err)
			}
			if tc.want.err != nil && err != nil && tc.want.err.Error() != err.Error() {
				t.Errorf("Delete(...): -want error, +got error: \nwant: %v\ngot: %v", tc.want.err, err)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("Delete(...): -want CR, +got CR:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("Delete(...): -want result, +got result:\n%s", diff)
			}
		})
	}
}

func TestDisconnect(t *testing.T) {
	e := &external{}
	err := e.Disconnect(context.Background())
	if err != nil {
		t.Errorf("Disconnect(...): unexpected error: %v", err)
	}
}
