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

package runners

import (
	"context"
	"net/http"
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	runners "github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/runners"
	runnersfake "github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/runners/fake"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/users"
	usersfake "github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/users/fake"
)

var (
	unexpecedItem resource.Managed

	errBoom           = errors.New("boom")
	projectID         = 1234
	runnerID          = 1
	extName           = "1"
	extNameAnnotation = map[string]string{meta.AnnotationKeyExternalName: extName}
)

type args struct {
	runnerClient     runners.RunnerClient
	userRunnerClient users.RunnerClient
	kube             client.Client
	cr               resource.Managed
}

type RunnerModifier func(*v1alpha1.Runner)

func withConditions(c ...xpv1.Condition) RunnerModifier {
	return func(cr *v1alpha1.Runner) { cr.Status.ConditionedStatus.Conditions = c }
}

func withProjectID() RunnerModifier {
	return func(r *v1alpha1.Runner) { r.Spec.ForProvider.ProjectID = &projectID }
}

func withAnnotations(a map[string]string) RunnerModifier {
	return func(p *v1alpha1.Runner) { meta.AddAnnotations(p, a) }
}

func withSpec(s v1alpha1.RunnerParameters) RunnerModifier {
	return func(r *v1alpha1.Runner) { r.Spec.ForProvider = s }
}

func withExternalName(n string) RunnerModifier {
	return func(r *v1alpha1.Runner) { meta.SetExternalName(r, n) }
}

func withConnectionSecretRef() RunnerModifier {
	return func(r *v1alpha1.Runner) {
		r.Spec.WriteConnectionSecretToReference = common.TestCreateLocalSecretReference("test-secret")
	}
}

func runner(m ...RunnerModifier) *v1alpha1.Runner {
	cr := &v1alpha1.Runner{}
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
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(errNotRunner),
			},
		},
		"ProviderConfigRefNotGivenError": {
			args: args{
				cr:   runner(),
				kube: &test.MockClient{MockGet: test.NewMockGetFn(nil)},
			},
			want: want{
				cr:  runner(),
				err: errors.New("providerConfigRef is not given"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &connector{kube: tc.kube, newGitlabClientFn: nil}
			o, err := c.Connect(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
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
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(errNotRunner),
			},
		},
		"NoExternalName": {
			args: args{
				cr: runner(
					withSpec(v1alpha1.RunnerParameters{ProjectID: &projectID}),
				),
			},
			want: want{
				cr: runner(
					withSpec(v1alpha1.RunnerParameters{ProjectID: &projectID}),
				),
				err:    nil,
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"NotIDExternalName": {
			args: args{
				runnerClient: &runnersfake.MockClient{
					MockGetRunnerDetails: func(rid any, options ...gitlab.RequestOptionFunc) (*gitlab.RunnerDetails, *gitlab.Response, error) {
						return &gitlab.RunnerDetails{}, &gitlab.Response{}, nil
					},
				},
				cr: runner(
					withExternalName("fr"),
					withProjectID(),
				),
			},
			want: want{
				cr: runner(
					withExternalName("fr"),
					withProjectID(),
				),
				err: errors.New(errIDNotInt),
			},
		},
		"NoProjectID": {
			args: args{
				cr: runner(
					withExternalName(extName),
				),
			},
			want: want{
				cr: runner(
					withExternalName(extName),
				),
				err:    errors.New(errMissingProjectID),
				result: managed.ExternalObservation{},
			},
		},
		"FailedGetRequest": {
			args: args{
				runnerClient: &runnersfake.MockClient{
					MockGetRunnerDetails: func(rid any, options ...gitlab.RequestOptionFunc) (*gitlab.RunnerDetails, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 400}}, errBoom
					},
				},
				cr: runner(
					withExternalName(extName),
					withProjectID(),
				),
			},
			want: want{
				cr: runner(
					withAnnotations(extNameAnnotation),
					withProjectID(),
				),
				result: managed.ExternalObservation{ResourceExists: false},
				err:    errors.Wrap(errBoom, errGetFailed),
			},
		},
		"ErrGet404": {
			args: args{
				runnerClient: &runnersfake.MockClient{
					MockGetRunnerDetails: func(rid any, options ...gitlab.RequestOptionFunc) (*gitlab.RunnerDetails, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errBoom
					},
				},
				cr: runner(
					withExternalName(extName),
					withProjectID(),
					withSpec(v1alpha1.RunnerParameters{ProjectID: &projectID}),
				),
			},
			want: want{
				cr: runner(
					withExternalName(extName),
					withProjectID(),
					withSpec(v1alpha1.RunnerParameters{ProjectID: &projectID}),
				),
				result: managed.ExternalObservation{ResourceExists: false},
				err:    nil,
			},
		},
		"SuccessfulAvailable": {
			args: args{
				runnerClient: &runnersfake.MockClient{
					MockGetRunnerDetails: func(rid any, options ...gitlab.RequestOptionFunc) (*gitlab.RunnerDetails, *gitlab.Response, error) {
						return nil, &gitlab.Response{}, nil
					},
				},
				cr: runner(
					withProjectID(),
					withExternalName(extName),
					withSpec(v1alpha1.RunnerParameters{ProjectID: &projectID}),
				),
			},
			want: want{
				cr: runner(
					withConditions(xpv1.Available()),
					withProjectID(),
					withExternalName(extName),
					withSpec(v1alpha1.RunnerParameters{ProjectID: &projectID}),
					// withStatus(v1alpha1.RunnerObservation{Name: name}),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: false,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.runnerClient, userRunnerClient: tc.userRunnerClient}
			o, err := e.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type want struct {
		cr  resource.Managed
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(errNotRunner),
			},
		},
		"NoExternalName": {
			args: args{
				cr: runner(
					withProjectID(),
				),
			},
			want: want{
				cr: runner(
					withProjectID(),
				),
				err: errors.New(errMissingExternalName),
			},
		},
		"NoProjectID": {
			args: args{
				cr: runner(
					withExternalName(extName),
				),
			},
			want: want{
				cr: runner(
					withExternalName(extName),
				),
				err: errors.New(errMissingProjectID),
			},
		},
		"FailedDeletion": {
			args: args{
				runnerClient: &runnersfake.MockClient{
					MockDeleteRegisteredRunnerByID: func(rid int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: runner(
					withProjectID(),
					withExternalName(extName),
					withSpec(v1alpha1.RunnerParameters{ProjectID: &projectID}),
				),
			},
			want: want{
				cr: runner(
					withProjectID(),
					withExternalName(extName),
					withSpec(v1alpha1.RunnerParameters{ProjectID: &projectID}),
				),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
		"SuccessfulDeletion": {
			args: args{
				runnerClient: &runnersfake.MockClient{
					MockDeleteRegisteredRunnerByID: func(rid int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: runner(
					withProjectID(),
					withExternalName(extName),
					withSpec(v1alpha1.RunnerParameters{ProjectID: &projectID}),
				),
			},
			want: want{
				cr: runner(
					withProjectID(),
					withExternalName(extName),
					withSpec(v1alpha1.RunnerParameters{ProjectID: &projectID}),
				),
				err: nil,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.runnerClient, userRunnerClient: tc.userRunnerClient}
			_, err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
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
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(errNotRunner),
			},
		},
		"NoProjectID": {
			args: args{
				cr: runner(
					withExternalName(extName),
				),
			},
			want: want{
				cr: runner(
					withExternalName(extName),
				),
				err:    errors.New(errMissingProjectID),
				result: managed.ExternalCreation{},
			},
		},
		"MissingConnectionSecret": {
			args: args{
				cr: runner(
					withProjectID(),
					withSpec(v1alpha1.RunnerParameters{ProjectID: &projectID}),
				),
			},
			want: want{
				cr: runner(
					withProjectID(),
					withSpec(v1alpha1.RunnerParameters{ProjectID: &projectID}),
				),
				err:    errors.New(errMissingConnectionSecret),
				result: managed.ExternalCreation{},
			},
		},
		"FailedCreation": {
			args: args{
				userRunnerClient: &usersfake.MockClient{
					MockCreateUserRunner: func(opts *gitlab.CreateUserRunnerOptions, options ...gitlab.RequestOptionFunc) (*gitlab.UserRunner, *gitlab.Response, error) {
						return &gitlab.UserRunner{}, &gitlab.Response{}, errBoom
					},
				},
				cr: runner(
					withProjectID(),
					withConnectionSecretRef(),
					withSpec(v1alpha1.RunnerParameters{ProjectID: &projectID}),
				),
			},
			want: want{
				cr: runner(
					withProjectID(),
					withConnectionSecretRef(),
					withSpec(v1alpha1.RunnerParameters{ProjectID: &projectID}),
				),
				err: errors.Wrap(errBoom, errCreateFailed),
			},
		},
		"SuccessfulCreation": {
			args: args{
				userRunnerClient: &usersfake.MockClient{
					MockCreateUserRunner: func(opts *gitlab.CreateUserRunnerOptions, options ...gitlab.RequestOptionFunc) (*gitlab.UserRunner, *gitlab.Response, error) {
						return &gitlab.UserRunner{ID: runnerID}, &gitlab.Response{}, nil
					},
				},
				cr: runner(
					withProjectID(),
					withConnectionSecretRef(),
					withSpec(v1alpha1.RunnerParameters{ProjectID: &projectID}),
				),
			},
			want: want{
				cr: runner(
					withProjectID(),
					withConnectionSecretRef(),
					withSpec(v1alpha1.RunnerParameters{ProjectID: &projectID}),
					withExternalName(extName),
					withConditions(xpv1.Creating()),
				),
				err: nil,
				result: managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{
						"token": []byte(""),
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.runnerClient, userRunnerClient: tc.userRunnerClient}
			o, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
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
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(errNotRunner),
			},
		},
		"NoProjectID": {
			args: args{
				cr: runner(
					withExternalName(extName),
				),
			},
			want: want{
				cr: runner(
					withExternalName(extName),
				),
				err:    errors.New(errMissingProjectID),
				result: managed.ExternalUpdate{},
			},
		},
		"NoExternalName": {
			args: args{
				cr: runner(
					withProjectID(),
				),
			},
			want: want{
				cr: runner(
					withProjectID(),
				),
				err:    errors.New(errMissingExternalName),
				result: managed.ExternalUpdate{},
			},
		},
		"NotIDExternalName": {
			args: args{
				cr: runner(
					withExternalName("fr"),
					withProjectID(),
				),
			},
			want: want{
				cr: runner(
					withExternalName("fr"),
					withProjectID(),
				),
				err: errors.New(errIDNotInt),
			},
		},
		"FailedUpdate": {
			args: args{
				runnerClient: &runnersfake.MockClient{
					MockUpdateRunnerDetails: func(rid any, opt *gitlab.UpdateRunnerDetailsOptions, options ...gitlab.RequestOptionFunc) (*gitlab.RunnerDetails, *gitlab.Response, error) {
						return nil, &gitlab.Response{}, errBoom
					},
				},
				cr: runner(
					withProjectID(),
					withExternalName(extName),
					withSpec(v1alpha1.RunnerParameters{ProjectID: &projectID}),
				),
			},
			want: want{
				cr: runner(
					withProjectID(),
					withExternalName(extName),
					withSpec(v1alpha1.RunnerParameters{ProjectID: &projectID}),
				),
				err: errors.Wrap(errBoom, errUpdateFailed),
			},
		},
		"SuccessfulUpdate": {
			args: args{
				runnerClient: &runnersfake.MockClient{
					MockUpdateRunnerDetails: func(rid any, opt *gitlab.UpdateRunnerDetailsOptions, options ...gitlab.RequestOptionFunc) (*gitlab.RunnerDetails, *gitlab.Response, error) {
						return &gitlab.RunnerDetails{}, &gitlab.Response{}, nil
					},
				},
				cr: runner(
					withProjectID(),
					withExternalName(extName),
					withSpec(v1alpha1.RunnerParameters{ProjectID: &projectID}),
				),
			},
			want: want{
				cr: runner(
					withProjectID(),
					withExternalName(extName),
					withSpec(v1alpha1.RunnerParameters{ProjectID: &projectID}),
				),
				err:    nil,
				result: managed.ExternalUpdate{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.runnerClient, userRunnerClient: tc.userRunnerClient}
			o, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
