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

	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
)

func TestGenerateProtectRepositoryBranchesOptions(t *testing.T) {
	accessLevel30 := v1alpha1.AccessLevelValue(30)
	accessLevel40 := v1alpha1.AccessLevelValue(40)

	cases := map[string]struct {
		params *v1alpha1.ProtectedBranchParameters
		want   *gitlab.ProtectRepositoryBranchesOptions
	}{
		"UserIDOnlyPushAccessLevelIsPreserved": {
			// Regression test for crossplane-contrib/provider-gitlab#267: a
			// userId-only rule must not be silently dropped in favor of an
			// access-level rule, which caused desired/observed state to
			// permanently diverge and the provider to thrash the resource.
			params: &v1alpha1.ProtectedBranchParameters{
				PushAccessLevels: []*v1alpha1.BranchAccessDescription{
					{UserID: ptr.To(int64(30584717))},
				},
			},
			want: &gitlab.ProtectRepositoryBranchesOptions{
				Name: ptr.To("main"),
				AllowedToPush: &[]*gitlab.BranchPermissionOptions{
					{UserID: ptr.To(int64(30584717))},
				},
			},
		},
		"MultipleAccessLevelsAreAllPreserved": {
			params: &v1alpha1.ProtectedBranchParameters{
				PushAccessLevels: []*v1alpha1.BranchAccessDescription{
					{UserID: ptr.To(int64(30584717))},
					{AccessLevel: &accessLevel40},
				},
			},
			want: &gitlab.ProtectRepositoryBranchesOptions{
				Name: ptr.To("main"),
				AllowedToPush: &[]*gitlab.BranchPermissionOptions{
					{UserID: ptr.To(int64(30584717))},
					{AccessLevel: (*gitlab.AccessLevelValue)(&accessLevel40)},
				},
			},
		},
		"GroupIDIsPreserved": {
			params: &v1alpha1.ProtectedBranchParameters{
				MergeAccessLevels: []*v1alpha1.BranchAccessDescription{
					{GroupID: ptr.To(int64(42)), AccessLevel: &accessLevel30},
				},
			},
			want: &gitlab.ProtectRepositoryBranchesOptions{
				Name: ptr.To("main"),
				AllowedToMerge: &[]*gitlab.BranchPermissionOptions{
					{GroupID: ptr.To(int64(42)), AccessLevel: (*gitlab.AccessLevelValue)(&accessLevel30)},
				},
			},
		},
		"NoAccessLevelsProduceNoAllowedToFields": {
			params: &v1alpha1.ProtectedBranchParameters{},
			want: &gitlab.ProtectRepositoryBranchesOptions{
				Name: ptr.To("main"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateProtectRepositoryBranchesOptions("main", tc.params)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateProtectRepositoryBranchesOptions() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
