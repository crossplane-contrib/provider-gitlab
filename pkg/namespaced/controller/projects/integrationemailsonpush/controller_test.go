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

package integrationemailsonpush

import (
	"context"
	"net/http"
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
	"k8s.io/utils/ptr"

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
	testRecipients       = "dev@example.com ops@example.com"
)

type emailsOnPushModifier func(*v1alpha1.IntegrationEmailsOnPush)

func withProjectID(id int64) emailsOnPushModifier {
	return func(r *v1alpha1.IntegrationEmailsOnPush) {
		r.Spec.ForProvider.ProjectID = &id
	}
}

func withRecipients(s string) emailsOnPushModifier {
	return func(r *v1alpha1.IntegrationEmailsOnPush) {
		r.Spec.ForProvider.Recipients = s
	}
}

func withDisableDiffs(v bool) emailsOnPushModifier {
	return func(r *v1alpha1.IntegrationEmailsOnPush) {
		r.Spec.ForProvider.DisableDiffs = &v
	}
}

func withSendFromCommitterEmail(v bool) emailsOnPushModifier {
	return func(r *v1alpha1.IntegrationEmailsOnPush) {
		r.Spec.ForProvider.SendFromCommitterEmail = &v
	}
}

func withPushEvents(v bool) emailsOnPushModifier {
	return func(r *v1alpha1.IntegrationEmailsOnPush) {
		r.Spec.ForProvider.PushEvents = &v
	}
}

func withTagPushEvents(v bool) emailsOnPushModifier {
	return func(r *v1alpha1.IntegrationEmailsOnPush) {
		r.Spec.ForProvider.TagPushEvents = &v
	}
}

func withBranchesToBeNotified(s string) emailsOnPushModifier {
	return func(r *v1alpha1.IntegrationEmailsOnPush) {
		r.Spec.ForProvider.BranchesToBeNotified = &s
	}
}

func withConditions(c ...v2.Condition) emailsOnPushModifier {
	return func(cr *v1alpha1.IntegrationEmailsOnPush) {
		cr.Status.ConditionedStatus.Conditions = c
	}
}

func withStatus(s v1alpha1.IntegrationEmailsOnPushObservation) emailsOnPushModifier {
	return func(r *v1alpha1.IntegrationEmailsOnPush) {
		r.Status.AtProvider = s
	}
}

// errMsg compares errors by message because the controller and pkg/common use different error packages.
func errMsg(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func integrationEmailsOnPush(m ...emailsOnPushModifier) *v1alpha1.IntegrationEmailsOnPush {
	cr := &v1alpha1.IntegrationEmailsOnPush{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func fullSpec() []emailsOnPushModifier {
	return []emailsOnPushModifier{
		withProjectID(testProjectID),
		withRecipients(testRecipients),
		withDisableDiffs(false),
		withSendFromCommitterEmail(true),
		withPushEvents(true),
		withTagPushEvents(false),
		withBranchesToBeNotified("protected"),
	}
}

func remoteService(active bool, recipients string) *gitlab.EmailsOnPushService {
	return &gitlab.EmailsOnPushService{
		Service: gitlab.Service{
			ID:            testIntegID,
			Title:         "Emails on push",
			Slug:          "emails-on-push",
			Active:        active,
			PushEvents:    true,
			TagPushEvents: false,
		},
		Properties: &gitlab.EmailsOnPushServiceProperties{
			Recipients:             recipients,
			DisableDiffs:           false,
			SendFromCommitterEmail: true,
			BranchesToBeNotified:   "protected",
		},
	}
}

func observedStatus(recipients string) v1alpha1.IntegrationEmailsOnPushObservation {
	return v1alpha1.IntegrationEmailsOnPushObservation{
		CommonIntegrationObservation: commonv1alpha1.CommonIntegrationObservation{
			ID:                             ptr.To(testIntegID),
			Title:                          ptr.To("Emails on push"),
			Slug:                           ptr.To("emails-on-push"),
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
			PushEvents:                     ptr.To(true),
			TagPushEvents:                  ptr.To(false),
			VulnerabilityEvents:            ptr.To(false),
			WikiPageEvents:                 ptr.To(false),
			CommentOnEventEnabled:          ptr.To(false),
			Inherited:                      ptr.To(false),
		},
		Recipients:             recipients,
		DisableDiffs:           false,
		SendFromCommitterEmail: true,
		BranchesToBeNotified:   "protected",
	}
}

func TestConnect(t *testing.T) {
	cases := map[string]struct {
		cr      resource.Managed
		wantErr error
	}{
		"InvalidInput": {
			cr:      unexpectedItem,
			wantErr: errors.New(errNotIntegrationEmailsOnPush),
		},
		"ProviderConfigRefNotGivenError": {
			cr:      integrationEmailsOnPush(),
			wantErr: errors.New("providerConfigRef is not given"),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &connector{kube: &test.MockClient{MockGet: test.NewMockGetFn(nil)}}
			o, err := c.Connect(context.Background(), tc.cr)
			if diff := cmp.Diff(errMsg(tc.wantErr), errMsg(err)); diff != "" {
				t.Errorf("Connect(...): -want error, +got error:\n%s", diff)
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

	getReturns := func(svc *gitlab.EmailsOnPushService, status int, err error) projects.EmailsOnPushClient {
		return &fake.MockEmailsOnPushClient{
			MockGetEmailsOnPushService: func(pid any, _ ...gitlab.RequestOptionFunc) (*gitlab.EmailsOnPushService, *gitlab.Response, error) {
				if pid != testProjectID {
					return nil, nil, errors.Errorf("unexpected project id %v", pid)
				}
				return svc, &gitlab.Response{Response: &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}}}}, err
			},
		}
	}

	cases := map[string]struct {
		client projects.EmailsOnPushClient
		cr     resource.Managed
		want   want
	}{
		"InvalidInput": {
			cr: unexpectedItem,
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errNotIntegrationEmailsOnPush),
			},
		},
		"NilProjectID": {
			cr: integrationEmailsOnPush(),
			want: want{
				cr:  integrationEmailsOnPush(),
				err: errors.New(errProjectIDMissing),
			},
		},
		"NotFound": {
			client: getReturns(nil, http.StatusNotFound, errBoom),
			cr:     integrationEmailsOnPush(withProjectID(testProjectID), withRecipients(testRecipients)),
			want: want{
				cr:     integrationEmailsOnPush(withProjectID(testProjectID), withRecipients(testRecipients)),
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"GetFailed": {
			client: getReturns(nil, http.StatusInternalServerError, errBoom),
			cr:     integrationEmailsOnPush(withProjectID(testProjectID), withRecipients(testRecipients)),
			want: want{
				cr:  integrationEmailsOnPush(withProjectID(testProjectID), withRecipients(testRecipients)),
				err: errors.Wrap(errBoom, errGetFailed),
			},
		},
		"NilService": {
			client: getReturns(nil, http.StatusOK, nil),
			cr:     integrationEmailsOnPush(withProjectID(testProjectID), withRecipients(testRecipients)),
			want: want{
				cr:     integrationEmailsOnPush(withProjectID(testProjectID), withRecipients(testRecipients)),
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"InactiveServiceAfterDisable": {
			client: getReturns(remoteService(false, ""), http.StatusOK, nil),
			cr:     integrationEmailsOnPush(withProjectID(testProjectID), withRecipients(testRecipients)),
			want: want{
				cr:     integrationEmailsOnPush(withProjectID(testProjectID), withRecipients(testRecipients)),
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"NilProperties": {
			client: getReturns(&gitlab.EmailsOnPushService{Service: gitlab.Service{Active: true}}, http.StatusOK, nil),
			cr:     integrationEmailsOnPush(withProjectID(testProjectID), withRecipients(testRecipients)),
			want: want{
				cr:     integrationEmailsOnPush(withProjectID(testProjectID), withRecipients(testRecipients)),
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"SuccessUpToDate": {
			client: getReturns(remoteService(true, testRecipients), http.StatusOK, nil),
			cr:     integrationEmailsOnPush(fullSpec()...),
			want: want{
				cr: integrationEmailsOnPush(append(fullSpec(),
					withConditions(v2.Available()),
					withStatus(observedStatus(testRecipients)),
				)...),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"SuccessLateInitialized": {
			client: getReturns(remoteService(true, testRecipients), http.StatusOK, nil),
			cr:     integrationEmailsOnPush(withProjectID(testProjectID), withRecipients(testRecipients)),
			want: want{
				cr: integrationEmailsOnPush(append(fullSpec(),
					withConditions(v2.Available()),
					withStatus(observedStatus(testRecipients)),
				)...),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
				},
			},
		},
		"DriftedRecipients": {
			client: getReturns(remoteService(true, "intruder@example.com"), http.StatusOK, nil),
			cr:     integrationEmailsOnPush(fullSpec()...),
			want: want{
				cr: integrationEmailsOnPush(append(fullSpec(),
					withConditions(v2.Available()),
					withStatus(observedStatus("intruder@example.com")),
				)...),
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
			o, err := e.Observe(context.Background(), tc.cr)

			if diff := cmp.Diff(errMsg(tc.want.err), errMsg(err)); diff != "" {
				t.Errorf("Observe(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.cr, test.EquateConditions()); diff != "" {
				t.Errorf("Observe(...): -want CR, +got CR:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("Observe(...): -want result, +got result:\n%s", diff)
			}
		})
	}
}

// recordSet returns a client whose SetEmailsOnPushService stores its arguments in pid and opts.
func recordSet(pid *any, opts **gitlab.SetEmailsOnPushServiceOptions, err error) *fake.MockEmailsOnPushClient {
	return &fake.MockEmailsOnPushClient{
		MockSetEmailsOnPushService: func(p any, o *gitlab.SetEmailsOnPushServiceOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.EmailsOnPushService, *gitlab.Response, error) {
			*pid = p
			*opts = o
			return &gitlab.EmailsOnPushService{}, &gitlab.Response{}, err
		},
	}
}

func TestCreate(t *testing.T) {
	type want struct {
		cr      resource.Managed
		err     error
		gotPID  any
		gotOpts *gitlab.SetEmailsOnPushServiceOptions
	}

	cases := map[string]struct {
		cr     resource.Managed
		setErr error
		want   want
	}{
		"InvalidInput": {
			cr: unexpectedItem,
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errNotIntegrationEmailsOnPush),
			},
		},
		"NilProjectID": {
			cr: integrationEmailsOnPush(),
			want: want{
				cr:  integrationEmailsOnPush(),
				err: errors.New(errProjectIDMissing),
			},
		},
		"SuccessfulCreate": {
			cr: integrationEmailsOnPush(fullSpec()...),
			want: want{
				cr:     integrationEmailsOnPush(append(fullSpec(), withConditions(v2.Creating()))...),
				gotPID: testProjectID,
				gotOpts: &gitlab.SetEmailsOnPushServiceOptions{
					Recipients:             ptr.To(testRecipients),
					DisableDiffs:           ptr.To(false),
					SendFromCommitterEmail: ptr.To(true),
					PushEvents:             ptr.To(true),
					TagPushEvents:          ptr.To(false),
					BranchesToBeNotified:   ptr.To("protected"),
				},
			},
		},
		"CreateFailed": {
			cr:     integrationEmailsOnPush(withProjectID(testProjectID), withRecipients(testRecipients)),
			setErr: errBoom,
			want: want{
				cr:      integrationEmailsOnPush(withProjectID(testProjectID), withRecipients(testRecipients), withConditions(v2.Creating())),
				err:     errors.Wrap(errBoom, errCreateFailed),
				gotPID:  testProjectID,
				gotOpts: &gitlab.SetEmailsOnPushServiceOptions{Recipients: ptr.To(testRecipients)},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var gotPID any
			var gotOpts *gitlab.SetEmailsOnPushServiceOptions
			e := &external{client: recordSet(&gotPID, &gotOpts, tc.setErr)}
			o, err := e.Create(context.Background(), tc.cr)

			if diff := cmp.Diff(errMsg(tc.want.err), errMsg(err)); diff != "" {
				t.Errorf("Create(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.cr, test.EquateConditions()); diff != "" {
				t.Errorf("Create(...): -want CR, +got CR:\n%s", diff)
			}
			if diff := cmp.Diff(managed.ExternalCreation{}, o); diff != "" {
				t.Errorf("Create(...): -want result, +got result:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.gotPID, gotPID); diff != "" {
				t.Errorf("Create(...): -want project id, +got project id:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.gotOpts, gotOpts); diff != "" {
				t.Errorf("Create(...): -want options, +got options:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type want struct {
		cr      resource.Managed
		err     error
		gotPID  any
		gotOpts *gitlab.SetEmailsOnPushServiceOptions
	}

	cases := map[string]struct {
		cr     resource.Managed
		setErr error
		want   want
	}{
		"InvalidInput": {
			cr: unexpectedItem,
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errNotIntegrationEmailsOnPush),
			},
		},
		"NilProjectID": {
			cr: integrationEmailsOnPush(),
			want: want{
				cr:  integrationEmailsOnPush(),
				err: errors.New(errProjectIDMissing),
			},
		},
		"SuccessfulUpdate": {
			cr: integrationEmailsOnPush(withProjectID(testProjectID), withRecipients(testRecipients), withDisableDiffs(true)),
			want: want{
				cr:     integrationEmailsOnPush(withProjectID(testProjectID), withRecipients(testRecipients), withDisableDiffs(true)),
				gotPID: testProjectID,
				gotOpts: &gitlab.SetEmailsOnPushServiceOptions{
					Recipients:   ptr.To(testRecipients),
					DisableDiffs: ptr.To(true),
				},
			},
		},
		"UpdateFailed": {
			cr:     integrationEmailsOnPush(withProjectID(testProjectID), withRecipients(testRecipients)),
			setErr: errBoom,
			want: want{
				cr:      integrationEmailsOnPush(withProjectID(testProjectID), withRecipients(testRecipients)),
				err:     errors.Wrap(errBoom, errUpdateFailed),
				gotPID:  testProjectID,
				gotOpts: &gitlab.SetEmailsOnPushServiceOptions{Recipients: ptr.To(testRecipients)},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var gotPID any
			var gotOpts *gitlab.SetEmailsOnPushServiceOptions
			e := &external{client: recordSet(&gotPID, &gotOpts, tc.setErr)}
			o, err := e.Update(context.Background(), tc.cr)

			if diff := cmp.Diff(errMsg(tc.want.err), errMsg(err)); diff != "" {
				t.Errorf("Update(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.cr, test.EquateConditions()); diff != "" {
				t.Errorf("Update(...): -want CR, +got CR:\n%s", diff)
			}
			if diff := cmp.Diff(managed.ExternalUpdate{}, o); diff != "" {
				t.Errorf("Update(...): -want result, +got result:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.gotPID, gotPID); diff != "" {
				t.Errorf("Update(...): -want project id, +got project id:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.gotOpts, gotOpts); diff != "" {
				t.Errorf("Update(...): -want options, +got options:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type want struct {
		err    error
		gotPID any
	}

	cases := map[string]struct {
		cr        resource.Managed
		deleteErr error
		want      want
	}{
		"InvalidInput": {
			cr:   unexpectedItem,
			want: want{err: errors.New(errNotIntegrationEmailsOnPush)},
		},
		"NilProjectID": {
			cr:   integrationEmailsOnPush(),
			want: want{err: errors.New(errProjectIDMissing)},
		},
		"SuccessfulDelete": {
			cr:   integrationEmailsOnPush(withProjectID(testProjectID), withRecipients(testRecipients)),
			want: want{gotPID: testProjectID},
		},
		"DeleteFailed": {
			cr:        integrationEmailsOnPush(withProjectID(testProjectID), withRecipients(testRecipients)),
			deleteErr: errBoom,
			want:      want{err: errors.Wrap(errBoom, errDeleteFailed), gotPID: testProjectID},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var gotPID any
			e := &external{client: &fake.MockEmailsOnPushClient{
				MockDeleteEmailsOnPushService: func(pid any, _ ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
					gotPID = pid
					return &gitlab.Response{}, tc.deleteErr
				},
			}}
			o, err := e.Delete(context.Background(), tc.cr)

			if diff := cmp.Diff(errMsg(tc.want.err), errMsg(err)); diff != "" {
				t.Errorf("Delete(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(managed.ExternalDelete{}, o); diff != "" {
				t.Errorf("Delete(...): -want result, +got result:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.gotPID, gotPID); diff != "" {
				t.Errorf("Delete(...): -want project id, +got project id:\n%s", diff)
			}
		})
	}
}
