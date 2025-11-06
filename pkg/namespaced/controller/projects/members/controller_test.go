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

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects/fake"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/users"
)

var (
	unexpecedItem resource.Managed
	errBoom       = errors.New("boom")
	projectID     = 0
	username      = "username"
	userID        = 123
	name          = "name"
	state         = "state"
	avatarURL     = "http://avatarURL"
	webURL        = "http://webURL"
	email         = "email@gmail.com"
	accessLevel   = gitlab.AccessLevelValue(30)
	now           = time.Now()
	expiresAt     = gitlab.ISOTime(now.AddDate(0, 0, 7*3))
	expiresAtNew  = gitlab.ISOTime(now.AddDate(0, 0, 7*4))
)

type args struct {
	projectMember projects.MemberClient
	user          users.UserClient
	kube          client.Client
	cr            resource.Managed
}

func withAccessLevel(i int) projectModifier {
	return func(r *v1alpha1.Member) { r.Spec.ForProvider.AccessLevel = v1alpha1.AccessLevelValue(i) }
}

func withExpiresAt(s string) projectModifier {
	return func(r *v1alpha1.Member) { r.Spec.ForProvider.ExpiresAt = &s }
}

type projectModifier func(*v1alpha1.Member)

func withConditions(c ...xpv1.Condition) projectModifier {
	return func(cr *v1alpha1.Member) { cr.Status.ConditionedStatus.Conditions = c }
}

func withProjectID() projectModifier {
	return func(r *v1alpha1.Member) { r.Spec.ForProvider.ProjectID = &projectID }
}

func withStatus(s v1alpha1.MemberObservation) projectModifier {
	return func(r *v1alpha1.Member) { r.Status.AtProvider = s }
}

func withSpec(s v1alpha1.MemberParameters) projectModifier {
	return func(r *v1alpha1.Member) { r.Spec.ForProvider = s }
}

func projectMember(m ...projectModifier) *v1alpha1.Member {
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
				cr:   projectMember(),
				kube: &test.MockClient{MockGet: test.NewMockGetFn(nil)},
			},
			want: want{
				cr:  projectMember(),
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
		"ErrProjectIDMissing": {
			args: args{
				cr: projectMember(),
			},
			want: want{
				cr: projectMember(),
				result: managed.ExternalObservation{
					ResourceExists:          false,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
				err: errors.New(errProjectIDMissing),
			},
		},
		"ErrGet404": {
			args: args{
				projectMember: &fake.MockClient{
					MockGetMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errBoom
					},
				},
				cr: projectMember(
					withProjectID(),
					withSpec(v1alpha1.MemberParameters{
						UserID:    &userID,
						ProjectID: &projectID,
					})),
			},
			want: want{
				cr: projectMember(
					withProjectID(),
					withSpec(v1alpha1.MemberParameters{
						UserID:    &userID,
						ProjectID: &projectID,
					})),
				result: managed.ExternalObservation{ResourceExists: false},
				err:    nil,
			},
		},
		"NoUserIDandNoUserName": {
			args: args{
				projectMember: &fake.MockClient{
					MockGetMember: func(pid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
						return nil, nil, errBoom
					},
				},
				cr: projectMember(
					withProjectID(),
					withSpec(v1alpha1.MemberParameters{
						UserID:   nil,
						UserName: nil,
					})),
			},
			want: want{
				cr: projectMember(
					withProjectID(),
					withSpec(v1alpha1.MemberParameters{
						UserID:   nil,
						UserName: nil,
					})),
				result: managed.ExternalObservation{},
				err:    errors.New(errProjectIDMissing),
			},
		},
		"ErrGet": {
			args: args{
				projectMember: &fake.MockClient{
					MockGetMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
						return nil, nil, errBoom
					},
				},
				cr: projectMember(withProjectID()),
			},
			want: want{
				cr:     projectMember(withProjectID()),
				result: managed.ExternalObservation{ResourceExists: false},
				err:    errors.New(errUserInfoMissing),
			},
		},
		"SuccessfulAvailable": {
			args: args{
				projectMember: &fake.MockClient{
					MockGetMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
						return &gitlab.ProjectMember{}, &gitlab.Response{}, nil
					},
				},
				cr: projectMember(
					withProjectID(),
					withSpec(v1alpha1.MemberParameters{
						UserID:    &userID,
						ProjectID: &projectID,
					}),
				),
			},
			want: want{
				cr: projectMember(
					withConditions(xpv1.Available()),
					withProjectID(),
					withSpec(v1alpha1.MemberParameters{
						UserID:    &userID,
						ProjectID: &projectID,
					}),
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
				projectMember: &fake.MockClient{
					MockGetMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
						return &gitlab.ProjectMember{
							AccessLevel: accessLevel,
						}, &gitlab.Response{}, nil
					},
				},
				cr: projectMember(
					withProjectID(),
					withSpec(v1alpha1.MemberParameters{
						UserID:    &userID,
						ProjectID: &projectID,
					}),
					withAccessLevel(10),
				),
			},
			want: want{
				cr: projectMember(
					withConditions(xpv1.Available()),
					withSpec(v1alpha1.MemberParameters{
						UserID:    &userID,
						ProjectID: &projectID,
					}),
					withProjectID(),
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
				projectMember: &fake.MockClient{
					MockGetMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
						return &gitlab.ProjectMember{ExpiresAt: &expiresAt}, &gitlab.Response{}, nil
					},
				},
				cr: projectMember(
					withProjectID(),
					withExpiresAt(expiresAtNew.String()),
					withSpec(v1alpha1.MemberParameters{
						UserID:    &userID,
						ProjectID: &projectID,
					}),
				),
			},
			want: want{
				cr: projectMember(
					withConditions(xpv1.Available()),
					withExpiresAt(expiresAtNew.String()),
					withSpec(v1alpha1.MemberParameters{
						UserID:    &userID,
						ProjectID: &projectID,
					}),
					withProjectID(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
			},
		},
		"NoUserIDSuccess": {
			args: args{
				projectMember: &fake.MockClient{
					MockGetMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
						return &gitlab.ProjectMember{}, &gitlab.Response{}, nil
					},
				},
				user: &fake.MockClient{
					MockListUsers: func(opt *gitlab.ListUsersOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.User, *gitlab.Response, error) {
						return []*gitlab.User{{ID: userID}}, &gitlab.Response{}, nil
					},
				},
				cr: projectMember(
					withProjectID(),
					withSpec(v1alpha1.MemberParameters{
						UserName:  &username,
						ProjectID: &projectID,
					})),
			},
			want: want{
				cr: projectMember(
					withProjectID(),
					withConditions(xpv1.Available()),
					withSpec(v1alpha1.MemberParameters{
						UserName:  &username,
						UserID:    &userID,
						ProjectID: &projectID,
					}),
					withStatus(v1alpha1.MemberObservation{}),
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
			e := &external{kube: tc.kube, client: tc.projectMember, userClient: tc.user}
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
				projectMember: &fake.MockClient{
					MockAddMember: func(gid interface{}, opt *gitlab.AddProjectMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
						return &gitlab.ProjectMember{
							ID:          projectID,
							Username:    username,
							Email:       email,
							Name:        name,
							State:       state,
							AvatarURL:   avatarURL,
							WebURL:      webURL,
							AccessLevel: accessLevel,
						}, &gitlab.Response{}, nil
					},
				},
				cr: projectMember(
					withSpec(v1alpha1.MemberParameters{ProjectID: &projectID}),
				),
			},
			want: want{
				cr: projectMember(
					withProjectID(),
					withSpec(v1alpha1.MemberParameters{ProjectID: &projectID}),
				),
				result: managed.ExternalCreation{},
			},
		},
		"SuccessfulCreationWithExpiresAt": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				projectMember: &fake.MockClient{
					MockAddMember: func(gid interface{}, opt *gitlab.AddProjectMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
						return &gitlab.ProjectMember{
							ID:          projectID,
							Username:    username,
							Email:       email,
							Name:        name,
							State:       state,
							AvatarURL:   avatarURL,
							WebURL:      webURL,
							AccessLevel: accessLevel,
							ExpiresAt:   &expiresAt,
						}, &gitlab.Response{}, nil
					},
				},
				cr: projectMember(
					withSpec(v1alpha1.MemberParameters{ProjectID: &projectID}),
				),
			},
			want: want{
				cr: projectMember(
					withProjectID(),
					withSpec(v1alpha1.MemberParameters{ProjectID: &projectID}),
				),
				result: managed.ExternalCreation{},
			},
		},
		"FailedCreation": {
			args: args{
				projectMember: &fake.MockClient{
					MockAddMember: func(gid interface{}, opt *gitlab.AddProjectMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
						return &gitlab.ProjectMember{}, &gitlab.Response{}, errBoom
					},
				},
				cr: projectMember(
					withSpec(v1alpha1.MemberParameters{ProjectID: &projectID}),
				),
			},
			want: want{
				cr: projectMember(
					withProjectID(),
					withSpec(v1alpha1.MemberParameters{ProjectID: &projectID}),
				),
				err: errors.Wrap(errBoom, errCreateFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.projectMember}
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
				projectMember: &fake.MockClient{
					MockEditMember: func(gid interface{}, user int, opt *gitlab.EditProjectMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
						return &gitlab.ProjectMember{
							ID:          projectID,
							Username:    username,
							Email:       email,
							Name:        name,
							State:       state,
							AvatarURL:   avatarURL,
							WebURL:      webURL,
							AccessLevel: accessLevel,
						}, &gitlab.Response{}, nil
					},
				},
				cr: projectMember(
					withProjectID(),
					withStatus(v1alpha1.MemberObservation{Username: "new username"}),
					withSpec(v1alpha1.MemberParameters{
						UserID:    &userID,
						ProjectID: &projectID,
					}),
				),
			},
			want: want{
				cr: projectMember(
					withProjectID(),
					withStatus(v1alpha1.MemberObservation{Username: "new username"}),
					withSpec(v1alpha1.MemberParameters{
						UserID:    &userID,
						ProjectID: &projectID,
					}),
				),
			},
		},
		"FailedUpdate": {
			args: args{
				projectMember: &fake.MockClient{
					MockEditMember: func(gid interface{}, user int, opt *gitlab.EditProjectMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
						return &gitlab.ProjectMember{}, &gitlab.Response{}, errBoom
					},
				},
				cr: projectMember(withProjectID()),
			},
			want: want{
				cr:  projectMember(withProjectID()),
				err: errors.New(errUserInfoMissing),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.projectMember}
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
				projectMember: &fake.MockClient{
					MockDeleteMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: projectMember(
					withProjectID(),
					withSpec(v1alpha1.MemberParameters{
						UserID:    &userID,
						ProjectID: &projectID,
					})),
			},
			want: want{
				cr: projectMember(
					withProjectID(),
					withSpec(v1alpha1.MemberParameters{
						UserID:    &userID,
						ProjectID: &projectID,
					})),
				err: nil,
			},
		},
		"FailedDeletion": {
			args: args{
				projectMember: &fake.MockClient{
					MockDeleteMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: projectMember(
					withSpec(v1alpha1.MemberParameters{ProjectID: &projectID})),
			},
			want: want{
				cr: projectMember(
					withSpec(v1alpha1.MemberParameters{ProjectID: &projectID})),
				err: errors.New(errUserInfoMissing),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.projectMember}
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
