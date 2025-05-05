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
	"net/http"
	"reflect"
	"testing"
	"time"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups/fake"
)

var (
	unexpecedItem      resource.Managed
	path               = "path/to/group"
	name               = "example-group"
	displayName        = "Example Group"
	groupAccessLevel   = 40
	groupID            = 1234
	groupIDtwo         = 123456
	extName            = "1234"
	errBoom            = errors.New("boom")
	expiresAt          = time.Now()
	expiresAtIso       = (gitlab.ISOTime)(expiresAt)
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

func withName(s string) groupModifier {
	return func(r *v1alpha1.Group) { r.Spec.ForProvider.Name = &s }
}

func withPath(s string) groupModifier {
	return func(r *v1alpha1.Group) { r.Spec.ForProvider.Path = s }
}

func withPermanentlyRemove(b *bool) groupModifier {
	return func(r *v1alpha1.Group) { r.Spec.ForProvider.PermanentlyRemove = b }
}

func withFullPathToRemove(s *string) groupModifier {
	return func(r *v1alpha1.Group) { r.Spec.ForProvider.FullPathToRemove = s }
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
		s := ""
		p.Spec.ForProvider = v1alpha1.GroupParameters{
			MembershipLock:                 &f,
			ShareWithGroupLock:             &f,
			RequireTwoFactorAuth:           &f,
			TwoFactorGracePeriod:           &i,
			AutoDevopsEnabled:              &f,
			EmailsEnabled:                  &f,
			MentionsDisabled:               &f,
			LFSEnabled:                     &f,
			RequestAccessEnabled:           &f,
			ParentID:                       &i,
			SharedRunnersMinutesLimit:      &i,
			ExtraSharedRunnersMinutesLimit: &i,
		}
		p.Status.AtProvider = v1alpha1.GroupObservation{
			ID:        &i,
			AvatarURL: &s,
			WebURL:    &s,
			FullName:  &s,
			FullPath:  &s,
			LDAPCN:    &s,
		}
	}
}

func withStatus(s v1alpha1.GroupObservation) groupModifier {
	return func(r *v1alpha1.Group) { r.Status.AtProvider = s }
}

func withAnnotations(a map[string]string) groupModifier {
	return func(p *v1alpha1.Group) { meta.AddAnnotations(p, a) }
}

func withSharedWithGroups(s []v1alpha1.SharedWithGroups) groupModifier {
	return func(g *v1alpha1.Group) { g.Spec.ForProvider.SharedWithGroups = s }
}

func withSharedWithGroupsObservation(s []v1alpha1.SharedWithGroupsObservation) groupModifier {
	return func(g *v1alpha1.Group) { g.Status.AtProvider.SharedWithGroups = s }
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
	gitlabSubGroupCreationLevelNew := gitlab.Ptr("owner")
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
				err: errors.New(errIDNotInt),
			},
		},
		"FailedGetRequest": {
			args: args{
				group: &fake.MockClient{
					MockGetGroup: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						return &gitlab.Group{}, &gitlab.Response{Response: &http.Response{StatusCode: 400}}, errBoom
					},
				},
				cr: group(withExternalName(extName)),
			},
			want: want{
				cr:  group(withAnnotations(extNameAnnotation)),
				err: errors.Wrap(errBoom, errGetFailed),
			},
		},
		"ErrGet404": {
			args: args{
				group: &fake.MockClient{
					MockGetGroup: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						return &gitlab.Group{}, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errBoom
					},
				},
				cr: group(withExternalName(extName)),
			},
			want: want{
				cr:  group(withAnnotations(extNameAnnotation)),
				err: nil,
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
							Path:         "",
							RunnersToken: "token",
							SharedWithGroups: []gitlab.SharedWithGroup{
								{
									GroupID:          groupID,
									GroupName:        name,
									GroupFullPath:    path,
									GroupAccessLevel: 40,
									ExpiresAt:        &expiresAtIso,
								},
							},
						}, &gitlab.Response{}, nil
					},
				},
				cr: group(
					withClientDefaultValues(),
					withExternalName(extName),
					withSharedWithGroups(
						[]v1alpha1.SharedWithGroups{
							{
								GroupID:          &groupID,
								GroupAccessLevel: 40,
							},
						},
					),
				),
			},
			want: want{
				cr: group(
					withConditions(xpv1.Available()),
					withPath(path),
					withAnnotations(extNameAnnotation),
					withStatus(v1alpha1.GroupObservation{}),
					withClientDefaultValues(),
					withSharedWithGroupsObservation(
						[]v1alpha1.SharedWithGroupsObservation{
							{
								GroupID:          &groupID,
								GroupName:        &name,
								GroupFullPath:    &path,
								GroupAccessLevel: &groupAccessLevel,
								ExpiresAt:        &metav1.Time{Time: time.Time(expiresAtIso)},
							},
						},
					),
					withSharedWithGroups(
						[]v1alpha1.SharedWithGroups{
							{
								GroupID:          &groupID,
								GroupAccessLevel: 40,
								ExpiresAt:        &metav1.Time{Time: expiresAt},
							},
						},
					),
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
							SubGroupCreationLevel:          gitlab.SubGroupCreationLevelValue(*gitlabSubGroupCreationLevelNew),
							MembershipLock:                 false,
							ShareWithGroupLock:             false,
							RequireTwoFactorAuth:           false,
							TwoFactorGracePeriod:           0,
							AutoDevopsEnabled:              false,
							EmailsEnabled:                  false,
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
		"SubGroupCreationLevel":          gitlab.SubGroupCreationLevelValue(*gitlabSubGroupCreationLevelNew),
		"Visibility":                     gitlabVisibilityNew,
		"ShareWithGroupLock":             true,
		"RequireTwoFactorAuth":           true,
		"TwoFactorGracePeriod":           1,
		"AutoDevopsEnabled":              true,
		"EmailsEnabled":                  true,
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

		if name == "Name" {
			argsGroupModifier = append(argsGroupModifier, withName(displayName))
			wantGroupModifier = append(wantGroupModifier, withName(displayName))
		}

		if name == "Path" {
			argsGroupModifier = append(argsGroupModifier, withPath(path))
			wantGroupModifier = append(wantGroupModifier, withPath(path))
		}

		if name == "Description" {
			argsGroupModifier = append(argsGroupModifier, withDescription(&description))
			wantGroupModifier = append(wantGroupModifier, withDescription(&description))
		}

		gitlabGroup := &gitlab.Group{
			Name:                  name,
			Visibility:            gitlab.VisibilityValue(visibility),
			ProjectCreationLevel:  *gitlab.ProjectCreationLevel(gitlab.ProjectCreationLevelValue(projectCreationLevel)),
			SubGroupCreationLevel: *gitlab.Ptr(gitlab.SubGroupCreationLevelValue(subGroupCreationLevel)),
		}
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
						return &gitlab.Group{Name: extName, Path: extName, ID: groupID}, &gitlab.Response{}, nil
					},
				},
				cr: group(withAnnotations(extNameAnnotation)),
			},
			want: want{
				cr:     group(withExternalName(extName)),
				result: managed.ExternalCreation{},
			},
		},
		"FailedCreation": {
			args: args{
				group: &fake.MockClient{
					MockCreateGroup: func(opt *gitlab.CreateGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						return &gitlab.Group{}, &gitlab.Response{}, errBoom
					},
				},
				cr: group(withStatus(v1alpha1.GroupObservation{ID: &groupID})),
			},
			want: want{
				cr:  group(withStatus(v1alpha1.GroupObservation{ID: &groupID})),
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
				cr: group(withStatus(v1alpha1.GroupObservation{ID: &groupID}), withExternalName("1234")),
			},
			want: want{
				cr: group(
					withStatus(v1alpha1.GroupObservation{ID: &groupID}),
					withExternalName("1234"),
				),
			},
		},
		"SharedWithGroups": {
			args: args{
				group: &fake.MockClient{
					MockUpdateGroup: func(pid interface{}, opt *gitlab.UpdateGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						return &gitlab.Group{
							ID: groupID,
							SharedWithGroups: []gitlab.SharedWithGroup{
								{
									GroupID: groupID,
								},
							},
						}, nil, nil
					},
					MockShareGroupWithGroup: func(gid interface{}, opt *gitlab.ShareGroupWithGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						return nil, nil, nil
					},
				},
				cr: group(
					withStatus(v1alpha1.GroupObservation{ID: &groupID}),
					withSharedWithGroups([]v1alpha1.SharedWithGroups{
						{
							GroupID:          &groupID,
							GroupAccessLevel: 40,
							ExpiresAt:        &metav1.Time{Time: expiresAt},
						},
						{
							GroupID: &groupIDtwo,
						},
					}),
				),
			},
			want: want{
				cr: group(
					withStatus(v1alpha1.GroupObservation{ID: &groupID}),
					withSharedWithGroups([]v1alpha1.SharedWithGroups{
						{
							GroupID:          &groupID,
							GroupAccessLevel: 40,
							ExpiresAt:        &metav1.Time{Time: expiresAt},
						},
						{
							GroupID: &groupIDtwo,
						},
					}),
				),
				result: managed.ExternalUpdate{},
				err:    nil,
			},
		},
		"UnsharedWithGroups": {
			args: args{
				group: &fake.MockClient{
					MockUpdateGroup: func(pid interface{}, opt *gitlab.UpdateGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						return &gitlab.Group{
							SharedWithGroups: []gitlab.SharedWithGroup{
								{GroupID: groupID},
								{GroupID: 123456},
							},
						}, nil, nil
					},
					MockUnshareGroupFromGroup: func(gid interface{}, groupID int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return nil, nil
					},
				},
				cr: group(
					withSharedWithGroups([]v1alpha1.SharedWithGroups{
						{GroupID: &groupIDtwo},
					}),
				),
			},
			want: want{
				cr: group(
					withSharedWithGroups([]v1alpha1.SharedWithGroups{
						{GroupID: &groupIDtwo},
					}),
				),
				result: managed.ExternalUpdate{},
				err:    nil,
			},
		},
		"SharedWithGroupsFailed": {
			args: args{
				group: &fake.MockClient{
					MockShareGroupWithGroup: func(gid interface{}, opt *gitlab.ShareGroupWithGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						return nil, nil, errBoom
					},
					MockUpdateGroup: func(pid interface{}, opt *gitlab.UpdateGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						return &gitlab.Group{}, nil, nil
					},
				},
				cr: group(
					withSharedWithGroups(
						[]v1alpha1.SharedWithGroups{
							{
								GroupID: &groupID,
							},
						},
					),
				),
			},
			want: want{
				cr: group(
					withSharedWithGroups(
						[]v1alpha1.SharedWithGroups{
							{
								GroupID: &groupID,
							},
						},
					),
				),
				err:    errors.Wrapf(errBoom, errShareFailed, groupID),
				result: managed.ExternalUpdate{},
			},
		},
		"UnsharedWithGroupsFailed": {
			args: args{
				group: &fake.MockClient{
					MockUpdateGroup: func(pid interface{}, opt *gitlab.UpdateGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						return &gitlab.Group{
							SharedWithGroups: []gitlab.SharedWithGroup{
								{GroupID: groupID},
								{GroupID: 123456},
							},
						}, nil, nil
					},
					MockUnshareGroupFromGroup: func(gid interface{}, groupID int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return nil, errBoom
					},
				},
				cr: group(),
			},
			want: want{
				cr:     group(),
				result: managed.ExternalUpdate{},
				err:    errors.Wrapf(errBoom, errUnshareFailed, groupID),
			},
		},
		"FailedUpdate": {
			args: args{
				group: &fake.MockClient{
					MockUpdateGroup: func(pid interface{}, opt *gitlab.UpdateGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						return &gitlab.Group{}, &gitlab.Response{}, errBoom
					},
				},
				cr: group(withStatus(v1alpha1.GroupObservation{ID: &groupID})),
			},
			want: want{
				cr:  group(withStatus(v1alpha1.GroupObservation{ID: &groupID})),
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
	type deleteGroupCalls struct {
		Pid interface{}
		Opt *gitlab.DeleteGroupOptions
	}
	var recordedCalls []deleteGroupCalls

	type want struct {
		cr    resource.Managed
		calls []deleteGroupCalls
		err   error
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
					MockDeleteGroup: func(pid interface{}, opt *gitlab.DeleteGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						recordedCalls = append(recordedCalls, deleteGroupCalls{Pid: pid, Opt: opt})
						return &gitlab.Response{}, nil
					},
				},
				cr: group(withExternalName("0")),
			},
			want: want{
				cr:    group(withExternalName("0")),
				calls: []deleteGroupCalls{{Pid: "0", Opt: &gitlab.DeleteGroupOptions{}}},
				err:   nil,
			},
		},
		"FailedDeletion": {
			args: args{
				group: &fake.MockClient{
					MockDeleteGroup: func(pid interface{}, opt *gitlab.DeleteGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return nil, errBoom
					},
				},
				cr: group(),
			},
			want: want{
				cr:  group(),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
		"SuccessfulPermanentlyDeletion": {
			args: args{
				group: &fake.MockClient{
					MockDeleteGroup: func(pid interface{}, opt *gitlab.DeleteGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						recordedCalls = append(recordedCalls, deleteGroupCalls{Pid: pid, Opt: opt})
						return &gitlab.Response{}, nil
					},
				},
				cr: group(
					withExternalName("0"),
					withPermanentlyRemove(gitlab.Ptr(true)),
					withPath("group"),
					withFullPathToRemove(gitlab.Ptr("path/to/group")),
					withStatus(v1alpha1.GroupObservation{FullPath: gitlab.Ptr("path/to/group")})),
			},
			want: want{
				cr: group(
					withExternalName("0"),
					withPermanentlyRemove(gitlab.Ptr(true)),
					withPath("group"),
					withFullPathToRemove(gitlab.Ptr("path/to/group")),
					withStatus(v1alpha1.GroupObservation{FullPath: gitlab.Ptr("path/to/group")})),
				calls: []deleteGroupCalls{
					{Pid: "0", Opt: &gitlab.DeleteGroupOptions{}},
					{Pid: "0", Opt: &gitlab.DeleteGroupOptions{PermanentlyRemove: gitlab.Ptr(true), FullPath: gitlab.Ptr("path/to/group")}},
				},
				err: nil,
			},
		},
		"SuccessfulPermanentlyTopLevelGroupDeletion": {
			args: args{
				group: &fake.MockClient{
					MockDeleteGroup: func(pid interface{}, opt *gitlab.DeleteGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						recordedCalls = append(recordedCalls, deleteGroupCalls{Pid: pid, Opt: opt})
						return &gitlab.Response{}, nil
					},
				},
				cr: group(
					withExternalName("0"),
					withPermanentlyRemove(gitlab.Ptr(true)),
					withPath("top-level-group"),
					withStatus(v1alpha1.GroupObservation{FullPath: gitlab.Ptr("top-level-group")})),
			},
			want: want{
				cr: group(
					withExternalName("0"),
					withPermanentlyRemove(gitlab.Ptr(true)),
					withPath("top-level-group"),
					withStatus(v1alpha1.GroupObservation{FullPath: gitlab.Ptr("top-level-group")}),
				),
				calls: []deleteGroupCalls{
					{Pid: "0", Opt: &gitlab.DeleteGroupOptions{}},
				},
				err: nil,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			recordedCalls = nil
			e := &external{kube: tc.kube, client: tc.group}
			_, err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.calls, recordedCalls, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
