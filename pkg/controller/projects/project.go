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
	"context"
	"reflect"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "github.com/xanzy/go-gitlab"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
)

const (
	errNotProject       = "managed resource is not a Gitlab project custom resource"
	errGetFailed        = "cannot get Gitlab project"
	errKubeUpdateFailed = "cannot update Gitlab project custom resource"
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
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient()), managed.NewNameAsExternalName(mgr.GetClient())),
			// managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
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

	prj, _, err := e.client.GetProject(meta.GetExternalName(cr), nil)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	projects.LateInitialize(&cr.Spec.ForProvider, prj)
	if !reflect.DeepEqual(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(context.Background(), cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}

	cr.Status.AtProvider = projects.GenerateObservation(prj)
	cr.Status.SetConditions(runtimev1alpha1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        projects.IsProjectUpToDate(&cr.Spec.ForProvider, prj),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
		ConnectionDetails:       managed.ConnectionDetails{"runnersToken": []byte(prj.RunnersToken)},
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {

	cr, ok := mg.(*v1alpha1.Project)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotProject)
	}

	name := meta.GetExternalName(cr)

	// Create new project
	p := &gitlab.CreateProjectOptions{
		Name:                             &name,
		Path:                             cr.Spec.ForProvider.Path,
		NamespaceID:                      cr.Spec.ForProvider.NamespaceID,
		DefaultBranch:                    cr.Spec.ForProvider.DefaultBranch,
		Description:                      cr.Spec.ForProvider.Description,
		IssuesAccessLevel:                stringToAccessControlValue(cr.Spec.ForProvider.IssuesAccessLevel),
		RepositoryAccessLevel:            stringToAccessControlValue(cr.Spec.ForProvider.RepositoryAccessLevel),
		MergeRequestsAccessLevel:         stringToAccessControlValue(cr.Spec.ForProvider.MergeRequestsAccessLevel),
		ForkingAccessLevel:               stringToAccessControlValue(cr.Spec.ForProvider.ForkingAccessLevel),
		BuildsAccessLevel:                stringToAccessControlValue(cr.Spec.ForProvider.BuildsAccessLevel),
		WikiAccessLevel:                  stringToAccessControlValue(cr.Spec.ForProvider.WikiAccessLevel),
		SnippetsAccessLevel:              stringToAccessControlValue(cr.Spec.ForProvider.SnippetsAccessLevel),
		PagesAccessLevel:                 stringToAccessControlValue(cr.Spec.ForProvider.PagesAccessLevel),
		EmailsDisabled:                   cr.Spec.ForProvider.EmailsDisabled,
		ResolveOutdatedDiffDiscussions:   cr.Spec.ForProvider.ResolveOutdatedDiffDiscussions,
		ContainerRegistryEnabled:         cr.Spec.ForProvider.ContainerRegistryEnabled,
		SharedRunnersEnabled:             cr.Spec.ForProvider.SharedRunnersEnabled,
		Visibility:                       stringToVisibilityLevel(cr.Spec.ForProvider.Visibility),
		ImportURL:                        cr.Spec.ForProvider.ImportURL,
		PublicBuilds:                     cr.Spec.ForProvider.PublicBuilds,
		OnlyAllowMergeIfPipelineSucceeds: cr.Spec.ForProvider.OnlyAllowMergeIfPipelineSucceeds,
		OnlyAllowMergeIfAllDiscussionsAreResolved: cr.Spec.ForProvider.OnlyAllowMergeIfAllDiscussionsAreResolved,
		MergeMethod:                              stringToMergeMethod(cr.Spec.ForProvider.MergeMethod),
		RemoveSourceBranchAfterMerge:             cr.Spec.ForProvider.RemoveSourceBranchAfterMerge,
		LFSEnabled:                               cr.Spec.ForProvider.LFSEnabled,
		RequestAccessEnabled:                     cr.Spec.ForProvider.RequestAccessEnabled,
		TagList:                                  &cr.Spec.ForProvider.TagList,
		PrintingMergeRequestLinkEnabled:          cr.Spec.ForProvider.PrintingMergeRequestLinkEnabled,
		BuildGitStrategy:                         cr.Spec.ForProvider.BuildGitStrategy,
		BuildTimeout:                             cr.Spec.ForProvider.BuildTimeout,
		AutoCancelPendingPipelines:               cr.Spec.ForProvider.AutoCancelPendingPipelines,
		BuildCoverageRegex:                       cr.Spec.ForProvider.BuildCoverageRegex,
		CIConfigPath:                             cr.Spec.ForProvider.CIConfigPath,
		AutoDevopsEnabled:                        cr.Spec.ForProvider.AutoDevopsEnabled,
		AutoDevopsDeployStrategy:                 cr.Spec.ForProvider.AutoDevopsDeployStrategy,
		ApprovalsBeforeMerge:                     cr.Spec.ForProvider.ApprovalsBeforeMerge,
		ExternalAuthorizationClassificationLabel: cr.Spec.ForProvider.ExternalAuthorizationClassificationLabel,
		Mirror:                                   cr.Spec.ForProvider.Mirror,
		MirrorTriggerBuilds:                      cr.Spec.ForProvider.MirrorTriggerBuilds,
		InitializeWithReadme:                     cr.Spec.ForProvider.InitializeWithReadme,
		TemplateName:                             cr.Spec.ForProvider.TemplateName,
		TemplateProjectID:                        cr.Spec.ForProvider.TemplateProjectID,
		UseCustomTemplate:                        cr.Spec.ForProvider.UseCustomTemplate,
		GroupWithProjectTemplatesID:              cr.Spec.ForProvider.GroupWithProjectTemplatesID,
		PackagesEnabled:                          cr.Spec.ForProvider.PackagesEnabled,
		ServiceDeskEnabled:                       cr.Spec.ForProvider.ServiceDeskEnabled,
		AutocloseReferencedIssues:                cr.Spec.ForProvider.AutocloseReferencedIssues,
	}

	_, _, err := e.client.CreateProject(p)
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	return nil
}

func stringToVisibilityLevel(from *v1alpha1.VisibilityValue) *gitlab.VisibilityValue {
	return (*gitlab.VisibilityValue)(&from)
}

func stringToAccessControlValue(s *v1alpha1.AccessControlValue) *gitlab.AccessControlValue {
	lookup := map[string]gitlab.VisibilityValue{
		"disabled": gitlab.DisabledAccessControl,
		"enabled":  gitlab.EnabledAccessControl,
		"private":  gitlab.PrivateAccessControl,
		"public":   gitlab.PublicAccessControl,
	}

	value, ok := lookup[s]
	if !ok {
		return nil
	}
	return value
}

func stringToMergeMethod(s *v1alpha1.MergeMethodValue) *gitlab.MergeMethodValue {
	lookup := map[string]gitlab.VisibilityValue{
		"merge":        gitlab.NoFastForwardMerge,
		"ff":           gitlab.FastForwardMerge,
		"rebase_merge": gitlab.RebaseMerge,
	}

	value, ok := lookup[s]
	if !ok {
		return nil
	}
	return value
}
