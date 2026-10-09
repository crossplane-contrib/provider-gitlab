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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Secret hash annotations let controllers detect changes to write-only secret
// values (passwords, webhooks, tokens) that the GitLab API never returns. A
// keyed HMAC of the value last applied to GitLab is stored on the managed
// resource, and compared with the HMAC of the value currently held in the
// secret. The HMAC key comes from the ProviderConfig secretHashKeySecretRef;
// without it nothing is stored and such changes are not detected.

const (
	// MinSecretHashKeyLength is the minimum length, in bytes, of the HMAC key.
	MinSecretHashKeyLength = 32

	secretHashAnnotationPrefix = "gitlab.crossplane.io/"
	secretHashAnnotationSuffix = "-secret-hash"

	ErrSecretHashKey         = "cannot resolve secret hash key from ProviderConfig secretHashKeySecretRef"
	ErrSecretHashKeyTooShort = "secret hash key referenced by ProviderConfig secretHashKeySecretRef must be at least %d bytes long"
	ErrPersistSecretHash     = "cannot persist secret hash annotation"
)

// SecretHashAnnotationKey returns the annotation key under which the hash of
// the secret identified by name is stored, e.g. "webhook" gives
// "gitlab.crossplane.io/webhook-secret-hash".
func SecretHashAnnotationKey(name string) string {
	return secretHashAnnotationPrefix + name + secretHashAnnotationSuffix
}

// A SecretHasher records and compares keyed hashes of write-only secret values
// in managed resource annotations. A nil or zero SecretHasher is disabled: it
// stores nothing and reports every value as up to date.
type SecretHasher struct {
	key []byte
}

// NewSecretHasher returns a SecretHasher keyed with the secret referenced by
// cfg.SecretHashKeySecretRef, or a disabled SecretHasher when cfg does not
// reference one. When mg is being deleted, which never needs the key, a key
// that cannot be resolved yields a disabled SecretHasher instead of an error,
// so that a misconfigured key cannot block deletion.
func NewSecretHasher(ctx context.Context, kube client.Client, mg resource.Managed, cfg *Config) (*SecretHasher, error) {
	if cfg == nil || cfg.SecretHashKeySecretRef == nil {
		return &SecretHasher{}, nil
	}
	h, err := resolveSecretHasher(ctx, kube, mg, cfg.SecretHashKeySecretRef)
	if err != nil && meta.WasDeleted(mg) {
		return &SecretHasher{}, nil
	}
	return h, err
}

func resolveSecretHasher(ctx context.Context, kube client.Client, mg resource.Managed, ref *v2.SecretKeySelector) (*SecretHasher, error) {
	key, err := GetTokenValueFromSecret(ctx, kube, mg, ref)
	if err != nil {
		return nil, errors.Wrap(err, ErrSecretHashKey)
	}
	return NewSecretHasherWithKey([]byte(*key))
}

// NewSecretHasherWithKey returns a SecretHasher keyed with key, which must be
// at least MinSecretHashKeyLength bytes long.
func NewSecretHasherWithKey(key []byte) (*SecretHasher, error) {
	if len(key) < MinSecretHashKeyLength {
		return nil, errors.Errorf(ErrSecretHashKeyTooShort, MinSecretHashKeyLength)
	}
	return &SecretHasher{key: key}, nil
}

// Enabled returns true if the SecretHasher has a key and thus records hashes.
func (h *SecretHasher) Enabled() bool {
	return h != nil && len(h.key) > 0
}

// Hash returns the hex encoded HMAC-SHA256 of value.
func (h *SecretHasher) Hash(value string) string {
	mac := hmac.New(sha256.New, h.key)
	mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

// IsUpToDate returns true if the hash annotation of the secret identified by
// name on o matches value. A missing annotation is reported as not up to date,
// so that the secret is (re)applied and its hash recorded. A disabled
// SecretHasher always returns true.
func (h *SecretHasher) IsUpToDate(o metav1.Object, name, value string) bool {
	if !h.Enabled() {
		return true
	}
	stored, ok := o.GetAnnotations()[SecretHashAnnotationKey(name)]
	return ok && hmac.Equal([]byte(stored), []byte(h.Hash(value)))
}

// Set records the hash of value for the secret identified by name on o, in
// memory only. It is suitable for use in Create, where the managed reconciler
// persists all annotations afterwards. It is a no-op when disabled.
func (h *SecretHasher) Set(o metav1.Object, name, value string) {
	if !h.Enabled() {
		return
	}
	meta.AddAnnotations(o, map[string]string{SecretHashAnnotationKey(name): h.Hash(value)})
}

// Persist records the hash of value for the secret identified by name on o
// and patches it to the API server. It must be used in Update, where the
// managed reconciler does not persist annotations. It is a no-op when disabled
// or when the stored hash is already up to date.
//
// The patch is sent from a copy of o, so that the API server response does not
// overwrite the in-memory status of o (e.g. observations made during Observe),
// which the managed reconciler persists after Update. Only the new
// resourceVersion is carried back to o so that the subsequent status update
// does not conflict.
func (h *SecretHasher) Persist(ctx context.Context, kube client.Client, o client.Object, name, value string) error {
	if !h.Enabled() || h.IsUpToDate(o, name, value) {
		return nil
	}
	p := o.DeepCopyObject().(client.Object)
	patch := client.MergeFrom(o.DeepCopyObject().(client.Object))
	h.Set(p, name, value)
	h.Set(o, name, value)
	if err := kube.Patch(ctx, p, patch); err != nil {
		return errors.Wrap(err, ErrPersistSecretHash)
	}
	o.SetResourceVersion(p.GetResourceVersion())
	return nil
}
