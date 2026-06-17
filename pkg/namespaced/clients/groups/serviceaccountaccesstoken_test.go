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

package groups

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
)

// stubSAATClient is a minimal ServiceAccountAccessTokenClient used to exercise
// GetServiceAccountAccessToken without importing the fake package (which would
// create an import cycle for tests in this package).
type stubSAATClient struct {
	list func(gid any, serviceAccount int64, opt *gitlab.ListServiceAccountPersonalAccessTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.PersonalAccessToken, *gitlab.Response, error)
}

func (s *stubSAATClient) ListServiceAccountPersonalAccessTokens(gid any, serviceAccount int64, opt *gitlab.ListServiceAccountPersonalAccessTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return s.list(gid, serviceAccount, opt, options...)
}

func (s *stubSAATClient) CreateServiceAccountPersonalAccessToken(_ any, _ int64, _ *gitlab.CreateServiceAccountPersonalAccessTokenOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return nil, nil, nil
}

func (s *stubSAATClient) RevokeServiceAccountPersonalAccessToken(_ any, _, _ int64, _ ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return nil, nil
}

func (s *stubSAATClient) RotateServiceAccountPersonalAccessToken(_ any, _, _ int64, _ *gitlab.RotateServiceAccountPersonalAccessTokenOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return nil, nil, nil
}

func (s *stubSAATClient) RotateServiceAccountSelf(_ *gitlab.RotatePersonalAccessTokenOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return nil, nil, nil
}

func (s *stubSAATClient) GetServiceAccountSelf(_ ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return nil, nil, nil
}

func (s *stubSAATClient) RevokeServiceAccountSelf(_ ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return nil, nil
}

func TestGenerateCreateServiceAccountAccessTokenOptions(t *testing.T) {
	name := "Name"
	var expiresAt time.Time
	renewalPeriodDays30 := 30
	scopes := []string{"scope1", "scope2"}
	description := "token description"
	type args struct {
		name       string
		parameters *v1alpha1.ServiceAccountAccessTokenParameters
	}
	cases := map[string]struct {
		args        args
		want        *gitlab.CreateServiceAccountPersonalAccessTokenOptions
		wantDaysOut *int
	}{
		"AllFields": {
			args: args{
				name: name,
				parameters: &v1alpha1.ServiceAccountAccessTokenParameters{
					Name:        name,
					ExpiresAt:   &v1.Time{Time: expiresAt},
					Scopes:      scopes,
					Description: &description,
				},
			},
			want: &gitlab.CreateServiceAccountPersonalAccessTokenOptions{
				Name:        &name,
				ExpiresAt:   (*gitlab.ISOTime)(&expiresAt),
				Scopes:      &scopes,
				Description: &description,
			},
		},
		"WithRenewalPeriodDays": {
			args: args{
				name: name,
				parameters: &v1alpha1.ServiceAccountAccessTokenParameters{
					Name:              name,
					RenewalPeriodDays: &renewalPeriodDays30,
					Scopes:            scopes,
				},
			},
			want: &gitlab.CreateServiceAccountPersonalAccessTokenOptions{
				Name:   &name,
				Scopes: &scopes,
			},
			wantDaysOut: &renewalPeriodDays30,
		},
		"NeitherExpiresAtNorRenewalPeriodDays": {
			args: args{
				name: name,
				parameters: &v1alpha1.ServiceAccountAccessTokenParameters{
					Name:   name,
					Scopes: scopes,
				},
			},
			want: &gitlab.CreateServiceAccountPersonalAccessTokenOptions{
				Name:      &name,
				ExpiresAt: nil,
				Scopes:    &scopes,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateCreateServiceAccountAccessTokenOptions(tc.args.parameters)

			if tc.want.Name != nil && got.Name != nil && *tc.want.Name != *got.Name {
				t.Errorf("Name: want %v, got %v", *tc.want.Name, *got.Name)
			}

			switch {
			case tc.wantDaysOut != nil:
				if got.ExpiresAt == nil {
					t.Errorf("ExpiresAt: want ~%d days from now, got nil", *tc.wantDaysOut)
				} else {
					gotTime := time.Time(*got.ExpiresAt)
					expectedDay := time.Now().UTC().AddDate(0, 0, *tc.wantDaysOut).Truncate(24 * time.Hour)
					gotDay := gotTime.UTC().Truncate(24 * time.Hour)
					if !expectedDay.Equal(gotDay) {
						t.Errorf("ExpiresAt: want ~%v, got %v", expectedDay, gotDay)
					}
				}
			case tc.want.ExpiresAt != nil && got.ExpiresAt != nil:
				wantTime := time.Time(*tc.want.ExpiresAt)
				gotTime := time.Time(*got.ExpiresAt)
				if !wantTime.Truncate(24 * time.Hour).Equal(gotTime.Truncate(24 * time.Hour)) {
					t.Errorf("ExpiresAt: want %v, got %v", wantTime, gotTime)
				}
			case tc.want.ExpiresAt != got.ExpiresAt:
				t.Errorf("ExpiresAt: want %v, got %v", tc.want.ExpiresAt, got.ExpiresAt)
			}

			if diff := cmp.Diff(tc.want.Scopes, got.Scopes); diff != "" {
				t.Errorf("Scopes: -want, +got:\n%s", diff)
			}

			if tc.want.Description != nil && got.Description != nil && *tc.want.Description != *got.Description {
				t.Errorf("Description: want %v, got %v", *tc.want.Description, *got.Description)
			}
		})
	}
}

func TestGenerateServiceAccountAccessTokenObservation(t *testing.T) {
	expiresAt := time.Now().UTC().Truncate(time.Second)
	createdAt := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	lastUsedAt := time.Now().UTC().Add(-time.Minute).Truncate(time.Second)

	got := GenerateServiceAccountAccessTokenObservation(&gitlab.PersonalAccessToken{
		ID:          123,
		Name:        "token-name",
		Description: "token-description",
		UserID:      99,
		Scopes:      []string{"read_repository"},
		ExpiresAt:   (*gitlab.ISOTime)(&expiresAt),
		Active:      true,
		CreatedAt:   &createdAt,
		LastUsedAt:  &lastUsedAt,
		Revoked:     false,
	})

	if got.ID != 123 || got.Name != "token-name" || got.Description != "token-description" || got.UserID != 99 || !got.Active || got.Revoked {
		t.Fatalf("unexpected observation scalar fields: %+v", got)
	}
	if diff := cmp.Diff([]string{"read_repository"}, got.Scopes); diff != "" {
		t.Fatalf("unexpected scopes: %s", diff)
	}
	if got.ExpiresAt == nil || !got.ExpiresAt.Time.Equal(expiresAt) {
		t.Fatalf("unexpected expiresAt: %v", got.ExpiresAt)
	}
	if got.CreatedAt == nil || !got.CreatedAt.Time.Equal(createdAt) {
		t.Fatalf("unexpected createdAt: %v", got.CreatedAt)
	}
	if got.LastUsedAt == nil || !got.LastUsedAt.Time.Equal(lastUsedAt) {
		t.Fatalf("unexpected lastUsedAt: %v", got.LastUsedAt)
	}
}

func TestGenerateRotateServiceAccountAccessTokenOptions(t *testing.T) {
	expiresAt := time.Now().UTC().Truncate(time.Second)
	renewalPeriodDays30 := 30

	cases := map[string]struct {
		params      *v1alpha1.ServiceAccountAccessTokenParameters
		want        *gitlab.RotateServiceAccountPersonalAccessTokenOptions
		wantDaysOut *int
	}{
		"WithExpiresAt": {
			params: &v1alpha1.ServiceAccountAccessTokenParameters{ExpiresAt: &v1.Time{Time: expiresAt}},
			want:   &gitlab.RotateServiceAccountPersonalAccessTokenOptions{ExpiresAt: (*gitlab.ISOTime)(&expiresAt)},
		},
		"WithRenewalPeriodDays": {
			params:      &v1alpha1.ServiceAccountAccessTokenParameters{RenewalPeriodDays: &renewalPeriodDays30},
			want:        &gitlab.RotateServiceAccountPersonalAccessTokenOptions{},
			wantDaysOut: &renewalPeriodDays30,
		},
		"NeitherExpiresAtNorRenewalPeriodDays": {
			params: &v1alpha1.ServiceAccountAccessTokenParameters{},
			want:   &gitlab.RotateServiceAccountPersonalAccessTokenOptions{ExpiresAt: nil},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateRotateServiceAccountAccessTokenOptions(tc.params)

			switch {
			case tc.wantDaysOut != nil:
				if got.ExpiresAt == nil {
					t.Errorf("ExpiresAt: want ~%d days from now, got nil", *tc.wantDaysOut)
				} else {
					gotDay := time.Time(*got.ExpiresAt).UTC().Truncate(24 * time.Hour)
					expectedDay := time.Now().UTC().AddDate(0, 0, *tc.wantDaysOut).Truncate(24 * time.Hour)
					if !expectedDay.Equal(gotDay) {
						t.Errorf("ExpiresAt: want ~%v, got %v", expectedDay, gotDay)
					}
				}
			case tc.want.ExpiresAt != nil && got.ExpiresAt != nil:
				if !time.Time(*tc.want.ExpiresAt).Truncate(24 * time.Hour).Equal(time.Time(*got.ExpiresAt).Truncate(24 * time.Hour)) {
					t.Errorf("ExpiresAt: want %v, got %v", tc.want.ExpiresAt, got.ExpiresAt)
				}
			case tc.want.ExpiresAt != got.ExpiresAt:
				t.Errorf("ExpiresAt: want %v, got %v", tc.want.ExpiresAt, got.ExpiresAt)
			}
		})
	}
}

func TestShouldRotateServiceAccountAccessToken(t *testing.T) {
	expiresAt := time.Now().UTC().Truncate(time.Second)
	otherExpiresAt := expiresAt.Add(24 * time.Hour)

	cases := map[string]struct {
		params *v1alpha1.ServiceAccountAccessTokenParameters
		at     *gitlab.PersonalAccessToken
		want   bool
	}{
		"NilToken": {
			params: &v1alpha1.ServiceAccountAccessTokenParameters{},
			at:     nil,
			want:   true,
		},
		"InactiveToken": {
			params: &v1alpha1.ServiceAccountAccessTokenParameters{},
			at:     &gitlab.PersonalAccessToken{Active: false},
			want:   true,
		},
		"ActiveNoDesiredExpiry": {
			params: &v1alpha1.ServiceAccountAccessTokenParameters{},
			at:     &gitlab.PersonalAccessToken{Active: true},
			want:   false,
		},
		"ActiveMatchingExpiresAt": {
			params: &v1alpha1.ServiceAccountAccessTokenParameters{ExpiresAt: &v1.Time{Time: expiresAt}},
			at:     &gitlab.PersonalAccessToken{Active: true, ExpiresAt: ptrToISOTime(expiresAt)},
			want:   false,
		},
		"ActiveMismatchingExpiresAt": {
			params: &v1alpha1.ServiceAccountAccessTokenParameters{ExpiresAt: &v1.Time{Time: expiresAt}},
			at:     &gitlab.PersonalAccessToken{Active: true, ExpiresAt: ptrToISOTime(otherExpiresAt)},
			want:   true,
		},
		"ActiveWithRenewalPeriodDaysPastTwoThirds": {
			params: &v1alpha1.ServiceAccountAccessTokenParameters{RenewalPeriodDays: func() *int { v := 30; return &v }()},
			at: &gitlab.PersonalAccessToken{
				Active:    true,
				CreatedAt: ptrToTime(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)),
				ExpiresAt: ptrToISOTime(time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC)),
			},
			want: true,
		},
		"ActiveWithRenewBeforeDaysDue": {
			params: &v1alpha1.ServiceAccountAccessTokenParameters{
				RenewalPeriodDays: func() *int { v := 365; return &v }(),
				RenewBeforeDays:   func() *int { v := 365; return &v }(),
			},
			at: &gitlab.PersonalAccessToken{
				Active:    true,
				CreatedAt: ptrToTime(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)),
				ExpiresAt: ptrToISOTime(time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC)),
			},
			want: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := ShouldRotateServiceAccountAccessToken(tc.params, tc.at)
			if got != tc.want {
				t.Errorf("ShouldRotateServiceAccountAccessToken() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGetServiceAccountAccessTokenFiltersByActiveState(t *testing.T) {
	var gotState *string
	c := &stubSAATClient{
		list: func(_ any, _ int64, opt *gitlab.ListServiceAccountPersonalAccessTokensOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.PersonalAccessToken, *gitlab.Response, error) {
			gotState = opt.State
			return []*gitlab.PersonalAccessToken{{ID: 1}}, &gitlab.Response{NextPage: 0}, nil
		},
	}

	if _, _, err := GetServiceAccountAccessToken(c, 1, 57, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotState == nil {
		t.Fatal("expected list options State to be set, got nil")
	}
	if *gotState != "active" {
		t.Fatalf("expected State=active, got %q", *gotState)
	}
}

func TestGetServiceAccountAccessToken(t *testing.T) {
	errBoom := errors.New("boom")

	cases := map[string]struct {
		list      func(gid any, serviceAccount int64, opt *gitlab.ListServiceAccountPersonalAccessTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.PersonalAccessToken, *gitlab.Response, error)
		tokenID   int64
		wantID    int64
		wantNil   bool
		wantErr   bool
		wantPages int
	}{
		"FoundOnFirstPage": {
			list: func(gid any, sa int64, opt *gitlab.ListServiceAccountPersonalAccessTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.PersonalAccessToken, *gitlab.Response, error) {
				return []*gitlab.PersonalAccessToken{{ID: 1}, {ID: 2}}, &gitlab.Response{NextPage: 0}, nil
			},
			tokenID: 2,
			wantID:  2,
		},
		"FoundOnSecondPage": {
			list: func() func(gid any, sa int64, opt *gitlab.ListServiceAccountPersonalAccessTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.PersonalAccessToken, *gitlab.Response, error) {
				calls := 0
				return func(gid any, sa int64, opt *gitlab.ListServiceAccountPersonalAccessTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.PersonalAccessToken, *gitlab.Response, error) {
					calls++
					if calls == 1 {
						return []*gitlab.PersonalAccessToken{{ID: 1}}, &gitlab.Response{NextPage: 2}, nil
					}
					return []*gitlab.PersonalAccessToken{{ID: 42}}, &gitlab.Response{NextPage: 0}, nil
				}
			}(),
			tokenID: 42,
			wantID:  42,
		},
		"NotFound": {
			list: func(gid any, sa int64, opt *gitlab.ListServiceAccountPersonalAccessTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.PersonalAccessToken, *gitlab.Response, error) {
				return []*gitlab.PersonalAccessToken{{ID: 1}}, &gitlab.Response{NextPage: 0}, nil
			},
			tokenID: 99,
			wantNil: true,
		},
		"ListError": {
			list: func(gid any, sa int64, opt *gitlab.ListServiceAccountPersonalAccessTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.PersonalAccessToken, *gitlab.Response, error) {
				return nil, nil, errBoom
			},
			tokenID: 1,
			wantErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &stubSAATClient{list: tc.list}
			got, _, err := GetServiceAccountAccessToken(c, 1, 57, tc.tokenID)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantNil {
				if got != nil {
					t.Fatalf("expected nil token, got %+v", got)
				}
				return
			}
			if got == nil || got.ID != tc.wantID {
				t.Fatalf("want token ID %d, got %+v", tc.wantID, got)
			}
		})
	}
}
