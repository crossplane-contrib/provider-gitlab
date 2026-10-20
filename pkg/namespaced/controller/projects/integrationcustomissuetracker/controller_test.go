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

package integrationcustomissuetracker

import (
	"context"
	"net/http"
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/errors"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
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

	testProjectID   int64 = 123
	testIntegID     int64 = 456
	testProjectURL        = "https://tracker.example.com/project"
	testIssuesURL         = "https://tracker.example.com/issues/:id"
	testNewIssueURL       = "https://tracker.example.com/issues/new"
)

type args struct {
	client projects.CustomIssueTrackerClient
	kube   client.Client
	cr     resource.Managed
}

type modifier func(*v1alpha1.IntegrationCustomIssueTracker)

func withProjectID(id int64) modifier {
	return func(r *v1alpha1.IntegrationCustomIssueTracker) {
		r.Spec.ForProvider.ProjectID = &id
	}
}

func withURLs(projectURL, issuesURL, newIssueURL string) modifier {
	return func(r *v1alpha1.IntegrationCustomIssueTracker) {
		r.Spec.ForProvider.ProjectURL = projectURL
		r.Spec.ForProvider.IssuesURL = issuesURL
		r.Spec.ForProvider.NewIssueURL = newIssueURL
	}
}

func withConditions(c ...v2.Condition) modifier {
	return func(cr *v1alpha1.IntegrationCustomIssueTracker) {
		cr.Status.ConditionedStatus.Conditions = c
	}
}

func withStatus(s v1alpha1.IntegrationCustomIssueTrackerObservation) modifier {
	return func(r *v1alpha1.IntegrationCustomIssueTracker) {
		r.Status.AtProvider = s
	}
}

func integrationCustomIssueTracker(m ...modifier) *v1alpha1.IntegrationCustomIssueTracker {
	cr := &v1alpha1.IntegrationCustomIssueTracker{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func activeService(projectURL string) *gitlab.CustomIssueTrackerService {
	return &gitlab.CustomIssueTrackerService{
		Service: gitlab.Service{
			ID:     testIntegID,
			Title:  "Custom issue tracker",
			Slug:   "custom-issue-tracker",
			Active: true,
		},
		Properties: &gitlab.CustomIssueTrackerServiceProperties{
			ProjectURL:  projectURL,
			IssuesURL:   testIssuesURL,
			NewIssueURL: testNewIssueURL,
		},
	}
}

func observation(projectURL string) v1alpha1.IntegrationCustomIssueTrackerObservation {
	return v1alpha1.IntegrationCustomIssueTrackerObservation{
		CommonIntegrationObservation: commonv1alpha1.CommonIntegrationObservation{
			ID:                             ptr.To(testIntegID),
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
		ProjectURL:  projectURL,
		IssuesURL:   testIssuesURL,
		NewIssueURL: testNewIssueURL,
	}
}

func TestConnect(t *testing.T) {
	cases := map[string]struct {
		args
		want error
	}{
		"InvalidInput": {
			args: args{cr: unexpectedItem},
			want: errors.New(errNotIntegrationCustomIssueTracker),
		},
		"ProviderConfigRefNotGivenError": {
			args: args{
				cr:   integrationCustomIssueTracker(),
				kube: &test.MockClient{MockGet: test.NewMockGetFn(nil)},
			},
			want: errors.New("providerConfigRef is not given"),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &connector{kube: tc.kube, newGitlabClientFn: nil}
			o, err := c.Connect(context.Background(), tc.args.cr)
			if err == nil || err.Error() != tc.want.Error() {
				t.Errorf("Connect(...): want error %q, got %v", tc.want, err)
			}
			if o != nil {
				t.Errorf("Connect(...): want nil client, got %v", o)
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
			args: args{cr: unexpectedItem},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errNotIntegrationCustomIssueTracker),
			},
		},
		"NilProjectID": {
			args: args{cr: integrationCustomIssueTracker()},
			want: want{
				cr:  integrationCustomIssueTracker(),
				err: errors.New(errProjectIDMissing),
			},
		},
		"NotFound": {
			args: args{
				client: &fake.MockCustomIssueTrackerClient{
					MockGetCustomIssueTrackerService: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.CustomIssueTrackerService, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errBoom
					},
				},
				cr: integrationCustomIssueTracker(withProjectID(testProjectID)),
			},
			want: want{
				cr:     integrationCustomIssueTracker(withProjectID(testProjectID)),
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"GetFailed": {
			args: args{
				client: &fake.MockCustomIssueTrackerClient{
					MockGetCustomIssueTrackerService: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.CustomIssueTrackerService, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusInternalServerError}}, errBoom
					},
				},
				cr: integrationCustomIssueTracker(withProjectID(testProjectID)),
			},
			want: want{
				cr:  integrationCustomIssueTracker(withProjectID(testProjectID)),
				err: errors.Wrap(errBoom, errGetFailed),
			},
		},
		"Inactive": {
			args: args{
				client: &fake.MockCustomIssueTrackerClient{
					MockGetCustomIssueTrackerService: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.CustomIssueTrackerService, *gitlab.Response, error) {
						return &gitlab.CustomIssueTrackerService{
							Service:    gitlab.Service{ID: testIntegID, Active: false},
							Properties: &gitlab.CustomIssueTrackerServiceProperties{},
						}, &gitlab.Response{}, nil
					},
				},
				cr: integrationCustomIssueTracker(withProjectID(testProjectID)),
			},
			want: want{
				cr:     integrationCustomIssueTracker(withProjectID(testProjectID)),
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"SuccessUpToDate": {
			args: args{
				client: &fake.MockCustomIssueTrackerClient{
					MockGetCustomIssueTrackerService: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.CustomIssueTrackerService, *gitlab.Response, error) {
						return activeService(testProjectURL), &gitlab.Response{}, nil
					},
				},
				cr: integrationCustomIssueTracker(
					withProjectID(testProjectID),
					withURLs(testProjectURL, testIssuesURL, testNewIssueURL),
				),
			},
			want: want{
				cr: integrationCustomIssueTracker(
					withProjectID(testProjectID),
					withURLs(testProjectURL, testIssuesURL, testNewIssueURL),
					withConditions(v2.Available()),
					withStatus(observation(testProjectURL)),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"SuccessNeedsUpdate": {
			args: args{
				client: &fake.MockCustomIssueTrackerClient{
					MockGetCustomIssueTrackerService: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.CustomIssueTrackerService, *gitlab.Response, error) {
						return activeService("https://drifted.example.com/project"), &gitlab.Response{}, nil
					},
				},
				cr: integrationCustomIssueTracker(
					withProjectID(testProjectID),
					withURLs(testProjectURL, testIssuesURL, testNewIssueURL),
				),
			},
			want: want{
				cr: integrationCustomIssueTracker(
					withProjectID(testProjectID),
					withURLs(testProjectURL, testIssuesURL, testNewIssueURL),
					withConditions(v2.Available()),
					withStatus(observation("https://drifted.example.com/project")),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.client}
			o, err := e.Observe(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Observe(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
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
		cr   resource.Managed
		pid  any
		opts *gitlab.SetCustomIssueTrackerServiceOptions
		err  error
	}

	cases := map[string]struct {
		args
		setErr error
		want
	}{
		"InvalidInput": {
			args: args{cr: unexpectedItem},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errNotIntegrationCustomIssueTracker),
			},
		},
		"NilProjectID": {
			args: args{cr: integrationCustomIssueTracker()},
			want: want{
				cr:  integrationCustomIssueTracker(),
				err: errors.New(errProjectIDMissing),
			},
		},
		"SuccessfulCreate": {
			args: args{
				cr: integrationCustomIssueTracker(
					withProjectID(testProjectID),
					withURLs(testProjectURL, testIssuesURL, testNewIssueURL),
				),
			},
			want: want{
				cr: integrationCustomIssueTracker(
					withProjectID(testProjectID),
					withURLs(testProjectURL, testIssuesURL, testNewIssueURL),
					withConditions(v2.Creating()),
				),
				pid: testProjectID,
				opts: &gitlab.SetCustomIssueTrackerServiceOptions{
					ProjectURL:  &testProjectURL,
					IssuesURL:   &testIssuesURL,
					NewIssueURL: &testNewIssueURL,
				},
			},
		},
		"CreateFailed": {
			args: args{
				cr: integrationCustomIssueTracker(
					withProjectID(testProjectID),
					withURLs(testProjectURL, testIssuesURL, testNewIssueURL),
				),
			},
			setErr: errBoom,
			want: want{
				cr: integrationCustomIssueTracker(
					withProjectID(testProjectID),
					withURLs(testProjectURL, testIssuesURL, testNewIssueURL),
					withConditions(v2.Creating()),
				),
				pid: testProjectID,
				opts: &gitlab.SetCustomIssueTrackerServiceOptions{
					ProjectURL:  &testProjectURL,
					IssuesURL:   &testIssuesURL,
					NewIssueURL: &testNewIssueURL,
				},
				err: errors.Wrap(errBoom, errCreateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var gotPID any
			var gotOpts *gitlab.SetCustomIssueTrackerServiceOptions
			e := &external{client: &fake.MockCustomIssueTrackerClient{
				MockSetCustomIssueTrackerService: func(pid any, opt *gitlab.SetCustomIssueTrackerServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.CustomIssueTrackerService, *gitlab.Response, error) {
					gotPID = pid
					gotOpts = opt
					return &gitlab.CustomIssueTrackerService{}, &gitlab.Response{}, tc.setErr
				},
			}}
			o, err := e.Create(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Create(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("Create(...): -want CR, +got CR:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.pid, gotPID); diff != "" {
				t.Errorf("Create(...): -want project ID, +got project ID:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.opts, gotOpts); diff != "" {
				t.Errorf("Create(...): -want options, +got options:\n%s", diff)
			}
			if diff := cmp.Diff(managed.ExternalCreation{}, o); diff != "" {
				t.Errorf("Create(...): -want result, +got result:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type want struct {
		pid  any
		opts *gitlab.SetCustomIssueTrackerServiceOptions
		err  error
	}

	cases := map[string]struct {
		args
		setErr error
		want
	}{
		"InvalidInput": {
			args: args{cr: unexpectedItem},
			want: want{err: errors.New(errNotIntegrationCustomIssueTracker)},
		},
		"NilProjectID": {
			args: args{cr: integrationCustomIssueTracker()},
			want: want{err: errors.New(errProjectIDMissing)},
		},
		"SuccessfulUpdate": {
			args: args{
				cr: integrationCustomIssueTracker(
					withProjectID(testProjectID),
					withURLs(testProjectURL, testIssuesURL, testNewIssueURL),
				),
			},
			want: want{
				pid: testProjectID,
				opts: &gitlab.SetCustomIssueTrackerServiceOptions{
					ProjectURL:  &testProjectURL,
					IssuesURL:   &testIssuesURL,
					NewIssueURL: &testNewIssueURL,
				},
			},
		},
		"UpdateFailed": {
			args: args{
				cr: integrationCustomIssueTracker(
					withProjectID(testProjectID),
					withURLs(testProjectURL, testIssuesURL, testNewIssueURL),
				),
			},
			setErr: errBoom,
			want: want{
				pid: testProjectID,
				opts: &gitlab.SetCustomIssueTrackerServiceOptions{
					ProjectURL:  &testProjectURL,
					IssuesURL:   &testIssuesURL,
					NewIssueURL: &testNewIssueURL,
				},
				err: errors.Wrap(errBoom, errUpdateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var gotPID any
			var gotOpts *gitlab.SetCustomIssueTrackerServiceOptions
			e := &external{client: &fake.MockCustomIssueTrackerClient{
				MockSetCustomIssueTrackerService: func(pid any, opt *gitlab.SetCustomIssueTrackerServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.CustomIssueTrackerService, *gitlab.Response, error) {
					gotPID = pid
					gotOpts = opt
					return &gitlab.CustomIssueTrackerService{}, &gitlab.Response{}, tc.setErr
				},
			}}
			o, err := e.Update(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Update(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.pid, gotPID); diff != "" {
				t.Errorf("Update(...): -want project ID, +got project ID:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.opts, gotOpts); diff != "" {
				t.Errorf("Update(...): -want options, +got options:\n%s", diff)
			}
			if diff := cmp.Diff(managed.ExternalUpdate{}, o); diff != "" {
				t.Errorf("Update(...): -want result, +got result:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type want struct {
		pid any
		err error
	}

	cases := map[string]struct {
		args
		deleteErr error
		want
	}{
		"InvalidInput": {
			args: args{cr: unexpectedItem},
			want: want{err: errors.New(errNotIntegrationCustomIssueTracker)},
		},
		"NilProjectID": {
			args: args{cr: integrationCustomIssueTracker()},
			want: want{err: errors.New(errProjectIDMissing)},
		},
		"SuccessfulDelete": {
			args: args{cr: integrationCustomIssueTracker(withProjectID(testProjectID))},
			want: want{pid: testProjectID},
		},
		"DeleteFailed": {
			args:      args{cr: integrationCustomIssueTracker(withProjectID(testProjectID))},
			deleteErr: errBoom,
			want: want{
				pid: testProjectID,
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var gotPID any
			e := &external{client: &fake.MockCustomIssueTrackerClient{
				MockDeleteCustomIssueTrackerService: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
					gotPID = pid
					return &gitlab.Response{}, tc.deleteErr
				},
			}}
			o, err := e.Delete(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Delete(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.pid, gotPID); diff != "" {
				t.Errorf("Delete(...): -want project ID, +got project ID:\n%s", diff)
			}
			if diff := cmp.Diff(managed.ExternalDelete{}, o); diff != "" {
				t.Errorf("Delete(...): -want result, +got result:\n%s", diff)
			}
		})
	}
}
