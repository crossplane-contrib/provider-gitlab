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
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
)

var (
	name                                 = "my-project"
	overrideName                         = "My Project"
	path                                 = "path/to/project"
	namespaceID                          = 1
	defaultBranch                        = "main"
	description                          = "my awesome project"
	issuesAccessLevel                    = "enabled"
	issuesAccessLevelv1alpha1            = v1alpha1.AccessControlValue(issuesAccessLevel)
	repositoryAccessLevel                = "enabled"
	repositoryAccessLevelv1alpha1        = v1alpha1.AccessControlValue(repositoryAccessLevel)
	mergeRequestsAccessLevel             = "enabled"
	mergeRequestsAccessLevelv1alpha1     = v1alpha1.AccessControlValue(mergeRequestsAccessLevel)
	forkingAccessLevel                   = "enabled"
	forkingAccessLevelv1alpha1           = v1alpha1.AccessControlValue(forkingAccessLevel)
	buildsAccessLevel                    = "disabled"
	buildsAccessLevelv1alpha1            = v1alpha1.AccessControlValue(buildsAccessLevel)
	wikiAccessLevel                      = "private"
	wikiAccessLevelv1alpha1              = v1alpha1.AccessControlValue(wikiAccessLevel)
	snippetsAccessLevel                  = "public"
	snippetsAccessLevelv1alpha1          = v1alpha1.AccessControlValue(snippetsAccessLevel)
	pagesAccessLevel                     = "enabled"
	pagesAccessLevelv1alpha1             = v1alpha1.AccessControlValue(pagesAccessLevel)
	operationsAccessLevel                = "public"
	operationsAccessLevelv1alpha1        = v1alpha1.AccessControlValue(operationsAccessLevel)
	containerRegistryAccessLevel         = "enabled"
	containerRegistryAccessLevelv1alpha1 = v1alpha1.AccessControlValue(pagesAccessLevel)
	emailsDisabled                       = true
	resolveOutdatedDiffDiscussions       = true
	cadence                              = "Cadence"
	keepN                                = 1
	olderThan                            = "OlderThan"
	nameRegexDelete                      = "NameRegexDelete"
	nameRegexKeep                        = "NameRegexKeep"
	enabled                              = false
	nextRunAt                            = time.Now()
	gitlabContainerExpirationPolicy      = gitlab.ContainerExpirationPolicy{
		Cadence:         cadence,
		KeepN:           keepN,
		OlderThan:       olderThan,
		NameRegexDelete: nameRegexDelete,
		NameRegexKeep:   nameRegexKeep,
		Enabled:         enabled,
		NextRunAt:       &nextRunAt,
	}
	v1alpha1ContainerExpirationPolicy = v1alpha1.ContainerExpirationPolicy{
		Cadence:         cadence,
		KeepN:           keepN,
		OlderThan:       olderThan,
		NameRegexDelete: nameRegexDelete,
		NameRegexKeep:   nameRegexKeep,
		Enabled:         enabled,
		NextRunAt:       &metav1.Time{Time: nextRunAt},
	}
	v1alpha1ContainerExpirationPolicyAttributes = v1alpha1.ContainerExpirationPolicyAttributes{
		Cadence:         &cadence,
		KeepN:           &keepN,
		OlderThan:       &olderThan,
		NameRegexDelete: &nameRegexDelete,
		NameRegexKeep:   &nameRegexKeep,
		Enabled:         &enabled,
	}
	gitlabContainerExpirationPolicyAttributes = gitlab.ContainerExpirationPolicyAttributes{
		Cadence:         &cadence,
		KeepN:           &keepN,
		OlderThan:       &olderThan,
		NameRegexDelete: &nameRegexDelete,
		NameRegexKeep:   &nameRegexKeep,
		Enabled:         &enabled,
	}
	sharedRunnersEnabled                      = true
	visibility                                = "private"
	visibilityv1alpha1                        = v1alpha1.VisibilityValue(visibility)
	importURL                                 = "import.url"
	publicBuilds                              = false
	allowMergeOnSkippedPipeline               = false
	onlyAllowMergeIfPipelineSucceeds          = true
	OnlyAllowMergeIfAllDiscussionsAreResolved = true
	mergeMethod                               = "merge"
	mergeMethodv1alpha1                       = v1alpha1.MergeMethodValue(mergeMethod)
	removeSourceBranchAfterMerge              = false
	lfsEnabled                                = true
	requestAccessEnabled                      = true
	topics                                    = []string{"tag1", "tag2"}
	printingMergeRequestLinkEnabled           = true
	buildGitStategy                           = "strategy"
	buildTimeout                              = 60
	autoCancelPendingPipelines                = "enabled"
	buildCoverageRegex                        = "some-regex"
	ciConfigPath                              = "path/to/ci/config"
	ciForwardDeploymentEnabled                = false
	ciDefaultGitDepth                         = 50
	autoDevopsEnabled                         = true
	autoDevopsDeployStrategy                  = "continuous"
	externalAuthorizationClassificationLabel  = "authz-label"
	mirror                                    = false
	mirrorUserID                              = 1
	mirrorTriggerBuilds                       = true
	initializeWithReadme                      = true
	templateName                              = "template"
	templateProjectID                         = 1
	useCustomTemplate                         = true
	groupWithProjectTemplatesID               = 1
	onlyMirrorProtectedBranches               = false
	mirrorOverwritesDivergedBranches          = false
	packagesEnabled                           = true
	serviceDeskEnabled                        = true
	autocloseReferencedIssues                 = true
	suggestionCommitMessage                   = "SuggestionCommitMessage"
	issuesTemplate                            = "IssuesTemplate"
	mergeRequestsTemplate                     = "MergeRequestsTemplate"
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
	nameWithNamespace := "name/to/cool-project"
	issuesAccessLevel := gitlab.EnabledAccessControl
	openIssuesCount := 3
	now := time.Now()
	creatorID := 1
	namespaceID := 3
	importStatus := "foo"
	importError := "none"
	permissionsProjectAccessAccessLevel := 1
	permissionsProjectAccessNotificationLevel := 2
	permissionsGroupAccessAccessLevel := 3
	permissionsGroupAccessNotificationLevel := 4
	markedForDeletionOn := gitlab.ISOTime(now)
	emptyRepo := false
	archived := false
	avatarURL := "https://AvatarURL"
	licenseURL := "https://LicenseURL"
	licenseKey := "Key"
	licenseName := "Name"
	licenseNickname := "Nickname"
	licenseHTMLURL := "HTMLURL"
	licenseSourceURL := "SourceURL"
	gitlabLicense := gitlab.ProjectLicense{
		Key:       licenseKey,
		Name:      licenseName,
		Nickname:  licenseNickname,
		HTMLURL:   licenseHTMLURL,
		SourceURL: licenseSourceURL,
	}
	v1alpha1License := v1alpha1.ProjectLicense{
		Key:       licenseKey,
		Name:      licenseName,
		Nickname:  licenseNickname,
		HTMLURL:   licenseHTMLURL,
		SourceURL: licenseSourceURL,
	}
	forksCount := 2
	starCount := 10000
	forkedFromProjectHTTPURL := "http://fork.url"
	serviceDeskAddress := "ServiceDeskAddress"
	sharedWithGroups := []struct {
		GroupID          int    `json:"group_id"`
		GroupName        string `json:"group_name"`
		GroupFullPath    string `json:"group_full_path"`
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
		LFSObjectsSize   int64 `json:"lfs_objects_size"`
		JobArtifactsSize int64 `json:"job_artifacts_size"`
	}{
		StorageSize:      10,
		RepositorySize:   20,
		LFSObjectsSize:   30,
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
					PublicJobs:    public,
					SSHURLToRepo:  sshURLToRepo,
					HTTPURLToRepo: httpURLToRepo,
					WebURL:        webURL,
					ReadmeURL:     readmeURL,
					Owner: &gitlab.User{
						Username:  owner,
						CreatedAt: &now,
					},
					PathWithNamespace:         pathWithNamespace,
					NameWithNamespace:         nameWithNamespace,
					IssuesAccessLevel:         issuesAccessLevel,
					OpenIssuesCount:           openIssuesCount,
					MergeRequestsAccessLevel:  gitlab.AccessControlValue(mergeRequestsAccessLevel),
					BuildsAccessLevel:         gitlab.AccessControlValue(buildsAccessLevel),
					WikiAccessLevel:           gitlab.AccessControlValue(wikiAccessLevel),
					SnippetsAccessLevel:       gitlab.AccessControlValue(snippetsAccessLevel),
					ContainerExpirationPolicy: &gitlabContainerExpirationPolicy,
					CreatedAt:                 &now,
					LastActivityAt:            &now,
					CreatorID:                 creatorID,
					Namespace: &gitlab.ProjectNamespace{
						ID: namespaceID,
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
					MarkedForDeletionOn: &markedForDeletionOn,
					EmptyRepo:           emptyRepo,
					Archived:            archived,
					AvatarURL:           avatarURL,
					LicenseURL:          licenseURL,
					License:             &gitlabLicense,
					ForksCount:          forksCount,
					StarCount:           starCount,
					ForkedFromProject: &gitlab.ForkParent{
						HTTPURLToRepo: forkedFromProjectHTTPURL,
					},
					ServiceDeskAddress: serviceDeskAddress,
					SharedWithGroups:   sharedWithGroups,
					Statistics: &gitlab.Statistics{
						StorageSize:      storageStatistics.StorageSize,
						RepositorySize:   storageStatistics.RepositorySize,
						LFSObjectsSize:   storageStatistics.LFSObjectsSize,
						JobArtifactsSize: storageStatistics.JobArtifactsSize,
						CommitCount:      int64(projectStatisticsCommitCount),
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
				PathWithNamespace:         pathWithNamespace,
				NameWithNamespace:         nameWithNamespace,
				IssuesAccessLevel:         issuesAccessLevelv1alpha1,
				OpenIssuesCount:           openIssuesCount,
				MergeRequestsAccessLevel:  mergeRequestsAccessLevelv1alpha1,
				BuildsAccessLevel:         buildsAccessLevelv1alpha1,
				WikiAccessLevel:           wikiAccessLevelv1alpha1,
				SnippetsAccessLevel:       snippetsAccessLevelv1alpha1,
				ContainerExpirationPolicy: &v1alpha1ContainerExpirationPolicy,
				CreatedAt:                 &metav1.Time{Time: now},
				LastActivityAt:            &metav1.Time{Time: now},
				CreatorID:                 creatorID,
				Namespace: &v1alpha1.ProjectNamespace{
					ID: namespaceID,
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
				MarkedForDeletionOn: &metav1.Time{Time: now},
				EmptyRepo:           emptyRepo,
				Archived:            archived,
				AvatarURL:           avatarURL,
				LicenseURL:          licenseURL,
				License:             &v1alpha1License,
				ForksCount:          forksCount,
				StarCount:           starCount,
				ForkedFromProject: &v1alpha1.ForkParent{
					HTTPURLToRepo: forkedFromProjectHTTPURL,
				},
				ServiceDeskAddress: serviceDeskAddress,
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
						LfsObjectsSize:   storageStatistics.LFSObjectsSize,
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
					PublicJobs:     public,
					CreatedAt:      &now,
					LastActivityAt: &now,
					Namespace: &gitlab.ProjectNamespace{
						ID: namespaceID,
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
					ID: namespaceID,
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

func TestGenerateCreateProjectOptions(t *testing.T) {
	type args struct {
		name       string
		parameters *v1alpha1.ProjectParameters
	}
	cases := map[string]struct {
		args args
		want *gitlab.CreateProjectOptions
	}{
		"AllFields": {
			args: args{
				name: name,
				parameters: &v1alpha1.ProjectParameters{
					Path:                                      &path,
					NamespaceID:                               &namespaceID,
					DefaultBranch:                             &defaultBranch,
					Description:                               &description,
					IssuesAccessLevel:                         &issuesAccessLevelv1alpha1,
					RepositoryAccessLevel:                     &repositoryAccessLevelv1alpha1,
					MergeRequestsAccessLevel:                  &mergeRequestsAccessLevelv1alpha1,
					ForkingAccessLevel:                        &forkingAccessLevelv1alpha1,
					BuildsAccessLevel:                         &buildsAccessLevelv1alpha1,
					WikiAccessLevel:                           &wikiAccessLevelv1alpha1,
					SnippetsAccessLevel:                       &snippetsAccessLevelv1alpha1,
					PagesAccessLevel:                          &pagesAccessLevelv1alpha1,
					OperationsAccessLevel:                     &operationsAccessLevelv1alpha1,
					EmailsDisabled:                            &emailsDisabled,
					ResolveOutdatedDiffDiscussions:            &resolveOutdatedDiffDiscussions,
					ContainerExpirationPolicyAttributes:       &v1alpha1ContainerExpirationPolicyAttributes,
					ContainerRegistryAccessLevel:              &containerRegistryAccessLevelv1alpha1,
					SharedRunnersEnabled:                      &sharedRunnersEnabled,
					Visibility:                                &visibilityv1alpha1,
					ImportURL:                                 &importURL,
					PublicBuilds:                              &publicBuilds,
					AllowMergeOnSkippedPipeline:               &allowMergeOnSkippedPipeline,
					OnlyAllowMergeIfPipelineSucceeds:          &onlyAllowMergeIfPipelineSucceeds,
					OnlyAllowMergeIfAllDiscussionsAreResolved: &OnlyAllowMergeIfAllDiscussionsAreResolved,
					MergeMethod:                               &mergeMethodv1alpha1,
					RemoveSourceBranchAfterMerge:              &removeSourceBranchAfterMerge,
					LFSEnabled:                                &lfsEnabled,
					RequestAccessEnabled:                      &requestAccessEnabled,
					Topics:                                    topics,
					PrintingMergeRequestLinkEnabled:           &printingMergeRequestLinkEnabled,
					BuildGitStrategy:                          &buildGitStategy,
					BuildTimeout:                              &buildTimeout,
					AutoCancelPendingPipelines:                &autoCancelPendingPipelines,
					BuildCoverageRegex:                        &buildCoverageRegex,
					CIConfigPath:                              &ciConfigPath,
					CIForwardDeploymentEnabled:                &ciForwardDeploymentEnabled,
					CIDefaultGitDepth:                         &ciDefaultGitDepth,
					AutoDevopsEnabled:                         &autoDevopsEnabled,
					AutoDevopsDeployStrategy:                  &autoDevopsDeployStrategy,
					ExternalAuthorizationClassificationLabel:  &externalAuthorizationClassificationLabel,
					Mirror:                                    &mirror,
					MirrorTriggerBuilds:                       &mirrorTriggerBuilds,
					InitializeWithReadme:                      &initializeWithReadme,
					TemplateName:                              &templateName,
					TemplateProjectID:                         &templateProjectID,
					UseCustomTemplate:                         &useCustomTemplate,
					GroupWithProjectTemplatesID:               &groupWithProjectTemplatesID,
					PackagesEnabled:                           &packagesEnabled,
					ServiceDeskEnabled:                        &serviceDeskEnabled,
					AutocloseReferencedIssues:                 &autocloseReferencedIssues,
					SuggestionCommitMessage:                   &suggestionCommitMessage,
					IssuesTemplate:                            &issuesTemplate,
					MergeRequestsTemplate:                     &mergeRequestsTemplate,
				},
			},
			want: &gitlab.CreateProjectOptions{
				Name:                                &name,
				Path:                                &path,
				NamespaceID:                         &namespaceID,
				DefaultBranch:                       &defaultBranch,
				Description:                         &description,
				IssuesAccessLevel:                   clients.AccessControlValueStringToGitlab(issuesAccessLevel),
				RepositoryAccessLevel:               clients.AccessControlValueStringToGitlab(repositoryAccessLevel),
				MergeRequestsAccessLevel:            clients.AccessControlValueStringToGitlab(mergeRequestsAccessLevel),
				ForkingAccessLevel:                  clients.AccessControlValueStringToGitlab(forkingAccessLevel),
				BuildsAccessLevel:                   clients.AccessControlValueStringToGitlab(buildsAccessLevel),
				WikiAccessLevel:                     clients.AccessControlValueStringToGitlab(wikiAccessLevel),
				SnippetsAccessLevel:                 clients.AccessControlValueStringToGitlab(snippetsAccessLevel),
				PagesAccessLevel:                    clients.AccessControlValueStringToGitlab(pagesAccessLevel),
				OperationsAccessLevel:               clients.AccessControlValueStringToGitlab(operationsAccessLevel),
				EmailsDisabled:                      &emailsDisabled,
				ResolveOutdatedDiffDiscussions:      &resolveOutdatedDiffDiscussions,
				ContainerExpirationPolicyAttributes: &gitlabContainerExpirationPolicyAttributes,
				ContainerRegistryAccessLevel:        clients.AccessControlValueStringToGitlab(containerRegistryAccessLevel),
				SharedRunnersEnabled:                &sharedRunnersEnabled,
				Visibility:                          clients.VisibilityValueStringToGitlab(visibility),
				ImportURL:                           &importURL,
				PublicBuilds:                        &publicBuilds,
				AllowMergeOnSkippedPipeline:         &allowMergeOnSkippedPipeline,
				OnlyAllowMergeIfPipelineSucceeds:    &onlyAllowMergeIfPipelineSucceeds,
				OnlyAllowMergeIfAllDiscussionsAreResolved: &OnlyAllowMergeIfAllDiscussionsAreResolved,
				MergeMethod:                              clients.MergeMethodStringToGitlab(mergeMethod),
				RemoveSourceBranchAfterMerge:             &removeSourceBranchAfterMerge,
				LFSEnabled:                               &lfsEnabled,
				RequestAccessEnabled:                     &requestAccessEnabled,
				TagList:                                  ptr.To[[]string](nil),
				Topics:                                   &topics,
				PrintingMergeRequestLinkEnabled:          &printingMergeRequestLinkEnabled,
				BuildGitStrategy:                         &buildGitStategy,
				BuildTimeout:                             &buildTimeout,
				AutoCancelPendingPipelines:               &autoCancelPendingPipelines,
				BuildCoverageRegex:                       &buildCoverageRegex,
				CIConfigPath:                             &ciConfigPath,
				CIForwardDeploymentEnabled:               &ciForwardDeploymentEnabled,
				AutoDevopsEnabled:                        &autoDevopsEnabled,
				AutoDevopsDeployStrategy:                 &autoDevopsDeployStrategy,
				ExternalAuthorizationClassificationLabel: &externalAuthorizationClassificationLabel,
				Mirror:                                   &mirror,
				MirrorTriggerBuilds:                      &mirrorTriggerBuilds,
				InitializeWithReadme:                     &initializeWithReadme,
				TemplateName:                             &templateName,
				TemplateProjectID:                        &templateProjectID,
				UseCustomTemplate:                        &useCustomTemplate,
				GroupWithProjectTemplatesID:              &groupWithProjectTemplatesID,
				PackagesEnabled:                          &packagesEnabled,
				ServiceDeskEnabled:                       &serviceDeskEnabled,
				AutocloseReferencedIssues:                &autocloseReferencedIssues,
				SuggestionCommitMessage:                  &suggestionCommitMessage,
				IssuesTemplate:                           &issuesTemplate,
				MergeRequestsTemplate:                    &mergeRequestsTemplate,
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
					Topics:                         topics,
					BuildTimeout:                   &buildTimeout,
				},
			},
			want: &gitlab.CreateProjectOptions{
				Name:                           &name,
				Path:                           &path,
				IssuesAccessLevel:              clients.AccessControlValueStringToGitlab(issuesAccessLevel),
				ResolveOutdatedDiffDiscussions: &resolveOutdatedDiffDiscussions,
				MergeMethod:                    clients.MergeMethodStringToGitlab(mergeMethod),
				TagList:                        ptr.To[[]string](nil),
				Topics:                         &topics,
				BuildTimeout:                   &buildTimeout,
			},
		},
		"NameOverride": {
			args: args{
				name: name,
				parameters: &v1alpha1.ProjectParameters{
					Name: &overrideName,
					// TODO: TagList specified and wanted because otherwise
					// test does not pass - (nil vs &nil - fix)
					Topics: topics,
				},
			},
			want: &gitlab.CreateProjectOptions{
				Name:    &overrideName,
				TagList: ptr.To[[]string](nil),
				Topics:  &topics,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateCreateProjectOptions(tc.args.name, tc.args.parameters)
			if diff := cmp.Diff(tc.want, got); diff != "" {
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
					Path:                                      &path,
					DefaultBranch:                             &defaultBranch,
					Description:                               &description,
					IssuesAccessLevel:                         &issuesAccessLevelv1alpha1,
					RepositoryAccessLevel:                     &repositoryAccessLevelv1alpha1,
					MergeRequestsAccessLevel:                  &mergeRequestsAccessLevelv1alpha1,
					ForkingAccessLevel:                        &forkingAccessLevelv1alpha1,
					BuildsAccessLevel:                         &buildsAccessLevelv1alpha1,
					WikiAccessLevel:                           &wikiAccessLevelv1alpha1,
					SnippetsAccessLevel:                       &snippetsAccessLevelv1alpha1,
					PagesAccessLevel:                          &pagesAccessLevelv1alpha1,
					OperationsAccessLevel:                     &operationsAccessLevelv1alpha1,
					EmailsDisabled:                            &emailsDisabled,
					ResolveOutdatedDiffDiscussions:            &resolveOutdatedDiffDiscussions,
					ContainerExpirationPolicyAttributes:       &v1alpha1ContainerExpirationPolicyAttributes,
					ContainerRegistryAccessLevel:              &containerRegistryAccessLevelv1alpha1,
					SharedRunnersEnabled:                      &sharedRunnersEnabled,
					Visibility:                                &visibilityv1alpha1,
					ImportURL:                                 &importURL,
					PublicBuilds:                              &publicBuilds,
					AllowMergeOnSkippedPipeline:               &allowMergeOnSkippedPipeline,
					OnlyAllowMergeIfPipelineSucceeds:          &onlyAllowMergeIfPipelineSucceeds,
					OnlyAllowMergeIfAllDiscussionsAreResolved: &OnlyAllowMergeIfAllDiscussionsAreResolved,
					MergeMethod:                               &mergeMethodv1alpha1,
					RemoveSourceBranchAfterMerge:              &removeSourceBranchAfterMerge,
					LFSEnabled:                                &lfsEnabled,
					RequestAccessEnabled:                      &requestAccessEnabled,
					Topics:                                    topics,
					BuildGitStrategy:                          &buildGitStategy,
					BuildTimeout:                              &buildTimeout,
					AutoCancelPendingPipelines:                &autoCancelPendingPipelines,
					BuildCoverageRegex:                        &buildCoverageRegex,
					CIConfigPath:                              &ciConfigPath,
					CIForwardDeploymentEnabled:                &ciForwardDeploymentEnabled,
					CIDefaultGitDepth:                         &ciDefaultGitDepth,
					AutoDevopsEnabled:                         &autoDevopsEnabled,
					AutoDevopsDeployStrategy:                  &autoDevopsDeployStrategy,
					ExternalAuthorizationClassificationLabel:  &externalAuthorizationClassificationLabel,
					Mirror:                                    &mirror,
					MirrorUserID:                              &mirrorUserID,
					MirrorTriggerBuilds:                       &mirrorTriggerBuilds,
					OnlyMirrorProtectedBranches:               &onlyMirrorProtectedBranches,
					MirrorOverwritesDivergedBranches:          &mirrorOverwritesDivergedBranches,
					PackagesEnabled:                           &packagesEnabled,
					ServiceDeskEnabled:                        &serviceDeskEnabled,
					AutocloseReferencedIssues:                 &autocloseReferencedIssues,
					SuggestionCommitMessage:                   &suggestionCommitMessage,
					IssuesTemplate:                            &issuesTemplate,
					MergeRequestsTemplate:                     &mergeRequestsTemplate,
				},
			},
			want: &gitlab.EditProjectOptions{
				Name:                                &name,
				Path:                                &path,
				DefaultBranch:                       &defaultBranch,
				Description:                         &description,
				IssuesAccessLevel:                   clients.AccessControlValueStringToGitlab(issuesAccessLevel),
				RepositoryAccessLevel:               clients.AccessControlValueStringToGitlab(repositoryAccessLevel),
				MergeRequestsAccessLevel:            clients.AccessControlValueStringToGitlab(mergeRequestsAccessLevel),
				ForkingAccessLevel:                  clients.AccessControlValueStringToGitlab(forkingAccessLevel),
				BuildsAccessLevel:                   clients.AccessControlValueStringToGitlab(buildsAccessLevel),
				WikiAccessLevel:                     clients.AccessControlValueStringToGitlab(wikiAccessLevel),
				SnippetsAccessLevel:                 clients.AccessControlValueStringToGitlab(snippetsAccessLevel),
				OperationsAccessLevel:               clients.AccessControlValueStringToGitlab(operationsAccessLevel),
				EmailsDisabled:                      &emailsDisabled,
				PagesAccessLevel:                    clients.AccessControlValueStringToGitlab(pagesAccessLevel),
				ResolveOutdatedDiffDiscussions:      &resolveOutdatedDiffDiscussions,
				ContainerExpirationPolicyAttributes: &gitlabContainerExpirationPolicyAttributes,
				ContainerRegistryAccessLevel:        clients.AccessControlValueStringToGitlab(containerRegistryAccessLevel),
				SharedRunnersEnabled:                &sharedRunnersEnabled,
				Visibility:                          clients.VisibilityValueStringToGitlab(visibility),
				ImportURL:                           &importURL,
				PublicJobs:                          &publicBuilds,
				AllowMergeOnSkippedPipeline:         &allowMergeOnSkippedPipeline,
				OnlyAllowMergeIfPipelineSucceeds:    &onlyAllowMergeIfPipelineSucceeds,
				OnlyAllowMergeIfAllDiscussionsAreResolved: &OnlyAllowMergeIfAllDiscussionsAreResolved,
				MergeMethod:                              clients.MergeMethodStringToGitlab(mergeMethod),
				RemoveSourceBranchAfterMerge:             &removeSourceBranchAfterMerge,
				LFSEnabled:                               &lfsEnabled,
				RequestAccessEnabled:                     &requestAccessEnabled,
				TagList:                                  ptr.To[[]string](nil),
				Topics:                                   &topics,
				BuildGitStrategy:                         &buildGitStategy,
				BuildTimeout:                             &buildTimeout,
				AutoCancelPendingPipelines:               &autoCancelPendingPipelines,
				BuildCoverageRegex:                       &buildCoverageRegex,
				CIConfigPath:                             &ciConfigPath,
				CIForwardDeploymentEnabled:               &ciForwardDeploymentEnabled,
				CIDefaultGitDepth:                        &ciDefaultGitDepth,
				AutoDevopsEnabled:                        &autoDevopsEnabled,
				AutoDevopsDeployStrategy:                 &autoDevopsDeployStrategy,
				ExternalAuthorizationClassificationLabel: &externalAuthorizationClassificationLabel,
				Mirror:                                   &mirror,
				MirrorUserID:                             &mirrorUserID,
				MirrorTriggerBuilds:                      &mirrorTriggerBuilds,
				OnlyMirrorProtectedBranches:              &onlyMirrorProtectedBranches,
				MirrorOverwritesDivergedBranches:         &mirrorOverwritesDivergedBranches,
				PackagesEnabled:                          &packagesEnabled,
				ServiceDeskEnabled:                       &serviceDeskEnabled,
				AutocloseReferencedIssues:                &autocloseReferencedIssues,
				SuggestionCommitMessage:                  &suggestionCommitMessage,
				IssuesTemplate:                           &issuesTemplate,
				MergeRequestsTemplate:                    &mergeRequestsTemplate,
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
					Topics:                         topics,
					BuildTimeout:                   &buildTimeout,
				},
			},
			want: &gitlab.EditProjectOptions{
				Name:                           &name,
				Path:                           &path,
				IssuesAccessLevel:              clients.AccessControlValueStringToGitlab(issuesAccessLevel),
				ResolveOutdatedDiffDiscussions: &resolveOutdatedDiffDiscussions,
				MergeMethod:                    clients.MergeMethodStringToGitlab(mergeMethod),
				TagList:                        ptr.To[[]string](nil),
				Topics:                         &topics,
				BuildTimeout:                   &buildTimeout,
			},
		},
		"NameOverride": {
			args: args{
				name: name,
				parameters: &v1alpha1.ProjectParameters{
					Name: &name,
					// TODO: TagList specified and wanted because otherwise
					// test does not pass - (nil vs &nil - fix)
					Topics: topics,
				},
			},
			want: &gitlab.EditProjectOptions{
				Name:    &name,
				TagList: ptr.To[[]string](nil),
				Topics:  &topics,
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
