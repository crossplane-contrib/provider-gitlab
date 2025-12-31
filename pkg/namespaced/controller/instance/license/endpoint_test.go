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

package license

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/instance/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
)

const connDetailsFmt = "connection details: -want, +got:\n%s"

func stringPtr(s string) *string { return &s }

func newScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	if err := corev1.AddToScheme(s); err != nil {
		t.Fatalf("AddToScheme(corev1): %v", err)
	}
	if err := v1alpha1.SchemeBuilder.AddToScheme(s); err != nil {
		t.Fatalf("AddToScheme(v1alpha1): %v", err)
	}
	return s
}

func newKube(t *testing.T, objs ...client.Object) client.Client {
	t.Helper()
	return fake.NewClientBuilder().WithScheme(newScheme(t)).WithObjects(objs...).Build()
}

func licenseCR() *v1alpha1.License {
	return &v1alpha1.License{ObjectMeta: metav1.ObjectMeta{Name: "l", Namespace: "default"}}
}

func TestGetLicenseFromSecrets(t *testing.T) {
	licenseValue := "my-license"
	licenseKey := "license"

	basicUser := "user"
	basicPass := "pass"
	token := "token"

	t.Run("FromEndpointURLWithTokenInSpec", func(t *testing.T) {
		// Server validates that token auth is used (Bearer)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := r.Header.Get("Authorization"); got != "Bearer "+token {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_, _ = w.Write([]byte(licenseValue))
		}))
		defer srv.Close()

		params := &v1alpha1.LicenseParameters{
			LicenseEndpointURL:      stringPtr(srv.URL),
			LicenseEndpointToken:    &token,
			LicenseEndpointUsername: &basicUser,
			LicenseEndpointPassword: &basicPass,
		}

		e := &external{kube: newKube(t)}
		cd, err := e.getLicenseFromSecrets(licenseCR(), context.Background(), params)

		want := managed.ConnectionDetails{keyLicense: []byte(licenseValue)}
		if diff := cmp.Diff(want, cd); diff != "" {
			t.Fatalf("connection details: -want, +got:\n%s", diff)
		}
		if diff := cmp.Diff(error(nil), err, test.EquateErrors()); diff != "" {
			t.Fatalf("error: -want, +got:\n%s", diff)
		}
	})

	t.Run("FromEndpointURLSecretRefWithBasicAuthSecrets", func(t *testing.T) {
		encoded := base64.StdEncoding.EncodeToString([]byte(basicUser + ":" + basicPass))
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := r.Header.Get("Authorization"); got != "Basic "+encoded {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_, _ = w.Write([]byte("  " + licenseValue + "  "))
		}))
		defer srv.Close()

		urlSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "url", Namespace: "default"}, Data: map[string][]byte{"url": []byte(srv.URL)}}
		userSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "user", Namespace: "default"}, Data: map[string][]byte{"u": []byte(basicUser)}}
		passSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "pass", Namespace: "default"}, Data: map[string][]byte{"p": []byte(basicPass)}}

		params := &v1alpha1.LicenseParameters{
			LicenseEndpointURLSecretRef:      common.TestCreateLocalSecretKeySelector("url", "url"),
			LicenseEndpointUsernameSecretRef: common.TestCreateLocalSecretKeySelector("user", "u"),
			LicenseEndpointPasswordSecretRef: common.TestCreateLocalSecretKeySelector("pass", "p"),
		}

		e := &external{kube: newKube(t, urlSecret, userSecret, passSecret)}
		cd, err := e.getLicenseFromSecrets(licenseCR(), context.Background(), params)

		want := managed.ConnectionDetails{keyLicense: []byte(licenseValue)}
		if diff := cmp.Diff(want, cd); diff != "" {
			t.Fatalf(connDetailsFmt, diff)
		}
		if diff := cmp.Diff(error(nil), err, test.EquateErrors()); diff != "" {
			t.Fatalf(errFmt, diff)
		}
	})

	cases := map[string]struct {
		kube    client.Client
		mg      *v1alpha1.License
		params  *v1alpha1.LicenseParameters
		wantCD  managed.ConnectionDetails
		wantErr error
	}{
		"FromLicenseSecretRef": {
			kube: newKube(t, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "src", Namespace: "default"}, Data: map[string][]byte{licenseKey: []byte(licenseValue)}}),
			mg:   licenseCR(),
			params: &v1alpha1.LicenseParameters{
				LicenseSecretRef: common.TestCreateLocalSecretKeySelector("src", licenseKey),
			},
			wantCD: managed.ConnectionDetails{keyLicense: []byte(licenseValue)},
		},
		"FromLicenseInSpec": {
			kube:   newKube(t),
			mg:     licenseCR(),
			params: &v1alpha1.LicenseParameters{License: &licenseValue},
			wantCD: managed.ConnectionDetails{keyLicense: []byte(licenseValue)},
		},
		"NoSourceProvided": {
			kube:    newKube(t),
			mg:      licenseCR(),
			params:  &v1alpha1.LicenseParameters{},
			wantErr: errors.New("no license source provided; please specify either LicenseEndpointURL, LicenseEndpointURLSecretRef, LicenseSecretRef, or License in the spec"),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube}
			cd, err := e.getLicenseFromSecrets(tc.mg, context.Background(), tc.params)
			if diff := cmp.Diff(tc.wantCD, cd); diff != "" {
				t.Errorf(connDetailsFmt, diff)
			}
			if tc.wantErr != nil {
				if err == nil {
					t.Errorf("expected error %q, got nil", tc.wantErr.Error())
				} else if err.Error() != tc.wantErr.Error() {
					t.Errorf("expected error %q, got %q", tc.wantErr.Error(), err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGetEndpointURL(t *testing.T) {
	e := &external{kube: newKube(t)}
	mg := licenseCR()
	params := &v1alpha1.LicenseParameters{LicenseEndpointURL: stringPtr("http://example")}

	_, err := e.getEndpointURL(mg, context.Background(), params, nil)
	if err == nil {
		t.Fatalf("expected error when connectionDetails is nil")
	}
}

func TestIsErrorFetchingLicenseFromEndpoint(t *testing.T) {
	if isErrorFetchingLicenseFromEndpoint(nil) {
		t.Fatalf("expected false for nil error")
	}
	if !isErrorFetchingLicenseFromEndpoint(errors.New(errFetchFromEndpoint)) {
		t.Fatalf("expected true for endpoint fetch error")
	}
	if isErrorFetchingLicenseFromEndpoint(errors.New("something else")) {
		t.Fatalf("expected false for other errors")
	}
}
