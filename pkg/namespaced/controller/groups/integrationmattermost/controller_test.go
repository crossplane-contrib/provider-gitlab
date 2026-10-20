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

	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	commonv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	groupclients "github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/groups"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/groups/fake"
)

var (
	unexpectedItem resource.Managed
	errBoom        = errors.New("boom")

	testGroupID int64 = 42
	testIntegID int64 = 999
	testChannel       = "gitlab-notifications"
	testWebHook       = "https://mattermost.example.com/hooks/abc123"
)

type args struct {
	mattermostClient groupclients.MattermostClient
	kube             client.Client
	cr               resource.Managed
}

type mattermostModifier func(*v1alpha1.IntegrationMattermost)

func withGroupID(id int64) mattermostModifier {
	return func(r *v1alpha1.IntegrationMattermost) {
		r.Spec.ForProvider.GroupID = &id
	}
}

func withChannel(c string) mattermostModifier {
	return func(r *v1alpha1.IntegrationMattermost) {
		r.Spec.ForProvider.Channel = &c
	}
}

func withWebHookSecretRef(name, key string) mattermostModifier {
	return func(r *v1alpha1.IntegrationMattermost) {
		r.Spec.ForProvider.WebHookSecretRef = &v2.LocalSecretKeySelector{
			LocalSecretReference: v2.LocalSecretReference{Name: name},
			Key:                  key,
		}
	}
}

func withConditions(c ...v2.Condition) mattermostModifier {
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

func withUseInheritedSettings(b bool) mattermostModifier {
	return func(r *v1alpha1.IntegrationMattermost) {
		r.Spec.ForProvider.UseInheritedSettings = &b
	}
}

func integrationMattermost(m ...mattermostModifier) *v1alpha1.IntegrationMattermost {
	cr := &v1alpha1.IntegrationMattermost{}
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

func activeMattermost() *groupclients.GroupMattermostIntegration {
	return &groupclients.GroupMattermostIntegration{
		Integration: gitlab.Integration{
			ID:     testIntegID,
			Title:  "Mattermost",
			Slug:   "mattermost",
			Active: true,
		},
		Properties: &groupclients.GroupMattermostIntegrationProperties{
			GroupMattermostIntegrationProperties: gitlab.GroupMattermostIntegrationProperties{Channel: testChannel},
		},
	}
}

func activeMattermostObservation() v1alpha1.IntegrationMattermostObservation {
	return v1alpha1.IntegrationMattermostObservation{
		CommonIntegrationObservation: commonv1alpha1.CommonIntegrationObservation{
			ID:                             ptr.To(testIntegID),
			Title:                          ptr.To("Mattermost"),
			Slug:                           ptr.To("mattermost"),
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
		Channel: testChannel,
	}
}

func inheritedMattermost(active bool) *groupclients.GroupMattermostIntegration {
	m := activeMattermost()
	m.Active = active
	m.Inherited = true
	return m
}

func inheritedMattermostObservation(active bool) v1alpha1.IntegrationMattermostObservation {
	o := activeMattermostObservation()
	o.Active = ptr.To(active)
	o.Inherited = ptr.To(true)
	return o
}

// checkSentOptions fails unless the webhook is sent and inherited settings
// are explicitly disabled.
func checkSentOptions(_ any, opt *gitlab.GroupMattermostIntegrationOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.GroupMattermostIntegration, *gitlab.Response, error) {
	if opt.WebHook == nil || *opt.WebHook != testWebHook {
		return nil, &gitlab.Response{}, errors.New("webhook was not sent")
	}
	if opt.UseInheritedSettings == nil || *opt.UseInheritedSettings {
		return nil, &gitlab.Response{}, errors.New("use_inherited_settings=false was not sent")
	}
	return &gitlab.GroupMattermostIntegration{Integration: gitlab.Integration{Active: true}}, &gitlab.Response{}, nil
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
		"DeletingResourceStillExists": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockGetGroupMattermostIntegration: func(gid any, options ...gitlab.RequestOptionFunc) (*groupclients.GroupMattermostIntegration, *gitlab.Response, error) {
						return &groupclients.GroupMattermostIntegration{
							Integration: gitlab.Integration{
								ID:     testIntegID,
								Title:  "Mattermost",
								Slug:   "mattermost",
								Active: true,
							},
							Properties: &groupclients.GroupMattermostIntegrationProperties{
								GroupMattermostIntegrationProperties: gitlab.GroupMattermostIntegrationProperties{Channel: testChannel},
							},
						}, &gitlab.Response{}, nil
					},
				},
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
					withDeletionTimestamp(time.Now()),
				),
			},
			want: want{
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
					withDeletionTimestamp(time.Now()),
					withConditions(v2.Available()),
					withStatus(v1alpha1.IntegrationMattermostObservation{
						CommonIntegrationObservation: commonv1alpha1.CommonIntegrationObservation{
							ID:                             ptr.To(testIntegID),
							Title:                          ptr.To("Mattermost"),
							Slug:                           ptr.To("mattermost"),
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
						Channel: testChannel,
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"DeletingResourceAlreadyGone": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockGetGroupMattermostIntegration: func(gid any, options ...gitlab.RequestOptionFunc) (*groupclients.GroupMattermostIntegration, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404, Header: http.Header{"Content-Type": {"application/json"}}}}, errBoom
					},
				},
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withDeletionTimestamp(time.Now()),
				),
			},
			want: want{
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withDeletionTimestamp(time.Now()),
				),
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"NilGroupID": {
			args: args{
				cr: integrationMattermost(),
			},
			want: want{
				cr:  integrationMattermost(),
				err: errors.New(errGroupIDMissing),
			},
		},
		"NotFound": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockGetGroupMattermostIntegration: func(gid any, options ...gitlab.RequestOptionFunc) (*groupclients.GroupMattermostIntegration, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404, Header: http.Header{"Content-Type": {"application/json"}}}}, errBoom
					},
				},
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
				),
			},
			want: want{
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
				),
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"GetFailed": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockGetGroupMattermostIntegration: func(gid any, options ...gitlab.RequestOptionFunc) (*groupclients.GroupMattermostIntegration, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 500}}, errBoom
					},
				},
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
				),
			},
			want: want{
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
				),
				result: managed.ExternalObservation{},
				err:    errors.Wrap(errBoom, errGetFailed),
			},
		},
		"NilIntegration": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockGetGroupMattermostIntegration: func(gid any, options ...gitlab.RequestOptionFunc) (*groupclients.GroupMattermostIntegration, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
					},
				},
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
				),
			},
			want: want{
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
				),
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"SuccessActive": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockGetGroupMattermostIntegration: func(gid any, options ...gitlab.RequestOptionFunc) (*groupclients.GroupMattermostIntegration, *gitlab.Response, error) {
						return activeMattermost(), &gitlab.Response{}, nil
					},
				},
				kube: mockKubeWithSecret(testWebHook),
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withWebHookSecretRef("my-secret", "webhook"),
					withChannel(testChannel),
				),
			},
			want: want{
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withWebHookSecretRef("my-secret", "webhook"),
					withChannel(testChannel),
					withConditions(v2.Available()),
					withStatus(activeMattermostObservation()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"SettingsDrift": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockGetGroupMattermostIntegration: func(gid any, options ...gitlab.RequestOptionFunc) (*groupclients.GroupMattermostIntegration, *gitlab.Response, error) {
						return activeMattermost(), &gitlab.Response{}, nil
					},
				},
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withWebHookSecretRef("my-secret", "webhook"),
					withChannel("other-channel"),
				),
			},
			want: want{
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withWebHookSecretRef("my-secret", "webhook"),
					withChannel("other-channel"),
					withConditions(v2.Available()),
					withStatus(activeMattermostObservation()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"InheritedAsRequested": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockGetGroupMattermostIntegration: func(gid any, options ...gitlab.RequestOptionFunc) (*groupclients.GroupMattermostIntegration, *gitlab.Response, error) {
						return inheritedMattermost(false), &gitlab.Response{}, nil
					},
				},
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel("ignored-while-inheriting"),
					withUseInheritedSettings(true),
				),
			},
			want: want{
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel("ignored-while-inheriting"),
					withUseInheritedSettings(true),
					withConditions(v2.Available()),
					withStatus(inheritedMattermostObservation(false)),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"InheritedNotRequested": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockGetGroupMattermostIntegration: func(gid any, options ...gitlab.RequestOptionFunc) (*groupclients.GroupMattermostIntegration, *gitlab.Response, error) {
						return inheritedMattermost(true), &gitlab.Response{}, nil
					},
				},
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
					withUseInheritedSettings(false),
				),
			},
			want: want{
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
					withUseInheritedSettings(false),
					withConditions(v2.Available()),
					withStatus(inheritedMattermostObservation(true)),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"InactiveInheritedNotRequested": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockGetGroupMattermostIntegration: func(gid any, options ...gitlab.RequestOptionFunc) (*groupclients.GroupMattermostIntegration, *gitlab.Response, error) {
						return inheritedMattermost(false), &gitlab.Response{}, nil
					},
				},
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
					withUseInheritedSettings(false),
				),
			},
			want: want{
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
					withUseInheritedSettings(false),
				),
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"InheritedLateInitialized": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockGetGroupMattermostIntegration: func(gid any, options ...gitlab.RequestOptionFunc) (*groupclients.GroupMattermostIntegration, *gitlab.Response, error) {
						return inheritedMattermost(true), &gitlab.Response{}, nil
					},
				},
				cr: integrationMattermost(withGroupID(testGroupID)),
			},
			want: want{
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withUseInheritedSettings(true),
					withConditions(v2.Available()),
					withStatus(inheritedMattermostObservation(true)),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
				},
			},
		},
		"DeletingInheritedIsGone": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockGetGroupMattermostIntegration: func(gid any, options ...gitlab.RequestOptionFunc) (*groupclients.GroupMattermostIntegration, *gitlab.Response, error) {
						return inheritedMattermost(true), &gitlab.Response{}, nil
					},
				},
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withUseInheritedSettings(true),
					withDeletionTimestamp(time.Now()),
				),
			},
			want: want{
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withUseInheritedSettings(true),
					withDeletionTimestamp(time.Now()),
				),
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"SuccessInactive": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockGetGroupMattermostIntegration: func(gid any, options ...gitlab.RequestOptionFunc) (*groupclients.GroupMattermostIntegration, *gitlab.Response, error) {
						return &groupclients.GroupMattermostIntegration{
							Integration: gitlab.Integration{
								ID:     testIntegID,
								Title:  "Mattermost",
								Slug:   "mattermost",
								Active: false,
							},
						}, &gitlab.Response{}, nil
					},
				},
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
				),
			},
			want: want{
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
				),
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.mattermostClient, kube: tc.kube}
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
		"NilGroupID": {
			args: args{
				cr: integrationMattermost(),
			},
			want: want{
				cr:  integrationMattermost(),
				err: errors.New(errGroupIDMissing),
			},
		},
		"WebHookResolutionFailed": {
			args: args{
				mattermostClient: &fake.MockClient{},
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(errBoom),
				},
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
					withWebHookSecretRef("my-secret", "webhook"),
				),
			},
			want: want{
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
					withWebHookSecretRef("my-secret", "webhook"),
					withConditions(v2.Creating()),
				),
				err: errors.Wrap(errors.Wrap(errors.Wrap(errBoom, "Cannot find referenced secret"), errWebHookMissing), errCreateFailed),
			},
		},
		"EmptyWebHook": {
			args: args{
				mattermostClient: &fake.MockClient{},
				kube:             mockKubeWithSecret(""),
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
					withWebHookSecretRef("my-secret", "webhook"),
				),
			},
			want: want{
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
					withWebHookSecretRef("my-secret", "webhook"),
					withConditions(v2.Creating()),
				),
				err: errors.Wrap(errors.New(errWebHookEmpty), errCreateFailed),
			},
		},
		"SuccessfulCreate": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockSetGroupMattermostIntegration: checkSentOptions,
				},
				kube: mockKubeWithSecret(testWebHook),
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
					withWebHookSecretRef("my-secret", "webhook"),
				),
			},
			want: want{
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
					withWebHookSecretRef("my-secret", "webhook"),
					withConditions(v2.Creating()),
				),
				result: managed.ExternalCreation{},
			},
		},
		"CreateFailed": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockSetGroupMattermostIntegration: func(gid any, opt *gitlab.GroupMattermostIntegrationOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMattermostIntegration, *gitlab.Response, error) {
						return nil, &gitlab.Response{}, errBoom
					},
				},
				kube: mockKubeWithSecret(testWebHook),
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
					withWebHookSecretRef("my-secret", "webhook"),
				),
			},
			want: want{
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
					withWebHookSecretRef("my-secret", "webhook"),
					withConditions(v2.Creating()),
				),
				result: managed.ExternalCreation{},
				err:    errors.Wrap(errBoom, errCreateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.mattermostClient, kube: tc.kube}
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
		"NilGroupID": {
			args: args{
				cr: integrationMattermost(),
			},
			want: want{
				cr:  integrationMattermost(),
				err: errors.New(errGroupIDMissing),
			},
		},
		"WebHookResolutionFailed": {
			args: args{
				mattermostClient: &fake.MockClient{},
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(errBoom),
				},
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
					withWebHookSecretRef("my-secret", "webhook"),
				),
			},
			want: want{
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
					withWebHookSecretRef("my-secret", "webhook"),
				),
				err: errors.Wrap(errors.Wrap(errors.Wrap(errBoom, "Cannot find referenced secret"), errWebHookMissing), errUpdateFailed),
			},
		},
		"EmptyWebHook": {
			args: args{
				mattermostClient: &fake.MockClient{},
				kube:             mockKubeWithSecret(""),
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
					withWebHookSecretRef("my-secret", "webhook"),
				),
			},
			want: want{
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
					withWebHookSecretRef("my-secret", "webhook"),
				),
				err: errors.Wrap(errors.New(errWebHookEmpty), errUpdateFailed),
			},
		},
		"SuccessfulUpdate": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockSetGroupMattermostIntegration: checkSentOptions,
				},
				kube: mockKubeWithSecret(testWebHook),
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
					withWebHookSecretRef("my-secret", "webhook"),
				),
			},
			want: want{
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
					withWebHookSecretRef("my-secret", "webhook"),
				),
				result: managed.ExternalUpdate{},
			},
		},
		"UpdateFailed": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockSetGroupMattermostIntegration: func(gid any, opt *gitlab.GroupMattermostIntegrationOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMattermostIntegration, *gitlab.Response, error) {
						return nil, &gitlab.Response{}, errBoom
					},
				},
				kube: mockKubeWithSecret(testWebHook),
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
					withWebHookSecretRef("my-secret", "webhook"),
				),
			},
			want: want{
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
					withWebHookSecretRef("my-secret", "webhook"),
				),
				result: managed.ExternalUpdate{},
				err:    errors.Wrap(errBoom, errUpdateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.mattermostClient, kube: tc.kube}
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
		"NilGroupID": {
			args: args{
				cr: integrationMattermost(),
			},
			want: want{
				cr:  integrationMattermost(),
				err: errors.New(errGroupIDMissing),
			},
		},
		"SuccessfulDelete": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockDeleteGroupMattermostIntegration: func(gid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
				),
			},
			want: want{
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
				),
				result: managed.ExternalDelete{},
			},
		},
		"DeleteFailed": {
			args: args{
				mattermostClient: &fake.MockClient{
					MockDeleteGroupMattermostIntegration: func(gid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
				),
			},
			want: want{
				cr: integrationMattermost(
					withGroupID(testGroupID),
					withChannel(testChannel),
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
