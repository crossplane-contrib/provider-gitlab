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

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/feature"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/statemetrics"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	secretstoreapi "github.com/crossplane-contrib/provider-gitlab/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/users"
	"github.com/crossplane-contrib/provider-gitlab/pkg/features"
)

const (
	errNotMember        = "managed resource is not a Gitlab Project Member custom resource"
	errCreateFailed     = "cannot create Gitlab Approval Rule"
	errUpdateFailed     = "cannot update Gitlab Approval Rule"
	errDeleteFailed     = "cannot delete Gitlab Approval Rule"
	errObserveFailed    = "cannot observe Gitlab Approval Rule"
	errProjectIDMissing = "ProjectID is missing"
)

// SetupRules adds a controller that reconciles Approval Rules.
func SetupRules(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.ApprovalRuleKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), secretstoreapi.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{
			kube:              mgr.GetClient(),
			newGitlabClientFn: projects.NewMemberClient,
			newUserClientFn:   users.NewUserClient,
		}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(feature.EnableBetaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.MemberGroupVersionKind),
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
		return nil, errors.New(errNotMember)
	}
	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg), userClient: c.newUserClientFn(*cfg)}, nil
}

type external struct {
	kube       client.Client
	client     projects.MemberClient
	userClient users.UserClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.ApprovalRule)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotMember)
	}
	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(errProjectIDMissing)
	}

	userID, err := cr.Spec.ForProvider.UserID, error(nil)
	if cr.Spec.ForProvider.UserID == nil {
		if cr.Spec.ForProvider.UserName == nil {
			return managed.ExternalObservation{}, errors.New(errUserInfoMissing)
		}
		userID, err = users.GetUserID(e.userClient, *cr.Spec.ForProvider.UserName)
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errFetchFailed)
		}
	}
	cr.Spec.ForProvider.UserID = userID

	projectMember, res, err := e.client.GetProjectMember(
		*cr.Spec.ForProvider.ProjectID,
		*cr.Spec.ForProvider.UserID,
	)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errObserveFailed)
	}

	cr.Status.AtProvider = projects.GenerateMemberObservation(projectMember)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isMemberUpToDate(&cr.Spec.ForProvider, projectMember),
		ResourceLateInitialized: false,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ApprovalRule)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotMember)
	}
	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalCreation{}, errors.New(errProjectIDMissing)
	}

	_, _, err := e.client.AddProjectMember(
		*cr.Spec.ForProvider.ProjectID,
		projects.GenerateAddMemberOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.ApprovalRule)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotMember)
	}
	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalUpdate{}, errors.New(errProjectIDMissing)
	}
	if cr.Spec.ForProvider.UserID == nil {
		return managed.ExternalUpdate{}, errors.New(errUserInfoMissing)
	}

	_, _, err := e.client.EditProjectMember(
		*cr.Spec.ForProvider.ProjectID,
		*cr.Spec.ForProvider.UserID,
		projects.GenerateEditMemberOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.ApprovalRule)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotMember)
	}
	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalDelete{}, errors.New(errProjectIDMissing)
	}
	if cr.Spec.ForProvider.UserID == nil {
		return managed.ExternalDelete{}, errors.New(errUserInfoMissing)
	}

	_, err := e.client.DeleteProjectMember(
		*cr.Spec.ForProvider.ProjectID,
		*cr.Spec.ForProvider.UserID,
		gitlab.WithContext(ctx),
	)
	return managed.ExternalDelete{}, errors.Wrap(err, errDeleteFailed)
}

func (e *external) Disconnect(ctx context.Context) error {
	// Disconnect is not implemented as it is a new method required by the SDK
	return nil
}

// isMemberUpToDate checks whether there is a change in any of the modifiable fields.
func isMemberUpToDate(p *v1alpha1.ApprovalRuleParameters, g *gitlab.ProjectMember) bool {
	if !cmp.Equal(int(p.AccessLevel), int(g.AccessLevel)) {
		return false
	}

	if !cmp.Equal(derefString(p.ExpiresAt), isoTimeToString(g.ExpiresAt)) {
		return false
	}

	return true
}

func derefString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// isoTimeToString checks if given return date in format iso8601 otherwise empty string.
func isoTimeToString(i interface{}) string {
	v, ok := i.(*gitlab.ISOTime)
	if ok && v != nil {
		return v.String()
	}
	return ""
}
