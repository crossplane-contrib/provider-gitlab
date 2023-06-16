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

package accesstokens

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
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects/fake"
)

var (
	errBoom        = errors.New("boom")
	id             = 0
	accessTokenID  = 1234
	sAccessTokenID = strconv.Itoa(accessTokenID)
	unexpecedItem  resource.Managed
	expiresAt      = time.Now()
	accessLevel    = 40
	token          = "Token"
	accessTokenObj = gitlab.ProjectAccessToken{
		ID:          accessTokenID,
		Name:        "Name",
		ExpiresAt:   (*gitlab.ISOTime)(&expiresAt),
		Token:       token,
		Scopes:      []string{"scope1", "scope2"},
		AccessLevel: 40, // Access level. Valid values are 10 (Guest), 20 (Reporter), 30 (Developer), 40 (Maintainer), and 50 (Owner). Defaults to 40.
	}

	accessTokens = []*gitlab.ProjectAccessToken{
		{
			ID:          accessTokenID,
			Name:        "Name",
			ExpiresAt:   (*gitlab.ISOTime)(&expiresAt),
			Token:       token,
			Scopes:      []string{"scope1", "scope2"},
			AccessLevel: gitlab.AccessLevelValue(accessLevel),
		},
	}

	extNameAnnotation = map[string]string{meta.AnnotationKeyExternalName: fmt.Sprint(accessTokenID)}
)

type args struct {
	accessToken gitlab.ProjectAccessTokensService
	kube        client.Client
	cr          resource.Managed
}

type accessTokenModifier func(*v1alpha1.AccessToken)

func withConditions(c ...xpv1.Condition) accessTokenModifier {
	return func(r *v1alpha1.AccessToken) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(fp v1alpha1.AccessTokenParameters) accessTokenModifier {
	return func(r *v1alpha1.AccessToken) { r.Spec.ForProvider = fp }
}

func withExternalName(accessTokenID string) accessTokenModifier {
	return func(r *v1alpha1.AccessToken) { meta.SetExternalName(r, accessTokenID) }
}

func withAnnotations(a map[string]string) accessTokenModifier {
	return func(p *v1alpha1.AccessToken) { meta.AddAnnotations(p, a) }
}

func accessToken(m ...accessTokenModifier) *v1alpha1.AccessToken {
	cr := &v1alpha1.AccessToken{}
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
				err: errors.New(errNotAccessToken),
			},
		},
		"NoExternalName": {
			args: args{
				cr: accessToken(),
			},
			want: want{
				cr: accessToken(),
				result: managed.ExternalObservation{
					ResourceExists:          false,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
			},
		},
		"NotIDExternalName": {
			args: args{
				accessToken: &fake.MockClient{
					MockListAccessTokens: func(pid interface{}, opt *gitlab.ListProjectAccessTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectAccessToken, *gitlab.Response, error) {
						return []*gitlab.ProjectAccessToken{}, &gitlab.Response{}, nil
					},
				},
				cr: accessToken(withExternalName("fr")),
			},
			want: want{
				cr:  accessToken(withExternalName("fr")),
				err: errors.New(errNotAccessToken),
			},
		},
		"FailedGetRequest": {
			args: args{
				accessToken: &fake.MockClient{
					MockListAccessTokens: func(pid interface{}, opt *gitlab.ListProjectAccessTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectAccessToken, *gitlab.Response, error) {
						return []*gitlab.ProjectAccessToken{}, &gitlab.Response{}, errBoom
					},
				},
				cr: accessToken(
					withExternalName(sAccessTokenID),
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &accessTokenID,
					}),
				),
			},
			want: want{
				cr: accessToken(
					withAnnotations(extNameAnnotation),
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &accessTokenID,
					}),
				),
				err: errors.Wrap(errBoom, errGetFailed),
			},
		},
		"AccessTokenNotFound": {
			args: args{
				accessToken: &fake.MockClient{
					MockListAccessTokens: func(pid interface{}, opt *gitlab.ListProjectAccessTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectAccessToken, *gitlab.Response, error) {
						return accessTokens, &gitlab.Response{}, nil
					},
				},
				cr: accessToken(
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID:   &accessTokenID,
						AccessLevel: &accessLevel,
						ExpiresAt:   &metav1.Time{Time: expiresAt},
					}),
					withExternalName("3"),
				),
			},
			want: want{
				cr: accessToken(
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID:   &accessTokenID,
						AccessLevel: &accessLevel,
						ExpiresAt:   &metav1.Time{Time: expiresAt},
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
				accessToken: &fake.MockClient{
					MockListAccessTokens: func(pid interface{}, opt *gitlab.ListProjectAccessTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectAccessToken, *gitlab.Response, error) {
						return accessTokens, &gitlab.Response{}, nil
					},
				},
				cr: accessToken(
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &accessTokenID,
					}),
					withExternalName(sAccessTokenID),
				),
			},
			want: want{
				cr: accessToken(
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID:   &accessTokenID,
						AccessLevel: &accessLevel,
						ExpiresAt:   &metav1.Time{Time: expiresAt},
					}),
					withConditions(xpv1.Available()),
					withExternalName(sAccessTokenID),
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
				accessToken: &fake.MockClient{
					MockListAccessTokens: func(pid interface{}, opt *gitlab.ListProjectAccessTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectAccessToken, *gitlab.Response, error) {
						return accessTokens, &gitlab.Response{}, nil
					},
				},
				cr: accessToken(
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID:   &accessTokenID,
						AccessLevel: &accessLevel,
						ExpiresAt:   &metav1.Time{Time: expiresAt},
					}),
					withExternalName(sAccessTokenID),
				),
			},
			want: want{
				cr: accessToken(
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID:   &accessTokenID,
						AccessLevel: &accessLevel,
						ExpiresAt:   &metav1.Time{Time: expiresAt},
					}),
					withConditions(xpv1.Available()),
					withExternalName(sAccessTokenID),
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
			e := &external{kube: tc.kube, client: tc.accessToken}
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
				err: errors.New(errNotAccessToken),
			},
		},
		"SuccessfulCreation": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				accessToken: &fake.MockClient{
					MockCreateAccessToken: func(pid interface{}, opt *gitlab.CreateProjectAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectAccessToken, *gitlab.Response, error) {
						return &accessTokenObj, &gitlab.Response{}, nil
					},
				},
				cr: accessToken(
					withAnnotations(extNameAnnotation),
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &accessTokenID,
					}),
				),
			},
			want: want{
				cr: accessToken(
					withExternalName(sAccessTokenID),
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &accessTokenID,
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
				accessToken: &fake.MockClient{
					MockCreateAccessToken: func(pid interface{}, opt *gitlab.CreateProjectAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectAccessToken, *gitlab.Response, error) {
						return &gitlab.ProjectAccessToken{}, &gitlab.Response{}, errBoom
					},
				},
				cr: accessToken(
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &accessTokenID,
					}),
					withExternalName("0"),
				),
			},
			want: want{
				cr: accessToken(
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &accessTokenID,
					}),
					withExternalName("0"),
				),
				err: errors.Wrap(errBoom, errCreateFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.accessToken}
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
				cr: accessToken(),
			},
			want: want{
				cr: accessToken(),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.accessToken}
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
				err: errors.New(errNotAccessToken),
			},
		},
		"SuccessfulDeletion": {
			args: args{
				accessToken: &fake.MockClient{
					MockDeleteAccessToken: func(pid interface{}, accessToken int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: accessToken(
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &accessTokenID,
					}),
					withExternalName(strconv.Itoa(id)),
				),
			},
			want: want{
				cr: accessToken(
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &accessTokenID,
					}),
					withExternalName(strconv.Itoa(id)),
				),
			},
		},
		"FailedDeletionErrNotAccessToken": {
			args: args{
				accessToken: &fake.MockClient{
					MockDeleteAccessToken: func(pid interface{}, accessToken int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: accessToken(
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &accessTokenID,
					}),
					withExternalName("test"),
				),
			},
			want: want{
				cr: accessToken(
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &accessTokenID,
					}),
					withExternalName("test"),
				),
				err: errors.New(errNotAccessToken),
			},
		},
		"FailedDeletion": {
			args: args{
				accessToken: &fake.MockClient{
					MockDeleteAccessToken: func(pid interface{}, accessToken int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: accessToken(
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &accessTokenID,
					}),
					withExternalName(strconv.Itoa(id)),
				),
			},
			want: want{
				cr: accessToken(
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &accessTokenID,
					}),
					withExternalName(strconv.Itoa(id)),
				),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.accessToken}
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
