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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/groups"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/groups/fake"
)

var (
	unexpecedItem resource.Managed

	errBoom     = errors.New("boom")
	cn          = "cn-group-example"
	groupID     = 1234
	groupAccess = gitlab.AccessLevelValue(10)
)

type LdapGroupLinkModifier func(*v1alpha1.LdapGroupLink)

func withConditions(c ...xpv1.Condition) LdapGroupLinkModifier {
	return func(cr *v1alpha1.LdapGroupLink) { cr.Status.ConditionedStatus.Conditions = c }
}

func withGroupID() LdapGroupLinkModifier {
	return func(r *v1alpha1.LdapGroupLink) { r.Spec.ForProvider.GroupID = &groupID }
}

func withStatus(s v1alpha1.LdapGroupLinkObservation) LdapGroupLinkModifier {
	return func(r *v1alpha1.LdapGroupLink) { r.Status.AtProvider = s }
}

func withGroupAccess(i int) LdapGroupLinkModifier {
	return func(r *v1alpha1.LdapGroupLink) { r.Spec.ForProvider.GroupAccess = v1alpha1.AccessLevelValue(i) }
}

func withSpec(s v1alpha1.LdapGroupLinkParameters) LdapGroupLinkModifier {
	return func(r *v1alpha1.LdapGroupLink) { r.Spec.ForProvider = s }
}

func withExternalName(n string) LdapGroupLinkModifier {
	return func(r *v1alpha1.LdapGroupLink) { meta.SetExternalName(r, n) }
}

func ldapGroupLink(m ...LdapGroupLinkModifier) *v1alpha1.LdapGroupLink {
	cr := &v1alpha1.LdapGroupLink{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

type args struct {
	ldapGroupLink groups.LdapGroupLinkClient
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
				err: errors.New(errNotLdapGroupLink),
			},
		},
		"ProviderConfigRefNotGivenError": {
			args: args{
				cr:   ldapGroupLink(),
				kube: &test.MockClient{MockGet: test.NewMockGetFn(nil)},
			},
			want: want{
				cr:  ldapGroupLink(),
				err: errors.New("providerConfigRef is not given"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &connector{kube: tc.kube, newGitlabClientFn: nil}
			o, err := c.Connect(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
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
				err: errors.New(errNotLdapGroupLink),
			},
		},
		"NoExternalName": {
			args: args{
				cr: ldapGroupLink(),
			},
			want: want{
				cr:     ldapGroupLink(),
				err:    nil,
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"NoGroupID": {
			args: args{
				cr: ldapGroupLink(
					withExternalName(cn),
				),
			},
			want: want{
				cr: ldapGroupLink(
					withExternalName(cn),
				),
				err:    errors.New(errMissingGroupID),
				result: managed.ExternalObservation{},
			},
		},
		"FailedGetRequest": {
			args: args{
				ldapGroupLink: &fake.MockClient{
					MockListGroupLDAPLinks: func(gid interface{}, options ...gitlab.RequestOptionFunc) ([]*gitlab.LDAPGroupLink, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 400}}, errBoom
					},
				},
				cr: ldapGroupLink(
					withGroupID(),
					withGroupAccess(10),
					withExternalName(cn),
					withSpec(v1alpha1.LdapGroupLinkParameters{GroupID: &groupID, CN: cn}),
				),
			},
			want: want{
				cr: ldapGroupLink(
					withGroupID(),
					withGroupAccess(10),
					withExternalName(cn),
					withSpec(v1alpha1.LdapGroupLinkParameters{GroupID: &groupID, CN: cn}),
				),
				result: managed.ExternalObservation{ResourceExists: false},
				err:    errors.Wrap(errBoom, errGetFailed),
			},
		},
		"ErrGet404": {
			args: args{
				ldapGroupLink: &fake.MockClient{
					MockListGroupLDAPLinks: func(gid interface{}, options ...gitlab.RequestOptionFunc) ([]*gitlab.LDAPGroupLink, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errors.New("404 Not Found")
					},
				},
				cr: ldapGroupLink(
					withGroupID(),
					withExternalName(cn),
					withGroupAccess(10),

					withSpec(v1alpha1.LdapGroupLinkParameters{GroupID: &groupID, CN: cn}),
				),
			},
			want: want{
				cr: ldapGroupLink(
					withGroupID(),
					withExternalName(cn),
					withGroupAccess(10),
					withSpec(v1alpha1.LdapGroupLinkParameters{GroupID: &groupID, CN: cn}),
				),
				result: managed.ExternalObservation{ResourceExists: false},
				err:    nil,
			},
		},
		"SuccessfulAvailable": {
			args: args{
				ldapGroupLink: &fake.MockClient{
					MockListGroupLDAPLinks: func(gid interface{}, options ...gitlab.RequestOptionFunc) ([]*gitlab.LDAPGroupLink, *gitlab.Response, error) {
						return []*gitlab.LDAPGroupLink{{CN: cn, GroupAccess: groupAccess}}, &gitlab.Response{}, nil
					},
				},
				cr: ldapGroupLink(
					withGroupID(),
					withExternalName(cn),
					withSpec(v1alpha1.LdapGroupLinkParameters{GroupID: &groupID, CN: cn, GroupAccess: v1alpha1.AccessLevelValue(groupAccess)}),
					withGroupAccess(10),
				),
			},
			want: want{
				cr: ldapGroupLink(
					withConditions(xpv1.Available()),
					withGroupID(),
					withExternalName(cn),
					withSpec(v1alpha1.LdapGroupLinkParameters{GroupID: &groupID, CN: cn, GroupAccess: v1alpha1.AccessLevelValue(groupAccess)}),
					withGroupAccess(10),
					withStatus(v1alpha1.LdapGroupLinkObservation{CN: cn}),
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
			e := &external{kube: tc.kube, client: tc.ldapGroupLink}
			o, err := e.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
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
				err: errors.New(errNotLdapGroupLink),
			},
		},
		"NoExternalName": {
			args: args{
				cr: ldapGroupLink(
					withGroupID(),
				),
			},
			want: want{
				cr: ldapGroupLink(
					withGroupID(),
				),
				err: errors.New(errMissingExternalName),
			},
		},
		"NoGroupID": {
			args: args{
				cr: ldapGroupLink(
					withExternalName(cn),
				),
			},
			want: want{
				cr: ldapGroupLink(
					withExternalName(cn),
				),
				err: errors.New(errMissingGroupID),
			},
		},
		"FailedDeletion": {
			args: args{
				ldapGroupLink: &fake.MockClient{
					MockDeleteGroupLDAPLinkWithCNOrFilter: func(pid interface{}, opts *gitlab.DeleteGroupLDAPLinkWithCNOrFilterOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: ldapGroupLink(
					withGroupID(),
					withExternalName(cn),
					withSpec(v1alpha1.LdapGroupLinkParameters{GroupID: &groupID, CN: cn}),
					withGroupAccess(10),
				),
			},
			want: want{
				cr: ldapGroupLink(
					withGroupID(),
					withExternalName(cn),
					withSpec(v1alpha1.LdapGroupLinkParameters{GroupID: &groupID, CN: cn}),
					withGroupAccess(10),
				),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
		"SuccessfulDeletion": {
			args: args{
				ldapGroupLink: &fake.MockClient{
					MockDeleteGroupLDAPLinkWithCNOrFilter: func(pid interface{}, opts *gitlab.DeleteGroupLDAPLinkWithCNOrFilterOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: ldapGroupLink(
					withGroupID(),
					withExternalName(cn),
					withSpec(v1alpha1.LdapGroupLinkParameters{GroupID: &groupID, CN: cn}),
					withGroupAccess(10),
				),
			},
			want: want{
				cr: ldapGroupLink(
					withGroupID(),
					withExternalName(cn),
					withSpec(v1alpha1.LdapGroupLinkParameters{GroupID: &groupID, CN: cn}),
					withGroupAccess(10),
				),
				err: nil,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.ldapGroupLink}
			_, err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
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
				err: errors.New(errNotLdapGroupLink),
			},
		},
		"NoGroupID": {
			args: args{
				cr: ldapGroupLink(
					withExternalName(cn),
				),
			},
			want: want{
				cr: ldapGroupLink(
					withExternalName(cn),
				),
				err:    errors.New(errMissingGroupID),
				result: managed.ExternalCreation{},
			},
		},
		"FailedCreation": {
			args: args{
				ldapGroupLink: &fake.MockClient{
					MockAddGroupLDAPLink: func(pid interface{}, opt *gitlab.AddGroupLDAPLinkOptions, options ...gitlab.RequestOptionFunc) (*gitlab.LDAPGroupLink, *gitlab.Response, error) {
						return &gitlab.LDAPGroupLink{}, &gitlab.Response{}, errBoom
					},
				},
				cr: ldapGroupLink(
					withGroupID(),
					withSpec(v1alpha1.LdapGroupLinkParameters{GroupID: &groupID, CN: cn}),
					withGroupAccess(10),
				),
			},
			want: want{
				cr: ldapGroupLink(
					withGroupID(),
					withSpec(v1alpha1.LdapGroupLinkParameters{GroupID: &groupID, CN: cn}),
					withGroupAccess(10),
				),
				err: errors.Wrap(errBoom, errCreateFailed),
			},
		},
		"SuccessfulCreation": {
			args: args{
				ldapGroupLink: &fake.MockClient{
					MockAddGroupLDAPLink: func(pid interface{}, opt *gitlab.AddGroupLDAPLinkOptions, options ...gitlab.RequestOptionFunc) (*gitlab.LDAPGroupLink, *gitlab.Response, error) {
						return &gitlab.LDAPGroupLink{CN: cn}, &gitlab.Response{}, nil
					},
				},
				cr: ldapGroupLink(
					withGroupID(),
					withSpec(v1alpha1.LdapGroupLinkParameters{GroupID: &groupID, CN: cn}),
					withGroupAccess(10),
				),
			},
			want: want{
				cr: ldapGroupLink(
					withGroupID(),
					withExternalName(cn),
					withSpec(v1alpha1.LdapGroupLinkParameters{GroupID: &groupID, CN: cn}),
					withGroupAccess(10),
				),
				err:    nil,
				result: managed.ExternalCreation{},
			},
		},
		"SuccessfulCreationWithMemberRoleID": {
			args: args{
				ldapGroupLink: &fake.MockClient{
					MockAddGroupLDAPLink: func(pid interface{}, opt *gitlab.AddGroupLDAPLinkOptions, options ...gitlab.RequestOptionFunc) (*gitlab.LDAPGroupLink, *gitlab.Response, error) {
						return &gitlab.LDAPGroupLink{CN: cn}, &gitlab.Response{}, nil
					},
				},
				cr: ldapGroupLink(
					withGroupID(),
					withSpec(v1alpha1.LdapGroupLinkParameters{GroupID: &groupID, CN: cn}),
					withGroupAccess(10),
				),
			},
			want: want{
				cr: ldapGroupLink(
					withGroupID(),
					withExternalName(cn),
					withSpec(v1alpha1.LdapGroupLinkParameters{GroupID: &groupID, CN: cn}),
					withGroupAccess(10),
				),
				err:    nil,
				result: managed.ExternalCreation{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.ldapGroupLink}
			o, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
