// /*
// Copyright 2021 The Crossplane Authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */

package accesstokens

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects/fake"
)

var (
	errBoom        = errors.New("boom")
	projectID      = ""
	wrongIDstr     = "fr"
	accessTokenID  = 1234
	sAccessTokenID = strconv.Itoa(accessTokenID)
	invalidInput   resource.Managed
	expiresAt      = time.Now().AddDate(0, 6, 0)
	accessLevel    = 40
	name           = "Access Token Name"
	token          = "Token"
	accessTokenObj = gitlab.ProjectAccessToken{
		ID:          accessTokenID,
		Name:        name,
		ExpiresAt:   (*gitlab.ISOTime)(&expiresAt),
		Token:       token,
		Scopes:      []string{"scope1", "scope2"},
		AccessLevel: 40, // Access level. Valid values are 10 (Guest), 20 (Reporter), 30 (Developer), 40 (Maintainer), and 50 (Owner). Defaults to 40.
	}

	extNameAnnotation = map[string]string{meta.AnnotationKeyExternalName: fmt.Sprint(accessTokenID)}
)

type args struct {
	accessTokenClient projects.AccessTokenClient
	kube              client.Client
	cr                resource.Managed
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
		"InvalidInput": {
			args: args{
				cr: invalidInput,
			},
			want: want{
				cr:  invalidInput,
				err: errors.New(errNotAccessToken),
			},
		},
		"NoExternalName": {
			args: args{
				cr: accessToken(),
			},
			want: want{
				cr:     accessToken(),
				result: managed.ExternalObservation{},
				err:    nil,
			},
		},
		"ExternalNameNotID": {
			args: args{
				cr: accessToken(withExternalName(wrongIDstr)),
			},
			want: want{
				cr:     accessToken(withExternalName(wrongIDstr)),
				result: managed.ExternalObservation{},
				err:    errors.Wrap(getConversionError(), errFailedParseID),
			},
		},
		"NoProjectID": {
			args: args{
				cr: accessToken(withExternalName(sAccessTokenID)),
			},
			want: want{
				cr:     accessToken(withExternalName(sAccessTokenID)),
				result: managed.ExternalObservation{},
				err:    errors.New(errMissingProjectID),
			},
		},
		"ErrGetAccessToken": {
			args: args{
				accessTokenClient: &fake.MockClient{
					MockGetProjectAccessToken: func(pid interface{}, id int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectAccessToken, *gitlab.Response, error) {
						return nil, nil, errBoom
					},
				},
				cr: accessToken(
					withExternalName(sAccessTokenID),
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &projectID,
					}),
				),
			},
			want: want{
				cr: accessToken(
					withExternalName(sAccessTokenID),
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &projectID,
					}),
				),
				result: managed.ExternalObservation{},
				err:    errors.Wrap(errBoom, errAccessTokentNotFound),
			},
		},
		"GetErr404": {
			args: args{
				accessTokenClient: &fake.MockClient{
					MockGetProjectAccessToken: func(pid interface{}, id int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectAccessToken, *gitlab.Response, error) {
						return nil, &gitlab.Response{
							Response: &http.Response{StatusCode: http.StatusNotFound},
						}, errBoom
					},
				},
				cr: accessToken(
					withExternalName(sAccessTokenID),
					withSpec(v1alpha1.AccessTokenParameters{ProjectID: &projectID}),
				),
			},
			want: want{
				cr: accessToken(
					withExternalName(sAccessTokenID),
					withSpec(v1alpha1.AccessTokenParameters{ProjectID: &projectID}),
				),
				result: managed.ExternalObservation{},
				err:    nil,
			},
		},
		"AccessTokenDoNotExist": {
			args: args{
				accessTokenClient: &fake.MockClient{
					MockGetProjectAccessToken: func(pid interface{}, id int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectAccessToken, *gitlab.Response, error) {
						return nil, nil, errBoom
					},
				},
				cr: accessToken(
					withExternalName(sAccessTokenID),
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &projectID,
					}),
				),
			},
			want: want{
				cr: accessToken(
					withExternalName(sAccessTokenID),
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &projectID,
					}),
				),
				result: managed.ExternalObservation{},
				err:    errors.Wrap(errBoom, errAccessTokentNotFound),
			},
		},
		"ResourceLateInitializedFalse": {
			args: args{
				accessTokenClient: &fake.MockClient{
					MockGetProjectAccessToken: func(pid interface{}, id int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectAccessToken, *gitlab.Response, error) {
						return &gitlab.ProjectAccessToken{}, &gitlab.Response{}, nil
					},
				},
				cr: accessToken(
					withExternalName(sAccessTokenID),
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID:   &projectID,
						AccessLevel: (*v1alpha1.AccessLevelValue)(&accessLevel),
						ExpiresAt:   &v1.Time{Time: expiresAt},
					}),
				),
			},
			want: want{
				cr: accessToken(
					withExternalName(sAccessTokenID),
					withConditions(xpv1.Available()),
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID:   &projectID,
						AccessLevel: (*v1alpha1.AccessLevelValue)(&accessLevel),
						ExpiresAt:   &v1.Time{Time: expiresAt},
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: false,
				},
			},
		},
		"ResourceLateInitializedTrue": {
			args: args{
				accessTokenClient: &fake.MockClient{
					MockGetProjectAccessToken: func(pid interface{}, id int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectAccessToken, *gitlab.Response, error) {
						return &gitlab.ProjectAccessToken{
							ExpiresAt:   accessTokenObj.ExpiresAt,
							AccessLevel: accessTokenObj.AccessLevel,
						}, &gitlab.Response{}, nil
					},
				},
				cr: accessToken(
					withExternalName(sAccessTokenID),
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &projectID,
					}),
				),
			},
			want: want{
				cr: accessToken(
					withExternalName(sAccessTokenID),
					withConditions(xpv1.Available()),
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID:   &projectID,
						ExpiresAt:   &v1.Time{Time: expiresAt},
						AccessLevel: (*v1alpha1.AccessLevelValue)(&accessLevel),
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
				},
			},
		},
		"TokenUpToDate": {
			args: args{
				accessTokenClient: &fake.MockClient{
					MockGetProjectAccessToken: func(pid interface{}, id int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectAccessToken, *gitlab.Response, error) {
						return &gitlab.ProjectAccessToken{}, &gitlab.Response{}, nil
					},
				},
				cr: accessToken(
					withExternalName(sAccessTokenID),
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID:   &projectID,
						AccessLevel: (*v1alpha1.AccessLevelValue)(&accessLevel),
						ExpiresAt:   &v1.Time{Time: expiresAt},
					}),
				),
			},
			want: want{
				cr: accessToken(
					withExternalName(sAccessTokenID),
					withConditions(xpv1.Available()),
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID:   &projectID,
						AccessLevel: (*v1alpha1.AccessLevelValue)(&accessLevel),
						ExpiresAt:   &v1.Time{Time: expiresAt},
					}),
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
			e := &external{kube: tc.kube, client: tc.accessTokenClient}
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
		"InvalidInput": {
			args: args{
				cr: invalidInput,
			},
			want: want{
				cr:     invalidInput,
				err:    errors.New(errNotAccessToken),
				result: managed.ExternalCreation{},
			},
		},
		"NoProjectID": {
			args: args{
				cr: accessToken(withExternalName(sAccessTokenID)),
			},
			want: want{
				cr:     accessToken(withExternalName(sAccessTokenID)),
				result: managed.ExternalCreation{},
				err:    errors.New(errMissingProjectID),
			},
		},
		"CreationFailedErr": {
			args: args{
				accessTokenClient: &fake.MockClient{
					MockCreateProjectAccessToken: func(pid interface{}, opt *gitlab.CreateProjectAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectAccessToken, *gitlab.Response, error) {
						return nil, nil, errBoom
					},
				},
				cr: accessToken(
					withExternalName(sAccessTokenID),
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &projectID,
					}),
				),
			},
			want: want{
				cr: accessToken(
					withExternalName(sAccessTokenID),
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &projectID,
					}),
				),
				result: managed.ExternalCreation{},
				err:    errors.Wrap(errBoom, errCreateFailed),
			},
		},
		"NoExternalName": {
			args: args{
				accessTokenClient: &fake.MockClient{
					MockCreateProjectAccessToken: func(pid interface{}, opt *gitlab.CreateProjectAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectAccessToken, *gitlab.Response, error) {
						return &gitlab.ProjectAccessToken{}, &gitlab.Response{}, errBoom
					},
				},
				cr: accessToken(
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &projectID,
					}),
					withExternalName("0"),
				),
			},
			want: want{
				cr: accessToken(
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &projectID,
					}),
					withExternalName("0"),
				),
				err: errors.Wrap(errBoom, errCreateFailed),
			},
		},
		"CreationSuccessful": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				accessTokenClient: &fake.MockClient{
					MockCreateProjectAccessToken: func(pid interface{}, opt *gitlab.CreateProjectAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectAccessToken, *gitlab.Response, error) {
						return &accessTokenObj, &gitlab.Response{}, nil
					},
				},
				cr: accessToken(
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &projectID,
					}),
					withAnnotations(extNameAnnotation),
				),
			},
			want: want{
				cr: accessToken(
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &projectID,
					}),
					withExternalName(sAccessTokenID),
				),
				result: managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{"token": []byte(token)},
				},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.accessTokenClient}
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
				cr:     accessToken(),
				result: managed.ExternalUpdate{},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.accessTokenClient}
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
		"InvalidInput": {
			args: args{
				cr: invalidInput,
			},
			want: want{
				cr:  invalidInput,
				err: errors.New(errNotAccessToken),
			},
		},
		"FailedDeletionExternalNameNotInt": {
			args: args{
				accessTokenClient: &fake.MockClient{
					MockRevokeProjectAccessToken: func(pid interface{}, id int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: accessToken(
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &projectID,
					}),
					withExternalName("test"),
				),
			},
			want: want{
				cr: accessToken(
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &projectID,
					}),
					withExternalName("test"),
				),
				err: errors.New(errExternalNameNotInt),
			},
		},
		"FailedDeletion": {
			args: args{
				accessTokenClient: &fake.MockClient{
					MockRevokeProjectAccessToken: func(pid interface{}, id int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: accessToken(
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &projectID,
					}),
					withExternalName(strconv.Itoa(0)),
				),
			},
			want: want{
				cr: accessToken(
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &projectID,
					}),
					withExternalName(strconv.Itoa(0)),
				),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
		"SuccessfulDeletion": {
			args: args{
				accessTokenClient: &fake.MockClient{
					MockRevokeProjectAccessToken: func(pid interface{}, id int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: accessToken(
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &projectID,
					}),
					withExternalName(strconv.Itoa(accessTokenID)),
				),
			},
			want: want{
				cr: accessToken(
					withSpec(v1alpha1.AccessTokenParameters{
						ProjectID: &projectID,
					}),
					withExternalName(strconv.Itoa(accessTokenID)),
				),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.accessTokenClient}
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

func getConversionError() error {
	_, err := strconv.Atoi(wrongIDstr)
	return err
}
