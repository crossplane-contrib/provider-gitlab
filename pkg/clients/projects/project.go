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
	"strings"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
)

const (
	errProjectNotFound = "404 Project Not Found"
)

// Client defines Gitlab Project service operations
type Client interface {
	GetProject(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error)
	CreateProject(opt *gitlab.CreateProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error)
	EditProject(pid interface{}, opt *gitlab.EditProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error)
	DeleteProject(pid interface{}, opt *gitlab.DeleteProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

	GetProjectPushRules(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error)
	EditProjectPushRule(pid interface{}, opt *gitlab.EditProjectPushRuleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error)
}

// NewProjectClient returns a new Gitlab Project service
func NewProjectClient(cfg clients.Config) Client {
	git := clients.NewClient(cfg)
	return git.Projects
}

// IsErrorProjectNotFound helper function to test for errProjectNotFound error.
func IsErrorProjectNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errProjectNotFound)
}

// GenerateObservation is used to produce v1alpha1.ProjectObservation from
// gitlab.Project.
func GenerateObservation(prj *gitlab.Project) v1alpha1.ProjectObservation { //nolint:gocyclo
	if prj == nil {
		return v1alpha1.ProjectObservation{}
	}

	o := v1alpha1.ProjectObservation{
		ID:                       prj.ID,
		Public:                   prj.PublicJobs,
		SSHURLToRepo:             prj.SSHURLToRepo,
		HTTPURLToRepo:            prj.HTTPURLToRepo,
		WebURL:                   prj.WebURL,
		ReadmeURL:                prj.ReadmeURL,
		NameWithNamespace:        prj.NameWithNamespace,
		PathWithNamespace:        prj.PathWithNamespace,
		IssuesAccessLevel:        v1alpha1.AccessControlValue(prj.IssuesAccessLevel),
		OpenIssuesCount:          prj.OpenIssuesCount,
		MergeRequestsAccessLevel: v1alpha1.AccessControlValue(prj.MergeRequestsAccessLevel),
		BuildsAccessLevel:        v1alpha1.AccessControlValue(prj.BuildsAccessLevel),
		WikiAccessLevel:          v1alpha1.AccessControlValue(prj.WikiAccessLevel),
		SnippetsAccessLevel:      v1alpha1.AccessControlValue(prj.SnippetsAccessLevel),
		CreatorID:                prj.CreatorID,
		ImportStatus:             prj.ImportStatus,
		ImportError:              prj.ImportError,
		Archived:                 prj.Archived,
		ForksCount:               prj.ForksCount,
		StarCount:                prj.StarCount,
		EmptyRepo:                prj.EmptyRepo,
		AvatarURL:                prj.AvatarURL,
		LicenseURL:               prj.LicenseURL,
		ServiceDeskAddress:       prj.ServiceDeskAddress,
	}

	if prj.ContainerExpirationPolicy != nil {
		o.ContainerExpirationPolicy = &v1alpha1.ContainerExpirationPolicy{
			Cadence:         prj.ContainerExpirationPolicy.Cadence,
			KeepN:           prj.ContainerExpirationPolicy.KeepN,
			OlderThan:       prj.ContainerExpirationPolicy.OlderThan,
			NameRegexDelete: prj.ContainerExpirationPolicy.NameRegexDelete,
			NameRegexKeep:   prj.ContainerExpirationPolicy.NameRegexKeep,
			Enabled:         prj.ContainerExpirationPolicy.Enabled,
			NextRunAt:       &metav1.Time{Time: *prj.ContainerExpirationPolicy.NextRunAt},
		}
	}

	if prj.License != nil {
		o.License = &v1alpha1.ProjectLicense{
			Key:       prj.License.Key,
			Name:      prj.License.Name,
			Nickname:  prj.License.Nickname,
			HTMLURL:   prj.License.HTMLURL,
			SourceURL: prj.License.SourceURL,
		}
	}

	if prj.CreatedAt != nil {
		o.CreatedAt = &metav1.Time{Time: *prj.CreatedAt}
	}
	if prj.LastActivityAt != nil {
		o.LastActivityAt = &metav1.Time{Time: *prj.LastActivityAt}
	}
	if prj.MarkedForDeletionOn != nil {
		o.MarkedForDeletionOn = &metav1.Time{Time: time.Time(*prj.MarkedForDeletionOn)}
	}

	if len(o.ComplianceFrameworks) == 0 && len(prj.ComplianceFrameworks) > 0 {
		o.ComplianceFrameworks = prj.ComplianceFrameworks
	}

	if len(o.CustomAttributes) == 0 && len(prj.CustomAttributes) > 0 {
		o.CustomAttributes = make([]v1alpha1.CustomAttribute, len(prj.CustomAttributes))
		for i, c := range prj.CustomAttributes {
			o.CustomAttributes[i].Key = c.Key
			o.CustomAttributes[i].Value = c.Value
		}
	}

	if prj.Statistics != nil {
		o.Statistics = &v1alpha1.ProjectStatistics{
			StorageStatistics: v1alpha1.StorageStatistics{
				StorageSize:      prj.Statistics.StorageSize,
				RepositorySize:   prj.Statistics.RepositorySize,
				LfsObjectsSize:   prj.Statistics.LFSObjectsSize,
				JobArtifactsSize: prj.Statistics.JobArtifactsSize,
			},
		}
	}

	if prj.Links != nil {
		o.Links = &v1alpha1.Links{
			Self:          prj.Links.Self,
			Issues:        prj.Links.Issues,
			MergeRequests: prj.Links.MergeRequests,
			RepoBranches:  prj.Links.RepoBranches,
			Labels:        prj.Links.Labels,
			Events:        prj.Links.Events,
			Members:       prj.Links.Members,
		}
	}

	if len(o.SharedWithGroups) == 0 && len(prj.SharedWithGroups) > 0 {
		o.SharedWithGroups = make([]v1alpha1.SharedWithGroups, len(prj.SharedWithGroups))
		for i, s := range prj.SharedWithGroups {
			o.SharedWithGroups[i].GroupID = s.GroupID
			o.SharedWithGroups[i].GroupName = s.GroupName
			o.SharedWithGroups[i].GroupAccessLevel = s.GroupAccessLevel
		}
	}

	if prj.ForkedFromProject != nil {
		o.ForkedFromProject = &v1alpha1.ForkParent{
			HTTPURLToRepo:     prj.ForkedFromProject.HTTPURLToRepo,
			ID:                prj.ForkedFromProject.ID,
			Name:              prj.ForkedFromProject.Name,
			NameWithNamespace: prj.ForkedFromProject.NameWithNamespace,
			Path:              prj.ForkedFromProject.Path,
			PathWithNamespace: prj.ForkedFromProject.PathWithNamespace,
			WebURL:            prj.ForkedFromProject.WebURL,
		}
	}

	if prj.Permissions != nil {
		o.Permissions = &v1alpha1.Permissions{}
		if prj.Permissions.ProjectAccess != nil {
			o.Permissions.ProjectAccess = &v1alpha1.ProjectAccess{
				AccessLevel:       v1alpha1.AccessLevelValue(prj.Permissions.ProjectAccess.AccessLevel),
				NotificationLevel: v1alpha1.NotificationLevelValue(prj.Permissions.ProjectAccess.NotificationLevel),
			}
		}
		if prj.Permissions.GroupAccess != nil {
			o.Permissions.GroupAccess = &v1alpha1.GroupAccess{
				AccessLevel:       v1alpha1.AccessLevelValue(prj.Permissions.GroupAccess.AccessLevel),
				NotificationLevel: v1alpha1.NotificationLevelValue(prj.Permissions.GroupAccess.NotificationLevel),
			}
		}
	}

	if prj.Namespace != nil {
		o.Namespace = &v1alpha1.ProjectNamespace{
			ID:        prj.Namespace.ID,
			Name:      prj.Namespace.Name,
			Path:      prj.Namespace.Path,
			Kind:      prj.Namespace.Kind,
			FullPath:  prj.Namespace.FullPath,
			AvatarURL: prj.Namespace.AvatarURL,
			WebURL:    prj.Namespace.WebURL,
		}
	}

	if prj.Owner != nil {
		o.Owner = GenerateOwnerObservation(prj.Owner)
	}

	return o
}

// GenerateOwnerObservation generates v1alpha.User from gitlab.User.
func GenerateOwnerObservation(usr *gitlab.User) *v1alpha1.User {
	o := &v1alpha1.User{
		ID:                        usr.ID,
		Username:                  usr.Username,
		Email:                     usr.Email,
		Name:                      usr.Name,
		State:                     usr.Name,
		WebURL:                    usr.WebURL,
		Bio:                       usr.Bio,
		Location:                  usr.Location,
		PublicEmail:               usr.PublicEmail,
		Skype:                     usr.Skype,
		Linkedin:                  usr.Linkedin,
		Twitter:                   usr.Twitter,
		WebsiteURL:                usr.WebsiteURL,
		Organization:              usr.Organization,
		ExternUID:                 usr.ExternUID,
		Provider:                  usr.Provider,
		ThemeID:                   usr.ThemeID,
		ColorSchemeID:             usr.ColorSchemeID,
		IsAdmin:                   usr.IsAdmin,
		AvatarURL:                 usr.AvatarURL,
		CanCreateGroup:            usr.CanCreateGroup,
		CanCreateProject:          usr.CanCreateProject,
		ProjectsLimit:             usr.ProjectsLimit,
		TwoFactorEnabled:          usr.TwoFactorEnabled,
		External:                  usr.External,
		PrivateProfile:            usr.PrivateProfile,
		SharedRunnersMinutesLimit: usr.SharedRunnersMinutesLimit,
	}
	if usr.CreatedAt != nil {
		o.CreatedAt = &metav1.Time{Time: *usr.CreatedAt}
	}
	if usr.LastActivityOn != nil {
		o.LastActivityOn = &metav1.Time{Time: time.Time(*usr.LastActivityOn)}
	}
	if usr.CurrentSignInAt != nil {
		o.CurrentSignInAt = &metav1.Time{Time: *usr.CurrentSignInAt}
	}
	if usr.LastSignInAt != nil {
		o.LastSignInAt = &metav1.Time{Time: *usr.LastSignInAt}
	}
	if usr.ConfirmedAt != nil {
		o.ConfirmedAt = &metav1.Time{Time: *usr.ConfirmedAt}
	}
	for i, c := range usr.CustomAttributes {
		o.CustomAttributes[i].Key = c.Key
		o.CustomAttributes[i].Value = c.Value
	}
	for i, id := range usr.Identities {
		o.Identities[i].Provider = id.Provider
		o.Identities[i].ExternUID = id.ExternUID
	}

	return o
}

// GenerateCreateProjectOptions generates project creation options
func GenerateCreateProjectOptions(name string, p *v1alpha1.ProjectParameters) *gitlab.CreateProjectOptions {
	// Name field overrides resource name
	if p.Name != nil {
		name = *p.Name
	}
	project := &gitlab.CreateProjectOptions{
		Name:                                &name,
		Path:                                p.Path,
		NamespaceID:                         p.NamespaceID,
		DefaultBranch:                       p.DefaultBranch,
		Description:                         p.Description,
		IssuesAccessLevel:                   clients.AccessControlValueV1alpha1ToGitlab(p.IssuesAccessLevel),
		RepositoryAccessLevel:               clients.AccessControlValueV1alpha1ToGitlab(p.RepositoryAccessLevel),
		MergeRequestsAccessLevel:            clients.AccessControlValueV1alpha1ToGitlab(p.MergeRequestsAccessLevel),
		ForkingAccessLevel:                  clients.AccessControlValueV1alpha1ToGitlab(p.ForkingAccessLevel),
		BuildsAccessLevel:                   clients.AccessControlValueV1alpha1ToGitlab(p.BuildsAccessLevel),
		WikiAccessLevel:                     clients.AccessControlValueV1alpha1ToGitlab(p.WikiAccessLevel),
		SnippetsAccessLevel:                 clients.AccessControlValueV1alpha1ToGitlab(p.SnippetsAccessLevel),
		PagesAccessLevel:                    clients.AccessControlValueV1alpha1ToGitlab(p.PagesAccessLevel),
		OperationsAccessLevel:               clients.AccessControlValueV1alpha1ToGitlab(p.OperationsAccessLevel),
		EmailsDisabled:                      p.EmailsDisabled,
		ResolveOutdatedDiffDiscussions:      p.ResolveOutdatedDiffDiscussions,
		ContainerExpirationPolicyAttributes: clients.ContainerExpirationPolicyAttributesV1alpha1ToGitlab(p.ContainerExpirationPolicyAttributes),
		ContainerRegistryAccessLevel:        clients.AccessControlValueV1alpha1ToGitlab(p.ContainerRegistryAccessLevel),
		SharedRunnersEnabled:                p.SharedRunnersEnabled,
		Visibility:                          clients.VisibilityValueV1alpha1ToGitlab(p.Visibility),
		ImportURL:                           p.ImportURL,
		PublicBuilds:                        p.PublicBuilds,
		AllowMergeOnSkippedPipeline:         p.AllowMergeOnSkippedPipeline,
		OnlyAllowMergeIfPipelineSucceeds:    p.OnlyAllowMergeIfPipelineSucceeds,
		OnlyAllowMergeIfAllDiscussionsAreResolved: p.OnlyAllowMergeIfAllDiscussionsAreResolved,
		MergeMethod:                              clients.MergeMethodV1alpha1ToGitlab(p.MergeMethod),
		RemoveSourceBranchAfterMerge:             p.RemoveSourceBranchAfterMerge,
		LFSEnabled:                               p.LFSEnabled,
		RequestAccessEnabled:                     p.RequestAccessEnabled,
		Topics:                                   &p.Topics,
		PrintingMergeRequestLinkEnabled:          p.PrintingMergeRequestLinkEnabled,
		BuildGitStrategy:                         p.BuildGitStrategy,
		BuildTimeout:                             p.BuildTimeout,
		AutoCancelPendingPipelines:               p.AutoCancelPendingPipelines,
		BuildCoverageRegex:                       p.BuildCoverageRegex,
		CIConfigPath:                             p.CIConfigPath,
		CIForwardDeploymentEnabled:               p.CIForwardDeploymentEnabled,
		AutoDevopsEnabled:                        p.AutoDevopsEnabled,
		AutoDevopsDeployStrategy:                 p.AutoDevopsDeployStrategy,
		ExternalAuthorizationClassificationLabel: p.ExternalAuthorizationClassificationLabel,
		Mirror:                                   p.Mirror,
		MirrorTriggerBuilds:                      p.MirrorTriggerBuilds,
		InitializeWithReadme:                     p.InitializeWithReadme,
		TemplateName:                             p.TemplateName,
		TemplateProjectID:                        p.TemplateProjectID,
		UseCustomTemplate:                        p.UseCustomTemplate,
		GroupWithProjectTemplatesID:              p.GroupWithProjectTemplatesID,
		PackagesEnabled:                          p.PackagesEnabled,
		ServiceDeskEnabled:                       p.ServiceDeskEnabled,
		AutocloseReferencedIssues:                p.AutocloseReferencedIssues,
		SuggestionCommitMessage:                  p.SuggestionCommitMessage,
		IssuesTemplate:                           p.IssuesTemplate,
		MergeRequestsTemplate:                    p.MergeRequestsTemplate,
	}
	return project
}

// GenerateEditProjectOptions generates project edit options
func GenerateEditProjectOptions(name string, p *v1alpha1.ProjectParameters) *gitlab.EditProjectOptions {
	// Name field overrides resource name
	if p.Name != nil {
		name = *p.Name
	}
	o := &gitlab.EditProjectOptions{
		Name:                                &name,
		Path:                                p.Path,
		DefaultBranch:                       p.DefaultBranch,
		Description:                         p.Description,
		IssuesAccessLevel:                   clients.AccessControlValueV1alpha1ToGitlab(p.IssuesAccessLevel),
		RepositoryAccessLevel:               clients.AccessControlValueV1alpha1ToGitlab(p.RepositoryAccessLevel),
		MergeRequestsAccessLevel:            clients.AccessControlValueV1alpha1ToGitlab(p.MergeRequestsAccessLevel),
		ForkingAccessLevel:                  clients.AccessControlValueV1alpha1ToGitlab(p.ForkingAccessLevel),
		BuildsAccessLevel:                   clients.AccessControlValueV1alpha1ToGitlab(p.BuildsAccessLevel),
		WikiAccessLevel:                     clients.AccessControlValueV1alpha1ToGitlab(p.WikiAccessLevel),
		SnippetsAccessLevel:                 clients.AccessControlValueV1alpha1ToGitlab(p.SnippetsAccessLevel),
		PagesAccessLevel:                    clients.AccessControlValueV1alpha1ToGitlab(p.PagesAccessLevel),
		OperationsAccessLevel:               clients.AccessControlValueV1alpha1ToGitlab(p.OperationsAccessLevel),
		EmailsDisabled:                      p.EmailsDisabled,
		ResolveOutdatedDiffDiscussions:      p.ResolveOutdatedDiffDiscussions,
		ContainerExpirationPolicyAttributes: clients.ContainerExpirationPolicyAttributesV1alpha1ToGitlab(p.ContainerExpirationPolicyAttributes),
		ContainerRegistryAccessLevel:        clients.AccessControlValueV1alpha1ToGitlab(p.ContainerRegistryAccessLevel),
		SharedRunnersEnabled:                p.SharedRunnersEnabled,
		Visibility:                          clients.VisibilityValueV1alpha1ToGitlab(p.Visibility),
		ImportURL:                           p.ImportURL,
		PublicBuilds:                        p.PublicBuilds,
		AllowMergeOnSkippedPipeline:         p.AllowMergeOnSkippedPipeline,
		OnlyAllowMergeIfPipelineSucceeds:    p.OnlyAllowMergeIfPipelineSucceeds,
		OnlyAllowMergeIfAllDiscussionsAreResolved: p.OnlyAllowMergeIfAllDiscussionsAreResolved,
		MergeMethod:                              clients.MergeMethodV1alpha1ToGitlab(p.MergeMethod),
		RemoveSourceBranchAfterMerge:             p.RemoveSourceBranchAfterMerge,
		LFSEnabled:                               p.LFSEnabled,
		RequestAccessEnabled:                     p.RequestAccessEnabled,
		Topics:                                   &p.Topics,
		BuildGitStrategy:                         p.BuildGitStrategy,
		BuildTimeout:                             p.BuildTimeout,
		AutoCancelPendingPipelines:               p.AutoCancelPendingPipelines,
		BuildCoverageRegex:                       p.BuildCoverageRegex,
		CIConfigPath:                             p.CIConfigPath,
		CIForwardDeploymentEnabled:               p.CIForwardDeploymentEnabled,
		CIDefaultGitDepth:                        p.CIDefaultGitDepth,
		AutoDevopsEnabled:                        p.AutoDevopsEnabled,
		AutoDevopsDeployStrategy:                 p.AutoDevopsDeployStrategy,
		ExternalAuthorizationClassificationLabel: p.ExternalAuthorizationClassificationLabel,
		Mirror:                                   p.Mirror,
		MirrorUserID:                             p.MirrorUserID,
		MirrorTriggerBuilds:                      p.MirrorTriggerBuilds,
		OnlyMirrorProtectedBranches:              p.OnlyMirrorProtectedBranches,
		MirrorOverwritesDivergedBranches:         p.MirrorOverwritesDivergedBranches,
		PackagesEnabled:                          p.PackagesEnabled,
		ServiceDeskEnabled:                       p.ServiceDeskEnabled,
		AutocloseReferencedIssues:                p.AutocloseReferencedIssues,
		SuggestionCommitMessage:                  p.SuggestionCommitMessage,
		IssuesTemplate:                           p.IssuesTemplate,
		MergeRequestsTemplate:                    p.MergeRequestsTemplate,
	}
	return o
}

func GenerateEditPushRulesOptions(p *v1alpha1.ProjectParameters) *gitlab.EditProjectPushRuleOptions {
	o := &gitlab.EditProjectPushRuleOptions{}
	if p.PushRules != nil {
		o.AuthorEmailRegex = p.PushRules.AuthorEmailRegex
		o.BranchNameRegex = p.PushRules.BranchNameRegex
		o.CommitCommitterCheck = p.PushRules.CommitCommitterCheck
		o.CommitCommitterNameCheck = p.PushRules.CommitCommitterNameCheck
		o.CommitMessageNegativeRegex = p.PushRules.CommitMessageNegativeRegex
		o.CommitMessageRegex = p.PushRules.CommitMessageRegex
		o.DenyDeleteTag = p.PushRules.DenyDeleteTag
		o.FileNameRegex = p.PushRules.FileNameRegex
		o.MaxFileSize = p.PushRules.MaxFileSize
		o.MemberCheck = p.PushRules.MemberCheck
		o.PreventSecrets = p.PushRules.PreventSecrets
		o.RejectNonDCOCommits = p.PushRules.RejectNonDCOCommits
		o.RejectUnsignedCommits = p.PushRules.RejectUnsignedCommits
	}
	return o
}
