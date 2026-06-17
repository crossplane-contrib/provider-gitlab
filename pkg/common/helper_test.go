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

package common

import (
	"context"
	"testing"

	clustergroupsv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/cluster/groups/v1alpha1"
	legacyv1beta1 "github.com/crossplane-contrib/provider-gitlab/apis/cluster/v1beta1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type mockManagedResource struct {
	resource.Managed
	namespace string
}

func (m *mockManagedResource) GetNamespace() string {
	return m.namespace
}

func TestGetTokenValueFromSecret(t *testing.T) {
	testToken := "test-token-value"
	testNamespace := "test-namespace"
	testSecretName := "test-secret"
	testKey := "token"

	type args struct {
		selector *v2.SecretKeySelector
		kube     client.Client
	}
	type want struct {
		token *string
		err   error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"SuccessfulTokenRetrieval": {
			args: args{
				selector: &v2.SecretKeySelector{
					Key: testKey,
					SecretReference: v2.SecretReference{
						Name:      testSecretName,
						Namespace: testNamespace,
					},
				},
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						*obj.(*corev1.Secret) = corev1.Secret{
							Data: map[string][]byte{
								testKey: []byte(testToken),
							},
						}
						return nil
					}),
				},
			},
			want: want{
				token: &testToken,
				err:   nil,
			},
		},
		"SecretNotFound": {
			args: args{
				selector: &v2.SecretKeySelector{
					Key: testKey,
					SecretReference: v2.SecretReference{
						Name:      testSecretName,
						Namespace: testNamespace,
					},
				},
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(errors.New("secret not found")),
				},
			},
			want: want{
				token: nil,
				err:   errors.Wrap(errors.New("secret not found"), ErrSecretNotFound),
			},
		},
		"KeyNotFoundInSecret": {
			args: args{
				selector: &v2.SecretKeySelector{
					Key: "wrong-key",
					SecretReference: v2.SecretReference{
						Name:      testSecretName,
						Namespace: testNamespace,
					},
				},
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						*obj.(*corev1.Secret) = corev1.Secret{
							Data: map[string][]byte{
								testKey: []byte(testToken), // Different key than requested
							},
						}
						return nil
					}),
				},
			},
			want: want{
				token: nil,
				err:   errors.New(ErrSecretKeyNotFound),
			},
		},
		"EmptySecretData": {
			args: args{
				selector: &v2.SecretKeySelector{
					Key: testKey,
					SecretReference: v2.SecretReference{
						Name:      testSecretName,
						Namespace: testNamespace,
					},
				},
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						*obj.(*corev1.Secret) = corev1.Secret{
							Data: map[string][]byte{}, // Empty data
						}
						return nil
					}),
				},
			},
			want: want{
				token: nil,
				err:   errors.New(ErrSecretKeyNotFound),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mg := &mockManagedResource{namespace: testNamespace}

			got, err := GetTokenValueFromSecret(context.Background(), tc.args.kube, mg, tc.args.selector)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("GetTokenValueFromSecret() error = %v, want %v\ndiff: %s", err, tc.want.err, diff)
			}

			if diff := cmp.Diff(tc.want.token, got); diff != "" {
				t.Errorf("GetTokenValueFromSecret() token = %v, want %v\ndiff: %s", got, tc.want.token, diff)
			}
		})
	}
}

func TestUseLegacyProviderConfigSetsCredentialsSecretRef(t *testing.T) {
	testToken := "test-token-value"
	selector := &v2.SecretKeySelector{
		Key: "token",
		SecretReference: v2.SecretReference{
			Name:      "test-secret",
			Namespace: "default",
		},
	}

	mg := &clustergroupsv1alpha1.Group{}
	mg.SetProviderConfigReference(&v2.Reference{Name: "test-provider-config"})

	kube := &test.MockClient{
		MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
			switch o := obj.(type) {
			case *legacyv1beta1.ProviderConfig:
				*o = legacyv1beta1.ProviderConfig{
					ObjectMeta: metav1.ObjectMeta{Name: "test-provider-config"},
					Spec: legacyv1beta1.ProviderConfigSpec{
						Credentials: legacyv1beta1.ProviderCredentials{
							Source: v2.CredentialsSourceSecret,
							CommonCredentialSelectors: v2.CommonCredentialSelectors{
								SecretRef: selector,
							},
						},
					},
				}
			case *corev1.Secret:
				*o = corev1.Secret{Data: map[string][]byte{"token": []byte(testToken)}}
			}
			return nil
		}),
		MockCreate: test.NewMockCreateFn(nil),
	}

	cfg, err := UseLegacyProviderConfig(context.Background(), kube, mg)
	if err != nil {
		t.Fatalf("UseLegacyProviderConfig() error = %v", err)
	}

	if cfg == nil {
		t.Fatal("UseLegacyProviderConfig() returned nil config")
	}

	if diff := cmp.Diff(selector, cfg.CredentialsSecretRef); diff != "" {
		t.Fatalf("CredentialsSecretRef mismatch (-want +got):\n%s", diff)
	}
}

func TestGetTokenValueFromLocalSecret(t *testing.T) {
	testToken := "test-local-token-value"
	testNamespace := "test-namespace"
	testSecretName := "test-secret"
	testKey := "token"

	type args struct {
		selector *v2.LocalSecretKeySelector
		kube     client.Client
	}
	type want struct {
		token *string
		err   error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"SuccessfulLocalTokenRetrieval": {
			args: args{
				selector: &v2.LocalSecretKeySelector{
					Key: testKey,
					LocalSecretReference: v2.LocalSecretReference{
						Name: testSecretName,
					},
				},
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						*obj.(*corev1.Secret) = corev1.Secret{
							Data: map[string][]byte{
								testKey: []byte(testToken),
							},
						}
						return nil
					}),
				},
			},
			want: want{
				token: &testToken,
				err:   nil,
			},
		},
		"LocalSecretNotFound": {
			args: args{
				selector: &v2.LocalSecretKeySelector{
					Key: testKey,
					LocalSecretReference: v2.LocalSecretReference{
						Name: testSecretName,
					},
				},
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(errors.New("local secret not found")),
				},
			},
			want: want{
				token: nil,
				err:   errors.Wrap(errors.New("local secret not found"), ErrSecretNotFound),
			},
		},
		"LocalKeyNotFound": {
			args: args{
				selector: &v2.LocalSecretKeySelector{
					Key: "wrong-key",
					LocalSecretReference: v2.LocalSecretReference{
						Name: testSecretName,
					},
				},
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						*obj.(*corev1.Secret) = corev1.Secret{
							Data: map[string][]byte{
								testKey: []byte(testToken), // Different key than requested
							},
						}
						return nil
					}),
				},
			},
			want: want{
				token: nil,
				err:   errors.New(ErrSecretKeyNotFound),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mg := &mockManagedResource{namespace: testNamespace}

			got, err := GetTokenValueFromLocalSecret(context.Background(), tc.args.kube, mg, tc.args.selector)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("GetTokenValueFromLocalSecret() error = %v, want %v\ndiff: %s", err, tc.want.err, diff)
			}

			if diff := cmp.Diff(tc.want.token, got); diff != "" {
				t.Errorf("GetTokenValueFromLocalSecret() token = %v, want %v\ndiff: %s", got, tc.want.token, diff)
			}
		})
	}
}

func TestResolvePublicJobsSetting(t *testing.T) {
	trueVal := true
	falseVal := false

	type args struct {
		publicBuilds *bool
		publicJobs   *bool
	}
	type want struct {
		value           *bool
		usingDeprecated bool
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"BothSet_PublicJobsWins": {
			args: args{
				publicBuilds: &trueVal,
				publicJobs:   &falseVal,
			},
			want: want{
				value:           &falseVal,
				usingDeprecated: false,
			},
		},
		"OnlyPublicBuilds_True": {
			args: args{
				publicBuilds: &trueVal,
				publicJobs:   nil,
			},
			want: want{
				value:           &trueVal,
				usingDeprecated: true,
			},
		},
		"OnlyPublicBuilds_False": {
			args: args{
				publicBuilds: &falseVal,
				publicJobs:   nil,
			},
			want: want{
				value:           &falseVal,
				usingDeprecated: true,
			},
		},
		"OnlyPublicJobs_True": {
			args: args{
				publicBuilds: nil,
				publicJobs:   &trueVal,
			},
			want: want{
				value:           &trueVal,
				usingDeprecated: false,
			},
		},
		"OnlyPublicJobs_False": {
			args: args{
				publicBuilds: nil,
				publicJobs:   &falseVal,
			},
			want: want{
				value:           &falseVal,
				usingDeprecated: false,
			},
		},
		"NeitherSet": {
			args: args{
				publicBuilds: nil,
				publicJobs:   nil,
			},
			want: want{
				value:           nil,
				usingDeprecated: false,
			},
		},
		"BothTrue_PublicJobsWins": {
			args: args{
				publicBuilds: &trueVal,
				publicJobs:   &trueVal,
			},
			want: want{
				value:           &trueVal,
				usingDeprecated: false,
			},
		},
		"BothFalse_PublicJobsWins": {
			args: args{
				publicBuilds: &falseVal,
				publicJobs:   &falseVal,
			},
			want: want{
				value:           &falseVal,
				usingDeprecated: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			gotValue, gotDeprecated := ResolvePublicJobsSetting(tc.args.publicBuilds, tc.args.publicJobs)

			if diff := cmp.Diff(tc.want.value, gotValue); diff != "" {
				t.Errorf("ResolvePublicJobsSetting() value diff (-want +got):\n%s", diff)
			}

			if gotDeprecated != tc.want.usingDeprecated {
				t.Errorf("ResolvePublicJobsSetting() usingDeprecated = %v, want %v", gotDeprecated, tc.want.usingDeprecated)
			}
		})
	}
}
