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

	gitlab "gitlab.com/gitlab-org/api/client-go/v3"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
)

func TestIsUsersUpToDate(t *testing.T) {
	cases := map[string]struct {
		cr   *v1alpha1.ApprovalRuleParameters
		in   *gitlab.ProjectApprovalRule
		want bool
	}{
		"NilBothEmptyUsers": {
			cr:   &v1alpha1.ApprovalRuleParameters{},
			in:   &gitlab.ProjectApprovalRule{Users: []*gitlab.BasicUser{}},
			want: true,
		},
		"NilBothWithUsers": {
			cr:   &v1alpha1.ApprovalRuleParameters{},
			in:   &gitlab.ProjectApprovalRule{Users: []*gitlab.BasicUser{{ID: 1, Username: "user1"}}},
			want: false,
		},
		"EmptyUsernamesEmptyUsers": {
			cr:   &v1alpha1.ApprovalRuleParameters{Usernames: &[]string{}},
			in:   &gitlab.ProjectApprovalRule{Users: []*gitlab.BasicUser{}},
			want: true,
		},
		"EmptyUserIDsEmptyUsers": {
			cr:   &v1alpha1.ApprovalRuleParameters{UserIDs: &[]int64{}},
			in:   &gitlab.ProjectApprovalRule{Users: []*gitlab.BasicUser{}},
			want: true,
		},
		"MatchingUsernamesOnly": {
			cr: &v1alpha1.ApprovalRuleParameters{Usernames: &[]string{"user1", "user2"}},
			in: &gitlab.ProjectApprovalRule{Users: []*gitlab.BasicUser{
				{ID: 1, Username: "user1"},
				{ID: 2, Username: "user2"},
			}},
			want: true,
		},
		"DifferentUsernames": {
			cr: &v1alpha1.ApprovalRuleParameters{Usernames: &[]string{"user1", "user2"}},
			in: &gitlab.ProjectApprovalRule{Users: []*gitlab.BasicUser{
				{ID: 1, Username: "user1"},
				{ID: 3, Username: "user3"},
			}},
			want: false,
		},
		"ExtraUsername": {
			cr: &v1alpha1.ApprovalRuleParameters{Usernames: &[]string{"user1"}},
			in: &gitlab.ProjectApprovalRule{Users: []*gitlab.BasicUser{
				{ID: 1, Username: "user1"},
				{ID: 2, Username: "user2"},
			}},
			want: false,
		},
		"MatchingUserIDsOnly": {
			cr: &v1alpha1.ApprovalRuleParameters{UserIDs: &[]int64{1, 2}},
			in: &gitlab.ProjectApprovalRule{Users: []*gitlab.BasicUser{
				{ID: 1, Username: "user1"},
				{ID: 2, Username: "user2"},
			}},
			want: true,
		},
		"DifferentUserIDs": {
			cr: &v1alpha1.ApprovalRuleParameters{UserIDs: &[]int64{1, 2}},
			in: &gitlab.ProjectApprovalRule{Users: []*gitlab.BasicUser{
				{ID: 1, Username: "user1"},
				{ID: 3, Username: "user3"},
			}},
			want: false,
		},
		"UserIDNotInGitLab": {
			cr: &v1alpha1.ApprovalRuleParameters{UserIDs: &[]int64{1, 2}},
			in: &gitlab.ProjectApprovalRule{Users: []*gitlab.BasicUser{
				{ID: 1, Username: "user1"},
			}},
			want: false,
		},
		"UserIDsAndUsernamesUnion": {
			cr: &v1alpha1.ApprovalRuleParameters{
				UserIDs:   &[]int64{1},
				Usernames: &[]string{"user2"},
			},
			in: &gitlab.ProjectApprovalRule{Users: []*gitlab.BasicUser{
				{ID: 1, Username: "user1"},
				{ID: 2, Username: "user2"},
			}},
			want: true,
		},
		"UserIDsAndUsernamesSameUser": {
			cr: &v1alpha1.ApprovalRuleParameters{
				UserIDs:   &[]int64{1},
				Usernames: &[]string{"user1"},
			},
			in: &gitlab.ProjectApprovalRule{Users: []*gitlab.BasicUser{
				{ID: 1, Username: "user1"},
			}},
			want: true,
		},
		"UserIDsAndUsernamesExtraUser": {
			cr: &v1alpha1.ApprovalRuleParameters{
				UserIDs:   &[]int64{1},
				Usernames: &[]string{"user2"},
			},
			in: &gitlab.ProjectApprovalRule{Users: []*gitlab.BasicUser{
				{ID: 1, Username: "user1"},
				{ID: 2, Username: "user2"},
				{ID: 3, Username: "user3"},
			}},
			want: false,
		},
		"UserIDsAndUsernamesMissingUsername": {
			cr: &v1alpha1.ApprovalRuleParameters{
				UserIDs:   &[]int64{1},
				Usernames: &[]string{"user2"},
			},
			in: &gitlab.ProjectApprovalRule{Users: []*gitlab.BasicUser{
				{ID: 1, Username: "user1"},
			}},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := isUsersUpToDate(tc.cr, tc.in)
			if got != tc.want {
				t.Errorf("isUsersUpToDate() = %v, want %v", got, tc.want)
			}
		})
	}
}
