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

package deploytokens

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
	errBoom        = errors.New("boom")
	id             = 0
	deployTokenID  = 1234
	sDeployTokenID = strconv.Itoa(deployTokenID)
	unexpecedItem  resource.Managed
	expiresAt      = time.Now()
	token          = "Token"
	username       = "Username"
	deployTokenObj = gitlab.DeployToken{
		ID:        deployTokenID,
		Name:      "Name",
		Username:  username,
		ExpiresAt: &expiresAt,
		Token:     token,
		Scopes:    []string{"scope1", "scope2"},
	}

	extNameAnnotation = map[string]string{meta.AnnotationKeyExternalName: fmt.Sprint(deployTokenID)}
)

type args struct {
	deployToken projects.DeployTokenClient
	kube        client.Client
	cr          resource.Managed
}

type deployTokenModifier func(*v1alpha1.DeployToken)

func withConditions(c ...xpv1.Condition) deployTokenModifier {
	return func(r *v1alpha1.DeployToken) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(fp v1alpha1.DeployTokenParameters) deployTokenModifier {
	return func(r *v1alpha1.DeployToken) { r.Spec.ForProvider = fp }
}

func withExternalName(deployTokenID string) deployTokenModifier {
	return func(r *v1alpha1.DeployToken) { meta.SetExternalName(r, deployTokenID) }
}

func withAnnotations(a map[string]string) deployTokenModifier {
	return func(p *v1alpha1.DeployToken) { meta.AddAnnotations(p, a) }
}

func deployToken(m ...deployTokenModifier) *v1alpha1.DeployToken {
	cr := &v1alpha1.DeployToken{}
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
				err: errors.New(errNotDeployToken),
			},
		},
		"NoExternalName": {
			args: args{
				cr: deployToken(),
			},
			want: want{
				cr: deployToken(),
				result: managed.ExternalObservation{
					ResourceExists:          false,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
			},
		},
		"NotIDExternalName": {
			args: args{
				cr: deployToken(withExternalName("fr")),
			},
			want: want{
				cr:  deployToken(withExternalName("fr")),
				err: errors.New(errIDnotInt),
			},
		},
		"FailedGetRequest": {
			args: args{
				deployToken: &fake.MockClient{
					MockGetProjectDeployToken: func(pid interface{}, deployToken int, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error) {
						return nil, nil, errBoom
					},
				},
				cr: deployToken(
					withExternalName(sDeployTokenID),
					withSpec(v1alpha1.DeployTokenParameters{
						ProjectID: &deployTokenID,
					}),
				),
			},
			want: want{
				cr: deployToken(
					withAnnotations(extNameAnnotation),
					withSpec(v1alpha1.DeployTokenParameters{
						ProjectID: &deployTokenID,
					}),
				),
				err: errors.Wrap(errBoom, errGetFailed),
			},
		},
		"DeployTokenNotFound": {
			args: args{
				deployToken: &fake.MockClient{
					MockGetProjectDeployToken: func(pid interface{}, deployToken int, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error) {
						return nil, &gitlab.Response{}, nil
					},
				},
				cr: deployToken(
					withSpec(v1alpha1.DeployTokenParameters{
						ProjectID: &deployTokenID,
						Username:  &username,
						ExpiresAt: &metav1.Time{Time: expiresAt},
					}),
					withExternalName("3"),
				),
			},
			want: want{
				cr: deployToken(
					withSpec(v1alpha1.DeployTokenParameters{
						ProjectID: &deployTokenID,
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
				deployToken: &fake.MockClient{
					MockGetProjectDeployToken: func(pid interface{}, deployToken int, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error) {
						return &deployTokenObj, &gitlab.Response{}, nil
					},
				},
				cr: deployToken(
					withSpec(v1alpha1.DeployTokenParameters{
						ProjectID: &deployTokenID,
					}),
					withExternalName(sDeployTokenID),
				),
			},
			want: want{
				cr: deployToken(
					withSpec(v1alpha1.DeployTokenParameters{
						ProjectID: &deployTokenID,
						Username:  &username,
						ExpiresAt: &metav1.Time{Time: expiresAt},
					}),
					withConditions(xpv1.Available()),
					withExternalName(sDeployTokenID),
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
				deployToken: &fake.MockClient{
					MockGetProjectDeployToken: func(pid interface{}, deployToken int, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error) {
						return &deployTokenObj, &gitlab.Response{}, nil
					},
				},
				cr: deployToken(
					withSpec(v1alpha1.DeployTokenParameters{
						ProjectID: &deployTokenID,
						Username:  &username,
						ExpiresAt: &metav1.Time{Time: expiresAt},
					}),
					withExternalName(sDeployTokenID),
				),
			},
			want: want{
				cr: deployToken(
					withSpec(v1alpha1.DeployTokenParameters{
						ProjectID: &deployTokenID,
						Username:  &username,
						ExpiresAt: &metav1.Time{Time: expiresAt},
					}),
					withConditions(xpv1.Available()),
					withExternalName(sDeployTokenID),
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
			e := &external{kube: tc.kube, client: tc.deployToken}
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
				err: errors.New(errNotDeployToken),
			},
		},
		"SuccessfulCreation": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				deployToken: &fake.MockClient{
					MockCreateDeployToken: func(pid interface{}, opt *gitlab.CreateProjectDeployTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error) {
						return &deployTokenObj, &gitlab.Response{}, nil
					},
				},
				cr: deployToken(
					withAnnotations(extNameAnnotation),
					withSpec(v1alpha1.DeployTokenParameters{
						ProjectID: &deployTokenID,
					}),
				),
			},
			want: want{
				cr: deployToken(
					withExternalName(sDeployTokenID),
					withSpec(v1alpha1.DeployTokenParameters{
						ProjectID: &deployTokenID,
					}),
				),
				result: managed.ExternalCreation{
					ExternalNameAssigned: true,
					ConnectionDetails:    managed.ConnectionDetails{"token": []byte("Token")},
				},
			},
		},
		"FailedCreation": {
			args: args{
				deployToken: &fake.MockClient{
					MockCreateDeployToken: func(pid interface{}, opt *gitlab.CreateProjectDeployTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error) {
						return &gitlab.DeployToken{}, &gitlab.Response{}, errBoom
					},
				},
				cr: deployToken(
					withSpec(v1alpha1.DeployTokenParameters{
						ProjectID: &deployTokenID,
					}),
					withExternalName("0"),
				),
			},
			want: want{
				cr: deployToken(
					withSpec(v1alpha1.DeployTokenParameters{
						ProjectID: &deployTokenID,
					}),
					withExternalName("0"),
				),
				err: errors.Wrap(errBoom, errCreateFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.deployToken}
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
				cr: deployToken(),
			},
			want: want{
				cr: deployToken(),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.deployToken}
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
				err: errors.New(errNotDeployToken),
			},
		},
		"SuccessfulDeletion": {
			args: args{
				deployToken: &fake.MockClient{
					MockDeleteDeployToken: func(pid interface{}, deployToken int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: deployToken(
					withSpec(v1alpha1.DeployTokenParameters{
						ProjectID: &deployTokenID,
					}),
					withExternalName(strconv.Itoa(id)),
				),
			},
			want: want{
				cr: deployToken(
					withSpec(v1alpha1.DeployTokenParameters{
						ProjectID: &deployTokenID,
					}),
					withExternalName(strconv.Itoa(id)),
				),
			},
		},
		"FailedDeletionErrNotDeployToken": {
			args: args{
				deployToken: &fake.MockClient{
					MockDeleteDeployToken: func(pid interface{}, deployToken int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: deployToken(
					withSpec(v1alpha1.DeployTokenParameters{
						ProjectID: &deployTokenID,
					}),
					withExternalName("test"),
				),
			},
			want: want{
				cr: deployToken(
					withSpec(v1alpha1.DeployTokenParameters{
						ProjectID: &deployTokenID,
					}),
					withExternalName("test"),
				),
				err: errors.New(errNotDeployToken),
			},
		},
		"FailedDeletion": {
			args: args{
				deployToken: &fake.MockClient{
					MockDeleteDeployToken: func(pid interface{}, deployToken int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: deployToken(
					withSpec(v1alpha1.DeployTokenParameters{
						ProjectID: &deployTokenID,
					}),
					withExternalName(strconv.Itoa(id)),
				),
			},
			want: want{
				cr: deployToken(
					withSpec(v1alpha1.DeployTokenParameters{
						ProjectID: &deployTokenID,
					}),
					withExternalName(strconv.Itoa(id)),
				),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.deployToken}
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
