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

package samlgrouplinks

import (
	"context"

	"github.com/crossplane/crossplane-runtime/v2/apis/common"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiCluster "github.com/crossplane-contrib/provider-gitlab/apis/cluster/groups/v1alpha1"
	apiNamespaced "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	sharedGroupsV1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/shared/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups"
)

const (
	ErrNotSamlGroupLink       = "managed resource is not a SamlGroupLink custom resource"
	ErrGetFailed              = "cannot get Gitlab SamlGroupLink"
	ErrCreateFailed           = "cannot create Gitlab SamlGroupLink"
	ErrDeleteFailed           = "cannot delete Gitlab SamlGroupLink"
	ErrSamlGroupLinktNotFound = "cannot find Gitlab SamlGroupLink"
	ErrMissingGroupID         = "missing Spec.ForProvider.GroupID"
	ErrMissingExternalName    = "external name annotation not found"
)

type External struct {
	Client groups.SamlGroupLinkClient
	Kube   client.Client
}

type options struct {
	externalName  string
	parameters    *sharedGroupsV1alpha1.SamlGroupLinkParameters
	atProvider    *sharedGroupsV1alpha1.SamlGroupLinkObservation
	setConditions func(c ...common.Condition)
	mg            resource.Managed
}

func (e *External) extractOptions(mg resource.Managed) (*options, error) {
	switch cr := mg.(type) {
	case *apiCluster.SamlGroupLink:
		return &options{
			externalName:  meta.GetExternalName(cr),
			parameters:    &cr.Spec.ForProvider.SamlGroupLinkParameters,
			atProvider:    &cr.Status.AtProvider,
			setConditions: cr.Status.SetConditions,
			mg:            mg,
		}, nil
	case *apiNamespaced.SamlGroupLink:
		return &options{
			externalName:  meta.GetExternalName(cr),
			parameters:    &cr.Spec.ForProvider.SamlGroupLinkParameters,
			atProvider:    &cr.Status.AtProvider,
			setConditions: cr.Status.SetConditions,
			mg:            mg,
		}, nil
	default:
		return nil, errors.New(ErrNotSamlGroupLink)
	}
}

func (e *External) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	o, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if o.externalName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	if o.parameters.GroupID == nil {
		return managed.ExternalObservation{}, errors.New(ErrMissingGroupID)
	}

	groupLink, _, err := e.Client.GetGroupSAMLLink(*o.parameters.GroupID, o.externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(groups.IsErrorSamlGroupLinkNotFound, err), ErrGetFailed)
	}

	*o.atProvider = groups.GenerateAddSamlGroupLinkObservation(groupLink)
	o.setConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isSamlGroupLinkUpToDate(o.parameters, groupLink),
		ResourceLateInitialized: false,
	}, nil
}

func (e *External) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	o, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	if o.parameters.GroupID == nil {
		return managed.ExternalCreation{}, errors.New(ErrMissingGroupID)
	}

	samlGroupLink, _, err := e.Client.AddGroupSAMLLink(
		*o.parameters.GroupID,
		groups.GenerateAddSamlGroupLinkOptions(o.parameters),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, ErrCreateFailed)
	}

	meta.SetExternalName(o.mg, samlGroupLink.Name)

	return managed.ExternalCreation{}, nil
}

func (e *External) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// not able to update SamlGroupLink
	return managed.ExternalUpdate{}, nil
}

func (e *External) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	o, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalDelete{}, err
	}

	if o.parameters.GroupID == nil {
		return managed.ExternalDelete{}, errors.New(ErrMissingGroupID)
	}

	if o.externalName == "" {
		return managed.ExternalDelete{}, errors.New(ErrMissingExternalName)
	}

	_, err = e.Client.DeleteGroupSAMLLink(
		*o.parameters.GroupID,
		o.externalName,
		nil,
		gitlab.WithContext(ctx),
	)
	return managed.ExternalDelete{}, errors.Wrap(err, ErrDeleteFailed)
}

func isSamlGroupLinkUpToDate(p *sharedGroupsV1alpha1.SamlGroupLinkParameters, g *gitlab.SAMLGroupLink) bool {
	if !cmp.Equal(int(p.AccessLevel), int(g.AccessLevel)) {
		return false
	}

	if !cmp.Equal(*p.Name, g.Name) {
		return false
	}
	return true
}

func (e *External) Disconnect(ctx context.Context) error {
	return nil
}
