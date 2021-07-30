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

package projects

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
)

func TestGenerateCreateProjectDeployTokenOptions(t *testing.T) {
	name := "Name"
	username := "Username"
	expiresAt := time.Now()
	scopes := []string{"scope1", "scope2"}
	type args struct {
		name       string
		parameters *v1alpha1.ProjectDeployTokenParameters
	}
	cases := map[string]struct {
		args args
		want *gitlab.CreateProjectDeployTokenOptions
	}{
		"AllFields": {
			args: args{
				name: name,
				parameters: &v1alpha1.ProjectDeployTokenParameters{
					Username:  &username,
					ExpiresAt: &v1.Time{Time: expiresAt},
					Scopes:    scopes,
				},
			},
			want: &gitlab.CreateProjectDeployTokenOptions{
				Name:      &name,
				Username:  &username,
				ExpiresAt: &expiresAt,
				Scopes:    scopes,
			},
		},
		"SomeFields": {
			args: args{
				name: name,
				parameters: &v1alpha1.ProjectDeployTokenParameters{
					Scopes: scopes,
				},
			},
			want: &gitlab.CreateProjectDeployTokenOptions{
				Name:   &name,
				Scopes: scopes,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateCreateProjectDeployTokenOptions(tc.args.name, tc.args.parameters)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
