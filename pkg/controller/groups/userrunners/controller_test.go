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

package userrunners

import (
	"context"
	"net/http"
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
	runners "github.com/crossplane-contrib/provider-gitlab/pkg/clients/runners"
	runnersfake "github.com/crossplane-contrib/provider-gitlab/pkg/clients/runners/fake"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/users"
	usersfake "github.com/crossplane-contrib/provider-gitlab/pkg/clients/users/fake"
)

var (
	unexpecedItem resource.Managed

	errBoom           = errors.New("boom")
	groupID           = 1234
	runnerID          = 1
	extName           = "1"
	extNameAnnotation = map[string]string{meta.AnnotationKeyExternalName: extName}
)

type args struct {
	runner     runners.RunnerClient
	userrunner users.UserRunnerClient
	kube       client.Client
	cr         resource.Managed
}

type UserRunnerModifier func(*v1alpha1.UserRunner)

func withConditions(c ...xpv1.Condition) UserRunnerModifier {
	return func(cr *v1alpha1.UserRunner) { cr.Status.ConditionedStatus.Conditions = c }
}

func withGroupID() UserRunnerModifier {
	return func(r *v1alpha1.UserRunner) { r.Spec.ForProvider.GroupID = &groupID }
}

func withStatus(s v1alpha1.UserRunnerObservation) UserRunnerModifier {
	return func(r *v1alpha1.UserRunner) { r.Status.AtProvider = s }
}

func withAnnotations(a map[string]string) UserRunnerModifier {
	return func(p *v1alpha1.UserRunner) { meta.AddAnnotations(p, a) }
}

func withSpec(s v1alpha1.UserRunnerParameters) UserRunnerModifier {
	return func(r *v1alpha1.UserRunner) { r.Spec.ForProvider = s }
}

func withExternalName(n string) UserRunnerModifier {
	return func(r *v1alpha1.UserRunner) { meta.SetExternalName(r, n) }
}

func withConnectionSecretRef() UserRunnerModifier {
	return func(r *v1alpha1.UserRunner) {
		r.Spec.WriteConnectionSecretToReference = &xpv1.SecretReference{
			Name:      "test-secret",
			Namespace: "default",
		}
	}
}

func userRunner(m ...UserRunnerModifier) *v1alpha1.UserRunner {
	cr := &v1alpha1.UserRunner{}
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
				err: errors.New(errNotUserRunner),
			},
		},
		"ProviderConfigRefNotGivenError": {
			args: args{
				cr:   userRunner(),
				kube: &test.MockClient{MockGet: test.NewMockGetFn(nil)},
			},
			want: want{
				cr:  userRunner(),
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
				err: errors.New(errNotUserRunner),
			},
		},
		"NoExternalName": {
			args: args{
				cr: userRunner(
					withSpec(v1alpha1.UserRunnerParameters{GroupID: &groupID}),
				),
			},
			want: want{
				cr: userRunner(
					withSpec(v1alpha1.UserRunnerParameters{GroupID: &groupID}),
				),
				err:    nil,
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"NotIDExternalName": {
			args: args{
				runner: &runnersfake.MockClient{
					MockGetRunnerDetails: func(rid any, options ...gitlab.RequestOptionFunc) (*gitlab.RunnerDetails, *gitlab.Response, error) {
						return &gitlab.RunnerDetails{}, &gitlab.Response{}, nil
					},
				},
				cr: userRunner(
					withExternalName("fr"),
					withGroupID(),
				),
			},
			want: want{
				cr: userRunner(
					withExternalName("fr"),
					withGroupID(),
				),
				err: errors.New(errIDNotInt),
			},
		},
		"NoGroupID": {
			args: args{
				cr: userRunner(
					withExternalName(extName),
				),
			},
			want: want{
				cr: userRunner(
					withExternalName(extName),
				),
				err:    errors.New(errMissingGroupID),
				result: managed.ExternalObservation{},
			},
		},
		"FailedGetRequest": {
			args: args{
				runner: &runnersfake.MockClient{
					MockGetRunnerDetails: func(rid any, options ...gitlab.RequestOptionFunc) (*gitlab.RunnerDetails, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 400}}, errBoom
					},
				},
				cr: userRunner(
					withExternalName(extName),
					withGroupID(),
				),
			},
			want: want{
				cr: userRunner(
					withAnnotations(extNameAnnotation),
					withGroupID(),
				),
				result: managed.ExternalObservation{ResourceExists: false},
				err:    errors.Wrap(errBoom, errGetFailed),
			},
		},
		"ErrGet404": {
			args: args{
				runner: &runnersfake.MockClient{
					MockGetRunnerDetails: func(rid any, options ...gitlab.RequestOptionFunc) (*gitlab.RunnerDetails, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errBoom
					},
				},
				cr: userRunner(
					withExternalName(extName),
					withGroupID(),
					withSpec(v1alpha1.UserRunnerParameters{GroupID: &groupID}),
				),
			},
			want: want{
				cr: userRunner(
					withExternalName(extName),
					withGroupID(),
					withSpec(v1alpha1.UserRunnerParameters{GroupID: &groupID}),
				),
				result: managed.ExternalObservation{ResourceExists: false},
				err:    nil,
			},
		},
		"SuccessfulAvailable": {
			args: args{
				runner: &runnersfake.MockClient{
					MockGetRunnerDetails: func(rid any, options ...gitlab.RequestOptionFunc) (*gitlab.RunnerDetails, *gitlab.Response, error) {
						return nil, &gitlab.Response{}, nil
					},
				},
				cr: userRunner(
					withGroupID(),
					withExternalName(extName),
					withSpec(v1alpha1.UserRunnerParameters{GroupID: &groupID}),
				),
			},
			want: want{
				cr: userRunner(
					withConditions(xpv1.Available()),
					withGroupID(),
					withExternalName(extName),
					withSpec(v1alpha1.UserRunnerParameters{GroupID: &groupID}),
					// withStatus(v1alpha1.UserRunnerObservation{Name: name}),
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
			e := &external{kube: tc.kube, client: tc.runner, userRunnerClient: tc.userrunner}
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
				err: errors.New(errNotUserRunner),
			},
		},
		"NoExternalName": {
			args: args{
				cr: userRunner(
					withGroupID(),
				),
			},
			want: want{
				cr: userRunner(
					withGroupID(),
				),
				err: errors.New(errMissingExternalName),
			},
		},
		"NoGroupID": {
			args: args{
				cr: userRunner(
					withExternalName(extName),
				),
			},
			want: want{
				cr: userRunner(
					withExternalName(extName),
				),
				err: errors.New(errMissingGroupID),
			},
		},
		"FailedDeletion": {
			args: args{
				runner: &runnersfake.MockClient{
					MockDeleteRegisteredRunnerByID: func(rid int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: userRunner(
					withGroupID(),
					withExternalName(extName),
					withSpec(v1alpha1.UserRunnerParameters{GroupID: &groupID}),
				),
			},
			want: want{
				cr: userRunner(
					withGroupID(),
					withExternalName(extName),
					withSpec(v1alpha1.UserRunnerParameters{GroupID: &groupID}),
				),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
		"SuccessfulDeletion": {
			args: args{
				runner: &runnersfake.MockClient{
					MockDeleteRegisteredRunnerByID: func(rid int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: userRunner(
					withGroupID(),
					withExternalName(extName),
					withSpec(v1alpha1.UserRunnerParameters{GroupID: &groupID}),
				),
			},
			want: want{
				cr: userRunner(
					withGroupID(),
					withExternalName(extName),
					withSpec(v1alpha1.UserRunnerParameters{GroupID: &groupID}),
				),
				err: nil,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.runner, userRunnerClient: tc.userrunner}
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
				err: errors.New(errNotUserRunner),
			},
		},
		"NoGroupID": {
			args: args{
				cr: userRunner(
					withExternalName(extName),
				),
			},
			want: want{
				cr: userRunner(
					withExternalName(extName),
				),
				err:    errors.New(errMissingGroupID),
				result: managed.ExternalCreation{},
			},
		},
		"MissingConnectionSecret": {
			args: args{
				cr: userRunner(
					withGroupID(),
					withSpec(v1alpha1.UserRunnerParameters{GroupID: &groupID}),
				),
			},
			want: want{
				cr: userRunner(
					withGroupID(),
					withSpec(v1alpha1.UserRunnerParameters{GroupID: &groupID}),
				),
				err:    errors.New(errMissingConnectionSecret),
				result: managed.ExternalCreation{},
			},
		},
		"FailedCreation": {
			args: args{
				userrunner: &usersfake.MockClient{
					MockCreateUserRunner: func(opts *gitlab.CreateUserRunnerOptions, options ...gitlab.RequestOptionFunc) (*gitlab.UserRunner, *gitlab.Response, error) {
						return &gitlab.UserRunner{}, &gitlab.Response{}, errBoom
					},
				},
				cr: userRunner(
					withGroupID(),
					withSpec(v1alpha1.UserRunnerParameters{GroupID: &groupID}),
					withConnectionSecretRef(),
				),
			},
			want: want{
				cr: userRunner(
					withGroupID(),
					withSpec(v1alpha1.UserRunnerParameters{GroupID: &groupID}),
					withConnectionSecretRef(),
				),
				err: errors.Wrap(errBoom, errCreateFailed),
			},
		},
		"SuccessfulCreation": {
			args: args{
				userrunner: &usersfake.MockClient{
					MockCreateUserRunner: func(opts *gitlab.CreateUserRunnerOptions, options ...gitlab.RequestOptionFunc) (*gitlab.UserRunner, *gitlab.Response, error) {
						return &gitlab.UserRunner{ID: runnerID}, &gitlab.Response{}, nil
					},
				},
				cr: userRunner(
					withGroupID(),
					withConnectionSecretRef(),
					withSpec(v1alpha1.UserRunnerParameters{GroupID: &groupID}),
				),
			},
			want: want{
				cr: userRunner(
					withGroupID(),
					withConnectionSecretRef(),
					withSpec(v1alpha1.UserRunnerParameters{GroupID: &groupID}),
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
			e := &external{kube: tc.kube, client: tc.runner, userRunnerClient: tc.userrunner}
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
				err: errors.New(errNotUserRunner),
			},
		},
		"NoGroupID": {
			args: args{
				cr: userRunner(
					withExternalName(extName),
				),
			},
			want: want{
				cr: userRunner(
					withExternalName(extName),
				),
				err:    errors.New(errMissingGroupID),
				result: managed.ExternalUpdate{},
			},
		},
		"NoExternalName": {
			args: args{
				cr: userRunner(
					withGroupID(),
				),
			},
			want: want{
				cr: userRunner(
					withGroupID(),
				),
				err:    errors.New(errMissingExternalName),
				result: managed.ExternalUpdate{},
			},
		},
		"NotIDExternalName": {
			args: args{
				cr: userRunner(
					withExternalName("fr"),
					withGroupID(),
				),
			},
			want: want{
				cr: userRunner(
					withExternalName("fr"),
					withGroupID(),
				),
				err: errors.New(errIDNotInt),
			},
		},
		"FailedUpdate": {
			args: args{
				runner: &runnersfake.MockClient{
					MockUpdateRunnerDetails: func(rid any, opt *gitlab.UpdateRunnerDetailsOptions, options ...gitlab.RequestOptionFunc) (*gitlab.RunnerDetails, *gitlab.Response, error) {
						return nil, &gitlab.Response{}, errBoom
					},
				},
				cr: userRunner(
					withGroupID(),
					withExternalName(extName),
					withSpec(v1alpha1.UserRunnerParameters{GroupID: &groupID}),
				),
			},
			want: want{
				cr: userRunner(
					withGroupID(),
					withExternalName(extName),
					withSpec(v1alpha1.UserRunnerParameters{GroupID: &groupID}),
				),
				err: errors.Wrap(errBoom, errUpdateFailed),
			},
		},
		"SuccessfulUpdate": {
			args: args{
				runner: &runnersfake.MockClient{
					MockUpdateRunnerDetails: func(rid any, opt *gitlab.UpdateRunnerDetailsOptions, options ...gitlab.RequestOptionFunc) (*gitlab.RunnerDetails, *gitlab.Response, error) {
						return &gitlab.RunnerDetails{}, &gitlab.Response{}, nil
					},
				},
				cr: userRunner(
					withGroupID(),
					withExternalName(extName),
					withSpec(v1alpha1.UserRunnerParameters{GroupID: &groupID}),
				),
			},
			want: want{
				cr: userRunner(
					withGroupID(),
					withExternalName(extName),
					withSpec(v1alpha1.UserRunnerParameters{GroupID: &groupID}),
				),
				err:    nil,
				result: managed.ExternalUpdate{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.runner, userRunnerClient: tc.userrunner}
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
