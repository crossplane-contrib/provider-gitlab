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

package projectsharegroups

import (
	"context"
	"net/http"
	"strconv"
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects/fake"
)

var (
	unexpecedItem resource.Managed
	errBoom       = errors.New("boom")
	projectIDStr  = "0"
	groupIDStr    = "123"
)

type args struct {
	project projects.Client
	cr      resource.Managed
}

type projectShareGroupModifier func(*v1alpha1.ProjectShareGroup)

func withConditions(c ...xpv1.Condition) projectShareGroupModifier {
	return func(cr *v1alpha1.ProjectShareGroup) { cr.Status.ConditionedStatus.Conditions = c }
}

func withProjectID(pID *string) projectShareGroupModifier {
	return func(r *v1alpha1.ProjectShareGroup) { r.Spec.ForProvider.ProjectID = pID }
}

func withGroupID(gID *string) projectShareGroupModifier {
	return func(r *v1alpha1.ProjectShareGroup) { r.Spec.ForProvider.GroupID = gID }
}

func withAccessLevel(a int) projectShareGroupModifier {
	return func(r *v1alpha1.ProjectShareGroup) { r.Spec.ForProvider.AccessLevel = a }
}

func projectShareGroup(m ...projectShareGroupModifier) *v1alpha1.ProjectShareGroup {
	cr := &v1alpha1.ProjectShareGroup{}
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
				err: errors.New(errNotProjectShareGroup),
			},
		},
		"ProjectIDMissing": {
			args: args{
				cr: projectShareGroup(withGroupID(&groupIDStr)),
			},
			want: want{
				cr:     projectShareGroup(withGroupID(&groupIDStr)),
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"GroupIDMissing": {
			args: args{
				cr: projectShareGroup(withProjectID(&projectIDStr)),
			},
			want: want{
				cr:     projectShareGroup(withProjectID(&projectIDStr)),
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"GetProjectError": {
			args: args{
				project: &fake.MockClient{
					MockGetProject: func(pid any, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return nil, nil, errBoom
					},
				},
				cr: projectShareGroup(withProjectID(&projectIDStr), withGroupID(&groupIDStr)),
			},
			want: want{
				cr:     projectShareGroup(withProjectID(&projectIDStr), withGroupID(&groupIDStr)),
				result: managed.ExternalObservation{ResourceExists: false},
				err:    errors.Wrap(errBoom, errGetProject),
			},
		},
		"ShareNotFound": {
			args: args{
				project: &fake.MockClient{
					MockGetProject: func(pid any, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{
							SharedWithGroups: []gitlab.ProjectSharedWithGroup{},
						}, &gitlab.Response{}, nil
					},
				},
				cr: projectShareGroup(withProjectID(&projectIDStr), withGroupID(&groupIDStr)),
			},
			want: want{
				cr:     projectShareGroup(withProjectID(&projectIDStr), withGroupID(&groupIDStr)),
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"SuccessfulAvailable": {
			args: args{
				project: &fake.MockClient{
					MockGetProject: func(pid any, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						gID, _ := strconv.ParseInt(groupIDStr, 10, 64)
						return &gitlab.Project{
							SharedWithGroups: []gitlab.ProjectSharedWithGroup{
								{
									GroupID:          gID,
									GroupAccessLevel: int64(gitlab.AccessLevelValue(30)),
								},
							},
						}, &gitlab.Response{}, nil
					},
				},
				cr: projectShareGroup(withProjectID(&projectIDStr), withGroupID(&groupIDStr), withAccessLevel(30)),
			},
			want: want{
				cr: projectShareGroup(
					withProjectID(&projectIDStr),
					withGroupID(&groupIDStr),
					withAccessLevel(30),
					withConditions(xpv1.Available()),
					func(cr *v1alpha1.ProjectShareGroup) { cr.Status.AtProvider.ID = "0-123" },
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"NotUpToDate": {
			args: args{
				project: &fake.MockClient{
					MockGetProject: func(pid any, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						gID, _ := strconv.ParseInt(groupIDStr, 10, 64)
						return &gitlab.Project{
							SharedWithGroups: []gitlab.ProjectSharedWithGroup{
								{
									GroupID:          gID,
									GroupAccessLevel: int64(gitlab.AccessLevelValue(40)), // Maintainer
								},
							},
						}, &gitlab.Response{}, nil
					},
				},
				cr: projectShareGroup(withProjectID(&projectIDStr), withGroupID(&groupIDStr), withAccessLevel(30)), // Developer
			},
			want: want{
				cr: projectShareGroup(
					withProjectID(&projectIDStr),
					withGroupID(&groupIDStr),
					withAccessLevel(30),
					withConditions(xpv1.Available()),
					func(cr *v1alpha1.ProjectShareGroup) { cr.Status.AtProvider.ID = "0-123" },
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.args.project}
			obs, err := e.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, obs); diff != "" {
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
				err: errors.New(errNotProjectShareGroup),
			},
		},
		"SuccessfulCreation": {
			args: args{
				project: &fake.MockClient{
					MockShareProjectWithGroup: func(pid any, opt *gitlab.ShareWithGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: projectShareGroup(withProjectID(&projectIDStr), withGroupID(&groupIDStr), withAccessLevel(30)),
			},
			want: want{
				cr:     projectShareGroup(withProjectID(&projectIDStr), withGroupID(&groupIDStr), withAccessLevel(30)),
				result: managed.ExternalCreation{},
			},
		},
		"FailedCreation": {
			args: args{
				project: &fake.MockClient{
					MockShareProjectWithGroup: func(pid any, opt *gitlab.ShareWithGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return nil, errBoom
					},
				},
				cr: projectShareGroup(withProjectID(&projectIDStr), withGroupID(&groupIDStr)),
			},
			want: want{
				cr:  projectShareGroup(withProjectID(&projectIDStr), withGroupID(&groupIDStr)),
				err: errors.Wrap(errBoom, errShareProject),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.args.project}
			_, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type want struct {
		cr  resource.Managed
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulUpdate": {
			args: args{
				project: &fake.MockClient{
					MockDeleteSharedProjectFromGroup: func(pid any, groupID int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: projectShareGroup(withProjectID(&projectIDStr), withGroupID(&groupIDStr)),
			},
			want: want{
				cr: projectShareGroup(withProjectID(&projectIDStr), withGroupID(&groupIDStr)),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.args.project}
			_, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
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
				err: errors.New(errNotProjectShareGroup),
			},
		},
		"SuccessfulDeletion": {
			args: args{
				project: &fake.MockClient{
					MockDeleteSharedProjectFromGroup: func(pid any, groupID int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: projectShareGroup(withProjectID(&projectIDStr), withGroupID(&groupIDStr)),
			},
			want: want{
				cr: projectShareGroup(withProjectID(&projectIDStr), withGroupID(&groupIDStr)),
			},
		},
		"FailedDeletion": {
			args: args{
				project: &fake.MockClient{
					MockDeleteSharedProjectFromGroup: func(pid any, groupID int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return nil, errBoom
					},
				},
				cr: projectShareGroup(withProjectID(&projectIDStr), withGroupID(&groupIDStr)),
			},
			want: want{
				cr:  projectShareGroup(withProjectID(&projectIDStr), withGroupID(&groupIDStr)),
				err: errors.Wrap(errBoom, errUnshareProject),
			},
		},
		"FailedDeletion404": {
			args: args{
				project: &fake.MockClient{
					MockDeleteSharedProjectFromGroup: func(pid any, groupID int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errBoom
					},
				},
				cr: projectShareGroup(withProjectID(&projectIDStr), withGroupID(&groupIDStr)),
			},
			want: want{
				cr:  projectShareGroup(withProjectID(&projectIDStr), withGroupID(&groupIDStr)),
				err: errors.Wrap(errBoom, errUnshareProject),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.args.project}
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
