/*
Copyright 2026 The Crossplane Authors.

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

	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeployTokenRenewAt returns the time at which a renewal-managed deploy token
// is due for replacement, or nil when renewalPeriodDays is unset or the token
// never expires. GitLab does not report when a deploy token was created, so its
// lifetime is taken to be renewalPeriodDays, the lifetime the provider gives
// every renewal-managed token it creates.
func DeployTokenRenewAt(renewalPeriodDays, renewBeforeDays *int, expiresAt *time.Time) *metav1.Time {
	if renewalPeriodDays == nil || expiresAt == nil {
		return nil
	}
	renewAt, ok := RotationTime(expiresAt.AddDate(0, 0, -*renewalPeriodDays), *expiresAt, renewBeforeDays)
	if !ok {
		return nil
	}
	return &metav1.Time{Time: renewAt}
}

// ShouldRenewDeployToken reports whether a renewal-managed deploy token must be
// replaced because it is revoked, expired, never expires, or has reached its
// renewal time.
func ShouldRenewDeployToken(renewalPeriodDays, renewBeforeDays *int, dt *gitlab.DeployToken, now time.Time) bool {
	if renewalPeriodDays == nil {
		return false
	}
	if dt.Revoked || dt.Expired {
		return true
	}
	renewAt := DeployTokenRenewAt(renewalPeriodDays, renewBeforeDays, dt.ExpiresAt)
	return renewAt == nil || !now.Before(renewAt.Time)
}
