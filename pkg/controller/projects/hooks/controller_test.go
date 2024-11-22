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

package hooks

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
	corev1 "k8s.io/api/core/v1"
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
	tokenValue    = "test"
	tokenSecret   = corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "test"},
		Data: map[string][]byte{
			"token": []byte(tokenValue),
		},
	}
)

type args struct {
	projecthook projects.HookClient
	kube        client.Client
	cr          *v1alpha1.Hook
}

type projectHookModifier func(*v1alpha1.Hook)

func withConditions(c ...xpv1.Condition) projectHookModifier {
	return func(r *v1alpha1.Hook) { r.Status.ConditionedStatus.Conditions = c }
}

func withDefaultValues() projectHookModifier {
	return func(ph *v1alpha1.Hook) {
		f := false
		ph.Spec.ForProvider = v1alpha1.HookParameters{
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
			Token: &v1alpha1.Token{
				SecretRef: &v1.SecretKeySelector{
					Key: "token", SecretReference: v1.SecretReference{Name: "test", Namespace: "test"},
				},
			},
		}
	}
}

func withProjectID(pid int) projectHookModifier {
	return func(r *v1alpha1.Hook) {
		r.Spec.ForProvider.ProjectID = &pid
	}
}

func withTokenRef() projectHookModifier {
	return func(r *v1alpha1.Hook) {
		r.Spec.ForProvider.Token = &v1alpha1.Token{
			SecretRef: &v1.SecretKeySelector{
				Key: "token", SecretReference: v1.SecretReference{Name: "test", Namespace: "test"},
			},
		}
	}
}

func withStatus(s v1alpha1.HookObservation) projectHookModifier {
	return func(r *v1alpha1.Hook) { r.Status.AtProvider = s }
}

func withExternalName(projectHookID int) projectHookModifier {
	return func(r *v1alpha1.Hook) { meta.SetExternalName(r, fmt.Sprint(projectHookID)) }
}

func projecthook(m ...projectHookModifier) *v1alpha1.Hook {
	cr := &v1alpha1.Hook{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.Hook
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
					MockGetHook: func(pid interface{}, projectHookID int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error) {
						return &gitlab.ProjectHook{}, &gitlab.Response{}, nil
					},
				},
				cr: projecthook(
					withDefaultValues(),
					withExternalName(projectHookID),
					withStatus(v1alpha1.HookObservation{
						ID:        projectHookID,
						CreatedAt: &metav1.Time{Time: createTime},
					}),
				),
			},
			want: want{
				cr: projecthook(
					withDefaultValues(),
					withExternalName(projectHookID),
					withConditions(xpv1.Available()),
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
					MockGetHook: func(pid interface{}, projectHookID int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error) {
						return &gitlab.ProjectHook{
							MergeRequestsEvents: true,
						}, &gitlab.Response{}, nil
					},
				},
				cr: projecthook(
					withDefaultValues(),
					withExternalName(projectHookID),
					withStatus(v1alpha1.HookObservation{
						ID:        projectHookID,
						CreatedAt: &metav1.Time{Time: createTime},
					}),
				),
			},
			want: want{
				cr: projecthook(
					withDefaultValues(),
					withExternalName(projectHookID),
					withConditions(xpv1.Available()),
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
					MockGetHook: func(pid interface{}, projectHookID int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error) {
						return &gitlab.ProjectHook{}, &gitlab.Response{}, nil
					},
				},
				cr: projecthook(
					withProjectID(projectID),
					withTokenRef(),
					withExternalName(projectHookID),
					withStatus(v1alpha1.HookObservation{
						ID:        projectHookID,
						CreatedAt: &metav1.Time{Time: createTime},
					}),
				),
			},
			want: want{
				cr: projecthook(
					withDefaultValues(),
					withExternalName(projectHookID),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
				},
			},
		},
		"ErrGet404": {
			args: args{
				projecthook: &fake.MockClient{
					MockGetHook: func(pid interface{}, hook int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errBoom
					},
				},
				cr: projecthook(
					withProjectID(projectID),
					withExternalName(projectHookID),
				),
			},
			want: want{
				cr: projecthook(
					withProjectID(projectID),
					withExternalName(projectHookID),
				),
				result: managed.ExternalObservation{},
				err:    nil,
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
		cr     *v1alpha1.Hook
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
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						*obj.(*corev1.Secret) = tokenSecret
						return nil
					}),
				},
				projecthook: &fake.MockClient{
					MockAddHook: func(pid interface{}, opt *gitlab.AddProjectHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error) {
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
					withConditions(xpv1.Creating()),
					withExternalName(projectHookID),
				),
				result: managed.ExternalCreation{},
			},
		},
		"FailedCreation": {
			args: args{
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						*obj.(*corev1.Secret) = tokenSecret
						return nil
					}),
				},
				projecthook: &fake.MockClient{
					MockAddHook: func(pid interface{}, opt *gitlab.AddProjectHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error) {
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
					withConditions(xpv1.Creating()),
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
		cr     *v1alpha1.Hook
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulEditHook": {
			args: args{
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						*obj.(*corev1.Secret) = tokenSecret
						return nil
					}),
				},
				projecthook: &fake.MockClient{
					MockEditHook: func(pid interface{}, hook int, opt *gitlab.EditProjectHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error) {
						return &gitlab.ProjectHook{}, &gitlab.Response{}, nil
					},
				},
				cr: projecthook(
					withExternalName(projectHookID),
					withProjectID(projectID),
					withTokenRef(),
					withStatus(v1alpha1.HookObservation{ID: projectHookID}),
				),
			},
			want: want{
				cr: projecthook(
					withExternalName(projectHookID),
					withTokenRef(),
					withProjectID(projectID),
					withStatus(v1alpha1.HookObservation{ID: projectHookID}),
				),
			},
		},
		"FailedEdit": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						*obj.(*corev1.Secret) = tokenSecret
						return nil
					}),
				},
				projecthook: &fake.MockClient{
					MockEditHook: func(pid interface{}, hook int, opt *gitlab.EditProjectHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error) {
						return &gitlab.ProjectHook{}, &gitlab.Response{}, errBoom
					},
				},
				cr: projecthook(
					withExternalName(projectHookID),
					withProjectID(projectID),
					withTokenRef(),
					withStatus(v1alpha1.HookObservation{ID: projectHookID}),
				),
			},
			want: want{
				cr: projecthook(
					withExternalName(projectHookID),
					withProjectID(projectID),
					withTokenRef(),
					withStatus(v1alpha1.HookObservation{ID: projectHookID}),
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
		cr  *v1alpha1.Hook
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulDeletion": {
			args: args{
				projecthook: &fake.MockClient{
					MockDeleteHook: func(pid interface{}, hook int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: projecthook(
					withProjectID(projectID),
					withStatus(v1alpha1.HookObservation{
						ID: projectHookID,
					}),
					withConditions(xpv1.Available()),
				),
			},
			want: want{
				cr: projecthook(
					withProjectID(projectID),
					withStatus(v1alpha1.HookObservation{
						ID: projectHookID,
					}),
					withConditions(xpv1.Deleting()),
				),
			},
		},
		"FailedDeletion": {
			args: args{
				projecthook: &fake.MockClient{
					MockDeleteHook: func(pid interface{}, hook int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: projecthook(
					withProjectID(projectID),
					withStatus(v1alpha1.HookObservation{
						ID: projectHookID,
					}),
					withConditions(xpv1.Available()),
				),
			},
			want: want{
				cr: projecthook(
					withProjectID(projectID),
					withStatus(v1alpha1.HookObservation{
						ID: projectHookID,
					}),
					withConditions(xpv1.Deleting()),
				),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
		"InvalidHookID": {
			args: args{
				projecthook: &fake.MockClient{
					MockDeleteHook: func(pid interface{}, hook int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: projecthook(
					withProjectID(projectID),
					withConditions(xpv1.Available()),
				),
			},
			want: want{
				cr: projecthook(
					withProjectID(projectID),
					withConditions(xpv1.Deleting()),
				),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.projecthook}
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
