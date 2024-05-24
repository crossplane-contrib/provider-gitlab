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

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups/fake"
)

var (
	errBoom          = errors.New("boom")
	groupID          = 5678
	variableKey      = "VARIABLE_KEY"
	variableValue    = "1234"
	variableType     = v1alpha1.VariableTypeEnvVar
	variableEnvScope = "*"
	f                = false
)

var (
	pv = gitlab.GroupVariable{
		Value:            variableValue,
		Key:              variableKey,
		EnvironmentScope: variableEnvScope,
		VariableType:     gitlab.VariableTypeValue(variableType),
		Protected:        f,
		Masked:           f,
		Raw:              f,
	}
)

type args struct {
	variable groups.VariableClient
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
			GroupID:          &groupID,
			Key:              variableKey,
			Value:            &variableValue,
			Protected:        &f,
			Masked:           &f,
			Raw:              &f,
			VariableType:     &variableType,
			EnvironmentScope: &variableEnvScope,
		}
	}
}

func withGroupID(pid int) variableModifier {
	return func(r *v1alpha1.Variable) {
		r.Spec.ForProvider.GroupID = &pid
	}
}

func withValue(value string) variableModifier {
	return func(r *v1alpha1.Variable) {
		r.Spec.ForProvider.Value = &value
	}
}

func withValueSecretRef(selector *xpv1.SecretKeySelector) variableModifier {
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

func withRaw(raw bool) variableModifier {
	return func(r *v1alpha1.Variable) {
		r.Spec.ForProvider.Raw = &raw
	}
}

func withVariableType(variableType v1alpha1.VariableType) variableModifier {
	return func(r *v1alpha1.Variable) {
		r.Spec.ForProvider.VariableType = &variableType
	}
}

func withEnvironmentScope(scope string) variableModifier {
	return func(r *v1alpha1.Variable) {
		r.Spec.ForProvider.EnvironmentScope = &scope
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
					MockGetGroupVariable: func(gid interface{}, key string, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error) {
						return &pv, &gitlab.Response{}, nil
					},
				},
				cr: variable(withDefaultValues()),
			},
			want: want{
				cr: variable(
					withDefaultValues(),
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
					MockGetGroupVariable: func(gid interface{}, key string, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error) {
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
					MockGetGroupVariable: func(gid interface{}, key string, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error) {
						rv := pv
						rv.Masked = true
						rv.VariableType = gitlab.FileVariableType
						return &rv, &gitlab.Response{}, nil
					},
				},
				cr: variable(
					withGroupID(groupID),
					withKey(variableKey),
					withValue(variableValue),
					withVariableType(v1alpha1.VariableTypeEnvVar),
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
					withVariableType(v1alpha1.VariableTypeEnvVar),
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
					MockGetGroupVariable: func(gid interface{}, key string, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error) {
						return &gitlab.GroupVariable{}, &gitlab.Response{Response: &http.Response{StatusCode: 400}}, errBoom
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
					MockGetGroupVariable: func(gid interface{}, key string, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error) {
						return &gitlab.GroupVariable{}, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errBoom
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
					MockGetGroupVariable: func(gid interface{}, key string, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error) {
						return &gitlab.GroupVariable{}, &gitlab.Response{}, nil
					},
				},
				cr: variable(
					// We don't want to use the default values used in the other tests since we want value to be empty and
					// Raw to be nil.
					withGroupID(groupID),
					withKey(variableKey),
					withValueSecretRef(&xpv1.SecretKeySelector{
						SecretReference: xpv1.SecretReference{},
						Key:             "blah",
					}),
					withEnvironmentScope("*"),
					withVariableType(v1alpha1.VariableTypeEnvVar),
				),
			},
			want: want{
				cr: variable(
					withDefaultValues(),
					withValueSecretRef(&xpv1.SecretKeySelector{
						SecretReference: xpv1.SecretReference{},
						Key:             "blah",
					}),
					withMasked(true),
					withRaw(true),
					withConditions(xpv1.Available()),
					withVariableType(v1alpha1.VariableTypeEnvVar),
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
					MockGetGroupVariable: func(gid interface{}, key string, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error) {
						return &gitlab.GroupVariable{}, &gitlab.Response{Response: &http.Response{StatusCode: 400}}, errors.New(errSecretKeyNotFound)
					},
				},
				cr: variable(
					withGroupID(groupID),
					withKey(variableKey),
					withValueSecretRef(&xpv1.SecretKeySelector{
						SecretReference: xpv1.SecretReference{},
						Key:             "bad",
					}),
				),
			},
			want: want{
				cr: variable(
					withGroupID(groupID),
					withKey(variableKey),
					withValueSecretRef(&xpv1.SecretKeySelector{
						SecretReference: xpv1.SecretReference{},
						Key:             "bad",
					}),
				),
				err: errors.Wrap(errors.New(errSecretKeyNotFound), errGetFailed),
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
					MockCreateGroupVariable: func(gid interface{}, opt *gitlab.CreateGroupVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error) {
						return &gitlab.GroupVariable{Key: variableKey}, &gitlab.Response{}, nil
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
					MockCreateGroupVariable: func(gid interface{}, opt *gitlab.CreateGroupVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error) {
						return &gitlab.GroupVariable{}, &gitlab.Response{}, errBoom
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
					MockCreateGroupVariable: func(gid interface{}, opt *gitlab.CreateGroupVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error) {
						return &gitlab.GroupVariable{}, &gitlab.Response{}, nil
					},
				},
				cr: variable(
					withGroupID(groupID),
					withKey(variableKey),
					withValueSecretRef(&xpv1.SecretKeySelector{
						SecretReference: xpv1.SecretReference{},
						Key:             "blah",
					}),
				),
			},
			want: want{
				cr: variable(
					withGroupID(groupID),
					withKey(variableKey),
					withConditions(xpv1.Creating()),
					withValueSecretRef(&xpv1.SecretKeySelector{
						SecretReference: xpv1.SecretReference{},
						Key:             "blah",
					}),
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
					MockCreateGroupVariable: func(gid interface{}, opt *gitlab.CreateGroupVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error) {
						return &gitlab.GroupVariable{}, &gitlab.Response{}, errors.New(errSecretKeyNotFound)
					},
				},
				cr: variable(
					withGroupID(groupID),
					withKey(variableKey),
					withValueSecretRef(&xpv1.SecretKeySelector{
						SecretReference: xpv1.SecretReference{},
						Key:             "bad",
					}),
				),
			},
			want: want{
				cr: variable(
					withGroupID(groupID),
					withKey(variableKey),
					withValueSecretRef(&xpv1.SecretKeySelector{
						SecretReference: xpv1.SecretReference{},
						Key:             "bad",
					})),
				err: errors.Wrap(errors.New(errSecretKeyNotFound), errCreateFailed),
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
		"SuccessfulEditGroup": {
			args: args{
				variable: &fake.MockClient{
					MockUpdateGroupVariable: func(gid interface{}, key string, opt *gitlab.UpdateGroupVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error) {
						return &gitlab.GroupVariable{}, &gitlab.Response{}, nil
					},
				},
				cr: variable(
					withKey(variableKey),
					withGroupID(groupID),
				),
			},
			want: want{
				cr: variable(
					withKey(variableKey),
					withGroupID(groupID),
				),
			},
		},
		"FailedEdit": {
			args: args{
				variable: &fake.MockClient{
					MockUpdateGroupVariable: func(gid interface{}, key string, opt *gitlab.UpdateGroupVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error) {
						return &gitlab.GroupVariable{}, &gitlab.Response{}, errBoom
					},
				},
				cr: variable(
					withKey(variableKey),
					withGroupID(groupID),
				),
			},
			want: want{
				cr: variable(
					withKey(variableKey),
					withGroupID(groupID),
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
					MockUpdateGroupVariable: func(gid interface{}, key string, opt *gitlab.UpdateGroupVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error) {
						return &gitlab.GroupVariable{}, &gitlab.Response{}, nil
					},
				},
				cr: variable(
					withGroupID(groupID),
					withKey(variableKey),
					withValueSecretRef(&xpv1.SecretKeySelector{
						SecretReference: xpv1.SecretReference{},
						Key:             "blah",
					}),
				),
			},
			want: want{
				cr: variable(
					withGroupID(groupID),
					withKey(variableKey),
					withValueSecretRef(&xpv1.SecretKeySelector{
						SecretReference: xpv1.SecretReference{},
						Key:             "blah",
					}),
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
					MockUpdateGroupVariable: func(gid interface{}, key string, opt *gitlab.UpdateGroupVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error) {
						return &gitlab.GroupVariable{}, &gitlab.Response{}, errors.New(errSecretKeyNotFound)
					},
				},
				cr: variable(
					withGroupID(groupID),
					withKey(variableKey),
					withValueSecretRef(&xpv1.SecretKeySelector{
						SecretReference: xpv1.SecretReference{},
						Key:             "bad",
					}),
				),
			},
			want: want{
				cr: variable(
					withGroupID(groupID),
					withKey(variableKey),
					withValueSecretRef(&xpv1.SecretKeySelector{
						SecretReference: xpv1.SecretReference{},
						Key:             "bad",
					})),
				err: errors.Wrap(errors.New(errSecretKeyNotFound), errUpdateFailed),
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
					MockRemoveGroupVariable: func(gid interface{}, key string, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: variable(
					withGroupID(groupID),
					withConditions(xpv1.Available()),
				),
			},
			want: want{
				cr: variable(
					withGroupID(groupID),
					withConditions(xpv1.Deleting()),
				),
			},
		},
		"FailedDeletion": {
			args: args{
				variable: &fake.MockClient{
					MockRemoveGroupVariable: func(gid interface{}, key string, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: variable(
					withGroupID(groupID),
					withConditions(xpv1.Available()),
				),
			},
			want: want{
				cr: variable(
					withGroupID(groupID),
					withConditions(xpv1.Deleting()),
				),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
		"InvalidVariableID": {
			args: args{
				variable: &fake.MockClient{
					MockRemoveGroupVariable: func(gid interface{}, key string, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: variable(
					withGroupID(groupID),
					withConditions(xpv1.Available()),
				),
			},
			want: want{
				cr: variable(
					withGroupID(groupID),
					withConditions(xpv1.Deleting()),
				),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.variable}
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
