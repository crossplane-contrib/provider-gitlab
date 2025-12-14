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

package groups

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	commonv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
)

const (
	testServiceAccountName     = "sa"
	testServiceAccountUsername = "sa-user"
	testServiceAccountEmail    = "sa@example.org"
)

func TestGenerateServiceAccountObservation(t *testing.T) {
	type args struct {
		user *gitlab.User
	}

	cases := map[string]struct {
		args args
		want v1alpha1.ServiceAccountObservation
	}{
		"Full": {
			args: args{user: &gitlab.User{ID: 123, Name: testServiceAccountName, Username: testServiceAccountUsername}},
			want: v1alpha1.ServiceAccountObservation{CommonServiceAccountObservation: commonv1alpha1.CommonServiceAccountObservation{ID: 123, Name: testServiceAccountName, Username: testServiceAccountUsername}},
		},
		"Empty": {
			args: args{user: &gitlab.User{}},
			want: v1alpha1.ServiceAccountObservation{CommonServiceAccountObservation: commonv1alpha1.CommonServiceAccountObservation{}},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateServiceAccountObservationFromUser(tc.args.user)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateServiceAccountObservation(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateUpdateServiceAccountOptions(t *testing.T) {
	params := v1alpha1.ServiceAccountParameters{
		CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{
			Name:     sPtr(testServiceAccountName),
			Username: sPtr(testServiceAccountUsername),
		},
	}
	got := GenerateUpdateServiceAccountOptions(&params)

	if got == nil {
		t.Fatal("GenerateUpdateServiceAccountOptions(): got nil")
	}
	if got.Name == nil || *got.Name != testServiceAccountName {
		t.Fatalf("GenerateUpdateServiceAccountOptions(): got.Name = %v, want %q", got.Name, testServiceAccountName)
	}
	if got.Username == nil || *got.Username != testServiceAccountUsername {
		t.Fatalf("GenerateUpdateServiceAccountOptions(): got.Username = %v, want %q", got.Username, testServiceAccountUsername)
	}
}

func TestGenerateServiceAccountCreateOptions(t *testing.T) {
	params := v1alpha1.ServiceAccountParameters{
		CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{
			Name:     sPtr(testServiceAccountName),
			Username: sPtr(testServiceAccountUsername),
			Email:    sPtr(testServiceAccountEmail),
		},
	}
	got := GenerateServiceAccountCreateOptions(&params)

	if got == nil {
		t.Fatal("GenerateServiceAccountCreateOptions(): got nil")
	}
	if got.Name == nil || *got.Name != testServiceAccountName {
		t.Fatalf("GenerateServiceAccountCreateOptions(): got.Name = %v, want %q", got.Name, testServiceAccountName)
	}
	if got.Username == nil || *got.Username != testServiceAccountUsername {
		t.Fatalf("GenerateServiceAccountCreateOptions(): got.Username = %v, want %q", got.Username, testServiceAccountUsername)
	}
	if got.Email == nil || *got.Email != testServiceAccountEmail {
		t.Fatalf("GenerateServiceAccountCreateOptions(): got.Email = %v, want %q", got.Email, testServiceAccountEmail)
	}
}

func TestIsServiceAccountUpToDate(t *testing.T) {
	type args struct {
		params *v1alpha1.ServiceAccountParameters
		user   *gitlab.User
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"NilParams": {
			args: args{params: nil, user: &gitlab.User{}},
			want: true,
		},
		"NameMismatch": {
			args: args{params: &v1alpha1.ServiceAccountParameters{CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{Name: sPtr("a"), Username: sPtr("u"), Email: sPtr("e")}}, user: &gitlab.User{Name: "b", Username: "u", Email: "e"}},
			want: false,
		},
		"UsernameMismatch": {
			args: args{params: &v1alpha1.ServiceAccountParameters{CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{Name: sPtr("a"), Username: sPtr("u"), Email: sPtr("e")}}, user: &gitlab.User{Name: "a", Username: "u2", Email: "e"}},
			want: false,
		},
		"EmailMismatch": {
			args: args{params: &v1alpha1.ServiceAccountParameters{CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{Name: sPtr("a"), Username: sPtr("u"), Email: sPtr("e")}}, user: &gitlab.User{Name: "a", Username: "u", Email: "e2"}},
			want: false,
		},
		"UpToDate": {
			args: args{params: &v1alpha1.ServiceAccountParameters{CommonServiceAccountParameters: commonv1alpha1.CommonServiceAccountParameters{Name: sPtr("a"), Username: sPtr("u"), Email: sPtr("e")}}, user: &gitlab.User{Name: "a", Username: "u", Email: "e"}},
			want: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsServiceAccountUpToDate(tc.args.params, tc.args.user)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("IsServiceAccountUpToDate(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestNewServiceAccountClient(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("NewServiceAccountClient() panicked: %v", r)
		}
	}()

	got := NewServiceAccountClient(common.Config{Token: "token"})
	if got == nil {
		t.Fatal("NewServiceAccountClient(): got nil")
	}
}

func sPtr(s string) *string {
	return &s
}
