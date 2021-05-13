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
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups/fake"
)

var (
	unexpecedItem     resource.Managed
	errBoom           = errors.New("boom")
	ID                = 0
	extName           = strconv.Itoa(ID)
	extNameAnnotation = map[string]string{meta.AnnotationKeyExternalName: extName}
	username          = "username"
	name              = "name"
	state             = "state"
	avatarURL         = "http://avatarURL"
	webURL            = "http://webURL"
	accessLevel       = gitlab.AccessLevelValue(30)
	now               = time.Now()
	expiresAt         = gitlab.ISOTime(now.AddDate(0, 0, 7*3))
	expiresAtNew      = gitlab.ISOTime(now.AddDate(0, 0, 7*4))
)

type args struct {
	groupMember groups.GroupMemberClient
	kube        client.Client
	cr          resource.Managed
}

func withAccessLevel(i int) groupModifier {
	return func(r *v1alpha1.GroupMember) { r.Spec.ForProvider.AccessLevel = v1alpha1.AccessLevelValue(i) }
}

func withExpiresAt(s string) groupModifier {
	return func(r *v1alpha1.GroupMember) { r.Spec.ForProvider.ExpiresAt = &s }
}

type groupModifier func(*v1alpha1.GroupMember)

func withConditions(c ...xpv1.Condition) groupModifier {
	return func(cr *v1alpha1.GroupMember) { cr.Status.ConditionedStatus.Conditions = c }
}

func withExternalName(n string) groupModifier {
	return func(r *v1alpha1.GroupMember) { meta.SetExternalName(r, n) }
}

func withStatus(s v1alpha1.GroupMemberObservation) groupModifier {
	return func(r *v1alpha1.GroupMember) { r.Status.AtProvider = s }
}

func withAnnotations(a map[string]string) groupModifier {
	return func(p *v1alpha1.GroupMember) { meta.AddAnnotations(p, a) }
}

func groupMember(m ...groupModifier) *v1alpha1.GroupMember {
	cr := &v1alpha1.GroupMember{}
	for _, f := range m {
		f(cr)
	}
	return cr
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
				err: errors.New(errNotGroupMember),
			},
		},
		"ProviderConfigRefNotGivenError": {
			args: args{
				cr:   groupMember(),
				kube: &test.MockClient{MockGet: test.NewMockGetFn(nil)},
			},
			want: want{
				cr:  groupMember(),
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
				err: errors.New(errNotGroupMember),
			},
		},
		"NoExternalName": {
			args: args{
				cr: groupMember(),
			},
			want: want{
				cr: groupMember(),
				result: managed.ExternalObservation{
					ResourceExists:          false,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
			},
		},
		"NotIDExternalName": {
			args: args{
				groupMember: &fake.MockClient{
					MockGetGroupMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
						return &gitlab.GroupMember{}, &gitlab.Response{}, nil
					},
				},
				cr: groupMember(withExternalName("fr")),
			},
			want: want{
				cr:  groupMember(withExternalName("fr")),
				err: errors.New(errNotGroupMember),
			},
		},
		"FailedGetRequest": {
			args: args{
				groupMember: &fake.MockClient{
					MockGetGroupMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
						return &gitlab.GroupMember{}, &gitlab.Response{}, errBoom
					},
				},
				cr: groupMember(withExternalName(extName)),
			},
			want: want{
				cr:  groupMember(withAnnotations(extNameAnnotation)),
				err: errors.Wrap(errBoom, errGetFailed),
			},
		},
		"SuccessfulAvailable": {
			args: args{
				groupMember: &fake.MockClient{
					MockGetGroupMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
						return &gitlab.GroupMember{}, &gitlab.Response{}, nil
					},
				},
				cr: groupMember(
					withExternalName(extName),
				),
			},
			want: want{
				cr: groupMember(
					withConditions(xpv1.Available()),
					withAnnotations(extNameAnnotation),
					withExternalName(extName),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: false,
				},
			},
		},
		"IsGroupUpToDateAccessLevel": {
			args: args{
				groupMember: &fake.MockClient{
					MockGetGroupMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
						return &gitlab.GroupMember{
							AccessLevel: accessLevel,
						}, &gitlab.Response{}, nil
					},
				},
				cr: groupMember(
					withExternalName(extName),
					withAccessLevel(10),
				),
			},
			want: want{
				cr: groupMember(
					withConditions(xpv1.Available()),
					withAnnotations(extNameAnnotation),
					withExternalName(extName),
					withAccessLevel(10),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
			},
		},
		"IsGroupUpToDateExpiresAt": {
			args: args{
				groupMember: &fake.MockClient{
					MockGetGroupMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
						return &gitlab.GroupMember{
							ExpiresAt: &expiresAt,
						}, &gitlab.Response{}, nil
					},
				},
				cr: groupMember(
					withExternalName(extName),
					withExpiresAt(expiresAtNew.String()),
				),
			},
			want: want{
				cr: groupMember(
					withConditions(xpv1.Available()),
					withAnnotations(extNameAnnotation),
					withExternalName(extName),
					withExpiresAt(expiresAtNew.String()),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.groupMember}
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
				err: errors.New(errNotGroupMember),
			},
		},
		"SuccessfulCreationWithoutExpiresAt": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				groupMember: &fake.MockClient{
					MockAddGroupMember: func(gid interface{}, opt *gitlab.AddGroupMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
						return &gitlab.GroupMember{
							ID:          ID,
							Username:    username,
							Name:        name,
							State:       state,
							AvatarURL:   avatarURL,
							WebURL:      webURL,
							AccessLevel: accessLevel,
						}, &gitlab.Response{}, nil
					},
				},
				cr: groupMember(withAnnotations(extNameAnnotation)),
			},
			want: want{
				cr:     groupMember(withExternalName(extName)),
				result: managed.ExternalCreation{ExternalNameAssigned: true},
			},
		},
		"SuccessfulCreationWithExpiresAt": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				groupMember: &fake.MockClient{
					MockAddGroupMember: func(gid interface{}, opt *gitlab.AddGroupMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
						return &gitlab.GroupMember{
							ID:          ID,
							Username:    username,
							Name:        name,
							State:       state,
							AvatarURL:   avatarURL,
							WebURL:      webURL,
							AccessLevel: accessLevel,
							ExpiresAt:   &expiresAt,
						}, &gitlab.Response{}, nil
					},
				},
				cr: groupMember(withAnnotations(extNameAnnotation)),
			},
			want: want{
				cr:     groupMember(withExternalName(extName)),
				result: managed.ExternalCreation{ExternalNameAssigned: true},
			},
		},
		"FailedCreation": {
			args: args{
				groupMember: &fake.MockClient{
					MockAddGroupMember: func(gid interface{}, opt *gitlab.AddGroupMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
						return &gitlab.GroupMember{}, &gitlab.Response{}, errBoom
					},
				},
				cr: groupMember(),
			},
			want: want{
				cr:  groupMember(),
				err: errors.Wrap(errBoom, errCreateFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.groupMember}
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

func TestUpdate(t *testing.T) {
	type want struct {
		cr     resource.Managed
		result managed.ExternalUpdate
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
				err: errors.New(errNotGroupMember),
			},
		},
		"SuccessfulUpdate": {
			args: args{
				groupMember: &fake.MockClient{
					MockEditGroupMember: func(gid interface{}, user int, opt *gitlab.EditGroupMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
						return &gitlab.GroupMember{
							ID:          ID,
							Username:    username,
							Name:        name,
							State:       state,
							AvatarURL:   avatarURL,
							WebURL:      webURL,
							AccessLevel: accessLevel,
						}, &gitlab.Response{}, nil
					},
				},
				cr: groupMember(
					withExternalName(extName),
					withStatus(v1alpha1.GroupMemberObservation{Username: "new username"}),
				),
			},
			want: want{
				cr: groupMember(
					withExternalName(extName),
					withStatus(v1alpha1.GroupMemberObservation{Username: "new username"}),
				),
			},
		},
		"FailedUpdate": {
			args: args{
				groupMember: &fake.MockClient{
					MockEditGroupMember: func(gid interface{}, user int, opt *gitlab.EditGroupMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
						return &gitlab.GroupMember{}, &gitlab.Response{}, errBoom
					},
				},
				cr: groupMember(),
			},
			want: want{
				cr:  groupMember(),
				err: errors.Wrap(errBoom, errUpdateFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.groupMember}
			o, err := e.Update(context.Background(), tc.args.cr)

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
				err: errors.New(errNotGroupMember),
			},
		},
		"SuccessfulDeletion": {
			args: args{
				groupMember: &fake.MockClient{
					MockRemoveGroupMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: groupMember(withExternalName(extName)),
			},
			want: want{
				cr:  groupMember(withExternalName(extName)),
				err: nil,
			},
		},
		"FailedDeletion": {
			args: args{
				groupMember: &fake.MockClient{
					MockRemoveGroupMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: groupMember(),
			},
			want: want{
				cr:  groupMember(),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.groupMember}
			err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}

}
