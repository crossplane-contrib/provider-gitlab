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
	gitlab "gitlab.com/gitlab-org/api/client-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-gitlab/apis/cluster/projects/v1alpha1"
)

var (
	projectID                = 0
	userID                   = 0
	accessLevel              = 10
	expiresAt                = "2021-05-04"
	email                    = "simpleemail@gmail.com"
	v1alpha1AccessLevelValue = v1alpha1.AccessLevelValue(accessLevel)
	gitlabAccessLevelValue   = gitlab.AccessLevelValue(accessLevel)
	createdAt                = time.Now()
)

func TestGenerateMemberObservation(t *testing.T) {
	username := "User Name"
	state := "State"
	avatarURL := "Avatar URL"
	webURL := "Web URL"
	name := "Name"
	type args struct {
		p *gitlab.ProjectMember
	}
	cases := map[string]struct {
		args args
		want v1alpha1.MemberObservation
	}{
		"Full": {
			args: args{
				p: &gitlab.ProjectMember{
					Username:  username,
					Name:      name,
					Email:     email,
					State:     state,
					CreatedAt: &createdAt,
					AvatarURL: avatarURL,
					WebURL:    webURL,
				},
			},
			want: v1alpha1.MemberObservation{
				Username:  username,
				Name:      name,
				Email:     email,
				State:     state,
				CreatedAt: &metav1.Time{Time: createdAt},
				AvatarURL: avatarURL,
				WebURL:    webURL,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateMemberObservation(tc.args.p)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateAddMemberOptions(t *testing.T) {
	type args struct {
		parameters *v1alpha1.MemberParameters
	}
	cases := map[string]struct {
		args args
		want *gitlab.AddProjectMemberOptions
	}{
		"AllFields": {
			args: args{
				parameters: &v1alpha1.MemberParameters{
					ProjectID:   &projectID,
					UserID:      &userID,
					AccessLevel: v1alpha1AccessLevelValue,
					ExpiresAt:   &expiresAt,
				},
			},
			want: &gitlab.AddProjectMemberOptions{
				UserID:      &userID,
				AccessLevel: &gitlabAccessLevelValue,
				ExpiresAt:   &expiresAt,
			},
		},
		"SomeFields": {
			args: args{
				parameters: &v1alpha1.MemberParameters{
					ProjectID:   &projectID,
					UserID:      &userID,
					AccessLevel: v1alpha1.AccessLevelValue(accessLevel),
				},
			},
			want: &gitlab.AddProjectMemberOptions{
				UserID:      &userID,
				AccessLevel: &gitlabAccessLevelValue,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateAddMemberOptions(tc.args.parameters)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateEditMemberOptions(t *testing.T) {
	type args struct {
		parameters *v1alpha1.MemberParameters
	}
	cases := map[string]struct {
		args args
		want *gitlab.EditProjectMemberOptions
	}{
		"AllFields": {
			args: args{
				parameters: &v1alpha1.MemberParameters{
					ProjectID:   &projectID,
					UserID:      &userID,
					AccessLevel: v1alpha1AccessLevelValue,
					ExpiresAt:   &expiresAt,
				},
			},
			want: &gitlab.EditProjectMemberOptions{
				AccessLevel: &gitlabAccessLevelValue,
				ExpiresAt:   &expiresAt,
			},
		},
		"SomeFields": {
			args: args{
				parameters: &v1alpha1.MemberParameters{
					ProjectID:   &projectID,
					UserID:      &userID,
					AccessLevel: v1alpha1AccessLevelValue,
				},
			},
			want: &gitlab.EditProjectMemberOptions{
				AccessLevel: &gitlabAccessLevelValue,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateEditMemberOptions(tc.args.parameters)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
