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

	apiCluster "github.com/crossplane-contrib/provider-gitlab/apis/cluster/projects/v1alpha1"
	apiNamespaced "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	sharedProjectsV1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/shared/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/users"
)

const (
	ErrNotMember        = "managed resource is not a Gitlab Project Member custom resource"
	ErrCreateFailed     = "cannot create Gitlab Project Member"
	ErrUpdateFailed     = "cannot update Gitlab Project Member"
	ErrDeleteFailed     = "cannot delete Gitlab Project Member"
	ErrObserveFailed    = "cannot observe Gitlab Project Member"
	ErrProjectIDMissing = "ProjectID is missing"
	ErrUserInfoMissing  = "UserID or UserName is missing"
	ErrFetchFailed      = "can not fetch userID by UserName"
)

type External struct {
	Client     projects.MemberClient
	UserClient users.UserClient
	Kube       client.Client
}

type options struct {
	parameters    *sharedProjectsV1alpha1.MemberParameters
	atProvider    *sharedProjectsV1alpha1.MemberObservation
	setConditions func(c ...common.Condition)
	mg            resource.Managed
}

func (e *External) extractOptions(mg resource.Managed) (*options, error) {
	switch cr := mg.(type) {
	case *apiCluster.Member:
		return &options{
			parameters:    &cr.Spec.ForProvider.MemberParameters,
			atProvider:    &cr.Status.AtProvider,
			setConditions: cr.SetConditions,
			mg:            mg,
		}, nil
	case *apiNamespaced.Member:
		return &options{
			parameters:    &cr.Spec.ForProvider.MemberParameters,
			atProvider:    &cr.Status.AtProvider,
			setConditions: cr.SetConditions,
			mg:            mg,
		}, nil
	default:
		return nil, errors.New(ErrNotMember)
	}
}

func (e *External) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if opts.parameters.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(ErrProjectIDMissing)
	}

	userID, err := opts.parameters.UserID, error(nil)
	if opts.parameters.UserID == nil {
		if opts.parameters.UserName == nil {
			return managed.ExternalObservation{}, errors.New(ErrUserInfoMissing)
		}
		userID, err = users.GetUserID(e.UserClient, *opts.parameters.UserName)
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, ErrFetchFailed)
		}
	}
	opts.parameters.UserID = userID

	projectMember, res, err := e.Client.GetProjectMember(*opts.parameters.ProjectID, *opts.parameters.UserID)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, ErrObserveFailed)
	}

	*opts.atProvider = projects.GenerateMemberObservation(projectMember)
	opts.setConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        e.isMemberUpToDate(opts.parameters, projectMember),
		ResourceLateInitialized: false,
	}, nil
}

func (e *External) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	if opts.parameters.ProjectID == nil {
		return managed.ExternalCreation{}, errors.New(ErrProjectIDMissing)
	}

	_, _, err = e.Client.AddProjectMember(
		*opts.parameters.ProjectID,
		projects.GenerateAddMemberOptions(opts.parameters),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, ErrCreateFailed)
	}

	return managed.ExternalCreation{}, nil
}

func (e *External) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	if opts.parameters.ProjectID == nil {
		return managed.ExternalUpdate{}, errors.New(ErrProjectIDMissing)
	}
	if opts.parameters.UserID == nil {
		return managed.ExternalUpdate{}, errors.New(ErrUserInfoMissing)
	}

	_, _, err = e.Client.EditProjectMember(
		*opts.parameters.ProjectID,
		*opts.parameters.UserID,
		projects.GenerateEditMemberOptions(opts.parameters),
		gitlab.WithContext(ctx),
	)
	return managed.ExternalUpdate{}, errors.Wrap(err, ErrUpdateFailed)
}

func (e *External) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	opts, err := e.extractOptions(mg)
	if err != nil {
		return managed.ExternalDelete{}, err
	}

	if opts.parameters.ProjectID == nil {
		return managed.ExternalDelete{}, errors.New(ErrProjectIDMissing)
	}
	if opts.parameters.UserID == nil {
		return managed.ExternalDelete{}, errors.New(ErrUserInfoMissing)
	}

	_, err = e.Client.DeleteProjectMember(
		*opts.parameters.ProjectID,
		*opts.parameters.UserID,
		gitlab.WithContext(ctx),
	)
	return managed.ExternalDelete{}, errors.Wrap(err, ErrDeleteFailed)
}

// isMemberUpToDate checks whether there is a change in any of the modifiable fields.
func (e *External) isMemberUpToDate(p *sharedProjectsV1alpha1.MemberParameters, g *gitlab.ProjectMember) bool {
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
