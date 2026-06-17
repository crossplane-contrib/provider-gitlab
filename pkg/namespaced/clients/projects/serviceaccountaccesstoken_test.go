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

package projects

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
)

// stubSAATClient exercises GetServiceAccountAccessToken pagination.
type stubSAATClient struct {
	list func(pid any, sa int64, opt *gitlab.ListProjectServiceAccountPersonalAccessTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.PersonalAccessToken, *gitlab.Response, error)
}

func (s *stubSAATClient) ListProjectServiceAccountPersonalAccessTokens(pid any, sa int64, opt *gitlab.ListProjectServiceAccountPersonalAccessTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return s.list(pid, sa, opt, options...)
}

func (s *stubSAATClient) CreateProjectServiceAccountPersonalAccessToken(_ any, _ int64, _ *gitlab.CreateProjectServiceAccountPersonalAccessTokenOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return nil, nil, nil
}

func (s *stubSAATClient) RevokeProjectServiceAccountPersonalAccessToken(_ any, _, _ int64, _ ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return nil, nil
}

func (s *stubSAATClient) RotateProjectServiceAccountPersonalAccessToken(_ any, _, _ int64, _ *gitlab.RotateProjectServiceAccountPersonalAccessTokenOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return nil, nil, nil
}

func (s *stubSAATClient) GetServiceAccountSelf(_ ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return nil, nil, nil
}

func (s *stubSAATClient) RotateServiceAccountSelf(_ *gitlab.RotatePersonalAccessTokenOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return nil, nil, nil
}

func (s *stubSAATClient) RevokeServiceAccountSelf(_ ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return nil, nil
}

func TestGenerateCreateServiceAccountAccessTokenOptions(t *testing.T) {
	name := "n"
	scopes := []string{"api", "self_rotate"}
	desc := "d"
	expiresAt := time.Now().UTC().Truncate(time.Second)
	renew := 30

	all := GenerateCreateServiceAccountAccessTokenOptions(&v1alpha1.ServiceAccountAccessTokenParameters{
		Name: name, Scopes: scopes, Description: &desc, ExpiresAt: &v1.Time{Time: expiresAt},
	})
	if *all.Name != name || all.Scopes == nil || *all.Description != desc || all.ExpiresAt == nil {
		t.Errorf("all-fields create options mismatch: %+v", all)
	}

	withRenew := GenerateCreateServiceAccountAccessTokenOptions(&v1alpha1.ServiceAccountAccessTokenParameters{
		Name: name, Scopes: scopes, RenewalPeriodDays: &renew,
	})
	if withRenew.ExpiresAt == nil {
		t.Errorf("expected ExpiresAt computed from renewalPeriodDays")
	}
}

func TestGenerateRotateServiceAccountAccessTokenOptions(t *testing.T) {
	expiresAt := time.Now().UTC().Truncate(time.Second)
	owner := GenerateRotateServiceAccountAccessTokenOptions(&v1alpha1.ServiceAccountAccessTokenParameters{ExpiresAt: &v1.Time{Time: expiresAt}})
	if owner.ExpiresAt == nil {
		t.Errorf("owner rotate ExpiresAt nil")
	}
	self := GenerateRotateServiceAccountSelfOptions(&v1alpha1.ServiceAccountAccessTokenParameters{ExpiresAt: &v1.Time{Time: expiresAt}})
	if self.ExpiresAt == nil {
		t.Errorf("self rotate ExpiresAt nil")
	}
}

func TestGenerateServiceAccountAccessTokenObservation(t *testing.T) {
	expiresAt := time.Now().UTC().Truncate(time.Second)
	createdAt := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	got := GenerateServiceAccountAccessTokenObservation(&gitlab.PersonalAccessToken{
		ID: 5, Name: "n", UserID: 9, Scopes: []string{"api"},
		ExpiresAt: (*gitlab.ISOTime)(&expiresAt), Active: true, CreatedAt: &createdAt,
	})
	if got.ID != 5 || got.Name != "n" || got.UserID != 9 || !got.Active {
		t.Fatalf("observation scalar mismatch: %+v", got)
	}
	if got.ExpiresAt == nil || !got.ExpiresAt.Time.Equal(expiresAt) {
		t.Fatalf("expiresAt mismatch: %v", got.ExpiresAt)
	}
	if diff := cmp.Diff(v1alpha1.ServiceAccountAccessTokenObservation{}, GenerateServiceAccountAccessTokenObservation(nil)); diff != "" {
		t.Fatalf("nil: -want, +got:\n%s", diff)
	}
}

func TestShouldRotateServiceAccountAccessToken(t *testing.T) {
	expiresAt := time.Now().UTC().Truncate(time.Second)
	cases := map[string]struct {
		p    *v1alpha1.ServiceAccountAccessTokenParameters
		at   *gitlab.PersonalAccessToken
		want bool
	}{
		"Nil":           {p: &v1alpha1.ServiceAccountAccessTokenParameters{}, at: nil, want: true},
		"Inactive":      {p: &v1alpha1.ServiceAccountAccessTokenParameters{}, at: &gitlab.PersonalAccessToken{Active: false}, want: true},
		"ActiveNoDrift": {p: &v1alpha1.ServiceAccountAccessTokenParameters{}, at: &gitlab.PersonalAccessToken{Active: true}, want: false},
		"MatchExpiry":   {p: &v1alpha1.ServiceAccountAccessTokenParameters{ExpiresAt: &v1.Time{Time: expiresAt}}, at: &gitlab.PersonalAccessToken{Active: true, ExpiresAt: ptrToISOTime(expiresAt)}, want: false},
		"DriftExpiry":   {p: &v1alpha1.ServiceAccountAccessTokenParameters{ExpiresAt: &v1.Time{Time: expiresAt}}, at: &gitlab.PersonalAccessToken{Active: true, ExpiresAt: ptrToISOTime(expiresAt.Add(48 * time.Hour))}, want: true},
		"RenewBeforeDue": {p: &v1alpha1.ServiceAccountAccessTokenParameters{
			RenewalPeriodDays: func() *int { v := 365; return &v }(),
			RenewBeforeDays:   func() *int { v := 365; return &v }(),
		}, at: &gitlab.PersonalAccessToken{Active: true, CreatedAt: ptrToTime(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)), ExpiresAt: ptrToISOTime(time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC))}, want: true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := ShouldRotateServiceAccountAccessToken(tc.p, tc.at); got != tc.want {
				t.Errorf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestGetServiceAccountAccessToken(t *testing.T) {
	errBoom := errors.New("boom")
	found := &stubSAATClient{list: func(_ any, _ int64, _ *gitlab.ListProjectServiceAccountPersonalAccessTokensOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.PersonalAccessToken, *gitlab.Response, error) {
		return []*gitlab.PersonalAccessToken{{ID: 1}, {ID: 2}}, &gitlab.Response{NextPage: 0}, nil
	}}
	got, _, err := GetServiceAccountAccessToken(found, 1, 7, 2)
	if err != nil || got == nil || got.ID != 2 {
		t.Fatalf("found: got=%+v err=%v", got, err)
	}

	notFound := &stubSAATClient{list: func(_ any, _ int64, _ *gitlab.ListProjectServiceAccountPersonalAccessTokensOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.PersonalAccessToken, *gitlab.Response, error) {
		return []*gitlab.PersonalAccessToken{{ID: 1}}, &gitlab.Response{NextPage: 0}, nil
	}}
	got, _, err = GetServiceAccountAccessToken(notFound, 1, 7, 99)
	if err != nil || got != nil {
		t.Fatalf("notfound: got=%+v err=%v", got, err)
	}

	errClient := &stubSAATClient{list: func(_ any, _ int64, _ *gitlab.ListProjectServiceAccountPersonalAccessTokensOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.PersonalAccessToken, *gitlab.Response, error) {
		return nil, nil, errBoom
	}}
	if _, _, err = GetServiceAccountAccessToken(errClient, 1, 7, 1); err == nil {
		t.Fatal("expected error")
	}
}
