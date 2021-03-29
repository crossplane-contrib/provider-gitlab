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
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
)

var (
	name                           = "example-group"
	path                           = "example/path/to/group"
	description                    = "group description"
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
	emailsDisabled                 = true
	mentionsDisabled               = true
	LFSEnabled                     = true
	requestAccessEnabled           = true
	parentID                       = 0
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
	gitlabStatistics = gitlab.StorageStatistics{
		StorageSize:      storageSize,
		RepositorySize:   repositorySize,
		LfsObjectsSize:   lfsObjectsSize,
		JobArtifactsSize: jobArtifactsSize,
	}
	LDAPAccess         = 0
	groupAccessLevel   = 50
	v1alpha1LDAPAccess = v1alpha1.AccessLevelValue(LDAPAccess)
	gitlabLDAPAccess   = gitlab.AccessLevelValue(LDAPAccess)
)

func TestGenerateObservation(t *testing.T) {
	id := 0
	webURL := "web.url"
	fullName := "Full name"
	fullPath := "Full path"
	now := time.Now()
	customAttributeKey := "Key_1"
	customAttributeValue := "Value_1"

	v1alpha1MarkedForDeletionOn := &metav1.Time{Time: now}
	gitlabMarkedForDeletionOn := gitlab.ISOTime(now)
	v1alpha1CreatedAt := &metav1.Time{Time: now}
	gitlabCreatedAt := now
	gitlabSharedWithGroupsExpireAt := gitlab.ISOTime(now)

	sharedWithGroups := []v1alpha1.SharedWithGroups{
		{
			GroupID:          0,
			GroupName:        "group name",
			GroupFullPath:    "group full path",
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
					}{
						{
							GroupID:          sharedWithGroups[0].GroupID,
							GroupName:        sharedWithGroups[0].GroupName,
							GroupFullPath:    sharedWithGroups[0].GroupFullPath,
							GroupAccessLevel: sharedWithGroups[0].GroupAccessLevel,
							ExpiresAt:        &gitlabSharedWithGroupsExpireAt,
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
				ID:         id,
				WebURL:     webURL,
				FullName:   fullName,
				FullPath:   fullPath,
				Statistics: &v1alpha1Statistics,
				CustomAttributes: []v1alpha1.CustomAttribute{
					{
						Key:   customAttributeKey,
						Value: customAttributeValue,
					},
				},
				SharedWithGroups: sharedWithGroups,
				LDAPAccess:       v1alpha1LDAPAccess,
				LDAPGroupLinks: []v1alpha1.LDAPGroupLink{
					{
						CN:          "CN",
						GroupAccess: v1alpha1.AccessLevelValue(groupAccessLevel),
						Provider:    "Provider",
					},
				},
				MarkedForDeletionOn: v1alpha1MarkedForDeletionOn,
				CreatedAt:           v1alpha1CreatedAt,
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

func TestLateInitialize(t *testing.T) {
	cases := map[string]struct {
		parameters *v1alpha1.GroupParameters
		group      *gitlab.Group
		want       *v1alpha1.GroupParameters
	}{
		"AllOptionalFields": {
			parameters: &v1alpha1.GroupParameters{},
			group: &gitlab.Group{
				Path:                           path,
				Description:                    description,
				MembershipLock:                 membershipLock,
				Visibility:                     gitlabVisibility,
				ShareWithGroupLock:             shareWithGroupLock,
				RequireTwoFactorAuth:           requireTwoFactorAuth,
				TwoFactorGracePeriod:           twoFactorGracePeriod,
				ProjectCreationLevel:           gitlabProjectCreationLevel,
				AutoDevopsEnabled:              autoDevopsEnabled,
				SubGroupCreationLevel:          gitlabSubGroupCreationLevel,
				EmailsDisabled:                 emailsDisabled,
				MentionsDisabled:               mentionsDisabled,
				LFSEnabled:                     LFSEnabled,
				RequestAccessEnabled:           requestAccessEnabled,
				ParentID:                       parentID,
				SharedRunnersMinutesLimit:      sharedRunnersMinutesLimit,
				ExtraSharedRunnersMinutesLimit: extraSharedRunnersMinutesLimit,
			},
			want: &v1alpha1.GroupParameters{
				Path:                           &path,
				Description:                    &description,
				MembershipLock:                 &membershipLock,
				Visibility:                     &v1alpha1Visibility,
				ShareWithGroupLock:             &shareWithGroupLock,
				RequireTwoFactorAuth:           &requireTwoFactorAuth,
				TwoFactorGracePeriod:           &twoFactorGracePeriod,
				ProjectCreationLevel:           &v1alpha1ProjectCreationLevel,
				AutoDevopsEnabled:              &autoDevopsEnabled,
				SubGroupCreationLevel:          &v1alpha1SubGroupCreationLevel,
				EmailsDisabled:                 &emailsDisabled,
				MentionsDisabled:               &mentionsDisabled,
				LFSEnabled:                     &LFSEnabled,
				RequestAccessEnabled:           &requestAccessEnabled,
				ParentID:                       &parentID,
				SharedRunnersMinutesLimit:      &sharedRunnersMinutesLimit,
				ExtraSharedRunnersMinutesLimit: &extraSharedRunnersMinutesLimit,
			},
		},
		"SomeFieldsDontOverwrite": {
			parameters: &v1alpha1.GroupParameters{
				Path:                 &path,
				Description:          &description,
				ProjectCreationLevel: &v1alpha1ProjectCreationLevel,
				RequestAccessEnabled: &requestAccessEnabled,
				ParentID:             &parentID,
			},
			group: &gitlab.Group{
				Path:                           path,
				Description:                    description,
				MembershipLock:                 membershipLock,
				Visibility:                     gitlabVisibility,
				ShareWithGroupLock:             shareWithGroupLock,
				RequireTwoFactorAuth:           requireTwoFactorAuth,
				TwoFactorGracePeriod:           twoFactorGracePeriod,
				ProjectCreationLevel:           gitlabProjectCreationLevel,
				AutoDevopsEnabled:              autoDevopsEnabled,
				SubGroupCreationLevel:          gitlabSubGroupCreationLevel,
				EmailsDisabled:                 emailsDisabled,
				MentionsDisabled:               mentionsDisabled,
				LFSEnabled:                     LFSEnabled,
				RequestAccessEnabled:           requestAccessEnabled,
				ParentID:                       parentID,
				SharedRunnersMinutesLimit:      sharedRunnersMinutesLimit,
				ExtraSharedRunnersMinutesLimit: extraSharedRunnersMinutesLimit,
			},
			want: &v1alpha1.GroupParameters{
				Path:                           &path,
				Description:                    &description,
				MembershipLock:                 &membershipLock,
				Visibility:                     &v1alpha1Visibility,
				ShareWithGroupLock:             &shareWithGroupLock,
				RequireTwoFactorAuth:           &requireTwoFactorAuth,
				TwoFactorGracePeriod:           &twoFactorGracePeriod,
				ProjectCreationLevel:           &v1alpha1ProjectCreationLevel,
				AutoDevopsEnabled:              &autoDevopsEnabled,
				SubGroupCreationLevel:          &v1alpha1SubGroupCreationLevel,
				EmailsDisabled:                 &emailsDisabled,
				MentionsDisabled:               &mentionsDisabled,
				LFSEnabled:                     &LFSEnabled,
				RequestAccessEnabled:           &requestAccessEnabled,
				ParentID:                       &parentID,
				SharedRunnersMinutesLimit:      &sharedRunnersMinutesLimit,
				ExtraSharedRunnersMinutesLimit: &extraSharedRunnersMinutesLimit,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitialize(tc.parameters, tc.group)
			if diff := cmp.Diff(tc.want, tc.parameters); diff != "" {
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
					Path:                           &path,
					Description:                    &description,
					MembershipLock:                 &membershipLock,
					Visibility:                     &v1alpha1Visibility,
					ShareWithGroupLock:             &shareWithGroupLock,
					RequireTwoFactorAuth:           &requireTwoFactorAuth,
					TwoFactorGracePeriod:           &twoFactorGracePeriod,
					ProjectCreationLevel:           &v1alpha1ProjectCreationLevel,
					AutoDevopsEnabled:              &autoDevopsEnabled,
					SubGroupCreationLevel:          &v1alpha1SubGroupCreationLevel,
					EmailsDisabled:                 &emailsDisabled,
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
				EmailsDisabled:                 &emailsDisabled,
				MentionsDisabled:               &mentionsDisabled,
				LFSEnabled:                     &LFSEnabled,
				RequestAccessEnabled:           &requestAccessEnabled,
				ParentID:                       &parentID,
				SharedRunnersMinutesLimit:      &sharedRunnersMinutesLimit,
				ExtraSharedRunnersMinutesLimit: &extraSharedRunnersMinutesLimit,
			},
		},
		"SomeFields": {
			args: args{
				name: name,
				parameters: &v1alpha1.GroupParameters{
					Path:               &path,
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
					Path:                           &path,
					Description:                    &description,
					MembershipLock:                 &membershipLock,
					Visibility:                     &v1alpha1Visibility,
					ShareWithGroupLock:             &shareWithGroupLock,
					RequireTwoFactorAuth:           &requireTwoFactorAuth,
					TwoFactorGracePeriod:           &twoFactorGracePeriod,
					ProjectCreationLevel:           &v1alpha1ProjectCreationLevel,
					AutoDevopsEnabled:              &autoDevopsEnabled,
					SubGroupCreationLevel:          &v1alpha1SubGroupCreationLevel,
					EmailsDisabled:                 &emailsDisabled,
					MentionsDisabled:               &mentionsDisabled,
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
				EmailsDisabled:                 &emailsDisabled,
				MentionsDisabled:               &mentionsDisabled,
				LFSEnabled:                     &LFSEnabled,
				RequestAccessEnabled:           &requestAccessEnabled,
				ParentID:                       &parentID,
				SharedRunnersMinutesLimit:      &sharedRunnersMinutesLimit,
				ExtraSharedRunnersMinutesLimit: &extraSharedRunnersMinutesLimit,
			},
		},
		"SomeFields": {
			args: args{
				name: name,
				parameters: &v1alpha1.GroupParameters{
					Path:           &path,
					Description:    &description,
					Visibility:     &v1alpha1Visibility,
					EmailsDisabled: &emailsDisabled,
				},
			},
			want: &gitlab.UpdateGroupOptions{
				Name:           &name,
				Path:           &path,
				Description:    &description,
				Visibility:     &gitlabVisibility,
				EmailsDisabled: &emailsDisabled,
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

func TestIsGroupUpToDate(t *testing.T) {
	type args struct {
		group *gitlab.Group
		p     *v1alpha1.GroupParameters
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				p: &v1alpha1.GroupParameters{
					Path:                           &path,
					Description:                    &description,
					MembershipLock:                 &membershipLock,
					Visibility:                     &v1alpha1Visibility,
					ShareWithGroupLock:             &shareWithGroupLock,
					RequireTwoFactorAuth:           &requireTwoFactorAuth,
					TwoFactorGracePeriod:           &twoFactorGracePeriod,
					ProjectCreationLevel:           &v1alpha1ProjectCreationLevel,
					AutoDevopsEnabled:              &autoDevopsEnabled,
					SubGroupCreationLevel:          &v1alpha1SubGroupCreationLevel,
					EmailsDisabled:                 &emailsDisabled,
					MentionsDisabled:               &mentionsDisabled,
					LFSEnabled:                     &LFSEnabled,
					RequestAccessEnabled:           &requestAccessEnabled,
					ParentID:                       &parentID,
					SharedRunnersMinutesLimit:      &sharedRunnersMinutesLimit,
					ExtraSharedRunnersMinutesLimit: &extraSharedRunnersMinutesLimit,
				},
				group: &gitlab.Group{
					Path:                           path,
					Description:                    description,
					MembershipLock:                 membershipLock,
					Visibility:                     gitlabVisibility,
					ShareWithGroupLock:             shareWithGroupLock,
					RequireTwoFactorAuth:           requireTwoFactorAuth,
					TwoFactorGracePeriod:           twoFactorGracePeriod,
					ProjectCreationLevel:           gitlabProjectCreationLevel,
					AutoDevopsEnabled:              autoDevopsEnabled,
					SubGroupCreationLevel:          gitlabSubGroupCreationLevel,
					EmailsDisabled:                 emailsDisabled,
					MentionsDisabled:               mentionsDisabled,
					LFSEnabled:                     LFSEnabled,
					RequestAccessEnabled:           requestAccessEnabled,
					ParentID:                       parentID,
					SharedRunnersMinutesLimit:      sharedRunnersMinutesLimit,
					ExtraSharedRunnersMinutesLimit: extraSharedRunnersMinutesLimit,
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				p: &v1alpha1.GroupParameters{
					Path:                           &path,
					Description:                    &description,
					MembershipLock:                 &membershipLock,
					Visibility:                     &v1alpha1Visibility,
					ShareWithGroupLock:             &shareWithGroupLock,
					RequireTwoFactorAuth:           &requireTwoFactorAuth,
					TwoFactorGracePeriod:           &twoFactorGracePeriod,
					ProjectCreationLevel:           &v1alpha1ProjectCreationLevel,
					AutoDevopsEnabled:              &autoDevopsEnabled,
					SubGroupCreationLevel:          &v1alpha1SubGroupCreationLevel,
					EmailsDisabled:                 &emailsDisabled,
					MentionsDisabled:               &mentionsDisabled,
					LFSEnabled:                     &LFSEnabled,
					RequestAccessEnabled:           &requestAccessEnabled,
					ParentID:                       &parentID,
					SharedRunnersMinutesLimit:      &sharedRunnersMinutesLimit,
					ExtraSharedRunnersMinutesLimit: &extraSharedRunnersMinutesLimit,
				},
				group: &gitlab.Group{
					Path:                           "/new/path/to/group",
					Description:                    "new descriptions",
					MembershipLock:                 membershipLock,
					Visibility:                     gitlabVisibility,
					ShareWithGroupLock:             shareWithGroupLock,
					RequireTwoFactorAuth:           requireTwoFactorAuth,
					TwoFactorGracePeriod:           twoFactorGracePeriod,
					ProjectCreationLevel:           gitlabProjectCreationLevel,
					AutoDevopsEnabled:              autoDevopsEnabled,
					SubGroupCreationLevel:          gitlabSubGroupCreationLevel,
					EmailsDisabled:                 emailsDisabled,
					MentionsDisabled:               mentionsDisabled,
					LFSEnabled:                     LFSEnabled,
					RequestAccessEnabled:           requestAccessEnabled,
					ParentID:                       parentID,
					SharedRunnersMinutesLimit:      sharedRunnersMinutesLimit,
					ExtraSharedRunnersMinutesLimit: extraSharedRunnersMinutesLimit,
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsGroupUpToDate(tc.args.p, tc.args.group)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
