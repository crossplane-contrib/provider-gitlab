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
	"fmt"
	"testing"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	namespacedgroupsv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
)

const (
	testSecretHashName  = "webhook"
	testSecretHashValue = "https://mattermost.example.com/hooks/abc123"
	testSecretHashKey   = "0123456789abcdef0123456789abcdef"
)

func secretHashObject(annotations map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Annotations: annotations}}
}

func testSecretHasher(key string) *SecretHasher {
	return &SecretHasher{key: []byte(key)}
}

func TestNewSecretHasherWithKey(t *testing.T) {
	if _, err := NewSecretHasherWithKey([]byte("too-short")); err == nil {
		t.Error("NewSecretHasherWithKey(): expected an error for a key shorter than MinSecretHashKeyLength")
	}
	h, err := NewSecretHasherWithKey([]byte(testSecretHashKey))
	if err != nil {
		t.Fatalf("NewSecretHasherWithKey(): unexpected error: %v", err)
	}
	if !h.Enabled() {
		t.Error("NewSecretHasherWithKey(): expected an enabled hasher")
	}
}

func TestSecretHashAnnotationKey(t *testing.T) {
	if got, want := SecretHashAnnotationKey("webhook"), "gitlab.crossplane.io/webhook-secret-hash"; got != want {
		t.Errorf("SecretHashAnnotationKey(): got %q, want %q", got, want)
	}
}

func TestNewSecretHasher(t *testing.T) {
	errBoom := errors.New("boom")
	ref := &v2.SecretKeySelector{
		Key:             "key",
		SecretReference: v2.SecretReference{Name: "hash-key", Namespace: "crossplane-system"},
	}
	kubeWithKey := func(key string) client.Client {
		return &test.MockClient{
			MockGet: func(_ context.Context, k client.ObjectKey, obj client.Object) error {
				if k.Name != ref.Name || k.Namespace != ref.Namespace {
					return errors.Errorf("unexpected secret %s", k)
				}
				obj.(*corev1.Secret).Data = map[string][]byte{"key": []byte(key)}
				return nil
			},
		}
	}

	deleting := &namespacedgroupsv1alpha1.IntegrationMattermost{}
	deleting.SetDeletionTimestamp(&metav1.Time{Time: time.Now()})

	cases := map[string]struct {
		kube        client.Client
		mg          resource.Managed
		cfg         *Config
		wantEnabled bool
		wantErr     error
	}{
		"NilConfig": {
			cfg:         nil,
			wantEnabled: false,
		},
		"NoKeyReferenced": {
			cfg:         &Config{},
			wantEnabled: false,
		},
		"KeyResolved": {
			kube:        kubeWithKey(testSecretHashKey),
			cfg:         &Config{SecretHashKeySecretRef: ref},
			wantEnabled: true,
		},
		"KeyTooShort": {
			kube:    kubeWithKey("too-short"),
			cfg:     &Config{SecretHashKeySecretRef: ref},
			wantErr: fmt.Errorf(ErrSecretHashKeyTooShort, MinSecretHashKeyLength),
		},
		"KeySecretMissing": {
			kube:    &test.MockClient{MockGet: test.NewMockGetFn(errBoom)},
			cfg:     &Config{SecretHashKeySecretRef: ref},
			wantErr: errors.Wrap(errors.Wrap(errBoom, ErrSecretNotFound), ErrSecretHashKey),
		},
		"KeySecretMissingWhileDeleting": {
			kube:        &test.MockClient{MockGet: test.NewMockGetFn(errBoom)},
			mg:          deleting,
			cfg:         &Config{SecretHashKeySecretRef: ref},
			wantEnabled: false,
		},
		"KeyTooShortWhileDeleting": {
			kube:        kubeWithKey("too-short"),
			mg:          deleting,
			cfg:         &Config{SecretHashKeySecretRef: ref},
			wantEnabled: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mg := tc.mg
			if mg == nil {
				mg = &namespacedgroupsv1alpha1.IntegrationMattermost{}
			}
			h, err := NewSecretHasher(context.Background(), tc.kube, mg, tc.cfg)
			if (tc.wantErr == nil) != (err == nil) || (err != nil && err.Error() != tc.wantErr.Error()) {
				t.Fatalf("NewSecretHasher(): want error %v, got %v", tc.wantErr, err)
			}
			if err == nil && h.Enabled() != tc.wantEnabled {
				t.Errorf("NewSecretHasher(): Enabled() = %v, want %v", h.Enabled(), tc.wantEnabled)
			}
		})
	}
}

func TestSecretHasherHash(t *testing.T) {
	h := testSecretHasher(testSecretHashKey)

	first, second := h.Hash(testSecretHashValue), h.Hash(testSecretHashValue)
	if first != second {
		t.Error("Hash(): expected identical hashes for identical key and value")
	}
	if h.Hash(testSecretHashValue) == h.Hash("other") {
		t.Error("Hash(): expected different hashes for different values")
	}
	if h.Hash(testSecretHashValue) == testSecretHasher("fedcba9876543210fedcba9876543210").Hash(testSecretHashValue) {
		t.Error("Hash(): expected different hashes for different keys")
	}
	// Known HMAC-SHA256 test vector (RFC 4231, test case 2).
	if got, want := testSecretHasher("Jefe").Hash("what do ya want for nothing?"), "5bdcc146bf60754e6a042426089575c75a003f089d2739839dec58b964ec3843"; got != want {
		t.Errorf("Hash(): got %q, want HMAC-SHA256 %q", got, want)
	}
}

func TestSecretHasherIsUpToDate(t *testing.T) {
	h := testSecretHasher(testSecretHashKey)
	key := SecretHashAnnotationKey(testSecretHashName)
	hash := h.Hash(testSecretHashValue)

	cases := map[string]struct {
		h     *SecretHasher
		o     metav1.Object
		value string
		want  bool
	}{
		"Matching": {
			h:     h,
			o:     secretHashObject(map[string]string{key: hash}),
			value: testSecretHashValue,
			want:  true,
		},
		"ValueChanged": {
			h:     h,
			o:     secretHashObject(map[string]string{key: hash}),
			value: "https://mattermost.example.com/hooks/rotated",
			want:  false,
		},
		"KeyRotated": {
			h:     testSecretHasher("fedcba9876543210fedcba9876543210"),
			o:     secretHashObject(map[string]string{key: hash}),
			value: testSecretHashValue,
			want:  false,
		},
		"AnnotationMissing": {
			h:     h,
			o:     secretHashObject(nil),
			value: testSecretHashValue,
			want:  false,
		},
		"OtherSecretAnnotated": {
			h:     h,
			o:     secretHashObject(map[string]string{SecretHashAnnotationKey("password"): hash}),
			value: testSecretHashValue,
			want:  false,
		},
		"DisabledAlwaysUpToDate": {
			h:     &SecretHasher{},
			o:     secretHashObject(nil),
			value: testSecretHashValue,
			want:  true,
		},
		"NilAlwaysUpToDate": {
			h:     nil,
			o:     secretHashObject(nil),
			value: testSecretHashValue,
			want:  true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := tc.h.IsUpToDate(tc.o, testSecretHashName, tc.value); got != tc.want {
				t.Errorf("IsUpToDate(): got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSecretHasherSet(t *testing.T) {
	cases := map[string]struct {
		h    *SecretHasher
		want map[string]string
	}{
		"Enabled": {
			h: testSecretHasher(testSecretHashKey),
			want: map[string]string{
				"keep": "me",
				SecretHashAnnotationKey(testSecretHashName): testSecretHasher(testSecretHashKey).Hash(testSecretHashValue),
			},
		},
		"Disabled": {
			h:    &SecretHasher{},
			want: map[string]string{"keep": "me"},
		},
		"Nil": {
			h:    nil,
			want: map[string]string{"keep": "me"},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			o := secretHashObject(map[string]string{"keep": "me"})
			tc.h.Set(o, testSecretHashName, testSecretHashValue)
			if diff := cmp.Diff(tc.want, o.GetAnnotations()); diff != "" {
				t.Errorf("Set(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestSecretHasherPersist(t *testing.T) {
	errBoom := errors.New("boom")
	h := testSecretHasher(testSecretHashKey)
	upToDate := secretHashObject(nil)
	h.Set(upToDate, testSecretHashName, testSecretHashValue)

	cases := map[string]struct {
		h            *SecretHasher
		o            *corev1.ConfigMap
		patchErr     error
		wantErr      error
		wantPatched  bool
		wantUpToDate bool
	}{
		"Patched": {
			h:            h,
			o:            secretHashObject(nil),
			wantPatched:  true,
			wantUpToDate: true,
		},
		"AlreadyUpToDate": {
			h:            h,
			o:            upToDate,
			wantPatched:  false,
			wantUpToDate: true,
		},
		"PatchFailed": {
			h:            h,
			o:            secretHashObject(nil),
			patchErr:     errBoom,
			wantErr:      errors.Wrap(errBoom, ErrPersistSecretHash),
			wantPatched:  true,
			wantUpToDate: true,
		},
		"Disabled": {
			h:           &SecretHasher{},
			o:           secretHashObject(nil),
			wantPatched: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			patched := false
			kube := &test.MockClient{
				MockPatch: func(_ context.Context, _ client.Object, p client.Patch, _ ...client.PatchOption) error {
					patched = true
					if p.Type() != types.MergePatchType {
						t.Errorf("Persist(): expected a merge patch, got %s", p.Type())
					}
					return tc.patchErr
				},
			}

			err := tc.h.Persist(context.Background(), kube, tc.o, testSecretHashName, testSecretHashValue)
			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("Persist(): -want error, +got error:\n%s", diff)
			}
			if patched != tc.wantPatched {
				t.Errorf("Persist(): patched = %v, want %v", patched, tc.wantPatched)
			}
			if tc.wantUpToDate && !tc.h.IsUpToDate(tc.o, testSecretHashName, testSecretHashValue) {
				t.Error("Persist(): expected in-memory hash to be up to date")
			}
			if !tc.h.Enabled() && len(tc.o.GetAnnotations()) != 0 {
				t.Errorf("Persist(): expected no annotation when disabled, got %v", tc.o.GetAnnotations())
			}
		})
	}
}
