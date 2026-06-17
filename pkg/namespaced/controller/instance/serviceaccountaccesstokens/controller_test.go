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

package serviceaccountaccesstokens

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/instance/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/instance"
)

// mockClient is an inline implementation of instance.ServiceAccountAccessTokenClient.
type mockClient struct {
	MockCreatePersonalAccessToken        func(user int64, opt *gitlab.CreatePersonalAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error)
	MockGetSinglePersonalAccessTokenByID func(token int64, options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error)
	MockRotatePersonalAccessTokenByID    func(token int64, opt *gitlab.RotatePersonalAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error)
	MockRevokePersonalAccessTokenByID    func(token int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
	MockGetServiceAccountSelf            func(options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error)
	MockRotateServiceAccountSelf         func(opt *gitlab.RotatePersonalAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error)
	MockRevokeServiceAccountSelf         func(options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

var _ instance.ServiceAccountAccessTokenClient = &mockClient{}

func (m *mockClient) CreatePersonalAccessToken(user int64, opt *gitlab.CreatePersonalAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return m.MockCreatePersonalAccessToken(user, opt, options...)
}

func (m *mockClient) GetSinglePersonalAccessTokenByID(token int64, options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return m.MockGetSinglePersonalAccessTokenByID(token, options...)
}

func (m *mockClient) RotatePersonalAccessTokenByID(token int64, opt *gitlab.RotatePersonalAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return m.MockRotatePersonalAccessTokenByID(token, opt, options...)
}

func (m *mockClient) RevokePersonalAccessTokenByID(token int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return m.MockRevokePersonalAccessTokenByID(token, options...)
}

func (m *mockClient) GetServiceAccountSelf(options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return m.MockGetServiceAccountSelf(options...)
}

func (m *mockClient) RotateServiceAccountSelf(opt *gitlab.RotatePersonalAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return m.MockRotateServiceAccountSelf(opt, options...)
}

func (m *mockClient) RevokeServiceAccountSelf(options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return m.MockRevokeServiceAccountSelf(options...)
}

var (
	errBoom        = errors.New("boom")
	saID           = int64(57)
	wrongIDstr     = "fr"
	accessTokenID  = int64(1234)
	sAccessTokenID = strconv.FormatInt(accessTokenID, 10)
	invalidInput   resource.Managed
	expiresAt      = time.Now().AddDate(0, 6, 0)
	name           = "Access Token Name"
	token          = "Token"
	tokenObj       = gitlab.PersonalAccessToken{
		ID:        accessTokenID,
		Name:      name,
		UserID:    saID,
		ExpiresAt: (*gitlab.ISOTime)(&expiresAt),
		Token:     token,
		Active:    true,
		Revoked:   false,
		Scopes:    []string{"scope1", "scope2"},
	}
)

// ownerCond is the SelfManaged condition for owner mode.
func ownerCond() v2.Condition { return selfManagedCondition(false) }

type args struct {
	client instance.ServiceAccountAccessTokenClient
	kube   client.Client
	cr     resource.Managed
}

type tokenModifier func(*v1alpha1.ServiceAccountAccessToken)

func withConditions(c ...v2.Condition) tokenModifier {
	return func(r *v1alpha1.ServiceAccountAccessToken) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(fp v1alpha1.ServiceAccountAccessTokenParameters) tokenModifier {
	return func(r *v1alpha1.ServiceAccountAccessToken) { r.Spec.ForProvider = fp }
}

func withExternalName(n string) tokenModifier {
	return func(r *v1alpha1.ServiceAccountAccessToken) { meta.SetExternalName(r, n) }
}

func saToken(m ...tokenModifier) *v1alpha1.ServiceAccountAccessToken {
	cr := &v1alpha1.ServiceAccountAccessToken{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func ownerSpec() v1alpha1.ServiceAccountAccessTokenParameters {
	return v1alpha1.ServiceAccountAccessTokenParameters{
		ServiceAccountID: &saID,
		ExpiresAt:        &v1.Time{Time: expiresAt},
	}
}

// ---- owner-mode Observe ----

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
		"InvalidInput": {
			args: args{cr: invalidInput},
			want: want{cr: invalidInput, err: errors.New(errNotServiceAccountAccessToken)},
		},
		"NoExternalName": {
			args: args{cr: saToken()},
			want: want{cr: saToken(withConditions(ownerCond())), result: managed.ExternalObservation{}},
		},
		"ExternalNameNotID": {
			args: args{cr: saToken(withExternalName(wrongIDstr))},
			want: want{
				cr:  saToken(withExternalName(wrongIDstr), withConditions(ownerCond())),
				err: errors.Wrap(getConversionError(), errFailedParseID),
			},
		},
		"ErrGet": {
			args: args{
				client: &mockClient{
					MockGetSinglePersonalAccessTokenByID: func(_ int64, _ ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
						return nil, nil, errBoom
					},
				},
				cr: saToken(withExternalName(sAccessTokenID), withSpec(ownerSpec())),
			},
			want: want{
				cr:  saToken(withExternalName(sAccessTokenID), withSpec(ownerSpec()), withConditions(ownerCond())),
				err: errors.Wrap(errBoom, errAccessTokenNotFound),
			},
		},
		"Get404": {
			args: args{
				client: &mockClient{
					MockGetSinglePersonalAccessTokenByID: func(_ int64, _ ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errBoom
					},
				},
				cr: saToken(withExternalName(sAccessTokenID), withSpec(ownerSpec())),
			},
			want: want{
				cr:     saToken(withExternalName(sAccessTokenID), withSpec(ownerSpec()), withConditions(ownerCond())),
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"TokenUserMismatch": {
			args: args{
				client: &mockClient{
					MockGetSinglePersonalAccessTokenByID: func(_ int64, _ ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
						return &gitlab.PersonalAccessToken{ID: accessTokenID, UserID: 999, Active: true}, &gitlab.Response{}, nil
					},
				},
				cr: saToken(withExternalName(sAccessTokenID), withSpec(ownerSpec())),
			},
			want: want{
				cr:     saToken(withExternalName(sAccessTokenID), withSpec(ownerSpec()), withConditions(ownerCond())),
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"TokenUpToDate": {
			args: args{
				client: &mockClient{
					MockGetSinglePersonalAccessTokenByID: func(_ int64, _ ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
						return &gitlab.PersonalAccessToken{ID: accessTokenID, UserID: saID, Active: true, ExpiresAt: (*gitlab.ISOTime)(&expiresAt)}, &gitlab.Response{}, nil
					},
				},
				cr: saToken(withExternalName(sAccessTokenID), withSpec(ownerSpec())),
			},
			want: want{
				cr:     saToken(withExternalName(sAccessTokenID), withSpec(ownerSpec()), withConditions(ownerCond(), v2.Available())),
				result: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.client}
			o, err := e.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("err: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions(), cmpopts.IgnoreFields(v1alpha1.ServiceAccountAccessTokenStatus{}, "AtProvider", "RenewAt")); diff != "" {
				t.Errorf("cr: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("result: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestObserveSetsAtProvider(t *testing.T) {
	cr := saToken(withExternalName(sAccessTokenID), withSpec(ownerSpec()))

	e := &external{client: &mockClient{
		MockGetSinglePersonalAccessTokenByID: func(_ int64, _ ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
			return &tokenObj, &gitlab.Response{}, nil
		},
	}}

	if _, err := e.Observe(context.Background(), cr); err != nil {
		t.Fatalf("Observe() error = %v", err)
	}

	want := instance.GenerateServiceAccountAccessTokenObservation(&tokenObj)
	if diff := cmp.Diff(want, cr.Status.AtProvider); diff != "" {
		t.Fatalf("AtProvider mismatch (-want +got):\n%s", diff)
	}
}

// ---- self-mode Observe ----

func TestObserveSelf(t *testing.T) {
	type want struct {
		result         managed.ExternalObservation
		err            error
		wantExternalID string
	}

	cases := map[string]struct {
		client instance.ServiceAccountAccessTokenClient
		cr     *v1alpha1.ServiceAccountAccessToken
		want   want
	}{
		"UpToDateAutoAdopt": {
			client: &mockClient{
				MockGetServiceAccountSelf: func(_ ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
					return &gitlab.PersonalAccessToken{ID: accessTokenID, Active: true}, &gitlab.Response{}, nil
				},
			},
			cr: saToken(withSpec(v1alpha1.ServiceAccountAccessTokenParameters{})),
			want: want{
				result:         managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true},
				wantExternalID: sAccessTokenID,
			},
		},
		"RotationDue": {
			client: &mockClient{
				MockGetServiceAccountSelf: func(_ ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
					return &gitlab.PersonalAccessToken{ID: accessTokenID, Active: false}, &gitlab.Response{}, nil
				},
			},
			cr: saToken(withSpec(v1alpha1.ServiceAccountAccessTokenParameters{})),
			want: want{
				result:         managed.ExternalObservation{ResourceExists: false},
				wantExternalID: sAccessTokenID,
			},
		},
		"DeadTokenTerminalError": {
			client: &mockClient{
				MockGetServiceAccountSelf: func(_ ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
					return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusUnauthorized}}, errBoom
				},
			},
			cr: saToken(withSpec(v1alpha1.ServiceAccountAccessTokenParameters{})),
			want: want{
				err: errors.Wrap(errBoom, errSelfInformFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.client, self: true}
			o, err := e.Observe(context.Background(), tc.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("err: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("result: -want, +got:\n%s", diff)
			}
			if c := tc.cr.Status.GetCondition(TypeSelfManaged); c.Reason != ReasonProviderConfigReferencesManagedToken {
				t.Errorf("expected SelfManaged condition reason %q, got %q", ReasonProviderConfigReferencesManagedToken, c.Reason)
			}
			if tc.want.wantExternalID != "" {
				if got := meta.GetExternalName(tc.cr); got != tc.want.wantExternalID {
					t.Errorf("external-name: want %q, got %q (auto-adopt)", tc.want.wantExternalID, got)
				}
			}
		})
	}
}

// ---- Create ----

func TestCreate(t *testing.T) {
	type want struct {
		cr     resource.Managed
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		self bool
		want want
	}{
		"InvalidInput": {
			args: args{cr: invalidInput},
			want: want{cr: invalidInput, err: errors.New(errNotServiceAccountAccessToken)},
		},
		"OwnerNoServiceAccountID": {
			args: args{cr: saToken(withSpec(v1alpha1.ServiceAccountAccessTokenParameters{Name: name, Scopes: []string{"api"}}))},
			want: want{cr: saToken(withSpec(v1alpha1.ServiceAccountAccessTokenParameters{Name: name, Scopes: []string{"api"}})), err: errors.New(errMissingServiceAccountID)},
		},
		"OwnerCreateFailed": {
			args: args{
				client: &mockClient{
					MockCreatePersonalAccessToken: func(_ int64, _ *gitlab.CreatePersonalAccessTokenOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
						return nil, nil, errBoom
					},
				},
				cr: saToken(withSpec(v1alpha1.ServiceAccountAccessTokenParameters{ServiceAccountID: &saID, Name: name, Scopes: []string{"api"}})),
			},
			want: want{
				cr:  saToken(withSpec(v1alpha1.ServiceAccountAccessTokenParameters{ServiceAccountID: &saID, Name: name, Scopes: []string{"api"}})),
				err: errors.Wrap(errBoom, errCreateFailed),
			},
		},
		"OwnerCreateSuccessful": {
			args: args{
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(nil)},
				client: &mockClient{
					MockCreatePersonalAccessToken: func(_ int64, _ *gitlab.CreatePersonalAccessTokenOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
						return &tokenObj, &gitlab.Response{}, nil
					},
				},
				cr: saToken(withSpec(v1alpha1.ServiceAccountAccessTokenParameters{ServiceAccountID: &saID, Name: name, Scopes: []string{"api"}})),
			},
			want: want{
				cr:     saToken(withSpec(v1alpha1.ServiceAccountAccessTokenParameters{ServiceAccountID: &saID, Name: name, Scopes: []string{"api"}}), withExternalName(sAccessTokenID)),
				result: managed.ExternalCreation{ConnectionDetails: managed.ConnectionDetails{"token": []byte(token)}},
			},
		},
		"OwnerRotationSuccessful": {
			args: args{
				client: &mockClient{
					MockRotatePersonalAccessTokenByID: func(_ int64, _ *gitlab.RotatePersonalAccessTokenOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
						return &tokenObj, &gitlab.Response{}, nil
					},
				},
				cr: saToken(withSpec(v1alpha1.ServiceAccountAccessTokenParameters{ServiceAccountID: &saID}), withExternalName(sAccessTokenID)),
			},
			want: want{
				cr:     saToken(withSpec(v1alpha1.ServiceAccountAccessTokenParameters{ServiceAccountID: &saID}), withExternalName(sAccessTokenID)),
				result: managed.ExternalCreation{ConnectionDetails: managed.ConnectionDetails{"token": []byte(token)}},
			},
		},
		"OwnerRotate404FreshCreate": {
			// Rotation returns 404 (token gone) -> fall through to a fresh create.
			args: args{
				client: &mockClient{
					MockRotatePersonalAccessTokenByID: func(_ int64, _ *gitlab.RotatePersonalAccessTokenOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errBoom
					},
					MockCreatePersonalAccessToken: func(_ int64, _ *gitlab.CreatePersonalAccessTokenOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
						return &tokenObj, &gitlab.Response{}, nil
					},
				},
				cr: saToken(withSpec(v1alpha1.ServiceAccountAccessTokenParameters{ServiceAccountID: &saID, Name: name, Scopes: []string{"api"}}), withExternalName(sAccessTokenID)),
			},
			want: want{
				cr:     saToken(withSpec(v1alpha1.ServiceAccountAccessTokenParameters{ServiceAccountID: &saID, Name: name, Scopes: []string{"api"}}), withExternalName(sAccessTokenID)),
				result: managed.ExternalCreation{ConnectionDetails: managed.ConnectionDetails{"token": []byte(token)}},
			},
		},
		"OwnerRotateOtherErrorSurfaced": {
			// Rotation fails with a non-404 error -> surface it, do NOT fresh-create
			// (tightened fallback: never orphan a possibly-still-valid token).
			args: args{
				client: &mockClient{
					MockRotatePersonalAccessTokenByID: func(_ int64, _ *gitlab.RotatePersonalAccessTokenOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusInternalServerError}}, errBoom
					},
				},
				cr: saToken(withSpec(v1alpha1.ServiceAccountAccessTokenParameters{ServiceAccountID: &saID, Name: name, Scopes: []string{"api"}}), withExternalName(sAccessTokenID)),
			},
			want: want{
				cr:  saToken(withSpec(v1alpha1.ServiceAccountAccessTokenParameters{ServiceAccountID: &saID, Name: name, Scopes: []string{"api"}}), withExternalName(sAccessTokenID)),
				err: errors.Wrap(errBoom, errRotateFailed),
			},
		},
		"SelfRotateSuccessful": {
			self: true,
			args: args{
				client: &mockClient{
					MockRotateServiceAccountSelf: func(_ *gitlab.RotatePersonalAccessTokenOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
						return &tokenObj, &gitlab.Response{}, nil
					},
				},
				cr: saToken(withSpec(v1alpha1.ServiceAccountAccessTokenParameters{})),
			},
			want: want{
				cr:     saToken(withSpec(v1alpha1.ServiceAccountAccessTokenParameters{}), withExternalName(sAccessTokenID)),
				result: managed.ExternalCreation{ConnectionDetails: managed.ConnectionDetails{"token": []byte(token)}},
			},
		},
		"SelfRotateFailed": {
			self: true,
			args: args{
				client: &mockClient{
					MockRotateServiceAccountSelf: func(_ *gitlab.RotatePersonalAccessTokenOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
						return nil, nil, errBoom
					},
				},
				cr: saToken(withSpec(v1alpha1.ServiceAccountAccessTokenParameters{})),
			},
			want: want{
				cr:  saToken(withSpec(v1alpha1.ServiceAccountAccessTokenParameters{})),
				err: errors.Wrap(errBoom, errSelfRotateFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.client, self: tc.self}
			o, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("err: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("cr: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("result: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	e := &external{}
	o, err := e.Update(context.Background(), saToken())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if diff := cmp.Diff(managed.ExternalUpdate{}, o); diff != "" {
		t.Errorf("r: -want, +got:\n%s", diff)
	}
}

// ---- Delete ----

func TestDelete(t *testing.T) {
	type want struct {
		cr  resource.Managed
		err error
	}
	cases := map[string]struct {
		args
		self bool
		want want
	}{
		"InvalidInput": {
			args: args{cr: invalidInput},
			want: want{cr: invalidInput, err: errors.New(errNotServiceAccountAccessToken)},
		},
		"OwnerExternalNameNotInt": {
			args: args{
				client: &mockClient{},
				cr:     saToken(withSpec(ownerSpec()), withExternalName("test")),
			},
			want: want{
				cr:  saToken(withSpec(ownerSpec()), withExternalName("test")),
				err: errors.New(errExternalNameNotInt),
			},
		},
		"OwnerFailedDeletion": {
			args: args{
				client: &mockClient{
					MockRevokePersonalAccessTokenByID: func(_ int64, _ ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: saToken(withSpec(ownerSpec()), withExternalName(strconv.Itoa(0))),
			},
			want: want{
				cr:  saToken(withSpec(ownerSpec()), withExternalName(strconv.Itoa(0))),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
		"OwnerSuccessfulDeletion": {
			args: args{
				client: &mockClient{
					MockRevokePersonalAccessTokenByID: func(_ int64, _ ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: saToken(withSpec(ownerSpec()), withExternalName(sAccessTokenID)),
			},
			want: want{
				cr: saToken(withSpec(ownerSpec()), withExternalName(sAccessTokenID)),
			},
		},
		"SelfSuccessfulRevoke": {
			self: true,
			args: args{
				client: &mockClient{
					MockRevokeServiceAccountSelf: func(_ ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: saToken(withExternalName(sAccessTokenID)),
			},
			want: want{cr: saToken(withExternalName(sAccessTokenID))},
		},
		"SelfRevokeFailed": {
			self: true,
			args: args{
				client: &mockClient{
					MockRevokeServiceAccountSelf: func(_ ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: saToken(withExternalName(sAccessTokenID)),
			},
			want: want{cr: saToken(withExternalName(sAccessTokenID)), err: errors.Wrap(errBoom, errDeleteFailed)},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.client, self: tc.self}
			_, err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("err: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("cr: -want, +got:\n%s", diff)
			}
		})
	}
}

func getConversionError() error {
	_, err := strconv.ParseInt(wrongIDstr, 10, 64)
	return err
}
