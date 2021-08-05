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
	"strconv"
	"testing"
	"time"

	"github.com/pkg/errors"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
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
	id            = 0
	deployKeyID   = 1234
	sDeployKeyID  = strconv.Itoa(deployKeyID)
	unexpecedItem resource.Managed
	createdAt     = time.Now()
	title         = "Title"
	key           = "key"
	canPush       = false
	canPushTrue   = true
	deployKeyObj  = gitlab.DeployKey{
		ID:        deployKeyID,
		Title:     title,
		Key:       key,
		CanPush:   &canPush,
		CreatedAt: &createdAt,
	}

	extNameAnnotation = map[string]string{meta.AnnotationKeyExternalName: fmt.Sprint(deployKeyID)}
)

type args struct {
	deployKey projects.DeployKeyClient
	kube      client.Client
	cr        resource.Managed
}

type deployKeyModifier func(*v1alpha1.DeployKey)

func withConditions(c ...xpv1.Condition) deployKeyModifier {
	return func(r *v1alpha1.DeployKey) { r.Status.ConditionedStatus.Conditions = c }
}

func withStatus(ap v1alpha1.DeployKeyObservation) deployKeyModifier {
	return func(r *v1alpha1.DeployKey) { r.Status.AtProvider = ap }
}

func withSpec(fp v1alpha1.DeployKeyParameters) deployKeyModifier {
	return func(r *v1alpha1.DeployKey) { r.Spec.ForProvider = fp }
}

func withExternalName(deployKeyID string) deployKeyModifier {
	return func(r *v1alpha1.DeployKey) { meta.SetExternalName(r, deployKeyID) }
}

func withAnnotations(a map[string]string) deployKeyModifier {
	return func(p *v1alpha1.DeployKey) { meta.AddAnnotations(p, a) }
}

func deployKey(m ...deployKeyModifier) *v1alpha1.DeployKey {
	cr := &v1alpha1.DeployKey{}
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
				err: errors.New(errNotDeployKey),
			},
		},
		"NoExternalName": {
			args: args{
				cr: deployKey(),
			},
			want: want{
				cr: deployKey(),
				result: managed.ExternalObservation{
					ResourceExists:          false,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
			},
		},
		"NotIDExternalName": {
			args: args{
				deployKey: &fake.MockClient{
					MockGetDeployKey: func(pid interface{}, deployKeyID int, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error) {
						return &gitlab.DeployKey{}, &gitlab.Response{}, nil
					},
				},
				cr: deployKey(withExternalName("fr")),
			},
			want: want{
				cr:  deployKey(withExternalName("fr")),
				err: errors.New(errNotDeployKey),
			},
		},
		"FailedGetRequest": {
			args: args{
				deployKey: &fake.MockClient{
					MockGetDeployKey: func(pid interface{}, deployKeyID int, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error) {
						return nil, &gitlab.Response{}, errBoom
					},
				},
				cr: deployKey(
					withExternalName(sDeployKeyID),
					withSpec(v1alpha1.DeployKeyParameters{
						ProjectID: &deployKeyID,
					}),
				),
			},
			want: want{
				cr: deployKey(
					withAnnotations(extNameAnnotation),
					withSpec(v1alpha1.DeployKeyParameters{
						ProjectID: &deployKeyID,
					}),
				),
				result: managed.ExternalObservation{ResourceExists: false},
				err:    nil,
			},
		},
		"IsDeployKeyUpToDateTitle": {
			args: args{
				deployKey: &fake.MockClient{
					MockGetDeployKey: func(pid interface{}, deployKeyID int, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error) {
						return &deployKeyObj, &gitlab.Response{}, nil
					},
				},
				cr: deployKey(
					withStatus(v1alpha1.DeployKeyObservation{
						ID:        deployKeyID,
						CreatedAt: &metav1.Time{Time: createdAt},
					}),
					withSpec(v1alpha1.DeployKeyParameters{
						ProjectID: &deployKeyID,
						Title:     "New Title",
						Key:       &key,
						CanPush:   &canPush,
					}),
					withExternalName(sDeployKeyID),
				),
			},
			want: want{
				cr: deployKey(
					withStatus(v1alpha1.DeployKeyObservation{
						ID:        deployKeyID,
						CreatedAt: &metav1.Time{Time: createdAt},
					}),
					withSpec(v1alpha1.DeployKeyParameters{
						ProjectID: &deployKeyID,
						Title:     "New Title",
						Key:       &key,
						CanPush:   &canPush,
					}),
					withConditions(xpv1.Available()),
					withExternalName(sDeployKeyID),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
			},
		},
		"IsDeployKeyUpToDateCanPush": {
			args: args{
				deployKey: &fake.MockClient{
					MockGetDeployKey: func(pid interface{}, deployKeyID int, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error) {
						return &deployKeyObj, &gitlab.Response{}, nil
					},
				},
				cr: deployKey(
					withStatus(v1alpha1.DeployKeyObservation{
						ID:        deployKeyID,
						CreatedAt: &metav1.Time{Time: createdAt},
					}),
					withSpec(v1alpha1.DeployKeyParameters{
						ProjectID: &deployKeyID,
						Title:     title,
						Key:       &key,
						CanPush:   &canPushTrue,
					}),
					withExternalName(sDeployKeyID),
				),
			},
			want: want{
				cr: deployKey(
					withStatus(v1alpha1.DeployKeyObservation{
						ID:        deployKeyID,
						CreatedAt: &metav1.Time{Time: createdAt},
					}),
					withSpec(v1alpha1.DeployKeyParameters{
						ProjectID: &deployKeyID,
						Title:     title,
						Key:       &key,
						CanPush:   &canPushTrue,
					}),
					withConditions(xpv1.Available()),
					withExternalName(sDeployKeyID),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
			},
		},
		"LateInitSuccess": {
			args: args{
				deployKey: &fake.MockClient{
					MockGetDeployKey: func(pid interface{}, deployKeyID int, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error) {
						return &deployKeyObj, &gitlab.Response{}, nil
					},
				},
				cr: deployKey(
					withStatus(v1alpha1.DeployKeyObservation{
						ID:        deployKeyID,
						CreatedAt: &metav1.Time{Time: createdAt},
					}),
					withSpec(v1alpha1.DeployKeyParameters{
						ProjectID: &deployKeyID,
					}),
					withExternalName(sDeployKeyID),
				),
			},
			want: want{
				cr: deployKey(
					withStatus(v1alpha1.DeployKeyObservation{
						ID:        deployKeyID,
						CreatedAt: &metav1.Time{Time: createdAt},
					}),
					withSpec(v1alpha1.DeployKeyParameters{
						ProjectID: &deployKeyID,
						CanPush:   &canPush,
					}),
					withConditions(xpv1.Available()),
					withExternalName(sDeployKeyID),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: true,
				},
			},
		},
		"SuccessfulAvailable": {
			args: args{
				deployKey: &fake.MockClient{
					MockGetDeployKey: func(pid interface{}, deployKeyID int, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error) {
						return &deployKeyObj, &gitlab.Response{}, nil
					},
				},
				cr: deployKey(
					withStatus(v1alpha1.DeployKeyObservation{
						ID:        deployKeyID,
						CreatedAt: &metav1.Time{Time: createdAt},
					}),
					withSpec(v1alpha1.DeployKeyParameters{
						ProjectID: &deployKeyID,
						Title:     title,
						Key:       &key,
						CanPush:   &canPush,
					}),
					withExternalName(sDeployKeyID),
				),
			},
			want: want{
				cr: deployKey(
					withStatus(v1alpha1.DeployKeyObservation{
						ID:        deployKeyID,
						CreatedAt: &metav1.Time{Time: createdAt},
					}),
					withSpec(v1alpha1.DeployKeyParameters{
						ProjectID: &deployKeyID,
						Title:     title,
						Key:       &key,
						CanPush:   &canPush,
					}),
					withConditions(xpv1.Available()),
					withExternalName(sDeployKeyID),
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
			e := &external{kube: tc.kube, client: tc.deployKey}
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
				err: errors.New(errNotDeployKey),
			},
		},
		"SuccessfulCreation": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				deployKey: &fake.MockClient{
					MockAddDeployKey: func(pid interface{}, opt *gitlab.AddDeployKeyOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error) {
						return &deployKeyObj, &gitlab.Response{}, nil
					},
				},
				cr: deployKey(
					withAnnotations(extNameAnnotation),
					withSpec(v1alpha1.DeployKeyParameters{
						ProjectID: &deployKeyID,
						Key:       &key,
					}),
				),
			},
			want: want{
				cr: deployKey(
					withExternalName(sDeployKeyID),
					withSpec(v1alpha1.DeployKeyParameters{
						ProjectID: &deployKeyID,
						Key:       &key,
					}),
				),
				result: managed.ExternalCreation{
					ExternalNameAssigned: true,
					ConnectionDetails:    managed.ConnectionDetails{},
				},
				err: nil,
			},
		},
		"FailedCreation": {
			args: args{
				deployKey: &fake.MockClient{
					MockAddDeployKey: func(pid interface{}, opt *gitlab.AddDeployKeyOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error) {
						return &gitlab.DeployKey{}, &gitlab.Response{}, errBoom
					},
				},
				cr: deployKey(
					withSpec(v1alpha1.DeployKeyParameters{
						ProjectID: &deployKeyID,
						Key:       &key,
					}),
					withStatus(v1alpha1.DeployKeyObservation{
						ID: 0,
					}),
				),
			},
			want: want{
				cr: deployKey(
					withSpec(v1alpha1.DeployKeyParameters{
						ProjectID: &deployKeyID,
						Key:       &key,
					}),
					withStatus(v1alpha1.DeployKeyObservation{
						ID: 0,
					}),
				),
				result: managed.ExternalCreation{},
				err:    errors.Wrap(errBoom, errCreateFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.deployKey}
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
				err: errors.New(errNotDeployKey),
			},
		},
		"SuccessfulUpdate": {
			args: args{
				deployKey: &fake.MockClient{
					MockUpdateDeployKey: func(pid interface{}, deployKeyID int, opt *gitlab.UpdateDeployKeyOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error) {
						return &deployKeyObj, &gitlab.Response{}, nil
					},
				},
				cr: deployKey(
					withSpec(v1alpha1.DeployKeyParameters{
						ProjectID: &deployKeyID,
						Title:     "New Title",
					}),
					withStatus(v1alpha1.DeployKeyObservation{
						ID: id,
					}),
				),
			},
			want: want{
				cr: deployKey(
					withSpec(v1alpha1.DeployKeyParameters{
						ProjectID: &deployKeyID,
						Title:     "New Title",
					}),
					withStatus(v1alpha1.DeployKeyObservation{
						ID: id,
					}),
				),
				result: managed.ExternalUpdate{},
				err:    nil,
			},
		},
		"FailedUpdate": {
			args: args{
				deployKey: &fake.MockClient{
					MockUpdateDeployKey: func(pid interface{}, deployKeyID int, opt *gitlab.UpdateDeployKeyOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error) {
						return &gitlab.DeployKey{}, &gitlab.Response{}, errBoom
					},
				},
				cr: deployKey(
					withSpec(v1alpha1.DeployKeyParameters{
						ProjectID: &deployKeyID,
					}),
					withStatus(v1alpha1.DeployKeyObservation{
						ID: id,
					}),
				),
			},
			want: want{
				cr: deployKey(
					withSpec(v1alpha1.DeployKeyParameters{
						ProjectID: &deployKeyID,
					}),
					withStatus(v1alpha1.DeployKeyObservation{
						ID: id,
					}),
				),
				result: managed.ExternalUpdate{},
				err:    errors.Wrap(errBoom, errUpdateFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.deployKey}
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
				err: errors.New(errNotDeployKey),
			},
		},
		"SuccessfulDeletion": {
			args: args{
				deployKey: &fake.MockClient{
					MockDeleteDeployKey: func(pid interface{}, deployKey int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: deployKey(
					withSpec(v1alpha1.DeployKeyParameters{
						ProjectID: &deployKeyID,
					}),
					withStatus(v1alpha1.DeployKeyObservation{
						ID: id,
					}),
				),
			},
			want: want{
				cr: deployKey(
					withSpec(v1alpha1.DeployKeyParameters{
						ProjectID: &deployKeyID,
					}),
					withStatus(v1alpha1.DeployKeyObservation{
						ID: id,
					}),
				),
			},
		},
		"FailedDeletion": {
			args: args{
				deployKey: &fake.MockClient{
					MockDeleteDeployKey: func(pid interface{}, deployKey int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: deployKey(
					withSpec(v1alpha1.DeployKeyParameters{
						ProjectID: &deployKeyID,
					}),
					withStatus(v1alpha1.DeployKeyObservation{
						ID: id,
					}),
				),
			},
			want: want{
				cr: deployKey(
					withSpec(v1alpha1.DeployKeyParameters{
						ProjectID: &deployKeyID,
					}),
					withStatus(v1alpha1.DeployKeyObservation{
						ID: id,
					}),
				),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.deployKey}
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
