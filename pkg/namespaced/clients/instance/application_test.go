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

package instance

import (
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/instance/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
)

func TestNewApplicationClient(t *testing.T) {
	cfg := common.Config{}
	c := NewApplicationClient(cfg)
	if c == nil {
		t.Fatal("NewApplicationClient returned nil")
	}
}

func TestGenerateApplicationObservation(t *testing.T) {
	appID := "client-app-id"
	callbackURL := "https://example.com/callback"

	cases := map[string]struct {
		app  *gitlab.Application
		want v1alpha1.ApplicationObservation
	}{
		"Nil": {
			app:  nil,
			want: v1alpha1.ApplicationObservation{},
		},
		"Full": {
			app: &gitlab.Application{
				ID:              1,
				ApplicationName: "my-app",
				ApplicationID:   appID,
				CallbackURL:     callbackURL,
				Confidential:    false,
				Scopes:          []string{"api", "read_user"},
			},
			want: v1alpha1.ApplicationObservation{
				ID:            1,
				Name:          "my-app",
				ApplicationID: appID,
				CallbackURL:   callbackURL,
				Confidential:  false,
				Scopes:        []string{"api", "read_user"},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateApplicationObservation(tc.app)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateApplicationObservation(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateCreateApplicationOptions(t *testing.T) {
	name := "my-app"
	redirectURI := "https://example.com/callback"
	scopes := []string{"api", "read_user"}

	cases := map[string]struct {
		params *v1alpha1.ApplicationParameters
		want   *gitlab.CreateApplicationOptions
	}{
		"RequiredFieldsOnly": {
			params: &v1alpha1.ApplicationParameters{
				Name:        name,
				RedirectURI: redirectURI,
				Scopes:      scopes,
			},
			want: &gitlab.CreateApplicationOptions{
				Name:         &name,
				RedirectURI:  &redirectURI,
				Scopes:       ptr.To(strings.Join(scopes, " ")),
				Confidential: nil,
			},
		},
		"WithConfidential": {
			params: &v1alpha1.ApplicationParameters{
				Name:         name,
				RedirectURI:  redirectURI,
				Scopes:       scopes,
				Confidential: ptr.To(true),
			},
			want: &gitlab.CreateApplicationOptions{
				Name:         &name,
				RedirectURI:  &redirectURI,
				Scopes:       ptr.To(strings.Join(scopes, " ")),
				Confidential: ptr.To(true),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateCreateApplicationOptions(tc.params)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateCreateApplicationOptions(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsApplicationUpToDate(t *testing.T) {
	cases := map[string]struct {
		params *v1alpha1.ApplicationParameters
		app    *gitlab.Application
		want   bool
	}{
		"NilParams": {
			params: nil,
			app:    &gitlab.Application{},
			want:   true,
		},
		"NilApp": {
			params: &v1alpha1.ApplicationParameters{Name: "app"},
			app:    nil,
			want:   false,
		},
		"AllFieldsMatch": {
			params: &v1alpha1.ApplicationParameters{
				Name:        "my-app",
				RedirectURI: "https://example.com",
			},
			app: &gitlab.Application{
				ApplicationName: "my-app",
				CallbackURL:     "https://example.com",
				Confidential:    false,
			},
			want: true,
		},
		"NameDiffers": {
			params: &v1alpha1.ApplicationParameters{Name: "app-a", RedirectURI: "https://example.com"},
			app:    &gitlab.Application{ApplicationName: "app-b", CallbackURL: "https://example.com"},
			want:   false,
		},
		"RedirectURIDiffers": {
			params: &v1alpha1.ApplicationParameters{Name: "app", RedirectURI: "https://a.example.com"},
			app:    &gitlab.Application{ApplicationName: "app", CallbackURL: "https://b.example.com"},
			want:   false,
		},
		"ConfidentialDiffers": {
			params: &v1alpha1.ApplicationParameters{
				Name:         "app",
				RedirectURI:  "https://example.com",
				Confidential: ptr.To(true),
			},
			app: &gitlab.Application{
				ApplicationName: "app",
				CallbackURL:     "https://example.com",
				Confidential:    false,
			},
			want: false,
		},
		"ConfidentialNilMatchesFalse": {
			params: &v1alpha1.ApplicationParameters{
				Name:        "app",
				RedirectURI: "https://example.com",
				// Confidential is nil — should not trigger diff
			},
			app: &gitlab.Application{
				ApplicationName: "app",
				CallbackURL:     "https://example.com",
				Confidential:    false,
			},
			want: true,
		},
		"ScopesDiffers": {
			params: &v1alpha1.ApplicationParameters{
				Name:        "app",
				RedirectURI: "https://example.com",
				Scopes:      []string{"api", "read_user"},
			},
			app: &gitlab.Application{
				ApplicationName: "app",
				CallbackURL:     "https://example.com",
				Scopes:          []string{"api"},
			},
			want: false,
		},
		"ScopesMatch": {
			params: &v1alpha1.ApplicationParameters{
				Name:        "app",
				RedirectURI: "https://example.com",
				Scopes:      []string{"read_user", "api"},
			},
			app: &gitlab.Application{
				ApplicationName: "app",
				CallbackURL:     "https://example.com",
				Scopes:          []string{"api", "read_user"},
			},
			want: true,
		},
		"ConfidentialMatchesTrue": {
			params: &v1alpha1.ApplicationParameters{
				Name:         "app",
				RedirectURI:  "https://example.com",
				Confidential: ptr.To(true),
			},
			app: &gitlab.Application{
				ApplicationName: "app",
				CallbackURL:     "https://example.com",
				Confidential:    true,
			},
			want: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsApplicationUpToDate(tc.params, tc.app)
			if got != tc.want {
				t.Errorf("IsApplicationUpToDate() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFindApplicationByID(t *testing.T) {
	app1 := &gitlab.Application{ID: 1, ApplicationName: "app-one"}
	app2 := &gitlab.Application{ID: 2, ApplicationName: "app-two"}

	cases := map[string]struct {
		apps []*gitlab.Application
		id   int64
		want *gitlab.Application
	}{
		"EmptyList": {
			apps: []*gitlab.Application{},
			id:   1,
			want: nil,
		},
		"NilEntry": {
			apps: []*gitlab.Application{nil, app2},
			id:   1,
			want: nil,
		},
		"Found": {
			apps: []*gitlab.Application{app1, app2},
			id:   1,
			want: app1,
		},
		"NotFound": {
			apps: []*gitlab.Application{app1, app2},
			id:   99,
			want: nil,
		},
		"FoundSecond": {
			apps: []*gitlab.Application{app1, app2},
			id:   2,
			want: app2,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := FindApplicationByID(tc.apps, tc.id)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("FindApplicationByID(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsRenewalDue(t *testing.T) {
	period := int64(30)
	pastDate := &metav1.Time{Time: time.Now().UTC().Add(-24 * time.Hour)}
	futureDate := &metav1.Time{Time: time.Now().UTC().Add(24 * time.Hour)}

	cases := map[string]struct {
		params        *v1alpha1.ApplicationParameters
		nextRenewalAt *metav1.Time
		want          bool
	}{
		"NilParams": {
			params: nil,
			want:   false,
		},
		"NoPeriodSet": {
			params: &v1alpha1.ApplicationParameters{Name: "app"},
			want:   false,
		},
		"PeriodSetScheduleUnset": {
			// An unset schedule is not yet initialised, so renewal is not due.
			params: &v1alpha1.ApplicationParameters{RenewalPeriodDays: &period},
			want:   false,
		},
		"PeriodSetScheduleInPast": {
			params:        &v1alpha1.ApplicationParameters{RenewalPeriodDays: &period},
			nextRenewalAt: pastDate,
			want:          true,
		},
		"PeriodSetScheduleInFuture": {
			params:        &v1alpha1.ApplicationParameters{RenewalPeriodDays: &period},
			nextRenewalAt: futureDate,
			want:          false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsApplicationRenewalDue(tc.params, tc.nextRenewalAt)
			if got != tc.want {
				t.Errorf("IsApplicationRenewalDue() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNextRenewalTime(t *testing.T) {
	before := time.Now().UTC()
	result := NextRenewalTime(30)
	after := time.Now().UTC()

	expectedMin := before.AddDate(0, 0, 30)
	expectedMax := after.AddDate(0, 0, 30)

	if result.Before(expectedMin) || result.After(expectedMax) {
		t.Errorf("NextRenewalTime(30) = %v, want between %v and %v", result, expectedMin, expectedMax)
	}
}
