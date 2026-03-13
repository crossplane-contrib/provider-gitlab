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

package protectedenvironments

import (
	"context"
	"net/http"
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects"
)

var (
	envName = "production"

	projectID = "1234"

	errBoom = errors.New("boom")

	unexpected resource.Managed
)

type args struct {
	env projects.ProtectedEnvironmentClient

	kube client.Client

	cr resource.Managed
}

type peModifier func(*v1alpha1.ProtectedEnvironment)

func withConditions(c ...xpv1.Condition) peModifier {

	return func(r *v1alpha1.ProtectedEnvironment) { r.Status.Conditions = c }

}

func withStatus(s v1alpha1.ProtectedEnvironmentObservation) peModifier {

	return func(r *v1alpha1.ProtectedEnvironment) { r.Status.AtProvider = s }

}

func withProjectID(id *string) peModifier {

	return func(r *v1alpha1.ProtectedEnvironment) { r.Spec.ForProvider.ProjectID = id }

}

func withName(name *string) peModifier {

	return func(r *v1alpha1.ProtectedEnvironment) { r.Spec.ForProvider.Name = name }

}

func withRequiredApprovalCount(n *int) peModifier {

	return func(r *v1alpha1.ProtectedEnvironment) { r.Spec.ForProvider.RequiredApprovalCount = ptr.To(int64(*n)) }

}

func protectedEnvironment(m ...peModifier) *v1alpha1.ProtectedEnvironment {

	cr := &v1alpha1.ProtectedEnvironment{}

	for _, f := range m {

		f(cr)

	}

	return cr

}

func withApprovalRules(rules *[]v1alpha1.EnvironmentApprovalRuleParameters) peModifier {

	return func(r *v1alpha1.ProtectedEnvironment) {

		r.Spec.ForProvider.ApprovalRules = rules

	}

}

// mockProtectedEnvironmentClient is a minimal test double for projects.ProtectedEnvironmentClient.

type mockProtectedEnvironmentClient struct {
	getFn func(pid interface{}, name string, options ...gitlab.RequestOptionFunc) (*gitlab.ProtectedEnvironment, *gitlab.Response, error)

	createFn func(pid interface{}, opt *gitlab.ProtectRepositoryEnvironmentsOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProtectedEnvironment, *gitlab.Response, error)

	updateFn func(pid interface{}, name string, opt *gitlab.UpdateProtectedEnvironmentsOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProtectedEnvironment, *gitlab.Response, error)

	deleteFn func(pid interface{}, name string, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

	getCalls int

	createCalls int

	updateCalls int

	deleteCalls int
}

func (m *mockProtectedEnvironmentClient) GetProtectedEnvironment(pid interface{}, name string, options ...gitlab.RequestOptionFunc) (*gitlab.ProtectedEnvironment, *gitlab.Response, error) {

	m.getCalls++

	return m.getFn(pid, name, options...)

}

func (m *mockProtectedEnvironmentClient) ProtectRepositoryEnvironments(pid interface{}, opt *gitlab.ProtectRepositoryEnvironmentsOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProtectedEnvironment, *gitlab.Response, error) {

	m.createCalls++

	return m.createFn(pid, opt, options...)

}

func (m *mockProtectedEnvironmentClient) UpdateProtectedEnvironments(pid interface{}, name string, opt *gitlab.UpdateProtectedEnvironmentsOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProtectedEnvironment, *gitlab.Response, error) {

	m.updateCalls++

	return m.updateFn(pid, name, opt, options...)

}

func (m *mockProtectedEnvironmentClient) UnprotectEnvironment(pid interface{}, name string, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {

	m.deleteCalls++

	return m.deleteFn(pid, name, options...)

}

func TestConnect(t *testing.T) {

	type want struct {
		result managed.ExternalClient

		err error
	}

	cases := map[string]struct {
		args

		want
	}{

		"InvalidInput": {

			args: args{cr: unexpected},

			want: want{err: errors.New(errNotProtectedEnvironment)},
		},

		"ProviderConfigMissing": {

			args: args{

				cr: protectedEnvironment(),

				kube: &test.MockClient{MockGet: test.NewMockGetFn(nil)},
			},

			want: want{err: errors.New("providerConfigRef is not given")},
		},
	}

	for name, tc := range cases {

		t.Run(name, func(t *testing.T) {

			c := &connector{

				kube: tc.kube,

				newGitlabClientFn: func(cfg common.Config) projects.ProtectedEnvironmentClient {

					return tc.env

				},
			}

			o, err := c.Connect(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {

				t.Errorf("err: -want, +got:\n%s", diff)

			}

			if diff := cmp.Diff(tc.want.result, o); diff != "" {

				t.Errorf("client: -want, +got:\n%s", diff)

			}

		})

	}

}

func TestObserve(t *testing.T) {

	type want struct {
		cr resource.Managed

		result managed.ExternalObservation

		err error
	}

	cases := map[string]struct {
		args

		want
	}{

		"InvalidInput": {

			args: args{cr: unexpected},

			want: want{

				cr: unexpected,

				err: errors.New(errNotProtectedEnvironment),
			},
		},

		"NoName": {

			args: args{

				cr: protectedEnvironment(),
			},

			want: want{

				cr: protectedEnvironment(),

				result: managed.ExternalObservation{

					ResourceExists: false,
				},
			},
		},

		"ProjectIDMissing": {

			args: args{

				cr: protectedEnvironment(withName(ptr.To(envName))),
			},

			want: want{

				cr: protectedEnvironment(withName(ptr.To(envName))),

				err: errors.New(errProjectIDMissing),
			},
		},

		"Get404": {

			args: args{

				env: &mockProtectedEnvironmentClient{

					getFn: func(pid interface{}, name string, _ ...gitlab.RequestOptionFunc) (*gitlab.ProtectedEnvironment, *gitlab.Response, error) {

						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errBoom

					},
				},

				cr: protectedEnvironment(

					withName(ptr.To(envName)),

					withProjectID(ptr.To(projectID)),
				),
			},

			want: want{

				cr: protectedEnvironment(

					withName(ptr.To(envName)),

					withProjectID(ptr.To(projectID)),
				),

				result: managed.ExternalObservation{ResourceExists: false},
			},
		},

		"FailedGetNon404": {

			args: args{

				env: &mockProtectedEnvironmentClient{

					getFn: func(pid interface{}, name string, _ ...gitlab.RequestOptionFunc) (*gitlab.ProtectedEnvironment, *gitlab.Response, error) {

						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 400}}, errBoom

					},
				},

				cr: protectedEnvironment(

					withName(ptr.To(envName)),

					withProjectID(ptr.To(projectID)),
				),
			},

			want: want{

				cr: protectedEnvironment(

					withName(ptr.To(envName)),

					withProjectID(ptr.To(projectID)),
				),

				result: managed.ExternalObservation{ResourceExists: false},

				err: errors.Wrap(resource.Ignore(projects.IsErrorProtectedEnvironmentNotFound, errBoom), errGetFailed),
			},
		},

		"SuccessfulAvailable": {

			args: args{

				env: &mockProtectedEnvironmentClient{

					getFn: func(pid interface{}, name string, _ ...gitlab.RequestOptionFunc) (*gitlab.ProtectedEnvironment, *gitlab.Response, error) {

						return &gitlab.ProtectedEnvironment{

							Name: envName,

							RequiredApprovalCount: 0,
						}, &gitlab.Response{}, nil

					},
				},

				cr: protectedEnvironment(

					withName(ptr.To(envName)),

					withProjectID(ptr.To(projectID)),
				),
			},

			want: want{

				cr: protectedEnvironment(

					withName(ptr.To(envName)),

					withProjectID(ptr.To(projectID)),

					withConditions(xpv1.Available()),

					withStatus(v1alpha1.ProtectedEnvironmentObservation{

						Name: ptr.To(envName),

						RequiredApprovalCount: ptr.To(int64(0)),
					}),
				),

				result: managed.ExternalObservation{

					ResourceExists: true,

					ResourceUpToDate: true,

					ResourceLateInitialized: false,
				},
			},
		},

		"LateInitSuccess": {

			args: args{

				env: &mockProtectedEnvironmentClient{

					getFn: func(pid interface{}, name string, _ ...gitlab.RequestOptionFunc) (*gitlab.ProtectedEnvironment, *gitlab.Response, error) {

						return &gitlab.ProtectedEnvironment{

							Name: envName,

							RequiredApprovalCount: 2,
						}, &gitlab.Response{}, nil

					},
				},

				cr: protectedEnvironment(

					withName(ptr.To(envName)),

					withProjectID(ptr.To(projectID)),
				),
			},

			want: want{

				cr: protectedEnvironment(

					withName(ptr.To(envName)),

					withProjectID(ptr.To(projectID)),

					withRequiredApprovalCount(ptr.To(2)),

					withConditions(xpv1.Available()),

					withStatus(v1alpha1.ProtectedEnvironmentObservation{

						Name: ptr.To(envName),

						RequiredApprovalCount: ptr.To(int64(2)),
					}),
				),

				result: managed.ExternalObservation{

					ResourceExists: true,

					ResourceUpToDate: true,

					ResourceLateInitialized: true,
				},
			},
		},
	}

	for name, tc := range cases {

		t.Run(name, func(t *testing.T) {

			e := &external{client: tc.env}

			o, err := e.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {

				t.Errorf("err: -want, +got:\n%s", diff)

			}

			if diff := cmp.Diff(tc.want.result, o); diff != "" {

				t.Errorf("obs: -want, +got:\n%s", diff)

			}

			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {

				t.Errorf("cr: -want, +got:\n%s", diff)

			}

		})

	}

}

func TestCreate(t *testing.T) {

	type want struct {
		cr resource.Managed

		result managed.ExternalCreation

		err error
	}

	cases := map[string]struct {
		args

		want
	}{

		"InvalidInput": {

			args: args{cr: unexpected},

			want: want{

				cr: unexpected,

				err: errors.New(errNotProtectedEnvironment),
			},
		},

		"NoName": {

			args: args{cr: protectedEnvironment()},

			want: want{

				cr: protectedEnvironment(),

				err: errors.New(errNameMissing),
			},
		},

		"ProjectIDMissing": {

			args: args{

				cr: protectedEnvironment(withName(ptr.To(envName))),
			},

			want: want{

				cr: protectedEnvironment(withName(ptr.To(envName))),

				err: errors.New(errProjectIDMissing),
			},
		},

		"SuccessfulCreation": {

			args: args{

				env: &mockProtectedEnvironmentClient{

					createFn: func(pid interface{}, opt *gitlab.ProtectRepositoryEnvironmentsOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.ProtectedEnvironment, *gitlab.Response, error) {

						return &gitlab.ProtectedEnvironment{Name: envName}, &gitlab.Response{}, nil

					},
				},

				cr: protectedEnvironment(

					withName(ptr.To(envName)),

					withProjectID(ptr.To(projectID)),
				),
			},

			want: want{

				cr: protectedEnvironment(

					withName(ptr.To(envName)),

					withProjectID(ptr.To(projectID)),

					withConditions(xpv1.Creating()),
				),

				result: managed.ExternalCreation{},
			},
		},

		"FailedCreation": {

			args: args{

				env: &mockProtectedEnvironmentClient{

					createFn: func(pid interface{}, opt *gitlab.ProtectRepositoryEnvironmentsOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.ProtectedEnvironment, *gitlab.Response, error) {

						return nil, &gitlab.Response{}, errBoom

					},
				},

				cr: protectedEnvironment(

					withName(ptr.To(envName)),

					withProjectID(ptr.To(projectID)),
				),
			},

			want: want{

				cr: protectedEnvironment(

					withName(ptr.To(envName)),

					withProjectID(ptr.To(projectID)),

					withConditions(xpv1.Creating()),
				),

				err: errors.Wrap(errBoom, errCreateFailed),
			},
		},
	}

	for name, tc := range cases {

		t.Run(name, func(t *testing.T) {

			e := &external{client: tc.env}

			o, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {

				t.Errorf("err: -want, +got:\n%s", diff)

			}

			if diff := cmp.Diff(tc.want.result, o); diff != "" {

				t.Errorf("res: -want, +got:\n%s", diff)

			}

			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {

				t.Errorf("cr: -want, +got:\n%s", diff)

			}

		})

	}

}

func TestUpdate(t *testing.T) {

	type want struct {
		err error
	}

	emptyRules := []v1alpha1.EnvironmentApprovalRuleParameters{}

	cases := map[string]struct {
		args

		want
	}{

		"InvalidInput": {

			args: args{cr: unexpected},

			want: want{err: errors.New(errNotProtectedEnvironment)},
		},

		"NoName": {

			args: args{cr: protectedEnvironment()},

			want: want{err: errors.New(errNameMissing)},
		},

		"ProjectIDMissing": {

			args: args{

				cr: protectedEnvironment(withName(ptr.To(envName))),
			},

			want: want{err: errors.New(errProjectIDMissing)},
		},

		"FailedGet": {

			args: args{

				env: &mockProtectedEnvironmentClient{

					getFn: func(pid interface{}, name string, _ ...gitlab.RequestOptionFunc) (*gitlab.ProtectedEnvironment, *gitlab.Response, error) {

						return nil, &gitlab.Response{}, errBoom

					},
				},

				cr: protectedEnvironment(

					withName(ptr.To(envName)),

					withProjectID(ptr.To(projectID)),
				),
			},

			want: want{err: errors.Wrap(errBoom, errGetFailed)},
		},

		"NoOpUpdate": {

			args: args{

				env: &mockProtectedEnvironmentClient{

					getFn: func(pid interface{}, name string, _ ...gitlab.RequestOptionFunc) (*gitlab.ProtectedEnvironment, *gitlab.Response, error) {

						// Same external state as spec (no managed fields set) => GenerateUpdate... should return nil

						return &gitlab.ProtectedEnvironment{

							Name: envName,

							RequiredApprovalCount: 0,
						}, &gitlab.Response{}, nil

					},

					updateFn: func(pid interface{}, name string, opt *gitlab.UpdateProtectedEnvironmentsOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.ProtectedEnvironment, *gitlab.Response, error) {

						return nil, &gitlab.Response{}, errors.New("must not be called")

					},
				},

				cr: protectedEnvironment(

					withName(ptr.To(envName)),

					withProjectID(ptr.To(projectID)),
				),
			},

			want: want{},
		},

		"SuccessfulUpdate": {

			args: args{

				env: &mockProtectedEnvironmentClient{

					getFn: func(pid interface{}, name string, _ ...gitlab.RequestOptionFunc) (*gitlab.ProtectedEnvironment, *gitlab.Response, error) {

						// external required approvals differ => delta should be non-nil when rules are explicitly managed empty

						return &gitlab.ProtectedEnvironment{

							Name: envName,

							RequiredApprovalCount: 0,
						}, &gitlab.Response{}, nil

					},

					updateFn: func(pid interface{}, name string, opt *gitlab.UpdateProtectedEnvironmentsOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.ProtectedEnvironment, *gitlab.Response, error) {

						return &gitlab.ProtectedEnvironment{

							Name: envName,

							RequiredApprovalCount: 2,
						}, &gitlab.Response{}, nil

					},
				},

				cr: protectedEnvironment(

					withName(ptr.To(envName)),

					withProjectID(ptr.To(projectID)),

					withApprovalRules(&emptyRules),

					withRequiredApprovalCount(ptr.To(2)),
				),
			},

			want: want{},
		},

		"FailedUpdate": {

			args: args{

				env: &mockProtectedEnvironmentClient{

					getFn: func(pid interface{}, name string, _ ...gitlab.RequestOptionFunc) (*gitlab.ProtectedEnvironment, *gitlab.Response, error) {

						return &gitlab.ProtectedEnvironment{

							Name: envName,

							RequiredApprovalCount: 0,
						}, &gitlab.Response{}, nil

					},

					updateFn: func(pid interface{}, name string, opt *gitlab.UpdateProtectedEnvironmentsOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.ProtectedEnvironment, *gitlab.Response, error) {

						return nil, &gitlab.Response{}, errBoom

					},
				},

				cr: protectedEnvironment(

					withName(ptr.To(envName)),

					withProjectID(ptr.To(projectID)),

					withApprovalRules(&emptyRules),

					withRequiredApprovalCount(ptr.To(2)),
				),
			},

			want: want{err: errors.Wrap(errBoom, errUpdateFailed)},
		},
	}

	for name, tc := range cases {

		t.Run(name, func(t *testing.T) {

			e := &external{client: tc.env}

			_, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {

				t.Errorf("err: -want, +got:\n%s", diff)

			}

		})

	}

}

func TestDelete(t *testing.T) {

	type want struct {
		cr resource.Managed

		result managed.ExternalDelete

		err error
	}

	cases := map[string]struct {
		args

		want
	}{

		"InvalidInput": {

			args: args{cr: unexpected},

			want: want{

				cr: unexpected,

				err: errors.New(errNotProtectedEnvironment),
			},
		},

		"NoName": {

			args: args{cr: protectedEnvironment()},

			want: want{

				cr: protectedEnvironment(),

				err: errors.New(errNameMissing),
			},
		},

		"ProjectIDMissing": {

			args: args{

				cr: protectedEnvironment(withName(ptr.To(envName))),
			},

			want: want{

				cr: protectedEnvironment(withName(ptr.To(envName))),

				err: errors.New(errProjectIDMissing),
			},
		},

		"SuccessfulDeletion": {

			args: args{

				env: &mockProtectedEnvironmentClient{

					deleteFn: func(pid interface{}, name string, _ ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {

						return &gitlab.Response{}, nil

					},
				},

				cr: protectedEnvironment(

					withName(ptr.To(envName)),

					withProjectID(ptr.To(projectID)),
				),
			},

			want: want{

				cr: protectedEnvironment(

					withName(ptr.To(envName)),

					withProjectID(ptr.To(projectID)),

					withConditions(xpv1.Deleting()),
				),

				result: managed.ExternalDelete{},
			},
		},

		"FailedDeletion": {

			args: args{

				env: &mockProtectedEnvironmentClient{

					deleteFn: func(pid interface{}, name string, _ ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {

						return &gitlab.Response{}, errBoom

					},
				},

				cr: protectedEnvironment(

					withName(ptr.To(envName)),

					withProjectID(ptr.To(projectID)),
				),
			},

			want: want{

				cr: protectedEnvironment(

					withName(ptr.To(envName)),

					withProjectID(ptr.To(projectID)),

					withConditions(xpv1.Deleting()),
				),

				result: managed.ExternalDelete{},

				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
	}

	for name, tc := range cases {

		t.Run(name, func(t *testing.T) {

			e := &external{client: tc.env}

			o, err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {

				t.Errorf("err: -want, +got:\n%s", diff)

			}

			if diff := cmp.Diff(tc.want.result, o); diff != "" {

				t.Errorf("res: -want, +got:\n%s", diff)

			}

			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {

				t.Errorf("cr: -want, +got:\n%s", diff)

			}

		})

	}

}

func TestIsResponseNotFound_Helper(t *testing.T) {

	// ensure we keep the same not-found behavior as the rest of controllers

	r := &gitlab.Response{Response: &http.Response{StatusCode: 404}}

	if !clients.IsResponseNotFound(r) {

		t.Fatalf("expected 404 to be treated as not found")

	}

}
