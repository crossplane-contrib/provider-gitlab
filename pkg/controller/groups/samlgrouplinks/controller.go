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

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/feature"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/statemetrics"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
	secretstoreapi "github.com/crossplane-contrib/provider-gitlab/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups"
	"github.com/crossplane-contrib/provider-gitlab/pkg/features"
)

const (
	errNotSamlGroupLink       = "managed resource is not a SamlGroupLink custom resource"
	errGetFailed              = "cannot get Gitlab SamlGroupLink"
	errCreateFailed           = "cannot create Gitlab SamlGroupLink"
	errDeleteFailed           = "cannot delete Gitlab SamlGroupLink"
	errSamlGroupLinktNotFound = "cannot find Gitlab SamlGroupLink"
	errMissingGroupID         = "missing Spec.ForProvider.GroupID"
	errMissingExternalName    = "external name annotation not found"
)

// SetupSamlGroupLink adds a controller that reconciles samlgrouplinks.
func SetupSamlGroupLink(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.SamlGroupLinkKind)
	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}

	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), secretstoreapi.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabClientFn: groups.NewSamlGroupLinkClient}),
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
		resource.ManagedKind(v1alpha1.SamlGroupLinkGroupVersionKind),
		reconcilerOpts...)

	if err := mgr.Add(statemetrics.NewMRStateRecorder(
		mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &v1alpha1.SamlGroupLinkList{}, o.MetricOptions.PollStateMetricInterval)); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.SamlGroupLink{}).
		Complete(r)
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg clients.Config) groups.SamlGroupLinkClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.SamlGroupLink)
	if !ok {
		return nil, errors.New(errNotSamlGroupLink)
	}

	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

type external struct {
	kube   client.Client
	client groups.SamlGroupLinkClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.SamlGroupLink)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotSamlGroupLink)
	}

	samlGroupName := meta.GetExternalName(cr)

	if samlGroupName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	if cr.Spec.ForProvider.GroupID == nil {
		return managed.ExternalObservation{}, errors.New(errMissingGroupID)
	}

	groupLink, _, err := e.client.GetGroupSAMLLink(*cr.Spec.ForProvider.GroupID, samlGroupName)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(groups.IsErrorSamlGroupLinkNotFound, err), errGetFailed)
	}

	cr.Status.AtProvider = groups.GenerateAddSamlGroupLinkObservation(groupLink)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isSamlGroupLinkUpToDate(&cr.Spec.ForProvider, groupLink),
		ResourceLateInitialized: false,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.SamlGroupLink)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotSamlGroupLink)
	}

	if cr.Spec.ForProvider.GroupID == nil {
		return managed.ExternalCreation{}, errors.New(errMissingGroupID)
	}

	samlGroupLink, _, err := e.client.AddGroupSAMLLink(
		*cr.Spec.ForProvider.GroupID,
		groups.GenerateAddSamlGroupLinkOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	meta.SetExternalName(cr, samlGroupLink.Name)

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// not able to update SamlGroupLink
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.SamlGroupLink)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotSamlGroupLink)
	}
	if cr.Spec.ForProvider.GroupID == nil {
		return managed.ExternalDelete{}, errors.New(errMissingGroupID)
	}

	samlGroupName := meta.GetExternalName(cr)

	if samlGroupName == "" {
		return managed.ExternalDelete{}, errors.New(errMissingExternalName)
	}

	_, err := e.client.DeleteGroupSAMLLink(
		*cr.Spec.ForProvider.GroupID,
		samlGroupName,
		nil,
		gitlab.WithContext(ctx),
	)
	return managed.ExternalDelete{}, errors.Wrap(err, errDeleteFailed)
}

func (e *external) Disconnect(ctx context.Context) error {
	// Disconnect is not implemented as it is a new method required by the SDK
	return nil
}

func isSamlGroupLinkUpToDate(p *v1alpha1.SamlGroupLinkParameters, g *gitlab.SAMLGroupLink) bool {
	if !cmp.Equal(int(p.AccessLevel), int(g.AccessLevel)) {
		return false
	}

	if !cmp.Equal(*p.Name, g.Name) {
		return false
	}
	return true
}
