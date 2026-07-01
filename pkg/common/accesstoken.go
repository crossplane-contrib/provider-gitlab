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
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// TimeToMetaTime returns nil if parameter is nil, otherwise metav1.Time value.
func TimeToMetaTime(t *time.Time) *metav1.Time {
	if t == nil {
		return nil
	}
	return &metav1.Time{Time: *t}
}

// GenerateRenewalExpiration returns an ISOTime set to `days` days from now.
func GenerateRenewalExpiration(days int) *gitlab.ISOTime {
	return (*gitlab.ISOTime)(ptr.To(time.Now().UTC().AddDate(0, 0, days)))
}

// ShouldRotateToken returns true when a token must be rotated based on its
// active state, desired expiry drift, or a renewal-managed token lifetime.
// When renewBeforeDays is non-nil the provider rotates once fewer than that
// many days remain before expiry; otherwise the default 2/3-lifetime
// threshold applies.
func ShouldRotateToken(active bool, createdAt *time.Time, actualExpiresAt *gitlab.ISOTime, desiredExpiresAt *time.Time, renewBeforeDays *int, now time.Time) bool {
	if !active {
		return true
	}

	if desiredExpiresAt != nil {
		if actualExpiresAt == nil {
			return true
		}
		return !SameDay(*desiredExpiresAt, time.Time(*actualExpiresAt))
	}

	if createdAt == nil || actualExpiresAt == nil {
		return false
	}

	return HasReachedRenewalTime(*createdAt, time.Time(*actualExpiresAt), now, renewBeforeDays)
}

// HasReachedRenewalTime returns true once the token is eligible for rotation.
// When renewBeforeDays is non-nil rotation is due once expiresAt minus that
// many days has been reached; otherwise the default 2/3-lifetime threshold applies.
func HasReachedRenewalTime(createdAt, expiresAt, now time.Time, renewBeforeDays *int) bool {
	renewAt, ok := RotationTime(createdAt, expiresAt, renewBeforeDays)
	if !ok {
		return false
	}
	return !now.UTC().Before(renewAt)
}

// RotationTime returns the time at which a renewal-managed token becomes
// eligible for rotation. When renewBeforeDays is non-nil the rotation time
// is expiresAt minus that many days; otherwise it is after 2/3 of the
// observed lifetime has elapsed.
func RotationTime(createdAt, expiresAt time.Time, renewBeforeDays *int) (time.Time, bool) {
	createdAt = createdAt.UTC()
	expiresAt = expiresAt.UTC()

	if renewBeforeDays != nil {
		return expiresAt.AddDate(0, 0, -*renewBeforeDays), true
	}

	lifetime := expiresAt.Sub(createdAt)
	if lifetime <= 0 {
		return time.Time{}, false
	}

	renewAt := createdAt.Add(lifetime * 2 / 3)
	return renewAt, true
}

// ComputeNextRotation returns the next rotation time for a renewal-managed
// token, or nil when it cannot be computed (e.g. explicit expiresAt mode
// or missing timestamps).
func ComputeNextRotation(createdAt *time.Time, expiresAt *gitlab.ISOTime, renewBeforeDays *int) *metav1.Time {
	if createdAt == nil || expiresAt == nil {
		return nil
	}
	t, ok := RotationTime(*createdAt, time.Time(*expiresAt), renewBeforeDays)
	if !ok {
		return nil
	}
	return &metav1.Time{Time: t}
}

// SameDay returns true if two times fall on the same UTC calendar day.
func SameDay(a, b time.Time) bool {
	a = a.UTC()
	b = b.UTC()
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

// IsSecretRenewalDue returns true when renewalPeriodDays is set and the renewal
// date stored in annotations[annotationKey] has been reached, is absent, or is malformed.
func IsSecretRenewalDue(renewalPeriodDays *int64, annotationKey string, annotations map[string]string) bool {
	if renewalPeriodDays == nil {
		return false
	}
	if annotations == nil {
		return true
	}
	dateStr, ok := annotations[annotationKey]
	if !ok {
		return true
	}
	renewalDate, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return true
	}
	return !time.Now().Before(renewalDate)
}

// NextSecretRenewalTime returns the UTC time that is periodDays days from now.
func NextSecretRenewalTime(periodDays int64) time.Time {
	return time.Now().UTC().AddDate(0, 0, int(periodDays))
}
