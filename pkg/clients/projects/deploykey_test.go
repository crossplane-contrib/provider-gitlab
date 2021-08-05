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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
)

var (
	title   = "Title"
	key     = "Key"
	canPush = false
)

func TestGenerateDeployKeyObservation(t *testing.T) {
	id := 0
	createdAt := time.Now()
	type args struct {
		p *gitlab.DeployKey
	}
	cases := map[string]struct {
		args args
		want v1alpha1.DeployKeyObservation
	}{
		"Full": {
			args: args{
				p: &gitlab.DeployKey{
					ID:        id,
					CreatedAt: &createdAt,
				},
			},
			want: v1alpha1.DeployKeyObservation{
				ID:        id,
				CreatedAt: &metav1.Time{Time: createdAt},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateDeployKeyObservation(tc.args.p)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateAddDeployKeyOptions(t *testing.T) {
	type args struct {
		name       string
		parameters *v1alpha1.DeployKeyParameters
	}
	cases := map[string]struct {
		args args
		want *gitlab.AddDeployKeyOptions
	}{
		"AllFields": {
			args: args{
				name: name,
				parameters: &v1alpha1.DeployKeyParameters{
					Title:   title,
					Key:     &key,
					CanPush: &canPush,
				},
			},
			want: &gitlab.AddDeployKeyOptions{
				Title:   &title,
				Key:     &key,
				CanPush: &canPush,
			},
		},
		"SomeFields": {
			args: args{
				name: name,
				parameters: &v1alpha1.DeployKeyParameters{
					Title: title,
					Key:   &key,
				},
			},
			want: &gitlab.AddDeployKeyOptions{
				Title: &title,
				Key:   &key,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateAddDeployKeyOptions(tc.args.parameters)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateUpdateDeployKeyOptions(t *testing.T) {
	type args struct {
		name       string
		parameters *v1alpha1.DeployKeyParameters
	}
	cases := map[string]struct {
		args args
		want *gitlab.UpdateDeployKeyOptions
	}{
		"AllFields": {
			args: args{
				name: name,
				parameters: &v1alpha1.DeployKeyParameters{
					Title:   title,
					CanPush: &canPush,
				},
			},
			want: &gitlab.UpdateDeployKeyOptions{
				Title:   &title,
				CanPush: &canPush,
			},
		},
		"SomeFields": {
			args: args{
				name: name,
				parameters: &v1alpha1.DeployKeyParameters{
					Title: title,
				},
			},
			want: &gitlab.UpdateDeployKeyOptions{
				Title: &title,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateUpdateDeployKeyOptions(tc.args.parameters)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
