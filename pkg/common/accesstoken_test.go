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
	createdAt := now.Add(-20 * time.Hour)
	expiresSoon := now.Add(4 * time.Hour)
	expiresLater := now.Add(16 * time.Hour)

	cases := map[string]struct {
		active           bool
		createdAt        *time.Time
		actualExpiresAt  *gitlab.ISOTime
		desiredExpiresAt *time.Time
		want             bool
	}{
		"Inactive":                   {active: false, want: true},
		"ActiveNoDesired":            {active: true, want: false},
		"ActiveMatching":             {active: true, actualExpiresAt: (*gitlab.ISOTime)(&now), desiredExpiresAt: &now, want: false},
		"ActiveMismatching":          {active: true, actualExpiresAt: (*gitlab.ISOTime)(&tomorrow), desiredExpiresAt: &now, want: true},
		"ActiveNilActual":            {active: true, actualExpiresAt: nil, desiredExpiresAt: &now, want: true},
		"RenewalThresholdReached":    {active: true, createdAt: &createdAt, actualExpiresAt: (*gitlab.ISOTime)(&expiresSoon), want: true},
		"RenewalThresholdNotReached": {active: true, createdAt: &createdAt, actualExpiresAt: (*gitlab.ISOTime)(&expiresLater), want: false},
		"RenewalWithoutCreatedAt":    {active: true, actualExpiresAt: (*gitlab.ISOTime)(&expiresSoon), want: false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := ShouldRotateToken(tc.active, tc.createdAt, tc.actualExpiresAt, tc.desiredExpiresAt); got != tc.want {
				t.Errorf("ShouldRotateToken() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestHasReachedRenewalTime(t *testing.T) {
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(30 * time.Hour)
	renewAt := createdAt.Add(20 * time.Hour)

	if HasReachedRenewalTime(createdAt, expiresAt, renewAt.Add(-time.Second)) {
		t.Fatal("expected false before renewal time")
	}
	if !HasReachedRenewalTime(createdAt, expiresAt, renewAt) {
		t.Fatal("expected true at renewal time")
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
