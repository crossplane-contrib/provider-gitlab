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

package approvalrules

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
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiCluster "github.com/crossplane-contrib/provider-gitlab/apis/cluster/projects/v1alpha1"
	apiNamespaced "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	sharedProjectsV1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/shared/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
)

const (
	ErrNotApprovalRule  = "managed resource is not a Gitlab Project Approval Rule custom resource"
	ErrCreateFailed     = "cannot create Gitlab Approval Rule"
	ErrUpdateFailed     = "cannot update Gitlab Approval Rule"
	ErrDeleteFailed     = "cannot delete Gitlab Approval Rule"
	ErrObserveFailed    = "cannot observe Gitlab Approval Rule"
	ErrProjectIDMissing = "ProjectID is missing"
	ErrIDNotInt         = "ID is not an integer"
)

type External struct {
	Client projects.ApprovalRulesClient
	Kube   client.Client
}

type options struct {
	externalName  string
	parameters    *sharedProjectsV1alpha1.ApprovalRuleParameters
	atProvider    *sharedProjectsV1alpha1.ApprovalRuleObservation
	setConditions func(c ...common.Condition)
	mg            resource.Managed
}

func (e *External) extractOptions(mg resource.Managed) (*options, error) {
	switch cr := mg.(type) {
	case *apiCluster.ApprovalRule:
		return &options{
			externalName:  meta.GetExternalName(cr),
			parameters:    &cr.Spec.ForProvider.ApprovalRuleParameters,
			atProvider:    &cr.Status.AtProvider,
			setConditions: cr.SetConditions,
			mg:            mg,
		}, nil
	case *apiNamespaced.ApprovalRule:
		return &options{
			externalName:  meta.GetExternalName(cr),
			parameters:    &cr.Spec.ForProvider.ApprovalRuleParameters,
			atProvider:    &cr.Status.AtProvider,
			setConditions: cr.SetConditions,
			mg:            mg,
		}, nil
	default:
		return nil, errors.New(ErrNotApprovalRule)
	}
}

func (e *External) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if opts.externalName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	id, err := strconv.Atoi(opts.externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.New(ErrIDNotInt)
	}

	if opts.parameters.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(ErrProjectIDMissing)
	}

	approvalRule, res, err := e.Client.GetProjectApprovalRule(*opts.parameters.ProjectID, id)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, ErrObserveFailed)
	}

	current := opts.parameters.DeepCopy()

	*opts.atProvider = sharedProjectsV1alpha1.ApprovalRuleObservation{}
	opts.setConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        projects.IsApprovalRuleUpToDate(opts.parameters, approvalRule),
		ResourceLateInitialized: !cmp.Equal(current, opts.parameters),
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

	opts.setConditions(xpv1.Creating())
	approvalRulesOptions := projects.GenerateCreateApprovalRulesOptions(opts.parameters)

	rule, _, err := e.Client.CreateProjectApprovalRule(*opts.parameters.ProjectID, approvalRulesOptions, gitlab.WithContext(ctx))
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, ErrCreateFailed)
	}

	err = e.updateExternalName(ctx, opts.mg, rule)
	return managed.ExternalCreation{}, err
}

func (e *External) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	if opts.parameters.ProjectID == nil {
		return managed.ExternalUpdate{}, errors.New(ErrProjectIDMissing)
	}

	ruleID, err := strconv.Atoi(opts.externalName)
	if err != nil {
		return managed.ExternalUpdate{}, errors.New(ErrIDNotInt)
	}

	_, _, err = e.Client.UpdateProjectApprovalRule(
		*opts.parameters.ProjectID,
		ruleID,
		projects.GenerateUpdateApprovalRulesOptions(opts.parameters),
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

	ruleID, err := strconv.Atoi(opts.externalName)
	if err != nil {
		return managed.ExternalDelete{}, errors.New(ErrIDNotInt)
	}

	_, err = e.Client.DeleteProjectApprovalRule(
		*opts.parameters.ProjectID,
		ruleID,
		gitlab.WithContext(ctx),
	)

	return managed.ExternalDelete{}, errors.Wrap(err, ErrDeleteFailed)
}

func (e *External) updateExternalName(ctx context.Context, mg resource.Managed, approvalRule *gitlab.ProjectApprovalRule) error {
	meta.SetExternalName(mg, strconv.Itoa(approvalRule.ID))
	return e.Kube.Update(ctx, mg)
}

func (e *External) Disconnect(ctx context.Context) error {
	return nil
}
