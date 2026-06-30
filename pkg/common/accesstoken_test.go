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

	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func intPtr(v int) *int { return &v }

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
	// Fixed dates for desiredExpiresAt / SameDay checks (no dependency on time.Now)
	fixedDay := time.Date(2026, time.June, 1, 12, 0, 0, 0, time.UTC)
	nextDay := fixedDay.Add(24 * time.Hour)

	// Fixed "now" used for all renewal-threshold tests
	now := time.Date(2026, time.May, 12, 0, 0, 0, 0, time.UTC)

	// Renewal threshold: 2/3 of lifetime already past
	renewalCreatedAt := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	renewalExpiresSoon := time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC)    // 2/3 threshold ~2024-09, before now
	renewalExpiresLater := time.Date(2099, time.December, 31, 0, 0, 0, 0, time.UTC) // 2/3 threshold ~2075, after now

	// renewBeforeDays override tests
	renewBeforeCreatedAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	renewBeforeExpiresFar := time.Date(2099, time.December, 31, 0, 0, 0, 0, time.UTC) // expiresAt - 5 days = ~2099-12-26, after now
	renewBeforeExpiresSoon := time.Date(2026, time.May, 15, 0, 0, 0, 0, time.UTC)     // expiresAt - 365 days = ~2025-05-16, before now

	cases := map[string]struct {
		active           bool
		createdAt        *time.Time
		actualExpiresAt  *gitlab.ISOTime
		desiredExpiresAt *time.Time
		renewBeforeDays  *int
		now              time.Time
		want             bool
	}{
		"Inactive":                   {active: false, now: now, want: true},
		"ActiveNoDesired":            {active: true, now: now, want: false},
		"ActiveMatching":             {active: true, actualExpiresAt: (*gitlab.ISOTime)(&fixedDay), desiredExpiresAt: &fixedDay, now: now, want: false},
		"ActiveMismatching":          {active: true, actualExpiresAt: (*gitlab.ISOTime)(&nextDay), desiredExpiresAt: &fixedDay, now: now, want: true},
		"ActiveNilActual":            {active: true, actualExpiresAt: nil, desiredExpiresAt: &fixedDay, now: now, want: true},
		"RenewalThresholdReached":    {active: true, createdAt: &renewalCreatedAt, actualExpiresAt: (*gitlab.ISOTime)(&renewalExpiresSoon), now: now, want: true},
		"RenewalThresholdNotReached": {active: true, createdAt: &renewalCreatedAt, actualExpiresAt: (*gitlab.ISOTime)(&renewalExpiresLater), now: now, want: false},
		"RenewalWithoutCreatedAt":    {active: true, actualExpiresAt: (*gitlab.ISOTime)(&renewalExpiresSoon), now: now, want: false},
		"RenewBeforeDaysNotYetDue": {
			active:          true,
			createdAt:       &renewBeforeCreatedAt,
			actualExpiresAt: (*gitlab.ISOTime)(&renewBeforeExpiresFar),
			renewBeforeDays: intPtr(5),
			now:             now,
			want:            false, // expiresAt - 5 days = ~2099-12-26, after now
		},
		"RenewBeforeDaysDue": {
			active:          true,
			createdAt:       &renewBeforeCreatedAt,
			actualExpiresAt: (*gitlab.ISOTime)(&renewBeforeExpiresSoon),
			renewBeforeDays: intPtr(365),
			now:             now,
			want:            true, // expiresAt - 365 days = ~2025-05-16, before now
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := ShouldRotateToken(tc.active, tc.createdAt, tc.actualExpiresAt, tc.desiredExpiresAt, tc.renewBeforeDays, tc.now); got != tc.want {
				t.Errorf("ShouldRotateToken() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRotationTime(t *testing.T) {
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC) // 30-day lifetime

	t.Run("DefaultTwoThirds", func(t *testing.T) {
		got, ok := RotationTime(createdAt, expiresAt, nil)
		if !ok {
			t.Fatal("expected ok=true")
		}
		// 2/3 of 30 days = 20 days → 2026-01-21
		want := time.Date(2026, time.January, 21, 0, 0, 0, 0, time.UTC)
		if !got.Equal(want) {
			t.Errorf("want %v, got %v", want, got)
		}
	})

	t.Run("RenewBeforeDaysOverride", func(t *testing.T) {
		got, ok := RotationTime(createdAt, expiresAt, intPtr(5))
		if !ok {
			t.Fatal("expected ok=true")
		}
		// expiresAt - 5 days = 2026-01-26
		want := time.Date(2026, time.January, 26, 0, 0, 0, 0, time.UTC)
		if !got.Equal(want) {
			t.Errorf("want %v, got %v", want, got)
		}
	})

	t.Run("NegativeLifetime", func(t *testing.T) {
		_, ok := RotationTime(expiresAt, createdAt, nil)
		if ok {
			t.Fatal("expected ok=false for negative lifetime")
		}
	})

	t.Run("RenewBeforeDaysWithNegativeLifetime", func(t *testing.T) {
		// renewBeforeDays always returns ok=true regardless of lifetime ordering
		got, ok := RotationTime(expiresAt, createdAt, intPtr(5))
		if !ok {
			t.Fatal("expected ok=true with renewBeforeDays set")
		}
		want := createdAt.AddDate(0, 0, -5)
		if !got.Equal(want) {
			t.Errorf("want %v, got %v", want, got)
		}
	})
}

func TestHasReachedRenewalTime(t *testing.T) {
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(30 * time.Hour)
	renewAt := createdAt.Add(20 * time.Hour) // 2/3 of 30h

	if HasReachedRenewalTime(createdAt, expiresAt, renewAt.Add(-time.Second), nil) {
		t.Fatal("expected false before renewal time")
	}
	if !HasReachedRenewalTime(createdAt, expiresAt, renewAt, nil) {
		t.Fatal("expected true at renewal time")
	}

	// With renewBeforeDays override: expiresAt is 30h from createdAt.
	// renewBeforeDays=1 → rotation at expiresAt - 24h = createdAt + 6h
	overrideRenewAt := expiresAt.Add(-24 * time.Hour)
	if HasReachedRenewalTime(createdAt, expiresAt, overrideRenewAt.Add(-time.Second), intPtr(1)) {
		t.Fatal("expected false before renewBeforeDays threshold")
	}
	if !HasReachedRenewalTime(createdAt, expiresAt, overrideRenewAt, intPtr(1)) {
		t.Fatal("expected true at renewBeforeDays threshold")
	}
}

func TestComputeNextRotation(t *testing.T) {
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC)

	t.Run("NilCreatedAt", func(t *testing.T) {
		if got := ComputeNextRotation(nil, (*gitlab.ISOTime)(&expiresAt), nil); got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("NilExpiresAt", func(t *testing.T) {
		if got := ComputeNextRotation(&createdAt, nil, nil); got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("DefaultTwoThirds", func(t *testing.T) {
		got := ComputeNextRotation(&createdAt, (*gitlab.ISOTime)(&expiresAt), nil)
		if got == nil {
			t.Fatal("expected non-nil")
		}
		want := time.Date(2026, time.January, 21, 0, 0, 0, 0, time.UTC)
		if !got.Time.Equal(want) {
			t.Errorf("want %v, got %v", want, got.Time)
		}
	})

	t.Run("WithRenewBeforeDays", func(t *testing.T) {
		got := ComputeNextRotation(&createdAt, (*gitlab.ISOTime)(&expiresAt), intPtr(5))
		if got == nil {
			t.Fatal("expected non-nil")
		}
		want := time.Date(2026, time.January, 26, 0, 0, 0, 0, time.UTC)
		if !got.Time.Equal(want) {
			t.Errorf("want %v, got %v", want, got.Time)
		}
	})
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

func TestIsSecretRenewalDue(t *testing.T) {
	period := int64(30)
	pastTime := &metav1.Time{Time: time.Now().UTC().Add(-24 * time.Hour)}
	futureTime := &metav1.Time{Time: time.Now().UTC().Add(24 * time.Hour)}

	cases := map[string]struct {
		renewalPeriodDays *int64
		nextRenewalAt     *metav1.Time
		want              bool
	}{
		"NilPeriod": {
			renewalPeriodDays: nil,
			want:              false,
		},
		"NilNextRenewalAt": {
			renewalPeriodDays: &period,
			want:              true,
		},
		"NextRenewalAtInPast": {
			renewalPeriodDays: &period,
			nextRenewalAt:     pastTime,
			want:              true,
		},
		"NextRenewalAtInFuture": {
			renewalPeriodDays: &period,
			nextRenewalAt:     futureTime,
			want:              false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsSecretRenewalDue(tc.renewalPeriodDays, tc.nextRenewalAt)
			if got != tc.want {
				t.Errorf("IsSecretRenewalDue() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNextSecretRenewalTime(t *testing.T) {
	before := time.Now().UTC()
	got := NextSecretRenewalTime(7)
	after := time.Now().UTC()

	earliest := before.AddDate(0, 0, 7)
	latest := after.AddDate(0, 0, 7)
	if got.Before(earliest) || got.After(latest) {
		t.Errorf("NextSecretRenewalTime(7) = %v, want between %v and %v", got, earliest, latest)
	}
}
