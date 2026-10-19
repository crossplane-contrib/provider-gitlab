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

package serviceaccounts

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/errors"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/google/go-cmp/cmp"
	pkgerrors "github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	commonv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/instance/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	groupsfake "github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/groups/fake"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/instance"
)

var (
	unexpectedItem resource.Managed
	errBoom        = errors.New("boom")
)

const invalidBaselinePermissionsFilter = "["

// testAvatar is the PNG signature, enough for content sniffing. GitLab serves it
// from testAvatarURL after an upload under its content addressed name.
var testAvatar = []byte("\x89PNG\r\n\x1a\n")

const (
	testAvatarName  = "4c4b6a3be1314ab86138bef4314dde022e600960d8689a2c8f8631802d20dab6.png"
	testAvatarURL   = "https://gitlab.example.com/uploads/-/system/user/avatar/123/" + testAvatarName + "?v=1"
	testGravatarURL = "https://secure.gravatar.com/avatar/abc?s=80&d=identicon"
)

// testSecretAvatar is the GIF signature an avatar Secret holds in the
// avatarSecretRef cases, distinct from testAvatar so a change is visible.
var testSecretAvatar = []byte("GIF89a")

const (
	testSecretAvatarName = "610f5ae4d76e332636a17bd357fd6ce99029316a99d320280d4d77a746bf29e8.gif"
	testAvatarSecret     = "sa-avatar"
	testAvatarSecretKey  = "avatar"
	errAvatarKeyMissing  = `cannot read avatar from key "avatar" of Secret "sa-avatar" in namespace "default": Cannot find key in referenced secret`
	errAvatarSecretGone  = `cannot read avatar from key "avatar" of Secret "sa-avatar" in namespace "default": Cannot find referenced secret: secrets "sa-avatar" not found`
)

// avatarKube returns a kube client that holds the avatar Secret with data, or
// no Secret at all when data is nil.
func avatarKube(data map[string][]byte) client.Client {
	b := fake.NewClientBuilder()
	if data != nil {
		b = b.WithObjects(&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: testAvatarSecret}, Data: data})
	}
	return b.Build()
}

func withAvatarSecretRef() serviceAccountModifier {
	return func(r *v1alpha1.ServiceAccount) {
		r.Namespace = "default"
		r.Spec.ForProvider.AvatarSecretRef = common.TestCreateLocalSecretKeySelector(testAvatarSecret, testAvatarSecretKey)
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// compileError returns the error raised by the standard library when compiling the given regular expression.
func compileError(pattern string) error {
	_, err := regexp.Compile(pattern)
	return err
}

const (
	testServiceAccountName     = "sa"
	testServiceAccountUsername = "sa-user"
	testServiceAccountEmail    = "sa@example.org"
)

// MockClient is a small mock for instance.ServiceAccountClient used by controller tests.
type MockClient struct {
	MockGetUser                  func(user int64, opt *gitlab.GetUserOptions, options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error)
	MockCreateServiceAccountUser func(opts *gitlab.CreateServiceAccountUserOptions, options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error)
	MockModifyUser               func(user int64, opt *gitlab.ModifyUserOptions, options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error)
	MockDeleteUser               func(user int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

func (m *MockClient) GetUser(user int64, opt *gitlab.GetUserOptions, options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
	return m.MockGetUser(user, opt, options...)
}

func (m *MockClient) CreateServiceAccountUser(opts *gitlab.CreateServiceAccountUserOptions, options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
	return m.MockCreateServiceAccountUser(opts, options...)
}

func (m *MockClient) ModifyUser(user int64, opt *gitlab.ModifyUserOptions, options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
	return m.MockModifyUser(user, opt, options...)
}

func (m *MockClient) DeleteUser(user int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return m.MockDeleteUser(user, options...)
}

type args struct {
	client instance.ServiceAccountClient
	kube   client.Client
	cr     resource.Managed
}

type serviceAccountModifier func(*v1alpha1.ServiceAccount)

func withExternalName(n string) serviceAccountModifier {
	return func(r *v1alpha1.ServiceAccount) { meta.SetExternalName(r, n) }
}

func withSpec(p v1alpha1.ServiceAccountParameters) serviceAccountModifier {
	return func(r *v1alpha1.ServiceAccount) { r.Spec.ForProvider = p }
}

func withConditions(c ...v2.Condition) serviceAccountModifier {
	return func(r *v1alpha1.ServiceAccount) { r.Status.SetConditions(c...) }
}

func withDeletionTimestamp(t metav1.Time) serviceAccountModifier {
	return func(r *v1alpha1.ServiceAccount) { r.ObjectMeta.DeletionTimestamp = &t }
}

func withAtProvider(o v1alpha1.ServiceAccountObservation) serviceAccountModifier {
	return func(r *v1alpha1.ServiceAccount) { r.Status.AtProvider = o }
}

func serviceAccount(m ...serviceAccountModifier) *v1alpha1.ServiceAccount {
	cr := &v1alpha1.ServiceAccount{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestConnect(t *testing.T) {
	type want struct {
		err error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"InValidInput": {
			args: args{cr: unexpectedItem},
			want: want{err: errors.New(errNotServiceAccount)},
		},
		// This is the most common connector edge case in this repo: no providerConfigRef.
		"ProviderConfigRefNotGivenError": {
			args: args{cr: serviceAccount(), kube: &test.MockClient{MockGet: test.NewMockGetFn(nil)}},
			want: want{err: pkgerrors.New("providerConfigRef is not given")},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &connector{kube: tc.args.kube, newGitlabClientFn: nil}
			got, err := c.Connect(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Connect(): -want, +got:\n%s", diff)
			}
			_ = got
		})
	}
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     resource.Managed
		result managed.ExternalObservation
		err    error
	}

	// Shared desired state used by multiple cases.
	desired := v1alpha1.ServiceAccountParameters{CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{Name: sPtr(testServiceAccountName), Username: sPtr(testServiceAccountUsername), Email: sPtr(testServiceAccountEmail)}}

	cases := map[string]struct {
		args args
		want want
	}{
		"InValidInput": {
			args: args{cr: unexpectedItem},
			want: want{cr: unexpectedItem, err: errors.New(errNotServiceAccount)},
		},
		"NoExternalName": {
			args: args{cr: serviceAccount(withSpec(desired))},
			want: want{cr: serviceAccount(withSpec(desired)), result: managed.ExternalObservation{ResourceExists: false}},
		},
		"NotIDExternalName": {
			args: args{client: &MockClient{MockGetUser: func(user int64, opt *gitlab.GetUserOptions, options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
				return &gitlab.User{}, &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
			}}, cr: serviceAccount(withExternalName("fr"), withSpec(desired))},
			want: want{cr: serviceAccount(withExternalName("fr"), withSpec(desired)), err: errors.New(errIDNotInt)},
		},
		"ErrGet": {
			args: args{client: &MockClient{MockGetUser: func(user int64, opt *gitlab.GetUserOptions, options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
				return nil, &gitlab.Response{Response: &http.Response{StatusCode: 500}}, errBoom
			}}, cr: serviceAccount(withExternalName("123"), withSpec(desired))},
			want: want{cr: serviceAccount(withExternalName("123"), withSpec(desired)), err: errors.Wrap(errBoom, errGetFailed)},
		},
		"ErrGet404": {
			args: args{client: &MockClient{MockGetUser: func(user int64, opt *gitlab.GetUserOptions, options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
				return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errors.New("not found")
			}}, cr: serviceAccount(withExternalName("123"), withSpec(desired))},
			want: want{cr: serviceAccount(withExternalName("123"), withSpec(desired)), result: managed.ExternalObservation{}},
		},
		"DeletingShouldAllowCRDeletion": func() struct {
			args args
			want want
		} {
			now := metav1.NewTime(time.Now())
			sa := serviceAccount(withExternalName("123"), withSpec(desired), withDeletionTimestamp(now))
			return struct {
				args args
				want want
			}{
				args: args{client: &MockClient{MockGetUser: func(user int64, opt *gitlab.GetUserOptions, options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
					return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errors.New("not found")
				}}, cr: sa},
				want: want{cr: sa, result: managed.ExternalObservation{}},
			}
		}(),
		"SuccessfulAvailableUpToDate": {
			args: args{client: &MockClient{MockGetUser: func(user int64, opt *gitlab.GetUserOptions, options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
				return &gitlab.User{ID: 123, Name: testServiceAccountName, Username: testServiceAccountUsername, Email: testServiceAccountEmail}, &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
			}}, cr: serviceAccount(withExternalName("123"), withSpec(desired))},
			want: want{
				cr: serviceAccount(
					withExternalName("123"),
					withSpec(desired),
					withConditions(v2.Available()),
					withAtProvider(instance.GenerateServiceAccountObservation(&gitlab.User{ID: 123, Name: testServiceAccountName, Username: testServiceAccountUsername, Email: testServiceAccountEmail})),
				),
				result: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true, ResourceLateInitialized: false},
			},
		},
		"SuccessfulAvailableNotUpToDate": {
			args: args{client: &MockClient{MockGetUser: func(user int64, opt *gitlab.GetUserOptions, options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
				return &gitlab.User{ID: 123, Name: testServiceAccountName, Username: testServiceAccountUsername, Email: "different@example.org"}, &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
			}}, cr: serviceAccount(withExternalName("123"), withSpec(desired))},
			want: want{
				cr: serviceAccount(
					withExternalName("123"),
					withSpec(desired),
					withConditions(v2.Available()),
					withAtProvider(instance.GenerateServiceAccountObservation(&gitlab.User{ID: 123, Name: testServiceAccountName, Username: testServiceAccountUsername, Email: "different@example.org"})),
				),
				result: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false, ResourceLateInitialized: false},
			},
		},
		"SuccessfulBaselinePermissionsSortsStatus": {
			args: args{
				client: &MockClient{MockGetUser: func(user int64, opt *gitlab.GetUserOptions, options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
					return &gitlab.User{ID: 123, Name: testServiceAccountName, Username: testServiceAccountUsername, Email: testServiceAccountEmail}, &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
				}},
				cr: serviceAccount(withExternalName("123"), withSpec(v1alpha1.ServiceAccountParameters{
					CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{Name: sPtr(testServiceAccountName), Username: sPtr(testServiceAccountUsername), Email: sPtr(testServiceAccountEmail)},
					BaselinePermissions:            ptr.To(accessLevelDeveloper),
				})),
			},
			want: want{
				cr: serviceAccount(
					withExternalName("123"),
					withSpec(v1alpha1.ServiceAccountParameters{
						CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{Name: sPtr(testServiceAccountName), Username: sPtr(testServiceAccountUsername), Email: sPtr(testServiceAccountEmail)},
						BaselinePermissions:            ptr.To(accessLevelDeveloper),
					}),
					withConditions(v2.Available()),
					func(r *v1alpha1.ServiceAccount) {
						r.Status.AtProvider = instance.GenerateServiceAccountObservation(&gitlab.User{ID: 123, Name: testServiceAccountName, Username: testServiceAccountUsername, Email: testServiceAccountEmail})
						r.Status.AtProvider.MissingMemberShipGroups = []int64{1, 3}
						r.Status.AtProvider.WrongPermissionsGroups = []int64{2, 4}
					},
				),
				result: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false, ResourceLateInitialized: false},
			},
		},
		"SuccessfulBaselinePermissionsFiltersGroups": {
			args: args{
				client: &MockClient{MockGetUser: func(user int64, opt *gitlab.GetUserOptions, options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
					return &gitlab.User{ID: 123, Name: testServiceAccountName, Username: testServiceAccountUsername, Email: testServiceAccountEmail}, &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
				}},
				cr: serviceAccount(withExternalName("123"), withSpec(v1alpha1.ServiceAccountParameters{
					CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{Name: sPtr(testServiceAccountName), Username: sPtr(testServiceAccountUsername), Email: sPtr(testServiceAccountEmail)},
					BaselinePermissions:            ptr.To(accessLevelDeveloper),
					BaselinePermissionsFilters:     []string{"^platform-"},
				})),
			},
			want: want{
				cr: serviceAccount(
					withExternalName("123"),
					withSpec(v1alpha1.ServiceAccountParameters{
						CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{Name: sPtr(testServiceAccountName), Username: sPtr(testServiceAccountUsername), Email: sPtr(testServiceAccountEmail)},
						BaselinePermissions:            ptr.To(accessLevelDeveloper),
						BaselinePermissionsFilters:     []string{"^platform-"},
					}),
					withConditions(v2.Available()),
					func(r *v1alpha1.ServiceAccount) {
						r.Status.AtProvider = instance.GenerateServiceAccountObservation(&gitlab.User{ID: 123, Name: testServiceAccountName, Username: testServiceAccountUsername, Email: testServiceAccountEmail})
						r.Status.AtProvider.MissingMemberShipGroups = []int64{1}
					},
				),
				result: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false, ResourceLateInitialized: false},
			},
		},
		"BaselinePermissionsInvalidFilter": {
			args: args{
				client: &MockClient{MockGetUser: func(user int64, opt *gitlab.GetUserOptions, options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
					return &gitlab.User{ID: 123, Name: testServiceAccountName, Username: testServiceAccountUsername, Email: testServiceAccountEmail}, &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
				}},
				cr: serviceAccount(withExternalName("123"), withSpec(v1alpha1.ServiceAccountParameters{
					CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{Name: sPtr(testServiceAccountName), Username: sPtr(testServiceAccountUsername), Email: sPtr(testServiceAccountEmail)},
					BaselinePermissions:            ptr.To(accessLevelDeveloper),
					BaselinePermissionsFilters:     []string{invalidBaselinePermissionsFilter},
				})),
			},
			want: want{
				cr: serviceAccount(
					withExternalName("123"),
					withSpec(v1alpha1.ServiceAccountParameters{
						CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{Name: sPtr(testServiceAccountName), Username: sPtr(testServiceAccountUsername), Email: sPtr(testServiceAccountEmail)},
						BaselinePermissions:            ptr.To(accessLevelDeveloper),
						BaselinePermissionsFilters:     []string{invalidBaselinePermissionsFilter},
					}),
					withConditions(v2.Available()),
					withAtProvider(instance.GenerateServiceAccountObservation(&gitlab.User{ID: 123, Name: testServiceAccountName, Username: testServiceAccountUsername, Email: testServiceAccountEmail})),
				),
				err: errors.Wrapf(compileError(invalidBaselinePermissionsFilter), errInvalidFilter, invalidBaselinePermissionsFilter),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.args.kube, client: tc.args.client}
			if name == "SuccessfulBaselinePermissionsFiltersGroups" {
				e.groupsClient = &groupsfake.MockClient{MockListGroups: func(opt *gitlab.ListGroupsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.Group, *gitlab.Response, error) {
					return []*gitlab.Group{{ID: 1, Path: "platform-tooling"}, {ID: 2, Path: "customers"}}, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}, NextPage: 0}, nil
				}}
				e.groupMemberClient = &groupsfake.MockClient{MockGetMember: func(gid interface{}, user int64, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
					if gid.(int64) != 1 {
						t.Fatalf("Observe() checked membership of filtered out group %v", gid)
					}
					return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, pkgerrors.New("not found")
				}}
			}
			if name == "SuccessfulBaselinePermissionsSortsStatus" {
				e.groupsClient = &groupsfake.MockClient{MockListGroups: func(opt *gitlab.ListGroupsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.Group, *gitlab.Response, error) {
					return []*gitlab.Group{{ID: 4}, {ID: 1}, {ID: 2}, {ID: 3}}, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}, NextPage: 0}, nil
				}}
				e.groupMemberClient = &groupsfake.MockClient{MockGetMember: func(gid interface{}, user int64, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
					switch gid.(int64) {
					case 4:
						return &gitlab.GroupMember{AccessLevel: gitlab.AccessLevelValue(accessLevelReporterValue)}, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					case 1:
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, pkgerrors.New("not found")
					case 2:
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					case 3:
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, pkgerrors.New("not found")
					default:
						return nil, nil, pkgerrors.New("unexpected group")
					}
				}}
			}
			got, err := e.Observe(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Observe(): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, got); diff != "" {
				t.Errorf("Observe(): -want result, +got result:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("Observe(): -want cr, +got cr:\n%s", diff)
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

	desired := v1alpha1.ServiceAccountParameters{CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{Name: sPtr(testServiceAccountName), Username: sPtr(testServiceAccountUsername), Email: sPtr(testServiceAccountEmail)}}

	cases := map[string]struct {
		args args
		want want
	}{
		"InValidInput": {
			args: args{cr: unexpectedItem},
			want: want{cr: unexpectedItem, err: errors.New(errNotServiceAccount)},
		},
		"ErrCreate": {
			args: args{client: &MockClient{MockCreateServiceAccountUser: func(opts *gitlab.CreateServiceAccountUserOptions, options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
				return nil, nil, errBoom
			}}, cr: serviceAccount(withSpec(desired))},
			want: want{cr: serviceAccount(withSpec(desired), withConditions(v2.Creating())), err: errors.Wrap(errBoom, errCreateFailed)},
		},
		"Successful": {
			args: args{client: &MockClient{MockCreateServiceAccountUser: func(opts *gitlab.CreateServiceAccountUserOptions, options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
				return &gitlab.User{}, &gitlab.Response{Response: &http.Response{StatusCode: 201}}, nil
			}}, cr: serviceAccount(withSpec(desired))},
			want: want{cr: serviceAccount(withSpec(desired), withConditions(v2.Creating()), withExternalName("0")), result: managed.ExternalCreation{}},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.args.kube, client: tc.args.client}
			got, err := e.Create(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Create(): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, got); diff != "" {
				t.Errorf("Create(): -want result, +got result:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("Create(): -want cr, +got cr:\n%s", diff)
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

	desired := v1alpha1.ServiceAccountParameters{CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{Name: sPtr(testServiceAccountName), Username: sPtr(testServiceAccountUsername), Email: sPtr(testServiceAccountEmail)}}

	cases := map[string]struct {
		args args
		want want
	}{
		"InValidInput": {
			args: args{cr: unexpectedItem},
			want: want{cr: unexpectedItem, err: errors.New(errNotServiceAccount)},
		},
		"NoExternalName": {
			args: args{client: &MockClient{MockModifyUser: func(user int64, opt *gitlab.ModifyUserOptions, options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
				return &gitlab.User{}, &gitlab.Response{}, nil
			}}, cr: serviceAccount(withSpec(desired))},
			want: want{cr: serviceAccount(withSpec(desired), withConditions(v2.Creating())), err: errors.New(errCreateFailed)},
		},
		"NotIDExternalName": {
			args: args{client: &MockClient{MockModifyUser: func(user int64, opt *gitlab.ModifyUserOptions, options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
				return &gitlab.User{}, &gitlab.Response{}, nil
			}}, cr: serviceAccount(withExternalName("fr"), withSpec(desired))},
			want: want{cr: serviceAccount(withExternalName("fr"), withSpec(desired), withConditions(v2.Creating())), err: errors.New(errIDNotInt)},
		},
		"ErrUpdate": {
			args: args{client: &MockClient{MockModifyUser: func(user int64, opt *gitlab.ModifyUserOptions, options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
				return nil, nil, errBoom
			}}, cr: serviceAccount(withExternalName("123"), withSpec(desired))},
			want: want{cr: serviceAccount(withExternalName("123"), withSpec(desired), withConditions(v2.Creating())), err: errors.Wrap(errBoom, errUpdateFailed)},
		},
		"Successful": {
			args: args{client: &MockClient{MockModifyUser: func(user int64, opt *gitlab.ModifyUserOptions, options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
				return &gitlab.User{}, &gitlab.Response{}, nil
			}}, cr: serviceAccount(withExternalName("123"), withSpec(desired))},
			want: want{cr: serviceAccount(withExternalName("123"), withSpec(desired), withConditions(v2.Creating())), result: managed.ExternalUpdate{}},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.args.kube, client: tc.args.client}
			got, err := e.Update(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Update(): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, got); diff != "" {
				t.Errorf("Update(): -want result, +got result:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("Update(): -want cr, +got cr:\n%s", diff)
			}
		})
	}
}

func TestUpdateAvatar(t *testing.T) {
	cases := map[string]struct {
		avatar       []byte
		avatarURL    string
		wantFilename string
		wantErr      bool
	}{
		"UploadsChangedAvatar": {avatar: testAvatar, avatarURL: testGravatarURL, wantFilename: testAvatarName},
		"SkipsCurrentAvatar":   {avatar: testAvatar, avatarURL: testAvatarURL},
		"SkipsUnmanagedAvatar": {avatarURL: testGravatarURL},
		"RejectsUnsupportedAvatar": {
			avatar:    []byte("not an image"),
			avatarURL: testGravatarURL,
			wantErr:   true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var got *gitlab.ModifyUserOptions
			e := &external{client: &MockClient{MockModifyUser: func(_ int64, opt *gitlab.ModifyUserOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
				got = opt
				return &gitlab.User{}, &gitlab.Response{}, nil
			}}}
			cr := serviceAccount(
				withExternalName("123"),
				withSpec(v1alpha1.ServiceAccountParameters{CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{Name: sPtr(testServiceAccountName), Avatar: tc.avatar}}),
				withAtProvider(v1alpha1.ServiceAccountObservation{CommonServiceAccountObservation: commonv1alpha1.CommonServiceAccountObservation{AvatarURL: tc.avatarURL}}),
			)

			_, err := e.Update(context.Background(), cr)
			if gotErr := err != nil; gotErr != tc.wantErr {
				t.Fatalf("Update() error = %v, want error %v", err, tc.wantErr)
			}
			if tc.wantErr {
				if !strings.HasPrefix(err.Error(), errUpdateAvatarFailed+": ") {
					t.Errorf("Update() error = %q, want prefix %q", err, errUpdateAvatarFailed)
				}
				if got != nil {
					t.Errorf("Update() called ModifyUser for an unsupported avatar")
				}
				return
			}
			gotFilename := ""
			if got.Avatar != nil {
				gotFilename = got.Avatar.Filename
			}
			if gotFilename != tc.wantFilename {
				t.Errorf("Update() uploaded avatar %q, want %q", gotFilename, tc.wantFilename)
			}
			if got.Name == nil || *got.Name != testServiceAccountName {
				t.Errorf("Update() dropped the name from ModifyUser options")
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type want struct {
		result managed.ExternalDelete
		err    error
	}

	desired := v1alpha1.ServiceAccountParameters{CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{Name: sPtr(testServiceAccountName), Username: sPtr(testServiceAccountUsername), Email: sPtr(testServiceAccountEmail)}}

	cases := map[string]struct {
		args args
		want want
	}{
		"InValidInput": {
			args: args{cr: unexpectedItem},
			want: want{err: errors.New(errNotServiceAccount)},
		},
		"NoExternalName": {
			args: args{client: &MockClient{MockDeleteUser: func(user int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
				return &gitlab.Response{}, nil
			}}, cr: serviceAccount(withSpec(desired))},
			want: want{result: managed.ExternalDelete{}, err: nil},
		},
		"NotIDExternalName": {
			args: args{client: &MockClient{MockDeleteUser: func(user int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
				return &gitlab.Response{}, nil
			}}, cr: serviceAccount(withExternalName("fr"), withSpec(desired))},
			want: want{err: errors.New(errIDNotInt)},
		},
		"ErrDelete": {
			args: args{client: &MockClient{MockDeleteUser: func(user int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
				return nil, errBoom
			}}, cr: serviceAccount(withExternalName("123"), withSpec(desired))},
			want: want{err: errors.Wrap(errBoom, errDeleteFailed)},
		},
		"Successful": {
			args: args{client: &MockClient{MockDeleteUser: func(user int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
				return &gitlab.Response{Response: &http.Response{StatusCode: 204}}, nil
			}}, cr: serviceAccount(withExternalName("123"), withSpec(desired))},
			want: want{result: managed.ExternalDelete{}, err: nil},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.args.kube, client: tc.args.client}
			got, err := e.Delete(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Delete(): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, got); diff != "" {
				t.Errorf("Delete(): -want result, +got result:\n%s", diff)
			}
		})
	}
}

func TestDisconnect(t *testing.T) {
	// Disconnect is currently a no-op required by the SDK.
	e := &external{}
	if err := e.Disconnect(context.Background()); err != nil {
		t.Errorf("Disconnect(): unexpected error: %v", err)
	}
}

func sPtr(s string) *string {
	return &s
}

func TestObserveAvatar(t *testing.T) {
	cases := map[string]struct {
		avatar    []byte
		avatarURL string
		want      managed.ExternalObservation
	}{
		"Unmanaged":     {avatarURL: testGravatarURL, want: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}},
		"AvatarMatches": {avatar: testAvatar, avatarURL: testAvatarURL, want: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}},
		"AvatarDiffers": {avatar: testAvatar, avatarURL: testGravatarURL, want: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: &MockClient{MockGetUser: func(_ int64, _ *gitlab.GetUserOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
				return &gitlab.User{ID: 123, Name: testServiceAccountName, AvatarURL: tc.avatarURL}, &gitlab.Response{}, nil
			}}}
			cr := serviceAccount(
				withExternalName("123"),
				withSpec(v1alpha1.ServiceAccountParameters{CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{Name: sPtr(testServiceAccountName), Avatar: tc.avatar}}),
			)

			got, err := e.Observe(context.Background(), cr)
			if err != nil {
				t.Fatalf("Observe() error = %v", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Observe(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestObserveAvatarSecretRef(t *testing.T) {
	cases := map[string]struct {
		secret  map[string][]byte
		want    managed.ExternalObservation
		wantErr string
	}{
		"SecretMatchesAvatar": {secret: map[string][]byte{testAvatarSecretKey: testAvatar}, want: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}},
		"SecretChanged":       {secret: map[string][]byte{testAvatarSecretKey: testSecretAvatar}, want: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}},
		"KeyMissing":          {secret: map[string][]byte{"logo": testAvatar}, wantErr: errAvatarKeyMissing},
		"SecretMissing":       {wantErr: errAvatarSecretGone},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: avatarKube(tc.secret), client: &MockClient{MockGetUser: func(_ int64, _ *gitlab.GetUserOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
				return &gitlab.User{ID: 123, Name: testServiceAccountName, AvatarURL: testAvatarURL}, &gitlab.Response{}, nil
			}}}
			cr := serviceAccount(
				withExternalName("123"),
				withSpec(v1alpha1.ServiceAccountParameters{CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{Name: sPtr(testServiceAccountName)}}),
				withAvatarSecretRef(),
			)

			got, err := e.Observe(context.Background(), cr)
			if gotErr := errString(err); gotErr != tc.wantErr {
				t.Errorf("Observe() error = %q, want %q", gotErr, tc.wantErr)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Observe(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdateAvatarSecretRef(t *testing.T) {
	cases := map[string]struct {
		secret       map[string][]byte
		wantFilename string
		wantErr      string
	}{
		"UploadsSecretAvatar": {secret: map[string][]byte{testAvatarSecretKey: testSecretAvatar}, wantFilename: testSecretAvatarName},
		"SecretMissing":       {wantErr: errAvatarSecretGone},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			gotFilename := ""
			e := &external{kube: avatarKube(tc.secret), client: &MockClient{MockModifyUser: func(_ int64, opt *gitlab.ModifyUserOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
				if opt.Avatar != nil {
					gotFilename = opt.Avatar.Filename
				}
				return &gitlab.User{}, &gitlab.Response{}, nil
			}}}
			cr := serviceAccount(
				withExternalName("123"),
				withSpec(v1alpha1.ServiceAccountParameters{CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{Name: sPtr(testServiceAccountName)}}),
				withAtProvider(v1alpha1.ServiceAccountObservation{CommonServiceAccountObservation: commonv1alpha1.CommonServiceAccountObservation{AvatarURL: testAvatarURL}}),
				withAvatarSecretRef(),
			)

			_, err := e.Update(context.Background(), cr)
			if gotErr := errString(err); gotErr != tc.wantErr {
				t.Errorf("Update() error = %q, want %q", gotErr, tc.wantErr)
			}
			if gotFilename != tc.wantFilename {
				t.Errorf("Update() uploaded avatar %q, want %q", gotFilename, tc.wantFilename)
			}
		})
	}
}
