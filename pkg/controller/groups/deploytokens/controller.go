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

package deploytokens

import (
	"context"
	"strconv"

	"github.com/xanzy/go-gitlab"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups"
)

const (
	errNotDeployToken = "managed resource is not a Gitlab deploytoken custom resource"
	errGetFailed      = "cannot get Gitlab deploytoken"
	errCreateFailed   = "cannot create Gitlab deploytoken"
	errDeleteFailed   = "cannot delete Gitlab deploytoken"
	errIDNotInt       = "ID is not integer value"
	errGroupIDMissing = "GroupID is missing"
)

// SetupDeployToken adds a controller that reconciles GroupDeployTokens.
func SetupDeployToken(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.DeployTokenKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.DeployToken{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.DeployTokenGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabClientFn: groups.NewDeployTokenClient}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg clients.Config) groups.DeployTokenClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.DeployToken)
	if !ok {
		return nil, errors.New(errNotDeployToken)
	}
	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

type external struct {
	kube   client.Client
	client groups.DeployTokenClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.DeployToken)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotDeployToken)
	}

	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	id, err := strconv.Atoi(externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.New(errIDNotInt)
	}

	if cr.Spec.ForProvider.GroupID == nil {
		return managed.ExternalObservation{}, errors.New(errGroupIDMissing)
	}

	dt, res, err := e.client.GetGroupDeployToken(*cr.Spec.ForProvider.GroupID, id)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	lateInitializeGroupDeployToken(&cr.Spec.ForProvider, dt)

	cr.Status.AtProvider = v1alpha1.DeployTokenObservation{}
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        true,
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.DeployToken)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotDeployToken)
	}

	if cr.Spec.ForProvider.GroupID == nil {
		return managed.ExternalCreation{}, errors.Wrap(errors.New("GroupId must be set directly or via reference"), errCreateFailed)
	}

	dt, _, err := e.client.CreateGroupDeployToken(
		*cr.Spec.ForProvider.GroupID,
		groups.GenerateCreateGroupDeployTokenOptions(cr.Name, &cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)

	connectionDetails := managed.ConnectionDetails{}
	connectionDetails["token"] = []byte(dt.Token)

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	meta.SetExternalName(cr, strconv.Itoa(dt.ID))
	return managed.ExternalCreation{ExternalNameAssigned: true, ConnectionDetails: connectionDetails}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// it's not possible to update a GroupDeployToken
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.DeployToken)
	if !ok {
		return errors.New(errNotDeployToken)
	}

	deployTokenID, err := strconv.Atoi(meta.GetExternalName(cr))

	if err != nil {
		return errors.New(errNotDeployToken)
	}

	if cr.Spec.ForProvider.GroupID == nil {
		return errors.New(errGroupIDMissing)
	}

	_, deleteError := e.client.DeleteGroupDeployToken(
		*cr.Spec.ForProvider.GroupID,
		deployTokenID,
		gitlab.WithContext(ctx),
	)

	return errors.Wrap(deleteError, errDeleteFailed)
}

// lateInitializeGroupDeployToken fills the empty fields in the deploy token spec with the
// values seen in gitlab deploy token.
func lateInitializeGroupDeployToken(in *v1alpha1.DeployTokenParameters, deployToken *gitlab.DeployToken) { // nolint:gocyclo
	if deployToken == nil {
		return
	}

	if in.Username == nil {
		in.Username = &deployToken.Username
	}

	if in.ExpiresAt == nil && deployToken.ExpiresAt != nil {
		in.ExpiresAt = &metav1.Time{Time: *deployToken.ExpiresAt}
	}
}
