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

package projectdeploytokens

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects/fake"
)

var (
	errBoom               = errors.New("boom")
	id                    = 0
	projectDeployTokenID  = 1234
	sProjectDeployTokenID = strconv.Itoa(projectDeployTokenID)
	unexpecedItem         resource.Managed
	expiresAt             = time.Now()
	token                 = "Token"
	username              = "Username"
	projectDeployTokenObj = gitlab.DeployToken{
		ID:        projectDeployTokenID,
		Name:      "Name",
		Username:  username,
		ExpiresAt: &expiresAt,
		Token:     token,
		Scopes:    []string{"scope1", "scope2"},
	}

	projectDeployTokens = []*gitlab.DeployToken{
		{
			ID:        projectDeployTokenID,
			Name:      "Name",
			Username:  username,
			ExpiresAt: &expiresAt,
			Token:     token,
			Scopes:    []string{"scope1", "scope2"},
		},
	}

	extNameAnnotation = map[string]string{meta.AnnotationKeyExternalName: fmt.Sprint(projectDeployTokenID)}
)

type args struct {
	projectDeployToken projects.ProjectDeployTokenClient
	kube               client.Client
	cr                 resource.Managed
}

type projectDeployTokenModifier func(*v1alpha1.ProjectDeployToken)

func withConditions(c ...xpv1.Condition) projectDeployTokenModifier {
	return func(r *v1alpha1.ProjectDeployToken) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(fp v1alpha1.ProjectDeployTokenParameters) projectDeployTokenModifier {
	return func(r *v1alpha1.ProjectDeployToken) { r.Spec.ForProvider = fp }
}

func withExternalName(projectDeployTokenID string) projectDeployTokenModifier {
	return func(r *v1alpha1.ProjectDeployToken) { meta.SetExternalName(r, projectDeployTokenID) }
}

func withAnnotations(a map[string]string) projectDeployTokenModifier {
	return func(p *v1alpha1.ProjectDeployToken) { meta.AddAnnotations(p, a) }
}

func projectDeployToken(m ...projectDeployTokenModifier) *v1alpha1.ProjectDeployToken {
	cr := &v1alpha1.ProjectDeployToken{}
	for _, f := range m {
		f(cr)
	}
	return cr
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
				err: errors.New(errNotProjectDeployToken),
			},
		},
		"NoExternalName": {
			args: args{
				cr: projectDeployToken(),
			},
			want: want{
				cr: projectDeployToken(),
				result: managed.ExternalObservation{
					ResourceExists:          false,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
			},
		},
		"NotIDExternalName": {
			args: args{
				projectDeployToken: &fake.MockClient{
					MockListProjectDeployTokens: func(pid interface{}, opt *gitlab.ListProjectDeployTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.DeployToken, *gitlab.Response, error) {
						return []*gitlab.DeployToken{}, &gitlab.Response{}, nil
					},
				},
				cr: projectDeployToken(withExternalName("fr")),
			},
			want: want{
				cr:  projectDeployToken(withExternalName("fr")),
				err: errors.New(errNotProjectDeployToken),
			},
		},
		"FailedGetRequest": {
			args: args{
				projectDeployToken: &fake.MockClient{
					MockListProjectDeployTokens: func(pid interface{}, opt *gitlab.ListProjectDeployTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.DeployToken, *gitlab.Response, error) {
						return []*gitlab.DeployToken{}, &gitlab.Response{}, errBoom
					},
				},
				cr: projectDeployToken(
					withExternalName(sProjectDeployTokenID),
					withSpec(v1alpha1.ProjectDeployTokenParameters{
						ProjectID: &projectDeployTokenID,
					}),
				),
			},
			want: want{
				cr: projectDeployToken(
					withAnnotations(extNameAnnotation),
					withSpec(v1alpha1.ProjectDeployTokenParameters{
						ProjectID: &projectDeployTokenID,
					}),
				),
				err: errors.Wrap(errBoom, errGetFailed),
			},
		},
		"ProjectDeployTokenNotFound": {
			args: args{
				projectDeployToken: &fake.MockClient{
					MockListProjectDeployTokens: func(pid interface{}, opt *gitlab.ListProjectDeployTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.DeployToken, *gitlab.Response, error) {
						return projectDeployTokens, &gitlab.Response{}, nil
					},
				},
				cr: projectDeployToken(
					withSpec(v1alpha1.ProjectDeployTokenParameters{
						ProjectID: &projectDeployTokenID,
						Username:  &username,
						ExpiresAt: &metav1.Time{Time: expiresAt},
					}),
					withExternalName("3"),
				),
			},
			want: want{
				cr: projectDeployToken(
					withSpec(v1alpha1.ProjectDeployTokenParameters{
						ProjectID: &projectDeployTokenID,
						Username:  &username,
						ExpiresAt: &metav1.Time{Time: expiresAt},
					}),
					withExternalName("3"),
				),
				result: managed.ExternalObservation{
					ResourceExists:          false,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
			},
		},
		"LateInitSuccess": {
			args: args{
				projectDeployToken: &fake.MockClient{
					MockListProjectDeployTokens: func(pid interface{}, opt *gitlab.ListProjectDeployTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.DeployToken, *gitlab.Response, error) {
						return projectDeployTokens, &gitlab.Response{}, nil
					},
				},
				cr: projectDeployToken(
					withSpec(v1alpha1.ProjectDeployTokenParameters{
						ProjectID: &projectDeployTokenID,
					}),
					withExternalName(sProjectDeployTokenID),
				),
			},
			want: want{
				cr: projectDeployToken(
					withSpec(v1alpha1.ProjectDeployTokenParameters{
						ProjectID: &projectDeployTokenID,
						Username:  &username,
						ExpiresAt: &metav1.Time{Time: expiresAt},
					}),
					withConditions(xpv1.Available()),
					withExternalName(sProjectDeployTokenID),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
				},
			},
		},
		"SuccessfulAvailable": {
			args: args{
				projectDeployToken: &fake.MockClient{
					MockListProjectDeployTokens: func(pid interface{}, opt *gitlab.ListProjectDeployTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.DeployToken, *gitlab.Response, error) {
						return projectDeployTokens, &gitlab.Response{}, nil
					},
				},
				cr: projectDeployToken(
					withSpec(v1alpha1.ProjectDeployTokenParameters{
						ProjectID: &projectDeployTokenID,
						Username:  &username,
						ExpiresAt: &metav1.Time{Time: expiresAt},
					}),
					withExternalName(sProjectDeployTokenID),
				),
			},
			want: want{
				cr: projectDeployToken(
					withSpec(v1alpha1.ProjectDeployTokenParameters{
						ProjectID: &projectDeployTokenID,
						Username:  &username,
						ExpiresAt: &metav1.Time{Time: expiresAt},
					}),
					withConditions(xpv1.Available()),
					withExternalName(sProjectDeployTokenID),
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
			e := &external{kube: tc.kube, client: tc.projectDeployToken}
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
				err: errors.New(errNotProjectDeployToken),
			},
		},
		"SuccessfulCreation": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				projectDeployToken: &fake.MockClient{
					MockCreateProjectDeployToken: func(pid interface{}, opt *gitlab.CreateProjectDeployTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error) {
						return &projectDeployTokenObj, &gitlab.Response{}, nil
					},
				},
				cr: projectDeployToken(
					withAnnotations(extNameAnnotation),
					withSpec(v1alpha1.ProjectDeployTokenParameters{
						ProjectID: &projectDeployTokenID,
					}),
				),
			},
			want: want{
				cr: projectDeployToken(
					withExternalName(sProjectDeployTokenID),
					withSpec(v1alpha1.ProjectDeployTokenParameters{
						ProjectID: &projectDeployTokenID,
					}),
				),
				result: managed.ExternalCreation{ExternalNameAssigned: true},
			},
		},
		"FailedCreation": {
			args: args{
				projectDeployToken: &fake.MockClient{
					MockCreateProjectDeployToken: func(pid interface{}, opt *gitlab.CreateProjectDeployTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error) {
						return &gitlab.DeployToken{}, &gitlab.Response{}, errBoom
					},
				},
				cr: projectDeployToken(
					withSpec(v1alpha1.ProjectDeployTokenParameters{
						ProjectID: &projectDeployTokenID,
					}),
					withExternalName("0"),
				),
			},
			want: want{
				cr: projectDeployToken(
					withSpec(v1alpha1.ProjectDeployTokenParameters{
						ProjectID: &projectDeployTokenID,
					}),
					withExternalName("0"),
				),
				err: errors.Wrap(errBoom, errCreateFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.projectDeployToken}
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
		"SuccessfulUpdate": {
			args: args{
				cr: projectDeployToken(),
			},
			want: want{
				cr: projectDeployToken(),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.projectDeployToken}
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
				err: errors.New(errNotProjectDeployToken),
			},
		},
		"SuccessfulDeletion": {
			args: args{
				projectDeployToken: &fake.MockClient{
					MockDeleteProjectDeployToken: func(pid interface{}, deployToken int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: projectDeployToken(
					withSpec(v1alpha1.ProjectDeployTokenParameters{
						ProjectID: &projectDeployTokenID,
					}),
					withExternalName(strconv.Itoa(id)),
				),
			},
			want: want{
				cr: projectDeployToken(
					withSpec(v1alpha1.ProjectDeployTokenParameters{
						ProjectID: &projectDeployTokenID,
					}),
					withExternalName(strconv.Itoa(id)),
				),
			},
		},
		"NotIDExternalName": {
			args: args{
				projectDeployToken: &fake.MockClient{
					MockDeleteProjectDeployToken: func(pid interface{}, deployToken int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: projectDeployToken(
					withSpec(v1alpha1.ProjectDeployTokenParameters{
						ProjectID: &projectDeployTokenID,
					}),
					withExternalName("test"),
				),
			},
			want: want{
				cr: projectDeployToken(
					withSpec(v1alpha1.ProjectDeployTokenParameters{
						ProjectID: &projectDeployTokenID,
					}),
					withExternalName("test"),
				),
				err: errors.New(errNotProjectDeployToken),
			},
		},
		"FailedDeletion": {
			args: args{
				projectDeployToken: &fake.MockClient{
					MockDeleteProjectDeployToken: func(pid interface{}, deployToken int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: projectDeployToken(
					withSpec(v1alpha1.ProjectDeployTokenParameters{
						ProjectID: &projectDeployTokenID,
					}),
					withExternalName(strconv.Itoa(id)),
				),
			},
			want: want{
				cr: projectDeployToken(
					withSpec(v1alpha1.ProjectDeployTokenParameters{
						ProjectID: &projectDeployTokenID,
					}),
					withExternalName(strconv.Itoa(id)),
				),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.projectDeployToken}
			err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
