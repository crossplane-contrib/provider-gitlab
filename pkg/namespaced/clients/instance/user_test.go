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
	"testing"

	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/instance/v1alpha1"
)

func TestGenerateUserObservation(t *testing.T) {
	got := GenerateUserObservation(&gitlab.User{
		ID:             42,
		Username:       "jdoe",
		Name:           "Jane Doe",
		Email:          "jdoe@example.org",
		State:          "blocked",
		IsAdmin:        true,
		CanCreateGroup: true,
		ProjectsLimit:  7,
		External:       true,
		Note:           "contractor",
		NamespaceID:    99,
		Bio:            "not observed",
	})
	want := v1alpha1.UserObservation{
		ID:             42,
		Username:       "jdoe",
		Name:           "Jane Doe",
		Email:          "jdoe@example.org",
		State:          "blocked",
		Admin:          true,
		CanCreateGroup: true,
		ProjectsLimit:  7,
		External:       true,
		Note:           "contractor",
		NamespaceID:    99,
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("GenerateUserObservation(): -want, +got:\n%s", diff)
	}
}

func TestGenerateCreateUserOptions(t *testing.T) {
	cases := map[string]struct {
		params   v1alpha1.UserParameters
		password *string
		want     *gitlab.CreateUserOptions
	}{
		"Password": {
			params: v1alpha1.UserParameters{
				Username:         "jdoe",
				Name:             "Jane Doe",
				Email:            "jdoe@example.org",
				SkipConfirmation: ptr.To(true),
				Admin:            ptr.To(false),
				CanCreateGroup:   ptr.To(true),
				ProjectsLimit:    ptr.To[int64](3),
				External:         ptr.To(true),
				Note:             ptr.To("contractor"),
				State:            ptr.To(v1alpha1.UserStateBlocked),
			},
			password: ptr.To("s3cret-pass"),
			want: &gitlab.CreateUserOptions{
				Username:         ptr.To("jdoe"),
				Name:             ptr.To("Jane Doe"),
				Email:            ptr.To("jdoe@example.org"),
				Password:         ptr.To("s3cret-pass"),
				SkipConfirmation: ptr.To(true),
				Admin:            ptr.To(false),
				CanCreateGroup:   ptr.To(true),
				ProjectsLimit:    ptr.To[int64](3),
				External:         ptr.To(true),
				Note:             ptr.To("contractor"),
			},
		},
		"ResetAndRandomPassword": {
			params: v1alpha1.UserParameters{
				Username:            "jdoe",
				Name:                "Jane Doe",
				Email:               "jdoe@example.org",
				ResetPassword:       ptr.To(true),
				ForceRandomPassword: ptr.To(true),
			},
			want: &gitlab.CreateUserOptions{
				Username:            ptr.To("jdoe"),
				Name:                ptr.To("Jane Doe"),
				Email:               ptr.To("jdoe@example.org"),
				ResetPassword:       ptr.To(true),
				ForceRandomPassword: ptr.To(true),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateCreateUserOptions(&tc.params, tc.password)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateCreateUserOptions(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateModifyUserOptions(t *testing.T) {
	got := GenerateModifyUserOptions(&v1alpha1.UserParameters{
		Username:            "jdoe",
		Name:                "Jane Doe",
		Email:               "jdoe@example.org",
		PasswordSecretRef:   &v2.LocalSecretKeySelector{Key: "password", LocalSecretReference: v2.LocalSecretReference{Name: "jdoe"}},
		ResetPassword:       ptr.To(true),
		ForceRandomPassword: ptr.To(true),
		SkipConfirmation:    ptr.To(true),
		Admin:               ptr.To(true),
		CanCreateGroup:      ptr.To(false),
		ProjectsLimit:       ptr.To[int64](0),
		External:            ptr.To(false),
		Note:                ptr.To(""),
		State:               ptr.To(v1alpha1.UserStateActive),
	})
	want := &gitlab.ModifyUserOptions{
		Username:           ptr.To("jdoe"),
		Name:               ptr.To("Jane Doe"),
		Email:              ptr.To("jdoe@example.org"),
		SkipReconfirmation: ptr.To(true),
		Admin:              ptr.To(true),
		CanCreateGroup:     ptr.To(false),
		ProjectsLimit:      ptr.To[int64](0),
		External:           ptr.To(false),
		Note:               ptr.To(""),
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("GenerateModifyUserOptions(): -want, +got:\n%s", diff)
	}
}

func TestIsUserUpToDate(t *testing.T) {
	observed := gitlab.User{
		Username:       "jdoe",
		Name:           "Jane Doe",
		Email:          "jdoe@example.org",
		State:          "active",
		IsAdmin:        false,
		CanCreateGroup: true,
		ProjectsLimit:  10,
		External:       false,
		Note:           "contractor",
	}
	desired := func(m func(p *v1alpha1.UserParameters)) *v1alpha1.UserParameters {
		p := &v1alpha1.UserParameters{
			Username:       "jdoe",
			Name:           "Jane Doe",
			Email:          "jdoe@example.org",
			Admin:          ptr.To(false),
			CanCreateGroup: ptr.To(true),
			ProjectsLimit:  ptr.To[int64](10),
			External:       ptr.To(false),
			Note:           ptr.To("contractor"),
			State:          ptr.To(v1alpha1.UserStateActive),
		}
		m(p)
		return p
	}

	cases := map[string]struct {
		params *v1alpha1.UserParameters
		want   bool
	}{
		"AllFieldsMatch":           {params: desired(func(*v1alpha1.UserParameters) {}), want: true},
		"EmailDiffersOnlyInCase":   {params: desired(func(p *v1alpha1.UserParameters) { p.Email = "JDoe@Example.org" }), want: true},
		"OptionalFieldsUnset":      {params: &v1alpha1.UserParameters{Username: "jdoe", Name: "Jane Doe", Email: "jdoe@example.org"}, want: true},
		"UsernameDiffers":          {params: desired(func(p *v1alpha1.UserParameters) { p.Username = "jdoe2" }), want: false},
		"NameDiffers":              {params: desired(func(p *v1alpha1.UserParameters) { p.Name = "Jane Roe" }), want: false},
		"EmailDiffers":             {params: desired(func(p *v1alpha1.UserParameters) { p.Email = "jroe@example.org" }), want: false},
		"AdminDiffers":             {params: desired(func(p *v1alpha1.UserParameters) { p.Admin = ptr.To(true) }), want: false},
		"CanCreateGroupDiffers":    {params: desired(func(p *v1alpha1.UserParameters) { p.CanCreateGroup = ptr.To(false) }), want: false},
		"ProjectsLimitDiffers":     {params: desired(func(p *v1alpha1.UserParameters) { p.ProjectsLimit = ptr.To[int64](0) }), want: false},
		"ExternalDiffers":          {params: desired(func(p *v1alpha1.UserParameters) { p.External = ptr.To(true) }), want: false},
		"NoteDiffers":              {params: desired(func(p *v1alpha1.UserParameters) { p.Note = ptr.To("") }), want: false},
		"StateDiffers":             {params: desired(func(p *v1alpha1.UserParameters) { p.State = ptr.To(v1alpha1.UserStateBlocked) }), want: false},
		"PasswordFieldsNotCompare": {params: desired(func(p *v1alpha1.UserParameters) { p.ResetPassword = ptr.To(true) }), want: true},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := IsUserUpToDate(tc.params, &observed); got != tc.want {
				t.Errorf("IsUserUpToDate() = %v, want %v", got, tc.want)
			}
		})
	}
}
