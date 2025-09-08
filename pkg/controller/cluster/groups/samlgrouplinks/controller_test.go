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
	"net/http"
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/cluster/groups/v1alpha1"
	sharedGroupsV1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/shared/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups/fake"
	shared "github.com/crossplane-contrib/provider-gitlab/pkg/controller/shared/groups/samlgrouplinks"
)

var (
	unexpecedItem resource.Managed

	errBoom     = errors.New("boom")
	name        = "Saml-example"
	groupID     = 1234
	accessLevel = gitlab.AccessLevelValue(10)
)

type SamlGroupLinkModifier func(*v1alpha1.SamlGroupLink)

func withConditions(c ...xpv1.Condition) SamlGroupLinkModifier {
	return func(cr *v1alpha1.SamlGroupLink) { cr.Status.Conditions = c }
}

func withGroupID() SamlGroupLinkModifier {
	return func(r *v1alpha1.SamlGroupLink) { r.Spec.ForProvider.GroupID = &groupID }
}

func withStatus(s sharedGroupsV1alpha1.SamlGroupLinkObservation) SamlGroupLinkModifier {
	return func(r *v1alpha1.SamlGroupLink) { r.Status.AtProvider = s }
}

func withAccessLevel(i int) SamlGroupLinkModifier {
	return func(r *v1alpha1.SamlGroupLink) {
		r.Spec.ForProvider.AccessLevel = sharedGroupsV1alpha1.AccessLevelValue(i)
	}
}

func withSpec(s sharedGroupsV1alpha1.SamlGroupLinkParameters) SamlGroupLinkModifier {
	return func(r *v1alpha1.SamlGroupLink) { r.Spec.ForProvider.SamlGroupLinkParameters = s }
}

func withExternalName(n string) SamlGroupLinkModifier {
	return func(r *v1alpha1.SamlGroupLink) { meta.SetExternalName(r, n) }
}

func withMemberRoleID(i int) SamlGroupLinkModifier {
	return func(r *v1alpha1.SamlGroupLink) { r.Spec.ForProvider.MemberRoleID = ptr.To(i) }
}

func samlGroupLink(m ...SamlGroupLinkModifier) *v1alpha1.SamlGroupLink {
	cr := &v1alpha1.SamlGroupLink{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

type args struct {
	samlGroupLink groups.SamlGroupLinkClient
	kube          client.Client
	cr            resource.Managed
}

func TestConnect(t *testing.T) {
	type want struct {
		cr     resource.Managed
		result managed.ExternalClient
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(shared.ErrNotSamlGroupLink),
			},
		},
		"ProviderConfigRefNotGivenError": {
			args: args{
				cr:   samlGroupLink(),
				kube: &test.MockClient{MockGet: test.NewMockGetFn(nil)},
			},
			want: want{
				cr:  samlGroupLink(),
				err: errors.New("providerConfigRef is not given"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &connector{kube: tc.kube, newGitlabClientFn: nil}
			o, err := c.Connect(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     resource.Managed
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(shared.ErrNotSamlGroupLink),
			},
		},
		"NoExternalName": {
			args: args{
				cr: samlGroupLink(),
			},
			want: want{
				cr:     samlGroupLink(),
				err:    nil,
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"NoGroupID": {
			args: args{
				cr: samlGroupLink(
					withExternalName(name),
				),
			},
			want: want{
				cr: samlGroupLink(
					withExternalName(name),
				),
				err:    errors.New(shared.ErrMissingGroupID),
				result: managed.ExternalObservation{},
			},
		},
		"FailedGetRequest": {
			args: args{
				samlGroupLink: &fake.MockClient{
					MockGetGroupSAMLLink: func(gid interface{}, samlGroupName string, options ...gitlab.RequestOptionFunc) (*gitlab.SAMLGroupLink, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 400}}, errBoom
					},
				},
				cr: samlGroupLink(
					withGroupID(),
					withAccessLevel(10),
					withExternalName(name),
					withSpec(sharedGroupsV1alpha1.SamlGroupLinkParameters{GroupID: &groupID, Name: &name}),
				),
			},
			want: want{
				cr: samlGroupLink(
					withGroupID(),
					withAccessLevel(10),
					withExternalName(name),
					withSpec(sharedGroupsV1alpha1.SamlGroupLinkParameters{GroupID: &groupID, Name: &name}),
				),
				result: managed.ExternalObservation{ResourceExists: false},
				err:    errors.Wrap(errBoom, shared.ErrGetFailed),
			},
		},
		"ErrGet404": {
			args: args{
				samlGroupLink: &fake.MockClient{
					MockGetGroupSAMLLink: func(gid interface{}, samlGroupName string, options ...gitlab.RequestOptionFunc) (*gitlab.SAMLGroupLink, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errors.New("Linked SAML group link not found")
					},
				},
				cr: samlGroupLink(
					withGroupID(),
					withExternalName(name),
					withAccessLevel(10),

					withSpec(sharedGroupsV1alpha1.SamlGroupLinkParameters{GroupID: &groupID, Name: &name}),
				),
			},
			want: want{
				cr: samlGroupLink(
					withGroupID(),
					withExternalName(name),
					withAccessLevel(10),
					withSpec(sharedGroupsV1alpha1.SamlGroupLinkParameters{GroupID: &groupID, Name: &name}),
				),
				result: managed.ExternalObservation{ResourceExists: false},
				err:    nil,
			},
		},
		"SuccessfulAvailable": {
			args: args{
				samlGroupLink: &fake.MockClient{
					MockGetGroupSAMLLink: func(gid interface{}, samlGroupName string, options ...gitlab.RequestOptionFunc) (*gitlab.SAMLGroupLink, *gitlab.Response, error) {
						return &gitlab.SAMLGroupLink{Name: name, AccessLevel: accessLevel}, &gitlab.Response{}, nil
					},
				},
				cr: samlGroupLink(
					withGroupID(),
					withExternalName(name),
					withSpec(sharedGroupsV1alpha1.SamlGroupLinkParameters{GroupID: &groupID, Name: &name}),
					withAccessLevel(10),
				),
			},
			want: want{
				cr: samlGroupLink(
					withConditions(xpv1.Available()),
					withGroupID(),
					withExternalName(name),
					withSpec(sharedGroupsV1alpha1.SamlGroupLinkParameters{GroupID: &groupID, Name: &name}),
					withAccessLevel(10),
					withStatus(sharedGroupsV1alpha1.SamlGroupLinkObservation{Name: name}),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: false,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &shared.External{Client: tc.samlGroupLink}
			o, err := e.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type want struct {
		cr  resource.Managed
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(shared.ErrNotSamlGroupLink),
			},
		},
		"NoExternalName": {
			args: args{
				cr: samlGroupLink(
					withGroupID(),
				),
			},
			want: want{
				cr: samlGroupLink(
					withGroupID(),
				),
				err: errors.New(shared.ErrMissingExternalName),
			},
		},
		"NoGroupID": {
			args: args{
				cr: samlGroupLink(
					withExternalName(name),
				),
			},
			want: want{
				cr: samlGroupLink(
					withExternalName(name),
				),
				err: errors.New(shared.ErrMissingGroupID),
			},
		},
		"FailedDeletion": {
			args: args{
				samlGroupLink: &fake.MockClient{
					MockDeleteGroupSAMLLink: func(pid interface{}, samlGroupName string, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: samlGroupLink(
					withGroupID(),
					withExternalName(name),
					withSpec(sharedGroupsV1alpha1.SamlGroupLinkParameters{GroupID: &groupID, Name: &name}),
					withAccessLevel(10),
				),
			},
			want: want{
				cr: samlGroupLink(
					withGroupID(),
					withExternalName(name),
					withSpec(sharedGroupsV1alpha1.SamlGroupLinkParameters{GroupID: &groupID, Name: &name}),
					withAccessLevel(10),
				),
				err: errors.Wrap(errBoom, shared.ErrDeleteFailed),
			},
		},
		"SuccessfulDeletion": {
			args: args{
				samlGroupLink: &fake.MockClient{
					MockDeleteGroupSAMLLink: func(pid interface{}, samlGroupName string, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: samlGroupLink(
					withGroupID(),
					withExternalName(name),
					withSpec(sharedGroupsV1alpha1.SamlGroupLinkParameters{GroupID: &groupID, Name: &name}),
					withAccessLevel(10),
				),
			},
			want: want{
				cr: samlGroupLink(
					withGroupID(),
					withExternalName(name),
					withSpec(sharedGroupsV1alpha1.SamlGroupLinkParameters{GroupID: &groupID, Name: &name}),
					withAccessLevel(10),
				),
				err: nil,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &shared.External{Client: tc.samlGroupLink}
			_, err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type want struct {
		cr     resource.Managed
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(shared.ErrNotSamlGroupLink),
			},
		},
		"NoGroupID": {
			args: args{
				cr: samlGroupLink(
					withExternalName(name),
				),
			},
			want: want{
				cr: samlGroupLink(
					withExternalName(name),
				),
				err:    errors.New(shared.ErrMissingGroupID),
				result: managed.ExternalCreation{},
			},
		},
		"FailedCreation": {
			args: args{
				samlGroupLink: &fake.MockClient{
					MockAddGroupSAMLLink: func(pid interface{}, opt *gitlab.AddGroupSAMLLinkOptions, options ...gitlab.RequestOptionFunc) (*gitlab.SAMLGroupLink, *gitlab.Response, error) {
						return &gitlab.SAMLGroupLink{}, &gitlab.Response{}, errBoom
					},
				},
				cr: samlGroupLink(
					withGroupID(),
					withSpec(sharedGroupsV1alpha1.SamlGroupLinkParameters{GroupID: &groupID, Name: &name}),
					withAccessLevel(10),
				),
			},
			want: want{
				cr: samlGroupLink(
					withGroupID(),
					withSpec(sharedGroupsV1alpha1.SamlGroupLinkParameters{GroupID: &groupID, Name: &name}),
					withAccessLevel(10),
				),
				err: errors.Wrap(errBoom, shared.ErrCreateFailed),
			},
		},
		"SuccessfulCreation": {
			args: args{
				samlGroupLink: &fake.MockClient{
					MockAddGroupSAMLLink: func(pid interface{}, opt *gitlab.AddGroupSAMLLinkOptions, options ...gitlab.RequestOptionFunc) (*gitlab.SAMLGroupLink, *gitlab.Response, error) {
						return &gitlab.SAMLGroupLink{Name: name}, &gitlab.Response{}, nil
					},
				},
				cr: samlGroupLink(
					withGroupID(),
					withSpec(sharedGroupsV1alpha1.SamlGroupLinkParameters{GroupID: &groupID, Name: &name}),
					withAccessLevel(10),
				),
			},
			want: want{
				cr: samlGroupLink(
					withGroupID(),
					withExternalName(name),
					withSpec(sharedGroupsV1alpha1.SamlGroupLinkParameters{GroupID: &groupID, Name: &name}),
					withAccessLevel(10),
				),
				err:    nil,
				result: managed.ExternalCreation{},
			},
		},
		"SuccessfulCreationWithMemberRoleID": {
			args: args{
				samlGroupLink: &fake.MockClient{
					MockAddGroupSAMLLink: func(pid interface{}, opt *gitlab.AddGroupSAMLLinkOptions, options ...gitlab.RequestOptionFunc) (*gitlab.SAMLGroupLink, *gitlab.Response, error) {
						return &gitlab.SAMLGroupLink{Name: name}, &gitlab.Response{}, nil
					},
				},
				cr: samlGroupLink(
					withGroupID(),
					withSpec(sharedGroupsV1alpha1.SamlGroupLinkParameters{GroupID: &groupID, Name: &name}),
					withAccessLevel(10),
					withMemberRoleID(10),
				),
			},
			want: want{
				cr: samlGroupLink(
					withGroupID(),
					withExternalName(name),
					withSpec(sharedGroupsV1alpha1.SamlGroupLinkParameters{GroupID: &groupID, Name: &name}),
					withAccessLevel(10),
					withMemberRoleID(10),
				),
				err:    nil,
				result: managed.ExternalCreation{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &shared.External{Client: tc.samlGroupLink}
			o, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
