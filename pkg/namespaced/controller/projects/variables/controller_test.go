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

package variables

import (
	"context"
	"net/http"
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/errors"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	commonv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects/fake"
)

var (
	errBoom             = errors.New("boom")
	projectID           = int64(5678)
	variableKey         = "VARIABLE_KEY"
	variableValue       = "1234"
	variableType        = commonv1alpha1.VariableTypeEnvVar
	variableEnvScope    = "*"
	variableDescription = "desc"
	f                   = false
)

var (
	pv = gitlab.ProjectVariable{
		Value:            variableValue,
		Key:              variableKey,
		Description:      variableDescription,
		EnvironmentScope: variableEnvScope,
		VariableType:     gitlab.VariableTypeValue(variableType),
		Protected:        f,
		Masked:           f,
		Raw:              f,
	}
)

type args struct {
	variable projects.VariableClient
	kube     client.Client
	cr       *v1alpha1.Variable
}

type variableModifier func(*v1alpha1.Variable)

func withConditions(c ...xpv1.Condition) variableModifier {
	return func(r *v1alpha1.Variable) { r.Status.ConditionedStatus.Conditions = c }
}

func withDefaultValues() variableModifier {
	return func(pv *v1alpha1.Variable) {
		pv.Spec.ForProvider = v1alpha1.VariableParameters{
			ProjectID: &projectID,
			CommonVariableParameters: commonv1alpha1.CommonVariableParameters{
				Key:          variableKey,
				Value:        &variableValue,
				Description:  &variableDescription,
				Protected:    &f,
				Masked:       &f,
				Raw:          &f,
				VariableType: &variableType,
			},
			EnvironmentScope: &variableEnvScope,
		}
	}
}

func withProjectID(pid int64) variableModifier {
	return func(r *v1alpha1.Variable) {
		r.Spec.ForProvider.ProjectID = &pid
	}
}

func withValue(value string) variableModifier {
	return func(r *v1alpha1.Variable) {
		r.Spec.ForProvider.Value = &value
	}
}

func withValueSecretRef(selector *xpv1.LocalSecretKeySelector) variableModifier {
	return func(r *v1alpha1.Variable) {
		r.Spec.ForProvider.ValueSecretRef = selector
	}
}

func withKey(key string) variableModifier {
	return func(r *v1alpha1.Variable) {
		r.Spec.ForProvider.Key = key
	}
}

func withMasked(masked bool) variableModifier {
	return func(r *v1alpha1.Variable) {
		r.Spec.ForProvider.Masked = &masked
	}
}

func withDescription(description string) variableModifier {
	return func(r *v1alpha1.Variable) {
		r.Spec.ForProvider.Description = &description
	}
}

func withRaw(raw bool) variableModifier {
	return func(r *v1alpha1.Variable) {
		r.Spec.ForProvider.Raw = &raw
	}
}

func withVariableType(variableType commonv1alpha1.VariableType) variableModifier {
	return func(r *v1alpha1.Variable) {
		r.Spec.ForProvider.VariableType = &variableType
	}
}

func withEnvironmentScope(scope string) variableModifier {
	return func(r *v1alpha1.Variable) {
		r.Spec.ForProvider.EnvironmentScope = &scope
	}
}

var deletionTime = metav1.Now()

func withDeletionTimestamp() variableModifier {
	return func(r *v1alpha1.Variable) {
		r.ObjectMeta.DeletionTimestamp = &deletionTime
	}
}

func variable(m ...variableModifier) *v1alpha1.Variable {
	cr := &v1alpha1.Variable{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.Variable
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				variable: &fake.MockClient{
					MockGetVariable: func(pid interface{}, key string, opt *gitlab.GetProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error) {
						return &pv, &gitlab.Response{}, nil
					},
				},
				cr: variable(withDefaultValues()),
			},
			want: want{
				cr: variable(
					withDefaultValues(),
					withDescription(variableDescription),
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
				variable: &fake.MockClient{
					MockGetVariable: func(pid interface{}, key string, opt *gitlab.GetProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error) {
						rv := pv
						rv.Value = "not-up-to-date"
						return &rv, &gitlab.Response{}, nil
					},
				},
				cr: variable(
					withDefaultValues(),
					withValue("blah"),
				),
			},
			want: want{
				cr: variable(
					withDefaultValues(),
					withValue("blah"),
					withDescription(variableDescription),
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
				variable: &fake.MockClient{
					MockGetVariable: func(pid interface{}, key string, opt *gitlab.GetProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error) {
						rv := pv
						rv.Masked = true
						rv.VariableType = gitlab.FileVariableType
						return &rv, &gitlab.Response{}, nil
					},
				},
				cr: variable(
					withProjectID(projectID),
					withKey(variableKey),
					withValue(variableValue),
					withVariableType(commonv1alpha1.VariableTypeEnvVar),
					withRaw(false),
				),
			},
			want: want{
				cr: variable(
					withDefaultValues(),
					withKey(variableKey),
					// We expect the masked value to be late-inited to true
					withMasked(true),
					// We expect the variable type value to be unchanged,
					// as it was already set in the existing CR.
					withVariableType(commonv1alpha1.VariableTypeEnvVar),
					withDescription(variableDescription),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists: true,
					// Resource is not up to date as local and remote
					// variableType setting do not match.
					ResourceUpToDate:        false,
					ResourceLateInitialized: true,
				},
			},
		},
		"GetError": {
			args: args{
				variable: &fake.MockClient{
					MockGetVariable: func(pid interface{}, key string, opt *gitlab.GetProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error) {
						return &gitlab.ProjectVariable{}, &gitlab.Response{Response: &http.Response{StatusCode: 400}}, errBoom
					},
				},
				cr: variable(
					withDefaultValues(),
					withValue("blah"),
				),
			},
			want: want{
				cr: variable(
					withDefaultValues(),
					withValue("blah"),
				),
				result: managed.ExternalObservation{},
				err:    errors.Wrap(errBoom, errGetFailed),
			},
		},
		"ErrGet404": {
			args: args{
				variable: &fake.MockClient{
					MockGetVariable: func(pid interface{}, key string, opt *gitlab.GetProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error) {
						return &gitlab.ProjectVariable{}, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errBoom
					},
				},
				cr: variable(
					withDefaultValues(),
					withValue("blah"),
				),
			},
			want: want{
				cr: variable(
					withDefaultValues(),
					withValue("blah"),
				),
				result: managed.ExternalObservation{},
				err:    nil,
			},
		},
		"ValueSecretRef": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret, ok := obj.(*corev1.Secret)
						if !ok {
							return errors.Wrapf(errBoom, "unexpected object type %T, expected %T", obj, secret)
						}

						secret.Data = map[string][]byte{
							"blah": []byte(variableValue),
						}

						return nil
					},
				},
				variable: &fake.MockClient{
					MockGetVariable: func(pid interface{}, key string, opt *gitlab.GetProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error) {
						return &pv, &gitlab.Response{}, nil
					},
				},
				cr: variable(
					// We don't want to use the default values used in the other tests since we want value to be empty and
					// Raw to be nil.
					withProjectID(projectID),
					withKey(variableKey),
					withValueSecretRef(common.TestCreateLocalSecretKeySelector("", "blah")),
					withEnvironmentScope("*"),
					withVariableType(commonv1alpha1.VariableTypeEnvVar),
				),
			},
			want: want{
				cr: variable(
					withDefaultValues(),
					withValueSecretRef(common.TestCreateLocalSecretKeySelector("", "blah")),
					withMasked(true),
					withRaw(true),
					withDescription(variableDescription),
					withConditions(xpv1.Available()),
					withVariableType(commonv1alpha1.VariableTypeEnvVar),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceLateInitialized: true,
				},
			},
		},
		"ValueSecretRefWrongKey": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret, ok := obj.(*corev1.Secret)
						if !ok {
							return errors.Wrapf(errBoom, "unexpected object type %T, expected %T", obj, secret)
						}

						secret.Data = map[string][]byte{
							"blah": []byte(variableValue),
						}

						return nil
					},
				},
				variable: &fake.MockClient{
					MockGetVariable: func(pid interface{}, key string, opt *gitlab.GetProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error) {
						return &gitlab.ProjectVariable{}, &gitlab.Response{Response: &http.Response{StatusCode: 400}}, errors.New(common.ErrSecretKeyNotFound)
					},
				},
				cr: variable(
					withProjectID(projectID),
					withKey(variableKey),
					withValueSecretRef(common.TestCreateLocalSecretKeySelector("", "bad")),
				),
			},
			want: want{
				cr: variable(
					withProjectID(projectID),
					withKey(variableKey),
					withValueSecretRef(common.TestCreateLocalSecretKeySelector("", "bad")),
				),
				err: errors.Wrap(errors.New(common.ErrSecretKeyNotFound), errGetFailed),
			},
		},
		"DeletingEarlyReturnSkipsSecret": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error { return errBoom },
				},
				variable: &fake.MockClient{
					MockGetVariable: func(pid interface{}, key string, opt *gitlab.GetProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error) {
						return &pv, &gitlab.Response{}, nil
					},
				},
				cr: variable(
					withProjectID(projectID),
					withKey(variableKey),
					withValueSecretRef(common.TestCreateLocalSecretKeySelector("", "blah")),
					withDeletionTimestamp(),
					withConditions(xpv1.Deleting()),
				),
			},
			want: want{
				cr: variable(
					withProjectID(projectID),
					withKey(variableKey),
					withValueSecretRef(common.TestCreateLocalSecretKeySelector("", "blah")),
					withDeletionTimestamp(),
					withConditions(xpv1.Deleting()),
				),
				result: managed.ExternalObservation{ResourceExists: true},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.variable}
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
		cr     *v1alpha1.Variable
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
				variable: &fake.MockClient{
					MockCreateVariable: func(pid interface{}, opt *gitlab.CreateProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error) {
						return &gitlab.ProjectVariable{Key: variableKey}, &gitlab.Response{}, nil
					},
				},
				cr: variable(
					withDefaultValues(),
				),
			},
			want: want{
				cr: variable(
					withDefaultValues(),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalCreation{},
			},
		},
		"FailedCreation": {
			args: args{
				variable: &fake.MockClient{
					MockCreateVariable: func(pid interface{}, opt *gitlab.CreateProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error) {
						return &gitlab.ProjectVariable{}, &gitlab.Response{}, errBoom
					},
				},
				cr: variable(
					withDefaultValues(),
				),
			},
			want: want{
				cr: variable(
					withDefaultValues(),
					withConditions(xpv1.Creating()),
				),
				err: errors.Wrap(errBoom, errCreateFailed),
			},
		},
		"ValueSecretRef": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret := obj.(*corev1.Secret)
						secret.Data = map[string][]byte{
							"blah": []byte(variableValue),
						}

						return nil
					},
				},
				variable: &fake.MockClient{
					MockCreateVariable: func(pid interface{}, opt *gitlab.CreateProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error) {
						return &pv, &gitlab.Response{}, nil
					},
				},
				cr: variable(
					withProjectID(projectID),
					withKey(variableKey),
					withValueSecretRef(common.TestCreateLocalSecretKeySelector("", "blah")),
				),
			},
			want: want{
				cr: variable(
					withProjectID(projectID),
					withKey(variableKey),
					withConditions(xpv1.Creating()),
					withValueSecretRef(common.TestCreateLocalSecretKeySelector("", "blah")),
					withValue(variableValue),
					withMasked(true),
					withRaw(true),
				),
			},
		},
		"ValueSecretRefWrongKey": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret := obj.(*corev1.Secret)
						secret.Data = map[string][]byte{
							"blah": []byte(variableValue),
						}

						return nil
					},
				},
				variable: &fake.MockClient{
					MockCreateVariable: func(pid interface{}, opt *gitlab.CreateProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error) {
						return &gitlab.ProjectVariable{}, &gitlab.Response{}, errors.New(common.ErrSecretKeyNotFound)
					},
				},
				cr: variable(
					withProjectID(projectID),
					withKey(variableKey),
					withValueSecretRef(common.TestCreateLocalSecretKeySelector("", "bad")),
				),
			},
			want: want{
				cr: variable(
					withProjectID(projectID),
					withKey(variableKey),
					withValueSecretRef(common.TestCreateLocalSecretKeySelector("", "bad")),
				),
				err: errors.Wrap(errors.New(common.ErrSecretKeyNotFound), errCreateFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.variable}
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
		cr     *v1alpha1.Variable
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulEditProject": {
			args: args{
				variable: &fake.MockClient{
					MockUpdateVariable: func(pid interface{}, key string, opt *gitlab.UpdateProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error) {
						return &gitlab.ProjectVariable{}, &gitlab.Response{}, nil
					},
				},
				cr: variable(
					withKey(variableKey),
					withProjectID(projectID),
				),
			},
			want: want{
				cr: variable(
					withKey(variableKey),
					withProjectID(projectID),
				),
			},
		},
		"FailedEdit": {
			args: args{
				variable: &fake.MockClient{
					MockUpdateVariable: func(pid interface{}, key string, opt *gitlab.UpdateProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error) {
						return &gitlab.ProjectVariable{}, &gitlab.Response{}, errBoom
					},
				},
				cr: variable(
					withKey(variableKey),
					withProjectID(projectID),
				),
			},
			want: want{
				cr: variable(
					withKey(variableKey),
					withProjectID(projectID),
				),
				err: errors.Wrap(errBoom, errUpdateFailed),
			},
		},
		"ValueSecretRef": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret, ok := obj.(*corev1.Secret)
						if !ok {
							return errors.Wrapf(errBoom, "unexpected object type %T, expected %T", obj, secret)
						}

						secret.Data = map[string][]byte{
							"blah": []byte(variableValue),
						}

						return nil
					},
				},
				variable: &fake.MockClient{
					MockUpdateVariable: func(pid interface{}, key string, opt *gitlab.UpdateProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error) {
						return &pv, &gitlab.Response{}, nil
					},
				},
				cr: variable(
					withProjectID(projectID),
					withKey(variableKey),
					withValueSecretRef(common.TestCreateLocalSecretKeySelector("", "blah")),
				),
			},
			want: want{
				cr: variable(
					withProjectID(projectID),
					withKey(variableKey),
					withValueSecretRef(common.TestCreateLocalSecretKeySelector("", "blah")),
					withValue(variableValue),
					withMasked(true),
					withRaw(true),
				),
			},
		},
		"ValueSecretRefWrongKey": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret, ok := obj.(*corev1.Secret)
						if !ok {
							return errors.Wrapf(errBoom, "unexpected object type %T, expected %T", obj, secret)
						}

						secret.Data = map[string][]byte{
							"blah": []byte(variableValue),
						}

						return nil
					},
				},
				variable: &fake.MockClient{
					MockUpdateVariable: func(pid interface{}, key string, opt *gitlab.UpdateProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error) {
						return &gitlab.ProjectVariable{}, &gitlab.Response{}, errors.New(common.ErrSecretKeyNotFound)
					},
				},
				cr: variable(
					withProjectID(projectID),
					withKey(variableKey),
					withValueSecretRef(common.TestCreateLocalSecretKeySelector("", "bad")),
				),
			},
			want: want{
				cr: variable(
					withProjectID(projectID),
					withKey(variableKey),
					withValueSecretRef(common.TestCreateLocalSecretKeySelector("", "bad")),
				),
				err: errors.Wrap(errors.New(common.ErrSecretKeyNotFound), errUpdateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.variable}
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
		cr  *v1alpha1.Variable
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulDeletion": {
			args: args{
				variable: &fake.MockClient{
					MockRemoveVariable: func(pid interface{}, key string, opt *gitlab.RemoveProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: variable(
					withProjectID(projectID),
					withConditions(xpv1.Available()),
				),
			},
			want: want{
				cr: variable(
					withProjectID(projectID),
					withConditions(xpv1.Deleting()),
				),
			},
		},
		"FailedDeletion": {
			args: args{
				variable: &fake.MockClient{
					MockRemoveVariable: func(pid interface{}, key string, opt *gitlab.RemoveProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: variable(
					withProjectID(projectID),
					withConditions(xpv1.Available()),
				),
			},
			want: want{
				cr: variable(
					withProjectID(projectID),
					withConditions(xpv1.Deleting()),
				),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
		"InvalidVariableID": {
			args: args{
				variable: &fake.MockClient{
					MockRemoveVariable: func(pid interface{}, key string, opt *gitlab.RemoveProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: variable(
					withProjectID(projectID),
					withConditions(xpv1.Available()),
				),
			},
			want: want{
				cr: variable(
					withProjectID(projectID),
					withConditions(xpv1.Deleting()),
				),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.variable}
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
