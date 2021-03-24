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

package deploykeys

import (
	"context"
	"fmt"
	"testing"
	"time"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/pkg/errors"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/deploykeys"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/deploykeys/fake"
)

var (
	errBoom     = errors.New("boom")
	createTime  = time.Now()
	projectID   = 5678
	deployKeyID = 1234
	title       = "example title"
	key         = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCy9s3kTnxuXKdHWAow344QlX87DdYRf6qZihisX3l55Z0TWikzESWdwbW31f6GIUh9hnRk7/VWp94bLaF/dMw1XSUhCOA8SQW3adQ0AoOOL/ZFqYu6puQDEuoVbb4wRBC7/wCKYIcNDVMVm4Nlrp0Ow3dqCOHuI3DQ8oa34QttbT15Nec0xsfFwxiebi0ZBUD1hg3SNeqEUpwmqUTQYA8LGPYHtqiAZEh55pN695/irBe2BXODvJSuUt//L5SQC2pnLSxjMoa5zK97RXrwXObvDpBZvf2l6yFhktgL+OT7VPQaOEHjVjO2ZRuUeTXOD2FkYhnEvhobXuyw/blY4DpK/jMUPHViX9yWcvxlBb/5piOZbToFsYF58n9+t73IbcHdzfDj0kZakxoCtikX4TUQg3WtaRvDzq3sBvG6u9QQUPOvtxJkJEj7aZVXqmfo+9kUiiPYYWpWfzqLT2sB0PDMMBfu62VK0m8jUUE937Wi29ezDGrHiSgP5aF5KE2G0mc="
	canPush     = false
)

type args struct {
	deploykey deploykeys.Client
	kube      client.Client
	cr        *v1alpha1.DeployKey
}

type deployKeyModifier func(*v1alpha1.DeployKey)

func withConditions(c ...xpv1.Condition) deployKeyModifier {
	return func(r *v1alpha1.DeployKey) { r.Status.ConditionedStatus.Conditions = c }
}

func withDefaultValues() deployKeyModifier {
	return func(dk *v1alpha1.DeployKey) {
		dk.Spec.ForProvider = v1alpha1.DeployKeyParameters{
			ProjectID: &projectID,
			Title:     &title,
			Key:       &key,
			CanPush:   &canPush,
		}
	}
}

func withProjectID(pid int) deployKeyModifier {
	return func(r *v1alpha1.DeployKey) {
		r.Spec.ForProvider.ProjectID = &pid
	}

}
func withCanPush(v bool) deployKeyModifier {
	return func(r *v1alpha1.DeployKey) {
		r.Spec.ForProvider.CanPush = &v
	}
}
func withKey(s string) deployKeyModifier {
	return func(r *v1alpha1.DeployKey) {
		r.Spec.ForProvider.Key = &s
	}
}

func withStatus(s v1alpha1.DeployKeyObservation) deployKeyModifier {
	return func(r *v1alpha1.DeployKey) { r.Status.AtProvider = s }
}

func withExternalName(deployKeyID int) deployKeyModifier {
	return func(r *v1alpha1.DeployKey) { meta.SetExternalName(r, fmt.Sprint(deployKeyID)) }
}

func deploykey(m ...deployKeyModifier) *v1alpha1.DeployKey {
	cr := &v1alpha1.DeployKey{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.DeployKey
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				deploykey: &fake.MockClient{
					MockGetDeployKey: func(pid interface{}, deployKeyID int, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error) {
						return &gitlab.DeployKey{
							Title:   title,
							CanPush: &canPush,
						}, &gitlab.Response{}, nil
					},
				},
				cr: deploykey(
					withDefaultValues(),
					withExternalName(deployKeyID),
					withStatus(v1alpha1.DeployKeyObservation{
						ID:        deployKeyID,
						CreatedAt: &metav1.Time{Time: createTime},
					}),
				),
			},
			want: want{
				cr: deploykey(
					withDefaultValues(),
					withExternalName(deployKeyID),
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
				deploykey: &fake.MockClient{
					MockGetDeployKey: func(pid interface{}, deployKeyID int, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error) {
						return &gitlab.DeployKey{
							CanPush: v1alpha1.Bool(true),
						}, &gitlab.Response{}, nil
					},
				},
				cr: deploykey(
					withDefaultValues(),
					withExternalName(deployKeyID),
					withStatus(v1alpha1.DeployKeyObservation{
						ID:        deployKeyID,
						CreatedAt: &metav1.Time{Time: createTime},
					}),
				),
			},
			want: want{
				cr: deploykey(
					withDefaultValues(),
					withExternalName(deployKeyID),
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
				deploykey: &fake.MockClient{
					MockGetDeployKey: func(pid interface{}, deployKeyID int, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error) {
						return &gitlab.DeployKey{
							CanPush: &canPush,
						}, &gitlab.Response{}, nil
					},
				},
				cr: deploykey(
					withProjectID(projectID),
					withExternalName(deployKeyID),
					withStatus(v1alpha1.DeployKeyObservation{
						ID:        deployKeyID,
						CreatedAt: &metav1.Time{Time: createTime},
					}),
				),
			},
			want: want{
				cr: deploykey(
					withProjectID(projectID),
					withExternalName(deployKeyID),
					withCanPush(canPush),
					withKey(""),
					withConditions(xpv1.Available()),
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
			e := &external{kube: tc.kube, client: tc.deploykey}
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
		cr     *v1alpha1.DeployKey
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
				deploykey: &fake.MockClient{
					MockAddDeployKey: func(pid interface{}, opt *gitlab.AddDeployKeyOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error) {
						return &gitlab.DeployKey{ID: deployKeyID}, &gitlab.Response{}, nil
					},
				},
				cr: deploykey(
					withDefaultValues(),
				),
			},
			want: want{
				cr: deploykey(
					withDefaultValues(),
					withConditions(xpv1.Creating()),
					withExternalName(deployKeyID),
				),
				result: managed.ExternalCreation{
					ExternalNameAssigned: true,
					ConnectionDetails:    managed.ConnectionDetails{},
				},
			},
		},
		"FailedCreation": {
			args: args{
				deploykey: &fake.MockClient{
					MockAddDeployKey: func(pid interface{}, opt *gitlab.AddDeployKeyOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error) {
						return &gitlab.DeployKey{}, &gitlab.Response{}, errBoom
					},
				},
				cr: deploykey(
					withDefaultValues(),
				),
			},
			want: want{
				cr: deploykey(
					withDefaultValues(),
					withConditions(xpv1.Creating()),
				),
				err: errors.Wrap(errBoom, errCreateFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.deploykey}
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
		cr     *v1alpha1.DeployKey
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulEditProject": {
			args: args{
				deploykey: &fake.MockClient{
					MockUpdateDeployKey: func(pid interface{}, deployKeyID int, opt *gitlab.UpdateDeployKeyOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error) {
						return &gitlab.DeployKey{}, &gitlab.Response{}, nil
					},
				},
				cr: deploykey(
					withExternalName(deployKeyID),
					withProjectID(projectID),
					withStatus(v1alpha1.DeployKeyObservation{ID: deployKeyID}),
				),
			},
			want: want{
				cr: deploykey(
					withExternalName(deployKeyID),
					withProjectID(projectID),
					withStatus(v1alpha1.DeployKeyObservation{ID: deployKeyID}),
				),
			},
		},
		"FailedEdit": {
			args: args{
				deploykey: &fake.MockClient{
					MockUpdateDeployKey: func(pid interface{}, deployKeyID int, opt *gitlab.UpdateDeployKeyOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error) {
						return &gitlab.DeployKey{}, &gitlab.Response{}, errBoom
					},
				},
				cr: deploykey(
					withExternalName(deployKeyID),
					withProjectID(projectID),
					withStatus(v1alpha1.DeployKeyObservation{ID: deployKeyID}),
				),
			},
			want: want{
				cr: deploykey(
					withExternalName(deployKeyID),
					withProjectID(projectID),
					withStatus(v1alpha1.DeployKeyObservation{ID: deployKeyID}),
				),
				err: errors.Wrap(errBoom, errUpdateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.deploykey}
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
		cr  *v1alpha1.DeployKey
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulDeletion": {
			args: args{
				deploykey: &fake.MockClient{
					MockDeleteDeployKey: func(pid interface{}, deployKeyID int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: deploykey(
					withProjectID(projectID),
					withStatus(v1alpha1.DeployKeyObservation{
						ID: deployKeyID,
					}),
					withConditions(xpv1.Available()),
				),
			},
			want: want{
				cr: deploykey(
					withProjectID(projectID),
					withStatus(v1alpha1.DeployKeyObservation{
						ID: deployKeyID,
					}),
					withConditions(xpv1.Deleting()),
				),
			},
		},
		"FailedDeletion": {
			args: args{
				deploykey: &fake.MockClient{
					MockDeleteDeployKey: func(pid interface{}, deployKeyID int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: deploykey(
					withProjectID(projectID),
					withStatus(v1alpha1.DeployKeyObservation{
						ID: deployKeyID,
					}),
					withConditions(xpv1.Available()),
				),
			},
			want: want{
				cr: deploykey(
					withProjectID(projectID),
					withStatus(v1alpha1.DeployKeyObservation{
						ID: deployKeyID,
					}),
					withConditions(xpv1.Deleting()),
				),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
		"InvaliddeployKeyID": {
			args: args{
				deploykey: &fake.MockClient{
					MockDeleteDeployKey: func(pid interface{}, deployKeyID int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: deploykey(
					withProjectID(projectID),
					withConditions(xpv1.Available()),
				),
			},
			want: want{
				cr: deploykey(
					withProjectID(projectID),
					withConditions(xpv1.Deleting()),
				),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.deploykey}
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
