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

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/statemetrics"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects"
)

const (
	errNotProject              = "managed resource is not a Gitlab project custom resource"
	errKubeUpdateFailed        = "cannot update Gitlab project custom resource"
	errCreateFailed            = "cannot create Gitlab project"
	errUpdateFailed            = "cannot update Gitlab project"
	errUpdatePushRulesFailed   = "cannot update Gitlab project push rules"
	errDeleteFailed            = "cannot delete Gitlab project"
	errGetFailed               = "cannot retrieve Gitlab project with"
	errGetPushRulesFailed      = "cannot retrieve Gitlab project push rules"
	errLateInitialize          = "cannot late-initialize Gitlab project"
	errLateInitializePushRules = "cannot late-initialize Gitlab project push rules"
	errCheckPushRulesUpToDate  = "cannot compare project push rules"
)

// SetupProject adds a controller that reconciles Projects.
func SetupProject(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.ProjectGroupKind)

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabClientFn: projects.NewProjectClient}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}

	if o.Features.Enabled(feature.EnableBetaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.ProjectGroupVersionKind),
		reconcilerOpts...)

	if err := mgr.Add(statemetrics.NewMRStateRecorder(
		mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &v1alpha1.ProjectList{}, o.MetricOptions.PollStateMetricInterval)); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Project{}).
		Complete(r)
}

// SetupProjectGated adds a controller with CRD gate support.
func SetupProjectGated(mgr ctrl.Manager, o controller.Options) error {
	o.Gate.Register(func() {
		if err := SetupProject(mgr, o); err != nil {
			mgr.GetLogger().Error(err, "unable to setup reconciler", "gvk", v1alpha1.ProjectGroupVersionKind.String())
		}
	}, v1alpha1.ProjectGroupVersionKind)
	return nil
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg common.Config) projects.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Project)
	if !ok {
		return nil, errors.New(errNotProject)
	}
	cfg, err := common.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

type external struct {
	kube   client.Client
	client projects.Client

	cache struct {
		externalPushRules   *v1alpha1.PushRules
		isPushRulesUpToDate bool
	}
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { //nolint:gocyclo
	cr, ok := mg.(*v1alpha1.Project)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotProject)
	}

	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	projectID, err := strconv.Atoi(externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.New(errNotProject)
	}

	prj, res, err := e.client.GetProject(projectID, nil)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}

	// Check if the project is in a pending deletion state and either remove the
	// finalizer if specified or keep tracking it.
	//
	// Mark the resource as unavailable if the project is in a deletion state but
	// managed resource is not.
	if prj.MarkedForDeletionOn != nil {
		if meta.WasDeleted(cr) {
			if ptr.Deref(cr.Spec.ForProvider.RemoveFinalizerOnPendingDeletion, false) {
				return managed.ExternalObservation{}, nil
			}
			cr.SetConditions(xpv1.Deleting().WithMessage("Project is in pending deletion state"))
		} else {
			cr.SetConditions(xpv1.Unavailable().WithMessage("Project is in pending deletion state but this managed resource is not"))
		}
	} else {
		cr.Status.SetConditions(xpv1.Available())
	}

	current := cr.Spec.ForProvider.DeepCopy()
	if err := e.lateInitialize(ctx, cr, prj); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errLateInitialize)
	}

	e.cache.isPushRulesUpToDate, err = e.isPushRulesUpToDate(ctx, cr)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCheckPushRulesUpToDate)
	}

	cr.Status.AtProvider = projects.GenerateObservation(prj)
	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isProjectUpToDate(&cr.Spec.ForProvider, prj) && e.cache.isPushRulesUpToDate,
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
		ConnectionDetails:       managed.ConnectionDetails{"runnersToken": []byte(prj.RunnersToken)},
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Project)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotProject)
	}

	prj, _, err := e.client.CreateProject(
		projects.GenerateCreateProjectOptions(cr.Name, &cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

<<<<<<< HEAD
	meta.SetExternalName(cr, strconv.FormatInt(prj.ID, 10))
=======
	meta.SetExternalName(cr, strconv.FormatInt(int64(prj.ID), 10))
>>>>>>> 77c306d (feat: migrate CRD types from *int to *int64)
	return managed.ExternalCreation{}, errors.Wrap(err, errKubeUpdateFailed)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Project)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotProject)
	}

	_, _, err := e.client.EditProject(
		meta.GetExternalName(cr),
		projects.GenerateEditProjectOptions(cr.Name, &cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}

	if !e.cache.isPushRulesUpToDate {
		// Only attempt to update push rules if the feature is supported
		// (either we have cached rules or push rules are specified in spec)
		if e.cache.externalPushRules != nil || cr.Spec.ForProvider.PushRules != nil {
			_, _, err := e.client.EditProjectPushRule(
				meta.GetExternalName(cr),
				projects.GenerateEditPushRulesOptions(&cr.Spec.ForProvider),
				gitlab.WithContext(ctx),
			)
			if err != nil {
				return managed.ExternalUpdate{}, errors.Wrap(err, errUpdatePushRulesFailed)
			}
		}
		// If push rules are not supported (e.g., GitLab Community Edition) and
		// none are specified in spec, we skip updating them
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.Project)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotProject)
	}

	_, err := e.client.DeleteProject(meta.GetExternalName(cr), &gitlab.DeleteProjectOptions{}, gitlab.WithContext(ctx))
	// if the project is for some reason already marked for deletion, we ignore the error and continue to delete the project permanently
	if err != nil && !strings.Contains(err.Error(), "Deletion pending.") {
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteFailed)
	}

	if cr.Spec.ForProvider.PermanentlyRemove != nil && *cr.Spec.ForProvider.PermanentlyRemove {
		_, err = e.client.DeleteProject(meta.GetExternalName(cr), &gitlab.DeleteProjectOptions{
			PermanentlyRemove: cr.Spec.ForProvider.PermanentlyRemove,
			FullPath:          &cr.Status.AtProvider.PathWithNamespace,
		}, gitlab.WithContext(ctx))
	}
	return managed.ExternalDelete{}, errors.Wrap(err, errDeleteFailed)
}

func (e *external) Disconnect(ctx context.Context) error {
	// Disconnect is not implemented as it is a new method required by the SDK
	return nil
}

// lateInitialize fills the empty fields in the project spec with the
// values seen in gitlab.Project.
func (e *external) lateInitialize(ctx context.Context, cr *v1alpha1.Project, project *gitlab.Project) error { //nolint:gocyclo
	in := &cr.Spec.ForProvider

	if project == nil {
		return nil
	}
	if in.AllowMergeOnSkippedPipeline == nil {
		in.AllowMergeOnSkippedPipeline = &project.AllowMergeOnSkippedPipeline
	}
<<<<<<< HEAD
	//nolint:staticcheck // SA1019 ApprovalsBeforeMerge is deprecated by GitLab API, will migrate to Merge Request Approvals API later
	if in.ApprovalsBeforeMerge == nil && project.ApprovalsBeforeMerge != 0 {
		val := project.ApprovalsBeforeMerge
		in.ApprovalsBeforeMerge = &val
=======
	if in.ApprovalsBeforeMerge == nil && project.ApprovalsBeforeMerge != 0 { //nolint:staticcheck
		val := project.ApprovalsBeforeMerge
		in.ApprovalsBeforeMerge = &val //nolint:staticcheck
>>>>>>> 77c306d (feat: migrate CRD types from *int to *int64)
	}
	if in.AutocloseReferencedIssues == nil {
		in.AutocloseReferencedIssues = &project.AutocloseReferencedIssues
	}

	in.BuildCoverageRegex = clients.LateInitializeStringPtr(in.BuildCoverageRegex, project.BuildCoverageRegex)
	in.BuildsAccessLevel = clients.LateInitializeAccessControlValue(in.BuildsAccessLevel, project.BuildsAccessLevel)
	in.CIConfigPath = clients.LateInitializeStringPtr(in.CIConfigPath, project.CIConfigPath)

	if in.CIDefaultGitDepth == nil && project.CIDefaultGitDepth != 0 {
		val := project.CIDefaultGitDepth
		in.CIDefaultGitDepth = &val
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
		val := project.MirrorUserID
		in.MirrorUserID = &val
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

	// Late-initialize publicJobs and publicBuilds for backward compatibility
	// Only initialize if both fields are unset
	//nolint:staticcheck // We intentionally use the deprecated field for backward compatibility
	if in.PublicBuilds == nil && in.PublicJobs == nil {
		in.PublicJobs = &project.PublicJobs
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

	if err := e.lateInitializePushRules(ctx, cr); err != nil {
		return errors.Wrap(err, errLateInitializePushRules)
	}

	return nil
}

func (e *external) lateInitializePushRules(ctx context.Context, cr *v1alpha1.Project) error {
	pr, err := e.getProjectPushRules(ctx, cr)
	if err != nil || pr == nil {
		return err
	}

	// Only late-initialize push rules if they are not explicitly set to nil in the spec
	// and if no push rules configuration exists yet
	if cr.Spec.ForProvider.PushRules == nil {
		// Don't late-initialize if user explicitly wants no push rules
		// We can distinguish this by checking if this is a new resource or if push rules were explicitly removed
		// For now, we'll be conservative and not late-initialize push rules to avoid unwanted behavior
		return nil
	}

	cr.Spec.ForProvider.PushRules = &v1alpha1.PushRules{
		AuthorEmailRegex:           clients.LateInitialize(cr.Spec.ForProvider.PushRules.AuthorEmailRegex, pr.AuthorEmailRegex),
		BranchNameRegex:            clients.LateInitialize(cr.Spec.ForProvider.PushRules.BranchNameRegex, pr.BranchNameRegex),
		CommitCommitterCheck:       clients.LateInitialize(cr.Spec.ForProvider.PushRules.CommitCommitterCheck, pr.CommitCommitterCheck),
		CommitCommitterNameCheck:   clients.LateInitialize(cr.Spec.ForProvider.PushRules.CommitCommitterNameCheck, pr.CommitCommitterNameCheck),
		CommitMessageNegativeRegex: clients.LateInitialize(cr.Spec.ForProvider.PushRules.CommitMessageNegativeRegex, pr.CommitMessageNegativeRegex),
		CommitMessageRegex:         clients.LateInitialize(cr.Spec.ForProvider.PushRules.CommitMessageRegex, pr.CommitMessageRegex),
		DenyDeleteTag:              clients.LateInitialize(cr.Spec.ForProvider.PushRules.DenyDeleteTag, pr.DenyDeleteTag),
		FileNameRegex:              clients.LateInitialize(cr.Spec.ForProvider.PushRules.FileNameRegex, pr.FileNameRegex),
		MaxFileSize:                clients.LateInitialize(cr.Spec.ForProvider.PushRules.MaxFileSize, pr.MaxFileSize),
		MemberCheck:                clients.LateInitialize(cr.Spec.ForProvider.PushRules.MemberCheck, pr.MemberCheck),
		PreventSecrets:             clients.LateInitialize(cr.Spec.ForProvider.PushRules.PreventSecrets, pr.PreventSecrets),
		RejectUnsignedCommits:      clients.LateInitialize(cr.Spec.ForProvider.PushRules.RejectUnsignedCommits, pr.RejectUnsignedCommits),
		RejectNonDCOCommits:        clients.LateInitialize(cr.Spec.ForProvider.PushRules.RejectNonDCOCommits, pr.RejectNonDCOCommits),
	}
	return nil
}

func (e *external) getProjectPushRules(ctx context.Context, cr *v1alpha1.Project) (*v1alpha1.PushRules, error) {
	if e.cache.externalPushRules != nil {
		return e.cache.externalPushRules, nil
	}
	res, resp, err := e.client.GetProjectPushRules(
		meta.GetExternalName(cr),
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
		return nil, errors.Wrap(err, errGetPushRulesFailed)
	}
	e.cache.externalPushRules = &v1alpha1.PushRules{
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
func isProjectUpToDate(p *v1alpha1.ProjectParameters, g *gitlab.Project) bool { //nolint:gocyclo
	if p.Name != nil && !cmp.Equal(*p.Name, g.Name) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.AllowMergeOnSkippedPipeline, g.AllowMergeOnSkippedPipeline) {
		return false
	}
	if !clients.IsInt64EqualToInt64Ptr(p.ApprovalsBeforeMerge, g.ApprovalsBeforeMerge) { //nolint:staticcheck
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
	if !clients.IsInt64EqualToInt64Ptr(p.CIDefaultGitDepth, g.CIDefaultGitDepth) {
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
	if !clients.IsInt64EqualToInt64Ptr(p.MirrorUserID, g.MirrorUserID) {
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
	// Use the resolved publicJobs value for comparison
	//nolint:staticcheck // We intentionally use the deprecated field for backward compatibility
	effectiveValue, _ := common.ResolvePublicJobsSetting(p.PublicBuilds, p.PublicJobs)
	if !clients.IsBoolEqualToBoolPtr(effectiveValue, g.PublicJobs) {
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

func isPushRulesEmpty(rules *v1alpha1.PushRules) bool { //nolint:gocyclo
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

func (e *external) isPushRulesUpToDate(ctx context.Context, cr *v1alpha1.Project) (bool, error) {
	current, err := e.getProjectPushRules(ctx, cr)
	if err != nil {
		return false, err
	}

	// If push rules are not available (e.g., GitLab Community Edition),
	// consider them up to date if no push rules are specified in the spec
	if current == nil {
		return cr.Spec.ForProvider.PushRules == nil, nil
	}

	// Check if current push rules are effectively empty (all default values)
	isCurrentEmpty := isPushRulesEmpty(current)

	// If current rules are empty and spec has no rules, they're up to date
	if isCurrentEmpty && cr.Spec.ForProvider.PushRules == nil {
		return true, nil
	}

	// If push rules are available in GitLab but not specified in spec,
	// they need to be cleared (not up to date) - unless they're already effectively empty
	if cr.Spec.ForProvider.PushRules == nil {
		return isCurrentEmpty, nil
	}

	// Both exist, compare them
	return cmp.Equal(cr.Spec.ForProvider.PushRules, current), nil
}
