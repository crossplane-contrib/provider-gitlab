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
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	commonv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/instance/v1alpha1"
	groupsfake "github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/groups/fake"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/instance"
)

var (
	unexpectedItem resource.Managed
	errBoom        = errors.New("boom")
)

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
				client: &MockClient{MockGetUser: func(user int64, opt gitlab.GetUsersOptions, options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
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
					withConditions(xpv1.Available()),
					func(r *v1alpha1.ServiceAccount) {
						r.Status.AtProvider = instance.GenerateServiceAccountObservation(&gitlab.User{ID: 123, Name: testServiceAccountName, Username: testServiceAccountUsername, Email: testServiceAccountEmail})
						r.Status.AtProvider.MissingMemberShipGroups = []int64{1, 3}
						r.Status.AtProvider.WrongPermissionsGroups = []int64{2, 4}
					},
				),
				result: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false, ResourceLateInitialized: false},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.args.kube, client: tc.args.client}
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
