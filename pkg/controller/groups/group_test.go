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

package groups

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups/fake"
)

var (
	unexpecedItem      resource.Managed
	path               = "path/to/group"
	name               = "example-group"
	groupID            = 1234
	extName            = strconv.Itoa(groupID)
	errBoom            = errors.New("boom")
	extNameAnnotation  = map[string]string{meta.AnnotationKeyExternalName: extName}
	visibility         = "private"
	v1alpha1Visibility = v1alpha1.VisibilityValue(visibility)

	projectCreationLevel         = "developer"
	v1alpha1ProjectCreationLevel = v1alpha1.ProjectCreationLevelValue(projectCreationLevel)

	subGroupCreationLevel         = "maintainer"
	v1alpha1SubGroupCreationLevel = v1alpha1.SubGroupCreationLevelValue(subGroupCreationLevel)
)

type args struct {
	group groups.Client
	kube  client.Client
	cr    resource.Managed
}

type groupModifier func(*v1alpha1.Group)

func withConditions(c ...xpv1.Condition) groupModifier {
	return func(cr *v1alpha1.Group) { cr.Status.ConditionedStatus.Conditions = c }
}

func withPath(s string) groupModifier {
	return func(r *v1alpha1.Group) { r.Spec.ForProvider.Path = s }
}

func withDescription(s *string) groupModifier {
	return func(r *v1alpha1.Group) { r.Spec.ForProvider.Description = s }
}

func withProjectCreationLevel(s *v1alpha1.ProjectCreationLevelValue) groupModifier {
	return func(r *v1alpha1.Group) { r.Spec.ForProvider.ProjectCreationLevel = s }
}

func withVisibility(s *v1alpha1.VisibilityValue) groupModifier {
	return func(r *v1alpha1.Group) { r.Spec.ForProvider.Visibility = s }
}

func withSubGroupCreationLevel(s *v1alpha1.SubGroupCreationLevelValue) groupModifier {
	return func(r *v1alpha1.Group) { r.Spec.ForProvider.SubGroupCreationLevel = s }
}

func withExternalName(n string) groupModifier {
	return func(r *v1alpha1.Group) { meta.SetExternalName(r, n) }
}

// Use for testing. When ResourceLateInitialized should it be false
func withClientDefaultValues() groupModifier {
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

func withStatus(s v1alpha1.GroupObservation) groupModifier {
	return func(r *v1alpha1.Group) { r.Status.AtProvider = s }
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

func TestConnect(t *testing.T) {
	type want struct {
		cr     resource.Managed
		result managed.ExternalClient
		err    error
		// client newGitlabClientFn
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
				err: errors.New(errNotGroup),
			},
		},
		"ProviderConfigRefNotGivenError": {
			args: args{
				cr:   group(),
				kube: &test.MockClient{MockGet: test.NewMockGetFn(nil)},
			},
			want: want{
				cr:  group(),
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
	description := "description"

	gitlabVisibilityNew := gitlab.VisibilityValue("public")
	v1alpha1VisibilityNew := v1alpha1.VisibilityValue("public")
	gitlabProjectCreationLevelNew := gitlab.ProjectCreationLevel("noone")
	v1alpha1ProjectCreationLevelNew := v1alpha1.ProjectCreationLevelValue("noone")
	gitlabSubGroupCreationLevelNew := gitlab.SubGroupCreationLevel("owner")
	v1alpha1SubGroupCreationLevelNew := v1alpha1.SubGroupCreationLevelValue("owner")

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
				err: errors.New(errNotGroup),
			},
		},
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
					withClientDefaultValues(),
					withExternalName(extName),
				),
			},
			want: want{
				cr: group(
					withClientDefaultValues(),
					withConditions(xpv1.Available()),
					withPath(path),
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
		"LateInitialized": {
			args: args{
				group: &fake.MockClient{
					MockGetGroup: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						return &gitlab.Group{
							Path:                           path,
							Description:                    description,
							Visibility:                     gitlabVisibilityNew,
							ProjectCreationLevel:           *gitlabProjectCreationLevelNew,
							SubGroupCreationLevel:          *gitlabSubGroupCreationLevelNew,
							MembershipLock:                 false,
							ShareWithGroupLock:             false,
							RequireTwoFactorAuth:           false,
							TwoFactorGracePeriod:           0,
							AutoDevopsEnabled:              false,
							EmailsDisabled:                 false,
							MentionsDisabled:               false,
							LFSEnabled:                     false,
							RequestAccessEnabled:           false,
							ParentID:                       0,
							SharedRunnersMinutesLimit:      0,
							ExtraSharedRunnersMinutesLimit: 0,
						}, &gitlab.Response{}, nil
					},
				},
				cr: group(withExternalName("0")),
			},
			want: want{
				cr: group(
					withExternalName("0"),
					withClientDefaultValues(),
					withPath(path),
					withConditions(xpv1.Available()),
					withDescription(&description),
					withVisibility(&v1alpha1VisibilityNew),
					withProjectCreationLevel(&v1alpha1ProjectCreationLevelNew),
					withSubGroupCreationLevel(&v1alpha1SubGroupCreationLevelNew),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
					ConnectionDetails:       managed.ConnectionDetails{"runnersToken": []byte("")},
				},
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
					withPath(""),
					withClientDefaultValues(),
					withExternalName(extName),
				),
			},
			want: want{
				cr: group(
					withPath(""),
					withClientDefaultValues(),
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
	}

	isGroupUpToDateCases := map[string]interface{}{
		"Name":                           "name",
		"Path":                           "/new/group/path",
		"Description":                    "description v2",
		"MembershipLock":                 true,
		"ProjectCreationLevel":           *gitlabProjectCreationLevelNew,
		"SubGroupCreationLevel":          *gitlabSubGroupCreationLevelNew,
		"Visibility":                     gitlabVisibilityNew,
		"ShareWithGroupLock":             true,
		"RequireTwoFactorAuth":           true,
		"TwoFactorGracePeriod":           1,
		"AutoDevopsEnabled":              true,
		"EmailsDisabled":                 true,
		"MentionsDisabled":               true,
		"LFSEnabled":                     true,
		"RequestAccessEnabled":           true,
		"ParentID":                       1,
		"SharedRunnersMinutesLimit":      1,
		"ExtraSharedRunnersMinutesLimit": 1,
	}

	for name, value := range isGroupUpToDateCases {
		argsGroupModifier := []groupModifier{
			withClientDefaultValues(),
			withExternalName("0"),
			withVisibility(&v1alpha1Visibility),
			withProjectCreationLevel(&v1alpha1ProjectCreationLevel),
			withSubGroupCreationLevel(&v1alpha1SubGroupCreationLevel),
		}
		wantGroupModifier := []groupModifier{
			withClientDefaultValues(),
			withExternalName("0"),
			withConditions(xpv1.Available()),
			withVisibility(&v1alpha1Visibility),
			withProjectCreationLevel(&v1alpha1ProjectCreationLevel),
			withSubGroupCreationLevel(&v1alpha1SubGroupCreationLevel),
		}

		if name == "Path" {
			argsGroupModifier = append(argsGroupModifier, withPath(path))
			wantGroupModifier = append(wantGroupModifier, withPath(path))
		}

		if name == "Description" {
			argsGroupModifier = append(argsGroupModifier, withDescription(&description))
			wantGroupModifier = append(wantGroupModifier, withDescription(&description))
		}

		gitlabGroup := &gitlab.Group{Name: name}
		structValue := reflect.ValueOf(gitlabGroup).Elem()
		structFieldValue := structValue.FieldByName(name)
		val := reflect.ValueOf(value)

		structFieldValue.Set(val)
		cases["IsGroupUpToDate"+name] = struct {
			args
			want
		}{
			args: args{
				group: &fake.MockClient{
					MockGetGroup: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						return gitlabGroup, &gitlab.Response{}, nil
					},
				},
				cr: group(argsGroupModifier...),
			},
			want: want{
				cr: group(wantGroupModifier...),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
					ConnectionDetails:       managed.ConnectionDetails{"runnersToken": []byte("")},
				},
			},
		}
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
				err: errors.New(errNotGroup),
			},
		},
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
				cr: group(withAnnotations(extNameAnnotation)),
			},
			want: want{
				cr:     group(withExternalName("0")),
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
				cr: group(withStatus(v1alpha1.GroupObservation{ID: 0})),
			},
			want: want{
				cr:  group(withStatus(v1alpha1.GroupObservation{ID: 0})),
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
				err: errors.New(errNotGroup),
			},
		},
		"SuccessfulUpdate": {
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
		"FailedUpdate": {
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
				err: errors.New(errNotGroup),
			},
		},
		"SuccessfulDeletion": {
			args: args{
				group: &fake.MockClient{
					MockDeleteGroup: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: group(withExternalName("0")),
			},
			want: want{
				cr:  group(withExternalName("0")),
				err: nil,
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
				cr:  group(),
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
