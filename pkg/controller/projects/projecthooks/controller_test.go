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

package projecthooks

import (
	"context"
	"fmt"
	"testing"
	"time"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/pkg/errors"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects/fake"
)

var (
	errBoom       = errors.New("boom")
	createTime    = time.Now()
	projectID     = 5678
	projectHookID = 1234
)

type args struct {
	projecthook projects.ProjectHookClient
	kube        client.Client
	cr          *v1alpha1.ProjectHook
}

type projectHookModifier func(*v1alpha1.ProjectHook)

func withConditions(c ...runtimev1alpha1.Condition) projectHookModifier {
	return func(r *v1alpha1.ProjectHook) { r.Status.ConditionedStatus.Conditions = c }
}

func withDefaultValues() projectHookModifier {
	return func(ph *v1alpha1.ProjectHook) {
		f := false
		ph.Spec.ForProvider = v1alpha1.ProjectHookParameters{
			URL:                      nil,
			ConfidentialNoteEvents:   &f,
			ProjectID:                &projectID,
			PushEvents:               &f,
			PushEventsBranchFilter:   nil,
			IssuesEvents:             &f,
			ConfidentialIssuesEvents: &f,
			MergeRequestsEvents:      &f,
			TagPushEvents:            &f,
			NoteEvents:               &f,
			JobEvents:                &f,
			PipelineEvents:           &f,
			WikiPageEvents:           &f,
			EnableSSLVerification:    &f,
			Token:                    nil,
		}
	}
}

func withProjectID(pid int) projectHookModifier {
	return func(r *v1alpha1.ProjectHook) {
		r.Spec.ForProvider.ProjectID = &pid
	}
}

func withStatus(s v1alpha1.ProjectHookObservation) projectHookModifier {
	return func(r *v1alpha1.ProjectHook) { r.Status.AtProvider = s }
}

func withExternalName(projectHookID int) projectHookModifier {
	return func(r *v1alpha1.ProjectHook) { meta.SetExternalName(r, fmt.Sprint(projectHookID)) }
}

func projecthook(m ...projectHookModifier) *v1alpha1.ProjectHook {
	cr := &v1alpha1.ProjectHook{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.ProjectHook
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				projecthook: &fake.MockClient{
					MockGetProjectHook: func(pid interface{}, projectHookID int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error) {
						return &gitlab.ProjectHook{}, &gitlab.Response{}, nil
					},
				},
				cr: projecthook(
					withDefaultValues(),
					withExternalName(projectHookID),
					withStatus(v1alpha1.ProjectHookObservation{
						ID:        projectHookID,
						CreatedAt: &metav1.Time{Time: createTime},
					}),
				),
			},
			want: want{
				cr: projecthook(
					withDefaultValues(),
					withExternalName(projectHookID),
					withConditions(runtimev1alpha1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"NotUpToDate": {
			args: args{
				projecthook: &fake.MockClient{
					MockGetProjectHook: func(pid interface{}, projectHookID int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error) {
						return &gitlab.ProjectHook{
							MergeRequestsEvents: true,
						}, &gitlab.Response{}, nil
					},
				},
				cr: projecthook(
					withDefaultValues(),
					withExternalName(projectHookID),
					withStatus(v1alpha1.ProjectHookObservation{
						ID:        projectHookID,
						CreatedAt: &metav1.Time{Time: createTime},
					}),
				),
			},
			want: want{
				cr: projecthook(
					withDefaultValues(),
					withExternalName(projectHookID),
					withConditions(runtimev1alpha1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"LateInitSuccess": {
			args: args{
				projecthook: &fake.MockClient{
					MockGetProjectHook: func(pid interface{}, projectHookID int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error) {
						return &gitlab.ProjectHook{}, &gitlab.Response{}, nil
					},
				},
				cr: projecthook(
					withProjectID(projectID),
					withExternalName(projectHookID),
					withStatus(v1alpha1.ProjectHookObservation{
						ID:        projectHookID,
						CreatedAt: &metav1.Time{Time: createTime},
					}),
				),
			},
			want: want{
				cr: projecthook(
					withDefaultValues(),
					withExternalName(projectHookID),
					withConditions(runtimev1alpha1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.projecthook}
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
		cr     *v1alpha1.ProjectHook
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulCreation": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				projecthook: &fake.MockClient{
					MockAddProjectHook: func(pid interface{}, opt *gitlab.AddProjectHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error) {
						return &gitlab.ProjectHook{ID: projectHookID}, &gitlab.Response{}, nil
					},
				},
				cr: projecthook(
					withDefaultValues(),
				),
			},
			want: want{
				cr: projecthook(
					withDefaultValues(),
					withConditions(runtimev1alpha1.Creating()),
					withExternalName(projectHookID),
				),
				result: managed.ExternalCreation{},
			},
		},
		"FailedCreation": {
			args: args{
				projecthook: &fake.MockClient{
					MockAddProjectHook: func(pid interface{}, opt *gitlab.AddProjectHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error) {
						return &gitlab.ProjectHook{}, &gitlab.Response{}, errBoom
					},
				},
				cr: projecthook(
					withDefaultValues(),
				),
			},
			want: want{
				cr: projecthook(
					withDefaultValues(),
					withConditions(runtimev1alpha1.Creating()),
				),
				err: errors.Wrap(errBoom, errCreateFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.projecthook}
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
		cr     *v1alpha1.ProjectHook
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulEditProject": {
			args: args{
				projecthook: &fake.MockClient{
					MockEditProjectHook: func(pid interface{}, hook int, opt *gitlab.EditProjectHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error) {
						return &gitlab.ProjectHook{}, &gitlab.Response{}, nil
					},
				},
				cr: projecthook(
					withExternalName(projectHookID),
					withProjectID(projectID),
					withStatus(v1alpha1.ProjectHookObservation{ID: projectHookID}),
				),
			},
			want: want{
				cr: projecthook(
					withExternalName(projectHookID),
					withProjectID(projectID),
					withStatus(v1alpha1.ProjectHookObservation{ID: projectHookID}),
				),
			},
		},
		"FailedEdit": {
			args: args{
				projecthook: &fake.MockClient{
					MockEditProjectHook: func(pid interface{}, hook int, opt *gitlab.EditProjectHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error) {
						return &gitlab.ProjectHook{}, &gitlab.Response{}, errBoom
					},
				},
				cr: projecthook(
					withExternalName(projectHookID),
					withProjectID(projectID),
					withStatus(v1alpha1.ProjectHookObservation{ID: projectHookID}),
				),
			},
			want: want{
				cr: projecthook(
					withExternalName(projectHookID),
					withProjectID(projectID),
					withStatus(v1alpha1.ProjectHookObservation{ID: projectHookID}),
				),
				err: errors.Wrap(errBoom, errUpdateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.projecthook}
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
		cr  *v1alpha1.ProjectHook
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulDeletion": {
			args: args{
				projecthook: &fake.MockClient{
					MockDeleteProjectHook: func(pid interface{}, hook int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: projecthook(
					withProjectID(projectID),
					withStatus(v1alpha1.ProjectHookObservation{
						ID: projectHookID,
					}),
					withConditions(runtimev1alpha1.Available()),
				),
			},
			want: want{
				cr: projecthook(
					withProjectID(projectID),
					withStatus(v1alpha1.ProjectHookObservation{
						ID: projectHookID,
					}),
					withConditions(runtimev1alpha1.Deleting()),
				),
			},
		},
		"FailedDeletion": {
			args: args{
				projecthook: &fake.MockClient{
					MockDeleteProjectHook: func(pid interface{}, hook int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: projecthook(
					withProjectID(projectID),
					withStatus(v1alpha1.ProjectHookObservation{
						ID: projectHookID,
					}),
					withConditions(runtimev1alpha1.Available()),
				),
			},
			want: want{
				cr: projecthook(
					withProjectID(projectID),
					withStatus(v1alpha1.ProjectHookObservation{
						ID: projectHookID,
					}),
					withConditions(runtimev1alpha1.Deleting()),
				),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
		"InvalidProjectHookID": {
			args: args{
				projecthook: &fake.MockClient{
					MockDeleteProjectHook: func(pid interface{}, hook int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: projecthook(
					withProjectID(projectID),
					withConditions(runtimev1alpha1.Available()),
				),
			},
			want: want{
				cr: projecthook(
					withProjectID(projectID),
					withConditions(runtimev1alpha1.Deleting()),
				),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.projecthook}
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
