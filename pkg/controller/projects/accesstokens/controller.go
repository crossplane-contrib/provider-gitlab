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
	"strings"

	"github.com/xanzy/go-gitlab"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
)

const (
	errNotAccessToken             = "managed resource is not a Gitlab accesstoken custom resource"
	errGetFailed                  = "cannot get Gitlab accesstoken"
	errCreateFailed               = "cannot create Gitlab accesstoken"
	errDeleteFailed               = "cannot delete Gitlab accesstoken"
	errProjecAccessTokentNotFound = "cannot find Gitlab accesstoken"
)

// SetupAccessToken adds a controller that reconciles ProjectAccessTokens.
func SetupAccessToken(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.AccessTokenKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.AccessToken{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.AccessTokenGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabClientFn: newAccessTokenClient}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg clients.Config) gitlab.ProjectAccessTokensService
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.AccessToken)
	if !ok {
		return nil, errors.New(errNotAccessToken)
	}
	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

type external struct {
	kube   client.Client
	client gitlab.ProjectAccessTokensService
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.AccessToken)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotAccessToken)
	}

	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	projectAccessTokenID, err := strconv.Atoi(externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.New(errNotAccessToken)
	}

	listProjectAccessTokensOptions := gitlab.ListProjectAccessTokensOptions{}
	accesstokenArr, _, err := e.client.ListProjectAccessTokens(*cr.Spec.ForProvider.ProjectID, &listProjectAccessTokensOptions)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(isErrorProjectAccessTokenNotFound, err), errGetFailed)
	}

	dt := findAccessToken(projectAccessTokenID, accesstokenArr)

	if dt == nil {
		return managed.ExternalObservation{}, nil
	}

	current := cr.Spec.ForProvider.DeepCopy()
	lateInitializeProjectAccessToken(&cr.Spec.ForProvider, dt)

	cr.Status.AtProvider = v1alpha1.AccessTokenObservation{}
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        true,
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.AccessToken)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotAccessToken)
	}

	dt, _, err := e.client.CreateProjectAccessToken(
		*cr.Spec.ForProvider.ProjectID,
		generateCreateProjectAccessTokenOptions(cr.Name, &cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	connectionDetails := managed.ConnectionDetails{}
	connectionDetails["token"] = []byte(dt.Token)

	meta.SetExternalName(cr, strconv.Itoa(dt.ID))
	return managed.ExternalCreation{ExternalNameAssigned: true, ConnectionDetails: connectionDetails}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// it's not possible to update a ProjectAccessToken
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.AccessToken)
	if !ok {
		return errors.New(errNotAccessToken)
	}

	accessTokenID, err := strconv.Atoi(meta.GetExternalName(cr))

	if err != nil {
		return errors.New(errNotAccessToken)
	}

	_, deleteError := e.client.RevokeProjectAccessToken(
		*cr.Spec.ForProvider.ProjectID,
		accessTokenID,
		gitlab.WithContext(ctx),
	)

	return errors.Wrap(deleteError, errDeleteFailed)
}

// lateInitializeProjectAccessToken fills the empty fields in the access token spec with the
// values seen in gitlab access token.
func lateInitializeProjectAccessToken(in *v1alpha1.AccessTokenParameters, accessToken *gitlab.ProjectAccessToken) { // nolint:gocyclo
	if accessToken == nil {
		return
	}

	if in.AccessLevel == nil {
		in.AccessLevel = (*v1alpha1.AccessLevelValue)(&accessToken.AccessLevel)
	}

	if in.ExpiresAt == nil && accessToken.ExpiresAt != nil {
		in.ExpiresAt = &metav1.Time{Time: in.ExpiresAt.Time}
	}
}

// findAccessToken try to find a access token with the ID in the access token array,
// if found return a access token otherwise return nil.
func findAccessToken(accessTokenID int, accessTokens []*gitlab.ProjectAccessToken) *gitlab.ProjectAccessToken {
	for _, v := range accessTokens {
		if v.ID == accessTokenID {
			return v
		}
	}
	return nil
}

//-- What was from interface file --//

// IsErrorProjectAccessTokenNotFound helper function to test for errProjectAccessTokenNotFound error.
func isErrorProjectAccessTokenNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errProjecAccessTokentNotFound)
}

// NewAccessTokenClient returns a new Gitlab ProjectAccessToken service
func newAccessTokenClient(cfg clients.Config) gitlab.ProjectAccessTokensService {
	git := clients.NewClient(cfg)
	return *git.ProjectAccessTokens
}

// GenerateCreateProjectAccessTokenOptions generates project creation options
func generateCreateProjectAccessTokenOptions(name string, p *v1alpha1.AccessTokenParameters) *gitlab.CreateProjectAccessTokenOptions {
	accesstoken := &gitlab.CreateProjectAccessTokenOptions{
		Name:   &name,
		Scopes: &p.Scopes,
	}

	if p.ExpiresAt != nil {
		accesstoken.ExpiresAt = (*gitlab.ISOTime)(&p.ExpiresAt.Time)
	}

	if p.AccessLevel != nil {
		accesstoken.AccessLevel = (*gitlab.AccessLevelValue)(p.AccessLevel)
	}

	return accesstoken
}
