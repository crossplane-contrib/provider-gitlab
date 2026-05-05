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

	gitlab "gitlab.com/gitlab-org/api/client-go"
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
// active state and whether the actual expiry matches the desired expiry.
func ShouldRotateToken(active bool, actualExpiresAt *gitlab.ISOTime, desiredExpiresAt *time.Time) bool {
	if !active {
		return true
	}

	if desiredExpiresAt != nil {
		if actualExpiresAt == nil {
			return true
		}
		return !SameDay(*desiredExpiresAt, time.Time(*actualExpiresAt))
	}

	return false
}

// SameDay returns true if two times fall on the same UTC calendar day.
func SameDay(a, b time.Time) bool {
	a = a.UTC()
	b = b.UTC()
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}
