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

package ldapgrouplinks

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
	gitlab "gitlab.com/gitlab-org/api/client-go"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
	secretstoreapi "github.com/crossplane-contrib/provider-gitlab/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups"
	"github.com/crossplane-contrib/provider-gitlab/pkg/features"
)

const (
	errNotLdapGroupLink       = "managed resource is not a LdapGroupLink custom resource"
	errGetFailed              = "cannot get Gitlab LdapGroupLink"
	errCreateFailed           = "cannot create Gitlab LdapGroupLink"
	errDeleteFailed           = "cannot delete Gitlab LdapGroupLink"
	errLdapGroupLinktNotFound = "cannot find Gitlab LdapGroupLink"
	errMissingGroupID         = "missing Spec.ForProvider.GroupID"
	errMissingExternalName    = "external name annotation not found"
)

// SetupLdapGroupLink adds a controller that reconciles ldapgrouplinks.
func SetupLdapGroupLink(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.LdapGroupLinkGroupKind)
	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}

	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), secretstoreapi.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabClientFn: groups.NewLdapGroupLinkClient}),
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
		resource.ManagedKind(v1alpha1.LdapGroupLinkGroupVersionKind),
		reconcilerOpts...)

	if err := mgr.Add(statemetrics.NewMRStateRecorder(
		mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &v1alpha1.LdapGroupLinkList{}, o.MetricOptions.PollStateMetricInterval)); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.LdapGroupLink{}).
		Complete(r)
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg clients.Config) groups.LdapGroupLinkClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.LdapGroupLink)
	if !ok {
		return nil, errors.New(errNotLdapGroupLink)
	}

	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

type external struct {
	kube   client.Client
	client groups.LdapGroupLinkClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.LdapGroupLink)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotLdapGroupLink)
	}

	ldapCN := meta.GetExternalName(cr)

	if ldapCN == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	if cr.Spec.ForProvider.GroupID == nil {
		return managed.ExternalObservation{}, errors.New(errMissingGroupID)
	}

	groupLinks, _, err := e.client.ListGroupLDAPLinks(*cr.Spec.ForProvider.GroupID)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(groups.IsErrorLdapGroupLinkNotFound, err), errGetFailed)
	}

	found := false
	groupLink := &gitlab.LDAPGroupLink{}
	for _, gl := range groupLinks {
		if gl.CN == ldapCN {
			groupLink = gl
			found = true
			break
		}
	}

	if !found {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	cr.Status.AtProvider = groups.GenerateAddLdapGroupLinkObservation(groupLink)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isLdapGroupLinkUpToDate(&cr.Spec.ForProvider, groupLink),
		ResourceLateInitialized: false,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.LdapGroupLink)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotLdapGroupLink)
	}

	if cr.Spec.ForProvider.GroupID == nil {
		return managed.ExternalCreation{}, errors.New(errMissingGroupID)
	}

	ldapGroupLink, _, err := e.client.AddGroupLDAPLink(
		*cr.Spec.ForProvider.GroupID,
		groups.GenerateAddLdapGroupLinkOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	meta.SetExternalName(cr, ldapGroupLink.CN)

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// recreate object
	_, errDelete := e.Delete(ctx, mg)
	if errDelete != nil {
		return managed.ExternalUpdate{}, errDelete
	}

	_, errCreate := e.Create(ctx, mg)
	if errCreate != nil {
		return managed.ExternalUpdate{}, errCreate
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.LdapGroupLink)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotLdapGroupLink)
	}
	if cr.Spec.ForProvider.GroupID == nil {
		return managed.ExternalDelete{}, errors.New(errMissingGroupID)
	}

	ldapCN := meta.GetExternalName(cr)

	if ldapCN == "" {
		return managed.ExternalDelete{}, errors.New(errMissingExternalName)
	}

	_, err := e.client.DeleteGroupLDAPLinkWithCNOrFilter(
		*cr.Spec.ForProvider.GroupID,
		groups.GenerateDeleteGroupLDAPLinkWithCNOrFilterOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)

	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteFailed)
	}

	return managed.ExternalDelete{}, nil
}

func (e *external) Disconnect(ctx context.Context) error {
	// Disconnect is not implemented as it is a new method required by the SDK
	return nil
}

func isLdapGroupLinkUpToDate(p *v1alpha1.LdapGroupLinkParameters, g *gitlab.LDAPGroupLink) bool {
	if !cmp.Equal(p.CN, g.CN) {
		return false
	}

	if !cmp.Equal(int(p.GroupAccess), int(g.GroupAccess)) {
		return false
	}

	return true
}
