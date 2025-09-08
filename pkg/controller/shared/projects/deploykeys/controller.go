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

package deploykeys

import (
	"context"
	"strconv"

	"github.com/crossplane/crossplane-runtime/v2/apis/common"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiCluster "github.com/crossplane-contrib/provider-gitlab/apis/cluster/projects/v1alpha1"
	apiNamespaced "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	sharedProjectsV1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/shared/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
)

const (
	ErrNotDeployKey     = "managed resource is not a Gitlab deploy key custom resource"
	ErrNotFound         = "404 project deploy key not found"
	ErrGetFail          = "cannot get Gitlab deploy key"
	ErrCreateFail       = "cannot create Gitlab deploy key"
	ErrUpdateFail       = "cannot update Gitlab deploy key"
	ErrDeleteFail       = "cannot delete Gitlab deploy key"
	ErrKeyMissing       = "missing key ref value"
	ErrIDNotAnInt       = "external-name is not an int"
	ErrProjectIDMissing = "missing project ID"
)

type External struct {
	Client projects.DeployKeyClient
	Kube   client.Client
}

type options struct {
	externalName  string
	parameters    *sharedProjectsV1alpha1.DeployKeyParameters
	keySecretRef  *xpv1.SecretKeySelector
	atProvider    *sharedProjectsV1alpha1.DeployKeyObservation
	setConditions func(c ...common.Condition)
	mg            resource.Managed
}

func (e *External) extractOptions(mg resource.Managed) (*options, error) {
	switch cr := mg.(type) {
	case *apiCluster.DeployKey:
		return &options{
			externalName:  meta.GetExternalName(cr),
			parameters:    &cr.Spec.ForProvider.DeployKeyParameters,
			keySecretRef:  &cr.Spec.ForProvider.KeySecretRef,
			atProvider:    &cr.Status.AtProvider,
			setConditions: cr.SetConditions,
			mg:            mg,
		}, nil
	case *apiNamespaced.DeployKey:
		return &options{
			externalName:  meta.GetExternalName(cr),
			parameters:    &cr.Spec.ForProvider.DeployKeyParameters,
			keySecretRef:  cr.Spec.ForProvider.KeySecretRef.ToSecretKeySelector(cr.GetNamespace()),
			atProvider:    &cr.Status.AtProvider,
			setConditions: cr.SetConditions,
			mg:            mg,
		}, nil
	default:
		return nil, errors.New(ErrNotDeployKey)
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

	if opts.externalName == "" {
		return managed.ExternalObservation{}, nil
	}

	id, err := strconv.Atoi(opts.externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.New(ErrIDNotAnInt)
	}

	dk, res, err := e.Client.GetDeployKey(*opts.parameters.ProjectID, id)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, ErrGetFail)
	}

	currentState := opts.parameters.DeepCopy()
	lateInitializeProjectDeployKey(opts.parameters, dk)
	isLateInitialized := !cmp.Equal(currentState, opts.parameters)

	*opts.atProvider = sharedProjectsV1alpha1.DeployKeyObservation{
		ID:        &dk.ID,
		CreatedAt: clients.TimeToMetaTime(dk.CreatedAt),
	}

	opts.setConditions(xpv1.Available())
	isUpToDate := e.isUpToDate(opts.parameters, dk)

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isUpToDate,
		ResourceLateInitialized: isLateInitialized,
	}, nil
}

func (e *External) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	if opts.parameters.ProjectID == nil {
		return managed.ExternalCreation{}, errors.New(ErrProjectIDMissing)
	}

	keySecretRef := opts.keySecretRef

	namespacedName := types.NamespacedName{
		Namespace: keySecretRef.Namespace,
		Name:      keySecretRef.Name,
	}

	secret := &corev1.Secret{}
	err = e.Kube.Get(ctx, namespacedName, secret)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, ErrKeyMissing)
	}

	keyResponse, _, err := e.Client.AddDeployKey(
		*opts.parameters.ProjectID,
		e.generateCreateOptions(string(secret.Data[keySecretRef.Key]), opts.parameters),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, ErrCreateFail)
	}

	id := strconv.Itoa(keyResponse.ID)
	meta.SetExternalName(opts.mg, id)

	return managed.ExternalCreation{}, nil
}

func (e *External) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	if opts.parameters.ProjectID == nil {
		return managed.ExternalUpdate{}, errors.New(ErrProjectIDMissing)
	}

	id, err := strconv.Atoi(opts.externalName)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, ErrIDNotAnInt)
	}

	_, _, er := e.Client.UpdateDeployKey(
		*opts.parameters.ProjectID,
		id,
		e.generateUpdateOptions(opts.parameters),
	)

	return managed.ExternalUpdate{}, errors.Wrap(er, ErrUpdateFail)
}

func (e *External) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalDelete{}, err
	}

	if opts.parameters.ProjectID == nil {
		return managed.ExternalDelete{}, errors.New(ErrProjectIDMissing)
	}

	keyID, err := strconv.Atoi(opts.externalName)
	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, ErrIDNotAnInt)
	}

	_, err = e.Client.DeleteDeployKey(*opts.parameters.ProjectID, keyID)

	return managed.ExternalDelete{}, errors.Wrap(err, ErrDeleteFail)
}

func lateInitializeProjectDeployKey(local *sharedProjectsV1alpha1.DeployKeyParameters, external *gitlab.ProjectDeployKey) {
	if external == nil {
		return
	}

	if local.CanPush == nil {
		local.CanPush = &external.CanPush
	}
}

func (e *External) generateCreateOptions(externalName string, params *sharedProjectsV1alpha1.DeployKeyParameters) *gitlab.AddDeployKeyOptions {
	return &gitlab.AddDeployKeyOptions{
		Key:     &externalName,
		Title:   &params.Title,
		CanPush: params.CanPush,
	}
}

func (e *External) generateUpdateOptions(params *sharedProjectsV1alpha1.DeployKeyParameters) *gitlab.UpdateDeployKeyOptions {
	return &gitlab.UpdateDeployKeyOptions{
		Title:   &params.Title,
		CanPush: params.CanPush,
	}
}

func (e *External) isUpToDate(params *sharedProjectsV1alpha1.DeployKeyParameters, dk *gitlab.ProjectDeployKey) bool {
	isCanPushUpToDate := ptr.Equal(params.CanPush, &dk.CanPush)
	isTitleUpToDate := params.Title == dk.Title

	return isCanPushUpToDate && isTitleUpToDate
}

func (e *External) Disconnect(ctx context.Context) error {
	return nil
}
