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

package groupmembers

import (
	"context"
	"strconv"

	"github.com/xanzy/go-gitlab"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups"
)

const (
	errNotGroupMember = "managed resource is not a Gitlab Group Member custom resource"
	errGetFailed      = "cannot get Gitlab Group Member"
	errCreateFailed   = "cannot create Gitlab Group Member"
	errUpdateFailed   = "cannot update Gitlab Group Member"
	errDeleteFailed   = "cannot delete Gitlab Group Member"
)

// SetupGroupMember adds a controller that reconciles Group Members.
func SetupGroupMember(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.GroupMemberKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.GroupMember{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.GroupMemberKubernetesGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabClientFn: groups.NewGroupMemberClient}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg clients.Config) groups.GroupMemberClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.GroupMember)
	if !ok {
		return nil, errors.New(errNotGroupMember)
	}
	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

type external struct {
	kube   client.Client
	client groups.GroupMemberClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.GroupMember)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotGroupMember)
	}

	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	groupID, err := strconv.Atoi(externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.New(errNotGroupMember)
	}

	groupMember, _, err := e.client.GetGroupMember(groupID, cr.Spec.ForProvider.UserID)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(groups.IsErrorGroupMemberNotFound, err), errGetFailed)
	}

	cr.Status.AtProvider = groups.GenerateGroupMemberObservation(groupMember)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isGroupMemberUpToDate(&cr.Spec.ForProvider, groupMember),
		ResourceLateInitialized: false,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.GroupMember)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotGroupMember)
	}

	_, _, err := e.client.AddGroupMember(
		cr.Spec.ForProvider.GroupID,
		groups.GenerateAddGroupMemberOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	meta.SetExternalName(cr, strconv.Itoa(cr.Spec.ForProvider.GroupID))
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.GroupMember)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotGroupMember)
	}

	_, _, err := e.client.EditGroupMember(
		meta.GetExternalName(cr),
		cr.Spec.ForProvider.UserID,
		groups.GenerateEditGroupMemberOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.GroupMember)
	if !ok {
		return errors.New(errNotGroupMember)
	}

	_, err := e.client.RemoveGroupMember(
		meta.GetExternalName(cr),
		cr.Spec.ForProvider.UserID,
		gitlab.WithContext(ctx),
	)
	return errors.Wrap(err, errDeleteFailed)
}

// isGroupMemberUpToDate checks whether there is a change in any of the modifiable fields.
func isGroupMemberUpToDate(p *v1alpha1.GroupMemberParameters, g *gitlab.GroupMember) bool {

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
