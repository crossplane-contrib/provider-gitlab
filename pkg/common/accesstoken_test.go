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
	"testing"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func TestSameDay(t *testing.T) {
	cases := map[string]struct {
		a, b time.Time
		want bool
	}{
		"Same": {
			a:    time.Date(2026, 6, 15, 8, 0, 0, 0, time.UTC),
			b:    time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
			want: true,
		},
		"Different": {
			a:    time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
			b:    time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC),
			want: false,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := SameDay(tc.a, tc.b); got != tc.want {
				t.Errorf("SameDay() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestShouldRotateToken(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	tomorrow := now.Add(24 * time.Hour)

	cases := map[string]struct {
		active           bool
		actualExpiresAt  *gitlab.ISOTime
		desiredExpiresAt *time.Time
		want             bool
	}{
		"Inactive":          {active: false, want: true},
		"ActiveNoDesired":   {active: true, want: false},
		"ActiveMatching":    {active: true, actualExpiresAt: (*gitlab.ISOTime)(&now), desiredExpiresAt: &now, want: false},
		"ActiveMismatching": {active: true, actualExpiresAt: (*gitlab.ISOTime)(&tomorrow), desiredExpiresAt: &now, want: true},
		"ActiveNilActual":   {active: true, actualExpiresAt: nil, desiredExpiresAt: &now, want: true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := ShouldRotateToken(tc.active, tc.actualExpiresAt, tc.desiredExpiresAt); got != tc.want {
				t.Errorf("ShouldRotateToken() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGenerateRenewalExpiration(t *testing.T) {
	got := GenerateRenewalExpiration(30)
	if got == nil {
		t.Fatal("expected non-nil")
	}
	gotTime := time.Time(*got)
	expected := time.Now().UTC().AddDate(0, 0, 30).Truncate(24 * time.Hour)
	gotDay := gotTime.UTC().Truncate(24 * time.Hour)
	if !expected.Equal(gotDay) {
		t.Errorf("want ~%v, got %v", expected, gotDay)
	}
}
