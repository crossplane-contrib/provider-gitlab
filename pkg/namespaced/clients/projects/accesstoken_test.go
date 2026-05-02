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
	gitlab "gitlab.com/gitlab-org/api/client-go"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
)

func TestGenerateCreateProjectAccessTokenOptions(t *testing.T) {
	name := "Name"
	var expiresAt time.Time
	renewalPeriodDays30 := 30
	scopes := []string{"scope1", "scope2"}
	accessLevel := v1alpha1.AccessLevelValue(40)
	gitlabAccessLevel := gitlab.AccessLevelValue(40)
	description := "token description"

	type args struct {
		name       string
		parameters *v1alpha1.AccessTokenParameters
	}

	cases := map[string]struct {
		args        args
		want        *gitlab.CreateProjectAccessTokenOptions
		wantDaysOut *int // expected days from now for ExpiresAt when using RenewalPeriodDays
	}{
		"AllFields": {
			args: args{
				name: name,
				parameters: &v1alpha1.AccessTokenParameters{
					Name:        name,
					AccessLevel: &accessLevel,
					ExpiresAt:   &v1.Time{Time: expiresAt},
					Scopes:      scopes,
					Description: &description,
				},
			},
			want: &gitlab.CreateProjectAccessTokenOptions{
				Name:        &name,
				AccessLevel: &gitlabAccessLevel,
				ExpiresAt:   (*gitlab.ISOTime)(&expiresAt),
				Scopes:      &scopes,
				Description: &description,
			},
		},
		"WithRenewalPeriodDays": {
			args: args{
				name: name,
				parameters: &v1alpha1.AccessTokenParameters{
					Name:              name,
					AccessLevel:       &accessLevel,
					RenewalPeriodDays: &renewalPeriodDays30,
					Scopes:            scopes,
				},
			},
			want: &gitlab.CreateProjectAccessTokenOptions{
				Name:        &name,
				AccessLevel: &gitlabAccessLevel,
				Scopes:      &scopes,
			},
			wantDaysOut: &renewalPeriodDays30,
		},
		"NeitherExpiresAtNorRenewalPeriodDays": {
			args: args{
				name: name,
				parameters: &v1alpha1.AccessTokenParameters{
					Name:        name,
					AccessLevel: &accessLevel,
					Scopes:      scopes,
				},
			},
			want: &gitlab.CreateProjectAccessTokenOptions{
				Name:        &name,
				AccessLevel: &gitlabAccessLevel,
				ExpiresAt:   nil,
				Scopes:      &scopes,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateCreateProjectAccessTokenOptions(tc.args.name, tc.args.parameters)

			if tc.want.Name != nil && got.Name != nil && *tc.want.Name != *got.Name {
				t.Errorf("Name: want %v, got %v", *tc.want.Name, *got.Name)
			}

			if tc.want.AccessLevel != nil && got.AccessLevel != nil && *tc.want.AccessLevel != *got.AccessLevel {
				t.Errorf("AccessLevel: want %v, got %v", *tc.want.AccessLevel, *got.AccessLevel)
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
				wantDay := wantTime.Truncate(24 * time.Hour)
				gotDay := gotTime.Truncate(24 * time.Hour)
				if !wantDay.Equal(gotDay) {
					t.Errorf("ExpiresAt: want %v, got %v", wantDay, gotDay)
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

func TestGenerateProjectAccessTokenObservation(t *testing.T) {
	tokenID := int64(123)
	got := GenerateProjectAccessTokenObservation(&gitlab.ProjectAccessToken{PersonalAccessToken: gitlab.PersonalAccessToken{ID: tokenID}})

	if got.TokenID == nil || *got.TokenID != tokenID {
		t.Fatalf("unexpected token id: %v", got.TokenID)
	}
}

func TestGenerateRotateProjectAccessTokenOptions(t *testing.T) {
	expiresAt := time.Now().UTC().Truncate(time.Second)
	renewalPeriodDays30 := 30

	cases := map[string]struct {
		params      *v1alpha1.AccessTokenParameters
		want        *gitlab.RotateProjectAccessTokenOptions
		wantDaysOut *int
	}{
		"WithExpiresAt": {
			params: &v1alpha1.AccessTokenParameters{ExpiresAt: &v1.Time{Time: expiresAt}},
			want:   &gitlab.RotateProjectAccessTokenOptions{ExpiresAt: (*gitlab.ISOTime)(&expiresAt)},
		},
		"WithRenewalPeriodDays": {
			params:      &v1alpha1.AccessTokenParameters{RenewalPeriodDays: &renewalPeriodDays30},
			want:        &gitlab.RotateProjectAccessTokenOptions{},
			wantDaysOut: &renewalPeriodDays30,
		},
		"NeitherExpiresAtNorRenewalPeriodDays": {
			params: &v1alpha1.AccessTokenParameters{},
			want:   &gitlab.RotateProjectAccessTokenOptions{ExpiresAt: nil},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateRotateProjectAccessTokenOptions(tc.params)

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
				wantDay := wantTime.Truncate(24 * time.Hour)
				gotDay := gotTime.Truncate(24 * time.Hour)
				if !wantDay.Equal(gotDay) {
					t.Errorf("ExpiresAt: want %v, got %v", wantDay, gotDay)
				}
			case tc.want.ExpiresAt != got.ExpiresAt:
				t.Errorf("ExpiresAt: want %v, got %v", tc.want.ExpiresAt, got.ExpiresAt)
			}
		})
	}
}

func TestShouldRotateProjectAccessToken(t *testing.T) {
	expiresAt := time.Now().UTC().Truncate(time.Second)
	otherExpiresAt := expiresAt.Add(24 * time.Hour)

	cases := map[string]struct {
		params *v1alpha1.AccessTokenParameters
		at     *gitlab.ProjectAccessToken
		want   bool
	}{
		"NilToken": {
			params: &v1alpha1.AccessTokenParameters{},
			at:     nil,
			want:   true,
		},
		"InactiveToken": {
			params: &v1alpha1.AccessTokenParameters{},
			at:     &gitlab.ProjectAccessToken{PersonalAccessToken: gitlab.PersonalAccessToken{Active: false}},
			want:   true,
		},
		"ActiveNoDesiredExpiry": {
			params: &v1alpha1.AccessTokenParameters{},
			at:     &gitlab.ProjectAccessToken{PersonalAccessToken: gitlab.PersonalAccessToken{Active: true}},
			want:   false,
		},
		"ActiveMatchingExpiresAt": {
			params: &v1alpha1.AccessTokenParameters{ExpiresAt: &v1.Time{Time: expiresAt}},
			at: &gitlab.ProjectAccessToken{PersonalAccessToken: gitlab.PersonalAccessToken{
				Active:    true,
				ExpiresAt: ptrToISOTime(expiresAt),
			}},
			want: false,
		},
		"ActiveMismatchingExpiresAt": {
			params: &v1alpha1.AccessTokenParameters{ExpiresAt: &v1.Time{Time: expiresAt}},
			at: &gitlab.ProjectAccessToken{PersonalAccessToken: gitlab.PersonalAccessToken{
				Active:    true,
				ExpiresAt: ptrToISOTime(otherExpiresAt),
			}},
			want: true,
		},
		"ActiveNoActualExpiryButDesiredSet": {
			params: &v1alpha1.AccessTokenParameters{ExpiresAt: &v1.Time{Time: expiresAt}},
			at:     &gitlab.ProjectAccessToken{PersonalAccessToken: gitlab.PersonalAccessToken{Active: true}},
			want:   true,
		},
		"ActiveSameDayDifferentTime": {
			params: &v1alpha1.AccessTokenParameters{ExpiresAt: &v1.Time{Time: time.Date(2026, time.June, 15, 8, 0, 0, 0, time.UTC)}},
			at: &gitlab.ProjectAccessToken{PersonalAccessToken: gitlab.PersonalAccessToken{
				Active:    true,
				ExpiresAt: ptrToISOTime(time.Date(2026, time.June, 15, 0, 0, 0, 0, time.UTC)),
			}},
			want: false,
		},
		"ActiveWithRenewalPeriodDays": {
			params: &v1alpha1.AccessTokenParameters{RenewalPeriodDays: func() *int { v := 30; return &v }()},
			at: &gitlab.ProjectAccessToken{PersonalAccessToken: gitlab.PersonalAccessToken{
				Active:    true,
				ExpiresAt: ptrToISOTime(time.Now().UTC().AddDate(0, 0, 20)),
			}},
			want: false, // active → no rotation until it expires
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := ShouldRotateAccessToken(tc.params, tc.at)
			if got != tc.want {
				t.Errorf("ShouldRotateAccessToken() = %v, want %v", got, tc.want)
			}
		})
	}
}

func ptrToISOTime(t time.Time) *gitlab.ISOTime { return (*gitlab.ISOTime)(&t) }
