/*
Copyright 2026 The Crossplane Authors.

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

package repositoryfiles

import (
	"context"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/statemetrics"
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects"
)

const (
	errNotRepositoryFile   = "managed resource is not a Gitlab repository file custom resource"
	errProjectIDMissing    = "ProjectID is missing"
	errGetFailed           = "cannot get Gitlab repository file"
	errCreateFailed        = "cannot create Gitlab repository file"
	errUpdateFailed        = "cannot update Gitlab repository file"
	errDeleteFailed        = "cannot delete Gitlab repository file"
	errContentFailed       = "cannot resolve Gitlab repository file content"
	errContentMissing      = "one of content or contentSecretRef must be set"
	errContentAmbiguous    = "content and contentSecretRef are mutually exclusive"
	errContentSecretFailed = "cannot read contentSecretRef"
)

// SetupRepositoryFile adds a controller that reconciles RepositoryFiles.
func SetupRepositoryFile(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.RepositoryFileGroupKind)
	managementPolicies := o.Features.Enabled(feature.EnableBetaManagementPolicies)

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnector(&connector{
			kube:                      mgr.GetClient(),
			newGitlabClientFn:         projects.NewRepositoryFileClient,
			managementPoliciesEnabled: managementPolicies,
		}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}

	if managementPolicies {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.RepositoryFileGroupVersionKind),
		reconcilerOpts...)

	if err := mgr.Add(statemetrics.NewMRStateRecorder(
		mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &v1alpha1.RepositoryFileList{}, o.MetricOptions.PollStateMetricInterval)); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.RepositoryFile{}).
		Complete(r)
}

// SetupRepositoryFileGated adds a controller with CRD gate support.
func SetupRepositoryFileGated(mgr ctrl.Manager, o controller.Options) error {
	o.Gate.Register(func() {
		if err := SetupRepositoryFile(mgr, o); err != nil {
			mgr.GetLogger().Error(err, "unable to setup reconciler", "gvk", v1alpha1.RepositoryFileGroupVersionKind.String())
		}
	}, v1alpha1.RepositoryFileGroupVersionKind)
	return nil
}

type connector struct {
	kube                      client.Client
	newGitlabClientFn         func(cfg common.Config) projects.RepositoryFileClient
	managementPoliciesEnabled bool
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.RepositoryFile)
	if !ok {
		return nil, errors.New(errNotRepositoryFile)
	}
	cfg, err := common.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{
		kube:                      c.kube,
		client:                    c.newGitlabClientFn(*cfg),
		managementPoliciesEnabled: c.managementPoliciesEnabled,
	}, nil
}

type external struct {
	kube                      client.Client
	client                    projects.RepositoryFileClient
	managementPoliciesEnabled bool
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.RepositoryFile)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRepositoryFile)
	}
	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(errProjectIDMissing)
	}

	file, res, err := e.client.GetFileMetaData(
		*cr.Spec.ForProvider.ProjectID,
		cr.Spec.ForProvider.FilePath,
		projects.GenerateGetFileMetaDataOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	projects.LateInitializeRepositoryFile(&cr.Spec.ForProvider, file)
	cr.Status.AtProvider = projects.GenerateRepositoryFileObservation(file)
	cr.Status.SetConditions(v2.Available())

	o := managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        true,
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}

	// Drift only matters if we are allowed to correct it. Not resolving the
	// content otherwise means its source may go away once the file exists,
	// e.g. a one-off bootstrap Secret used with ["Observe", "Create"].
	policy := managed.NewManagementPoliciesResolver(e.managementPoliciesEnabled, cr.GetManagementPolicies())
	if !policy.ShouldUpdate() {
		return o, nil
	}

	content, err := e.content(ctx, cr)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errContentFailed)
	}
	o.ResourceUpToDate = projects.IsRepositoryFileUpToDate(&cr.Spec.ForProvider, file, content)
	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.RepositoryFile)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRepositoryFile)
	}
	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalCreation{}, errors.New(errProjectIDMissing)
	}

	content, err := e.content(ctx, cr)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errContentFailed)
	}

	cr.Status.SetConditions(v2.Creating())
	_, _, err = e.client.CreateFile(
		*cr.Spec.ForProvider.ProjectID,
		cr.Spec.ForProvider.FilePath,
		projects.GenerateCreateFileOptions(&cr.Spec.ForProvider, content),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	// The file is identified by spec.forProvider; the external name is
	// informational only.
	meta.SetExternalName(cr, cr.Spec.ForProvider.FilePath)
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.RepositoryFile)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRepositoryFile)
	}
	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalUpdate{}, errors.New(errProjectIDMissing)
	}

	content, err := e.content(ctx, cr)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errContentFailed)
	}

	// Observe has just refreshed LastCommitID, so GitLab rejects the update
	// if someone else committed to the file in between.
	_, _, err = e.client.UpdateFile(
		*cr.Spec.ForProvider.ProjectID,
		cr.Spec.ForProvider.FilePath,
		projects.GenerateUpdateFileOptions(&cr.Spec.ForProvider, content, cr.Status.AtProvider.LastCommitID),
		gitlab.WithContext(ctx),
	)
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.RepositoryFile)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotRepositoryFile)
	}
	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalDelete{}, errors.New(errProjectIDMissing)
	}

	cr.Status.SetConditions(v2.Deleting())
	res, err := e.client.DeleteFile(
		*cr.Spec.ForProvider.ProjectID,
		cr.Spec.ForProvider.FilePath,
		projects.GenerateDeleteFileOptions(&cr.Spec.ForProvider, cr.Status.AtProvider.LastCommitID),
		gitlab.WithContext(ctx),
	)
	if err != nil && !clients.IsResponseNotFound(res) {
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteFailed)
	}
	return managed.ExternalDelete{}, nil
}

func (e *external) Disconnect(_ context.Context) error {
	return nil
}

// content returns the bytes the file should contain.
func (e *external) content(ctx context.Context, cr *v1alpha1.RepositoryFile) ([]byte, error) {
	p := &cr.Spec.ForProvider
	switch {
	case p.Content != nil && p.ContentSecretRef != nil:
		return nil, errors.New(errContentAmbiguous)
	case p.Content != nil:
		return projects.DecodeRepositoryFileContent(*p.Content, p.Encoding)
	case p.ContentSecretRef != nil:
		v, err := common.GetTokenValueFromLocalSecret(ctx, e.kube, cr, p.ContentSecretRef)
		if err != nil {
			return nil, errors.Wrap(err, errContentSecretFailed)
		}
		return projects.DecodeRepositoryFileContent(*v, p.Encoding)
	default:
		return nil, errors.New(errContentMissing)
	}
}
