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

	"github.com/crossplane/crossplane-runtime/v2/apis/common"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiCluster "github.com/crossplane-contrib/provider-gitlab/apis/cluster/groups/v1alpha1"
	apiNamespaced "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	sharedGroupsV1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/shared/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/users"
)

const (
	ErrNotMember       = "managed resource is not a Gitlab Group Member custom resource"
	ErrIDNotInt        = "ID is not an integer value"
	ErrCreateFailed    = "cannot create Gitlab Group Member"
	ErrUpdateFailed    = "cannot update Gitlab Group Member"
	ErrDeleteFailed    = "cannot delete Gitlab Group Member"
	ErrGetFailed       = "cannot get Gitlab Group Member"
	ErrMissingGroupID  = "Group ID not set"
	ErrMissingUserInfo = "UserID or UserName not set"
	ErrFetchFailed     = "can not fetch userID by userName"
)

type External struct {
	Client     groups.MemberClient
	UserClient users.UserClient
	Kube       client.Client
}

type options struct {
	parameters    *sharedGroupsV1alpha1.MemberParameters
	atProvider    *sharedGroupsV1alpha1.MemberObservation
	setConditions func(c ...common.Condition)
	mg            resource.Managed
}

func (e *External) extractOptions(mg resource.Managed) (*options, error) {
	switch cr := mg.(type) {
	case *apiCluster.Member:
		return &options{
			parameters:    &cr.Spec.ForProvider.MemberParameters,
			atProvider:    &cr.Status.AtProvider,
			setConditions: cr.Status.SetConditions,
			mg:            mg,
		}, nil
	case *apiNamespaced.Member:
		return &options{
			parameters:    &cr.Spec.ForProvider.MemberParameters,
			atProvider:    &cr.Status.AtProvider,
			setConditions: cr.Status.SetConditions,
			mg:            mg,
		}, nil
	default:
		return nil, errors.New(ErrNotMember)
	}
}

func (e *External) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	o, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if o.parameters.GroupID == nil {
		return managed.ExternalObservation{}, errors.New(ErrMissingGroupID)
	}

	userID := o.parameters.UserID
	if userID == nil {
		if o.parameters.UserName == nil {
			return managed.ExternalObservation{}, errors.New(ErrMissingUserInfo)
		}
		resolvedUserID, err := users.GetUserID(e.UserClient, *o.parameters.UserName)
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, ErrFetchFailed)
		}
		userID = resolvedUserID
		o.parameters.UserID = userID
	}

	groupMember, res, err := e.Client.GetGroupMember(*o.parameters.GroupID, *userID)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, ErrGetFailed)
	}

	*o.atProvider = groups.GenerateMemberObservation(groupMember)

	o.setConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isMemberUpToDate(o.parameters, groupMember),
		ResourceLateInitialized: false,
	}, nil
}

func (e *External) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	o, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	if o.parameters.GroupID == nil {
		return managed.ExternalCreation{}, errors.New(ErrMissingGroupID)
	}

	_, _, err = e.Client.AddGroupMember(
		*o.parameters.GroupID,
		groups.GenerateAddMemberOptions(o.parameters),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, ErrCreateFailed)
	}

	return managed.ExternalCreation{}, nil
}

func (e *External) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	o, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	if o.parameters.GroupID == nil {
		return managed.ExternalUpdate{}, errors.New(ErrMissingGroupID)
	}

	if o.parameters.UserID == nil {
		return managed.ExternalUpdate{}, errors.New(ErrMissingUserInfo)
	}

	_, _, err = e.Client.EditGroupMember(
		*o.parameters.GroupID,
		*o.parameters.UserID,
		groups.GenerateEditMemberOptions(o.parameters),
		gitlab.WithContext(ctx),
	)
	return managed.ExternalUpdate{}, errors.Wrap(err, ErrUpdateFailed)
}

func (e *External) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	o, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalDelete{}, err
	}

	if o.parameters.GroupID == nil {
		return managed.ExternalDelete{}, errors.New(ErrMissingGroupID)
	}

	if o.parameters.UserID == nil {
		return managed.ExternalDelete{}, errors.New(ErrMissingUserInfo)
	}

	_, err = e.Client.RemoveGroupMember(*o.parameters.GroupID, *o.parameters.UserID, nil, gitlab.WithContext(ctx))
	return managed.ExternalDelete{}, errors.Wrap(err, ErrDeleteFailed)
}

// isMemberUpToDate checks whether there is a change in any of the modifiable fields.
func isMemberUpToDate(p *sharedGroupsV1alpha1.MemberParameters, g *gitlab.GroupMember) bool {
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

func (e *External) Disconnect(ctx context.Context) error {
	return nil
}
