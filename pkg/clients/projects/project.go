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

	"github.com/xanzy/go-gitlab"
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
	DeleteProject(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
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
func GenerateObservation(prj *gitlab.Project) v1alpha1.ProjectObservation { // nolint:gocyclo
	if prj == nil {
		return v1alpha1.ProjectObservation{}
	}

	o := v1alpha1.ProjectObservation{
		ID:                   prj.ID,
		Public:               prj.Public,
		SSHURLToRepo:         prj.SSHURLToRepo,
		HTTPURLToRepo:        prj.HTTPURLToRepo,
		WebURL:               prj.WebURL,
		ReadmeURL:            prj.ReadmeURL,
		NameWithNamespace:    prj.NameWithNamespace,
		PathWithNamespace:    prj.PathWithNamespace,
		IssuesEnabled:        prj.IssuesEnabled,
		OpenIssuesCount:      prj.OpenIssuesCount,
		MergeRequestsEnabled: prj.MergeRequestsEnabled,
		JobsEnabled:          prj.JobsEnabled,
		WikiEnabled:          prj.WikiEnabled,
		SnippetsEnabled:      prj.SnippetsEnabled,
		CreatorID:            prj.CreatorID,
		ImportStatus:         prj.ImportStatus,
		ImportError:          prj.ImportError,
		Archived:             prj.Archived,
		ForksCount:           prj.ForksCount,
		StarCount:            prj.StarCount,
		RunnersToken:         prj.RunnersToken,
		EmptyRepo:            prj.EmptyRepo,
		AvatarURL:            prj.AvatarURL,
		LicenseURL:           prj.LicenseURL,
		ServiceDeskAddress:   prj.ServiceDeskAddress,
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
	if prj.MarkedForDeletionAt != nil {
		o.MarkedForDeletionAt = &metav1.Time{Time: time.Time(*prj.MarkedForDeletionAt)}
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
				LfsObjectsSize:   prj.Statistics.LfsObjectsSize,
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
		o.Owner = &v1alpha1.User{
			ID:                        prj.Owner.ID,
			Username:                  prj.Owner.Username,
			Email:                     prj.Owner.Email,
			Name:                      prj.Owner.Name,
			State:                     prj.Owner.Name,
			WebURL:                    prj.Owner.WebURL,
			Bio:                       prj.Owner.Bio,
			Location:                  prj.Owner.Location,
			PublicEmail:               prj.Owner.PublicEmail,
			Skype:                     prj.Owner.Skype,
			Linkedin:                  prj.Owner.Linkedin,
			Twitter:                   prj.Owner.Twitter,
			WebsiteURL:                prj.Owner.WebsiteURL,
			Organization:              prj.Owner.Organization,
			ExternUID:                 prj.Owner.ExternUID,
			Provider:                  prj.Owner.Provider,
			ThemeID:                   prj.Owner.ThemeID,
			ColorSchemeID:             prj.Owner.ColorSchemeID,
			IsAdmin:                   prj.Owner.IsAdmin,
			AvatarURL:                 prj.Owner.AvatarURL,
			CanCreateGroup:            prj.Owner.CanCreateGroup,
			CanCreateProject:          prj.Owner.CanCreateProject,
			ProjectsLimit:             prj.Owner.ProjectsLimit,
			TwoFactorEnabled:          prj.Owner.TwoFactorEnabled,
			External:                  prj.Owner.External,
			PrivateProfile:            prj.Owner.PrivateProfile,
			SharedRunnersMinutesLimit: prj.Owner.SharedRunnersMinutesLimit,
		}
		if prj.Owner.CreatedAt != nil {
			o.Owner.CreatedAt = &metav1.Time{Time: *prj.Owner.CreatedAt}
		}
		if prj.Owner.LastActivityOn != nil {
			o.Owner.LastActivityOn = &metav1.Time{Time: time.Time(*prj.Owner.LastActivityOn)}
		}
		if prj.Owner.CurrentSignInAt != nil {
			o.Owner.CurrentSignInAt = &metav1.Time{Time: *prj.Owner.CurrentSignInAt}
		}
		if prj.Owner.LastSignInAt != nil {
			o.Owner.LastSignInAt = &metav1.Time{Time: *prj.Owner.LastSignInAt}
		}
		if prj.Owner.ConfirmedAt != nil {
			o.Owner.ConfirmedAt = &metav1.Time{Time: *prj.Owner.ConfirmedAt}
		}
		for i, c := range prj.Owner.CustomAttributes {
			o.Owner.CustomAttributes[i].Key = c.Key
			o.Owner.CustomAttributes[i].Value = c.Value
		}
		for i, id := range prj.Owner.Identities {
			o.Owner.Identities[i].Provider = id.Provider
			o.Owner.Identities[i].ExternUID = id.ExternUID
		}
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
		ContainerRegistryEnabled:            p.ContainerRegistryEnabled,
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
		TagList:                                  &p.TagList,
		PrintingMergeRequestLinkEnabled:          p.PrintingMergeRequestLinkEnabled,
		BuildGitStrategy:                         p.BuildGitStrategy,
		BuildTimeout:                             p.BuildTimeout,
		AutoCancelPendingPipelines:               p.AutoCancelPendingPipelines,
		BuildCoverageRegex:                       p.BuildCoverageRegex,
		CIConfigPath:                             p.CIConfigPath,
		CIForwardDeploymentEnabled:               p.CIForwardDeploymentEnabled,
		AutoDevopsEnabled:                        p.AutoDevopsEnabled,
		AutoDevopsDeployStrategy:                 p.AutoDevopsDeployStrategy,
		ApprovalsBeforeMerge:                     p.ApprovalsBeforeMerge,
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
		ContainerRegistryEnabled:            p.ContainerRegistryEnabled,
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
		TagList:                                  &p.TagList,
		BuildGitStrategy:                         p.BuildGitStrategy,
		BuildTimeout:                             p.BuildTimeout,
		AutoCancelPendingPipelines:               p.AutoCancelPendingPipelines,
		BuildCoverageRegex:                       p.BuildCoverageRegex,
		CIConfigPath:                             p.CIConfigPath,
		CIForwardDeploymentEnabled:               p.CIForwardDeploymentEnabled,
		CIDefaultGitDepth:                        p.CIDefaultGitDepth,
		AutoDevopsEnabled:                        p.AutoDevopsEnabled,
		AutoDevopsDeployStrategy:                 p.AutoDevopsDeployStrategy,
		ApprovalsBeforeMerge:                     p.ApprovalsBeforeMerge,
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
