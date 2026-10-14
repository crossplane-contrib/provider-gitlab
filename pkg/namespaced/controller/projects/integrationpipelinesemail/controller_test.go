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

package integrationpipelinesemail

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

	testProjectID  int64 = 123
	testIntegID    int64 = 456
	testRecipient        = "dev@example.com"
	testRecipients       = []string{testRecipient}
)

type args struct {
	client projects.PipelinesEmailClient
	kube   client.Client
	cr     resource.Managed
}

type modifier func(*v1alpha1.IntegrationPipelinesEmail)

func withProjectID(id int64) modifier {
	return func(r *v1alpha1.IntegrationPipelinesEmail) {
		r.Spec.ForProvider.ProjectID = &id
	}
}

func withRecipients(rs ...string) modifier {
	return func(r *v1alpha1.IntegrationPipelinesEmail) {
		r.Spec.ForProvider.Recipients = rs
	}
}

func withNotifyOnlyBrokenPipelines(v bool) modifier {
	return func(r *v1alpha1.IntegrationPipelinesEmail) {
		r.Spec.ForProvider.NotifyOnlyBrokenPipelines = &v
	}
}

func withBranchesToBeNotified(v string) modifier {
	return func(r *v1alpha1.IntegrationPipelinesEmail) {
		r.Spec.ForProvider.BranchesToBeNotified = &v
	}
}

func withPipelineEvents(v bool) modifier {
	return func(r *v1alpha1.IntegrationPipelinesEmail) {
		r.Spec.ForProvider.PipelineEvents = &v
	}
}

func withConditions(c ...v2.Condition) modifier {
	return func(cr *v1alpha1.IntegrationPipelinesEmail) {
		cr.Status.ConditionedStatus.Conditions = c
	}
}

func withStatus(s v1alpha1.IntegrationPipelinesEmailObservation) modifier {
	return func(r *v1alpha1.IntegrationPipelinesEmail) {
		r.Status.AtProvider = s
	}
}

func integrationPipelinesEmail(m ...modifier) *v1alpha1.IntegrationPipelinesEmail {
	cr := &v1alpha1.IntegrationPipelinesEmail{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func remoteService(recipients string) *gitlab.PipelinesEmailService {
	return &gitlab.PipelinesEmailService{
		Service: gitlab.Service{
			ID:             testIntegID,
			Title:          "Pipeline status emails",
			Slug:           "pipelines-email",
			Active:         true,
			PipelineEvents: true,
		},
		Properties: &gitlab.PipelinesEmailProperties{
			Recipients:                recipients,
			NotifyOnlyBrokenPipelines: true,
			BranchesToBeNotified:      "default",
		},
	}
}

func observation(recipients ...string) v1alpha1.IntegrationPipelinesEmailObservation {
	return v1alpha1.IntegrationPipelinesEmailObservation{
		CommonIntegrationObservation: commonv1alpha1.CommonIntegrationObservation{
			ID:                             ptr.To(testIntegID),
			Title:                          ptr.To("Pipeline status emails"),
			Slug:                           ptr.To("pipelines-email"),
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
			PipelineEvents:                 ptr.To(true),
			PushEvents:                     ptr.To(false),
			TagPushEvents:                  ptr.To(false),
			VulnerabilityEvents:            ptr.To(false),
			WikiPageEvents:                 ptr.To(false),
			CommentOnEventEnabled:          ptr.To(false),
			Inherited:                      ptr.To(false),
		},
		Recipients:                recipients,
		NotifyOnlyBrokenPipelines: true,
		BranchesToBeNotified:      "default",
	}
}

func TestConnect(t *testing.T) {
	type want struct {
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
				err: errors.New(errNotIntegrationPipelinesEmail),
			},
		},
		"ProviderConfigRefNotGivenError": {
			args: args{
				cr:   integrationPipelinesEmail(),
				kube: &test.MockClient{MockGet: test.NewMockGetFn(nil)},
			},
			want: want{
				err: errors.New("providerConfigRef is not given"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &connector{kube: tc.kube, newGitlabClientFn: nil}
			o, err := c.Connect(context.Background(), tc.args.cr)

			if err == nil {
				t.Fatalf("Connect(...): want error %q, got nil", tc.want.err)
			}
			if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
				t.Errorf("Connect(...): -want error, +got error:\n%s", diff)
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
				err: errors.New(errNotIntegrationPipelinesEmail),
			},
		},
		"NilProjectID": {
			args: args{
				cr: integrationPipelinesEmail(),
			},
			want: want{
				cr:  integrationPipelinesEmail(),
				err: errors.New(errProjectIDMissing),
			},
		},
		"NotFound": {
			args: args{
				client: &fake.MockClient{
					MockGetPipelinesEmailService: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.PipelinesEmailService, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errBoom
					},
				},
				cr: integrationPipelinesEmail(withProjectID(testProjectID), withRecipients(testRecipients...)),
			},
			want: want{
				cr:     integrationPipelinesEmail(withProjectID(testProjectID), withRecipients(testRecipients...)),
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"GetFailed": {
			args: args{
				client: &fake.MockClient{
					MockGetPipelinesEmailService: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.PipelinesEmailService, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 500}}, errBoom
					},
				},
				cr: integrationPipelinesEmail(withProjectID(testProjectID), withRecipients(testRecipients...)),
			},
			want: want{
				cr:  integrationPipelinesEmail(withProjectID(testProjectID), withRecipients(testRecipients...)),
				err: errors.Wrap(errBoom, errGetFailed),
			},
		},
		"InactiveIntegration": {
			args: args{
				client: &fake.MockClient{
					MockGetPipelinesEmailService: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.PipelinesEmailService, *gitlab.Response, error) {
						return &gitlab.PipelinesEmailService{Properties: &gitlab.PipelinesEmailProperties{}}, &gitlab.Response{}, nil
					},
				},
				cr: integrationPipelinesEmail(withProjectID(testProjectID), withRecipients(testRecipients...)),
			},
			want: want{
				cr:     integrationPipelinesEmail(withProjectID(testProjectID), withRecipients(testRecipients...)),
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"SuccessUpToDate": {
			args: args{
				client: &fake.MockClient{
					MockGetPipelinesEmailService: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.PipelinesEmailService, *gitlab.Response, error) {
						return remoteService(testRecipient), &gitlab.Response{}, nil
					},
				},
				cr: integrationPipelinesEmail(
					withProjectID(testProjectID),
					withRecipients(testRecipients...),
					withNotifyOnlyBrokenPipelines(true),
					withBranchesToBeNotified("default"),
					withPipelineEvents(true),
				),
			},
			want: want{
				cr: integrationPipelinesEmail(
					withProjectID(testProjectID),
					withRecipients(testRecipients...),
					withNotifyOnlyBrokenPipelines(true),
					withBranchesToBeNotified("default"),
					withPipelineEvents(true),
					withConditions(v2.Available()),
					withStatus(observation(testRecipient)),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"SuccessLateInitialized": {
			args: args{
				client: &fake.MockClient{
					MockGetPipelinesEmailService: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.PipelinesEmailService, *gitlab.Response, error) {
						return remoteService(testRecipient), &gitlab.Response{}, nil
					},
				},
				cr: integrationPipelinesEmail(withProjectID(testProjectID), withRecipients(testRecipients...)),
			},
			want: want{
				cr: integrationPipelinesEmail(
					withProjectID(testProjectID),
					withRecipients(testRecipients...),
					withNotifyOnlyBrokenPipelines(true),
					withBranchesToBeNotified("default"),
					withPipelineEvents(true),
					withConditions(v2.Available()),
					withStatus(observation(testRecipient)),
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
				client: &fake.MockClient{
					MockGetPipelinesEmailService: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.PipelinesEmailService, *gitlab.Response, error) {
						return remoteService("other@example.com"), &gitlab.Response{}, nil
					},
				},
				cr: integrationPipelinesEmail(
					withProjectID(testProjectID),
					withRecipients(testRecipients...),
					withNotifyOnlyBrokenPipelines(true),
					withBranchesToBeNotified("default"),
					withPipelineEvents(true),
				),
			},
			want: want{
				cr: integrationPipelinesEmail(
					withProjectID(testProjectID),
					withRecipients(testRecipients...),
					withNotifyOnlyBrokenPipelines(true),
					withBranchesToBeNotified("default"),
					withPipelineEvents(true),
					withConditions(v2.Available()),
					withStatus(observation("other@example.com")),
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
		opts *gitlab.SetPipelinesEmailServiceOptions
		err  error
	}

	var sent *gitlab.SetPipelinesEmailServiceOptions
	capture := func(err error) *fake.MockClient {
		return &fake.MockClient{
			MockSetPipelinesEmailService: func(pid any, opt *gitlab.SetPipelinesEmailServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelinesEmailService, *gitlab.Response, error) {
				sent = opt
				return &gitlab.PipelinesEmailService{}, &gitlab.Response{}, err
			},
		}
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
				err: errors.New(errNotIntegrationPipelinesEmail),
			},
		},
		"NilProjectID": {
			args: args{
				cr: integrationPipelinesEmail(),
			},
			want: want{
				cr:  integrationPipelinesEmail(),
				err: errors.New(errProjectIDMissing),
			},
		},
		"SuccessfulCreate": {
			args: args{
				client: capture(nil),
				cr: integrationPipelinesEmail(
					withProjectID(testProjectID),
					withRecipients("a@example.com", "b@example.com"),
					withBranchesToBeNotified("all"),
				),
			},
			want: want{
				cr: integrationPipelinesEmail(
					withProjectID(testProjectID),
					withRecipients("a@example.com", "b@example.com"),
					withBranchesToBeNotified("all"),
					withConditions(v2.Creating()),
				),
				opts: &gitlab.SetPipelinesEmailServiceOptions{
					Recipients:           ptr.To("a@example.com,b@example.com"),
					BranchesToBeNotified: ptr.To("all"),
				},
			},
		},
		"CreateFailed": {
			args: args{
				client: capture(errBoom),
				cr:     integrationPipelinesEmail(withProjectID(testProjectID), withRecipients(testRecipients...)),
			},
			want: want{
				cr: integrationPipelinesEmail(
					withProjectID(testProjectID),
					withRecipients(testRecipients...),
					withConditions(v2.Creating()),
				),
				opts: &gitlab.SetPipelinesEmailServiceOptions{Recipients: ptr.To(testRecipient)},
				err:  errors.Wrap(errBoom, errCreateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			sent = nil
			e := &external{client: tc.client}
			o, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Create(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("Create(...): -want CR, +got CR:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.opts, sent); diff != "" {
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
		opts *gitlab.SetPipelinesEmailServiceOptions
		err  error
	}

	var sent *gitlab.SetPipelinesEmailServiceOptions
	capture := func(err error) *fake.MockClient {
		return &fake.MockClient{
			MockSetPipelinesEmailService: func(pid any, opt *gitlab.SetPipelinesEmailServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelinesEmailService, *gitlab.Response, error) {
				sent = opt
				return &gitlab.PipelinesEmailService{}, &gitlab.Response{}, err
			},
		}
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
				err: errors.New(errNotIntegrationPipelinesEmail),
			},
		},
		"NilProjectID": {
			args: args{
				cr: integrationPipelinesEmail(),
			},
			want: want{
				err: errors.New(errProjectIDMissing),
			},
		},
		"SuccessfulUpdate": {
			args: args{
				client: capture(nil),
				cr: integrationPipelinesEmail(
					withProjectID(testProjectID),
					withRecipients(testRecipients...),
					withNotifyOnlyBrokenPipelines(false),
					withPipelineEvents(true),
				),
			},
			want: want{
				opts: &gitlab.SetPipelinesEmailServiceOptions{
					Recipients:                ptr.To(testRecipient),
					NotifyOnlyBrokenPipelines: ptr.To(false),
					PipelineEvents:            ptr.To(true),
				},
			},
		},
		"UpdateFailed": {
			args: args{
				client: capture(errBoom),
				cr:     integrationPipelinesEmail(withProjectID(testProjectID), withRecipients(testRecipients...)),
			},
			want: want{
				opts: &gitlab.SetPipelinesEmailServiceOptions{Recipients: ptr.To(testRecipient)},
				err:  errors.Wrap(errBoom, errUpdateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			sent = nil
			e := &external{client: tc.client}
			o, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Update(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.opts, sent); diff != "" {
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

	var deleted any
	capture := func(err error) *fake.MockClient {
		return &fake.MockClient{
			MockDeletePipelinesEmailService: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
				deleted = pid
				return &gitlab.Response{}, err
			},
		}
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
				err: errors.New(errNotIntegrationPipelinesEmail),
			},
		},
		"NilProjectID": {
			args: args{
				cr: integrationPipelinesEmail(),
			},
			want: want{
				err: errors.New(errProjectIDMissing),
			},
		},
		"SuccessfulDelete": {
			args: args{
				client: capture(nil),
				cr:     integrationPipelinesEmail(withProjectID(testProjectID), withRecipients(testRecipients...)),
			},
			want: want{
				pid: testProjectID,
			},
		},
		"DeleteFailed": {
			args: args{
				client: capture(errBoom),
				cr:     integrationPipelinesEmail(withProjectID(testProjectID), withRecipients(testRecipients...)),
			},
			want: want{
				pid: testProjectID,
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			deleted = nil
			e := &external{client: tc.client}
			o, err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Delete(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.pid, deleted); diff != "" {
				t.Errorf("Delete(...): -want project ID, +got project ID:\n%s", diff)
			}
			if diff := cmp.Diff(managed.ExternalDelete{}, o); diff != "" {
				t.Errorf("Delete(...): -want result, +got result:\n%s", diff)
			}
		})
	}
}
