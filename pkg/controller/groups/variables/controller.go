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

package variables

import (
	"context"

	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups"
)

const (
	errNotVariable       = "managed resource is not a Gitlab variable custom resource"
	errGetFailed         = "cannot get Gitlab variable"
	errCreateFailed      = "cannot create Gitlab variable"
	errUpdateFailed      = "cannot update Gitlab variable"
	errDeleteFailed      = "cannot delete Gitlab variable"
	errGetSecretFailed   = "cannot get secret for Gitlab variable value"
	errSecretKeyNotFound = "cannot find key in secret for Gitlab variable value"
	errGroupIDMissing    = "GroupID is missing"
)

// SetupVariable adds a controller that reconciles Variables.
func SetupVariable(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.VariableKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Variable{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.VariableGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabClientFn: groups.NewVariableClient}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg clients.Config) groups.VariableClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Variable)
	if !ok {
		return nil, errors.New(errNotVariable)
	}
	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

type external struct {
	kube   client.Client
	client groups.VariableClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Variable)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotVariable)
	}
	if cr.Spec.ForProvider.GroupID == nil {
		return managed.ExternalObservation{}, errors.New(errGroupIDMissing)
	}

	variable, res, err := e.client.GetVariable(
		*cr.Spec.ForProvider.GroupID,
		cr.Spec.ForProvider.Key,
		gitlab.WithContext(ctx))

	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}

	if cr.Spec.ForProvider.ValueSecretRef != nil {
		if err = e.updateVariableFromSecret(ctx, cr.Spec.ForProvider.ValueSecretRef, &cr.Spec.ForProvider); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errUpdateFailed)
		}
	}

	current := cr.Spec.ForProvider.DeepCopy()
	groups.LateInitializeVariable(&cr.Spec.ForProvider, variable)

	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        groups.IsVariableUpToDate(&cr.Spec.ForProvider, variable),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Variable)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotVariable)
	}

	if cr.Spec.ForProvider.ValueSecretRef != nil {
		if err := e.updateVariableFromSecret(ctx, cr.Spec.ForProvider.ValueSecretRef, &cr.Spec.ForProvider); err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
		}
	}
	if cr.Spec.ForProvider.GroupID == nil {
		return managed.ExternalCreation{}, errors.New(errGroupIDMissing)
	}

	cr.Status.SetConditions(xpv1.Creating())
	_, _, err := e.client.CreateVariable(
		*cr.Spec.ForProvider.GroupID,
		groups.GenerateCreateVariableOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx))

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Variable)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotVariable)
	}

	if cr.Spec.ForProvider.ValueSecretRef != nil {
		if err := e.updateVariableFromSecret(ctx, cr.Spec.ForProvider.ValueSecretRef, &cr.Spec.ForProvider); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
		}
	}
	if cr.Spec.ForProvider.GroupID == nil {
		return managed.ExternalUpdate{}, errors.New(errGroupIDMissing)
	}

	_, _, err := e.client.UpdateVariable(
		*cr.Spec.ForProvider.GroupID,
		cr.Spec.ForProvider.Key,
		groups.GenerateUpdateVariableOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Variable)
	if !ok {
		return errors.New(errNotVariable)
	}

	if cr.Spec.ForProvider.GroupID == nil {
		return errors.New(errGroupIDMissing)
	}

	cr.Status.SetConditions(xpv1.Deleting())
	_, err := e.client.RemoveVariable(
		*cr.Spec.ForProvider.GroupID,
		cr.Spec.ForProvider.Key,
		gitlab.WithContext(ctx),
	)
	return errors.Wrap(err, errDeleteFailed)
}

func (e *external) updateVariableFromSecret(ctx context.Context, selector *xpv1.SecretKeySelector, params *v1alpha1.VariableParameters) error {
	// Fetch the Kubernetes secret.
	secret := &corev1.Secret{}
	nn := types.NamespacedName{
		Namespace: selector.Namespace,
		Name:      selector.Name,
	}

	err := e.kube.Get(ctx, nn, secret)
	if err != nil {
		return errors.Wrap(err, errGetSecretFailed)
	}

	// Obtain the data from the secret.
	raw, ok := secret.Data[selector.Key]
	if raw == nil || !ok {
		return errors.New(errSecretKeyNotFound)
	}

	// Mask variable if it hasn't already been explicitly configured.
	if params.Masked == nil {
		params.Masked = gitlab.Bool(true)
	}

	// Make variable raw if it hasn't already been explicitly configured.
	if params.Raw == nil {
		params.Raw = gitlab.Bool(true)
	}

	value := string(raw)
	params.Value = &value

	return nil
}
