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
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeployTokenRenewAt(t *testing.T) {
	expiresAt := time.Date(2026, time.July, 31, 12, 0, 0, 0, time.UTC)

	cases := map[string]struct {
		renewalPeriodDays *int
		renewBeforeDays   *int
		expiresAt         *time.Time
		want              *metav1.Time
	}{
		"NoRenewalPeriod": {
			expiresAt: &expiresAt,
			want:      nil,
		},
		"NeverExpires": {
			renewalPeriodDays: intPtr(30),
			want:              nil,
		},
		"TwoThirdsOfPeriod": {
			renewalPeriodDays: intPtr(30),
			expiresAt:         &expiresAt,
			want:              &metav1.Time{Time: time.Date(2026, time.July, 21, 12, 0, 0, 0, time.UTC)},
		},
		"RenewBeforeDays": {
			renewalPeriodDays: intPtr(30),
			renewBeforeDays:   intPtr(3),
			expiresAt:         &expiresAt,
			want:              &metav1.Time{Time: time.Date(2026, time.July, 28, 12, 0, 0, 0, time.UTC)},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := DeployTokenRenewAt(tc.renewalPeriodDays, tc.renewBeforeDays, tc.expiresAt)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("DeployTokenRenewAt(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestShouldRenewDeployToken(t *testing.T) {
	expiresAt := time.Date(2026, time.July, 31, 12, 0, 0, 0, time.UTC)
	beforeRenewAt := time.Date(2026, time.July, 21, 11, 59, 59, 0, time.UTC)
	atRenewAt := time.Date(2026, time.July, 21, 12, 0, 0, 0, time.UTC)

	cases := map[string]struct {
		renewalPeriodDays *int
		renewBeforeDays   *int
		dt                *gitlab.DeployToken
		now               time.Time
		want              bool
	}{
		"NoRenewalPeriodNeverRenews": {
			dt:   &gitlab.DeployToken{Expired: true},
			now:  atRenewAt,
			want: false,
		},
		"BeforeRenewAt": {
			renewalPeriodDays: intPtr(30),
			dt:                &gitlab.DeployToken{ExpiresAt: &expiresAt},
			now:               beforeRenewAt,
			want:              false,
		},
		"AtRenewAt": {
			renewalPeriodDays: intPtr(30),
			dt:                &gitlab.DeployToken{ExpiresAt: &expiresAt},
			now:               atRenewAt,
			want:              true,
		},
		"RenewBeforeDaysNotReached": {
			renewalPeriodDays: intPtr(30),
			renewBeforeDays:   intPtr(3),
			dt:                &gitlab.DeployToken{ExpiresAt: &expiresAt},
			now:               atRenewAt,
			want:              false,
		},
		"NeverExpires": {
			renewalPeriodDays: intPtr(30),
			dt:                &gitlab.DeployToken{},
			now:               beforeRenewAt,
			want:              true,
		},
		"Revoked": {
			renewalPeriodDays: intPtr(30),
			dt:                &gitlab.DeployToken{ExpiresAt: &expiresAt, Revoked: true},
			now:               beforeRenewAt,
			want:              true,
		},
		"Expired": {
			renewalPeriodDays: intPtr(30),
			dt:                &gitlab.DeployToken{ExpiresAt: &expiresAt, Expired: true},
			now:               beforeRenewAt,
			want:              true,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := ShouldRenewDeployToken(tc.renewalPeriodDays, tc.renewBeforeDays, tc.dt, tc.now); got != tc.want {
				t.Errorf("ShouldRenewDeployToken() = %v, want %v", got, tc.want)
			}
		})
	}
}
