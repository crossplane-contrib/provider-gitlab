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

	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	commonv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/instance/v1alpha1"
)

const testEmail = "name@example.org"

func TestGenerateServiceAccountObservation(t *testing.T) {
	type args struct {
		user *gitlab.User
	}

	cases := map[string]struct {
		args args
		want v1alpha1.ServiceAccountObservation
	}{
		"Full": {
			args: args{user: &gitlab.User{ID: 123, Name: "sa", Username: "sa-user"}},
			want: v1alpha1.ServiceAccountObservation{
				CommonServiceAccountObservation: commonv1alpha1.CommonServiceAccountObservation{ID: 123, Name: "sa", Username: "sa-user"},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateServiceAccountObservation(tc.args.user)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateServiceAccountObservation(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateUpdateServiceAccountOptions(t *testing.T) {
	type args struct {
		params *v1alpha1.ServiceAccountParameters
	}

	name := "name"
	username := "user"
	email := testEmail

	cases := map[string]struct {
		args args
		want *gitlab.ModifyUserOptions
	}{
		"AllFields": {
			args: args{params: &v1alpha1.ServiceAccountParameters{CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{Name: &name, Username: &username, Email: &email}}},
			want: &gitlab.ModifyUserOptions{Name: &name, Username: &username, Email: &email},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateUpdateServiceAccountOptions(tc.args.params)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateUpdateServiceAccountOptions(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateServiceAccountCreateOptions(t *testing.T) {
	type args struct {
		params *v1alpha1.ServiceAccountParameters
	}

	name := "name"
	username := "user"
	email := testEmail

	cases := map[string]struct {
		args args
		want *gitlab.CreateServiceAccountUserOptions
	}{
		"AllFields": {
			args: args{params: &v1alpha1.ServiceAccountParameters{CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{Name: &name, Username: &username, Email: &email}}},
			want: &gitlab.CreateServiceAccountUserOptions{Name: &name, Username: &username, Email: &email},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateServiceAccountCreateOptions(tc.args.params)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateServiceAccountCreateOptions(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsServiceAccountUpToDate(t *testing.T) {
	name := "name"
	username := "user"
	email := testEmail

	type args struct {
		params *v1alpha1.ServiceAccountParameters
		user   *gitlab.User
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		// Nil desired state should be treated as up-to-date.
		"NilParams": {
			args: args{params: nil, user: &gitlab.User{}},
			want: true,
		},
		"AllFieldsMatch": {
			args: args{params: &v1alpha1.ServiceAccountParameters{CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{Name: &name, Username: &username, Email: &email}}, user: &gitlab.User{Name: name, Username: username, Email: email}},
			want: true,
		},
		"NameDiffers": {
			args: args{params: &v1alpha1.ServiceAccountParameters{CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{Name: sPtr("a"), Username: &username, Email: &email}}, user: &gitlab.User{Name: "b", Username: username, Email: email}},
			want: false,
		},
		"UsernameDiffers": {
			args: args{params: &v1alpha1.ServiceAccountParameters{CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{Name: &name, Username: sPtr("a"), Email: &email}}, user: &gitlab.User{Name: name, Username: "b", Email: email}},
			want: false,
		},
		"EmailDiffers": {
			args: args{params: &v1alpha1.ServiceAccountParameters{CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{Name: &name, Username: &username, Email: sPtr("a@example.org")}}, user: &gitlab.User{Name: name, Username: username, Email: "b@example.org"}},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsServiceAccountUpToDate(tc.args.params, tc.args.user)
			if got != tc.want {
				t.Errorf("IsServiceAccountUpToDate() = %v, want %v", got, tc.want)
			}
		})
	}
}

func sPtr(s string) *string {
	return &s
}
