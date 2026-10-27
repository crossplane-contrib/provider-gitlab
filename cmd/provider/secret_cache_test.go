/*
Copyright 2026 The Crossplane Authors.

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

package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	apisNamespaced "github.com/crossplane-contrib/provider-gitlab/apis/namespaced"
	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
)

// secretNamespace is the only namespace allowing Secret writes in the
// fixture.
const secretNamespace = "team"

// secretAPI records real manager-client requests and denies all Secret
// list/watch
// requests. Its payloads are synthetic; no Kubernetes cluster is contacted.
type secretAPI struct {
	mu       sync.Mutex
	secrets  map[string]*corev1.Secret
	requests []string
}

// ServeHTTP implements only the namespaced Secret operations used by the
// test.
func (a *secretAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.requests = append(a.requests, r.Method+" "+r.URL.Path)

	w.Header().Set("Content-Type", "application/json")

	collection := "/api/v1/namespaces/" + secretNamespace + "/secrets"
	switch {
	case r.Method == http.MethodPost && r.URL.Path == collection:
		secret := new(corev1.Secret)

		err := json.NewDecoder(r.Body).Decode(secret)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		a.secrets[secret.Name] = secret

		w.WriteHeader(http.StatusCreated)
		writeSecretJSON(w, secret)
	case r.URL.Path == collection+"/credentials" || r.URL.Path == collection+"/analysis":
		a.serveNamedSecret(w, r)
	default:
		writeSecretStatus(w, http.StatusForbidden, metav1.StatusReasonForbidden)
	}
}

// serveNamedSecret supports the native publisher's GET, POST and merge
// PATCH flow.
func (a *secretAPI) serveNamedSecret(w http.ResponseWriter, r *http.Request) {
	name := "analysis"
	if r.URL.Path == "/api/v1/namespaces/"+secretNamespace+"/secrets/credentials" {
		name = "credentials"
	}

	secret, exists := a.secrets[name]
	if !exists {
		writeSecretStatus(w, http.StatusNotFound, metav1.StatusReasonNotFound)

		return
	}

	switch r.Method {
	case http.MethodGet:
		writeSecretJSON(w, secret)
	case http.MethodPatch:
		// The runtime publisher supplies the complete desired object as a merge
		// patch. Preserve the API object's metadata, including its controller.
		desired := new(corev1.Secret)

		err := json.NewDecoder(r.Body).Decode(desired)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		if desired.Data != nil {
			secret.Data = desired.Data
		}

		writeSecretJSON(w, secret)
	default:
		writeSecretStatus(w, http.StatusForbidden, metav1.StatusReasonForbidden)
	}
}

// writeSecretStatus returns a Kubernetes API error without exposing
// request data.
func writeSecretStatus(w http.ResponseWriter, code int, reason metav1.StatusReason) {
	w.WriteHeader(code)
	writeSecretJSON(w, &metav1.Status{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Status"},
		Status:   metav1.StatusFailure, Reason: reason, Code: int32(code),
		Message: "synthetic Secret API response",
	})
}

// newSecretTestManager constructs the real controller-runtime delegating
// client.
func newSecretTestManager(t *testing.T, api *secretAPI, cached bool) ctrl.Manager {
	t.Helper()

	server := httptest.NewServer(api)
	t.Cleanup(server.Close)

	scheme := runtime.NewScheme()

	err := corev1.AddToScheme(scheme)
	if err != nil {
		t.Fatal(err)
	}

	err = apisNamespaced.AddToScheme(scheme)
	if err != nil {
		t.Fatal(err)
	}

	mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{corev1.SchemeGroupVersion})
	mapper.Add(corev1.SchemeGroupVersion.WithKind("Secret"), meta.RESTScopeNamespace)
	mapper.Add(corev1.SchemeGroupVersion.WithKind("Pod"), meta.RESTScopeNamespace)

	mgr, err := ctrl.NewManager(&rest.Config{Host: server.URL, ContentConfig: rest.ContentConfig{ContentType: "application/json", AcceptContentTypes: "application/json"}}, ctrl.Options{
		Scheme:         scheme,
		MapperProvider: func(*rest.Config, *http.Client) (meta.RESTMapper, error) { return mapper, nil },
		Client:         secretCacheClientOptions(cached),
		Metrics:        metricsserver.Options{BindAddress: "0"},
	})
	if err != nil {
		t.Fatal(err)
	}

	return mgr
}

// TestSecretCacheEnabled preserves cached Secret reads and other resource
// caches.
func TestSecretCacheEnabled(t *testing.T) {
	t.Parallel()

	api := &secretAPI{}
	mgr := newSecretTestManager(t, api, true)

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()

	err := mgr.GetClient().Get(ctx, types.NamespacedName{Namespace: secretNamespace, Name: "credentials"}, &corev1.Secret{})
	if err == nil {
		t.Fatal("an unstarted Secret cache unexpectedly served a read")
	}

	if len(api.requests) != 0 {
		t.Fatal("cached read unexpectedly reached the API")
	}
}

// TestUncachedSecretLifecycle exercises credential rotation and the native
// crossplane-runtime AccessToken connection publisher with Secret
// list/watch denied.
func TestUncachedSecretLifecycle(t *testing.T) {
	t.Parallel()

	api := &secretAPI{secrets: map[string]*corev1.Secret{
		"credentials": {
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
			ObjectMeta: metav1.ObjectMeta{Namespace: secretNamespace, Name: "credentials"},
			Data:       map[string][]byte{"token": []byte("synthetic-initial")},
		},
	}}
	mgr := newSecretTestManager(t, api, false)

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	// Pods still use the cache; before it starts there must be no API request.
	podErr := mgr.GetClient().Get(ctx, types.NamespacedName{Namespace: secretNamespace, Name: "workload"}, &corev1.Pod{})
	if podErr == nil {
		t.Fatal("an unstarted Pod cache unexpectedly served a read")
	}

	if len(api.requests) != 0 {
		t.Fatal("Pod cache was bypassed")
	}

	done := make(chan error, 1)
	go func() { done <- mgr.GetCache().Start(ctx) }()

	t.Cleanup(func() {
		cancel()

		err := <-done
		if err != nil {
			t.Error(err)
		}
	})

	token := &v1alpha1.AccessToken{ObjectMeta: metav1.ObjectMeta{Namespace: secretNamespace, Name: "token", UID: "test-owner"}}
	token.SetWriteConnectionSecretToReference(&xpv1.LocalSecretReference{Name: "analysis"})
	selector := &xpv1.SecretKeySelector{SecretReference: xpv1.SecretReference{Namespace: secretNamespace, Name: "credentials"}, Key: "token"}

	for _, value := range []string{"synthetic-initial", "synthetic-rotated"} {
		api.mu.Lock()
		api.secrets["credentials"].Data["token"] = []byte(value)
		api.mu.Unlock()

		got, err := common.GetTokenValueFromSecret(ctx, mgr.GetClient(), token, selector)
		if err != nil {
			t.Fatal(err)
		}

		if got == nil || *got != value {
			t.Fatal("credential point-read did not observe the current value")
		}
	}

	publisher := managed.NewAPILocalSecretPublisher(mgr.GetClient(), mgr.GetScheme())

	for _, value := range []string{"synthetic-created-token", "synthetic-renewed-token"} {
		details := managed.ConnectionDetails{"token": []byte(value)}

		changed, err := publisher.PublishConnection(ctx, token, details)
		if err != nil || !changed {
			t.Fatalf("connection publication failed: changed=%v err=%v", changed, err)
		}

		got, err := common.GetTokenValueFromLocalSecret(ctx, mgr.GetClient(), token, &xpv1.LocalSecretKeySelector{LocalSecretReference: *token.GetWriteConnectionSecretToReference(), Key: "token"})
		if err != nil {
			t.Fatal(err)
		}

		if got == nil || *got != value {
			t.Fatal("connection Secret did not contain the current token")
		}

		changed, err = publisher.PublishConnection(ctx, token, details)
		if err != nil || changed {
			t.Fatalf("identical connection details were not a no-op: changed=%v err=%v", changed, err)
		}
	}

	// Observe returns no new token value. The native merge patch omits Data,
	// preserving the previously published value instead of clearing it.
	_, err := publisher.PublishConnection(ctx, token, managed.ConnectionDetails{})
	if err != nil {
		t.Fatal(err)
	}

	retained, err := common.GetTokenValueFromLocalSecret(ctx, mgr.GetClient(), token, &xpv1.LocalSecretKeySelector{LocalSecretReference: *token.GetWriteConnectionSecretToReference(), Key: "token"})
	if err != nil {
		t.Fatal(err)
	}

	if retained == nil || *retained != "synthetic-renewed-token" {
		t.Fatal("observation without new details erased the connection token")
	}

	err = publisher.UnpublishConnection(ctx, token, nil)
	if err != nil {
		t.Fatal(err)
	}

	deniedSelector := selector.DeepCopy()
	deniedSelector.Namespace = "other-team"

	_, err = common.GetTokenValueFromSecret(ctx, mgr.GetClient(), token, deniedSelector)
	if !apierrors.IsForbidden(err) {
		t.Fatalf("credential authorization denial was not preserved: %v", err)
	}

	deniedToken := token.DeepCopy()
	deniedToken.Namespace = "other-team"

	_, err = publisher.PublishConnection(ctx, deniedToken, managed.ConnectionDetails{"token": []byte("synthetic-denied")})
	if !apierrors.IsForbidden(err) {
		t.Fatalf("publisher authorization denial was not preserved: %v", err)
	}

	assertSecretRequests(t, api)
}

// assertSecretRequests rejects list/watch and checks create/patch were
// exercised.
func assertSecretRequests(t *testing.T, api *secretAPI) {
	t.Helper()
	api.mu.Lock()
	defer api.mu.Unlock()

	for _, request := range api.requests {
		switch request {
		case "GET /api/v1/pods", "GET /api/v1/namespaces/team/secrets/credentials", "GET /api/v1/namespaces/team/secrets/analysis", "POST /api/v1/namespaces/team/secrets", "PATCH /api/v1/namespaces/team/secrets/analysis", "GET /api/v1/namespaces/other-team/secrets/credentials", "GET /api/v1/namespaces/other-team/secrets/analysis":
		default:
			t.Fatalf("unexpected request (including Secret list/watch): %s", request)
		}
	}

	for _, request := range []string{"POST /api/v1/namespaces/team/secrets", "PATCH /api/v1/namespaces/team/secrets/analysis"} {
		if !slices.Contains(api.requests, request) {
			t.Fatalf("native publisher did not exercise %s", request)
		}
	}
}

// writeSecretJSON encodes synthetic responses and checks serialization
// errors.
func writeSecretJSON(w http.ResponseWriter, value any) {
	err := json.NewEncoder(w).Encode(value)
	if err != nil {
		http.Error(w, "cannot encode synthetic response", http.StatusInternalServerError)

		return
	}
}
