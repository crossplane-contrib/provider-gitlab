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

package project

import (
	"context"
	"strconv"
	"time"

	"github.com/xanzy/go-gitlab"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-gitlab/apis/deploytoken/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/deploytoken"
)

const (
	errNotProjectDeployToken = "managed resource is not a Gitlab deploytoken custom resource"
	errGetFailed             = "cannot get Gitlab deploytoken"
	errKubeUpdateFailed      = "cannot update Gitlab deploytoken custom resource"
	errCreateFailed          = "cannot create Gitlab deploytoken"
	errUpdateFailed          = "cannot update Gitlab deploytoken"
	errDeleteFailed          = "cannot delete Gitlab deploytoken"
)

// SetupProjectDeployToken adds a controller that reconciles ProjectDeployTokens.
func SetupProjectDeployToken(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.ProjectDeployTokenKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.ProjectDeployToken{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.ProjectDeployTokenGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabClientFn: deploytoken.NewProjectDeployTokenClient}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg clients.Config) deploytoken.ProjectDeployTokenClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.ProjectDeployToken)
	if !ok {
		return nil, errors.New(errNotProjectDeployToken)
	}
	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

type external struct {
	kube   client.Client
	client deploytoken.ProjectDeployTokenClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.ProjectDeployToken)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotProjectDeployToken)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	deployTokenId, err := strconv.Atoi(meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalObservation{}, errors.New(errNotProjectDeployToken)
	}

	dt := &gitlab.DeployToken{}
	listProjectDeployTokensOptions := gitlab.ListProjectDeployTokensOptions{}
	deploytokenArr, _, err := e.client.ListProjectDeployTokens(*cr.Spec.ForProvider.ProjectID, &listProjectDeployTokensOptions)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(deploytoken.IsErrorProjectNotFound, err), errGetFailed)
	}

	for _, v := range deploytokenArr {
		if v.ID == deployTokenId {
			dt = v
			break
		}
	}

	if dt.Name == "" {
		return managed.ExternalObservation{}, nil
	}

	current := cr.Spec.ForProvider.DeepCopy()
	lateInitializeProjectDeployToken(&cr.Spec.ForProvider, dt)

	cr.Status.AtProvider = generateProjectDeployTokenObservation(dt)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        true,
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ProjectDeployToken)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotProjectDeployToken)
	}

	dt, _, err := e.client.CreateProjectDeployToken(*cr.Spec.ForProvider.ProjectID, generateCreateProjectDeployTokenOptions(&cr.Spec.ForProvider), gitlab.WithContext(ctx))
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	connectionDetails := managed.ConnectionDetails{}
	connectionDetails["token"] = []byte(dt.Token)

	meta.SetExternalName(cr, strconv.Itoa(dt.ID))

	return managed.ExternalCreation{
		ExternalNameAssigned: true,
		ConnectionDetails:    connectionDetails,
	}, errors.Wrap(err, errKubeUpdateFailed)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// it's not possible to update a ProjectDeployToken
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.ProjectDeployToken)
	if !ok {
		return errors.New(errNotProjectDeployToken)
	}

	_, err := e.client.DeleteProjectDeployToken(*cr.Spec.ForProvider.ProjectID, cr.Status.AtProvider.ID, gitlab.WithContext(ctx))
	return errors.Wrap(err, errDeleteFailed)
}

// lateInitializeProjectDeployToken fills the empty fields in the deploy token spec with the
// values seen in gitlab deploy token.
func lateInitializeProjectDeployToken(in *v1alpha1.ProjectDeployTokenParameters, deployToken *gitlab.DeployToken) { // nolint:gocyclo
	if deployToken == nil {
		return
	}

	if in.Username == nil {
		in.Username = &deployToken.Username
	}

	if in.ExpiresAt == nil && deployToken.ExpiresAt != nil {
		in.ExpiresAt = &metav1.Time{Time: time.Time(*deployToken.ExpiresAt)}
	}
}

// generateProjectDeployTokenObservation is used to produce v1alpha1.ProjectDeployTokenObservation from
// gitlab.Project.
func generateProjectDeployTokenObservation(deploytoken *gitlab.DeployToken) v1alpha1.ProjectDeployTokenObservation { // nolint:gocyclo
	if deploytoken == nil {
		return v1alpha1.ProjectDeployTokenObservation{}
	}

	o := v1alpha1.ProjectDeployTokenObservation{
		ID: deploytoken.ID,
	}

	return o
}

// generateCreateProjectDeployTokenOptions generates project creation options
func generateCreateProjectDeployTokenOptions(p *v1alpha1.ProjectDeployTokenParameters) *gitlab.CreateProjectDeployTokenOptions {
	deploytoken := &gitlab.CreateProjectDeployTokenOptions{
		Name:   &p.Name,
		Scopes: p.Scopes,
	}

	if p.ExpiresAt != nil {
		deploytoken.ExpiresAt = &p.ExpiresAt.Time
	}

	if p.Username != nil {
		deploytoken.Username = p.Username
	}

	return deploytoken
}
