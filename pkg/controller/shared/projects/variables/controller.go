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

	"github.com/crossplane/crossplane-runtime/v2/apis/common"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/errors"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiCluster "github.com/crossplane-contrib/provider-gitlab/apis/cluster/projects/v1alpha1"
	apiNamespaced "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	sharedProjectsV1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/shared/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
)

const (
	ErrNotVariable       = "managed resource is not a Gitlab variable custom resource"
	ErrGetFailed         = "cannot get Gitlab variable"
	ErrCreateFailed      = "cannot create Gitlab variable"
	ErrUpdateFailed      = "cannot update Gitlab variable"
	ErrDeleteFailed      = "cannot delete Gitlab variable"
	ErrGetSecretFailed   = "cannot get secret for Gitlab variable value"
	ErrSecretKeyNotFound = "cannot find key in secret for Gitlab variable value"
	ErrProjectIDMissing  = "ProjectID is missing"
)

type External struct {
	Client projects.VariableClient
	Kube   client.Client
}

type options struct {
	parameters     *sharedProjectsV1alpha1.VariableParameters
	valueSecretRef *xpv1.SecretKeySelector
	setConditions  func(c ...common.Condition)
	mg             resource.Managed
}

func (e *External) extractOptions(mg resource.Managed) (*options, error) {
	switch cr := mg.(type) {
	case *apiCluster.Variable:
		return &options{
			parameters:     &cr.Spec.ForProvider.VariableParameters,
			valueSecretRef: cr.Spec.ForProvider.ValueSecretRef,
			setConditions:  cr.SetConditions,
			mg:             mg,
		}, nil
	case *apiNamespaced.Variable:
		var valueSecretRef *xpv1.SecretKeySelector
		if cr.Spec.ForProvider.ValueSecretRef != nil {
			valueSecretRef = cr.Spec.ForProvider.ValueSecretRef.ToSecretKeySelector(cr.GetNamespace())
		}
		return &options{
			parameters:     &cr.Spec.ForProvider.VariableParameters,
			valueSecretRef: valueSecretRef,
			setConditions:  cr.SetConditions,
			mg:             mg,
		}, nil
	default:
		return nil, errors.New(ErrNotVariable)
	}
}

func (e *External) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if opts.parameters.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(ErrProjectIDMissing)
	}

	variable, res, err := e.Client.GetVariable(
		*opts.parameters.ProjectID,
		opts.parameters.Key,
		projects.GenerateGetVariableOptions(opts.parameters),
		gitlab.WithContext(ctx))
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, ErrGetFailed)
	}

	if opts.valueSecretRef != nil {
		if err = e.updateVariableFromSecret(ctx, opts.valueSecretRef, opts.parameters); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, ErrUpdateFailed)
		}
	}

	current := opts.parameters.DeepCopy()
	projects.LateInitializeVariable(opts.parameters, variable)

	opts.setConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        projects.IsVariableUpToDate(opts.parameters, variable),
		ResourceLateInitialized: !cmp.Equal(current, opts.parameters),
	}, nil
}

func (e *External) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	if opts.valueSecretRef != nil {
		if err := e.updateVariableFromSecret(ctx, opts.valueSecretRef, opts.parameters); err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, ErrCreateFailed)
		}
	}
	if opts.parameters.ProjectID == nil {
		return managed.ExternalCreation{}, errors.New(ErrProjectIDMissing)
	}

	opts.setConditions(xpv1.Creating())
	_, _, err = e.Client.CreateVariable(
		*opts.parameters.ProjectID,
		projects.GenerateCreateVariableOptions(opts.parameters),
		gitlab.WithContext(ctx))
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, ErrCreateFailed)
	}
	return managed.ExternalCreation{}, nil
}

func (e *External) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	if opts.valueSecretRef != nil {
		if err := e.updateVariableFromSecret(ctx, opts.valueSecretRef, opts.parameters); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, ErrUpdateFailed)
		}
	}
	if opts.parameters.ProjectID == nil {
		return managed.ExternalUpdate{}, errors.New(ErrProjectIDMissing)
	}

	_, _, err = e.Client.UpdateVariable(
		*opts.parameters.ProjectID,
		opts.parameters.Key,
		projects.GenerateUpdateVariableOptions(opts.parameters),
		gitlab.WithContext(ctx),
	)
	return managed.ExternalUpdate{}, errors.Wrap(err, ErrUpdateFailed)
}

func (e *External) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalDelete{}, err
	}

	if opts.parameters.ProjectID == nil {
		return managed.ExternalDelete{}, errors.New(ErrProjectIDMissing)
	}

	opts.setConditions(xpv1.Deleting())
	_, err = e.Client.RemoveVariable(
		*opts.parameters.ProjectID,
		opts.parameters.Key,
		projects.GenerateRemoveVariableOptions(opts.parameters),
		gitlab.WithContext(ctx),
	)
	return managed.ExternalDelete{}, errors.Wrap(err, ErrDeleteFailed)
}

func (e *External) updateVariableFromSecret(ctx context.Context, selector *xpv1.SecretKeySelector, params *sharedProjectsV1alpha1.VariableParameters) error {
	// Fetch the Kubernetes secret.
	secret := &corev1.Secret{}
	nn := types.NamespacedName{
		Namespace: selector.Namespace,
		Name:      selector.Name,
	}

	err := e.Kube.Get(ctx, nn, secret)
	if err != nil {
		return errors.Wrap(err, ErrGetSecretFailed)
	}

	// Obtain the data from the secret.
	raw, ok := secret.Data[selector.Key]
	if raw == nil || !ok {
		return errors.New(ErrSecretKeyNotFound)
	}

	// Mask variable if it hasn't already been explicitly configured.
	if params.Masked == nil {
		params.Masked = gitlab.Ptr(true)
	}

	// Make variable raw if it hasn't already been explicitly configured.
	if params.Raw == nil {
		params.Raw = gitlab.Ptr(true)
	}

	value := string(raw)
	params.Value = &value

	return nil
}

func (e *External) Disconnect(ctx context.Context) error {
	return nil
}
