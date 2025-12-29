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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects"
)

const (
	errNotPipelineSchedule            = "managed resource is not a PipelineSchedule"
	errIDNotAnInt                     = "managed resource ID is not an integer"
	errNoProjectID                    = "managed resource mising project ID value"
	errExternalNameMissing            = "managed resource missing external name value"
	errGetPipelineSchedule            = "failed to get PipelineSchedule"
	errCreatePipelineSchedule         = "failed to create PipelineSchedule"
	errUpdatePipelineSchedule         = "failed to update PipelineSchedule"
	errDeletePipelineSchedule         = "failed to delete PipelineSchedule"
	errCreatePipelineScheduleVariable = "failed to create PipelineScheduleVariable %v"
	errUpdatePipelineScheduleVariable = "failed to update PipelineScheduleVariable %v"
	errDeletePipelineScheduleVariable = "failed to delete PipelineScheduleVariable %v"
)

// SetupPipelineSchedule adds a controller that reconciles PipelineSchedule.
func SetupPipelineSchedule(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.PipelineScheduleGroupKind)

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabClientFn: newPipelineScheduleClient}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}

	if o.Features.Enabled(feature.EnableBetaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.PipelineScheduleGroupVersionKind),
		reconcilerOpts...)

	if err := mgr.Add(statemetrics.NewMRStateRecorder(
		mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &v1alpha1.PipelineScheduleList{}, o.MetricOptions.PollStateMetricInterval)); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.PipelineSchedule{}).
		Complete(r)
}

// SetupPipelineScheduleGated adds a controller with CRD gate support.
func SetupPipelineScheduleGated(mgr ctrl.Manager, o controller.Options) error {
	o.Gate.Register(func() {
		if err := SetupPipelineSchedule(mgr, o); err != nil {
			mgr.GetLogger().Error(err, "unable to setup reconciler", "gvk", v1alpha1.PipelineScheduleGroupVersionKind.String())
		}
	}, v1alpha1.PipelineScheduleGroupVersionKind)
	return nil
}

type external struct {
	kube   client.Client
	client projects.PipelineScheduleClient
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(c common.Config) projects.PipelineScheduleClient
}

// Connect implements managed.ExternalConnecter.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.PipelineSchedule)
	if !ok {
		return nil, errors.New(errNotPipelineSchedule)
	}

	conf, err := common.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}

	return &external{
		kube:   c.kube,
		client: c.newGitlabClientFn(*conf),
	}, nil
}

// Observe implements managed.ExternalClient.
func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.PipelineSchedule)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotPipelineSchedule)
	}

	idstr := meta.GetExternalName(cr)
	if idstr == "" {
		return managed.ExternalObservation{}, nil
	}

	id, err := strconv.Atoi(idstr)
	if err != nil {
		return managed.ExternalObservation{}, errors.New(errIDNotAnInt)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(errNoProjectID)
	}

	ps, res, err := e.client.GetPipelineSchedule(*cr.Spec.ForProvider.ProjectID, int64(id))
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetPipelineSchedule)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	lateInitialize(&cr.Spec.ForProvider, ps)
	generateObservation(cr, ps)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isUpToDate(cr, ps),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

// Create implements managed.ExternalClient.
func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.PipelineSchedule)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotPipelineSchedule)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalCreation{}, errors.New(errNoProjectID)
	}

	opt := &gitlab.CreatePipelineScheduleOptions{
		Description:  &cr.Spec.ForProvider.Description,
		Ref:          &cr.Spec.ForProvider.Ref,
		Cron:         &cr.Spec.ForProvider.Cron,
		CronTimezone: cr.Spec.ForProvider.CronTimezone,
		Active:       cr.Spec.ForProvider.Active,
	}

	ps, _, err := e.client.CreatePipelineSchedule(*cr.Spec.ForProvider.ProjectID, opt)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreatePipelineSchedule)
	}

<<<<<<< HEAD
	meta.SetExternalName(cr, strconv.FormatInt(ps.ID, 10))
=======
	meta.SetExternalName(cr, strconv.FormatInt(int64(ps.ID), 10))
>>>>>>> 77c306d (feat: migrate CRD types from *int to *int64)

	for _, v := range cr.Spec.ForProvider.Variables {
		opt := &gitlab.CreatePipelineScheduleVariableOptions{
			Key:          &v.Key,
			Value:        &v.Value,
			VariableType: (*gitlab.VariableTypeValue)(v.VariableType),
		}
		_, _, err := e.client.CreatePipelineScheduleVariable(
			*cr.Spec.ForProvider.ProjectID,
			int64(ps.ID),
			opt,
		)
		if err != nil {
			return managed.ExternalCreation{}, errors.Wrapf(err, errCreatePipelineScheduleVariable, v)
		}
	}

	return managed.ExternalCreation{}, nil
}

// Update implements managed.ExternalClient.
func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) { //nolint:gocyclo
	cr, ok := mg.(*v1alpha1.PipelineSchedule)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotPipelineSchedule)
	}
	extName := meta.GetExternalName(cr)
	if extName == "" {
		return managed.ExternalUpdate{}, errors.New(errExternalNameMissing)
	}
	id, err := strconv.Atoi(extName)
	if err != nil {
		return managed.ExternalUpdate{}, errors.New(errIDNotAnInt)
	}
	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalUpdate{}, errors.New(errNoProjectID)
	}

	opt := &gitlab.EditPipelineScheduleOptions{
		Description:  &cr.Spec.ForProvider.Description,
		Ref:          &cr.Spec.ForProvider.Ref,
		Cron:         &cr.Spec.ForProvider.Cron,
		CronTimezone: cr.Spec.ForProvider.CronTimezone,
		Active:       cr.Spec.ForProvider.Active,
	}

	ps, _, err := e.client.EditPipelineSchedule(
		*cr.Spec.ForProvider.ProjectID,
		int64(id),
		opt,
	)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdatePipelineSchedule)
	}

	if hasVariables(cr, ps) {
		ps, _, err := e.client.GetPipelineSchedule(*cr.Spec.ForProvider.ProjectID, int64(id))
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errGetPipelineSchedule)
		}
		for _, v := range cr.Spec.ForProvider.Variables {
			if notSaved(v, ps.Variables) {
				opt := &gitlab.CreatePipelineScheduleVariableOptions{
					Key:          &v.Key,
					Value:        &v.Value,
					VariableType: (*gitlab.VariableTypeValue)(v.VariableType),
				}
				_, _, err := e.client.CreatePipelineScheduleVariable(
					*cr.Spec.ForProvider.ProjectID,
					int64(ps.ID),
					opt,
				)
				if err != nil {
					return managed.ExternalUpdate{}, errors.Wrapf(err, errCreatePipelineScheduleVariable, v)
				}
			}
			if notUpdated(v, ps.Variables) {
				opt := &gitlab.EditPipelineScheduleVariableOptions{
					Value:        &v.Value,
					VariableType: (*gitlab.VariableTypeValue)(v.VariableType),
				}
				_, _, err := e.client.EditPipelineScheduleVariable(
					*cr.Spec.ForProvider.ProjectID,
					int64(ps.ID),
					v.Key,
					opt,
				)
				if err != nil {
					return managed.ExternalUpdate{}, errors.Wrapf(err, errUpdatePipelineScheduleVariable, v)
				}
			}
		}
		for _, v := range ps.Variables {
			if notDeleted(v, cr.Spec.ForProvider.Variables) {
				_, _, err := e.client.DeletePipelineScheduleVariable(
					*cr.Spec.ForProvider.ProjectID,
					int64(ps.ID),
					v.Key,
				)
				if err != nil {
					return managed.ExternalUpdate{}, errors.Wrapf(err, errDeletePipelineScheduleVariable, v)
				}
			}
		}
	}

	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdatePipelineSchedule)
}

// Delete implements managed.ExternalClient.
func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.PipelineSchedule)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotPipelineSchedule)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalDelete{}, errors.New(errNoProjectID)
	}

	id, err := strconv.Atoi(meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalDelete{}, errors.New(errIDNotAnInt)
	}

	_, err = e.client.DeletePipelineSchedule(
		*cr.Spec.ForProvider.ProjectID,
		int64(id),
	)

	return managed.ExternalDelete{}, errors.Wrap(err, errDeletePipelineSchedule)
}

func (e *external) Disconnect(ctx context.Context) error {
	// Disconnect is not implemented as it is a new method required by the SDK
	return nil
}

func newPipelineScheduleClient(c common.Config) projects.PipelineScheduleClient {
	return common.NewClient(c).PipelineSchedules
}

func lateInitialize(cr *v1alpha1.PipelineScheduleParameters, ps *gitlab.PipelineSchedule) {
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
		varr := make([]v1alpha1.PipelineVariable, len(ps.Variables))
		for i, vv := range ps.Variables {
			varr[i] = v1alpha1.PipelineVariable{
				Key:          vv.Key,
				Value:        vv.Value,
				VariableType: (*string)(&vv.VariableType),
			}
		}
		cr.Variables = varr
	}
}

func isUpToDate(cr *v1alpha1.PipelineSchedule, ps *gitlab.PipelineSchedule) bool {
	if cr.Spec.ForProvider.Cron != ps.Cron {
		return false
	}
	if cr.Spec.ForProvider.Description != ps.Description {
		return false
	}
	if !clients.IsStringEqualToStringPtr(cr.Spec.ForProvider.CronTimezone, ps.CronTimezone) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(cr.Spec.ForProvider.Active, ps.Active) {
		return false
	}
	if !isVariablesUpToDate(cr.Spec.ForProvider.Variables, ps.Variables) {
		return false
	}

	return true
}

func isVariablesUpToDate(crv []v1alpha1.PipelineVariable, inv []*gitlab.PipelineVariable) bool {
	if len(crv) != len(inv) {
		return false
	}
	for _, v := range crv {
		if notSaved(v, inv) {
			return false
		}
	}
	for _, v := range inv {
		if notDeleted(v, crv) {
			return false
		}
	}
	return true
}

func notDeleted(inv *gitlab.PipelineVariable, crvArr []v1alpha1.PipelineVariable) bool {
	for _, v := range crvArr {
		if inv.Key == v.Key {
			return false
		}
	}
	return true
}

func notSaved(crv v1alpha1.PipelineVariable, invArr []*gitlab.PipelineVariable) bool {
	for _, v := range invArr {
		if crv.Key == v.Key {
			return false
		}
	}
	return true
}

func notUpdated(crv v1alpha1.PipelineVariable, invArr []*gitlab.PipelineVariable) bool {
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

func generateObservation(cr *v1alpha1.PipelineSchedule, ps *gitlab.PipelineSchedule) {
	id64 := int64(ps.ID)
	o := v1alpha1.PipelineScheduleObservation{
<<<<<<< HEAD
		ID:           &ps.ID,
=======
		ID:           common.Int64ToIntPtr(&id64),
>>>>>>> 77c306d (feat: migrate CRD types from *int to *int64)
		LastPipeline: convertLastPipeline(ps.LastPipeline),
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
	cr.Status.AtProvider = o
}

func hasVariables(cr *v1alpha1.PipelineSchedule, ps *gitlab.PipelineSchedule) bool {
	return cr.Spec.ForProvider.Variables != nil || ps.Variables != nil
}

func convertLastPipeline(lp *gitlab.LastPipeline) *v1alpha1.LastPipeline {
	if lp == nil {
		return nil
	}
	return &v1alpha1.LastPipeline{
<<<<<<< HEAD
		ID:     lp.ID,
=======
		ID:     int(lp.ID),
>>>>>>> 77c306d (feat: migrate CRD types from *int to *int64)
		SHA:    lp.SHA,
		Ref:    lp.Ref,
		Status: lp.Status,
		WebURL: lp.WebURL,
	}
}
