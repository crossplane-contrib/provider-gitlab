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

package projects

import (
	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

var (
	name                                      = "my-project"
	path                                      = "path/to/project"
	defaultBranch                             = "main"
	description                               = "my awesome project"
	issuesAccessLevel                         = "enabled"
	issuesAccessLevelv1alpha1                 = v1alpha1.AccessControlValue(issuesAccessLevel)
	repositoryAccessLevel                     = "enabled"
	repositoryAccessLevelv1alpha1             = v1alpha1.AccessControlValue(repositoryAccessLevel)
	mergeRequestsAccessLevel                  = "enabled"
	mergeRequestsAccessLevelv1alpha1          = v1alpha1.AccessControlValue(mergeRequestsAccessLevel)
	forkingAccessLevel                        = "enabled"
	forkingAccessLevelv1alpha1                = v1alpha1.AccessControlValue(forkingAccessLevel)
	buildsAccessLevel                         = "disabled"
	buildsAccessLevelv1alpha1                 = v1alpha1.AccessControlValue(buildsAccessLevel)
	wikiAccessLevel                           = "private"
	wikiAccessLevelv1alpha1                   = v1alpha1.AccessControlValue(wikiAccessLevel)
	snippetsAccessLevel                       = "public"
	snippetsAccessLevelv1alpha1               = v1alpha1.AccessControlValue(snippetsAccessLevel)
	pagesAccessLevel                          = "enabled"
	pagesAccessLevelv1alpha1                  = v1alpha1.AccessControlValue(pagesAccessLevel)
	resolveOutdatedDiffDiscussions            = true
	containerRegistryEnabled                  = true
	sharedRunnersEnabled                      = true
	visibility                                = "private"
	visibilityv1alpha1                        = v1alpha1.VisibilityValue(visibility)
	publicBuilds                              = false
	onlyAllowMergeIfPipelineSucceeds          = true
	OnlyAllowMergeIfAllDiscussionsAreResolved = true
	mergeMethod                               = "merge"
	mergeMethodv1alpha1                       = v1alpha1.MergeMethodValue(mergeMethod)
	removeSourceBranchAfterMerge              = false
	lfsEnabled                                = true
	requestAccessEnabled                      = true
	tagList                                   = []string{"tag1", "tag2"}
	buildGitStategy                           = "strategy"
	buildTimeout                              = 60
	ciConfigPath                              = "path/to/ci/config"
	ciDefaultGitDepth                         = 50
	approvalsBeforeMerge                      = 0
	externalAuthorizationClassificationLabel  = "authz-label"
	mirror                                    = false
	mirrorUserID                              = 1
	mirrorTriggerBuilds                       = true
	onlyMirrorProtectedBranches               = false
	mirrorOverwritesDivergedBranches          = false
	packagesEnabled                           = true
	serviceDeskEnabled                        = true
	autocloseReferencedIssues                 = true
)

func TestGenerateObservation(t *testing.T) {
	id := 0
	public := true
	sshURLToRepo := "ssh:url"
	httpURLToRepo := "http://url"
	webURL := "web.url"
	readmeURL := "readme.url"
	owner := "chief"
	pathWithNamespace := "path/to/cool-project"
	issuesEnabled := true
	openIssuesCount := 3
	mergeRequestsEnabled := true
	jobsEnabled := false
	wikiEnabled := false
	snippetsEnabled := true
	now := time.Now()
	creatorID := 1
	namespaceId := 3
	importStatus := "foo"
	importError := "none"
	permissionsProjectAccessAccessLevel := 1
	permissionsProjectAccessNotificationLevel := 2
	permissionsGroupAccessAccessLevel := 3
	permissionsGroupAccessNotificationLevel := 4
	markedForDeletionAt := gitlab.ISOTime(now)
	archived := false
	forksCount := 2
	starCount := 10000
	// TODO(knappek): verify that Kubernetes secret has been created
	//runnersToken := "secrettoken"
	forkedFromProjectHTTPURL := "http://fork.url"
	sharedWithGroups := []struct {
		GroupID          int    `json:"group_id"`
		GroupName        string `json:"group_name"`
		GroupAccessLevel int    `json:"group_access_level"`
	}{
		{
			GroupID:          0,
			GroupName:        "sharedgroup",
			GroupAccessLevel: 1,
		},
	}
	storageStatistics := struct {
		StorageSize      int64 `json:"storage_size"`
		RepositorySize   int64 `json:"repository_size"`
		LfsObjectsSize   int64 `json:"lfs_objects_size"`
		JobArtifactsSize int64 `json:"job_artifacts_size"`
	}{
		StorageSize:      10,
		RepositorySize:   20,
		LfsObjectsSize:   30,
		JobArtifactsSize: 40,
	}
	projectStatisticsCommitCount := 0
	linksSelf := "selflink"
	customAttributesKey := "customAttrKey"
	customAttributesValue := "customAttrValue"
	complianceFrameworks := []string{"framework1", "framework2"}

	type args struct {
		p *gitlab.Project
	}
	cases := map[string]struct {
		args args
		want v1alpha1.ProjectObservation
	}{
		"Full": {
			args: args{
				p: &gitlab.Project{
					ID:            id,
					Public:        public,
					SSHURLToRepo:  sshURLToRepo,
					HTTPURLToRepo: httpURLToRepo,
					WebURL:        webURL,
					ReadmeURL:     readmeURL,
					Owner: &gitlab.User{
						Username:  owner,
						CreatedAt: &now,
					},
					PathWithNamespace:    pathWithNamespace,
					IssuesEnabled:        issuesEnabled,
					OpenIssuesCount:      openIssuesCount,
					MergeRequestsEnabled: mergeRequestsEnabled,
					JobsEnabled:          jobsEnabled,
					WikiEnabled:          wikiEnabled,
					SnippetsEnabled:      snippetsEnabled,
					CreatedAt:            &now,
					LastActivityAt:       &now,
					CreatorID:            creatorID,
					Namespace: &gitlab.ProjectNamespace{
						ID: namespaceId,
					},
					ImportStatus: importStatus,
					ImportError:  importError,
					Permissions: &gitlab.Permissions{
						ProjectAccess: &gitlab.ProjectAccess{
							AccessLevel:       gitlab.AccessLevelValue(permissionsProjectAccessAccessLevel),
							NotificationLevel: gitlab.NotificationLevelValue(permissionsProjectAccessNotificationLevel),
						},
						GroupAccess: &gitlab.GroupAccess{
							AccessLevel:       gitlab.AccessLevelValue(permissionsGroupAccessAccessLevel),
							NotificationLevel: gitlab.NotificationLevelValue(permissionsGroupAccessNotificationLevel),
						},
					},
					MarkedForDeletionAt: &markedForDeletionAt,
					Archived:            archived,
					ForksCount:          forksCount,
					StarCount:           starCount,
					//RunnersToken:         ,
					ForkedFromProject: &gitlab.ForkParent{
						HTTPURLToRepo: forkedFromProjectHTTPURL,
					},
					SharedWithGroups: sharedWithGroups,
					Statistics: &gitlab.ProjectStatistics{
						StorageStatistics: storageStatistics,
						CommitCount:       projectStatisticsCommitCount,
					},
					Links: &gitlab.Links{
						Self: linksSelf,
					},
					CIDefaultGitDepth: ciDefaultGitDepth,
					CustomAttributes: []*gitlab.CustomAttribute{
						{
							Key:   customAttributesKey,
							Value: customAttributesValue,
						},
					},
					ComplianceFrameworks: complianceFrameworks,
				},
			},
			want: v1alpha1.ProjectObservation{
				ID:            id,
				Public:        public,
				SSHURLToRepo:  sshURLToRepo,
				HTTPURLToRepo: httpURLToRepo,
				WebURL:        webURL,
				ReadmeURL:     readmeURL,
				Owner: &v1alpha1.User{
					Username:  owner,
					CreatedAt: &metav1.Time{Time: now},
				},
				PathWithNamespace:    pathWithNamespace,
				IssuesEnabled:        issuesEnabled,
				OpenIssuesCount:      openIssuesCount,
				MergeRequestsEnabled: mergeRequestsEnabled,
				JobsEnabled:          jobsEnabled,
				WikiEnabled:          wikiEnabled,
				SnippetsEnabled:      snippetsEnabled,
				CreatedAt:            &metav1.Time{Time: now},
				LastActivityAt:       &metav1.Time{Time: now},
				CreatorID:            creatorID,
				Namespace: &v1alpha1.ProjectNamespace{
					ID: namespaceId,
				},
				ImportStatus: importStatus,
				ImportError:  importError,
				Permissions: &v1alpha1.Permissions{
					ProjectAccess: &v1alpha1.ProjectAccess{
						AccessLevel:       v1alpha1.AccessLevelValue(permissionsProjectAccessAccessLevel),
						NotificationLevel: v1alpha1.NotificationLevelValue(permissionsProjectAccessNotificationLevel),
					},
					GroupAccess: &v1alpha1.GroupAccess{
						AccessLevel:       v1alpha1.AccessLevelValue(permissionsGroupAccessAccessLevel),
						NotificationLevel: v1alpha1.NotificationLevelValue(permissionsGroupAccessNotificationLevel),
					},
				},
				MarkedForDeletionAt: &metav1.Time{Time: now},
				Archived:            archived,
				ForksCount:          forksCount,
				StarCount:           starCount,
				//RunnersToken:         ,
				ForkedFromProject: &v1alpha1.ForkParent{
					HTTPURLToRepo: forkedFromProjectHTTPURL,
				},
				SharedWithGroups: []v1alpha1.SharedWithGroups{
					{
						GroupID:          sharedWithGroups[0].GroupID,
						GroupName:        sharedWithGroups[0].GroupName,
						GroupAccessLevel: sharedWithGroups[0].GroupAccessLevel,
					},
				},
				Statistics: &v1alpha1.ProjectStatistics{
					StorageStatistics: v1alpha1.StorageStatistics{
						StorageSize:      storageStatistics.StorageSize,
						RepositorySize:   storageStatistics.RepositorySize,
						LfsObjectsSize:   storageStatistics.LfsObjectsSize,
						JobArtifactsSize: storageStatistics.JobArtifactsSize,
					},
					CommitCount: projectStatisticsCommitCount,
				},
				Links: &v1alpha1.Links{
					Self: linksSelf,
				},
				CustomAttributes: []v1alpha1.CustomAttribute{
					{
						Key:   customAttributesKey,
						Value: customAttributesValue,
					},
				},
				ComplianceFrameworks: complianceFrameworks,
			},
		},
		"NullPermissions": {
			args: args{
				p: &gitlab.Project{
					ID:             id,
					Public:         public,
					CreatedAt:      &now,
					LastActivityAt: &now,
					Namespace: &gitlab.ProjectNamespace{
						ID: namespaceId,
					},
					Permissions: &gitlab.Permissions{
						ProjectAccess: nil,
						GroupAccess: &gitlab.GroupAccess{
							AccessLevel:       gitlab.AccessLevelValue(permissionsGroupAccessAccessLevel),
							NotificationLevel: gitlab.NotificationLevelValue(permissionsGroupAccessNotificationLevel),
						},
					},
				},
			},
			want: v1alpha1.ProjectObservation{
				ID:             id,
				Public:         public,
				CreatedAt:      &metav1.Time{Time: now},
				LastActivityAt: &metav1.Time{Time: now},
				Namespace: &v1alpha1.ProjectNamespace{
					ID: namespaceId,
				},
				Permissions: &v1alpha1.Permissions{
					GroupAccess: &v1alpha1.GroupAccess{
						AccessLevel:       v1alpha1.AccessLevelValue(permissionsGroupAccessAccessLevel),
						NotificationLevel: v1alpha1.NotificationLevelValue(permissionsGroupAccessNotificationLevel),
					},
				},
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
		parameters *v1alpha1.ProjectParameters
		project    *gitlab.Project
		want       *v1alpha1.ProjectParameters
	}{
		"AllOptionalFields": {
			parameters: &v1alpha1.ProjectParameters{},
			project: &gitlab.Project{
				Path:                             path,
				DefaultBranch:                    defaultBranch,
				Description:                      description,
				IssuesAccessLevel:                gitlab.AccessControlValue(issuesAccessLevel),
				RepositoryAccessLevel:            gitlab.AccessControlValue(repositoryAccessLevel),
				MergeRequestsAccessLevel:         gitlab.AccessControlValue(mergeRequestsAccessLevel),
				ForkingAccessLevel:               gitlab.AccessControlValue(forkingAccessLevel),
				BuildsAccessLevel:                gitlab.AccessControlValue(buildsAccessLevel),
				WikiAccessLevel:                  gitlab.AccessControlValue(wikiAccessLevel),
				SnippetsAccessLevel:              gitlab.AccessControlValue(snippetsAccessLevel),
				PagesAccessLevel:                 gitlab.AccessControlValue(pagesAccessLevel),
				ResolveOutdatedDiffDiscussions:   resolveOutdatedDiffDiscussions,
				ContainerRegistryEnabled:         containerRegistryEnabled,
				SharedRunnersEnabled:             sharedRunnersEnabled,
				Visibility:                       gitlab.VisibilityValue(visibility),
				PublicBuilds:                     publicBuilds,
				OnlyAllowMergeIfPipelineSucceeds: onlyAllowMergeIfPipelineSucceeds,
				OnlyAllowMergeIfAllDiscussionsAreResolved: OnlyAllowMergeIfAllDiscussionsAreResolved,
				RemoveSourceBranchAfterMerge:              removeSourceBranchAfterMerge,
				LFSEnabled:                                lfsEnabled,
				RequestAccessEnabled:                      requestAccessEnabled,
				MergeMethod:                               gitlab.MergeMethodValue(mergeMethod),
				TagList:                                   tagList,
				CIConfigPath:                              ciConfigPath,
				CIDefaultGitDepth:                         ciDefaultGitDepth,
				Mirror:                                    mirror,
				MirrorUserID:                              mirrorUserID,
				MirrorTriggerBuilds:                       mirrorTriggerBuilds,
				OnlyMirrorProtectedBranches:               onlyMirrorProtectedBranches,
				MirrorOverwritesDivergedBranches:          mirrorOverwritesDivergedBranches,
				PackagesEnabled:                           packagesEnabled,
				ServiceDeskEnabled:                        serviceDeskEnabled,
				AutocloseReferencedIssues:                 autocloseReferencedIssues,
			},
			want: &v1alpha1.ProjectParameters{
				Path:                             &path,
				DefaultBranch:                    &defaultBranch,
				Description:                      &description,
				IssuesAccessLevel:                &issuesAccessLevelv1alpha1,
				RepositoryAccessLevel:            &repositoryAccessLevelv1alpha1,
				MergeRequestsAccessLevel:         &mergeRequestsAccessLevelv1alpha1,
				ForkingAccessLevel:               &forkingAccessLevelv1alpha1,
				BuildsAccessLevel:                &buildsAccessLevelv1alpha1,
				WikiAccessLevel:                  &wikiAccessLevelv1alpha1,
				SnippetsAccessLevel:              &snippetsAccessLevelv1alpha1,
				PagesAccessLevel:                 &pagesAccessLevelv1alpha1,
				ResolveOutdatedDiffDiscussions:   &resolveOutdatedDiffDiscussions,
				ContainerRegistryEnabled:         &containerRegistryEnabled,
				SharedRunnersEnabled:             &sharedRunnersEnabled,
				Visibility:                       &visibilityv1alpha1,
				PublicBuilds:                     &publicBuilds,
				OnlyAllowMergeIfPipelineSucceeds: &onlyAllowMergeIfPipelineSucceeds,
				OnlyAllowMergeIfAllDiscussionsAreResolved: &OnlyAllowMergeIfAllDiscussionsAreResolved,
				RemoveSourceBranchAfterMerge:              &removeSourceBranchAfterMerge,
				LFSEnabled:                                &lfsEnabled,
				RequestAccessEnabled:                      &requestAccessEnabled,
				MergeMethod:                               &mergeMethodv1alpha1,
				TagList:                                   tagList,
				CIConfigPath:                              &ciConfigPath,
				CIDefaultGitDepth:                         &ciDefaultGitDepth,
				Mirror:                                    &mirror,
				MirrorUserID:                              &mirrorUserID,
				MirrorTriggerBuilds:                       &mirrorTriggerBuilds,
				OnlyMirrorProtectedBranches:               &onlyMirrorProtectedBranches,
				MirrorOverwritesDivergedBranches:          &mirrorOverwritesDivergedBranches,
				PackagesEnabled:                           &packagesEnabled,
				ServiceDeskEnabled:                        &serviceDeskEnabled,
				AutocloseReferencedIssues:                 &autocloseReferencedIssues,
			},
		},
		"SomeFieldsDontOverwrite": {
			parameters: &v1alpha1.ProjectParameters{
				Path:                           &path,
				Description:                    &description,
				DefaultBranch:                  &defaultBranch,
				ResolveOutdatedDiffDiscussions: &resolveOutdatedDiffDiscussions,
				PublicBuilds:                   &publicBuilds,
				TagList:                        tagList,
			},
			project: &gitlab.Project{
				Path:                             path,
				DefaultBranch:                    defaultBranch,
				Description:                      description,
				IssuesAccessLevel:                gitlab.AccessControlValue(issuesAccessLevel),
				RepositoryAccessLevel:            gitlab.AccessControlValue(repositoryAccessLevel),
				MergeRequestsAccessLevel:         gitlab.AccessControlValue(mergeRequestsAccessLevel),
				ForkingAccessLevel:               gitlab.AccessControlValue(forkingAccessLevel),
				BuildsAccessLevel:                gitlab.AccessControlValue(buildsAccessLevel),
				WikiAccessLevel:                  gitlab.AccessControlValue(wikiAccessLevel),
				SnippetsAccessLevel:              gitlab.AccessControlValue(snippetsAccessLevel),
				PagesAccessLevel:                 gitlab.AccessControlValue(pagesAccessLevel),
				ResolveOutdatedDiffDiscussions:   !resolveOutdatedDiffDiscussions,
				ContainerRegistryEnabled:         containerRegistryEnabled,
				SharedRunnersEnabled:             sharedRunnersEnabled,
				Visibility:                       gitlab.VisibilityValue(visibility),
				PublicBuilds:                     !publicBuilds,
				OnlyAllowMergeIfPipelineSucceeds: onlyAllowMergeIfPipelineSucceeds,
				OnlyAllowMergeIfAllDiscussionsAreResolved: OnlyAllowMergeIfAllDiscussionsAreResolved,
				RemoveSourceBranchAfterMerge:              removeSourceBranchAfterMerge,
				LFSEnabled:                                lfsEnabled,
				RequestAccessEnabled:                      requestAccessEnabled,
				MergeMethod:                               gitlab.MergeMethodValue(mergeMethod),
				TagList:                                   tagList,
				CIConfigPath:                              ciConfigPath,
				CIDefaultGitDepth:                         ciDefaultGitDepth,
				Mirror:                                    mirror,
				MirrorUserID:                              mirrorUserID,
				MirrorTriggerBuilds:                       mirrorTriggerBuilds,
				OnlyMirrorProtectedBranches:               onlyMirrorProtectedBranches,
				MirrorOverwritesDivergedBranches:          mirrorOverwritesDivergedBranches,
				PackagesEnabled:                           packagesEnabled,
				ServiceDeskEnabled:                        serviceDeskEnabled,
				AutocloseReferencedIssues:                 autocloseReferencedIssues,
			},
			want: &v1alpha1.ProjectParameters{
				Path:                             &path,
				DefaultBranch:                    &defaultBranch,
				Description:                      &description,
				IssuesAccessLevel:                &issuesAccessLevelv1alpha1,
				RepositoryAccessLevel:            &repositoryAccessLevelv1alpha1,
				MergeRequestsAccessLevel:         &mergeRequestsAccessLevelv1alpha1,
				ForkingAccessLevel:               &forkingAccessLevelv1alpha1,
				BuildsAccessLevel:                &buildsAccessLevelv1alpha1,
				WikiAccessLevel:                  &wikiAccessLevelv1alpha1,
				SnippetsAccessLevel:              &snippetsAccessLevelv1alpha1,
				PagesAccessLevel:                 &pagesAccessLevelv1alpha1,
				ResolveOutdatedDiffDiscussions:   &resolveOutdatedDiffDiscussions,
				ContainerRegistryEnabled:         &containerRegistryEnabled,
				SharedRunnersEnabled:             &sharedRunnersEnabled,
				Visibility:                       &visibilityv1alpha1,
				PublicBuilds:                     &publicBuilds,
				OnlyAllowMergeIfPipelineSucceeds: &onlyAllowMergeIfPipelineSucceeds,
				OnlyAllowMergeIfAllDiscussionsAreResolved: &OnlyAllowMergeIfAllDiscussionsAreResolved,
				RemoveSourceBranchAfterMerge:              &removeSourceBranchAfterMerge,
				LFSEnabled:                                &lfsEnabled,
				RequestAccessEnabled:                      &requestAccessEnabled,
				MergeMethod:                               &mergeMethodv1alpha1,
				TagList:                                   tagList,
				CIConfigPath:                              &ciConfigPath,
				CIDefaultGitDepth:                         &ciDefaultGitDepth,
				Mirror:                                    &mirror,
				MirrorUserID:                              &mirrorUserID,
				MirrorTriggerBuilds:                       &mirrorTriggerBuilds,
				OnlyMirrorProtectedBranches:               &onlyMirrorProtectedBranches,
				MirrorOverwritesDivergedBranches:          &mirrorOverwritesDivergedBranches,
				PackagesEnabled:                           &packagesEnabled,
				ServiceDeskEnabled:                        &serviceDeskEnabled,
				AutocloseReferencedIssues:                 &autocloseReferencedIssues,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitialize(tc.parameters, tc.project)
			if diff := cmp.Diff(tc.want, tc.parameters); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateEditProjectOptions(t *testing.T) {
	type args struct {
		name       string
		parameters *v1alpha1.ProjectParameters
	}
	cases := map[string]struct {
		args args
		want *gitlab.EditProjectOptions
	}{
		"AllFields": {
			args: args{
				name: name,
				parameters: &v1alpha1.ProjectParameters{
					Path:                             &path,
					Description:                      &description,
					DefaultBranch:                    &defaultBranch,
					IssuesAccessLevel:                &issuesAccessLevelv1alpha1,
					RepositoryAccessLevel:            &repositoryAccessLevelv1alpha1,
					MergeRequestsAccessLevel:         &mergeRequestsAccessLevelv1alpha1,
					ForkingAccessLevel:               &forkingAccessLevelv1alpha1,
					BuildsAccessLevel:                &buildsAccessLevelv1alpha1,
					WikiAccessLevel:                  &wikiAccessLevelv1alpha1,
					SnippetsAccessLevel:              &snippetsAccessLevelv1alpha1,
					PagesAccessLevel:                 &pagesAccessLevelv1alpha1,
					ResolveOutdatedDiffDiscussions:   &resolveOutdatedDiffDiscussions,
					ContainerRegistryEnabled:         &containerRegistryEnabled,
					SharedRunnersEnabled:             &sharedRunnersEnabled,
					Visibility:                       &visibilityv1alpha1,
					PublicBuilds:                     &publicBuilds,
					OnlyAllowMergeIfPipelineSucceeds: &onlyAllowMergeIfPipelineSucceeds,
					OnlyAllowMergeIfAllDiscussionsAreResolved: &OnlyAllowMergeIfAllDiscussionsAreResolved,
					RemoveSourceBranchAfterMerge:              &removeSourceBranchAfterMerge,
					LFSEnabled:                                &lfsEnabled,
					RequestAccessEnabled:                      &requestAccessEnabled,
					MergeMethod:                               &mergeMethodv1alpha1,
					TagList:                                   tagList,
					BuildGitStrategy:                          &buildGitStategy,
					BuildTimeout:                              &buildTimeout,
					CIConfigPath:                              &ciConfigPath,
					CIDefaultGitDepth:                         &ciDefaultGitDepth,
					Mirror:                                    &mirror,
					MirrorUserID:                              &mirrorUserID,
					MirrorTriggerBuilds:                       &mirrorTriggerBuilds,
					OnlyMirrorProtectedBranches:               &onlyMirrorProtectedBranches,
					MirrorOverwritesDivergedBranches:          &mirrorOverwritesDivergedBranches,
					PackagesEnabled:                           &packagesEnabled,
					ServiceDeskEnabled:                        &serviceDeskEnabled,
					AutocloseReferencedIssues:                 &autocloseReferencedIssues,
				},
			},
			want: &gitlab.EditProjectOptions{
				Name:                             &name,
				Path:                             &path,
				DefaultBranch:                    &defaultBranch,
				Description:                      &description,
				IssuesAccessLevel:                clients.AccessControlValueStringToGitlab(issuesAccessLevel),
				RepositoryAccessLevel:            clients.AccessControlValueStringToGitlab(repositoryAccessLevel),
				MergeRequestsAccessLevel:         clients.AccessControlValueStringToGitlab(mergeRequestsAccessLevel),
				ForkingAccessLevel:               clients.AccessControlValueStringToGitlab(forkingAccessLevel),
				BuildsAccessLevel:                clients.AccessControlValueStringToGitlab(buildsAccessLevel),
				WikiAccessLevel:                  clients.AccessControlValueStringToGitlab(wikiAccessLevel),
				SnippetsAccessLevel:              clients.AccessControlValueStringToGitlab(snippetsAccessLevel),
				PagesAccessLevel:                 clients.AccessControlValueStringToGitlab(pagesAccessLevel),
				ResolveOutdatedDiffDiscussions:   &resolveOutdatedDiffDiscussions,
				ContainerRegistryEnabled:         &containerRegistryEnabled,
				SharedRunnersEnabled:             &sharedRunnersEnabled,
				Visibility:                       clients.VisibilityValueStringToGitlab(visibility),
				PublicBuilds:                     &publicBuilds,
				OnlyAllowMergeIfPipelineSucceeds: &onlyAllowMergeIfPipelineSucceeds,
				OnlyAllowMergeIfAllDiscussionsAreResolved: &OnlyAllowMergeIfAllDiscussionsAreResolved,
				RemoveSourceBranchAfterMerge:              &removeSourceBranchAfterMerge,
				LFSEnabled:                                &lfsEnabled,
				RequestAccessEnabled:                      &requestAccessEnabled,
				MergeMethod:                               clients.MergeMethodStringToGitlab(mergeMethod),
				TagList:                                   &tagList,
				BuildGitStrategy:                          &buildGitStategy,
				BuildTimeout:                              &buildTimeout,
				CIConfigPath:                              &ciConfigPath,
				CIDefaultGitDepth:                         &ciDefaultGitDepth,
				Mirror:                                    &mirror,
				MirrorUserID:                              &mirrorUserID,
				MirrorTriggerBuilds:                       &mirrorTriggerBuilds,
				OnlyMirrorProtectedBranches:               &onlyMirrorProtectedBranches,
				MirrorOverwritesDivergedBranches:          &mirrorOverwritesDivergedBranches,
				PackagesEnabled:                           &packagesEnabled,
				ServiceDeskEnabled:                        &serviceDeskEnabled,
				AutocloseReferencedIssues:                 &autocloseReferencedIssues,
			},
		},
		"SomeFields": {
			args: args{
				name: name,
				parameters: &v1alpha1.ProjectParameters{
					Path:                           &path,
					IssuesAccessLevel:              &issuesAccessLevelv1alpha1,
					ResolveOutdatedDiffDiscussions: &resolveOutdatedDiffDiscussions,
					MergeMethod:                    &mergeMethodv1alpha1,
					TagList:                        tagList,
					BuildTimeout:                   &buildTimeout,
				},
			},
			want: &gitlab.EditProjectOptions{
				Name:                           &name,
				Path:                           &path,
				IssuesAccessLevel:              clients.AccessControlValueStringToGitlab(issuesAccessLevel),
				ResolveOutdatedDiffDiscussions: &resolveOutdatedDiffDiscussions,
				MergeMethod:                    clients.MergeMethodStringToGitlab(mergeMethod),
				TagList:                        &tagList,
				BuildTimeout:                   &buildTimeout,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateEditProjectOptions(tc.args.name, tc.args.parameters)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsProjectUpToDate(t *testing.T) {
	type args struct {
		project *gitlab.Project
		p       *v1alpha1.ProjectParameters
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				p: &v1alpha1.ProjectParameters{
					Path:                             &path,
					DefaultBranch:                    &defaultBranch,
					Description:                      &description,
					IssuesAccessLevel:                &issuesAccessLevelv1alpha1,
					RepositoryAccessLevel:            &repositoryAccessLevelv1alpha1,
					MergeRequestsAccessLevel:         &mergeRequestsAccessLevelv1alpha1,
					ForkingAccessLevel:               &forkingAccessLevelv1alpha1,
					BuildsAccessLevel:                &buildsAccessLevelv1alpha1,
					WikiAccessLevel:                  &wikiAccessLevelv1alpha1,
					SnippetsAccessLevel:              &snippetsAccessLevelv1alpha1,
					PagesAccessLevel:                 &pagesAccessLevelv1alpha1,
					ResolveOutdatedDiffDiscussions:   &resolveOutdatedDiffDiscussions,
					ContainerRegistryEnabled:         &containerRegistryEnabled,
					SharedRunnersEnabled:             &sharedRunnersEnabled,
					Visibility:                       &visibilityv1alpha1,
					PublicBuilds:                     &publicBuilds,
					OnlyAllowMergeIfPipelineSucceeds: &onlyAllowMergeIfPipelineSucceeds,
					OnlyAllowMergeIfAllDiscussionsAreResolved: &OnlyAllowMergeIfAllDiscussionsAreResolved,
					MergeMethod:                  &mergeMethodv1alpha1,
					RemoveSourceBranchAfterMerge: &removeSourceBranchAfterMerge,
					LFSEnabled:                   &lfsEnabled,
					RequestAccessEnabled:         &requestAccessEnabled,
					TagList:                      tagList,
					CIConfigPath:                 &ciConfigPath,
					ApprovalsBeforeMerge:         &approvalsBeforeMerge,
					Mirror:                       &mirror,
					MirrorTriggerBuilds:          &mirrorTriggerBuilds,
					PackagesEnabled:              &packagesEnabled,
					ServiceDeskEnabled:           &serviceDeskEnabled,
					AutocloseReferencedIssues:    &autocloseReferencedIssues,
				},
				project: &gitlab.Project{
					Path:                             path,
					DefaultBranch:                    defaultBranch,
					Description:                      description,
					IssuesAccessLevel:                gitlab.AccessControlValue(issuesAccessLevel),
					RepositoryAccessLevel:            gitlab.AccessControlValue(repositoryAccessLevel),
					MergeRequestsAccessLevel:         gitlab.AccessControlValue(mergeRequestsAccessLevel),
					ForkingAccessLevel:               gitlab.AccessControlValue(forkingAccessLevel),
					BuildsAccessLevel:                gitlab.AccessControlValue(buildsAccessLevel),
					WikiAccessLevel:                  gitlab.AccessControlValue(wikiAccessLevel),
					SnippetsAccessLevel:              gitlab.AccessControlValue(snippetsAccessLevel),
					PagesAccessLevel:                 gitlab.AccessControlValue(pagesAccessLevel),
					ResolveOutdatedDiffDiscussions:   resolveOutdatedDiffDiscussions,
					ContainerRegistryEnabled:         containerRegistryEnabled,
					SharedRunnersEnabled:             sharedRunnersEnabled,
					Visibility:                       gitlab.VisibilityValue(visibility),
					PublicBuilds:                     publicBuilds,
					OnlyAllowMergeIfPipelineSucceeds: onlyAllowMergeIfPipelineSucceeds,
					OnlyAllowMergeIfAllDiscussionsAreResolved: OnlyAllowMergeIfAllDiscussionsAreResolved,
					MergeMethod:                  gitlab.MergeMethodValue(mergeMethod),
					RemoveSourceBranchAfterMerge: removeSourceBranchAfterMerge,
					LFSEnabled:                   lfsEnabled,
					RequestAccessEnabled:         requestAccessEnabled,
					TagList:                      tagList,
					CIConfigPath:                 ciConfigPath,
					ApprovalsBeforeMerge:         approvalsBeforeMerge,
					Mirror:                       mirror,
					MirrorTriggerBuilds:          mirrorTriggerBuilds,
					PackagesEnabled:              packagesEnabled,
					ServiceDeskEnabled:           serviceDeskEnabled,
					AutocloseReferencedIssues:    autocloseReferencedIssues,
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				p: &v1alpha1.ProjectParameters{
					Path:                             &path,
					DefaultBranch:                    &defaultBranch,
					Description:                      &description,
					IssuesAccessLevel:                &issuesAccessLevelv1alpha1,
					RepositoryAccessLevel:            &repositoryAccessLevelv1alpha1,
					MergeRequestsAccessLevel:         &mergeRequestsAccessLevelv1alpha1,
					ForkingAccessLevel:               &forkingAccessLevelv1alpha1,
					BuildsAccessLevel:                &buildsAccessLevelv1alpha1,
					WikiAccessLevel:                  &wikiAccessLevelv1alpha1,
					SnippetsAccessLevel:              &snippetsAccessLevelv1alpha1,
					PagesAccessLevel:                 &pagesAccessLevelv1alpha1,
					ResolveOutdatedDiffDiscussions:   &resolveOutdatedDiffDiscussions,
					ContainerRegistryEnabled:         &containerRegistryEnabled,
					SharedRunnersEnabled:             &sharedRunnersEnabled,
					Visibility:                       &visibilityv1alpha1,
					PublicBuilds:                     &publicBuilds,
					OnlyAllowMergeIfPipelineSucceeds: &onlyAllowMergeIfPipelineSucceeds,
					OnlyAllowMergeIfAllDiscussionsAreResolved: &OnlyAllowMergeIfAllDiscussionsAreResolved,
					MergeMethod:                  &mergeMethodv1alpha1,
					RemoveSourceBranchAfterMerge: &removeSourceBranchAfterMerge,
					LFSEnabled:                   &lfsEnabled,
					RequestAccessEnabled:         &requestAccessEnabled,
					TagList:                      tagList,
					CIConfigPath:                 &ciConfigPath,
					ApprovalsBeforeMerge:         &approvalsBeforeMerge,
					Mirror:                       &mirror,
					MirrorTriggerBuilds:          &mirrorTriggerBuilds,
					PackagesEnabled:              &packagesEnabled,
					ServiceDeskEnabled:           &serviceDeskEnabled,
					AutocloseReferencedIssues:    &autocloseReferencedIssues,
				},
				project: &gitlab.Project{
					Path:                             "some/other/path",
					DefaultBranch:                    defaultBranch,
					Description:                      "some other description",
					IssuesAccessLevel:                gitlab.AccessControlValue(issuesAccessLevel),
					RepositoryAccessLevel:            gitlab.AccessControlValue(repositoryAccessLevel),
					MergeRequestsAccessLevel:         gitlab.AccessControlValue(mergeRequestsAccessLevel),
					ForkingAccessLevel:               gitlab.AccessControlValue(forkingAccessLevel),
					BuildsAccessLevel:                gitlab.AccessControlValue(buildsAccessLevel),
					WikiAccessLevel:                  gitlab.AccessControlValue(wikiAccessLevel),
					SnippetsAccessLevel:              gitlab.AccessControlValue(snippetsAccessLevel),
					PagesAccessLevel:                 gitlab.AccessControlValue(pagesAccessLevel),
					ResolveOutdatedDiffDiscussions:   resolveOutdatedDiffDiscussions,
					ContainerRegistryEnabled:         containerRegistryEnabled,
					SharedRunnersEnabled:             !sharedRunnersEnabled,
					Visibility:                       gitlab.VisibilityValue(visibility),
					PublicBuilds:                     publicBuilds,
					OnlyAllowMergeIfPipelineSucceeds: onlyAllowMergeIfPipelineSucceeds,
					OnlyAllowMergeIfAllDiscussionsAreResolved: OnlyAllowMergeIfAllDiscussionsAreResolved,
					MergeMethod:                  gitlab.MergeMethodValue(mergeMethod),
					RemoveSourceBranchAfterMerge: removeSourceBranchAfterMerge,
					LFSEnabled:                   lfsEnabled,
					RequestAccessEnabled:         requestAccessEnabled,
					TagList:                      []string{"some", "other", "tags"},
					CIConfigPath:                 ciConfigPath,
					ApprovalsBeforeMerge:         approvalsBeforeMerge + 1,
					Mirror:                       mirror,
					MirrorTriggerBuilds:          !mirrorTriggerBuilds,
					PackagesEnabled:              packagesEnabled,
					ServiceDeskEnabled:           !serviceDeskEnabled,
					AutocloseReferencedIssues:    autocloseReferencedIssues,
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsProjectUpToDate(tc.args.p, tc.args.project)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
