/*
Copyright 2020 The Crossplane Authors.

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

package accesstokens

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
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/groups"
)

const (
	errNotAccessToken      = "managed resource is not a Gitlab accesstoken custom resource"
	errExternalNameNotInt  = "custom resource external name is not an integer"
	errFailedParseID       = "cannot parse Access Token ID to int"
	errCreateFailed        = "cannot create Gitlab accesstoken"
	errDeleteFailed        = "cannot delete Gitlab accesstoken"
	errAccessTokenNotFound = "cannot find Gitlab accesstoken"
	errMissingGroupID      = "missing Spec.ForProvider.GroupID"
)

// SetupAccessToken adds a controller that reconciles GroupAccessTokens.
func SetupAccessToken(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.AccessTokenGroupKind)

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabClientFn: groups.NewAccessTokenClient}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}

	if o.Features.Enabled(feature.EnableBetaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.AccessTokenGroupVersionKind),
		reconcilerOpts...)

	if err := mgr.Add(statemetrics.NewMRStateRecorder(
		mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &v1alpha1.AccessTokenList{}, o.MetricOptions.PollStateMetricInterval)); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.AccessToken{}).
		Complete(r)
}

// SetupAccessTokenGated adds a controller with CRD gate support.
func SetupAccessTokenGated(mgr ctrl.Manager, o controller.Options) error {
	o.Gate.Register(func() {
		if err := SetupAccessToken(mgr, o); err != nil {
			mgr.GetLogger().Error(err, "unable to setup reconciler", "gvk", v1alpha1.AccessTokenGroupVersionKind.String())
		}
	}, v1alpha1.AccessTokenGroupVersionKind)
	return nil
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg common.Config) groups.AccessTokenClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.AccessToken)
	if !ok {
		return nil, errors.New(errNotAccessToken)
	}
	cfg, err := common.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

type external struct {
	kube   client.Client
	client groups.AccessTokenClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.AccessToken)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotAccessToken)
	}

	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalObservation{}, nil
	}

	accessTokenID, err := strconv.Atoi(externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errFailedParseID)
	}

	if cr.Spec.ForProvider.GroupID == nil {
		return managed.ExternalObservation{}, errors.New(errMissingGroupID)
	}

	at, res, err := e.client.GetGroupAccessToken(*cr.Spec.ForProvider.GroupID, int64(accessTokenID))
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{
				ResourceExists: false,
			}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errAccessTokenNotFound)
	}

	cr.Status.AtProvider = groups.GenerateGroupAccessTokenObservation(at)

	if groups.ShouldRotateAccessToken(&cr.Spec.ForProvider, at) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        true,
		ResourceLateInitialized: false,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.AccessToken)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotAccessToken)
	}

	if cr.Spec.ForProvider.GroupID == nil {
		return managed.ExternalCreation{}, errors.New(errMissingGroupID)
	}

	// If a token ID is already set as the external name, try rotating it first.
	// Rotation atomically revokes the old token and issues a new one.
	if existingID, err := strconv.ParseInt(meta.GetExternalName(cr), 10, 64); err == nil && existingID > 0 {
		at, _, rotErr := e.client.RotateGroupAccessToken(
			*cr.Spec.ForProvider.GroupID,
			existingID,
			groups.GenerateRotateGroupAccessTokenOptions(&cr.Spec.ForProvider),
			gitlab.WithContext(ctx),
		)
		if rotErr == nil {
			meta.SetExternalName(cr, strconv.FormatInt(at.ID, 10))
			return managed.ExternalCreation{
				ConnectionDetails: managed.ConnectionDetails{"token": []byte(at.Token)},
			}, nil
		}
		// Rotation failed (e.g. token was already revoked externally); fall through to fresh create.
	}

	at, _, err := e.client.CreateGroupAccessToken(
		*cr.Spec.ForProvider.GroupID,
		groups.GenerateCreateGroupAccessTokenOptions(cr.Name, &cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	meta.SetExternalName(cr, strconv.FormatInt(at.ID, 10))
	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{"token": []byte(at.Token)},
	}, nil
}

// Update is currently a no-op as the only updatable field is expiration, which is handled by rotation in Observe.
func (e *external) Update(_ context.Context, _ resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.AccessToken)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotAccessToken)
	}

	accessTokenID, err := strconv.Atoi(meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalDelete{}, errors.New(errExternalNameNotInt)
	}

	if cr.Spec.ForProvider.GroupID == nil {
		return managed.ExternalDelete{}, errors.New(errMissingGroupID)
	}
	_, err = e.client.RevokeGroupAccessToken(
		*cr.Spec.ForProvider.GroupID,
		int64(accessTokenID),
		gitlab.WithContext(ctx),
	)

	return managed.ExternalDelete{}, errors.Wrap(err, errDeleteFailed)
}

func (e *external) Disconnect(ctx context.Context) error {
	// Disconnect is not implemented as it is a new method required by the SDK
	return nil
}
