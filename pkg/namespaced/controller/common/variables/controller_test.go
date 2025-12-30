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

package variables_test

import (
	"context"
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	commonv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	instancev1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/instance/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/common/variables"
)

func TestUpdateVariableFromSecret(t *testing.T) {
	secretKey := "token"
	secretValue := "s3cr3t"

	// The helper resolves LocalSecretKeySelector by using the managed resource namespace.
	mg := &instancev1alpha1.Variable{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}}

	type args struct {
		kube     client.Client
		selector *xpv1.LocalSecretKeySelector
		params   *commonv1alpha1.CommonVariableParameters
	}

	cases := map[string]struct {
		args args
		want *commonv1alpha1.CommonVariableParameters
		err  error
	}{
		"SelectorNil": {
			args: args{
				kube:     &test.MockClient{},
				selector: nil,
				params:   &commonv1alpha1.CommonVariableParameters{},
			},
			err: errors.New(common.ErrSecretSelectorNil),
		},
		"SuccessfulSetsValueAndDefaults": {
			args: args{
				kube: &test.MockClient{MockGet: func(_ context.Context, _ client.ObjectKey, obj client.Object) error {
					secret, ok := obj.(*corev1.Secret)
					if !ok {
						return errors.Errorf("unexpected object type %T", obj)
					}
					secret.Data = map[string][]byte{secretKey: []byte(secretValue)}
					return nil
				}},
				selector: common.TestCreateLocalSecretKeySelector("ignored", secretKey),
				params:   &commonv1alpha1.CommonVariableParameters{},
			},
			want: &commonv1alpha1.CommonVariableParameters{
				Value:  &secretValue,
				Masked: gitlab.Ptr(true),
				Raw:    gitlab.Ptr(true),
			},
		},
		"DoesNotOverrideExplicitMaskedRaw": {
			args: args{
				kube: &test.MockClient{MockGet: func(_ context.Context, _ client.ObjectKey, obj client.Object) error {
					secret, ok := obj.(*corev1.Secret)
					if !ok {
						return errors.Errorf("unexpected object type %T", obj)
					}
					secret.Data = map[string][]byte{secretKey: []byte(secretValue)}
					return nil
				}},
				selector: common.TestCreateLocalSecretKeySelector("ignored", secretKey),
				params: &commonv1alpha1.CommonVariableParameters{
					Masked: gitlab.Ptr(false),
					Raw:    gitlab.Ptr(false),
				},
			},
			want: &commonv1alpha1.CommonVariableParameters{
				Value:  &secretValue,
				Masked: gitlab.Ptr(false),
				Raw:    gitlab.Ptr(false),
			},
		},
		"WrongKey": {
			args: args{
				kube: &test.MockClient{MockGet: func(_ context.Context, _ client.ObjectKey, obj client.Object) error {
					secret, ok := obj.(*corev1.Secret)
					if !ok {
						return errors.Errorf("unexpected object type %T", obj)
					}
					secret.Data = map[string][]byte{"other": []byte(secretValue)}
					return nil
				}},
				selector: common.TestCreateLocalSecretKeySelector("ignored", secretKey),
				params:   &commonv1alpha1.CommonVariableParameters{},
			},
			err: errors.New(common.ErrSecretKeyNotFound),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := variables.UpdateVariableFromSecret(tc.args.kube, mg, context.Background(), tc.args.selector, tc.args.params)
			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Fatalf("UpdateVariableFromSecret(...): -want err, +got err:\n%s", diff)
			}

			if tc.want == nil {
				return
			}

			// Compare only the fields this helper is responsible for.
			cmpOpts := []cmp.Option{
				cmpopts.IgnoreFields(commonv1alpha1.CommonVariableParameters{}, "Key", "Protected", "VariableType", "Description"),
			}
			if diff := cmp.Diff(tc.want, tc.args.params, cmpOpts...); diff != "" {
				t.Errorf("UpdateVariableFromSecret(...): -want params, +got params:\n%s", diff)
			}
		})
	}
}
