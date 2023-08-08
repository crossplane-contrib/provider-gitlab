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

	"github.com/xanzy/go-gitlab"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
)

const (
	errNotProject       = "managed resource is not a Gitlab project custom resource"
	errKubeUpdateFailed = "cannot update Gitlab project custom resource"
	errCreateFailed     = "cannot create Gitlab project"
	errUpdateFailed     = "cannot update Gitlab project"
	errDeleteFailed     = "cannot delete Gitlab project"
	errGetFailed        = "cannot retrieve Gitlab project with"
)

// SetupProject adds a controller that reconciles Projects.
func SetupProject(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.ProjectKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Project{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.ProjectGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabClientFn: projects.NewProjectClient}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg clients.Config) projects.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Project)
	if !ok {
		return nil, errors.New(errNotProject)
	}
	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

type external struct {
	kube   client.Client
	client projects.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
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

	current := cr.Spec.ForProvider.DeepCopy()
	lateInitialize(&cr.Spec.ForProvider, prj)

	cr.Status.AtProvider = projects.GenerateObservation(prj)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isProjectUpToDate(&cr.Spec.ForProvider, prj),
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

	meta.SetExternalName(cr, strconv.Itoa(prj.ID))
	return managed.ExternalCreation{ExternalNameAssigned: true}, errors.Wrap(err, errKubeUpdateFailed)
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

	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Project)
	if !ok {
		return errors.New(errNotProject)
	}

	_, err := e.client.DeleteProject(meta.GetExternalName(cr), gitlab.WithContext(ctx))
	return errors.Wrap(err, errDeleteFailed)
}

// lateInitialize fills the empty fields in the project spec with the
// values seen in gitlab.Project.
func lateInitialize(in *v1alpha1.ProjectParameters, project *gitlab.Project) { // nolint:gocyclo
	if project == nil {
		return
	}
	if in.AllowMergeOnSkippedPipeline == nil {
		in.AllowMergeOnSkippedPipeline = &project.AllowMergeOnSkippedPipeline
	}
	if in.ApprovalsBeforeMerge == nil {
		in.ApprovalsBeforeMerge = &project.ApprovalsBeforeMerge
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
	if in.ContainerRegistryEnabled == nil {
		in.ContainerRegistryEnabled = &project.ContainerRegistryEnabled
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
	if in.MirrorUserID == nil {
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

	if len(in.TagList) == 0 && len(project.TagList) > 0 {
		in.TagList = project.TagList
	}

	in.Visibility = clients.LateInitializeVisibilityValue(in.Visibility, project.Visibility)
	in.WikiAccessLevel = clients.LateInitializeAccessControlValue(in.WikiAccessLevel, project.WikiAccessLevel)
}

// isProjectUpToDate checks whether there is a change in any of the modifiable fields.
func isProjectUpToDate(p *v1alpha1.ProjectParameters, g *gitlab.Project) bool { // nolint:gocyclo
	if p.Name != nil && !cmp.Equal(*p.Name, g.Name) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.AllowMergeOnSkippedPipeline, g.AllowMergeOnSkippedPipeline) {
		return false
	}
	if !clients.IsIntEqualToIntPtr(p.ApprovalsBeforeMerge, g.ApprovalsBeforeMerge) {
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
	if !clients.IsBoolEqualToBoolPtr(p.ContainerRegistryEnabled, g.ContainerRegistryEnabled) {
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
	if !cmp.Equal(p.TagList, g.TagList, cmpopts.EquateEmpty()) {
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
