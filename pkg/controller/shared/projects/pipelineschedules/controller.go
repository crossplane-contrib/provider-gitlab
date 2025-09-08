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

package pipelineschedules

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiCluster "github.com/crossplane-contrib/provider-gitlab/apis/cluster/projects/v1alpha1"
	apiNamespaced "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	sharedProjectsV1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/shared/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
)

const (
	ErrNotPipelineSchedule            = "managed resource is not a PipelineSchedule"
	ErrIDNotAnInt                     = "managed resource ID is not an integer"
	ErrNoProjectID                    = "managed resource mising project ID value"
	ErrExternalNameMissing            = "managed resource missing external name value"
	ErrGetPipelineSchedule            = "failed to get PipelineSchedule"
	ErrCreatePipelineSchedule         = "failed to create PipelineSchedule"
	ErrUpdatePipelineSchedule         = "failed to update PipelineSchedule"
	ErrDeletePipelineSchedule         = "failed to delete PipelineSchedule"
	ErrCreatePipelineScheduleVariable = "failed to create PipelineScheduleVariable %v"
	ErrUpdatePipelineScheduleVariable = "failed to update PipelineScheduleVariable %v"
	ErrDeletePipelineScheduleVariable = "failed to delete PipelineScheduleVariable %v"
)

type External struct {
	Client projects.PipelineScheduleClient
	Kube   client.Client
}

type options struct {
	externalName  string
	parameters    *sharedProjectsV1alpha1.PipelineScheduleParameters
	atProvider    *sharedProjectsV1alpha1.PipelineScheduleObservation
	setConditions func(c ...common.Condition)
	mg            resource.Managed
}

func (e *External) extractOptions(mg resource.Managed) (*options, error) {
	switch cr := mg.(type) {
	case *apiCluster.PipelineSchedule:
		return &options{
			externalName:  meta.GetExternalName(cr),
			parameters:    &cr.Spec.ForProvider.PipelineScheduleParameters,
			atProvider:    &cr.Status.AtProvider,
			setConditions: cr.SetConditions,
			mg:            mg,
		}, nil
	case *apiNamespaced.PipelineSchedule:
		return &options{
			externalName:  meta.GetExternalName(cr),
			parameters:    &cr.Spec.ForProvider.PipelineScheduleParameters,
			atProvider:    &cr.Status.AtProvider,
			setConditions: cr.SetConditions,
			mg:            mg,
		}, nil
	default:
		return nil, errors.New(ErrNotPipelineSchedule)
	}
}

func (e *External) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if opts.externalName == "" {
		return managed.ExternalObservation{}, nil
	}

	id, err := strconv.Atoi(opts.externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.New(ErrIDNotAnInt)
	}

	if opts.parameters.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(ErrNoProjectID)
	}

	ps, res, err := e.Client.GetPipelineSchedule(*opts.parameters.ProjectID, id)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, ErrGetPipelineSchedule)
	}

	current := opts.parameters.DeepCopy()
	e.lateInitialize(opts.parameters, ps)
	e.generateObservation(opts.atProvider, ps)
	opts.setConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        e.isUpToDate(opts.parameters, ps),
		ResourceLateInitialized: !cmp.Equal(current, opts.parameters),
	}, nil
}

func (e *External) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	if opts.parameters.ProjectID == nil {
		return managed.ExternalCreation{}, errors.New(ErrNoProjectID)
	}

	opt := &gitlab.CreatePipelineScheduleOptions{
		Description:  &opts.parameters.Description,
		Ref:          &opts.parameters.Ref,
		Cron:         &opts.parameters.Cron,
		CronTimezone: opts.parameters.CronTimezone,
		Active:       opts.parameters.Active,
	}

	ps, _, err := e.Client.CreatePipelineSchedule(*opts.parameters.ProjectID, opt)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, ErrCreatePipelineSchedule)
	}

	meta.SetExternalName(opts.mg, strconv.Itoa(ps.ID))

	for _, v := range opts.parameters.Variables {
		opt := &gitlab.CreatePipelineScheduleVariableOptions{
			Key:          &v.Key,
			Value:        &v.Value,
			VariableType: (*gitlab.VariableTypeValue)(v.VariableType),
		}
		_, _, err := e.Client.CreatePipelineScheduleVariable(*opts.parameters.ProjectID, ps.ID, opt)
		if err != nil {
			return managed.ExternalCreation{}, errors.Wrapf(err, ErrCreatePipelineScheduleVariable, v)
		}
	}

	return managed.ExternalCreation{}, nil
}

func (e *External) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) { //nolint:gocyclo
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	if opts.externalName == "" {
		return managed.ExternalUpdate{}, errors.New(ErrExternalNameMissing)
	}

	id, err := strconv.Atoi(opts.externalName)
	if err != nil {
		return managed.ExternalUpdate{}, errors.New(ErrIDNotAnInt)
	}

	if opts.parameters.ProjectID == nil {
		return managed.ExternalUpdate{}, errors.New(ErrNoProjectID)
	}

	opt := &gitlab.EditPipelineScheduleOptions{
		Description:  &opts.parameters.Description,
		Ref:          &opts.parameters.Ref,
		Cron:         &opts.parameters.Cron,
		CronTimezone: opts.parameters.CronTimezone,
		Active:       opts.parameters.Active,
	}

	ps, _, err := e.Client.EditPipelineSchedule(*opts.parameters.ProjectID, id, opt)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, ErrUpdatePipelineSchedule)
	}

	if e.hasVariables(opts.parameters, ps) {
		ps, _, err := e.Client.GetPipelineSchedule(*opts.parameters.ProjectID, id)
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, ErrGetPipelineSchedule)
		}
		for _, v := range opts.parameters.Variables {
			if e.notSaved(v, ps.Variables) {
				opt := &gitlab.CreatePipelineScheduleVariableOptions{
					Key:          &v.Key,
					Value:        &v.Value,
					VariableType: (*gitlab.VariableTypeValue)(v.VariableType),
				}
				_, _, err := e.Client.CreatePipelineScheduleVariable(*opts.parameters.ProjectID, ps.ID, opt)
				if err != nil {
					return managed.ExternalUpdate{}, errors.Wrapf(err, ErrCreatePipelineScheduleVariable, v)
				}
			}
			if e.notUpdated(v, ps.Variables) {
				opt := &gitlab.EditPipelineScheduleVariableOptions{
					Value:        &v.Value,
					VariableType: (*gitlab.VariableTypeValue)(v.VariableType),
				}
				_, _, err := e.Client.EditPipelineScheduleVariable(*opts.parameters.ProjectID, ps.ID, v.Key, opt)
				if err != nil {
					return managed.ExternalUpdate{}, errors.Wrapf(err, ErrUpdatePipelineScheduleVariable, v)
				}
			}
		}
		for _, v := range ps.Variables {
			if e.notDeleted(v, opts.parameters.Variables) {
				_, _, err := e.Client.DeletePipelineScheduleVariable(*opts.parameters.ProjectID, ps.ID, v.Key)
				if err != nil {
					return managed.ExternalUpdate{}, errors.Wrapf(err, ErrDeletePipelineScheduleVariable, v)
				}
			}
		}
	}

	return managed.ExternalUpdate{}, errors.Wrap(err, ErrUpdatePipelineSchedule)
}

func (e *External) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalDelete{}, err
	}

	if opts.parameters.ProjectID == nil {
		return managed.ExternalDelete{}, errors.New(ErrNoProjectID)
	}

	id, err := strconv.Atoi(opts.externalName)
	if err != nil {
		return managed.ExternalDelete{}, errors.New(ErrIDNotAnInt)
	}

	_, err = e.Client.DeletePipelineSchedule(*opts.parameters.ProjectID, id)

	return managed.ExternalDelete{}, errors.Wrap(err, ErrDeletePipelineSchedule)
}

func (e *External) lateInitialize(cr *sharedProjectsV1alpha1.PipelineScheduleParameters, ps *gitlab.PipelineSchedule) {
	if ps == nil {
		return
	}
	if cr.CronTimezone == nil && ps.CronTimezone != "" {
		cr.CronTimezone = &ps.CronTimezone
	}
	if cr.Active == nil {
		cr.Active = &ps.Active
	}
	if cr.Variables == nil && len(ps.Variables) > 0 {
		varr := make([]sharedProjectsV1alpha1.PipelineVariable, len(ps.Variables))
		for i, vv := range ps.Variables {
			varr[i] = sharedProjectsV1alpha1.PipelineVariable{
				Key:          vv.Key,
				Value:        vv.Value,
				VariableType: (*string)(&vv.VariableType),
			}
		}
		cr.Variables = varr
	}
}

func (e *External) isUpToDate(cr *sharedProjectsV1alpha1.PipelineScheduleParameters, ps *gitlab.PipelineSchedule) bool {
	if cr.Cron != ps.Cron {
		return false
	}
	if cr.Description != ps.Description {
		return false
	}
	if !clients.IsStringEqualToStringPtr(cr.CronTimezone, ps.CronTimezone) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(cr.Active, ps.Active) {
		return false
	}
	if !e.isVariablesUpToDate(cr.Variables, ps.Variables) {
		return false
	}

	return true
}

func (e *External) isVariablesUpToDate(crv []sharedProjectsV1alpha1.PipelineVariable, inv []*gitlab.PipelineVariable) bool {
	if len(crv) != len(inv) {
		return false
	}
	for _, v := range crv {
		if e.notSaved(v, inv) {
			return false
		}
	}
	for _, v := range inv {
		if e.notDeleted(v, crv) {
			return false
		}
	}
	return true
}

func (e *External) notDeleted(inv *gitlab.PipelineVariable, crvArr []sharedProjectsV1alpha1.PipelineVariable) bool {
	for _, v := range crvArr {
		if inv.Key == v.Key {
			return false
		}
	}
	return true
}

func (e *External) notSaved(crv sharedProjectsV1alpha1.PipelineVariable, invArr []*gitlab.PipelineVariable) bool {
	for _, v := range invArr {
		if crv.Key == v.Key {
			return false
		}
	}
	return true
}

func (e *External) notUpdated(crv sharedProjectsV1alpha1.PipelineVariable, invArr []*gitlab.PipelineVariable) bool {
	victim := gitlab.PipelineVariable{
		Key:   crv.Key,
		Value: crv.Value,
	}

	if crv.VariableType != nil {
		victim.VariableType = gitlab.VariableTypeValue(*crv.VariableType)
	}

	for _, v := range invArr {
		if victim.Key == v.Key && victim != *v {
			return true
		}
	}

	return false
}

func (e *External) generateObservation(o *sharedProjectsV1alpha1.PipelineScheduleObservation, ps *gitlab.PipelineSchedule) {
	*o = sharedProjectsV1alpha1.PipelineScheduleObservation{
		ID:           &ps.ID,
		LastPipeline: (*sharedProjectsV1alpha1.LastPipeline)(ps.LastPipeline),
	}
	if ps.Owner != nil {
		o.Owner = projects.GenerateOwnerObservation(ps.Owner)
	}
	if ps.NextRunAt != nil {
		o.NextRunAt = &metav1.Time{Time: *ps.NextRunAt}
	}
	if ps.CreatedAt != nil {
		o.CreatedAt = &metav1.Time{Time: *ps.CreatedAt}
	}
	if ps.UpdatedAt != nil {
		o.UpdatedAt = &metav1.Time{Time: *ps.UpdatedAt}
	}
}

func (e *External) hasVariables(cr *sharedProjectsV1alpha1.PipelineScheduleParameters, ps *gitlab.PipelineSchedule) bool {
	return cr.Variables != nil || ps.Variables != nil
}

func (e *External) Disconnect(ctx context.Context) error {
	return nil
}
