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

func TestIsProtectedBranchUpToDate(t *testing.T) {
	accessLevel30 := v1alpha1.AccessLevelValue(30)
	accessLevel40 := v1alpha1.AccessLevelValue(40)

	cases := map[string]struct {
		params *v1alpha1.ProtectedBranchParameters
		pb     *gitlab.ProtectedBranch
		want   bool
	}{
		"UserIDOnlyRuleMatchesRegardlessOfObservedAccessLevel": {
			// Regression test for #267: GitLab reports a real AccessLevel alongside
			// a UserID-scoped rule, so it must not be required to match too.
			params: &v1alpha1.ProtectedBranchParameters{
				PushAccessLevels: []*v1alpha1.BranchAccessDescription{
					{UserID: ptr.To(int64(30584717))},
				},
			},
			pb: &gitlab.ProtectedBranch{
				PushAccessLevels: []*gitlab.BranchAccessDescription{
					{AccessLevel: gitlab.AccessLevelValue(40), UserID: 30584717},
				},
			},
			want: true,
		},
		"UserIDOnlyRuleDoesNotMatchDifferentUser": {
			params: &v1alpha1.ProtectedBranchParameters{
				PushAccessLevels: []*v1alpha1.BranchAccessDescription{
					{UserID: ptr.To(int64(30584717))},
				},
			},
			pb: &gitlab.ProtectedBranch{
				PushAccessLevels: []*gitlab.BranchAccessDescription{
					{AccessLevel: gitlab.AccessLevelValue(40), UserID: 1},
				},
			},
			want: false,
		},
		"AccessLevelOnlyRuleStillRequiresAccessLevelMatch": {
			params: &v1alpha1.ProtectedBranchParameters{
				PushAccessLevels: []*v1alpha1.BranchAccessDescription{
					{AccessLevel: &accessLevel40},
				},
			},
			pb: &gitlab.ProtectedBranch{
				PushAccessLevels: []*gitlab.BranchAccessDescription{
					{AccessLevel: gitlab.AccessLevelValue(30)},
				},
			},
			want: false,
		},
		"DuplicateDesiredRulesDoNotShareOneObservedMatch": {
			// Two identical desired entries must each claim a distinct observed entry.
			params: &v1alpha1.ProtectedBranchParameters{
				PushAccessLevels: []*v1alpha1.BranchAccessDescription{
					{UserID: ptr.To(int64(100))},
					{UserID: ptr.To(int64(100))},
				},
			},
			pb: &gitlab.ProtectedBranch{
				PushAccessLevels: []*gitlab.BranchAccessDescription{
					{AccessLevel: gitlab.AccessLevelValue(40), UserID: 100},
					{AccessLevel: gitlab.AccessLevelValue(40), UserID: 200},
				},
			},
			want: false,
		},
		"NilDesiredEntriesAreFilteredWithoutPanic": {
			// A nil desired entry must be filtered out rather than dereferenced.
			params: &v1alpha1.ProtectedBranchParameters{
				PushAccessLevels: []*v1alpha1.BranchAccessDescription{
					nil,
					{UserID: ptr.To(int64(300))},
				},
			},
			pb: &gitlab.ProtectedBranch{
				PushAccessLevels: []*gitlab.BranchAccessDescription{
					{AccessLevel: gitlab.AccessLevelValue(40), UserID: 300},
				},
			},
			want: true,
		},
		"UnconstrainedRuleDoesNotStealSlotNeededByMoreSpecificRule": {
			// An unconstrained entry must not grab the slot a more specific entry needs.
			params: &v1alpha1.ProtectedBranchParameters{
				PushAccessLevels: []*v1alpha1.BranchAccessDescription{
					{},
					{AccessLevel: &accessLevel30},
				},
			},
			pb: &gitlab.ProtectedBranch{
				PushAccessLevels: []*gitlab.BranchAccessDescription{
					{AccessLevel: gitlab.AccessLevelValue(30)},
					{AccessLevel: gitlab.AccessLevelValue(40)},
				},
			},
			want: true,
		},
		"UnconstrainedRuleOrderReversedGivesSameAnswer": {
			// Same state as above, reordered: the result must not depend on input order.
			params: &v1alpha1.ProtectedBranchParameters{
				PushAccessLevels: []*v1alpha1.BranchAccessDescription{
					{AccessLevel: &accessLevel30},
					{},
				},
			},
			pb: &gitlab.ProtectedBranch{
				PushAccessLevels: []*gitlab.BranchAccessDescription{
					{AccessLevel: gitlab.AccessLevelValue(30)},
					{AccessLevel: gitlab.AccessLevelValue(40)},
				},
			},
			want: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsProtectedBranchUpToDate(tc.params, tc.pb)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("IsProtectedBranchUpToDate() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
