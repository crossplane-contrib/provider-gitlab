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

package members

import (
	"context"

	"github.com/xanzy/go-gitlab"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/users"
)

const (
	errNotMember       = "managed resource is not a Gitlab Group Member custom resource"
	errIDNotInt        = "ID is not an integer value"
	errCreateFailed    = "cannot create Gitlab Group Member"
	errUpdateFailed    = "cannot update Gitlab Group Member"
	errDeleteFailed    = "cannot delete Gitlab Group Member"
	errGetFailed       = "cannot get Gitlab Group Member"
	errMissingGroupID  = "Group ID not set"
	errMissingUserInfo = "UserID or UserName not set"
	errFetchFailed     = "can not fetch userID by userName"
)

// SetupMember adds a controller that reconciles Group Members.
func SetupMember(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.MemberKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Member{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.MemberKubernetesGroupVersionKind),
			managed.WithExternalConnecter(&connector{
				kube:              mgr.GetClient(),
				newGitlabClientFn: groups.NewMemberClient,
				newUserClientFn:   users.NewUserClient}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg clients.Config) groups.MemberClient
	newUserClientFn   func(cfg clients.Config) users.UserClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Member)
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
	client     groups.MemberClient
	userClient users.UserClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Member)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotMember)
	}

	if cr.Spec.ForProvider.GroupID == nil {
		return managed.ExternalObservation{}, errors.New(errMissingGroupID)
	}

	userID, err := cr.Spec.ForProvider.UserID, error(nil)
	if cr.Spec.ForProvider.UserID == nil {
		if cr.Spec.ForProvider.UserName == nil {
			return managed.ExternalObservation{}, errors.New(errMissingUserInfo)
		}
		userID, err = users.GetUserID(e.userClient, *cr.Spec.ForProvider.UserName)
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errFetchFailed)
		}
	}
	cr.Spec.ForProvider.UserID = userID

	groupMember, res, err := e.client.GetGroupMember(
		*cr.Spec.ForProvider.GroupID,
		*cr.Spec.ForProvider.UserID,
	)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}

	cr.Status.AtProvider = groups.GenerateMemberObservation(groupMember)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isMemberUpToDate(&cr.Spec.ForProvider, groupMember),
		ResourceLateInitialized: false,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Member)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotMember)
	}

	if cr.Spec.ForProvider.GroupID == nil {
		return managed.ExternalCreation{}, errors.New(errMissingGroupID)
	}

	_, _, err := e.client.AddGroupMember(
		*cr.Spec.ForProvider.GroupID,
		groups.GenerateAddMemberOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Member)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotMember)
	}

	if cr.Spec.ForProvider.GroupID == nil {
		return managed.ExternalUpdate{}, errors.New(errMissingGroupID)
	}

	if cr.Spec.ForProvider.UserID == nil {
		return managed.ExternalUpdate{}, errors.New(errMissingUserInfo)
	}

	_, _, err := e.client.EditGroupMember(
		*cr.Spec.ForProvider.GroupID,
		*cr.Spec.ForProvider.UserID,
		groups.GenerateEditMemberOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Member)
	if !ok {
		return errors.New(errNotMember)
	}

	if cr.Spec.ForProvider.GroupID == nil {
		return errors.New(errMissingGroupID)
	}

	if cr.Spec.ForProvider.UserID == nil {
		return errors.New(errMissingUserInfo)
	}

	_, err := e.client.RemoveGroupMember(
		*cr.Spec.ForProvider.GroupID,
		*cr.Spec.ForProvider.UserID,
		nil,
		gitlab.WithContext(ctx),
	)
	return errors.Wrap(err, errDeleteFailed)
}

// isMemberUpToDate checks whether there is a change in any of the modifiable fields.
func isMemberUpToDate(p *v1alpha1.MemberParameters, g *gitlab.GroupMember) bool {

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
