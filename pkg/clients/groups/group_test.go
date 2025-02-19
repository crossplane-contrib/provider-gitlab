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

package groups

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/gitlab-org/api/client-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
)

var (
	name                           = "example-group"
	path                           = "example/path/to/group"
	description                    = "group description"
	ID                             = 123456
	membershipLock                 = true
	visibility                     = "private"
	v1alpha1Visibility             = v1alpha1.VisibilityValue(visibility)
	gitlabVisibility               = gitlab.VisibilityValue(visibility)
	shareWithGroupLock             = true
	requireTwoFactorAuth           = false
	twoFactorGracePeriod           = 48
	projectCreationLevel           = "developer"
	v1alpha1ProjectCreationLevel   = v1alpha1.ProjectCreationLevelValue(projectCreationLevel)
	gitlabProjectCreationLevel     = gitlab.ProjectCreationLevelValue(projectCreationLevel)
	autoDevopsEnabled              = true
	subGroupCreationLevel          = "maintainer"
	v1alpha1SubGroupCreationLevel  = v1alpha1.SubGroupCreationLevelValue(subGroupCreationLevel)
	gitlabSubGroupCreationLevel    = gitlab.SubGroupCreationLevelValue(subGroupCreationLevel)
	emailsEnabled                  = true
	mentionsDisabled               = true
	LFSEnabled                     = true
	requestAccessEnabled           = true
	parentID                       = 0
	parentIDint                    = 0
	sharedRunnersMinutesLimit      = 0
	extraSharedRunnersMinutesLimit = 0
	storageSize                    = int64(10)
	repositorySize                 = int64(20)
	lfsObjectsSize                 = int64(30)
	jobArtifactsSize               = int64(40)
	v1alpha1Statistics             = v1alpha1.StorageStatistics{
		StorageSize:      storageSize,
		RepositorySize:   repositorySize,
		LfsObjectsSize:   lfsObjectsSize,
		JobArtifactsSize: jobArtifactsSize,
	}
	gitlabStatistics = gitlab.Statistics{
		StorageSize:      storageSize,
		RepositorySize:   repositorySize,
		LFSObjectsSize:   lfsObjectsSize,
		JobArtifactsSize: jobArtifactsSize,
	}
	LDAPAccess       = 0
	groupAccessLevel = 50
	gitlabLDAPAccess = gitlab.AccessLevelValue(LDAPAccess)
)

func TestGenerateObservation(t *testing.T) {
	id := 0
	webURL := "web.url"
	fullName := "Full name"
	fullPath := "Full path"
	now := time.Now()
	customAttributeKey := "Key_1"
	customAttributeValue := "Value_1"
	i := 0
	s := ""

	v1alpha1MarkedForDeletionOn := &metav1.Time{Time: now}
	gitlabMarkedForDeletionOn := gitlab.ISOTime(now)
	v1alpha1CreatedAt := &metav1.Time{Time: now}
	gitlabCreatedAt := now
	gitlabSharedWithGroupsExpireAt := gitlab.ISOTime(now)

	sharedWithGroups := []v1alpha1.SharedWithGroups{
		{
			GroupID:          &ID,
			GroupAccessLevel: 1,
			ExpiresAt:        &metav1.Time{Time: now},
		},
	}
	type args struct {
		p *gitlab.Group
	}
	cases := map[string]struct {
		args args
		want v1alpha1.GroupObservation
	}{
		"Full": {
			args: args{
				p: &gitlab.Group{
					ID:         id,
					WebURL:     webURL,
					FullName:   fullName,
					FullPath:   fullPath,
					Statistics: &gitlabStatistics,
					CustomAttributes: []*gitlab.CustomAttribute{
						{
							Key:   customAttributeKey,
							Value: customAttributeValue,
						},
					},
					SharedWithGroups: []struct {
						GroupID          int             `json:"group_id"`
						GroupName        string          `json:"group_name"`
						GroupFullPath    string          `json:"group_full_path"`
						GroupAccessLevel int             `json:"group_access_level"`
						ExpiresAt        *gitlab.ISOTime `json:"expires_at"`
						MemberRoleID     int             `json:"member_role_id"`
					}{
						{
							GroupID:          ID,
							GroupName:        name,
							GroupFullPath:    path,
							GroupAccessLevel: sharedWithGroups[0].GroupAccessLevel,
							ExpiresAt:        &gitlabSharedWithGroupsExpireAt,
							MemberRoleID:     0,
						},
					},
					LDAPAccess: gitlabLDAPAccess,
					LDAPGroupLinks: []*gitlab.LDAPGroupLink{
						{
							CN:          "CN",
							GroupAccess: gitlab.AccessLevelValue(groupAccessLevel),
							Provider:    "Provider",
						},
					},
					MarkedForDeletionOn: &gitlabMarkedForDeletionOn,
					CreatedAt:           &gitlabCreatedAt,
				},
			},
			want: v1alpha1.GroupObservation{
				ID:         &id,
				AvatarURL:  &s,
				WebURL:     &webURL,
				FullName:   &fullName,
				FullPath:   &fullPath,
				Statistics: &v1alpha1Statistics,
				CustomAttributes: []v1alpha1.CustomAttribute{
					{
						Key:   customAttributeKey,
						Value: customAttributeValue,
					},
				},
				LDAPCN:     &s,
				LDAPAccess: nil,
				LDAPGroupLinks: []v1alpha1.LDAPGroupLink{
					{
						CN:          "CN",
						GroupAccess: v1alpha1.AccessLevelValue(groupAccessLevel),
						Provider:    "Provider",
					},
				},
				MarkedForDeletionOn: v1alpha1MarkedForDeletionOn,
				CreatedAt:           v1alpha1CreatedAt,
				SharedWithGroups: []v1alpha1.SharedWithGroupsObservation{
					{
						GroupID:          &ID,
						GroupName:        &name,
						GroupFullPath:    &path,
						GroupAccessLevel: &sharedWithGroups[0].GroupAccessLevel,
						ExpiresAt:        &metav1.Time{Time: time.Time(gitlabSharedWithGroupsExpireAt)},
					},
				},
			},
		},
		"SharedWithGroupExpiresAtIsNil": {
			args: args{
				p: &gitlab.Group{
					SharedWithGroups: []struct {
						GroupID          int             `json:"group_id"`
						GroupName        string          `json:"group_name"`
						GroupFullPath    string          `json:"group_full_path"`
						GroupAccessLevel int             `json:"group_access_level"`
						ExpiresAt        *gitlab.ISOTime `json:"expires_at"`
						MemberRoleID     int             `json:"member_role_id"`
					}{
						{
							ExpiresAt: nil,
						},
					},
				},
			},
			want: v1alpha1.GroupObservation{
				ID:        &i,
				AvatarURL: &s,
				WebURL:    &s,
				FullName:  &s,
				FullPath:  &s,
				LDAPCN:    &s,

				SharedWithGroups: []v1alpha1.SharedWithGroupsObservation{{
					GroupID:          &i,
					GroupName:        &s,
					GroupFullPath:    &s,
					GroupAccessLevel: &i,
					ExpiresAt:        nil,
				}},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateObservation(tc.args.p)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateCreateGroupOptions(t *testing.T) {
	type args struct {
		name       string
		parameters *v1alpha1.GroupParameters
	}
	cases := map[string]struct {
		args args
		want *gitlab.CreateGroupOptions
	}{
		"AllFields": {
			args: args{
				name: name,
				parameters: &v1alpha1.GroupParameters{
					Path:                           path,
					Description:                    &description,
					MembershipLock:                 &membershipLock,
					Visibility:                     &v1alpha1Visibility,
					ShareWithGroupLock:             &shareWithGroupLock,
					RequireTwoFactorAuth:           &requireTwoFactorAuth,
					TwoFactorGracePeriod:           &twoFactorGracePeriod,
					ProjectCreationLevel:           &v1alpha1ProjectCreationLevel,
					AutoDevopsEnabled:              &autoDevopsEnabled,
					SubGroupCreationLevel:          &v1alpha1SubGroupCreationLevel,
					EmailsEnabled:                  &emailsEnabled,
					MentionsDisabled:               &mentionsDisabled,
					LFSEnabled:                     &LFSEnabled,
					RequestAccessEnabled:           &requestAccessEnabled,
					ParentID:                       &parentID,
					SharedRunnersMinutesLimit:      &sharedRunnersMinutesLimit,
					ExtraSharedRunnersMinutesLimit: &extraSharedRunnersMinutesLimit,
				},
			},
			want: &gitlab.CreateGroupOptions{
				Name:                           &name,
				Path:                           &path,
				Description:                    &description,
				MembershipLock:                 &membershipLock,
				Visibility:                     &gitlabVisibility,
				ShareWithGroupLock:             &shareWithGroupLock,
				RequireTwoFactorAuth:           &requireTwoFactorAuth,
				TwoFactorGracePeriod:           &twoFactorGracePeriod,
				ProjectCreationLevel:           &gitlabProjectCreationLevel,
				AutoDevopsEnabled:              &autoDevopsEnabled,
				SubGroupCreationLevel:          &gitlabSubGroupCreationLevel,
				EmailsEnabled:                  &emailsEnabled,
				MentionsDisabled:               &mentionsDisabled,
				LFSEnabled:                     &LFSEnabled,
				RequestAccessEnabled:           &requestAccessEnabled,
				ParentID:                       &parentIDint,
				SharedRunnersMinutesLimit:      &sharedRunnersMinutesLimit,
				ExtraSharedRunnersMinutesLimit: &extraSharedRunnersMinutesLimit,
			},
		},
		"SomeFields": {
			args: args{
				name: name,
				parameters: &v1alpha1.GroupParameters{
					Path:               path,
					Description:        &description,
					MembershipLock:     &membershipLock,
					Visibility:         &v1alpha1Visibility,
					ShareWithGroupLock: &shareWithGroupLock,
				},
			},
			want: &gitlab.CreateGroupOptions{
				Name:               &name,
				Path:               &path,
				Description:        &description,
				MembershipLock:     &membershipLock,
				Visibility:         &gitlabVisibility,
				ShareWithGroupLock: &shareWithGroupLock,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateCreateGroupOptions(tc.args.name, tc.args.parameters)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateEditGroupOptions(t *testing.T) {
	type args struct {
		name       string
		parameters *v1alpha1.GroupParameters
	}
	cases := map[string]struct {
		args args
		want *gitlab.UpdateGroupOptions
	}{
		"AllFields": {
			args: args{
				name: name,
				parameters: &v1alpha1.GroupParameters{
					Path:                           path,
					Description:                    &description,
					MembershipLock:                 &membershipLock,
					Visibility:                     &v1alpha1Visibility,
					ShareWithGroupLock:             &shareWithGroupLock,
					RequireTwoFactorAuth:           &requireTwoFactorAuth,
					TwoFactorGracePeriod:           &twoFactorGracePeriod,
					ProjectCreationLevel:           &v1alpha1ProjectCreationLevel,
					AutoDevopsEnabled:              &autoDevopsEnabled,
					SubGroupCreationLevel:          &v1alpha1SubGroupCreationLevel,
					MentionsDisabled:               &mentionsDisabled,
					EmailsEnabled:                  &emailsEnabled,
					LFSEnabled:                     &LFSEnabled,
					RequestAccessEnabled:           &requestAccessEnabled,
					ParentID:                       &parentID,
					SharedRunnersMinutesLimit:      &sharedRunnersMinutesLimit,
					ExtraSharedRunnersMinutesLimit: &extraSharedRunnersMinutesLimit,
				},
			},
			want: &gitlab.UpdateGroupOptions{
				Name:                           &name,
				Path:                           &path,
				Description:                    &description,
				MembershipLock:                 &membershipLock,
				Visibility:                     &gitlabVisibility,
				ShareWithGroupLock:             &shareWithGroupLock,
				RequireTwoFactorAuth:           &requireTwoFactorAuth,
				TwoFactorGracePeriod:           &twoFactorGracePeriod,
				ProjectCreationLevel:           &gitlabProjectCreationLevel,
				AutoDevopsEnabled:              &autoDevopsEnabled,
				SubGroupCreationLevel:          &gitlabSubGroupCreationLevel,
				EmailsEnabled:                  &emailsEnabled,
				MentionsDisabled:               &mentionsDisabled,
				LFSEnabled:                     &LFSEnabled,
				RequestAccessEnabled:           &requestAccessEnabled,
				SharedRunnersMinutesLimit:      &sharedRunnersMinutesLimit,
				ExtraSharedRunnersMinutesLimit: &extraSharedRunnersMinutesLimit,
			},
		},
		"SomeFields": {
			args: args{
				name: name,
				parameters: &v1alpha1.GroupParameters{
					Path:          path,
					Description:   &description,
					Visibility:    &v1alpha1Visibility,
					EmailsEnabled: &emailsEnabled,
				},
			},
			want: &gitlab.UpdateGroupOptions{
				Name:          &name,
				Path:          &path,
				Description:   &description,
				Visibility:    &gitlabVisibility,
				EmailsEnabled: &emailsEnabled,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateEditGroupOptions(tc.args.name, tc.args.parameters)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
