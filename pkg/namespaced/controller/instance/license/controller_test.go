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
	"net/http"
	"net/http/httptest"
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xperrors "github.com/crossplane/crossplane-runtime/v2/pkg/errors"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/instance/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/instance"
)

const (
	errFmt = "error: -want, +got:\n%s"
	crFmt  = "cr: -want, +got:\n%s"
	outFmt = "out: -want, +got:\n%s"

	testLicenseKey = "my-license"
	testSecretName = "conn-secret"
	errNotSecret   = "object is not a secret"
)

var (
	unexpectedItem resource.Managed
	errBoom        = errors.New("boom")
)

type mockClient struct {
	MockGetLicense    func(options ...gitlab.RequestOptionFunc) (*gitlab.License, *gitlab.Response, error)
	MockAddLicense    func(opt *gitlab.AddLicenseOptions, options ...gitlab.RequestOptionFunc) (*gitlab.License, *gitlab.Response, error)
	MockDeleteLicense func(licenseID int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

func (m *mockClient) GetLicense(options ...gitlab.RequestOptionFunc) (*gitlab.License, *gitlab.Response, error) {
	return m.MockGetLicense(options...)
}

func (m *mockClient) AddLicense(opt *gitlab.AddLicenseOptions, options ...gitlab.RequestOptionFunc) (*gitlab.License, *gitlab.Response, error) {
	return m.MockAddLicense(opt, options...)
}

func (m *mockClient) DeleteLicense(licenseID int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return m.MockDeleteLicense(licenseID, options...)
}

type args struct {
	client instance.LicenseClient
	kube   client.Client
	cr     resource.Managed
}

type modifier func(*v1alpha1.License)

func withNamespace(ns string) modifier {
	return func(cr *v1alpha1.License) { cr.Namespace = ns }
}

func withWriteConnectionSecretRef(name string) modifier {
	return func(cr *v1alpha1.License) {
		cr.Spec.WriteConnectionSecretToReference = common.TestCreateLocalSecretReference(name)
	}
}

func withExternalName(n string) modifier {
	return func(cr *v1alpha1.License) { meta.SetExternalName(cr, n) }
}

func withConditions(c ...xpv1.Condition) modifier {
	return func(cr *v1alpha1.License) { cr.Status.SetConditions(c...) }
}

func withSpec(p v1alpha1.LicenseParameters) modifier {
	return func(cr *v1alpha1.License) { cr.Spec.ForProvider = p }
}

func withAtProvider(o v1alpha1.LicenseObservation) modifier {
	return func(cr *v1alpha1.License) { cr.Status.AtProvider = o }
}

func withDeletionTimestamp(t metav1.Time) modifier {
	return func(cr *v1alpha1.License) { cr.DeletionTimestamp = &t }
}

func license(m ...modifier) *v1alpha1.License {
	cr := &v1alpha1.License{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func kubeClient(t *testing.T, objs ...client.Object) client.Client {
	t.Helper()
	s := runtime.NewScheme()
	_ = corev1.AddToScheme(s)
	_ = v1alpha1.SchemeBuilder.AddToScheme(s)
	return fake.NewClientBuilder().WithScheme(s).WithObjects(objs...).Build()
}

func mockSecret(name, namespace string, data map[string][]byte) func(client.Object) error {
	return func(obj client.Object) error {
		s, ok := obj.(*corev1.Secret)
		if !ok {
			return errors.New(errNotSecret)
		}
		s.Name = name
		s.Namespace = namespace
		s.Data = data
		return nil
	}
}

func TestObserve(t *testing.T) {
	now := metav1.Now()
	type want struct {
		cr  resource.Managed
		obs managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"InvalidInput": {
			args: args{cr: unexpectedItem},
			want: want{cr: unexpectedItem, err: errors.New(errNotLicense)},
		},
		"NoExternalName": {
			args: args{cr: license(withNamespace("default"))},
			want: want{cr: license(withNamespace("default")), obs: managed.ExternalObservation{ResourceExists: false}},
		},
		"ErrGet": {
			args: args{
				client: &mockClient{MockGetLicense: func(options ...gitlab.RequestOptionFunc) (*gitlab.License, *gitlab.Response, error) {
					return nil, &gitlab.Response{Response: &http.Response{StatusCode: 500}}, errBoom
				}},
				cr: license(withNamespace("default"), withExternalName("1")),
			},
			want: want{cr: license(withNamespace("default"), withExternalName("1")), err: errors.Wrap(errBoom, errGetFailed)},
		},
		"NotFound": {
			args: args{
				client: &mockClient{MockGetLicense: func(options ...gitlab.RequestOptionFunc) (*gitlab.License, *gitlab.Response, error) {
					return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errors.New("not found")
				}},
				cr: license(withNamespace("default"), withExternalName("1")),
			},
			want: want{cr: license(withNamespace("default"), withExternalName("1")), obs: managed.ExternalObservation{}},
		},
		"NotIDExternalName": {
			args: args{
				client: &mockClient{MockGetLicense: func(options ...gitlab.RequestOptionFunc) (*gitlab.License, *gitlab.Response, error) {
					return &gitlab.License{ID: 2, Expired: false}, &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
				}},
				cr: license(withNamespace("default"), withExternalName("1")),
			},
			want: want{cr: license(withNamespace("default"), withExternalName("1")), obs: managed.ExternalObservation{}},
		},
		"SuccessfulAvailable": {
			args: args{
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, mockSecret(testSecretName, "default", map[string][]byte{
						"license": []byte(testLicenseKey),
					})),
				},
				client: &mockClient{MockGetLicense: func(options ...gitlab.RequestOptionFunc) (*gitlab.License, *gitlab.Response, error) {
					return &gitlab.License{ID: 1, Expired: false}, &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
				}},
				cr: license(
					withNamespace("default"),
					withExternalName("1"),
					withSpec(v1alpha1.LicenseParameters{License: strPtr(testLicenseKey)}),
					withWriteConnectionSecretRef(testSecretName),
				),
			},
			want: want{
				cr: license(
					withNamespace("default"),
					withExternalName("1"),
					withSpec(v1alpha1.LicenseParameters{License: strPtr(testLicenseKey)}),
					withWriteConnectionSecretRef(testSecretName),
					withConditions(xpv1.Available()),
					withAtProvider(instance.GenerateLicenseObservation(&gitlab.License{ID: 1, Expired: false})),
				),
				obs: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true, ResourceLateInitialized: false},
			},
		},
		"LicenseKeyMismatch": {
			args: args{
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, mockSecret(testSecretName, "default", map[string][]byte{
						"license": []byte("old-license"),
					})),
				},
				client: &mockClient{MockGetLicense: func(options ...gitlab.RequestOptionFunc) (*gitlab.License, *gitlab.Response, error) {
					return &gitlab.License{ID: 1, Expired: false}, &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
				}},
				cr: license(
					withNamespace("default"),
					withExternalName("1"),
					withSpec(v1alpha1.LicenseParameters{License: strPtr(testLicenseKey)}),
					withWriteConnectionSecretRef(testSecretName),
				),
			},
			want: want{
				cr: license(
					withNamespace("default"),
					withExternalName("1"),
					withSpec(v1alpha1.LicenseParameters{License: strPtr(testLicenseKey)}),
					withWriteConnectionSecretRef(testSecretName),
				),
				obs: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false, ResourceLateInitialized: false},
			},
		},
		"MissingConnectionSecret": {
			args: args{
				client: &mockClient{MockGetLicense: func(options ...gitlab.RequestOptionFunc) (*gitlab.License, *gitlab.Response, error) {
					return &gitlab.License{ID: 1, Expired: false}, &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
				}},
				cr: license(
					withNamespace("default"),
					withExternalName("1"),
					withSpec(v1alpha1.LicenseParameters{License: strPtr(testLicenseKey)}),
				),
			},
			want: want{
				cr: license(
					withNamespace("default"),
					withExternalName("1"),
					withSpec(v1alpha1.LicenseParameters{License: strPtr(testLicenseKey)}),
				),
				err: errors.New(errMissingConnectionSecret),
			},
		},
		"MissingLicenseSource": {
			args: args{
				client: &mockClient{MockGetLicense: func(options ...gitlab.RequestOptionFunc) (*gitlab.License, *gitlab.Response, error) {
					return &gitlab.License{ID: 1, Expired: false}, &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
				}},
				cr: license(
					withNamespace("default"),
					withExternalName("1"),
					withWriteConnectionSecretRef(testSecretName),
				),
			},
			want: want{
				cr: license(
					withNamespace("default"),
					withExternalName("1"),
					withWriteConnectionSecretRef(testSecretName),
				),
				err: errors.Wrap(xperrors.New("no license source provided; please specify either LicenseEndpointURL, LicenseEndpointURLSecretRef, LicenseSecretRef, or License in the spec"), errMissingLicenseKey),
			},
		},
		"EndpointFetchFail": {
			args: args{
				client: &mockClient{MockGetLicense: func(options ...gitlab.RequestOptionFunc) (*gitlab.License, *gitlab.Response, error) {
					return &gitlab.License{ID: 1, Expired: false}, &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
				}},
				cr: license(
					withNamespace("default"),
					withExternalName("1"),
					withSpec(v1alpha1.LicenseParameters{LicenseEndpointURL: strPtr("http://invalid-url")}),
					withWriteConnectionSecretRef(testSecretName),
				),
			},
			want: want{
				cr: license(
					withNamespace("default"),
					withExternalName("1"),
					withSpec(v1alpha1.LicenseParameters{LicenseEndpointURL: strPtr("http://invalid-url")}),
					withWriteConnectionSecretRef(testSecretName),
					withConditions(xpv1.Available()),
					withAtProvider(instance.GenerateLicenseObservation(&gitlab.License{ID: 1, Expired: false})),
				),
				obs: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true, ResourceLateInitialized: false},
			},
		},
		"DeletingLicenseGone": {
			args: args{
				client: &mockClient{
					MockDeleteLicense: func(licenseID int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errors.New("not found")
					},
				},
				cr: license(
					withNamespace("default"),
					withExternalName("1"),
					withSpec(v1alpha1.LicenseParameters{License: strPtr(testLicenseKey)}),
					withWriteConnectionSecretRef(testSecretName),
					withDeletionTimestamp(now),
				),
			},
			want: want{
				cr: license(
					withNamespace("default"),
					withExternalName("1"),
					withSpec(v1alpha1.LicenseParameters{License: strPtr(testLicenseKey)}),
					withWriteConnectionSecretRef(testSecretName),
					withDeletionTimestamp(now),
				),
				obs: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"DeletingLicenseDeleted": {
			args: args{
				client: &mockClient{
					MockDeleteLicense: func(licenseID int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{Response: &http.Response{StatusCode: 204}}, nil
					},
				},
				cr: license(
					withNamespace("default"),
					withExternalName("1"),
					withSpec(v1alpha1.LicenseParameters{License: strPtr(testLicenseKey)}),
					withWriteConnectionSecretRef(testSecretName),
					withDeletionTimestamp(now),
				),
			},
			want: want{
				cr: license(
					withNamespace("default"),
					withExternalName("1"),
					withSpec(v1alpha1.LicenseParameters{License: strPtr(testLicenseKey)}),
					withWriteConnectionSecretRef(testSecretName),
					withDeletionTimestamp(now),
				),
				obs: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"DeletingError": {
			args: args{
				client: &mockClient{
					MockDeleteLicense: func(licenseID int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{Response: &http.Response{StatusCode: 500}}, errBoom
					},
				},
				cr: license(
					withNamespace("default"),
					withExternalName("1"),
					withSpec(v1alpha1.LicenseParameters{License: strPtr(testLicenseKey)}),
					withWriteConnectionSecretRef(testSecretName),
					withDeletionTimestamp(now),
				),
			},
			want: want{
				cr: license(
					withNamespace("default"),
					withExternalName("1"),
					withSpec(v1alpha1.LicenseParameters{License: strPtr(testLicenseKey)}),
					withWriteConnectionSecretRef(testSecretName),
					withDeletionTimestamp(now),
				),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.args.kube, client: tc.args.client}
			obs, err := e.Observe(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf(errFmt, diff)
			}
			if diff := cmp.Diff(tc.want.obs, obs); diff != "" {
				t.Errorf("observation: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, cmpopts.EquateEmpty(), test.EquateConditions()); diff != "" {
				t.Errorf(crFmt, diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type want struct {
		cr  resource.Managed
		out managed.ExternalCreation
		err error
	}

	licenseValue := "my-license"
	srcSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "src", Namespace: "default"}, Data: map[string][]byte{"license": []byte(licenseValue)}}

	cases := map[string]struct {
		args args
		want want
	}{
		"InvalidInput": {
			args: args{cr: unexpectedItem},
			want: want{cr: unexpectedItem, err: errors.New(errNotLicense)},
		},
		"MissingConnectionSecretRef": {
			args: args{cr: license(withNamespace("default"), withSpec(v1alpha1.LicenseParameters{License: &licenseValue}))},
			want: want{cr: license(withNamespace("default"), withSpec(v1alpha1.LicenseParameters{License: &licenseValue})), err: errors.New(errMissingConnectionSecret)},
		},
		"MissingLicenseSource": {
			args: args{cr: license(withNamespace("default"), withWriteConnectionSecretRef("conn"))},
			want: want{cr: license(withNamespace("default"), withWriteConnectionSecretRef("conn")), err: errors.Wrap(xperrors.New("no license source provided; please specify either LicenseEndpointURL, LicenseEndpointURLSecretRef, LicenseSecretRef, or License in the spec"), errMissingLicenseKey)},
		},
		"Successful": {
			args: args{
				kube: kubeClient(t, srcSecret),
				client: &mockClient{MockAddLicense: func(opt *gitlab.AddLicenseOptions, options ...gitlab.RequestOptionFunc) (*gitlab.License, *gitlab.Response, error) {
					if opt == nil || opt.License == nil || *opt.License != licenseValue {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 400}}, errors.New("bad opt")
					}
					return &gitlab.License{ID: 7}, &gitlab.Response{Response: &http.Response{StatusCode: 201}}, nil
				}},
				cr: license(
					withNamespace("default"),
					withWriteConnectionSecretRef("conn"),
					withSpec(v1alpha1.LicenseParameters{LicenseSecretRef: common.TestCreateLocalSecretKeySelector("src", "license")}),
				),
			},
			want: want{
				cr: license(
					withNamespace("default"),
					withWriteConnectionSecretRef("conn"),
					withSpec(v1alpha1.LicenseParameters{LicenseSecretRef: common.TestCreateLocalSecretKeySelector("src", "license")}),
					withConditions(xpv1.Creating()),
					withExternalName("7"),
				),
				out: managed.ExternalCreation{ConnectionDetails: managed.ConnectionDetails{keyLicense: []byte(licenseValue)}},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.args.kube, client: tc.args.client}
			out, err := e.Create(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf(errFmt, diff)
			}
			if diff := cmp.Diff(tc.want.out, out, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf(outFmt, diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, cmpopts.EquateEmpty(), test.EquateConditions()); diff != "" {
				t.Errorf(crFmt, diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type want struct {
		cr  resource.Managed
		out managed.ExternalUpdate
		err error
	}

	oldKey := "old"
	newKey := "new"
	connSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "conn", Namespace: "default"}, Data: map[string][]byte{keyLicense: []byte(oldKey)}}

	cases := map[string]struct {
		args args
		want want
	}{
		"InvalidInput": {
			args: args{cr: unexpectedItem},
			want: want{cr: unexpectedItem, err: errors.New(errNotLicense)},
		},
		"MissingConnectionSecretRef": {
			args: args{cr: license(withNamespace("default"), withSpec(v1alpha1.LicenseParameters{License: &newKey}))},
			want: want{cr: license(withNamespace("default"), withSpec(v1alpha1.LicenseParameters{License: &newKey})), err: errors.New(errMissingConnectionSecret)},
		},
		"SkipWhenEndpointFetchFails": func() struct {
			args args
			want want
		} {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}))
			t.Cleanup(srv.Close)

			url := srv.URL
			u := license(
				withNamespace("default"),
				withWriteConnectionSecretRef("conn"),
				withSpec(v1alpha1.LicenseParameters{LicenseEndpointURL: &url}),
			)
			return struct {
				args args
				want want
			}{
				args: args{kube: kubeClient(t, connSecret), client: &mockClient{MockAddLicense: func(*gitlab.AddLicenseOptions, ...gitlab.RequestOptionFunc) (*gitlab.License, *gitlab.Response, error) {
					return nil, &gitlab.Response{Response: &http.Response{StatusCode: 500}}, errors.New("should not be called")
				}}, cr: u},
				want: want{cr: u, out: managed.ExternalUpdate{}},
			}
		}(),
		"NoUpdateWhenUnchanged": {
			args: args{
				kube: kubeClient(t, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "conn", Namespace: "default"}, Data: map[string][]byte{keyLicense: []byte(newKey)}}),
				client: &mockClient{MockAddLicense: func(*gitlab.AddLicenseOptions, ...gitlab.RequestOptionFunc) (*gitlab.License, *gitlab.Response, error) {
					return nil, &gitlab.Response{Response: &http.Response{StatusCode: 500}}, errors.New("should not be called")
				}},
				cr: license(withNamespace("default"), withWriteConnectionSecretRef("conn"), withSpec(v1alpha1.LicenseParameters{License: &newKey})),
			},
			want: want{cr: license(withNamespace("default"), withWriteConnectionSecretRef("conn"), withSpec(v1alpha1.LicenseParameters{License: &newKey})), out: managed.ExternalUpdate{}},
		},
		"UpdateWhenChanged": {
			args: args{
				kube: kubeClient(t, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "conn", Namespace: "default"}, Data: map[string][]byte{keyLicense: []byte(oldKey)}}),
				client: &mockClient{MockAddLicense: func(opt *gitlab.AddLicenseOptions, options ...gitlab.RequestOptionFunc) (*gitlab.License, *gitlab.Response, error) {
					if opt == nil || opt.License == nil || *opt.License != newKey {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 400}}, errors.New("bad opt")
					}
					return &gitlab.License{ID: 9}, &gitlab.Response{Response: &http.Response{StatusCode: 201}}, nil
				}},
				cr: license(withNamespace("default"), withWriteConnectionSecretRef("conn"), withSpec(v1alpha1.LicenseParameters{License: &newKey})),
			},
			want: want{
				cr:  license(withNamespace("default"), withWriteConnectionSecretRef("conn"), withSpec(v1alpha1.LicenseParameters{License: &newKey}), withExternalName("9")),
				out: managed.ExternalUpdate{ConnectionDetails: managed.ConnectionDetails{keyLicense: []byte(newKey)}},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.args.kube, client: tc.args.client}
			out, err := e.Update(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf(errFmt, diff)
			}
			if diff := cmp.Diff(tc.want.out, out, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf(outFmt, diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, cmpopts.EquateEmpty(), test.EquateConditions()); diff != "" {
				t.Errorf(crFmt, diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type want struct {
		cr  resource.Managed
		out managed.ExternalDelete
		err error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"InvalidInput": {
			args: args{cr: unexpectedItem},
			want: want{cr: unexpectedItem, err: errors.New(errNotLicense)},
		},
		"MissingExternalName": {
			args: args{cr: license(withNamespace("default"))},
			want: want{cr: license(withNamespace("default")), err: errors.New(errMissingExternalName)},
		},
		"ExternalNameNotInt": {
			args: args{cr: license(withNamespace("default"), withExternalName("nope"))},
			want: want{cr: license(withNamespace("default"), withExternalName("nope")), err: errors.New(errIDNotInt)},
		},
		"ErrDelete": {
			args: args{
				client: &mockClient{MockDeleteLicense: func(licenseID int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
					return &gitlab.Response{Response: &http.Response{StatusCode: 500}}, errBoom
				}},
				cr: license(withNamespace("default"), withExternalName("1")),
			},
			want: want{cr: license(withNamespace("default"), withExternalName("1")), err: errors.Wrap(errBoom, errDeleteFailed)},
		},
		"ErrDelete404": {
			args: args{
				client: &mockClient{MockDeleteLicense: func(licenseID int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
					return &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errors.New("not found")
				}},
				cr: license(withNamespace("default"), withExternalName("1")),
			},
			want: want{cr: license(withNamespace("default"), withExternalName("1")), out: managed.ExternalDelete{}},
		},
		"Successful": {
			args: args{
				client: &mockClient{MockDeleteLicense: func(licenseID int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
					if licenseID != 1 {
						return &gitlab.Response{Response: &http.Response{StatusCode: 400}}, errors.New("wrong id")
					}
					return &gitlab.Response{Response: &http.Response{StatusCode: 204}}, nil
				}},
				cr: license(withNamespace("default"), withExternalName("1")),
			},
			want: want{cr: license(withNamespace("default"), withExternalName("1")), out: managed.ExternalDelete{}},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.args.kube, client: tc.args.client}
			out, err := e.Delete(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf(errFmt, diff)
			}
			if diff := cmp.Diff(tc.want.out, out, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf(outFmt, diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf(crFmt, diff)
			}
		})
	}
}

func TestDisconnect(t *testing.T) {
	e := &external{}
	if err := e.Disconnect(context.Background()); err != nil {
		t.Fatalf("expected nil error")
	}
}

func TestConnectorConnectInvalidInput(t *testing.T) {
	c := &connector{kube: kubeClient(t)}
	_, err := c.Connect(context.Background(), unexpectedItem)
	if diff := cmp.Diff(errors.New(errNotLicense), err, test.EquateErrors()); diff != "" {
		t.Fatalf("error: -want, +got:\n%s", diff)
	}
}

func TestIsResponseNotFound(t *testing.T) {
	if clients.IsResponseNotFound(nil) {
		t.Fatalf("expected false")
	}
	if !clients.IsResponseNotFound(&gitlab.Response{Response: &http.Response{StatusCode: 404}}) {
		t.Fatalf("expected true")
	}
	if clients.IsResponseNotFound(&gitlab.Response{Response: &http.Response{StatusCode: 500}}) {
		t.Fatalf("expected false")
	}
}

func strPtr(s string) *string { return &s }
