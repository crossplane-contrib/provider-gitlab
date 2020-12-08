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
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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

// LateInitialize fills the empty fields in the project spec with the
// values seen in gitlab.Project.
func LateInitialize(in *v1alpha1.ProjectParameters, project *gitlab.Project) { // nolint:gocyclo
	if project == nil {
		return
	}
	in.Path = clients.LateInitializeStringPtr(in.Path, project.Path)
	in.DefaultBranch = clients.LateInitializeStringPtr(in.DefaultBranch, project.DefaultBranch)
	in.Description = clients.LateInitializeStringPtr(in.Description, project.Description)
	in.IssuesAccessLevel = clients.LateInitializeAccessControlValue(in.IssuesAccessLevel, project.IssuesAccessLevel)
	in.RepositoryAccessLevel = clients.LateInitializeAccessControlValue(in.RepositoryAccessLevel, project.RepositoryAccessLevel)
	in.MergeRequestsAccessLevel = clients.LateInitializeAccessControlValue(in.MergeRequestsAccessLevel, project.MergeRequestsAccessLevel)
	in.ForkingAccessLevel = clients.LateInitializeAccessControlValue(in.ForkingAccessLevel, project.ForkingAccessLevel)
	in.BuildsAccessLevel = clients.LateInitializeAccessControlValue(in.BuildsAccessLevel, project.BuildsAccessLevel)
	in.WikiAccessLevel = clients.LateInitializeAccessControlValue(in.WikiAccessLevel, project.WikiAccessLevel)
	in.SnippetsAccessLevel = clients.LateInitializeAccessControlValue(in.SnippetsAccessLevel, project.SnippetsAccessLevel)
	in.PagesAccessLevel = clients.LateInitializeAccessControlValue(in.PagesAccessLevel, project.PagesAccessLevel)
	if in.ResolveOutdatedDiffDiscussions == nil {
		in.ResolveOutdatedDiffDiscussions = &project.ResolveOutdatedDiffDiscussions
	}
	if in.ContainerRegistryEnabled == nil {
		in.ContainerRegistryEnabled = &project.ContainerRegistryEnabled
	}
	if in.SharedRunnersEnabled == nil {
		in.SharedRunnersEnabled = &project.SharedRunnersEnabled
	}
	in.Visibility = clients.LateInitializeVisibilityValue(in.Visibility, project.Visibility)
	if in.PublicBuilds == nil {
		in.PublicBuilds = &project.PublicBuilds
	}
	if in.OnlyAllowMergeIfPipelineSucceeds == nil {
		in.OnlyAllowMergeIfPipelineSucceeds = &project.OnlyAllowMergeIfPipelineSucceeds
	}
	if in.OnlyAllowMergeIfAllDiscussionsAreResolved == nil {
		in.OnlyAllowMergeIfAllDiscussionsAreResolved = &project.OnlyAllowMergeIfAllDiscussionsAreResolved
	}
	if in.RemoveSourceBranchAfterMerge == nil {
		in.RemoveSourceBranchAfterMerge = &project.RemoveSourceBranchAfterMerge
	}
	if in.LFSEnabled == nil {
		in.LFSEnabled = &project.LFSEnabled
	}
	if in.RequestAccessEnabled == nil {
		in.RequestAccessEnabled = &project.RequestAccessEnabled
	}
	in.MergeMethod = clients.LateInitializeMergeMethodValue(in.MergeMethod, project.MergeMethod)
	if len(in.TagList) == 0 && len(project.TagList) > 0 {
		in.TagList = project.TagList
	}
	in.CIConfigPath = clients.LateInitializeStringPtr(in.CIConfigPath, project.CIConfigPath)
	if in.CIDefaultGitDepth == nil {
		in.CIDefaultGitDepth = &project.CIDefaultGitDepth
	}
	if in.Mirror == nil {
		in.Mirror = &project.Mirror
	}
	if in.MirrorUserID == nil {
		in.MirrorUserID = &project.MirrorUserID
	}
	if in.MirrorTriggerBuilds == nil {
		in.MirrorTriggerBuilds = &project.MirrorTriggerBuilds
	}
	if in.OnlyMirrorProtectedBranches == nil {
		in.OnlyMirrorProtectedBranches = &project.OnlyMirrorProtectedBranches
	}
	if in.MirrorOverwritesDivergedBranches == nil {
		in.MirrorOverwritesDivergedBranches = &project.MirrorOverwritesDivergedBranches
	}
	if in.PackagesEnabled == nil {
		in.PackagesEnabled = &project.PackagesEnabled
	}
	if in.ServiceDeskEnabled == nil {
		in.ServiceDeskEnabled = &project.ServiceDeskEnabled
	}
	if in.AutocloseReferencedIssues == nil {
		in.AutocloseReferencedIssues = &project.AutocloseReferencedIssues
	}
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
	project := &gitlab.CreateProjectOptions{
		Name:                             &name,
		Path:                             p.Path,
		NamespaceID:                      p.NamespaceID,
		DefaultBranch:                    p.DefaultBranch,
		Description:                      p.Description,
		IssuesAccessLevel:                clients.AccessControlValueV1alpha1ToGitlab(p.IssuesAccessLevel),
		RepositoryAccessLevel:            clients.AccessControlValueV1alpha1ToGitlab(p.RepositoryAccessLevel),
		MergeRequestsAccessLevel:         clients.AccessControlValueV1alpha1ToGitlab(p.MergeRequestsAccessLevel),
		ForkingAccessLevel:               clients.AccessControlValueV1alpha1ToGitlab(p.ForkingAccessLevel),
		BuildsAccessLevel:                clients.AccessControlValueV1alpha1ToGitlab(p.BuildsAccessLevel),
		WikiAccessLevel:                  clients.AccessControlValueV1alpha1ToGitlab(p.WikiAccessLevel),
		SnippetsAccessLevel:              clients.AccessControlValueV1alpha1ToGitlab(p.SnippetsAccessLevel),
		PagesAccessLevel:                 clients.AccessControlValueV1alpha1ToGitlab(p.PagesAccessLevel),
		EmailsDisabled:                   p.EmailsDisabled,
		ResolveOutdatedDiffDiscussions:   p.ResolveOutdatedDiffDiscussions,
		ContainerRegistryEnabled:         p.ContainerRegistryEnabled,
		SharedRunnersEnabled:             p.SharedRunnersEnabled,
		Visibility:                       clients.VisibilityValueV1alpha1ToGitlab(p.Visibility),
		ImportURL:                        p.ImportURL,
		PublicBuilds:                     p.PublicBuilds,
		OnlyAllowMergeIfPipelineSucceeds: p.OnlyAllowMergeIfPipelineSucceeds,
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
	}

	return project
}

// GenerateEditProjectOptions generates project edit options
func GenerateEditProjectOptions(name string, p *v1alpha1.ProjectParameters) *gitlab.EditProjectOptions {
	o := &gitlab.EditProjectOptions{
		Name:                             &name,
		Path:                             p.Path,
		DefaultBranch:                    p.DefaultBranch,
		Description:                      p.Description,
		IssuesAccessLevel:                clients.AccessControlValueV1alpha1ToGitlab(p.IssuesAccessLevel),
		RepositoryAccessLevel:            clients.AccessControlValueV1alpha1ToGitlab(p.RepositoryAccessLevel),
		MergeRequestsAccessLevel:         clients.AccessControlValueV1alpha1ToGitlab(p.MergeRequestsAccessLevel),
		ForkingAccessLevel:               clients.AccessControlValueV1alpha1ToGitlab(p.ForkingAccessLevel),
		BuildsAccessLevel:                clients.AccessControlValueV1alpha1ToGitlab(p.BuildsAccessLevel),
		WikiAccessLevel:                  clients.AccessControlValueV1alpha1ToGitlab(p.WikiAccessLevel),
		SnippetsAccessLevel:              clients.AccessControlValueV1alpha1ToGitlab(p.SnippetsAccessLevel),
		PagesAccessLevel:                 clients.AccessControlValueV1alpha1ToGitlab(p.PagesAccessLevel),
		EmailsDisabled:                   p.EmailsDisabled,
		ResolveOutdatedDiffDiscussions:   p.ResolveOutdatedDiffDiscussions,
		ContainerRegistryEnabled:         p.ContainerRegistryEnabled,
		SharedRunnersEnabled:             p.SharedRunnersEnabled,
		Visibility:                       clients.VisibilityValueV1alpha1ToGitlab(p.Visibility),
		ImportURL:                        p.ImportURL,
		PublicBuilds:                     p.PublicBuilds,
		OnlyAllowMergeIfPipelineSucceeds: p.OnlyAllowMergeIfPipelineSucceeds,
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
	}

	return o
}

// IsProjectUpToDate checks whether there is a change in any of the modifiable fields.
func IsProjectUpToDate(p *v1alpha1.ProjectParameters, g *gitlab.Project) bool { // nolint:gocyclo
	if !cmp.Equal(p.Path, clients.StringToPtr(g.Path)) {
		return false
	}
	if !cmp.Equal(p.DefaultBranch, clients.StringToPtr(g.DefaultBranch)) {
		return false
	}
	if !cmp.Equal(p.Description, clients.StringToPtr(g.Description)) {
		return false
	}
	if p.IssuesAccessLevel != nil {
		if !cmp.Equal(string(*p.IssuesAccessLevel), string(g.IssuesAccessLevel)) {
			return false
		}
	}
	if p.RepositoryAccessLevel != nil {
		if !cmp.Equal(string(*p.RepositoryAccessLevel), string(g.RepositoryAccessLevel)) {
			return false
		}
	}
	if p.MergeRequestsAccessLevel != nil {
		if !cmp.Equal(string(*p.MergeRequestsAccessLevel), string(g.MergeRequestsAccessLevel)) {
			return false
		}
	}
	if p.ForkingAccessLevel != nil {
		if !cmp.Equal(string(*p.ForkingAccessLevel), string(g.ForkingAccessLevel)) {
			return false
		}
	}
	if p.BuildsAccessLevel != nil {
		if !cmp.Equal(string(*p.BuildsAccessLevel), string(g.BuildsAccessLevel)) {
			return false
		}
	}
	if p.WikiAccessLevel != nil {
		if !cmp.Equal(string(*p.WikiAccessLevel), string(g.WikiAccessLevel)) {
			return false
		}
	}
	if p.SnippetsAccessLevel != nil {
		if !cmp.Equal(string(*p.SnippetsAccessLevel), string(g.SnippetsAccessLevel)) {
			return false
		}
	}
	if p.PagesAccessLevel != nil {
		if !cmp.Equal(string(*p.PagesAccessLevel), string(g.PagesAccessLevel)) {
			return false
		}
	}
	if !clients.IsBoolEqualToBoolPtr(p.ResolveOutdatedDiffDiscussions, g.ResolveOutdatedDiffDiscussions) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.ContainerRegistryEnabled, g.ContainerRegistryEnabled) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.SharedRunnersEnabled, g.SharedRunnersEnabled) {
		return false
	}

	if p.Visibility != nil {
		if !cmp.Equal(string(*p.Visibility), string(g.Visibility)) {
			return false
		}
	}
	if !clients.IsBoolEqualToBoolPtr(p.PublicBuilds, g.PublicBuilds) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.OnlyAllowMergeIfPipelineSucceeds, g.OnlyAllowMergeIfPipelineSucceeds) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.OnlyAllowMergeIfAllDiscussionsAreResolved, g.OnlyAllowMergeIfAllDiscussionsAreResolved) {
		return false
	}
	if p.MergeMethod != nil {
		if !cmp.Equal(string(*p.MergeMethod), string(g.MergeMethod)) {
			return false
		}
	}
	if !clients.IsBoolEqualToBoolPtr(p.RemoveSourceBranchAfterMerge, g.RemoveSourceBranchAfterMerge) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.LFSEnabled, g.LFSEnabled) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.RequestAccessEnabled, g.RequestAccessEnabled) {
		return false
	}
	if !cmp.Equal(p.TagList, g.TagList, cmpopts.EquateEmpty()) {
		return false
	}
	if p.CIConfigPath != nil {
		if !cmp.Equal(*p.CIConfigPath, g.CIConfigPath) {
			return false
		}
	}
	if !clients.IsIntEqualToIntPtr(p.ApprovalsBeforeMerge, g.ApprovalsBeforeMerge) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.Mirror, g.Mirror) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.MirrorTriggerBuilds, g.MirrorTriggerBuilds) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.PackagesEnabled, g.PackagesEnabled) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.ServiceDeskEnabled, g.ServiceDeskEnabled) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.AutocloseReferencedIssues, g.AutocloseReferencedIssues) {
		return false
	}

	return true
}
