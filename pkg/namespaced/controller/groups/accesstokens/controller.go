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
	"net/http"
	"strconv"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/statemetrics"
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	auth "github.com/crossplane-contrib/provider-gitlab/pkg/common/auth"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/groups"
)

const (
	errNotAccessToken      = "managed resource is not a Gitlab accesstoken custom resource"
	errExternalNameNotInt  = "custom resource external name is not an integer"
	errFailedParseID       = "cannot parse Access Token ID to int"
	errCreateFailed        = "cannot create Gitlab accesstoken"
	errRotateFailed        = "cannot rotate Gitlab accesstoken"
	errDeleteFailed        = "cannot delete Gitlab accesstoken"
	errAccessTokenNotFound = "cannot find Gitlab accesstoken"
	errMissingGroupID      = "missing Spec.ForProvider.GroupID"
	errSelfInformFailed    = "cannot read the self access token; the token referenced by the ProviderConfig may be expired or revoked - reseed the credentials secret with a valid token"
	errSelfRotateFailed    = "cannot self-rotate Gitlab access token"
	errSelfNameMismatch    = "the token referenced by the ProviderConfig does not have the name in spec.forProvider.name; refusing to manage it - check that the credentials secret and this resource point at the same token"

	// connectionSecretKeyToken is the connection-secret key under which the
	// token value is published. Self-mode detection requires the ProviderConfig
	// to read this exact key.
	connectionSecretKeyToken = "token"
)

// Condition surfaced on the managed resource to indicate which rotation mode is
// in effect. It is informational (not a Crossplane system condition) and is
// re-asserted on every reconcile.
const (
	// TypeSelfManaged is the condition type reporting self-managed mode.
	TypeSelfManaged v2.ConditionType = "SelfManaged"

	// ReasonProviderConfigReferencesManagedToken is set when the ProviderConfig
	// authenticates with the very token this resource manages (self-mode).
	ReasonProviderConfigReferencesManagedToken v2.ConditionReason = "ProviderConfigReferencesManagedToken"
	// ReasonOwnerManaged is set when the token is managed via group-owner endpoints.
	ReasonOwnerManaged v2.ConditionReason = "OwnerManaged"
)

// selfManagedCondition returns the SelfManaged condition for the given mode.
func selfManagedCondition(self bool) v2.Condition {
	c := v2.Condition{Type: TypeSelfManaged, LastTransitionTime: metav1.Now()}
	if self {
		c.Status = corev1.ConditionTrue
		c.Reason = ReasonProviderConfigReferencesManagedToken
		c.Message = "rotating via the self endpoints; the ProviderConfig authenticates with the token managed by this resource"
		return c
	}
	c.Status = corev1.ConditionFalse
	c.Reason = ReasonOwnerManaged
	return c
}

// SetupAccessToken adds a controller that reconciles GroupAccessTokens.
func SetupAccessToken(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.AccessTokenGroupKind)

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnector(&connector{kube: mgr.GetClient(), newGitlabClientFn: groups.NewAccessTokenClient}),
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
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg), self: isSelfManaged(cfg, cr)}, nil
}

// isSelfManaged reports whether the ProviderConfig this resource uses
// authenticates with the very token this resource writes to its connection
// secret. In that case the provider is acting as the token's bot user itself and
// can only use the self endpoints (self-inform / self-rotate / self-revoke).
//
// Detection requires a PersonalAccessToken credential whose secret reference
// matches the resource's writeConnectionSecretToRef (namespace + name) and the
// published token key.
func isSelfManaged(cfg *common.Config, cr *v1alpha1.AccessToken) bool {
	if cfg == nil || cfg.AuthMethod != auth.PersonalAccessToken {
		return false
	}
	ref := cfg.CredentialsSecretRef
	if ref == nil {
		return false
	}
	w := cr.GetWriteConnectionSecretToReference()
	if w == nil {
		return false
	}
	return ref.Namespace == cr.GetNamespace() &&
		ref.Name == w.Name &&
		ref.Key == connectionSecretKeyToken
}

type external struct {
	kube   client.Client
	client groups.AccessTokenClient
	// self indicates the ProviderConfig authenticates with the token this
	// resource manages, so only the self endpoints are usable.
	self bool
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.AccessToken)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotAccessToken)
	}

	cr.Status.SetConditions(selfManagedCondition(e.self))

	if e.self {
		return e.observeSelf(ctx, cr)
	}
	return e.observeOwner(ctx, cr)
}

// observeSelf observes the token via the self-inform endpoint. The managed token
// is whatever the ProviderConfig authenticates with, so the external name is
// auto-adopted from the response.
func (e *external) observeSelf(ctx context.Context, cr *v1alpha1.AccessToken) (managed.ExternalObservation, error) {
	at, res, err := e.client.GetSelf(gitlab.WithContext(ctx))
	if err != nil {
		// During deletion the self-revoke has already run, so the self
		// endpoints reject the (now revoked) credential with 401/403. That is
		// the expected terminal state: report the resource gone so the
		// finalizer is removed instead of wedging the delete in a retry loop.
		if meta.WasDeleted(cr) && clients.IsResponseUnauthorized(res) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		// A dead self-credential is unrecoverable by the provider: surface a
		// clear error rather than masquerading as "does not exist" (which would
		// push the resource into a doomed self-rotate loop).
		return managed.ExternalObservation{}, errors.Wrap(err, errSelfInformFailed)
	}

	// Safety: in self-managed mode we adopt whatever token the ProviderConfig
	// authenticates with. The self endpoints are not scoped by group, and the
	// self-inform response does not carry the token's group binding, so the only
	// spec-derived identity signal is the token name. Refuse to manage a token
	// whose name does not match, to catch a miswired credentials secret.
	if at.Name != cr.Spec.ForProvider.Name {
		return managed.ExternalObservation{}, errors.New(errSelfNameMismatch)
	}

	meta.SetExternalName(cr, strconv.FormatInt(at.ID, 10))
	cr.Status.AtProvider = groups.GenerateGroupAccessTokenObservationFromPAT(at)

	if cr.Spec.ForProvider.RenewalPeriodDays != nil {
		cr.Status.RenewAt = common.ComputeNextRotation(at.CreatedAt, at.ExpiresAt, cr.Spec.ForProvider.RenewBeforeDays)
	} else {
		cr.Status.RenewAt = nil
	}

	if groups.ShouldRotateAccessTokenFromPAT(&cr.Spec.ForProvider, at) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	cr.Status.SetConditions(v2.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) observeOwner(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
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

	at, res, err := e.client.GetGroupAccessToken(*cr.Spec.ForProvider.GroupID, int64(accessTokenID), gitlab.WithContext(ctx))
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{
				ResourceExists: false,
			}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errAccessTokenNotFound)
	}

	cr.Status.AtProvider = groups.GenerateGroupAccessTokenObservation(at)

	if cr.Spec.ForProvider.RenewalPeriodDays != nil {
		cr.Status.RenewAt = common.ComputeNextRotation(at.CreatedAt, at.ExpiresAt, cr.Spec.ForProvider.RenewBeforeDays)
	} else {
		cr.Status.RenewAt = nil
	}

	if groups.ShouldRotateAccessToken(&cr.Spec.ForProvider, at) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	cr.Status.SetConditions(v2.Available())

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

	if e.self {
		return e.createSelf(ctx, cr)
	}
	return e.createOwner(ctx, cr)
}

// createSelf rotates the token via the self endpoint. It is reached only when
// Observe reported a rotation is due; a brand-new token cannot be bootstrapped
// in self-mode (the token must already exist - it is the credential).
func (e *external) createSelf(ctx context.Context, cr *v1alpha1.AccessToken) (managed.ExternalCreation, error) {
	at, _, err := e.client.RotateSelf(
		groups.GenerateRotateSelfOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errSelfRotateFailed)
	}

	meta.SetExternalName(cr, strconv.FormatInt(at.ID, 10))
	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{connectionSecretKeyToken: []byte(at.Token)},
	}, nil
}

func (e *external) createOwner(ctx context.Context, cr *v1alpha1.AccessToken) (managed.ExternalCreation, error) {
	if cr.Spec.ForProvider.GroupID == nil {
		return managed.ExternalCreation{}, errors.New(errMissingGroupID)
	}

	// If a token ID is already set as the external name, try rotating it first.
	// Rotation atomically revokes the old token and issues a new one.
	if existingID, err := strconv.ParseInt(meta.GetExternalName(cr), 10, 64); err == nil && existingID > 0 {
		at, res, rotErr := e.client.RotateGroupAccessToken(
			*cr.Spec.ForProvider.GroupID,
			existingID,
			groups.GenerateRotateGroupAccessTokenOptions(&cr.Spec.ForProvider),
			gitlab.WithContext(ctx),
		)
		if rotErr == nil {
			meta.SetExternalName(cr, strconv.FormatInt(at.ID, 10))
			return managed.ExternalCreation{
				ConnectionDetails: managed.ConnectionDetails{connectionSecretKeyToken: []byte(at.Token)},
			}, nil
		}
		// If 401, try self-rotation (token with self_rotate scope can only rotate itself).
		if res != nil && res.StatusCode == http.StatusUnauthorized {
			pat, _, selfErr := e.client.RotateSelf(
				groups.GenerateRotateSelfOptions(&cr.Spec.ForProvider),
				gitlab.WithContext(ctx),
			)
			if selfErr == nil {
				meta.SetExternalName(cr, strconv.FormatInt(pat.ID, 10))
				return managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{connectionSecretKeyToken: []byte(pat.Token)},
				}, nil
			}
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
		ConnectionDetails: managed.ConnectionDetails{connectionSecretKeyToken: []byte(at.Token)},
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

	if e.self {
		// Revoke the token the ProviderConfig authenticates with. This breaks
		// every resource using that ProviderConfig; use deletionPolicy: Orphan
		// to keep the token.
		res, err := e.client.RevokeSelf(gitlab.WithContext(ctx))
		if err != nil {
			// A second reconcile hitting an already-revoked token can no longer
			// authenticate and gets 401/403. Treat that as an idempotent
			// success so the delete does not wedge on an already-completed
			// revoke.
			if clients.IsResponseUnauthorized(res) {
				return managed.ExternalDelete{}, nil
			}
			return managed.ExternalDelete{}, errors.Wrap(err, errDeleteFailed)
		}
		return managed.ExternalDelete{}, nil
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
