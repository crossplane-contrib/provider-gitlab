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

package projectmembers

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

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects/fake"
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
	email             = "email@gmail.com"
	accessLevel       = gitlab.AccessLevelValue(30)
	now               = time.Now()
	expiresAt         = gitlab.ISOTime(now.AddDate(0, 0, 7*3))
	expiresAtNew      = gitlab.ISOTime(now.AddDate(0, 0, 7*4))
)

type args struct {
	projectMember projects.ProjectMemberClient
	kube          client.Client
	cr            resource.Managed
}

func withAccessLevel(i int) projectModifier {
	return func(r *v1alpha1.ProjectMember) { r.Spec.ForProvider.AccessLevel = v1alpha1.AccessLevelValue(i) }
}

func withExpiresAt(s string) projectModifier {
	return func(r *v1alpha1.ProjectMember) { r.Spec.ForProvider.ExpiresAt = &s }
}

type projectModifier func(*v1alpha1.ProjectMember)

func withConditions(c ...xpv1.Condition) projectModifier {
	return func(cr *v1alpha1.ProjectMember) { cr.Status.ConditionedStatus.Conditions = c }
}

func withExternalName(n string) projectModifier {
	return func(r *v1alpha1.ProjectMember) { meta.SetExternalName(r, n) }
}

func withStatus(s v1alpha1.ProjectMemberObservation) projectModifier {
	return func(r *v1alpha1.ProjectMember) { r.Status.AtProvider = s }
}

func withSpec(s v1alpha1.ProjectMemberParameters) projectModifier {
	return func(r *v1alpha1.ProjectMember) { r.Spec.ForProvider = s }
}

func withAnnotations(a map[string]string) projectModifier {
	return func(p *v1alpha1.ProjectMember) { meta.AddAnnotations(p, a) }
}

func projectMember(m ...projectModifier) *v1alpha1.ProjectMember {
	cr := &v1alpha1.ProjectMember{}
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
				err: errors.New(errNotProjectMember),
			},
		},
		"ProviderConfigRefNotGivenError": {
			args: args{
				cr:   projectMember(),
				kube: &test.MockClient{MockGet: test.NewMockGetFn(nil)},
			},
			want: want{
				cr:  projectMember(),
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
				err: errors.New(errNotProjectMember),
			},
		},
		"NoExternalName": {
			args: args{
				cr: projectMember(),
			},
			want: want{
				cr: projectMember(),
				result: managed.ExternalObservation{
					ResourceExists:          false,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
			},
		},
		"NotIDExternalName": {
			args: args{
				projectMember: &fake.MockClient{
					MockGetProjectMember: func(pid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
						return &gitlab.ProjectMember{}, &gitlab.Response{}, nil
					},
				},
				cr: projectMember(withExternalName("fr")),
			},
			want: want{
				cr:  projectMember(withExternalName("fr")),
				err: errors.New(errNotProjectMember),
			},
		},
		"FailedGetRequest": {
			args: args{
				projectMember: &fake.MockClient{
					MockGetProjectMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
						return nil, &gitlab.Response{}, errBoom
					},
				},
				cr: projectMember(withExternalName(extName)),
			},
			want: want{
				cr:     projectMember(withAnnotations(extNameAnnotation)),
				result: managed.ExternalObservation{ResourceExists: false},
				err:    nil,
			},
		},
		"SuccessfulAvailable": {
			args: args{
				projectMember: &fake.MockClient{
					MockGetProjectMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
						return &gitlab.ProjectMember{}, &gitlab.Response{}, nil
					},
				},
				cr: projectMember(
					withExternalName(extName),
				),
			},
			want: want{
				cr: projectMember(
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
				projectMember: &fake.MockClient{
					MockGetProjectMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
						return &gitlab.ProjectMember{
							AccessLevel: accessLevel,
						}, &gitlab.Response{}, nil
					},
				},
				cr: projectMember(
					withExternalName(extName),
					withAccessLevel(10),
				),
			},
			want: want{
				cr: projectMember(
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
				projectMember: &fake.MockClient{
					MockGetProjectMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
						return &gitlab.ProjectMember{
							ExpiresAt: &expiresAt,
						}, &gitlab.Response{}, nil
					},
				},
				cr: projectMember(
					withExternalName(extName),
					withExpiresAt(expiresAtNew.String()),
				),
			},
			want: want{
				cr: projectMember(
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
			e := &external{kube: tc.kube, client: tc.projectMember}
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
				err: errors.New(errNotProjectMember),
			},
		},
		"SuccessfulCreationWithoutExpiresAt": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				projectMember: &fake.MockClient{
					MockAddProjectMember: func(gid interface{}, opt *gitlab.AddProjectMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
						return &gitlab.ProjectMember{
							ID:          ID,
							Username:    username,
							Email:       email,
							Name:        name,
							State:       state,
							AvatarURL:   avatarURL,
							WebURL:      webURL,
							AccessLevel: accessLevel,
						}, &gitlab.Response{}, nil
					},
				},
				cr: projectMember(
					withAnnotations(extNameAnnotation),
					withSpec(v1alpha1.ProjectMemberParameters{ProjectID: &ID}),
				),
			},
			want: want{
				cr: projectMember(
					withExternalName(extName),
					withSpec(v1alpha1.ProjectMemberParameters{ProjectID: &ID}),
				),
				result: managed.ExternalCreation{ExternalNameAssigned: true},
			},
		},
		"SuccessfulCreationWithExpiresAt": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				projectMember: &fake.MockClient{
					MockAddProjectMember: func(gid interface{}, opt *gitlab.AddProjectMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
						return &gitlab.ProjectMember{
							ID:          ID,
							Username:    username,
							Email:       email,
							Name:        name,
							State:       state,
							AvatarURL:   avatarURL,
							WebURL:      webURL,
							AccessLevel: accessLevel,
							ExpiresAt:   &expiresAt,
						}, &gitlab.Response{}, nil
					},
				},
				cr: projectMember(
					withAnnotations(extNameAnnotation),
					withSpec(v1alpha1.ProjectMemberParameters{ProjectID: &ID}),
				),
			},
			want: want{
				cr: projectMember(
					withExternalName(extName),
					withSpec(v1alpha1.ProjectMemberParameters{ProjectID: &ID}),
				),
				result: managed.ExternalCreation{ExternalNameAssigned: true},
			},
		},
		"FailedCreation": {
			args: args{
				projectMember: &fake.MockClient{
					MockAddProjectMember: func(gid interface{}, opt *gitlab.AddProjectMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
						return &gitlab.ProjectMember{}, &gitlab.Response{}, errBoom
					},
				},
				cr: projectMember(
					withAnnotations(extNameAnnotation),
					withSpec(v1alpha1.ProjectMemberParameters{ProjectID: &ID}),
				),
			},
			want: want{
				cr: projectMember(
					withExternalName(extName),
					withSpec(v1alpha1.ProjectMemberParameters{ProjectID: &ID}),
				),
				err: errors.Wrap(errBoom, errCreateFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.projectMember}
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
				err: errors.New(errNotProjectMember),
			},
		},
		"SuccessfulUpdate": {
			args: args{
				projectMember: &fake.MockClient{
					MockEditProjectMember: func(gid interface{}, user int, opt *gitlab.EditProjectMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
						return &gitlab.ProjectMember{
							ID:          ID,
							Username:    username,
							Email:       email,
							Name:        name,
							State:       state,
							AvatarURL:   avatarURL,
							WebURL:      webURL,
							AccessLevel: accessLevel,
						}, &gitlab.Response{}, nil
					},
				},
				cr: projectMember(
					withExternalName(extName),
					withStatus(v1alpha1.ProjectMemberObservation{Username: "new username"}),
				),
			},
			want: want{
				cr: projectMember(
					withExternalName(extName),
					withStatus(v1alpha1.ProjectMemberObservation{Username: "new username"}),
				),
			},
		},
		"FailedUpdate": {
			args: args{
				projectMember: &fake.MockClient{
					MockEditProjectMember: func(gid interface{}, user int, opt *gitlab.EditProjectMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
						return &gitlab.ProjectMember{}, &gitlab.Response{}, errBoom
					},
				},
				cr: projectMember(),
			},
			want: want{
				cr:  projectMember(),
				err: errors.Wrap(errBoom, errUpdateFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.projectMember}
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
				err: errors.New(errNotProjectMember),
			},
		},
		"SuccessfulDeletion": {
			args: args{
				projectMember: &fake.MockClient{
					MockDeleteProjectMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: projectMember(withExternalName(extName)),
			},
			want: want{
				cr:  projectMember(withExternalName(extName)),
				err: nil,
			},
		},
		"FailedDeletion": {
			args: args{
				projectMember: &fake.MockClient{
					MockDeleteProjectMember: func(gid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: projectMember(),
			},
			want: want{
				cr:  projectMember(),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.projectMember}
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
