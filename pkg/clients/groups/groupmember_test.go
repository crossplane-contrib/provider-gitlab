/*
Copyright 2020 The Crossplane Authors.

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
	"github.com/xanzy/go-gitlab"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
)

var (
	groupID                  = 0
	userID                   = 0
	accessLevel              = 10
	expiresAt                = "2021-05-04"
	v1alpha1AccessLevelValue = v1alpha1.AccessLevelValue(accessLevel)
	gitlabAccessLevelValue   = gitlab.AccessLevelValue(accessLevel)
)

func TestGenerateGroupMemberObservation(t *testing.T) {
	username := "User Name"
	state := "State"
	avatarURL := "Avatar URL"
	webURL := "Web URL"
	externUID := "ExternUID"
	provider := "Provider"
	samlProviderID := 0
	v1alpha1GroupSAMLIdentity := v1alpha1.GroupMemberSAMLIdentity{
		ExternUID:      externUID,
		Provider:       provider,
		SAMLProviderID: samlProviderID,
	}
	gitlabGroupSAMLIdentity := gitlab.GroupMemberSAMLIdentity{
		ExternUID:      externUID,
		Provider:       provider,
		SAMLProviderID: samlProviderID,
	}
	name := "Name"
	type args struct {
		p *gitlab.GroupMember
	}
	cases := map[string]struct {
		args args
		want v1alpha1.GroupMemberObservation
	}{
		"Full": {
			args: args{
				p: &gitlab.GroupMember{
					Username:          username,
					Name:              name,
					State:             state,
					AvatarURL:         avatarURL,
					WebURL:            webURL,
					GroupSAMLIdentity: &gitlabGroupSAMLIdentity,
				},
			},
			want: v1alpha1.GroupMemberObservation{
				Username:          username,
				Name:              name,
				State:             state,
				AvatarURL:         avatarURL,
				WebURL:            webURL,
				GroupSAMLIdentity: &v1alpha1GroupSAMLIdentity,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateGroupMemberObservation(tc.args.p)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateAddGroupMemberOptions(t *testing.T) {
	type args struct {
		parameters *v1alpha1.GroupMemberParameters
	}
	cases := map[string]struct {
		args args
		want *gitlab.AddGroupMemberOptions
	}{
		"AllFields": {
			args: args{
				parameters: &v1alpha1.GroupMemberParameters{
					GroupID:     groupID,
					UserID:      userID,
					AccessLevel: v1alpha1AccessLevelValue,
					ExpiresAt:   &expiresAt,
				},
			},
			want: &gitlab.AddGroupMemberOptions{
				UserID:      &userID,
				AccessLevel: &gitlabAccessLevelValue,
				ExpiresAt:   &expiresAt,
			},
		},
		"SomeFields": {
			args: args{
				parameters: &v1alpha1.GroupMemberParameters{
					GroupID:     groupID,
					UserID:      userID,
					AccessLevel: v1alpha1.AccessLevelValue(accessLevel),
				},
			},
			want: &gitlab.AddGroupMemberOptions{
				UserID:      &userID,
				AccessLevel: &gitlabAccessLevelValue,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateAddGroupMemberOptions(tc.args.parameters)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateEditGroupMemberOptions(t *testing.T) {
	type args struct {
		parameters *v1alpha1.GroupMemberParameters
	}
	cases := map[string]struct {
		args args
		want *gitlab.EditGroupMemberOptions
	}{
		"AllFields": {
			args: args{
				parameters: &v1alpha1.GroupMemberParameters{
					GroupID:     groupID,
					UserID:      userID,
					AccessLevel: v1alpha1AccessLevelValue,
					ExpiresAt:   &expiresAt,
				},
			},
			want: &gitlab.EditGroupMemberOptions{
				AccessLevel: &gitlabAccessLevelValue,
				ExpiresAt:   &expiresAt,
			},
		},
		"SomeFields": {
			args: args{
				parameters: &v1alpha1.GroupMemberParameters{
					GroupID:     groupID,
					UserID:      userID,
					AccessLevel: v1alpha1AccessLevelValue,
				},
			},
			want: &gitlab.EditGroupMemberOptions{
				AccessLevel: &gitlabAccessLevelValue,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateEditGroupMemberOptions(tc.args.parameters)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
