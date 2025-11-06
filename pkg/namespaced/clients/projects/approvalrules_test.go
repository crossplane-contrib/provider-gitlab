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

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
)

func TestIsUsernamesUpToDate(t *testing.T) {
	cases := map[string]struct {
		cr   *v1alpha1.ApprovalRuleParameters
		in   *gitlab.ProjectApprovalRule
		want bool
	}{
		"NilUsernamesEmptyUsers": {
			cr: &v1alpha1.ApprovalRuleParameters{
				Usernames: nil,
			},
			in: &gitlab.ProjectApprovalRule{
				Users: []*gitlab.BasicUser{},
			},
			want: true,
		},
		"EmptyUsernamesEmptyUsers": {
			cr: &v1alpha1.ApprovalRuleParameters{
				Usernames: &[]string{},
			},
			in: &gitlab.ProjectApprovalRule{
				Users: []*gitlab.BasicUser{},
			},
			want: true,
		},
		"MatchingUsernames": {
			cr: &v1alpha1.ApprovalRuleParameters{
				Usernames: &[]string{"user1", "user2"},
			},
			in: &gitlab.ProjectApprovalRule{
				Users: []*gitlab.BasicUser{
					{Username: "user1"},
					{Username: "user2"},
				},
			},
			want: true,
		},
		"DifferentUsernames": {
			cr: &v1alpha1.ApprovalRuleParameters{
				Usernames: &[]string{"user1", "user2"},
			},
			in: &gitlab.ProjectApprovalRule{
				Users: []*gitlab.BasicUser{
					{Username: "user1"},
					{Username: "user3"},
				},
			},
			want: false,
		},
		"DifferentLengths": {
			cr: &v1alpha1.ApprovalRuleParameters{
				Usernames: &[]string{"user1"},
			},
			in: &gitlab.ProjectApprovalRule{
				Users: []*gitlab.BasicUser{
					{Username: "user1"},
					{Username: "user2"},
				},
			},
			want: false,
		},
		"NilUsernamesWithUsers": {
			cr: &v1alpha1.ApprovalRuleParameters{
				Usernames: nil,
			},
			in: &gitlab.ProjectApprovalRule{
				Users: []*gitlab.BasicUser{
					{Username: "user1"},
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := isUsernamesUpToDate(tc.cr, tc.in)
			if got != tc.want {
				t.Errorf("isUsernamesUpToDate() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIsUserIDsUpToDate(t *testing.T) {
	cases := map[string]struct {
		cr   *v1alpha1.ApprovalRuleParameters
		in   *gitlab.ProjectApprovalRule
		want bool
	}{
		"NilUserIDsEmptyUsers": {
			cr: &v1alpha1.ApprovalRuleParameters{
				UserIDs: nil,
			},
			in: &gitlab.ProjectApprovalRule{
				Users: []*gitlab.BasicUser{},
			},
			want: true,
		},
		"EmptyUserIDsEmptyUsers": {
			cr: &v1alpha1.ApprovalRuleParameters{
				UserIDs: &[]int{},
			},
			in: &gitlab.ProjectApprovalRule{
				Users: []*gitlab.BasicUser{},
			},
			want: true,
		},
		"MatchingUserIDs": {
			cr: &v1alpha1.ApprovalRuleParameters{
				UserIDs: &[]int{1, 2},
			},
			in: &gitlab.ProjectApprovalRule{
				Users: []*gitlab.BasicUser{
					{ID: 1},
					{ID: 2},
				},
			},
			want: true,
		},
		"DifferentUserIDs": {
			cr: &v1alpha1.ApprovalRuleParameters{
				UserIDs: &[]int{1, 2},
			},
			in: &gitlab.ProjectApprovalRule{
				Users: []*gitlab.BasicUser{
					{ID: 1},
					{ID: 3},
				},
			},
			want: false,
		},
		"DifferentLengths": {
			cr: &v1alpha1.ApprovalRuleParameters{
				UserIDs: &[]int{1},
			},
			in: &gitlab.ProjectApprovalRule{
				Users: []*gitlab.BasicUser{
					{ID: 1},
					{ID: 2},
				},
			},
			want: false,
		},
		"NilUserIDsWithUsers": {
			cr: &v1alpha1.ApprovalRuleParameters{
				UserIDs: nil,
			},
			in: &gitlab.ProjectApprovalRule{
				Users: []*gitlab.BasicUser{
					{ID: 1},
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := isUserIDsUpToDate(tc.cr, tc.in)
			if got != tc.want {
				t.Errorf("isUserIDsUpToDate() = %v, want %v", got, tc.want)
			}
		})
	}
}
