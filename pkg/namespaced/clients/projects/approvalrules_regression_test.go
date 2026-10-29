/*
Copyright 2026 The Crossplane Authors.

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
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
)

func TestIsApprovalRuleUpToDateUserSelectors(t *testing.T) {
	alice := &gitlab.BasicUser{ID: 1, Username: "alice"}
	bob := &gitlab.BasicUser{ID: 2, Username: "bob"}
	cases := map[string]struct {
		ids       *[]int64
		usernames *[]string
		users     []*gitlab.BasicUser
		want      bool
	}{
		"UserIDsOnly": {
			ids: ptr.To([]int64{1}), users: []*gitlab.BasicUser{alice}, want: true,
		},
		"UsernamesOnly": {
			usernames: ptr.To([]string{"alice"}), users: []*gitlab.BasicUser{alice}, want: true,
		},
		"DistinctUsersUnion": {
			ids: ptr.To([]int64{1}), usernames: ptr.To([]string{"bob"}), users: []*gitlab.BasicUser{bob, alice}, want: true,
		},
		"OverlappingSelectors": {
			ids: ptr.To([]int64{1}), usernames: ptr.To([]string{"alice"}), users: []*gitlab.BasicUser{alice}, want: true,
		},
		"DuplicateIDs": {
			ids: ptr.To([]int64{1, 1}), users: []*gitlab.BasicUser{alice}, want: true,
		},
		"DuplicateUsernames": {
			usernames: ptr.To([]string{"alice", "alice"}), users: []*gitlab.BasicUser{alice}, want: true,
		},
		"MissingWantedUser": {
			ids: ptr.To([]int64{1}), usernames: ptr.To([]string{"bob"}), users: []*gitlab.BasicUser{alice}, want: false,
		},
		"UnexpectedUser": {
			ids: ptr.To([]int64{1}), users: []*gitlab.BasicUser{alice, bob}, want: false,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			spec := &v1alpha1.ApprovalRuleParameters{
				Name:              ptr.To("review-rule"),
				ApprovalsRequired: ptr.To(int64(2)),
				UserIDs:           tc.ids,
				Usernames:         tc.usernames,
			}
			observed := &gitlab.ProjectApprovalRule{
				Name:              "review-rule",
				ApprovalsRequired: 2,
				Users:             tc.users,
			}
			if got := IsApprovalRuleUpToDate(spec, observed); got != tc.want {
				t.Errorf("IsApprovalRuleUpToDate() = %v, want %v", got, tc.want)
			}
		})
	}
}
