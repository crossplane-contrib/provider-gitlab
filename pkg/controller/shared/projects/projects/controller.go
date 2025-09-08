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
	"context"
	"strconv"
	"strings"

	"github.com/crossplane/crossplane-runtime/v2/apis/common"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiCluster "github.com/crossplane-contrib/provider-gitlab/apis/cluster/projects/v1alpha1"
	apiNamespaced "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	sharedProjectsV1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/shared/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
)

const (
	ErrNotProject              = "managed resource is not a Gitlab project custom resource"
	ErrKubeUpdateFailed        = "cannot update Gitlab project custom resource"
	ErrCreateFailed            = "cannot create Gitlab project"
	ErrUpdateFailed            = "cannot update Gitlab project"
	ErrUpdatePushRulesFailed   = "cannot update Gitlab project push rules"
	ErrDeleteFailed            = "cannot delete Gitlab project"
	ErrGetFailed               = "cannot retrieve Gitlab project with"
	ErrGetPushRulesFailed      = "cannot retrieve Gitlab project push rules"
	ErrLateInitialize          = "cannot late-initialize Gitlab project"
	ErrLateInitializePushRules = "cannot late-initialize Gitlab project push rules"
	ErrCheckPushRulesUpToDate  = "cannot compare project push rules"
)

type External struct {
	Client projects.Client
	Kube   client.Client

	cache struct {
		externalPushRules   *sharedProjectsV1alpha1.PushRules
		isPushRulesUpToDate bool
	}
}

type options struct {
	externalName  string
	name          string
	parameters    *sharedProjectsV1alpha1.ProjectParameters
	atProvider    *sharedProjectsV1alpha1.ProjectObservation
	setConditions func(c ...common.Condition)
	mg            resource.Managed
}

func (e *External) extractOptions(mg resource.Managed) (*options, error) {
	switch cr := mg.(type) {
	case *apiCluster.Project:
		return &options{
			externalName:  meta.GetExternalName(cr),
			name:          cr.Name,
			parameters:    &cr.Spec.ForProvider.ProjectParameters,
			atProvider:    &cr.Status.AtProvider,
			setConditions: cr.SetConditions,
			mg:            mg,
		}, nil
	case *apiNamespaced.Project:
		return &options{
			externalName:  meta.GetExternalName(cr),
			name:          cr.Name,
			parameters:    &cr.Spec.ForProvider.ProjectParameters,
			atProvider:    &cr.Status.AtProvider,
			setConditions: cr.SetConditions,
			mg:            mg,
		}, nil
	default:
		return nil, errors.New(ErrNotProject)
	}
}

func (e *External) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { //nolint:gocyclo
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if opts.externalName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	projectID, err := strconv.Atoi(opts.externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.New(ErrNotProject)
	}

	prj, res, err := e.Client.GetProject(projectID, nil)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, ErrGetFailed)
	}

	// Check if the project is in a pending deletion state and either remove the
	// finalizer if specified or keep tracking it.
	//
	// Mark the resource as unavailable if the project is in a deletion state but
	// managed resource is not.
	if prj.MarkedForDeletionOn != nil {
		if meta.WasDeleted(opts.mg) {
			if ptr.Deref(opts.parameters.RemoveFinalizerOnPendingDeletion, false) {
				return managed.ExternalObservation{}, nil
			}
			opts.setConditions(xpv1.Deleting().WithMessage("Project is in pending deletion state"))
		} else {
			opts.setConditions(xpv1.Unavailable().WithMessage("Project is in pending deletion state but this managed resource is not"))
		}
	} else {
		opts.setConditions(xpv1.Available())
	}

	current := opts.parameters.DeepCopy()
	if err := e.lateInitialize(ctx, opts, prj); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, ErrLateInitialize)
	}

	e.cache.isPushRulesUpToDate, err = e.isPushRulesUpToDate(ctx, opts)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, ErrCheckPushRulesUpToDate)
	}

	*opts.atProvider = projects.GenerateObservation(prj)
	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        e.isProjectUpToDate(opts.parameters, prj) && e.cache.isPushRulesUpToDate,
		ResourceLateInitialized: !cmp.Equal(current, opts.parameters),
		ConnectionDetails:       managed.ConnectionDetails{"runnersToken": []byte(prj.RunnersToken)},
	}, nil
}

func (e *External) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	prj, _, err := e.Client.CreateProject(
		projects.GenerateCreateProjectOptions(opts.name, opts.parameters),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, ErrCreateFailed)
	}

	meta.SetExternalName(opts.mg, strconv.Itoa(prj.ID))
	return managed.ExternalCreation{}, nil
}

func (e *External) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	_, _, err = e.Client.EditProject(
		opts.externalName,
		projects.GenerateEditProjectOptions(opts.name, opts.parameters),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, ErrUpdateFailed)
	}

	if !e.cache.isPushRulesUpToDate {
		// Only attempt to update push rules if the feature is supported
		// (either we have cached rules or push rules are specified in spec)
		if e.cache.externalPushRules != nil || opts.parameters.PushRules != nil {
			_, _, err := e.Client.EditProjectPushRule(
				opts.externalName,
				projects.GenerateEditPushRulesOptions(opts.parameters),
				gitlab.WithContext(ctx),
			)
			if err != nil {
				return managed.ExternalUpdate{}, errors.Wrap(err, ErrUpdatePushRulesFailed)
			}
		}
		// If push rules are not supported (e.g., GitLab Community Edition) and
		// none are specified in spec, we skip updating them
	}
	return managed.ExternalUpdate{}, nil
}

func (e *External) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalDelete{}, err
	}

	_, err = e.Client.DeleteProject(opts.externalName, &gitlab.DeleteProjectOptions{}, gitlab.WithContext(ctx))
	// if the project is for some reason already marked for deletion, we ignore the error and continue to delete the project permanently
	if err != nil && !strings.Contains(err.Error(), "Deletion pending.") {
		return managed.ExternalDelete{}, errors.Wrap(err, ErrDeleteFailed)
	}

	if opts.parameters.PermanentlyRemove != nil && *opts.parameters.PermanentlyRemove {
		_, err = e.Client.DeleteProject(opts.externalName, &gitlab.DeleteProjectOptions{
			PermanentlyRemove: opts.parameters.PermanentlyRemove,
			FullPath:          &opts.atProvider.PathWithNamespace,
		}, gitlab.WithContext(ctx))
	}
	return managed.ExternalDelete{}, errors.Wrap(err, ErrDeleteFailed)
}

// lateInitialize fills the empty fields in the project spec with the
// values seen in gitlab.Project.
func (e *External) lateInitialize(ctx context.Context, opts *options, project *gitlab.Project) error { //nolint:gocyclo
	in := opts.parameters

	if project == nil {
		return nil
	}
	if in.AllowMergeOnSkippedPipeline == nil {
		in.AllowMergeOnSkippedPipeline = &project.AllowMergeOnSkippedPipeline
	}
	if in.ApprovalsBeforeMerge == nil {
		in.ApprovalsBeforeMerge = &project.ApprovalsBeforeMerge //nolint:staticcheck
	}
	if in.AutocloseReferencedIssues == nil {
		in.AutocloseReferencedIssues = &project.AutocloseReferencedIssues
	}

	in.BuildCoverageRegex = clients.LateInitializeStringPtr(in.BuildCoverageRegex, project.BuildCoverageRegex)
	in.BuildsAccessLevel = clients.LateInitializeAccessControlValue(in.BuildsAccessLevel, project.BuildsAccessLevel)
	in.CIConfigPath = clients.LateInitializeStringPtr(in.CIConfigPath, project.CIConfigPath)

	if in.CIDefaultGitDepth == nil {
		in.CIDefaultGitDepth = &project.CIDefaultGitDepth
	}
	if in.CIForwardDeploymentEnabled == nil {
		in.CIForwardDeploymentEnabled = &project.CIForwardDeploymentEnabled
	}
	if in.ContainerRegistryEnabled == nil { //nolint:staticcheck
		in.ContainerRegistryEnabled = &project.ContainerRegistryEnabled //nolint:staticcheck
	}
	if in.ContainerRegistryAccessLevel == nil {
		in.ContainerRegistryAccessLevel = clients.LateInitializeAccessControlValue(in.ContainerRegistryAccessLevel, project.ContainerRegistryAccessLevel)
	}

	in.DefaultBranch = clients.LateInitializeStringPtr(in.DefaultBranch, project.DefaultBranch)
	in.Description = clients.LateInitializeStringPtr(in.Description, project.Description)
	in.ForkingAccessLevel = clients.LateInitializeAccessControlValue(in.ForkingAccessLevel, project.ForkingAccessLevel)
	in.IssuesAccessLevel = clients.LateInitializeAccessControlValue(in.IssuesAccessLevel, project.IssuesAccessLevel)
	in.IssuesTemplate = clients.LateInitializeStringPtr(in.IssuesTemplate, project.IssuesTemplate)

	if in.LFSEnabled == nil {
		in.LFSEnabled = &project.LFSEnabled
	}

	in.MergeMethod = clients.LateInitializeMergeMethodValue(in.MergeMethod, project.MergeMethod)
	in.MergeRequestsAccessLevel = clients.LateInitializeAccessControlValue(in.MergeRequestsAccessLevel, project.MergeRequestsAccessLevel)
	in.MergeRequestsTemplate = clients.LateInitializeStringPtr(in.MergeRequestsTemplate, project.MergeRequestsTemplate)

	if in.Mirror == nil {
		in.Mirror = &project.Mirror
	}
	if in.MirrorOverwritesDivergedBranches == nil {
		in.MirrorOverwritesDivergedBranches = &project.MirrorOverwritesDivergedBranches
	}
	if in.MirrorTriggerBuilds == nil {
		in.MirrorTriggerBuilds = &project.MirrorTriggerBuilds
	}
	if in.MirrorUserID == nil && project.MirrorUserID != 0 { // since project.MirrorUserID is non-nullable, value `0` treated as `not set`
		in.MirrorUserID = &project.MirrorUserID
	}
	if in.OnlyAllowMergeIfAllDiscussionsAreResolved == nil {
		in.OnlyAllowMergeIfAllDiscussionsAreResolved = &project.OnlyAllowMergeIfAllDiscussionsAreResolved
	}
	if in.OnlyAllowMergeIfPipelineSucceeds == nil {
		in.OnlyAllowMergeIfPipelineSucceeds = &project.OnlyAllowMergeIfPipelineSucceeds
	}
	if in.OnlyMirrorProtectedBranches == nil {
		in.OnlyMirrorProtectedBranches = &project.OnlyMirrorProtectedBranches
	}

	in.OperationsAccessLevel = clients.LateInitializeAccessControlValue(in.OperationsAccessLevel, project.OperationsAccessLevel)

	if in.PackagesEnabled == nil {
		in.PackagesEnabled = &project.PackagesEnabled
	}

	in.PagesAccessLevel = clients.LateInitializeAccessControlValue(in.PagesAccessLevel, project.PagesAccessLevel)
	in.Path = clients.LateInitializeStringPtr(in.Path, project.Path)

	if in.PublicBuilds == nil {
		in.PublicBuilds = &project.PublicJobs
	}
	if in.RemoveSourceBranchAfterMerge == nil {
		in.RemoveSourceBranchAfterMerge = &project.RemoveSourceBranchAfterMerge
	}

	in.RepositoryAccessLevel = clients.LateInitializeAccessControlValue(in.RepositoryAccessLevel, project.RepositoryAccessLevel)

	if in.RequestAccessEnabled == nil {
		in.RequestAccessEnabled = &project.RequestAccessEnabled
	}
	if in.ResolveOutdatedDiffDiscussions == nil {
		in.ResolveOutdatedDiffDiscussions = &project.ResolveOutdatedDiffDiscussions
	}
	if in.ServiceDeskEnabled == nil {
		in.ServiceDeskEnabled = &project.ServiceDeskEnabled
	}
	if in.SharedRunnersEnabled == nil {
		in.SharedRunnersEnabled = &project.SharedRunnersEnabled
	}

	in.SnippetsAccessLevel = clients.LateInitializeAccessControlValue(in.SnippetsAccessLevel, project.SnippetsAccessLevel)
	in.SuggestionCommitMessage = clients.LateInitializeStringPtr(in.SuggestionCommitMessage, project.SuggestionCommitMessage)

	if len(in.TagList) == 0 && len(project.TagList) > 0 { //nolint:staticcheck
		in.TagList = project.TagList //nolint:staticcheck
	}
	if len(in.Topics) == 0 && len(project.Topics) > 0 {
		in.Topics = project.Topics
	}

	in.Visibility = clients.LateInitializeVisibilityValue(in.Visibility, project.Visibility)
	in.WikiAccessLevel = clients.LateInitializeAccessControlValue(in.WikiAccessLevel, project.WikiAccessLevel)

	if err := e.lateInitializePushRules(ctx, opts); err != nil {
		return errors.Wrap(err, ErrLateInitializePushRules)
	}

	return nil
}

func (e *External) lateInitializePushRules(ctx context.Context, opts *options) error {
	pr, err := e.getProjectPushRules(ctx, opts)
	if err != nil || pr == nil {
		return err
	}

	// Only late-initialize push rules if they are not explicitly set to nil in the spec
	// and if no push rules configuration exists yet
	if opts.parameters.PushRules == nil {
		// Don't late-initialize if user explicitly wants no push rules
		// We can distinguish this by checking if this is a new resource or if push rules were explicitly removed
		// For now, we'll be conservative and not late-initialize push rules to avoid unwanted behavior
		return nil
	}

	opts.parameters.PushRules = &sharedProjectsV1alpha1.PushRules{
		AuthorEmailRegex:           clients.LateInitialize(opts.parameters.PushRules.AuthorEmailRegex, pr.AuthorEmailRegex),
		BranchNameRegex:            clients.LateInitialize(opts.parameters.PushRules.BranchNameRegex, pr.BranchNameRegex),
		CommitCommitterCheck:       clients.LateInitialize(opts.parameters.PushRules.CommitCommitterCheck, pr.CommitCommitterCheck),
		CommitCommitterNameCheck:   clients.LateInitialize(opts.parameters.PushRules.CommitCommitterNameCheck, pr.CommitCommitterNameCheck),
		CommitMessageNegativeRegex: clients.LateInitialize(opts.parameters.PushRules.CommitMessageNegativeRegex, pr.CommitMessageNegativeRegex),
		CommitMessageRegex:         clients.LateInitialize(opts.parameters.PushRules.CommitMessageRegex, pr.CommitMessageRegex),
		DenyDeleteTag:              clients.LateInitialize(opts.parameters.PushRules.DenyDeleteTag, pr.DenyDeleteTag),
		FileNameRegex:              clients.LateInitialize(opts.parameters.PushRules.FileNameRegex, pr.FileNameRegex),
		MaxFileSize:                clients.LateInitialize(opts.parameters.PushRules.MaxFileSize, pr.MaxFileSize),
		MemberCheck:                clients.LateInitialize(opts.parameters.PushRules.MemberCheck, pr.MemberCheck),
		PreventSecrets:             clients.LateInitialize(opts.parameters.PushRules.PreventSecrets, pr.PreventSecrets),
		RejectUnsignedCommits:      clients.LateInitialize(opts.parameters.PushRules.RejectUnsignedCommits, pr.RejectUnsignedCommits),
		RejectNonDCOCommits:        clients.LateInitialize(opts.parameters.PushRules.RejectNonDCOCommits, pr.RejectNonDCOCommits),
	}
	return nil
}

func (e *External) getProjectPushRules(ctx context.Context, opts *options) (*sharedProjectsV1alpha1.PushRules, error) {
	if e.cache.externalPushRules != nil {
		return e.cache.externalPushRules, nil
	}
	res, resp, err := e.Client.GetProjectPushRules(
		opts.externalName,
		gitlab.WithContext(ctx),
	)
	if err != nil {
		// Push rules are not available in GitLab Community Edition (404 Not Found)
		// However, we need to be careful: 404 could mean either:
		// 1. Feature not available (Community Edition) - should be ignored
		// 2. No push rules configured (Premium/Enterprise) - might also return 404 in some GitLab versions
		//
		// Based on GitLab API patterns, when a feature is available but not configured,
		// it typically returns 200 with default/empty values rather than 404.
		// Therefore, 404 most likely means the feature is not available.
		if clients.IsResponseNotFound(resp) {
			return nil, nil
		}
		return nil, errors.Wrap(err, ErrGetPushRulesFailed)
	}
	e.cache.externalPushRules = &sharedProjectsV1alpha1.PushRules{
		AuthorEmailRegex:           &res.AuthorEmailRegex,
		BranchNameRegex:            &res.BranchNameRegex,
		CommitCommitterCheck:       &res.CommitCommitterCheck,
		CommitCommitterNameCheck:   &res.CommitCommitterNameCheck,
		CommitMessageNegativeRegex: &res.CommitMessageNegativeRegex,
		CommitMessageRegex:         &res.CommitMessageRegex,
		DenyDeleteTag:              &res.DenyDeleteTag,
		FileNameRegex:              &res.FileNameRegex,
		MaxFileSize:                &res.MaxFileSize,
		MemberCheck:                &res.MemberCheck,
		PreventSecrets:             &res.PreventSecrets,
		RejectUnsignedCommits:      &res.RejectUnsignedCommits,
		RejectNonDCOCommits:        &res.RejectNonDCOCommits,
	}
	return e.cache.externalPushRules, nil
}

// isProjectUpToDate checks whether there is a change in any of the modifiable fields.
func (e *External) isProjectUpToDate(p *sharedProjectsV1alpha1.ProjectParameters, g *gitlab.Project) bool { //nolint:gocyclo
	if p.Name != nil && !cmp.Equal(*p.Name, g.Name) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.AllowMergeOnSkippedPipeline, g.AllowMergeOnSkippedPipeline) {
		return false
	}
	if !clients.IsIntEqualToIntPtr(p.ApprovalsBeforeMerge, g.ApprovalsBeforeMerge) { //nolint:staticcheck
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.AutocloseReferencedIssues, g.AutocloseReferencedIssues) {
		return false
	}
	if !cmp.Equal(p.BuildCoverageRegex, clients.StringToPtr(g.BuildCoverageRegex)) {
		return false
	}
	if p.BuildsAccessLevel != nil && !cmp.Equal(string(*p.BuildsAccessLevel), string(g.BuildsAccessLevel)) {
		return false
	}
	if p.CIConfigPath != nil && !cmp.Equal(*p.CIConfigPath, g.CIConfigPath) {
		return false
	}
	if !clients.IsIntEqualToIntPtr(p.CIDefaultGitDepth, g.CIDefaultGitDepth) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.CIForwardDeploymentEnabled, g.CIForwardDeploymentEnabled) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.ContainerRegistryEnabled, g.ContainerRegistryEnabled) { //nolint:staticcheck
		return false
	}
	if p.ContainerRegistryAccessLevel != nil && !cmp.Equal(string(*p.ContainerRegistryAccessLevel), string(g.ContainerRegistryAccessLevel)) {
		return false
	}
	if !cmp.Equal(p.DefaultBranch, clients.StringToPtr(g.DefaultBranch)) {
		return false
	}
	if !cmp.Equal(p.Description, clients.StringToPtr(g.Description)) {
		return false
	}
	if p.ForkingAccessLevel != nil && !cmp.Equal(string(*p.ForkingAccessLevel), string(g.ForkingAccessLevel)) {
		return false
	}
	if p.IssuesAccessLevel != nil && !cmp.Equal(string(*p.IssuesAccessLevel), string(g.IssuesAccessLevel)) {
		return false
	}
	if !cmp.Equal(p.IssuesTemplate, clients.StringToPtr(g.IssuesTemplate)) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.LFSEnabled, g.LFSEnabled) {
		return false
	}
	if p.MergeMethod != nil && !cmp.Equal(string(*p.MergeMethod), string(g.MergeMethod)) {
		return false
	}
	if p.MergeRequestsAccessLevel != nil && !cmp.Equal(string(*p.MergeRequestsAccessLevel), string(g.MergeRequestsAccessLevel)) {
		return false
	}
	if !cmp.Equal(p.MergeRequestsTemplate, clients.StringToPtr(g.MergeRequestsTemplate)) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.Mirror, g.Mirror) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.MirrorOverwritesDivergedBranches, g.MirrorOverwritesDivergedBranches) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.MirrorTriggerBuilds, g.MirrorTriggerBuilds) {
		return false
	}
	if !clients.IsIntEqualToIntPtr(p.MirrorUserID, g.MirrorUserID) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.OnlyAllowMergeIfAllDiscussionsAreResolved, g.OnlyAllowMergeIfAllDiscussionsAreResolved) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.OnlyAllowMergeIfPipelineSucceeds, g.OnlyAllowMergeIfPipelineSucceeds) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.OnlyMirrorProtectedBranches, g.OnlyMirrorProtectedBranches) {
		return false
	}
	if p.OperationsAccessLevel != nil && !cmp.Equal(string(*p.OperationsAccessLevel), string(g.OperationsAccessLevel)) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.PackagesEnabled, g.PackagesEnabled) {
		return false
	}
	if p.PagesAccessLevel != nil && !cmp.Equal(string(*p.PagesAccessLevel), string(g.PagesAccessLevel)) {
		return false
	}
	if !cmp.Equal(p.Path, clients.StringToPtr(g.Path)) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.PublicBuilds, g.PublicJobs) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.RemoveSourceBranchAfterMerge, g.RemoveSourceBranchAfterMerge) {
		return false
	}
	if p.RepositoryAccessLevel != nil && !cmp.Equal(string(*p.RepositoryAccessLevel), string(g.RepositoryAccessLevel)) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.RequestAccessEnabled, g.RequestAccessEnabled) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.ResolveOutdatedDiffDiscussions, g.ResolveOutdatedDiffDiscussions) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.ServiceDeskEnabled, g.ServiceDeskEnabled) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.SharedRunnersEnabled, g.SharedRunnersEnabled) {
		return false
	}
	if p.SnippetsAccessLevel != nil && !cmp.Equal(string(*p.SnippetsAccessLevel), string(g.SnippetsAccessLevel)) {
		return false
	}
	if !cmp.Equal(p.SuggestionCommitMessage, clients.StringToPtr(g.SuggestionCommitMessage)) {
		return false
	}
	if !cmp.Equal(p.TagList, g.TagList, cmpopts.EquateEmpty()) { //nolint:staticcheck
		return false
	}
	if !cmp.Equal(p.Topics, g.Topics, cmpopts.EquateEmpty()) {
		return false
	}
	if p.Visibility != nil && !cmp.Equal(string(*p.Visibility), string(g.Visibility)) {
		return false
	}
	if p.WikiAccessLevel != nil && !cmp.Equal(string(*p.WikiAccessLevel), string(g.WikiAccessLevel)) {
		return false
	}
	return true
}

func (e *External) isPushRulesEmpty(rules *sharedProjectsV1alpha1.PushRules) bool { //nolint:gocyclo
	if rules == nil {
		return true
	}

	// Check string fields - if any is non-empty, rules are not empty
	if rules.AuthorEmailRegex != nil && *rules.AuthorEmailRegex != "" {
		return false
	}
	if rules.BranchNameRegex != nil && *rules.BranchNameRegex != "" {
		return false
	}
	if rules.CommitMessageNegativeRegex != nil && *rules.CommitMessageNegativeRegex != "" {
		return false
	}
	if rules.CommitMessageRegex != nil && *rules.CommitMessageRegex != "" {
		return false
	}
	if rules.FileNameRegex != nil && *rules.FileNameRegex != "" {
		return false
	}

	// Check boolean fields - if any is true, rules are not empty
	if rules.CommitCommitterCheck != nil && *rules.CommitCommitterCheck {
		return false
	}
	if rules.CommitCommitterNameCheck != nil && *rules.CommitCommitterNameCheck {
		return false
	}
	if rules.DenyDeleteTag != nil && *rules.DenyDeleteTag {
		return false
	}
	if rules.MemberCheck != nil && *rules.MemberCheck {
		return false
	}
	if rules.PreventSecrets != nil && *rules.PreventSecrets {
		return false
	}
	if rules.RejectNonDCOCommits != nil && *rules.RejectNonDCOCommits {
		return false
	}
	if rules.RejectUnsignedCommits != nil && *rules.RejectUnsignedCommits {
		return false
	}

	// Check integer fields - if any is non-zero, rules are not empty
	if rules.MaxFileSize != nil && *rules.MaxFileSize != 0 {
		return false
	}

	// If we reach here, all fields are empty/default
	return true
}

func (e *External) isPushRulesUpToDate(ctx context.Context, opts *options) (bool, error) {
	current, err := e.getProjectPushRules(ctx, opts)
	if err != nil {
		return false, err
	}

	// If push rules are not available (e.g., GitLab Community Edition),
	// consider them up to date if no push rules are specified in the spec
	if current == nil {
		return opts.parameters.PushRules == nil, nil
	}

	// Check if current push rules are effectively empty (all default values)
	isCurrentEmpty := e.isPushRulesEmpty(current)

	// If current rules are empty and spec has no rules, they're up to date
	if isCurrentEmpty && opts.parameters.PushRules == nil {
		return true, nil
	}

	// If push rules are available in GitLab but not specified in spec,
	// they need to be cleared (not up to date) - unless they're already effectively empty
	if opts.parameters.PushRules == nil {
		return isCurrentEmpty, nil
	}

	// Both exist, compare them
	return cmp.Equal(opts.parameters.PushRules, current), nil
}

func (e *External) Disconnect(ctx context.Context) error {
	return nil
}
