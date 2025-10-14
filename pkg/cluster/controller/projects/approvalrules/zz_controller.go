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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/cluster/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/cluster/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/cluster/clients/projects"
)

const (
	errNotApprovalRule  = "managed resource is not a Gitlab Project Approval Rule custom resource"
	errCreateFailed     = "cannot create Gitlab Approval Rule"
	errUpdateFailed     = "cannot update Gitlab Approval Rule"
	errDeleteFailed     = "cannot delete Gitlab Approval Rule"
	errObserveFailed    = "cannot observe Gitlab Approval Rule"
	errProjectIDMissing = "ProjectID is missing"
	errIDnotInt         = "ID is not an integer"
)

// SetupRules adds a controller that reconciles Approval Rules.
func SetupRules(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.ApprovalRuleKind)

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{
			kube:              mgr.GetClient(),
			newGitlabClientFn: projects.NewApprovalRulesClient,
		}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}

	if o.Features.Enabled(feature.EnableBetaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.ApprovalRuleGroupVersionKind),
		reconcilerOpts...)

	if err := mgr.Add(statemetrics.NewMRStateRecorder(
		mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &v1alpha1.ApprovalRuleList{}, o.MetricOptions.PollStateMetricInterval)); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.ApprovalRule{}).
		Complete(r)
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg clients.Config) projects.ApprovalRulesClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.ApprovalRule)
	if !ok {
		return nil, errors.New(errNotApprovalRule)
	}
	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

type external struct {
	kube   client.Client
	client projects.ApprovalRulesClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.ApprovalRule)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotApprovalRule)
	}

	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	id, err := strconv.Atoi(externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.New(errIDnotInt)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(errProjectIDMissing)
	}

	approvalRule, res, err := e.client.GetProjectApprovalRule(*cr.Spec.ForProvider.ProjectID, id)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}

		return managed.ExternalObservation{}, errors.Wrap(err, errObserveFailed)
	}

	current := cr.Spec.ForProvider.DeepCopy()

	cr.Status.AtProvider = v1alpha1.ApprovalRuleObservation{}
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        projects.IsApprovalRuleUpToDate(&cr.Spec.ForProvider, approvalRule),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil

}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ApprovalRule)

	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotApprovalRule)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalCreation{}, errors.New(errProjectIDMissing)
	}

	cr.Status.SetConditions(xpv1.Creating())
	approvalRulesOptions := projects.GenerateCreateApprovalRulesOptions(&cr.Spec.ForProvider)

	rule, _, err := e.client.CreateProjectApprovalRule(*cr.Spec.ForProvider.ProjectID, approvalRulesOptions, gitlab.WithContext(ctx))

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	err = e.updateExternalName(ctx, cr, rule)
	return managed.ExternalCreation{}, err
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.ApprovalRule)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotApprovalRule)
	}
	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalUpdate{}, errors.New(errProjectIDMissing)
	}

	ruleID, err := strconv.Atoi(meta.GetExternalName(cr))

	if err != nil {
		return managed.ExternalUpdate{}, errors.New(errIDnotInt)
	}

	_, _, err = e.client.UpdateProjectApprovalRule(
		*cr.Spec.ForProvider.ProjectID,
		ruleID,
		projects.GenerateUpdateApprovalRulesOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)

	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.ApprovalRule)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotApprovalRule)
	}
	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalDelete{}, errors.New(errProjectIDMissing)
	}

	ruleID, err := strconv.Atoi(meta.GetExternalName(cr))

	if err != nil {
		return managed.ExternalDelete{}, errors.New(errIDnotInt)
	}

	_, err = e.client.DeleteProjectApprovalRule(
		*cr.Spec.ForProvider.ProjectID,
		ruleID,
		gitlab.WithContext(ctx),
	)

	return managed.ExternalDelete{}, errors.Wrap(err, errDeleteFailed)
}

func (e *external) Disconnect(ctx context.Context) error {
	// Disconnect is not implemented as it is a new method required by the SDK
	return nil
}

func (e *external) updateExternalName(ctx context.Context, cr *v1alpha1.ApprovalRule, approvalRule *gitlab.ProjectApprovalRule) error {
	meta.SetExternalName(cr, strconv.Itoa(approvalRule.ID))
	return e.kube.Update(ctx, cr)
}
