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

package deploykeys

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"

	"github.com/crossplane-contrib/provider-gitlab/apis/v1alpha1"
)

var (
	projectID = 0
	title     = "example title"
	key       = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCy9s3kTnxuXKdHWAow344QlX87DdYRf6qZihisX3l55Z0TWikzESWdwbW31f6GIUh9hnRk7/VWp94bLaF/dMw1XSUhCOA8SQW3adQ0AoOOL/ZFqYu6puQDEuoVbb4wRBC7/wCKYIcNDVMVm4Nlrp0Ow3dqCOHuI3DQ8oa34QttbT15Nec0xsfFwxiebi0ZBUD1hg3SNeqEUpwmqUTQYA8LGPYHtqiAZEh55pN695/irBe2BXODvJSuUt//L5SQC2pnLSxjMoa5zK97RXrwXObvDpBZvf2l6yFhktgL+OT7VPQaOEHjVjO2ZRuUeTXOD2FkYhnEvhobXuyw/blY4DpK/jMUPHViX9yWcvxlBb/5piOZbToFsYF58n9+t73IbcHdzfDj0kZakxoCtikX4TUQg3WtaRvDzq3sBvG6u9QQUPOvtxJkJEj7aZVXqmfo+9kUiiPYYWpWfzqLT2sB0PDMMBfu62VK0m8jUUE937Wi29ezDGrHiSgP5aF5KE2G0mc="
	canPush   = false
)

func TestGenerateDeployKeyObservation(t *testing.T) {
	id := 0
	createdAt := time.Now()

	type args struct {
		ph *gitlab.DeployKey
	}

	cases := map[string]struct {
		args args
		want v1alpha1.DeployKeyObservation
	}{
		"Full": {
			args: args{
				ph: &gitlab.DeployKey{
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
			got := GenerateDeployKeyObservation(tc.args.ph)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
func TestLateInitializeDeployKey(t *testing.T) {
	cases := map[string]struct {
		parameters  *v1alpha1.DeployKeyParameters
		projecthook *gitlab.DeployKey
		want        *v1alpha1.DeployKeyParameters
	}{
		"AllOptionalFields": {
			parameters: &v1alpha1.DeployKeyParameters{},
			projecthook: &gitlab.DeployKey{
				CanPush: &canPush,
			},
			want: &v1alpha1.DeployKeyParameters{
				CanPush: &canPush,
				Key:     new(string),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeDeployKey(tc.parameters, tc.projecthook)
			if diff := cmp.Diff(tc.want, tc.parameters); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
func TestGenerateCreateDeployKeyOptions(t *testing.T) {
	type args struct {
		parameters *v1alpha1.DeployKeyParameters
	}
	cases := map[string]struct {
		args args
		want *gitlab.AddDeployKeyOptions
	}{
		"AllFields": {
			args: args{
				parameters: &v1alpha1.DeployKeyParameters{
					ProjectID: &projectID,
					Title:     &title,
					Key:       &key,
					CanPush:   &canPush,
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
				parameters: &v1alpha1.DeployKeyParameters{
					ProjectID: &projectID,
					Title:     &title,
					Key:       &key,
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
			got := GenerateCreateDeployKeyOptions(tc.args.parameters)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
func TestGenerateUpdateDeployKeyOptions(t *testing.T) {
	type args struct {
		parameters *v1alpha1.DeployKeyParameters
	}
	cases := map[string]struct {
		args args
		want *gitlab.UpdateDeployKeyOptions
	}{
		"AllFields": {
			args: args{
				parameters: &v1alpha1.DeployKeyParameters{
					ProjectID: &projectID,
					Title:     &title,
					CanPush:   &canPush,
				},
			},
			want: &gitlab.UpdateDeployKeyOptions{
				Title:   &title,
				CanPush: &canPush,
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
func TestIsDeployKeyUpToDate(t *testing.T) {
	type args struct {
		projecthook *gitlab.DeployKey
		p           *v1alpha1.DeployKeyParameters
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				p: &v1alpha1.DeployKeyParameters{
					ProjectID: &projectID,
					Title:     &title,
					Key:       &key,
					CanPush:   &canPush,
				},
				projecthook: &gitlab.DeployKey{
					Title:   title,
					Key:     key,
					CanPush: &canPush,
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				p: &v1alpha1.DeployKeyParameters{
					ProjectID: &projectID,
					Title:     &title,
					Key:       &key,
					CanPush:   &canPush,
				},
				projecthook: &gitlab.DeployKey{
					Title: "different title",
					Key:   "ssh-rsa BBBBB",
				},
			},
			want: false,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsDeployKeyUpToDate(tc.args.p, tc.args.projecthook)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}

}
