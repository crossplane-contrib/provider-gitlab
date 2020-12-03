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
<<<<<<< Updated upstream
=======
	"fmt"
	"net/http"
	"strings"
	"time"

>>>>>>> Stashed changes
	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	gitlab "github.com/xanzy/go-gitlab"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
<<<<<<< Updated upstream
	"time"
=======
>>>>>>> Stashed changes
)

// Client defines Gitlab Project service operations
type Client interface {
	GetProject(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error)
	CreateProject(opt *CreateProjectOptions, options ...RequestOptionFunc) (*Project, *Response, error)
}

// NewProjectClient returns a new Gitlab Project service
func NewProjectClient(cfg clients.Config) Client {
	git := clients.NewClient(cfg)
	return git.Projects
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
	in.Visibility = clients.LateInitializeVisibilityValue(in.Visibility, project.Visibility)
	in.MergeMethod = clients.LateInitializeMergeMethodValue(in.MergeMethod, project.MergeMethod)
	if len(in.TagList) == 0 && len(project.TagList) > 0 {
		in.TagList = project.TagList
	}
	in.CIConfigPath = clients.LateInitializeStringPtr(in.CIConfigPath, project.CIConfigPath)
}

// GenerateObservation is used to produce v1alpha1.ProjectObservation from
// gitlab.Project.
func GenerateObservation(prj *gitlab.Project) v1alpha1.ProjectObservation { // nolint:gocyclo
	if prj == nil {
		return v1alpha1.ProjectObservation{}
	}

	o := v1alpha1.ProjectObservation{
		ID:                               prj.ID,
		Public:                           prj.Public,
		SSHURLToRepo:                     prj.SSHURLToRepo,
		HTTPURLToRepo:                    prj.HTTPURLToRepo,
		WebURL:                           prj.WebURL,
		ReadmeURL:                        prj.ReadmeURL,
		PathWithNamespace:                prj.PathWithNamespace,
		IssuesEnabled:                    prj.IssuesEnabled,
		OpenIssuesCount:                  prj.OpenIssuesCount,
		MergeRequestsEnabled:             prj.MergeRequestsEnabled,
		JobsEnabled:                      prj.JobsEnabled,
		WikiEnabled:                      prj.WikiEnabled,
		SnippetsEnabled:                  prj.SnippetsEnabled,
		CreatorID:                        prj.CreatorID,
		ImportStatus:                     prj.ImportStatus,
		ImportError:                      prj.ImportError,
		Archived:                         prj.Archived,
		ForksCount:                       prj.ForksCount,
		StarCount:                        prj.StarCount,
		MirrorUserID:                     prj.MirrorUserID,
		OnlyMirrorProtectedBranches:      prj.OnlyMirrorProtectedBranches,
		MirrorOverwritesDivergedBranches: prj.MirrorOverwritesDivergedBranches,
		CIDefaultGitDepth:                prj.CIDefaultGitDepth,
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
			o.Owner.LastSignInAt = &metav1.Time{Time: time.Time(*prj.Owner.LastSignInAt)}
		}
		if prj.Owner.ConfirmedAt != nil {
			o.Owner.ConfirmedAt = &metav1.Time{Time: time.Time(*prj.Owner.ConfirmedAt)}
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

func stringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func isBoolEqualToBoolPtr(bp *bool, b bool) bool {
	if bp != nil {
		if !cmp.Equal(*bp, b) {
			return false
		}
	}
	return true
}

func isIntEqualToIntPtr(ip *int, i int) bool {
	if ip != nil {
		if !cmp.Equal(*ip, i) {
			return false
		}
	}
	return true
}

// IsProjectUpToDate checks whether there is a change in any of the modifiable fields.
func IsProjectUpToDate(p *v1alpha1.ProjectParameters, g *gitlab.Project) bool { // nolint:gocyclo
	if !cmp.Equal(p.Path, stringToPtr(g.Path)) {
		return false
	}
	if !cmp.Equal(p.DefaultBranch, stringToPtr(g.DefaultBranch)) {
		return false
	}
	if !cmp.Equal(p.Description, stringToPtr(g.Description)) {
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
	if !isBoolEqualToBoolPtr(p.ResolveOutdatedDiffDiscussions, g.ResolveOutdatedDiffDiscussions) {
		return false
	}
	if !isBoolEqualToBoolPtr(p.ContainerRegistryEnabled, g.ContainerRegistryEnabled) {
		return false
	}
	if !isBoolEqualToBoolPtr(p.SharedRunnersEnabled, g.SharedRunnersEnabled) {
		return false
	}

	if p.Visibility != nil {
		if !cmp.Equal(string(*p.Visibility), string(g.Visibility)) {
			return false
		}
	}
	if !isBoolEqualToBoolPtr(p.PublicBuilds, g.PublicBuilds) {
		return false
	}
	if !isBoolEqualToBoolPtr(p.OnlyAllowMergeIfPipelineSucceeds, g.OnlyAllowMergeIfPipelineSucceeds) {
		return false
	}
	if !isBoolEqualToBoolPtr(p.OnlyAllowMergeIfAllDiscussionsAreResolved, g.OnlyAllowMergeIfAllDiscussionsAreResolved) {
		return false
	}
	if p.MergeMethod != nil {
		if !cmp.Equal(string(*p.MergeMethod), string(g.MergeMethod)) {
			return false
		}
	}
	if !isBoolEqualToBoolPtr(p.RemoveSourceBranchAfterMerge, g.RemoveSourceBranchAfterMerge) {
		return false
	}
	if !isBoolEqualToBoolPtr(p.LFSEnabled, g.LFSEnabled) {
		return false
	}
	if !isBoolEqualToBoolPtr(p.RequestAccessEnabled, g.RequestAccessEnabled) {
		return false
	}
	if !cmp.Equal(p.TagList, g.TagList, cmpopts.EquateEmpty()) {
		return false
	}
	if p.CIConfigPath != nil {
		if !cmp.Equal(string(*p.CIConfigPath), string(g.CIConfigPath)) {
			return false
		}
	}
	if !isIntEqualToIntPtr(p.ApprovalsBeforeMerge, g.ApprovalsBeforeMerge) {
		return false
	}
	if !isBoolEqualToBoolPtr(p.Mirror, g.Mirror) {
		return false
	}
	if !isBoolEqualToBoolPtr(p.MirrorTriggerBuilds, g.MirrorTriggerBuilds) {
		return false
	}
	if !isBoolEqualToBoolPtr(p.PackagesEnabled, g.PackagesEnabled) {
		return false
	}
	if !isBoolEqualToBoolPtr(p.ServiceDeskEnabled, g.ServiceDeskEnabled) {
		return false
	}
	if !isBoolEqualToBoolPtr(p.AutocloseReferencedIssues, g.AutocloseReferencedIssues) {
		return false
	}

	return true
}
