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
	"time"

	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
)

func TestGenerateCreateGroupDeployTokenOptions(t *testing.T) {
	name := "Name"
	username := "Username"
	expiresAt := time.Now()
	scopes := []string{"scope1", "scope2"}
	type args struct {
		name       string
		parameters *v1alpha1.DeployTokenParameters
	}
	cases := map[string]struct {
		args args
		want *gitlab.CreateGroupDeployTokenOptions
	}{
		"AllFields": {
			args: args{
				name: name,
				parameters: &v1alpha1.DeployTokenParameters{
					Username:  &username,
					ExpiresAt: &v1.Time{Time: expiresAt},
					Scopes:    scopes,
				},
			},
			want: &gitlab.CreateGroupDeployTokenOptions{
				Name:      &name,
				Username:  &username,
				ExpiresAt: &expiresAt,
				Scopes:    &scopes,
			},
		},
		"SomeFields": {
			args: args{
				name: name,
				parameters: &v1alpha1.DeployTokenParameters{
					Scopes: scopes,
				},
			},
			want: &gitlab.CreateGroupDeployTokenOptions{
				Name:   &name,
				Scopes: &scopes,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateCreateGroupDeployTokenOptions(tc.args.name, tc.args.parameters)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateCreateGroupDeployTokenOptionsRenewal(t *testing.T) {
	renewalPeriodDays := 30
	got := GenerateCreateGroupDeployTokenOptions("Name", &v1alpha1.DeployTokenParameters{
		RenewalPeriodDays: &renewalPeriodDays,
		Scopes:            []string{"read_repository"},
	})
	if got.ExpiresAt == nil {
		t.Fatal("ExpiresAt = nil, want 30 days from now")
	}
	want := time.Now().UTC().AddDate(0, 0, 30)
	if d := got.ExpiresAt.Sub(want).Abs(); d > time.Minute {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, want)
	}
}
