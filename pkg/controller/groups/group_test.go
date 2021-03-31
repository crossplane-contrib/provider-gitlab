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

package groups

import (
	"context"
	"strconv"
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups/fake"
)

var (
	path              = "path/to/group"
	name              = "example-group"
	groupID           = 1234
	extName           = strconv.Itoa(groupID)
	errBoom           = errors.New("boom")
	extNameAnnotation = map[string]string{
		meta.AnnotationKeyExternalName: extName,
	}
)

type args struct {
	group groups.Client
	kube  client.Client
	cr    *v1alpha1.Group
}

type groupModifier func(*v1alpha1.Group)

func withConditions(c ...xpv1.Condition) groupModifier {
	return func(cr *v1alpha1.Group) { cr.Status.ConditionedStatus.Conditions = c }
}

func withPath(p *string) groupModifier {
	return func(r *v1alpha1.Group) { r.Spec.ForProvider.Path = p }
}

func withExternalName(n string) groupModifier {
	return func(r *v1alpha1.Group) { meta.SetExternalName(r, n) }
}

func withStatus(s v1alpha1.GroupObservation) groupModifier {
	return func(r *v1alpha1.Group) { r.Status.AtProvider = s }
}

func withDefaultValues() groupModifier {
	return func(p *v1alpha1.Group) {
		f := false
		i := 0
		p.Spec.ForProvider = v1alpha1.GroupParameters{
			MembershipLock:                 &f,
			ShareWithGroupLock:             &f,
			RequireTwoFactorAuth:           &f,
			TwoFactorGracePeriod:           &i,
			AutoDevopsEnabled:              &f,
			EmailsDisabled:                 &f,
			MentionsDisabled:               &f,
			LFSEnabled:                     &f,
			RequestAccessEnabled:           &f,
			ParentID:                       &i,
			SharedRunnersMinutesLimit:      &i,
			ExtraSharedRunnersMinutesLimit: &i,
		}
	}
}

func withAnnotations(a map[string]string) groupModifier {
	return func(p *v1alpha1.Group) { meta.AddAnnotations(p, a) }
}

func group(m ...groupModifier) *v1alpha1.Group {
	cr := &v1alpha1.Group{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.Group
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"NoExternalName": {
			args: args{
				cr: group(),
			},
			want: want{
				cr: group(),
				result: managed.ExternalObservation{
					ResourceExists:          false,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
			},
		},
		"NotIDExternalName": {
			args: args{
				group: &fake.MockClient{
					MockGetGroup: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						return &gitlab.Group{}, &gitlab.Response{}, nil
					},
				},
				cr: group(withExternalName("fr")),
			},
			want: want{
				cr:  group(withExternalName("fr")),
				err: errors.New(errNotGroup),
			},
		},
		"SuccessfulAvailable": {
			args: args{
				group: &fake.MockClient{
					MockGetGroup: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						return &gitlab.Group{Name: name}, &gitlab.Response{}, nil
					},
				},
				cr: group(
					withDefaultValues(),
					withExternalName(extName),
				),
			},
			want: want{
				cr: group(
					withDefaultValues(),
					withConditions(xpv1.Available()),
					withAnnotations(extNameAnnotation),
					withExternalName(extName),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: false,
					ConnectionDetails:       managed.ConnectionDetails{"runnersToken": []byte("")},
				},
			},
		},
		"FailedGetRequest": {
			args: args{
				group: &fake.MockClient{
					MockGetGroup: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						return &gitlab.Group{}, &gitlab.Response{}, errBoom
					},
				},
				cr: group(withExternalName(extName)),
			},
			want: want{
				cr:  group(withAnnotations(extNameAnnotation)),
				err: errors.Wrap(errBoom, errGetFailed),
			},
		},
		"LateInitSuccess": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				group: &fake.MockClient{
					MockGetGroup: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						return &gitlab.Group{
							Path:         path,
							RunnersToken: "token",
						}, &gitlab.Response{}, nil
					},
				},
				cr: group(
					withDefaultValues(),
					withExternalName(extName),
				),
			},
			want: want{
				cr: group(
					withDefaultValues(),
					withConditions(xpv1.Available()),
					withPath(&path),
					withAnnotations(extNameAnnotation),
					withStatus(v1alpha1.GroupObservation{RunnersToken: "token"}),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
					ConnectionDetails:       managed.ConnectionDetails{"runnersToken": []byte("token")},
				},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.group}
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
		cr     *v1alpha1.Group
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulCreation": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				group: &fake.MockClient{
					MockCreateGroup: func(opt *gitlab.CreateGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						return &gitlab.Group{Name: extName, Path: extName, ID: 0}, &gitlab.Response{}, nil
					},
				},
				cr: group(
					withAnnotations(extNameAnnotation),
				),
			},
			want: want{
				cr: group(
					withConditions(xpv1.Creating()),
					withExternalName("0"),
				),
				result: managed.ExternalCreation{ExternalNameAssigned: true},
			},
		},
		"FailedCreation": {
			args: args{
				group: &fake.MockClient{
					MockCreateGroup: func(opt *gitlab.CreateGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						return &gitlab.Group{}, &gitlab.Response{}, errBoom
					},
				},
				cr: group(
					withStatus(v1alpha1.GroupObservation{ID: 0}),
				),
			},
			want: want{
				cr: group(
					withConditions(xpv1.Creating()),
					withStatus(v1alpha1.GroupObservation{ID: 0}),
				),
				err: errors.Wrap(errBoom, errCreateFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.group}
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
		cr     *v1alpha1.Group
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulUpdateGroup": {
			args: args{
				group: &fake.MockClient{
					MockUpdateGroup: func(pid interface{}, opt *gitlab.UpdateGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						return &gitlab.Group{ID: 1234}, &gitlab.Response{}, nil
					},
				},
				cr: group(withStatus(v1alpha1.GroupObservation{ID: 1234}), withExternalName("1234")),
			},
			want: want{
				cr: group(
					withStatus(v1alpha1.GroupObservation{ID: 1234}),
					withExternalName("1234"),
				),
			},
		},
		"FailedEdit": {
			args: args{
				group: &fake.MockClient{
					MockUpdateGroup: func(pid interface{}, opt *gitlab.UpdateGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						return &gitlab.Group{}, &gitlab.Response{}, errBoom
					},
				},
				cr: group(withStatus(v1alpha1.GroupObservation{ID: 1234})),
			},
			want: want{
				cr:  group(withStatus(v1alpha1.GroupObservation{ID: 1234})),
				err: errors.Wrap(errBoom, errUpdateFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.group}
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
		cr  *v1alpha1.Group
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulDeletion": {
			args: args{
				group: &fake.MockClient{
					MockDeleteGroup: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: group(
					withConditions(xpv1.Available()),
				),
			},
			want: want{
				cr: group(
					withConditions(xpv1.Deleting()),
				),
			},
		},
		"FailedDeletion": {
			args: args{
				group: &fake.MockClient{
					MockDeleteGroup: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: group(),
			},
			want: want{
				cr: group(
					withConditions(xpv1.Deleting()),
				),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.group}
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
