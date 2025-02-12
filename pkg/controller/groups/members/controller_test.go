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

package members

import (
	"context"
	"net/http"
	"testing"
	"time"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"gitlab.com/gitlab-org/api/client-go"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups/fake"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/users"
)

var (
	unexpecedItem resource.Managed
	errBoom       = errors.New("boom")
	ID            = 0
	username      = "username"
	userID        = 123
	name          = "name"
	state         = "state"
	avatarURL     = "http://avatarURL"
	webURL        = "http://webURL"
	accessLevel   = gitlab.AccessLevelValue(30)
	now           = time.Now()
	expiresAt     = gitlab.ISOTime(now.AddDate(0, 0, 7*3))
	expiresAtNew  = gitlab.ISOTime(now.AddDate(0, 0, 7*4))
	groupID       = 1234
)

type args struct {
	groupMember groups.MemberClient
	user        users.UserClient
	kube        client.Client
	cr          resource.Managed
}

func withAccessLevel(i int) groupModifier {
	return func(r *v1alpha1.Member) { r.Spec.ForProvider.AccessLevel = v1alpha1.AccessLevelValue(i) }
}

func withExpiresAt(s string) groupModifier {
	return func(r *v1alpha1.Member) { r.Spec.ForProvider.ExpiresAt = &s }
}

type groupModifier func(*v1alpha1.Member)

func withConditions(c ...xpv1.Condition) groupModifier {
	return func(cr *v1alpha1.Member) { cr.Status.ConditionedStatus.Conditions = c }
}

func withGroupID() groupModifier {
	return func(r *v1alpha1.Member) { r.Spec.ForProvider.GroupID = &groupID }
}

func withStatus(s v1alpha1.MemberObservation) groupModifier {
	return func(r *v1alpha1.Member) { r.Status.AtProvider = s }
}

func withSpec(s v1alpha1.MemberParameters) groupModifier {
	return func(r *v1alpha1.Member) { r.Spec.ForProvider = s }
}

func groupMember(m ...groupModifier) *v1alpha1.Member {
	cr := &v1alpha1.Member{}
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
				err: errors.New(errNotMember),
			},
		},
		"ProviderConfigRefNotGivenError": {
			args: args{
				cr:   groupMember(),
				kube: &test.MockClient{MockGet: test.NewMockGetFn(nil)},
			},
			want: want{
				cr:  groupMember(),
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
				err: errors.New(errNotMember),
			},
		},
		"FailedGetRequest": {
			args: args{
				groupMember: &fake.MockClient{
					MockGetMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 400}}, errBoom
					},
				},
				cr: groupMember(
					withGroupID(),
					withSpec(v1alpha1.MemberParameters{UserID: &userID, GroupID: &groupID})),
			},
			want: want{
				cr: groupMember(
					withGroupID(),
					withSpec(v1alpha1.MemberParameters{UserID: &userID, GroupID: &groupID})),
				result: managed.ExternalObservation{ResourceExists: false},
				err:    errors.Wrap(errBoom, errGetFailed),
			},
		},
		"ErrGet404": {
			args: args{
				groupMember: &fake.MockClient{
					MockGetMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errBoom
					},
				},
				cr: groupMember(
					withGroupID(),
					withSpec(v1alpha1.MemberParameters{UserID: &userID, GroupID: &groupID})),
			},
			want: want{
				cr: groupMember(
					withGroupID(),
					withSpec(v1alpha1.MemberParameters{UserID: &userID, GroupID: &groupID})),
				result: managed.ExternalObservation{ResourceExists: false},
				err:    nil,
			},
		},
		"SuccessfulAvailable": {
			args: args{
				groupMember: &fake.MockClient{
					MockGetMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
						return &gitlab.GroupMember{}, &gitlab.Response{}, nil
					},
				},
				cr: groupMember(
					withGroupID(),
					withSpec(v1alpha1.MemberParameters{UserID: &userID, GroupID: &groupID}),
				),
			},
			want: want{
				cr: groupMember(
					withConditions(xpv1.Available()),
					withGroupID(),
					withSpec(v1alpha1.MemberParameters{UserID: &userID, GroupID: &groupID}),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: false,
				},
			},
		},
		"IsGroupUpToDateAccessLevel": {
			args: args{
				groupMember: &fake.MockClient{
					MockGetMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
						return &gitlab.GroupMember{
							AccessLevel: accessLevel,
						}, &gitlab.Response{}, nil
					},
				},
				cr: groupMember(
					withGroupID(),
					withSpec(v1alpha1.MemberParameters{UserID: &userID, GroupID: &groupID}),
					withAccessLevel(10),
				),
			},
			want: want{
				cr: groupMember(
					withConditions(xpv1.Available()),
					withGroupID(),
					withSpec(v1alpha1.MemberParameters{UserID: &userID, GroupID: &groupID}),
					withAccessLevel(10),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
			},
		},
		"IsGroupUpToDateExpiresAt": {
			args: args{
				groupMember: &fake.MockClient{
					MockGetMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
						return &gitlab.GroupMember{ExpiresAt: &expiresAt}, &gitlab.Response{}, nil
					},
				},
				cr: groupMember(
					withGroupID(),
					withSpec(v1alpha1.MemberParameters{UserID: &userID, GroupID: &groupID}),
					withExpiresAt(expiresAtNew.String()),
				),
			},
			want: want{
				cr: groupMember(
					withConditions(xpv1.Available()),
					withGroupID(),
					withSpec(v1alpha1.MemberParameters{UserID: &userID, GroupID: &groupID}),
					withExpiresAt(expiresAtNew.String()),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
			},
		},
		"NoUserIDandNoUserName": {
			args: args{
				groupMember: &fake.MockClient{
					MockGetMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
						return nil, nil, errBoom
					},
				},
				cr: groupMember(
					withSpec(v1alpha1.MemberParameters{
						GroupID:  &groupID,
						UserID:   nil,
						UserName: nil,
					}))},
			want: want{
				cr: groupMember(
					withSpec(v1alpha1.MemberParameters{
						GroupID:  &groupID,
						UserID:   nil,
						UserName: nil,
					})),
				result: managed.ExternalObservation{},
				err:    errors.New(errMissingUserInfo),
			},
		},
		"NoUserIDSuccess": {
			args: args{
				groupMember: &fake.MockClient{
					MockGetMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
						return &gitlab.GroupMember{}, &gitlab.Response{}, nil
					},
				},
				user: &fake.MockClient{
					MockListUsers: func(opt *gitlab.ListUsersOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.User, *gitlab.Response, error) {
						return []*gitlab.User{{ID: userID}}, &gitlab.Response{}, nil
					},
				},
				cr: groupMember(
					withGroupID(),
					withSpec(v1alpha1.MemberParameters{
						UserName: &username,
						GroupID:  &groupID,
					})),
			},
			want: want{
				cr: groupMember(
					withGroupID(),
					withConditions(xpv1.Available()),
					withStatus(v1alpha1.MemberObservation{}),
					withSpec(v1alpha1.MemberParameters{
						UserName: &username,
						UserID:   &userID,
						GroupID:  &groupID,
					})),
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
			e := &external{kube: tc.kube, client: tc.groupMember, userClient: tc.user}
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
				err: errors.New(errNotMember),
			},
		},
		"SuccessfulCreationWithoutExpiresAt": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				groupMember: &fake.MockClient{
					MockAddMember: func(gid interface{}, opt *gitlab.AddGroupMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
						return &gitlab.GroupMember{
							ID:          ID,
							Username:    username,
							Name:        name,
							State:       state,
							AvatarURL:   avatarURL,
							WebURL:      webURL,
							AccessLevel: accessLevel,
						}, &gitlab.Response{}, nil
					},
				},
				cr: groupMember(
					withSpec(v1alpha1.MemberParameters{GroupID: &ID}),
				),
			},
			want: want{
				cr: groupMember(
					withGroupID(),
					withSpec(v1alpha1.MemberParameters{GroupID: &ID}),
				),
				result: managed.ExternalCreation{},
			},
		},
		"SuccessfulCreationWithExpiresAt": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				groupMember: &fake.MockClient{
					MockAddMember: func(gid interface{}, opt *gitlab.AddGroupMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
						return &gitlab.GroupMember{
							ID:          ID,
							Username:    username,
							Name:        name,
							State:       state,
							AvatarURL:   avatarURL,
							WebURL:      webURL,
							AccessLevel: accessLevel,
							ExpiresAt:   &expiresAt,
						}, &gitlab.Response{}, nil
					},
				},
				cr: groupMember(
					withSpec(v1alpha1.MemberParameters{GroupID: &ID}),
				),
			},
			want: want{
				cr: groupMember(
					withGroupID(),
					withSpec(v1alpha1.MemberParameters{GroupID: &ID}),
				),
				result: managed.ExternalCreation{},
			},
		},
		"FailedCreation": {
			args: args{
				groupMember: &fake.MockClient{
					MockAddMember: func(gid interface{}, opt *gitlab.AddGroupMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
						return &gitlab.GroupMember{}, &gitlab.Response{}, errBoom
					},
				},
				cr: groupMember(
					withSpec(v1alpha1.MemberParameters{GroupID: &ID}),
				),
			},
			want: want{
				cr: groupMember(
					withGroupID(),
					withSpec(v1alpha1.MemberParameters{GroupID: &ID}),
				),
				err: errors.Wrap(errBoom, errCreateFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.groupMember}
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
				err: errors.New(errNotMember),
			},
		},
		"SuccessfulUpdate": {
			args: args{
				groupMember: &fake.MockClient{
					MockEditMember: func(gid interface{}, user int, opt *gitlab.EditGroupMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
						return &gitlab.GroupMember{
							ID:          ID,
							Username:    username,
							Name:        name,
							State:       state,
							AvatarURL:   avatarURL,
							WebURL:      webURL,
							AccessLevel: accessLevel,
						}, &gitlab.Response{}, nil
					},
				},
				cr: groupMember(
					withGroupID(),
					withSpec(v1alpha1.MemberParameters{UserID: &userID, GroupID: &groupID}),
					withStatus(v1alpha1.MemberObservation{Username: "new username"}),
				),
			},
			want: want{
				cr: groupMember(
					withGroupID(),
					withSpec(v1alpha1.MemberParameters{UserID: &userID, GroupID: &groupID}),
					withStatus(v1alpha1.MemberObservation{Username: "new username"}),
				),
			},
		},
		"FailedUpdate": {
			args: args{
				groupMember: &fake.MockClient{
					MockEditMember: func(gid interface{}, user int, opt *gitlab.EditGroupMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
						return &gitlab.GroupMember{}, &gitlab.Response{}, errBoom
					},
				},
				cr: groupMember(
					withGroupID(),
					withSpec(v1alpha1.MemberParameters{UserID: &userID, GroupID: &groupID})),
			},
			want: want{
				cr: groupMember(
					withGroupID(),
					withSpec(v1alpha1.MemberParameters{UserID: &userID, GroupID: &groupID})),
				err: errors.Wrap(errBoom, errUpdateFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.groupMember}
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
				err: errors.New(errNotMember),
			},
		},
		"SuccessfulDeletion": {
			args: args{
				groupMember: &fake.MockClient{
					MockRemoveMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: groupMember(
					withGroupID(),
					withSpec(v1alpha1.MemberParameters{UserID: &userID, GroupID: &groupID})),
			},
			want: want{
				cr: groupMember(
					withGroupID(),
					withSpec(v1alpha1.MemberParameters{UserID: &userID, GroupID: &groupID})),
				err: nil,
			},
		},
		"FailedDeletion": {
			args: args{
				groupMember: &fake.MockClient{
					MockRemoveMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: groupMember(withGroupID()),
			},
			want: want{
				cr:  groupMember(withGroupID()),
				err: errors.New(errMissingUserInfo),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.groupMember}
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
