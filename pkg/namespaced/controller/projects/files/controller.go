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

package files

import (
	"context"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/statemetrics"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects"
)

const (
	errNotFile          = "managed resource is not a Gitlab file custom resource"
	errGetFailed        = "cannot get Gitlab file"
	errCreateFailed     = "cannot create Gitlab file"
	errUpdateFailed     = "cannot update Gitlab file"
	errDeleteFailed     = "cannot delete Gitlab file"
	errProjectIDMissing = "ProjectID is missing"
)

// SetupFile adds a controller that reconciles Files.
func SetupFile(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.FileGroupKind)

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabFileClientFn: projects.NewFileClient, newGitlabCommitClientFn: projects.NewCommitsClient}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}

	if o.Features.Enabled(feature.EnableBetaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.FileGroupVersionKind),
		reconcilerOpts...)

	if err := mgr.Add(statemetrics.NewMRStateRecorder(
		mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &v1alpha1.FileList{}, o.MetricOptions.PollStateMetricInterval)); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.File{}).
		Complete(r)
}

// SetupFileGated adds a controller with CRD gate support.
func SetupFileGated(mgr ctrl.Manager, o controller.Options) error {
	o.Gate.Register(func() {
		if err := SetupFile(mgr, o); err != nil {
			mgr.GetLogger().Error(err, "unable to setup reconciler", "gvk", v1alpha1.FileGroupVersionKind.String())
		}
	}, v1alpha1.FileGroupVersionKind)
	return nil
}

type connector struct {
	kube                    client.Client
	newGitlabFileClientFn   func(cfg common.Config) projects.FileClient
	newGitlabCommitClientFn func(cfg common.Config) projects.CommitClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.File)
	if !ok {
		return nil, errors.New(errNotFile)
	}
	cfg, err := common.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, fileClient: c.newGitlabFileClientFn(*cfg), commitClient: c.newGitlabCommitClientFn(*cfg)}, nil
}

type external struct {
	kube         client.Client
	fileClient   projects.FileClient
	commitClient projects.CommitClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.File)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotFile)
	}

	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalObservation{}, nil
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(errProjectIDMissing)
	}

	file, res, err := e.fileClient.GetFile(
		*cr.Spec.ForProvider.ProjectID,
		*cr.Spec.ForProvider.FilePath,
		projects.GenerateGetFileOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx))

	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}

	commit, _, _ := e.commitClient.GetCommit(
		*cr.Spec.ForProvider.ProjectID,
		file.CommitID,
		projects.GenerateGetCommitOptions(),
		gitlab.WithContext(ctx))

	current := cr.Spec.ForProvider.DeepCopy()
	projects.LateInitializeFile(&cr.Spec.ForProvider, file, commit)

	cr.Status.AtProvider = projects.GenerateFileObservation(file)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        projects.IsFileUpToDate(&cr.Spec.ForProvider, file),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.File)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotFile)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalCreation{}, errors.New(errProjectIDMissing)
	}

	cr.Status.SetConditions(xpv1.Creating())
	file, _, err := e.fileClient.CreateFile(
		*cr.Spec.ForProvider.ProjectID,
		*cr.Spec.ForProvider.FilePath,
		projects.GenerateCreateFileOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx))

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	meta.SetExternalName(cr, file.FilePath)
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.File)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotFile)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalUpdate{}, errors.New(errProjectIDMissing)
	}

	_, _, err := e.fileClient.UpdateFile(
		*cr.Spec.ForProvider.ProjectID,
		*cr.Spec.ForProvider.FilePath,
		projects.GenerateUpdateFileOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.File)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotFile)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalDelete{}, errors.New(errProjectIDMissing)
	}

	cr.Status.SetConditions(xpv1.Deleting())
	_, err := e.fileClient.DeleteFile(
		*cr.Spec.ForProvider.ProjectID,
		*cr.Spec.ForProvider.FilePath,
		projects.GenerateDeleteFileOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	return managed.ExternalDelete{}, errors.Wrap(err, errDeleteFailed)
}

func (e *external) Disconnect(ctx context.Context) error {
	// Disconnect is not implemented as it is a new method required by the SDK
	return nil
}
