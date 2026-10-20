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

package integrationmicrosoftteams

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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	commonv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects/fake"
)

var (
	unexpectedItem resource.Managed
	errBoom        = errors.New("boom")

	testProjectID int64 = 123
	testIntegID   int64 = 789
	testWebhook         = "https://outlook.office.com/webhook/abc"
)

type teamsModifier func(*v1alpha1.IntegrationMicrosoftTeams)

func withProjectID(id int64) teamsModifier {
	return func(r *v1alpha1.IntegrationMicrosoftTeams) { r.Spec.ForProvider.ProjectID = &id }
}

func withWebhookSecretRef(name, key string) teamsModifier {
	return func(r *v1alpha1.IntegrationMicrosoftTeams) {
		r.Spec.ForProvider.WebhookSecretRef = v2.LocalSecretKeySelector{
			LocalSecretReference: v2.LocalSecretReference{Name: name},
			Key:                  key,
		}
	}
}

func withBranchesToBeNotified(b string) teamsModifier {
	return func(r *v1alpha1.IntegrationMicrosoftTeams) { r.Spec.ForProvider.BranchesToBeNotified = &b }
}

func withLateInit() teamsModifier {
	return func(r *v1alpha1.IntegrationMicrosoftTeams) {
		p := &r.Spec.ForProvider
		p.BranchesToBeNotified = ptr.To("default")
	}
}

func withConditions(c ...v2.Condition) teamsModifier {
	return func(r *v1alpha1.IntegrationMicrosoftTeams) { r.Status.ConditionedStatus.Conditions = c }
}

func withStatus(s v1alpha1.IntegrationMicrosoftTeamsObservation) teamsModifier {
	return func(r *v1alpha1.IntegrationMicrosoftTeams) { r.Status.AtProvider = s }
}

func integrationTeams(m ...teamsModifier) *v1alpha1.IntegrationMicrosoftTeams {
	cr := &v1alpha1.IntegrationMicrosoftTeams{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func mockKubeWithSecret(webhook string) client.Client {
	return &test.MockClient{
		MockGet: func(_ context.Context, _ client.ObjectKey, obj client.Object) error {
			secret, ok := obj.(*corev1.Secret)
			if !ok {
				return errors.Errorf("unexpected object type %T", obj)
			}
			secret.Data = map[string][]byte{"webhook": []byte(webhook)}
			return nil
		},
	}
}

func activeTeams() *gitlab.MicrosoftTeamsService {
	return &gitlab.MicrosoftTeamsService{
		Service: gitlab.Service{
			ID:             testIntegID,
			Title:          "Microsoft Teams notifications",
			Slug:           "microsoft-teams",
			Active:         true,
			PushEvents:     true,
			PipelineEvents: true,
		},
		Properties: &gitlab.MicrosoftTeamsServiceProperties{
			NotifyOnlyBrokenPipelines: true,
			BranchesToBeNotified:      "default",
		},
	}
}

func activeTeamsObservation() v1alpha1.IntegrationMicrosoftTeamsObservation {
	return v1alpha1.IntegrationMicrosoftTeamsObservation{
		CommonIntegrationObservation: commonv1alpha1.CommonIntegrationObservation{
			ID:                             ptr.To(testIntegID),
			Title:                          ptr.To("Microsoft Teams notifications"),
			Slug:                           ptr.To("microsoft-teams"),
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
			PushEvents:                     ptr.To(true),
			TagPushEvents:                  ptr.To(false),
			VulnerabilityEvents:            ptr.To(false),
			WikiPageEvents:                 ptr.To(false),
			CommentOnEventEnabled:          ptr.To(false),
			Inherited:                      ptr.To(false),
		},
		NotifyOnlyBrokenPipelines: true,
		BranchesToBeNotified:      "default",
	}
}

func checkErr(t *testing.T, method string, want, got error) {
	t.Helper()
	if (want == nil) != (got == nil) || (want != nil && want.Error() != got.Error()) {
		t.Errorf("%s(...): -want error, +got error: \nwant: %v\ngot: %v", method, want, got)
	}
}

func TestConnect(t *testing.T) {
	cases := map[string]struct {
		kube client.Client
		cr   resource.Managed
		err  error
	}{
		"InvalidInput": {
			cr:  unexpectedItem,
			err: errors.New(errNotIntegrationMicrosoftTeams),
		},
		"ProviderConfigRefNotGivenError": {
			kube: &test.MockClient{MockGet: test.NewMockGetFn(nil)},
			cr:   integrationTeams(),
			err:  errors.New("providerConfigRef is not given"),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &connector{kube: tc.kube}
			o, err := c.Connect(context.Background(), tc.cr)
			checkErr(t, "Connect", tc.err, err)
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

	get := func(svc *gitlab.MicrosoftTeamsService, status int, err error) *fake.MockMicrosoftTeamsClient {
		return &fake.MockMicrosoftTeamsClient{
			MockGetMicrosoftTeamsService: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.MicrosoftTeamsService, *gitlab.Response, error) {
				if pid != testProjectID {
					return nil, nil, errors.Errorf("unexpected project id %v", pid)
				}
				return svc, &gitlab.Response{Response: &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}}}}, err
			},
		}
	}
	inactive := activeTeams()
	inactive.Active = false
	drifted := activeTeams()
	drifted.Properties.BranchesToBeNotified = "all"
	driftedObs := activeTeamsObservation()
	driftedObs.BranchesToBeNotified = "all"

	cases := map[string]struct {
		client *fake.MockMicrosoftTeamsClient
		cr     resource.Managed
		want   want
	}{
		"InvalidInput": {
			cr:   unexpectedItem,
			want: want{cr: unexpectedItem, err: errors.New(errNotIntegrationMicrosoftTeams)},
		},
		"NilProjectID": {
			cr:   integrationTeams(),
			want: want{cr: integrationTeams(), err: errors.New(errProjectIDMissing)},
		},
		"NotFound": {
			client: get(nil, 404, errBoom),
			cr:     integrationTeams(withProjectID(testProjectID)),
			want:   want{cr: integrationTeams(withProjectID(testProjectID))},
		},
		"GetFailed": {
			client: get(nil, 500, errBoom),
			cr:     integrationTeams(withProjectID(testProjectID)),
			want: want{
				cr:  integrationTeams(withProjectID(testProjectID)),
				err: errors.Wrap(errBoom, errGetFailed),
			},
		},
		"Inactive": {
			client: get(inactive, 200, nil),
			cr:     integrationTeams(withProjectID(testProjectID)),
			want:   want{cr: integrationTeams(withProjectID(testProjectID))},
		},
		"SuccessUpToDate": {
			client: get(activeTeams(), 200, nil),
			cr:     integrationTeams(withProjectID(testProjectID), withLateInit()),
			want: want{
				cr: integrationTeams(
					withProjectID(testProjectID),
					withLateInit(),
					withConditions(v2.Available()),
					withStatus(activeTeamsObservation()),
				),
				result: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true},
			},
		},
		"SuccessLateInitialized": {
			client: get(activeTeams(), 200, nil),
			cr:     integrationTeams(withProjectID(testProjectID)),
			want: want{
				cr: integrationTeams(
					withProjectID(testProjectID),
					withLateInit(),
					withConditions(v2.Available()),
					withStatus(activeTeamsObservation()),
				),
				result: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true, ResourceLateInitialized: true},
			},
		},
		"SuccessNeedsUpdate": {
			client: get(drifted, 200, nil),
			cr:     integrationTeams(withProjectID(testProjectID), withLateInit()),
			want: want{
				cr: integrationTeams(
					withProjectID(testProjectID),
					withLateInit(),
					withConditions(v2.Available()),
					withStatus(driftedObs),
				),
				result: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.client}
			o, err := e.Observe(context.Background(), tc.cr)
			checkErr(t, "Observe", tc.want.err, err)
			if diff := cmp.Diff(tc.want.cr, tc.cr, test.EquateConditions()); diff != "" {
				t.Errorf("Observe(...): -want CR, +got CR:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("Observe(...): -want result, +got result:\n%s", diff)
			}
		})
	}
}

func TestCreateAndUpdate(t *testing.T) {
	type want struct {
		opts *gitlab.SetMicrosoftTeamsServiceOptions
		err  error
	}
	base := []teamsModifier{withProjectID(testProjectID), withWebhookSecretRef("teams", "webhook")}

	cases := map[string]struct {
		kube    client.Client
		cr      *v1alpha1.IntegrationMicrosoftTeams
		setErr  error
		wantCrt want
		wantUpd want
	}{
		"WebhookFromSecret": {
			kube: mockKubeWithSecret(testWebhook),
			cr:   integrationTeams(append(base, withBranchesToBeNotified("all"))...),
			wantCrt: want{opts: &gitlab.SetMicrosoftTeamsServiceOptions{
				WebHook:              ptr.To(testWebhook),
				BranchesToBeNotified: ptr.To("all"),
			}},
			wantUpd: want{opts: &gitlab.SetMicrosoftTeamsServiceOptions{
				WebHook:              ptr.To(testWebhook),
				BranchesToBeNotified: ptr.To("all"),
			}},
		},
		"WebhookResolutionFailed": {
			kube:    &test.MockClient{MockGet: test.NewMockGetFn(errBoom)},
			cr:      integrationTeams(base...),
			wantCrt: want{err: errors.Wrap(errors.Wrap(errors.Wrap(errBoom, "Cannot find referenced secret"), errWebhookMissing), errCreateFailed)},
			wantUpd: want{err: errors.Wrap(errors.Wrap(errors.Wrap(errBoom, "Cannot find referenced secret"), errWebhookMissing), errUpdateFailed)},
		},
		"SetFailed": {
			kube:   mockKubeWithSecret(testWebhook),
			cr:     integrationTeams(base...),
			setErr: errBoom,
			wantCrt: want{
				opts: &gitlab.SetMicrosoftTeamsServiceOptions{WebHook: ptr.To(testWebhook)},
				err:  errors.Wrap(errBoom, errCreateFailed),
			},
			wantUpd: want{
				opts: &gitlab.SetMicrosoftTeamsServiceOptions{WebHook: ptr.To(testWebhook)},
				err:  errors.Wrap(errBoom, errUpdateFailed),
			},
		},
		"NilProjectID": {
			cr:      integrationTeams(),
			wantCrt: want{err: errors.New(errProjectIDMissing)},
			wantUpd: want{err: errors.New(errProjectIDMissing)},
		},
	}

	for name, tc := range cases {
		var got *gitlab.SetMicrosoftTeamsServiceOptions
		mc := &fake.MockMicrosoftTeamsClient{
			MockSetMicrosoftTeamsService: func(pid any, opt *gitlab.SetMicrosoftTeamsServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.MicrosoftTeamsService, *gitlab.Response, error) {
				got = opt
				if pid != testProjectID {
					return nil, nil, errors.Errorf("unexpected project id %v", pid)
				}
				return &gitlab.MicrosoftTeamsService{}, &gitlab.Response{}, tc.setErr
			},
		}
		e := &external{client: mc, kube: tc.kube}

		t.Run(name+"/Create", func(t *testing.T) {
			got = nil
			_, err := e.Create(context.Background(), tc.cr.DeepCopy())
			checkErr(t, "Create", tc.wantCrt.err, err)
			if diff := cmp.Diff(tc.wantCrt.opts, got); diff != "" {
				t.Errorf("Create(...): -want opts, +got opts:\n%s", diff)
			}
		})
		t.Run(name+"/Update", func(t *testing.T) {
			got = nil
			_, err := e.Update(context.Background(), tc.cr.DeepCopy())
			checkErr(t, "Update", tc.wantUpd.err, err)
			if diff := cmp.Diff(tc.wantUpd.opts, got); diff != "" {
				t.Errorf("Update(...): -want opts, +got opts:\n%s", diff)
			}
		})
	}

	t.Run("InvalidInput", func(t *testing.T) {
		e := &external{}
		_, err := e.Create(context.Background(), unexpectedItem)
		checkErr(t, "Create", errors.New(errNotIntegrationMicrosoftTeams), err)
		_, err = e.Update(context.Background(), unexpectedItem)
		checkErr(t, "Update", errors.New(errNotIntegrationMicrosoftTeams), err)
	})

	t.Run("CreateSetsCreatingCondition", func(t *testing.T) {
		cr := integrationTeams(base...)
		e := &external{kube: mockKubeWithSecret(testWebhook), client: &fake.MockMicrosoftTeamsClient{
			MockSetMicrosoftTeamsService: func(pid any, opt *gitlab.SetMicrosoftTeamsServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.MicrosoftTeamsService, *gitlab.Response, error) {
				return &gitlab.MicrosoftTeamsService{}, &gitlab.Response{}, nil
			},
		}}
		if _, err := e.Create(context.Background(), cr); err != nil {
			t.Fatalf("Create(...): unexpected error: %v", err)
		}
		want := integrationTeams(append(base, withConditions(v2.Creating()))...)
		if diff := cmp.Diff(want, cr, test.EquateConditions()); diff != "" {
			t.Errorf("Create(...): -want CR, +got CR:\n%s", diff)
		}
	})
}

func TestDelete(t *testing.T) {
	cases := map[string]struct {
		cr  resource.Managed
		pid any
		err error
	}{
		"InvalidInput": {
			cr:  unexpectedItem,
			err: errors.New(errNotIntegrationMicrosoftTeams),
		},
		"NilProjectID": {
			cr:  integrationTeams(),
			err: errors.New(errProjectIDMissing),
		},
		"SuccessfulDelete": {
			cr:  integrationTeams(withProjectID(testProjectID)),
			pid: testProjectID,
		},
		"DeleteFailed": {
			cr:  integrationTeams(withProjectID(testProjectID)),
			pid: testProjectID,
			err: errors.Wrap(errBoom, errDeleteFailed),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var gotPID any
			e := &external{client: &fake.MockMicrosoftTeamsClient{
				MockDeleteMicrosoftTeamsService: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
					gotPID = pid
					if tc.err != nil {
						return &gitlab.Response{}, errBoom
					}
					return &gitlab.Response{}, nil
				},
			}}
			_, err := e.Delete(context.Background(), tc.cr)
			checkErr(t, "Delete", tc.err, err)
			if diff := cmp.Diff(tc.pid, gotPID); diff != "" {
				t.Errorf("Delete(...): -want pid, +got pid:\n%s", diff)
			}
		})
	}
}

func TestDisconnect(t *testing.T) {
	e := &external{}
	if err := e.Disconnect(context.Background()); err != nil {
		t.Errorf("Disconnect(...): unexpected error: %v", err)
	}
}
