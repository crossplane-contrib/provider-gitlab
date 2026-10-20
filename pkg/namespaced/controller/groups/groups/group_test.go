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
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/groups"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/groups/fake"
)

var (
	unexpecedItem      resource.Managed
	path               = "path/to/group"
	name               = "example-group"
	displayName        = "Example Group"
	groupAccessLevel   = int64(40)
	groupID            = int64(1234)
	groupIDtwo         = int64(123456)
	extName            = "1234"
	errBoom            = errors.New("boom")
	expiresAt          = time.Now()
	expiresAtIso       = (gitlab.ISOTime)(expiresAt)
	extNameAnnotation  = map[string]string{meta.AnnotationKeyExternalName: extName}
	fullPath           = "path/to/example-group"
	visibility         = "private"
	v1alpha1Visibility = v1alpha1.VisibilityValue(visibility)

	timeNow = time.Now()

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

func withConditions(c ...v2.Condition) groupModifier {
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

func withManagementPolicies(p ...v2.ManagementAction) groupModifier {
	return func(r *v1alpha1.Group) { r.SetManagementPolicies(p) }
}

// Use for testing. When ResourceLateInitialized should it be false
func withClientDefaultValues() groupModifier {
	return func(p *v1alpha1.Group) {
		f := false
		i64 := int64(0)
		s := ""
		p.Spec.ForProvider = v1alpha1.GroupParameters{
			MembershipLock:                 &f,
			ShareWithGroupLock:             &f,
			RequireTwoFactorAuth:           &f,
			TwoFactorGracePeriod:           nil,
			AutoDevopsEnabled:              &f,
			EmailsEnabled:                  &f,
			MentionsDisabled:               &f,
			LFSEnabled:                     &f,
			RequestAccessEnabled:           &f,
			ParentID:                       nil,
			SharedRunnersMinutesLimit:      nil,
			ExtraSharedRunnersMinutesLimit: nil,
		}
		p.Status.AtProvider = v1alpha1.GroupObservation{
			ID:        &i64,
			AvatarURL: &s,
			WebURL:    &s,
			FullName:  &s,
			FullPath:  &s,
			LDAPCN:    &s,
			Archived:  &f,
		}
	}
}

func withArchived(b *bool) groupModifier {
	return func(r *v1alpha1.Group) { r.Spec.ForProvider.Archived = b }
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
	gitlabProjectCreationLevelNew := gitlab.NoOneProjectCreation
	v1alpha1ProjectCreationLevelNew := v1alpha1.ProjectCreationLevelValue("noone")
	gitlabSubGroupCreationLevelNew := new("owner")
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
		"PathExternalNameObserveOnly": {
			args: args{
				group: &fake.MockClient{
					MockGetGroup: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						if pid != fullPath {
							return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404, Header: http.Header{"Content-Type": {"application/json"}}}}, errBoom
						}
						return &gitlab.Group{ID: groupID, Name: name}, &gitlab.Response{}, nil
					},
				},
				cr: group(
					withPath(""),
					withClientDefaultValues(),
					withExternalName(fullPath),
					withManagementPolicies(v2.ManagementActionObserve),
				),
			},
			want: want{
				cr: group(
					withPath(""),
					withClientDefaultValues(),
					withConditions(v2.Available()),
					withExternalName(fullPath),
					withManagementPolicies(v2.ManagementActionObserve),
					func(g *v1alpha1.Group) { g.Status.AtProvider.ID = &groupID },
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: false,
					ConnectionDetails:       managed.ConnectionDetails{"runnersToken": []byte("")},
				},
			},
		},
		"PathExternalNameNotFound": {
			args: args{
				group: &fake.MockClient{
					MockGetGroup: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404, Header: http.Header{"Content-Type": {"application/json"}}}}, errBoom
					},
				},
				cr: group(withExternalName(fullPath), withManagementPolicies(v2.ManagementActionObserve)),
			},
			want: want{
				cr:     group(withExternalName(fullPath), withManagementPolicies(v2.ManagementActionObserve)),
				result: managed.ExternalObservation{},
			},
		},
		"PathExternalNameWithWritePolicies": {
			args: args{
				group: &fake.MockClient{},
				cr:    group(withExternalName(fullPath), withManagementPolicies(v2.ManagementActionAll)),
			},
			want: want{
				cr:  group(withExternalName(fullPath), withManagementPolicies(v2.ManagementActionAll)),
				err: errors.New(common.ErrPathExternalNameNotObserveOnly),
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
						return &gitlab.Group{}, &gitlab.Response{Response: &http.Response{StatusCode: 404, Header: http.Header{"Content-Type": {"application/json"}}}}, errBoom
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
					withConditions(v2.Available()),
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
							ProjectCreationLevel:           gitlabProjectCreationLevelNew,
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
					withConditions(v2.Available()),
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
						if pid != extName {
							return nil, &gitlab.Response{}, errBoom
						}
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
					withConditions(v2.Available()),
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
		"DeleteOnPendingDeletion": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				group: &fake.MockClient{
					MockGetGroup: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						return &gitlab.Group{
							Name:                name,
							MarkedForDeletionOn: &gitlab.ISOTime{},
						}, &gitlab.Response{}, nil
					},
				},
				cr: group(
					withClientDefaultValues(),
					func(p *v1alpha1.Group) {
						p.DeletionTimestamp = &metav1.Time{Time: timeNow}
						p.Spec.ForProvider.RemoveFinalizerOnPendingDeletion = ptr.To(true)
					},
					withExternalName(extName),
				),
			},
			want: want{
				cr: group(
					withClientDefaultValues(),
					withExternalName(extName),
					func(p *v1alpha1.Group) {
						p.DeletionTimestamp = &metav1.Time{Time: timeNow}
						p.Spec.ForProvider.RemoveFinalizerOnPendingDeletion = ptr.To(true)
					},
				),
				result: managed.ExternalObservation{
					ResourceExists:          false,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
			},
		},
	}

	isGroupUpToDateCases := map[string]interface{}{
		"Name":                           "name",
		"Path":                           "/new/group/path",
		"Description":                    "description v2",
		"MembershipLock":                 true,
		"ProjectCreationLevel":           gitlabProjectCreationLevelNew,
		"SubGroupCreationLevel":          gitlab.SubGroupCreationLevelValue(*gitlabSubGroupCreationLevelNew),
		"Visibility":                     gitlabVisibilityNew,
		"ShareWithGroupLock":             true,
		"RequireTwoFactorAuth":           true,
		"TwoFactorGracePeriod":           int64(1),
		"AutoDevopsEnabled":              true,
		"EmailsEnabled":                  true,
		"MentionsDisabled":               true,
		"LFSEnabled":                     true,
		"RequestAccessEnabled":           true,
		"ParentID":                       int64(1),
		"SharedRunnersMinutesLimit":      int64(1),
		"ExtraSharedRunnersMinutesLimit": int64(1),
		"Archived":                       true,
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
			withConditions(v2.Available()),
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

		if name == "Archived" {
			argsGroupModifier = append(argsGroupModifier, withArchived(new(false)))
			wantGroupModifier = append(wantGroupModifier, withArchived(new(false)), func(g *v1alpha1.Group) {
				g.Status.AtProvider.Archived = new(true)
			})
		}

		gitlabGroup := &gitlab.Group{
			Name:                  name,
			Visibility:            gitlab.VisibilityValue(visibility),
			ProjectCreationLevel:  gitlab.ProjectCreationLevelValue(projectCreationLevel),
			SubGroupCreationLevel: gitlab.SubGroupCreationLevelValue(subGroupCreationLevel),
		}
		structValue := reflect.ValueOf(gitlabGroup).Elem()
		structFieldValue := structValue.FieldByName(name)
		val := reflect.ValueOf(value)

		structFieldValue.Set(val)

		// Special handling for runner minute limit fields - they get late-initialized when > 0
		expectedUpToDate := false
		expectedLateInitialized := false
		if (name == "SharedRunnersMinutesLimit" || name == "ExtraSharedRunnersMinutesLimit") && value.(int64) > 0 {
			expectedUpToDate = true
			expectedLateInitialized = true
			// Add the expected value to wantGroupModifier
			if name == "SharedRunnersMinutesLimit" {
				wantGroupModifier = append(wantGroupModifier, func(p *v1alpha1.Group) {
					v := value.(int64)
					p.Spec.ForProvider.SharedRunnersMinutesLimit = &v
				})
			} else {
				wantGroupModifier = append(wantGroupModifier, func(p *v1alpha1.Group) {
					v := value.(int64)
					p.Spec.ForProvider.ExtraSharedRunnersMinutesLimit = &v
				})
			}
		}

		// TwoFactorGracePeriod and ParentID also get late-initialized when > 0
		if (name == "TwoFactorGracePeriod" || name == "ParentID") && value.(int64) > 0 {
			expectedUpToDate = true
			expectedLateInitialized = true
			if name == "TwoFactorGracePeriod" {
				wantGroupModifier = append(wantGroupModifier, func(p *v1alpha1.Group) {
					v := value.(int64)
					p.Spec.ForProvider.TwoFactorGracePeriod = &v
				})
			} else {
				wantGroupModifier = append(wantGroupModifier, func(p *v1alpha1.Group) {
					v := value.(int64)
					p.Spec.ForProvider.ParentID = &v
				})
			}
		}

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
					ResourceUpToDate:        expectedUpToDate,
					ResourceLateInitialized: expectedLateInitialized,
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
					MockUnshareGroupFromGroup: func(gid interface{}, groupID int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
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
					MockUnshareGroupFromGroup: func(gid interface{}, groupID int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
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
	// Added to track the new GetGroup call
	type getGroupCalls struct {
		Pid interface{}
	}

	var recordedCalls []deleteGroupCalls
	var recordedGetCalls []getGroupCalls

	// The response GitLab gives when deleting a group that is already
	// scheduled for deletion.
	errGroupAlreadyMarked := &gitlab.ErrorResponse{
		StatusCode: http.StatusBadRequest,
		Response: &http.Response{
			StatusCode: http.StatusBadRequest,
			Request:    &http.Request{Method: http.MethodDelete, URL: &url.URL{Path: "/api/v4/groups/0"}},
		},
		Message: "{message: Group has already been marked for deletion}",
	}
	markedForDeletionOn := new(gitlab.ISOTime(time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)))

	type want struct {
		cr       resource.Managed
		calls    []deleteGroupCalls
		getCalls []getGroupCalls
		err      error
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
				cr:       group(withExternalName("0")),
				calls:    []deleteGroupCalls{{Pid: "0", Opt: &gitlab.DeleteGroupOptions{}}},
				getCalls: nil,
				err:      nil,
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
					MockGetGroup: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						recordedGetCalls = append(recordedGetCalls, getGroupCalls{Pid: pid})
						return &gitlab.Group{ID: 0, FullPath: "group-deletion_scheduled-0", MarkedForDeletionOn: markedForDeletionOn}, &gitlab.Response{}, nil
					},
				},
				cr: group(
					withExternalName("0"),
					withPermanentlyRemove(new(true)),
					withPath("group"),
					withFullPathToRemove(new("path/to/group")),
					withStatus(v1alpha1.GroupObservation{FullPath: new("path/to/group")})),
			},
			want: want{
				cr: group(
					withExternalName("0"),
					withPermanentlyRemove(new(true)),
					withPath("group"),
					withFullPathToRemove(new("path/to/group")),
					withStatus(v1alpha1.GroupObservation{FullPath: new("path/to/group")})),
				calls: []deleteGroupCalls{
					{Pid: "0", Opt: &gitlab.DeleteGroupOptions{}},
					{Pid: "0", Opt: &gitlab.DeleteGroupOptions{
						PermanentlyRemove: new(true), FullPath: new("group-deletion_scheduled-0"),
					}},
				},
				getCalls: []getGroupCalls{{Pid: "0"}},
				err:      nil,
			},
		},
		"SuccessfulPermanentlyDeletionWhenGroupNotFound": {
			args: args{
				group: &fake.MockClient{
					MockDeleteGroup: func(pid interface{}, opt *gitlab.DeleteGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						recordedCalls = append(recordedCalls, deleteGroupCalls{Pid: pid, Opt: opt})
						return &gitlab.Response{}, nil
					},
					MockGetGroup: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						recordedGetCalls = append(recordedGetCalls, getGroupCalls{Pid: pid})
						return nil, &gitlab.Response{Response: &http.Response{
							StatusCode: http.StatusNotFound,
							Header:     http.Header{"Content-Type": {"application/json"}},
						}}, fmt.Errorf("group not found: 404")
					},
				},
				cr: group(
					withExternalName("0"),
					withPermanentlyRemove(new(true)),
					withPath("group"),
					withFullPathToRemove(new("path/to/group")),
					withStatus(v1alpha1.GroupObservation{FullPath: new("path/to/group")})),
			},
			want: want{
				cr: group(
					withExternalName("0"),
					withPermanentlyRemove(new(true)),
					withPath("group"),
					withFullPathToRemove(new("path/to/group")),
					withStatus(v1alpha1.GroupObservation{FullPath: new("path/to/group")})),
				calls: []deleteGroupCalls{
					{Pid: "0", Opt: &gitlab.DeleteGroupOptions{}},
				},
				getCalls: []getGroupCalls{{Pid: "0"}},
				err:      nil,
			},
		},
		"FailedPermanentlyDeletionWhenGetGroupFails": {
			args: args{
				group: &fake.MockClient{
					MockDeleteGroup: func(pid interface{}, opt *gitlab.DeleteGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						recordedCalls = append(recordedCalls, deleteGroupCalls{Pid: pid, Opt: opt})
						return &gitlab.Response{}, nil
					},
					MockGetGroup: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						recordedGetCalls = append(recordedGetCalls, getGroupCalls{Pid: pid})
						return nil, nil, fmt.Errorf("some other problem")
					},
				},
				cr: group(
					withExternalName("0"),
					withPermanentlyRemove(new(true)),
					withPath("group"),
					withFullPathToRemove(new("path/to/group")),
					withStatus(v1alpha1.GroupObservation{FullPath: new("path/to/group")})),
			},
			want: want{
				cr: group(
					withExternalName("0"),
					withPermanentlyRemove(new(true)),
					withPath("group"),
					withFullPathToRemove(new("path/to/group")),
					withStatus(v1alpha1.GroupObservation{FullPath: new("path/to/group")})),
				calls: []deleteGroupCalls{
					{Pid: "0", Opt: &gitlab.DeleteGroupOptions{}},
				},
				getCalls: []getGroupCalls{{Pid: "0"}},
				err:      errors.Wrap(fmt.Errorf("some other problem"), "Cannot get group before deletion"),
			},
		},
		"FailedPermanentlyDeletionWhenGetGroupReturnsEmptyObject": {
			args: args{
				group: &fake.MockClient{
					MockDeleteGroup: func(pid interface{}, opt *gitlab.DeleteGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						recordedCalls = append(recordedCalls, deleteGroupCalls{Pid: pid, Opt: opt})
						return &gitlab.Response{}, nil
					},
					MockGetGroup: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						recordedGetCalls = append(recordedGetCalls, getGroupCalls{Pid: pid})
						return nil, &gitlab.Response{}, nil
					},
				},
				cr: group(
					withExternalName("0"),
					withPermanentlyRemove(new(true)),
					withPath("group"),
					withFullPathToRemove(new("path/to/group")),
					withStatus(v1alpha1.GroupObservation{FullPath: new("path/to/group")})),
			},
			want: want{
				cr: group(
					withExternalName("0"),
					withPermanentlyRemove(new(true)),
					withPath("group"),
					withFullPathToRemove(new("path/to/group")),
					withStatus(v1alpha1.GroupObservation{FullPath: new("path/to/group")})),
				calls: []deleteGroupCalls{
					{Pid: "0", Opt: &gitlab.DeleteGroupOptions{}},
				},
				getCalls: []getGroupCalls{{Pid: "0"}},
				err:      errors.New("GitLab API returned empty group object"),
			},
		},
		"AlreadyMarkedForDeletion": {
			args: args{
				group: &fake.MockClient{
					MockDeleteGroup: func(pid interface{}, opt *gitlab.DeleteGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						recordedCalls = append(recordedCalls, deleteGroupCalls{Pid: pid, Opt: opt})
						return nil, errGroupAlreadyMarked
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
		"AlreadyGone": {
			args: args{
				group: &fake.MockClient{
					MockDeleteGroup: func(pid interface{}, opt *gitlab.DeleteGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						recordedCalls = append(recordedCalls, deleteGroupCalls{Pid: pid, Opt: opt})
						return &gitlab.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errBoom
					},
				},
				cr: group(withExternalName("0"), withPermanentlyRemove(new(true)), withPath("group"),
					withStatus(v1alpha1.GroupObservation{FullPath: new("path/to/group")})),
			},
			want: want{
				cr: group(withExternalName("0"), withPermanentlyRemove(new(true)), withPath("group"),
					withStatus(v1alpha1.GroupObservation{FullPath: new("path/to/group")})),
				calls: []deleteGroupCalls{{Pid: "0", Opt: &gitlab.DeleteGroupOptions{}}},
				err:   nil,
			},
		},
		"PermanentlyRemoveAlreadyGone": {
			args: args{
				group: &fake.MockClient{
					MockDeleteGroup: func(pid interface{}, opt *gitlab.DeleteGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						recordedCalls = append(recordedCalls, deleteGroupCalls{Pid: pid, Opt: opt})
						if opt.PermanentlyRemove != nil {
							return &gitlab.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errBoom
						}
						return &gitlab.Response{}, nil
					},
					MockGetGroup: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						recordedGetCalls = append(recordedGetCalls, getGroupCalls{Pid: pid})
						return &gitlab.Group{ID: 0, FullPath: "group-deletion_scheduled-0", MarkedForDeletionOn: markedForDeletionOn}, &gitlab.Response{}, nil
					},
				},
				cr: group(withExternalName("0"), withPermanentlyRemove(new(true)), withPath("group"),
					withStatus(v1alpha1.GroupObservation{FullPath: new("path/to/group")})),
			},
			want: want{
				cr: group(withExternalName("0"), withPermanentlyRemove(new(true)), withPath("group"),
					withStatus(v1alpha1.GroupObservation{FullPath: new("path/to/group")})),
				calls: []deleteGroupCalls{
					{Pid: "0", Opt: &gitlab.DeleteGroupOptions{}},
					{Pid: "0", Opt: &gitlab.DeleteGroupOptions{PermanentlyRemove: new(true), FullPath: new("group-deletion_scheduled-0")}},
				},
				getCalls: []getGroupCalls{{Pid: "0"}},
				err:      nil,
			},
		},
		"PermanentlyRemoveWhenAlreadyMarked": {
			args: args{
				group: &fake.MockClient{
					MockDeleteGroup: func(pid interface{}, opt *gitlab.DeleteGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						recordedCalls = append(recordedCalls, deleteGroupCalls{Pid: pid, Opt: opt})
						if opt.PermanentlyRemove == nil {
							return nil, errGroupAlreadyMarked
						}
						return &gitlab.Response{}, nil
					},
					MockGetGroup: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						recordedGetCalls = append(recordedGetCalls, getGroupCalls{Pid: pid})
						return &gitlab.Group{ID: 0, FullPath: "group-deletion_scheduled-0", MarkedForDeletionOn: markedForDeletionOn}, &gitlab.Response{}, nil
					},
				},
				cr: group(
					withExternalName("0"),
					withPermanentlyRemove(new(true)),
					withPath("group"),
					withStatus(v1alpha1.GroupObservation{FullPath: new("path/to/group")})),
			},
			want: want{
				cr: group(
					withExternalName("0"),
					withPermanentlyRemove(new(true)),
					withPath("group"),
					withStatus(v1alpha1.GroupObservation{FullPath: new("path/to/group")})),
				calls: []deleteGroupCalls{
					{Pid: "0", Opt: &gitlab.DeleteGroupOptions{}},
					{Pid: "0", Opt: &gitlab.DeleteGroupOptions{PermanentlyRemove: new(true), FullPath: new("group-deletion_scheduled-0")}},
				},
				getCalls: []getGroupCalls{{Pid: "0"}},
				err:      nil,
			},
		},
		"PermanentlyRemoveWhenAncestorScheduled": {
			args: args{
				group: &fake.MockClient{
					MockDeleteGroup: func(pid interface{}, opt *gitlab.DeleteGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						recordedCalls = append(recordedCalls, deleteGroupCalls{Pid: pid, Opt: opt})
						return &gitlab.Response{}, nil
					},
					MockGetGroup: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
						recordedGetCalls = append(recordedGetCalls, getGroupCalls{Pid: pid})
						return &gitlab.Group{ID: 0, FullPath: "path/to/group"}, &gitlab.Response{}, nil
					},
				},
				cr: group(
					withExternalName("0"),
					withPermanentlyRemove(new(true)),
					withPath("group"),
					withStatus(v1alpha1.GroupObservation{FullPath: new("path/to/group")})),
			},
			want: want{
				cr: group(
					withExternalName("0"),
					withPermanentlyRemove(new(true)),
					withPath("group"),
					withStatus(v1alpha1.GroupObservation{FullPath: new("path/to/group")})),
				calls:    []deleteGroupCalls{{Pid: "0", Opt: &gitlab.DeleteGroupOptions{}}},
				getCalls: []getGroupCalls{{Pid: "0"}},
				err:      nil,
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
					withPermanentlyRemove(new(true)),
					withPath("top-level-group"),
					withStatus(v1alpha1.GroupObservation{FullPath: new("top-level-group")})),
			},
			want: want{
				cr: group(
					withExternalName("0"),
					withPermanentlyRemove(new(true)),
					withPath("top-level-group"),
					withStatus(v1alpha1.GroupObservation{FullPath: new("top-level-group")}),
				),
				calls: []deleteGroupCalls{
					{Pid: "0", Opt: &gitlab.DeleteGroupOptions{}},
				},
				getCalls: nil,
				err:      nil,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			recordedCalls = nil
			recordedGetCalls = nil
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
			if diff := cmp.Diff(tc.want.getCalls, recordedGetCalls, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

// archiveGitLab is a fake GitLab group that records update, archive, and
// unarchive calls in order. With AncestorArchived set, it rejects archive and
// unarchive calls the way GitLab does for a subgroup of an archived group.
// With RejectArchive set, it rejects archive calls.
type archiveGitLab struct {
	Archived         bool
	AncestorArchived bool
	RejectArchive    bool
	Description      string
	Calls            []string
}

var errAncestorArchived = errors.New("Cannot unarchive group since one of the ancestor groups is archived!")

func (g *archiveGitLab) client() *fake.MockClient {
	setArchived := func(call string, archived bool) (*gitlab.Response, error) {
		g.Calls = append(g.Calls, call)
		if g.AncestorArchived {
			return &gitlab.Response{}, errAncestorArchived
		}
		if archived && g.RejectArchive {
			return &gitlab.Response{}, errBoom
		}
		g.Archived = archived
		return &gitlab.Response{}, nil
	}
	return &fake.MockClient{
		MockGetGroup: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
			return &gitlab.Group{Path: path, Archived: g.Archived, Description: g.Description}, &gitlab.Response{}, nil
		},
		MockUpdateGroup: func(pid interface{}, opt *gitlab.UpdateGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
			g.Calls = append(g.Calls, "update")
			if opt.Description != nil {
				g.Description = *opt.Description
			}
			return &gitlab.Group{}, &gitlab.Response{}, nil
		},
		MockArchiveGroup: func(gid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
			return setArchived("archive", true)
		},
		MockUnarchiveGroup: func(gid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
			return setArchived("unarchive", false)
		},
	}
}

// TestArchiveConverges runs Observe and Update against a fake GitLab until the
// group is up to date or Update fails, and checks the archive state, the
// applied update, and the order of the GitLab calls.
func TestArchiveConverges(t *testing.T) {
	cases := map[string]struct {
		reason  string
		spec    v1alpha1.GroupParameters
		start   archiveGitLab
		want    archiveGitLab
		wantErr error
	}{
		"ArchiveAfterUpdate": {
			reason: "Archiving must run after the update.",
			spec:   v1alpha1.GroupParameters{Path: path, Archived: new(true), Description: new("new")},
			start:  archiveGitLab{Description: "old"},
			want:   archiveGitLab{Archived: true, Description: "new", Calls: []string{"update", "archive"}},
		},
		"ArchiveRejected": {
			reason:  "If GitLab rejects the archive, Update must fail after it updates.",
			spec:    v1alpha1.GroupParameters{Path: path, Archived: new(true), Description: new("new")},
			start:   archiveGitLab{RejectArchive: true, Description: "old"},
			want:    archiveGitLab{RejectArchive: true, Description: "new", Calls: []string{"update", "archive"}},
			wantErr: errors.Wrap(errBoom, errArchiveFailed),
		},
		"UnarchiveBeforeUpdate": {
			reason: "Unarchiving must run before the update.",
			spec:   v1alpha1.GroupParameters{Path: path, Archived: new(false), Description: new("new")},
			start:  archiveGitLab{Archived: true, Description: "old"},
			want:   archiveGitLab{Description: "new", Calls: []string{"unarchive", "update"}},
		},
		"UpdateStaysArchived": {
			reason: "A group that stays archived must not be unarchived to apply an update.",
			spec:   v1alpha1.GroupParameters{Path: path, Archived: new(true), Description: new("new")},
			start:  archiveGitLab{Archived: true, Description: "old"},
			want:   archiveGitLab{Archived: true, Description: "new", Calls: []string{"update"}},
		},
		"UnsetLeavesArchiveState": {
			reason: "Without archived in the spec, the archive state must not change.",
			spec:   v1alpha1.GroupParameters{Path: path, Description: new("new")},
			start:  archiveGitLab{Archived: true, Description: "old"},
			want:   archiveGitLab{Archived: true, Description: "new", Calls: []string{"update"}},
		},
		"AncestorArchived": {
			reason:  "If GitLab rejects the unarchive, Update must fail before it updates.",
			spec:    v1alpha1.GroupParameters{Path: path, Archived: new(false), Description: new("new")},
			start:   archiveGitLab{Archived: true, AncestorArchived: true, Description: "old"},
			want:    archiveGitLab{Archived: true, AncestorArchived: true, Description: "old", Calls: []string{"unarchive"}},
			wantErr: errors.Wrap(errAncestorArchived, errUnarchiveFailed),
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			gl := tc.start
			e := &external{client: gl.client()}
			cr := group(withExternalName(extName))
			cr.Spec.ForProvider = tc.spec

			var err error
			upToDate := false
			for range 4 {
				obs, oerr := e.Observe(context.Background(), cr)
				if oerr != nil {
					t.Fatalf("Observe(...): %v", oerr)
				}
				if upToDate = obs.ResourceUpToDate; upToDate {
					break
				}
				if _, err = e.Update(context.Background(), cr); err != nil {
					break
				}
			}

			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s\nUpdate(...): -want error, +got error:\n%s", tc.reason, diff)
			}
			if tc.wantErr == nil && !upToDate {
				t.Errorf("%s\ngroup never became up to date: %+v", tc.reason, gl)
			}
			if diff := cmp.Diff(tc.want, gl); diff != "" {
				t.Errorf("%s\nfinal GitLab state: -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}
