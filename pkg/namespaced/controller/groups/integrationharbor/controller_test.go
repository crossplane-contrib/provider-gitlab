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

package integrationharbor

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
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
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

	testGroupID   int64 = 42
	testIntegID   int64 = 999
	testHarborURL       = "https://demo.goharbor.io"
	testPassword        = "supersecret"
)

type args struct {
	harborClient groupclients.HarborClient
	kube         client.Client
	cr           resource.Managed
}

type harborModifier func(*v1alpha1.IntegrationHarbor)

func withGroupID(id int64) harborModifier {
	return func(r *v1alpha1.IntegrationHarbor) {
		r.Spec.ForProvider.GroupID = &id
	}
}

func withURL(u string) harborModifier {
	return func(r *v1alpha1.IntegrationHarbor) {
		r.Spec.ForProvider.URL = u
	}
}

func withPasswordSecretRef(name, key string) harborModifier {
	return func(r *v1alpha1.IntegrationHarbor) {
		r.Spec.ForProvider.PasswordSecretRef = v2.LocalSecretKeySelector{
			LocalSecretReference: v2.LocalSecretReference{Name: name},
			Key:                  key,
		}
	}
}

func withConditions(c ...v2.Condition) harborModifier {
	return func(cr *v1alpha1.IntegrationHarbor) {
		cr.Status.ConditionedStatus.Conditions = c
	}
}

func withStatus(s v1alpha1.IntegrationHarborObservation) harborModifier {
	return func(r *v1alpha1.IntegrationHarbor) {
		r.Status.AtProvider = s
	}
}

func withDeletionTimestamp(t time.Time) harborModifier {
	return func(r *v1alpha1.IntegrationHarbor) {
		r.ObjectMeta.DeletionTimestamp = &metav1.Time{Time: t}
	}
}

func integrationHarbor(m ...harborModifier) *v1alpha1.IntegrationHarbor {
	cr := &v1alpha1.IntegrationHarbor{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func mockKubeWithSecret(password string) client.Client {
	return &test.MockClient{
		MockGet: func(_ context.Context, _ client.ObjectKey, obj client.Object) error {
			secret, ok := obj.(*corev1.Secret)
			if !ok {
				return errors.Errorf("unexpected object type %T", obj)
			}
			secret.Data = map[string][]byte{"password": []byte(password)}
			return nil
		},
	}
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
				err: errors.New(errNotIntegrationHarbor),
			},
		},
		"ProviderConfigRefNotGivenError": {
			args: args{
				cr:   integrationHarbor(),
				kube: &test.MockClient{MockGet: test.NewMockGetFn(nil)},
			},
			want: want{
				cr:  integrationHarbor(),
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
				err: errors.New(errNotIntegrationHarbor),
			},
		},
		"DeletingResourceStillExists": {
			args: args{
				harborClient: &fake.MockClient{
					MockGetGroupHarborSettings: func(gid any, options ...gitlab.RequestOptionFunc) (*gitlab.HarborIntegration, *gitlab.Response, error) {
						return &gitlab.HarborIntegration{
							Integration: gitlab.Integration{
								ID:     testIntegID,
								Title:  "Harbor",
								Slug:   "harbor",
								Active: true,
							},
							Properties: gitlab.HarborIntegrationProperties{URL: testHarborURL},
						}, &gitlab.Response{}, nil
					},
				},
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
					withDeletionTimestamp(time.Now()),
				),
			},
			want: want{
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
					withDeletionTimestamp(time.Now()),
					withConditions(v2.Available()),
					withStatus(v1alpha1.IntegrationHarborObservation{
						CommonIntegrationObservation: commonv1alpha1.CommonIntegrationObservation{
							ID:                             ptr.To(testIntegID),
							Title:                          ptr.To("Harbor"),
							Slug:                           ptr.To("harbor"),
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
				harborClient: &fake.MockClient{
					MockGetGroupHarborSettings: func(gid any, options ...gitlab.RequestOptionFunc) (*gitlab.HarborIntegration, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errBoom
					},
				},
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withDeletionTimestamp(time.Now()),
				),
			},
			want: want{
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withDeletionTimestamp(time.Now()),
				),
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"NilGroupID": {
			args: args{
				cr: integrationHarbor(),
			},
			want: want{
				cr:  integrationHarbor(),
				err: errors.New(errGroupIDMissing),
			},
		},
		"NotFound": {
			args: args{
				harborClient: &fake.MockClient{
					MockGetGroupHarborSettings: func(gid any, options ...gitlab.RequestOptionFunc) (*gitlab.HarborIntegration, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errBoom
					},
				},
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
				),
			},
			want: want{
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
				),
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"GetFailed": {
			args: args{
				harborClient: &fake.MockClient{
					MockGetGroupHarborSettings: func(gid any, options ...gitlab.RequestOptionFunc) (*gitlab.HarborIntegration, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 500}}, errBoom
					},
				},
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
				),
			},
			want: want{
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
				),
				result: managed.ExternalObservation{},
				err:    errors.Wrap(errBoom, errGetFailed),
			},
		},
		"NilIntegration": {
			args: args{
				harborClient: &fake.MockClient{
					MockGetGroupHarborSettings: func(gid any, options ...gitlab.RequestOptionFunc) (*gitlab.HarborIntegration, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
					},
				},
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
				),
			},
			want: want{
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
				),
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"SuccessActive": {
			args: args{
				harborClient: &fake.MockClient{
					MockGetGroupHarborSettings: func(gid any, options ...gitlab.RequestOptionFunc) (*gitlab.HarborIntegration, *gitlab.Response, error) {
						return &gitlab.HarborIntegration{
							Integration: gitlab.Integration{
								ID:     testIntegID,
								Title:  "Harbor",
								Slug:   "harbor",
								Active: true,
							},
							Properties: gitlab.HarborIntegrationProperties{URL: testHarborURL},
						}, &gitlab.Response{}, nil
					},
				},
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
				),
			},
			want: want{
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
					withConditions(v2.Available()),
					withStatus(v1alpha1.IntegrationHarborObservation{
						CommonIntegrationObservation: commonv1alpha1.CommonIntegrationObservation{
							ID:                             ptr.To(testIntegID),
							Title:                          ptr.To("Harbor"),
							Slug:                           ptr.To("harbor"),
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
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"SuccessInactive": {
			args: args{
				harborClient: &fake.MockClient{
					MockGetGroupHarborSettings: func(gid any, options ...gitlab.RequestOptionFunc) (*gitlab.HarborIntegration, *gitlab.Response, error) {
						return &gitlab.HarborIntegration{
							Integration: gitlab.Integration{
								ID:     testIntegID,
								Title:  "Harbor",
								Slug:   "harbor",
								Active: false,
							},
						}, &gitlab.Response{}, nil
					},
				},
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
				),
			},
			want: want{
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
				),
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.harborClient}
			o, err := e.Observe(context.Background(), tc.args.cr)

			if (tc.want.err == nil) != (err == nil) {
				t.Errorf("Observe(...): -want error, +got error: \nwant: %v\ngot: %v", tc.want.err, err)
			}
			if tc.want.err != nil && err != nil && tc.want.err.Error() != err.Error() {
				t.Errorf("Observe(...): -want error, +got error: \nwant: %v\ngot: %v", tc.want.err, err)
			}

			opts := []cmp.Option{test.EquateConditions()}
			if cr, ok := tc.args.cr.(*v1alpha1.IntegrationHarbor); ok && cr != nil && !cr.ObjectMeta.DeletionTimestamp.IsZero() {
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
				err: errors.New(errNotIntegrationHarbor),
			},
		},
		"NilGroupID": {
			args: args{
				cr: integrationHarbor(),
			},
			want: want{
				cr:  integrationHarbor(),
				err: errors.New(errGroupIDMissing),
			},
		},
		"PasswordResolutionFailed": {
			args: args{
				harborClient: &fake.MockClient{},
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(errBoom),
				},
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
					withPasswordSecretRef("my-secret", "password"),
				),
			},
			want: want{
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
					withPasswordSecretRef("my-secret", "password"),
					withConditions(v2.Creating()),
				),
				err: errors.Wrap(errors.Wrap(errors.Wrap(errBoom, "Cannot find referenced secret"), errPasswordMissing), errCreateFailed),
			},
		},
		"SuccessfulCreate": {
			args: args{
				harborClient: &fake.MockClient{
					MockSetUpGroupHarbor: func(gid any, opt *gitlab.SetUpHarborOptions, options ...gitlab.RequestOptionFunc) (*gitlab.HarborIntegration, *gitlab.Response, error) {
						return &gitlab.HarborIntegration{Integration: gitlab.Integration{Active: true}}, &gitlab.Response{}, nil
					},
				},
				kube: mockKubeWithSecret(testPassword),
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
					withPasswordSecretRef("my-secret", "password"),
				),
			},
			want: want{
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
					withPasswordSecretRef("my-secret", "password"),
					withConditions(v2.Creating()),
				),
				result: managed.ExternalCreation{},
			},
		},
		"CreateFailed": {
			args: args{
				harborClient: &fake.MockClient{
					MockSetUpGroupHarbor: func(gid any, opt *gitlab.SetUpHarborOptions, options ...gitlab.RequestOptionFunc) (*gitlab.HarborIntegration, *gitlab.Response, error) {
						return nil, &gitlab.Response{}, errBoom
					},
				},
				kube: mockKubeWithSecret(testPassword),
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
					withPasswordSecretRef("my-secret", "password"),
				),
			},
			want: want{
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
					withPasswordSecretRef("my-secret", "password"),
					withConditions(v2.Creating()),
				),
				result: managed.ExternalCreation{},
				err:    errors.Wrap(errBoom, errCreateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.harborClient, kube: tc.kube}
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
				err: errors.New(errNotIntegrationHarbor),
			},
		},
		"NilGroupID": {
			args: args{
				cr: integrationHarbor(),
			},
			want: want{
				cr:  integrationHarbor(),
				err: errors.New(errGroupIDMissing),
			},
		},
		"PasswordResolutionFailed": {
			args: args{
				harborClient: &fake.MockClient{},
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(errBoom),
				},
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
					withPasswordSecretRef("my-secret", "password"),
				),
			},
			want: want{
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
					withPasswordSecretRef("my-secret", "password"),
				),
				err: errors.Wrap(errors.Wrap(errors.Wrap(errBoom, "Cannot find referenced secret"), errPasswordMissing), errUpdateFailed),
			},
		},
		"SuccessfulUpdate": {
			args: args{
				harborClient: &fake.MockClient{
					MockSetUpGroupHarbor: func(gid any, opt *gitlab.SetUpHarborOptions, options ...gitlab.RequestOptionFunc) (*gitlab.HarborIntegration, *gitlab.Response, error) {
						return &gitlab.HarborIntegration{Integration: gitlab.Integration{Active: true}}, &gitlab.Response{}, nil
					},
				},
				kube: mockKubeWithSecret(testPassword),
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
					withPasswordSecretRef("my-secret", "password"),
				),
			},
			want: want{
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
					withPasswordSecretRef("my-secret", "password"),
				),
				result: managed.ExternalUpdate{},
			},
		},
		"UpdateFailed": {
			args: args{
				harborClient: &fake.MockClient{
					MockSetUpGroupHarbor: func(gid any, opt *gitlab.SetUpHarborOptions, options ...gitlab.RequestOptionFunc) (*gitlab.HarborIntegration, *gitlab.Response, error) {
						return nil, &gitlab.Response{}, errBoom
					},
				},
				kube: mockKubeWithSecret(testPassword),
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
					withPasswordSecretRef("my-secret", "password"),
				),
			},
			want: want{
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
					withPasswordSecretRef("my-secret", "password"),
				),
				result: managed.ExternalUpdate{},
				err:    errors.Wrap(errBoom, errUpdateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.harborClient, kube: tc.kube}
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
				err: errors.New(errNotIntegrationHarbor),
			},
		},
		"NilGroupID": {
			args: args{
				cr: integrationHarbor(),
			},
			want: want{
				cr:  integrationHarbor(),
				err: errors.New(errGroupIDMissing),
			},
		},
		"SuccessfulDelete": {
			args: args{
				harborClient: &fake.MockClient{
					MockDisableGroupHarbor: func(gid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
				),
			},
			want: want{
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
				),
				result: managed.ExternalDelete{},
			},
		},
		"DeleteFailed": {
			args: args{
				harborClient: &fake.MockClient{
					MockDisableGroupHarbor: func(gid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
				),
			},
			want: want{
				cr: integrationHarbor(
					withGroupID(testGroupID),
					withURL(testHarborURL),
				),
				result: managed.ExternalDelete{},
				err:    errors.Wrap(errBoom, errDeleteFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.harborClient}
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
