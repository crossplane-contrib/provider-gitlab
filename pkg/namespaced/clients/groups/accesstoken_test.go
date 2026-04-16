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
	gitlab "gitlab.com/gitlab-org/api/client-go"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
)

func TestGenerateCreateGroupAccessTokenOptions(t *testing.T) {
	name := "Name"
	var expiresAt time.Time
	defaultExpiresAt := time.Now().AddDate(0, 0, 7)
	scopes := []string{"scope1", "scope2"}
	accessLevel := v1alpha1.AccessLevelValue(40)
	gitlabAccessLevel := gitlab.AccessLevelValue(40)
	description := "token description"
	type args struct {
		name       string
		parameters *v1alpha1.AccessTokenParameters
	}
	cases := map[string]struct {
		args args
		want *gitlab.CreateGroupAccessTokenOptions
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
			want: &gitlab.CreateGroupAccessTokenOptions{
				Name:        &name,
				AccessLevel: &gitlabAccessLevel,
				ExpiresAt:   (*gitlab.ISOTime)(&expiresAt),
				Scopes:      &scopes,
				Description: &description,
			},
		},
		"noExpiresAt": {
			args: args{
				name: name,
				parameters: &v1alpha1.AccessTokenParameters{
					Name:        name,
					AccessLevel: &accessLevel,
					ExpiresAt:   nil,
					Scopes:      scopes,
				},
			},
			want: &gitlab.CreateGroupAccessTokenOptions{
				Name:        &name,
				AccessLevel: &gitlabAccessLevel,
				ExpiresAt:   (*gitlab.ISOTime)(&defaultExpiresAt),
				Scopes:      &scopes,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateCreateGroupAccessTokenOptions(tc.args.name, tc.args.parameters)

			// Compare fields individually since gitlab.ISOTime has unexported fields
			if tc.want.Name != nil && got.Name != nil && *tc.want.Name != *got.Name {
				t.Errorf("Name: want %v, got %v", *tc.want.Name, *got.Name)
			}

			if tc.want.AccessLevel != nil && got.AccessLevel != nil && *tc.want.AccessLevel != *got.AccessLevel {
				t.Errorf("AccessLevel: want %v, got %v", *tc.want.AccessLevel, *got.AccessLevel)
			}

			if tc.want.ExpiresAt != nil && got.ExpiresAt != nil {
				wantTime := time.Time(*tc.want.ExpiresAt)
				gotTime := time.Time(*got.ExpiresAt)
				wantDay := wantTime.Truncate(24 * time.Hour)
				gotDay := gotTime.Truncate(24 * time.Hour)
				if !wantDay.Equal(gotDay) {
					t.Errorf("ExpiresAt: want %v, got %v", wantDay, gotDay)
				}
			} else if tc.want.ExpiresAt != got.ExpiresAt {
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

func TestGenerateGroupAccessTokenObservation(t *testing.T) {
	expiresAt := time.Now().UTC().Truncate(time.Second)
	createdAt := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)

	got := GenerateGroupAccessTokenObservation(&gitlab.GroupAccessToken{
		PersonalAccessToken: gitlab.PersonalAccessToken{
			ID:          123,
			Name:        "token-name",
			Description: "token-description",
			UserID:      99,
			Scopes:      []string{"read_repository"},
			ExpiresAt:   (*gitlab.ISOTime)(&expiresAt),
			Active:      true,
			CreatedAt:   &createdAt,
			Revoked:     false,
		},
		AccessLevel: 40,
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
	if got.AccessLevel != 40 {
		t.Fatalf("unexpected access level: %d", got.AccessLevel)
	}
}

func TestGenerateRotateGroupAccessTokenOptions(t *testing.T) {
	expiresAt := time.Now().UTC().Truncate(time.Second)
	defaultExpiresAt := time.Now().AddDate(0, 0, 7)

	cases := map[string]struct {
		params *v1alpha1.AccessTokenParameters
		want   *gitlab.RotateGroupAccessTokenOptions
	}{
		"WithExpiresAt": {
			params: &v1alpha1.AccessTokenParameters{ExpiresAt: &v1.Time{Time: expiresAt}},
			want:   &gitlab.RotateGroupAccessTokenOptions{ExpiresAt: (*gitlab.ISOTime)(&expiresAt)},
		},
		"WithoutExpiresAt": {
			params: &v1alpha1.AccessTokenParameters{},
			want:   &gitlab.RotateGroupAccessTokenOptions{ExpiresAt: (*gitlab.ISOTime)(&defaultExpiresAt)},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateRotateGroupAccessTokenOptions(tc.params)

			if tc.want.ExpiresAt != nil && got.ExpiresAt != nil {
				wantTime := time.Time(*tc.want.ExpiresAt)
				gotTime := time.Time(*got.ExpiresAt)
				wantDay := wantTime.Truncate(24 * time.Hour)
				gotDay := gotTime.Truncate(24 * time.Hour)
				if !wantDay.Equal(gotDay) {
					t.Errorf("ExpiresAt: want %v, got %v", wantDay, gotDay)
				}
			} else if tc.want.ExpiresAt != got.ExpiresAt {
				t.Errorf("ExpiresAt: want %v, got %v", tc.want.ExpiresAt, got.ExpiresAt)
			}
		})
	}
}

func TestIsAccessTokenUpToDate(t *testing.T) {
	expiresAt := time.Now().UTC().Truncate(time.Second)
	otherExpiresAt := expiresAt.Add(24 * time.Hour)

	cases := map[string]struct {
		params *v1alpha1.AccessTokenParameters
		at     *gitlab.GroupAccessToken
		want   bool
	}{
		"NilTokenNilParams": {
			params: nil,
			at:     nil,
			want:   true,
		},
		"NilTokenNoDesiredExpiry": {
			params: &v1alpha1.AccessTokenParameters{},
			at:     nil,
			want:   true,
		},
		"NilTokenWithDesiredExpiry": {
			params: &v1alpha1.AccessTokenParameters{ExpiresAt: &v1.Time{Time: expiresAt}},
			at:     nil,
			want:   false,
		},
		"MatchingExpiresAt": {
			params: &v1alpha1.AccessTokenParameters{ExpiresAt: &v1.Time{Time: expiresAt}},
			at: &gitlab.GroupAccessToken{PersonalAccessToken: gitlab.PersonalAccessToken{
				ExpiresAt: (*gitlab.ISOTime)(&expiresAt),
			}},
			want: true,
		},
		"MismatchingExpiresAt": {
			params: &v1alpha1.AccessTokenParameters{ExpiresAt: &v1.Time{Time: expiresAt}},
			at: &gitlab.GroupAccessToken{PersonalAccessToken: gitlab.PersonalAccessToken{
				ExpiresAt: (*gitlab.ISOTime)(&otherExpiresAt),
			}},
			want: false,
		},
		"ObservedNoExpiryDesiredHasExpiry": {
			params: &v1alpha1.AccessTokenParameters{ExpiresAt: &v1.Time{Time: expiresAt}},
			at:     &gitlab.GroupAccessToken{},
			want:   false,
		},
		"ObservedNoExpiryDesiredNoExpiry": {
			params: &v1alpha1.AccessTokenParameters{},
			at:     &gitlab.GroupAccessToken{},
			want:   true,
		},
		"ObservedHasExpiryDesiredNoExpiry": {
			params: &v1alpha1.AccessTokenParameters{},
			at: &gitlab.GroupAccessToken{PersonalAccessToken: gitlab.PersonalAccessToken{
				ExpiresAt: (*gitlab.ISOTime)(&expiresAt),
			}},
			want: true,
		},
		"SameDayDifferentTime": {
			params: &v1alpha1.AccessTokenParameters{ExpiresAt: &v1.Time{Time: time.Date(2026, time.June, 15, 8, 0, 0, 0, time.UTC)}},
			at: &gitlab.GroupAccessToken{PersonalAccessToken: gitlab.PersonalAccessToken{
				ExpiresAt: ptrToISOTime(time.Date(2026, time.June, 15, 0, 0, 0, 0, time.UTC)),
			}},
			want: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsAccessTokenUpToDate(tc.params, tc.at)
			if got != tc.want {
				t.Errorf("IsAccessTokenUpToDate() = %v, want %v", got, tc.want)
			}
		})
	}
}

func ptrToISOTime(t time.Time) *gitlab.ISOTime { return (*gitlab.ISOTime)(&t) }
