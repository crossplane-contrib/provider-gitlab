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

package protectedbranches

import (
	"context"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/statemetrics"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/cluster/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/cluster/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/cluster/clients/projects"
)

const (
	errNotProtectedBranch = "managed resource is not a GitLab protected branch custom resource"
	errProjectIDMissing   = "ProjectID is missing"
	errGetFailed          = "cannot get GitLab protected branch"
	errCreateFailed       = "cannot create GitLab protected branch"
	errDeleteFailed       = "cannot delete GitLab protected branch"
	errBranchNameMissing  = "branch name is missing from spec.forProvider.branchName"
)

// SetupProtectedBranch adds a controller that reconciles ProtectedBranches.
func SetupProtectedBranch(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.ProtectedBranchGroupKind)

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabClientFn: projects.NewProtectedBranchClient}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}

	if o.Features.Enabled(feature.EnableBetaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.ProtectedBranchGroupVersionKind),
		reconcilerOpts...)

	if err := mgr.Add(statemetrics.NewMRStateRecorder(
		mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &v1alpha1.ProtectedBranchList{}, o.MetricOptions.PollStateMetricInterval)); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.ProtectedBranch{}).
		Complete(r)
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg clients.Config) projects.ProtectedBranchClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.ProtectedBranch)
	if !ok {
		return nil, errors.New(errNotProtectedBranch)
	}
	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

type external struct {
	kube   client.Client
	client projects.ProtectedBranchClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.ProtectedBranch)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotProtectedBranch)
	}

	branchName := cr.Spec.ForProvider.BranchName
	if branchName == "" {
		return managed.ExternalObservation{}, nil
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(errProjectIDMissing)
	}

	protectedBranch, res, err := e.client.GetProtectedBranch(*cr.Spec.ForProvider.ProjectID, branchName)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(projects.IsErrorProtectedBranchNotFound, err), errGetFailed)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	projects.LateInitializeProtectedBranch(&cr.Spec.ForProvider, protectedBranch)

	cr.Status.AtProvider = projects.GenerateProtectedBranchObservation(protectedBranch)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        projects.IsProtectedBranchUpToDate(&cr.Spec.ForProvider, protectedBranch),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ProtectedBranch)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotProtectedBranch)
	}

	branchName := cr.Spec.ForProvider.BranchName
	if branchName == "" {
		return managed.ExternalCreation{}, errors.New(errBranchNameMissing)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalCreation{}, errors.New(errProjectIDMissing)
	}

	cr.Status.SetConditions(xpv1.Creating())

	protectOptions := projects.GenerateProtectRepositoryBranchesOptions(branchName, &cr.Spec.ForProvider)

	_, _, err := e.client.ProtectRepositoryBranches(*cr.Spec.ForProvider.ProjectID, protectOptions, gitlab.WithContext(ctx))
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.ProtectedBranch)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotProtectedBranch)
	}

	branchName := cr.Spec.ForProvider.BranchName
	if branchName == "" {
		return managed.ExternalUpdate{}, errors.New(errBranchNameMissing)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalUpdate{}, errors.New(errProjectIDMissing)
	}

	// GitLab doesn't have a direct "update" API for protected branches.
	// We need to unprotect and then protect again with new settings.
	_, err := e.client.UnprotectRepositoryBranches(*cr.Spec.ForProvider.ProjectID, branchName, gitlab.WithContext(ctx))
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "cannot unprotect branch for update")
	}

	protectOptions := projects.GenerateProtectRepositoryBranchesOptions(branchName, &cr.Spec.ForProvider)
	_, _, err = e.client.ProtectRepositoryBranches(*cr.Spec.ForProvider.ProjectID, protectOptions, gitlab.WithContext(ctx))
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "cannot re-protect branch after update")
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.ProtectedBranch)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotProtectedBranch)
	}

	branchName := cr.Spec.ForProvider.BranchName
	if branchName == "" {
		return managed.ExternalDelete{}, errors.New(errBranchNameMissing)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalDelete{}, errors.New(errProjectIDMissing)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.UnprotectRepositoryBranches(*cr.Spec.ForProvider.ProjectID, branchName, gitlab.WithContext(ctx))
	return managed.ExternalDelete{}, errors.Wrap(err, errDeleteFailed)
}

func (e *external) Disconnect(ctx context.Context) error {
	// Disconnect is not implemented as it is a new method required by the SDK
	return nil
}
