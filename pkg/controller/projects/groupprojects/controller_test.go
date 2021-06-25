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

package groupprojects

import (
	"context"
	"reflect"
	"strconv"
	"testing"

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
	path              = "some/path/to/repo"
	unexpecedItem     resource.Managed
	errBoom           = errors.New("boom")
	projectID         = 1234
	extName           = strconv.Itoa(projectID)
	extNameAnnotation = map[string]string{meta.AnnotationKeyExternalName: extName}
)

type args struct {
	project projects.Client
	kube    client.Client
	cr      resource.Managed
}

type projectModifier func(*v1alpha1.GroupProject)

func withConditions(c ...xpv1.Condition) projectModifier {
	return func(r *v1alpha1.GroupProject) { r.Status.ConditionedStatus.Conditions = c }
}

func withPath(p *string) projectModifier {
	return func(r *v1alpha1.GroupProject) { r.Spec.ForProvider.Path = p }
}

func withExternalName(projectID string) projectModifier {
	return func(r *v1alpha1.GroupProject) { meta.SetExternalName(r, projectID) }
}

func withStatus(s v1alpha1.GroupProjectObservation) projectModifier {
	return func(r *v1alpha1.GroupProject) { r.Status.AtProvider = s }
}

func withSpec(s v1alpha1.GroupProjectParameters) projectModifier {
	return func(r *v1alpha1.GroupProject) { r.Spec.ForProvider = s }
}

func withClientDefaultValues() projectModifier {
	return func(p *v1alpha1.GroupProject) {
		f := false
		i := 0
		p.Spec.ForProvider = v1alpha1.GroupProjectParameters{
			AllowMergeOnSkippedPipeline:               &f,
			CIForwardDeploymentEnabled:                &f,
			NamespaceID:                               &i,
			EmailsDisabled:                            &f,
			ResolveOutdatedDiffDiscussions:            &f,
			ContainerRegistryEnabled:                  &f,
			SharedRunnersEnabled:                      &f,
			PublicBuilds:                              &f,
			OnlyAllowMergeIfPipelineSucceeds:          &f,
			OnlyAllowMergeIfAllDiscussionsAreResolved: &f,
			RemoveSourceBranchAfterMerge:              &f,
			LFSEnabled:                                &f,
			RequestAccessEnabled:                      &f,
			PrintingMergeRequestLinkEnabled:           &f,
			BuildTimeout:                              &i,
			CIDefaultGitDepth:                         &i,
			AutoDevopsEnabled:                         &f,
			ApprovalsBeforeMerge:                      &i,
			Mirror:                                    &f,
			MirrorUserID:                              &i,
			MirrorTriggerBuilds:                       &f,
			OnlyMirrorProtectedBranches:               &f,
			MirrorOverwritesDivergedBranches:          &f,
			InitializeWithReadme:                      &f,
			TemplateProjectID:                         &i,
			UseCustomTemplate:                         &f,
			GroupWithProjectTemplatesID:               &i,
			PackagesEnabled:                           &f,
			ServiceDeskEnabled:                        &f,
			AutocloseReferencedIssues:                 &f,
		}
	}
}

func withAnnotations(a map[string]string) projectModifier {
	return func(p *v1alpha1.GroupProject) { meta.AddAnnotations(p, a) }
}

func project(m ...projectModifier) *v1alpha1.GroupProject {
	cr := &v1alpha1.GroupProject{}
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
				err: errors.New(errNotGroupProject),
			},
		},
		"ProviderConfigRefNotGivenError": {
			args: args{
				cr:   project(),
				kube: &test.MockClient{MockGet: test.NewMockGetFn(nil)},
			},
			want: want{
				cr:  project(),
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
				err: errors.New(errNotGroupProject),
			},
		},
		"NoExternalName": {
			args: args{
				cr: project(),
			},
			want: want{
				cr: project(),
				result: managed.ExternalObservation{
					ResourceExists:          false,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
			},
		},
		"NotIDExternalName": {
			args: args{
				project: &fake.MockClient{
					MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{}, &gitlab.Response{}, nil
					},
				},
				cr: project(withExternalName("fr")),
			},
			want: want{
				cr:  project(withExternalName("fr")),
				err: errors.New(errNotGroupProject),
			},
		},
		"FailedGetRequest": {
			args: args{
				project: &fake.MockClient{
					MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return nil, &gitlab.Response{}, errBoom
					},
				},
				cr: project(withExternalName(extName)),
			},
			want: want{
				cr:     project(withExternalName(extName)),
				result: managed.ExternalObservation{ResourceExists: false},
				err:    nil,
			},
		},
		"SuccessfulAvailable": {
			args: args{
				project: &fake.MockClient{
					MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{Name: "example-project"}, &gitlab.Response{}, nil
					},
				},
				cr: project(
					withClientDefaultValues(),
					withExternalName(extName),
				),
			},
			want: want{
				cr: project(
					withClientDefaultValues(),
					withExternalName(extName),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: false,
					ConnectionDetails:       managed.ConnectionDetails{"runnersToken": []byte("")},
				},
			},
		},
		"LateInitSuccess": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				project: &fake.MockClient{
					MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{Path: path, RunnersToken: "token"}, &gitlab.Response{}, nil
					},
				},
				cr: project(
					withClientDefaultValues(),
					withExternalName(extName),
				),
			},
			want: want{
				cr: project(
					withClientDefaultValues(),
					withConditions(xpv1.Available()),
					withPath(&path),
					withExternalName(extName),
					withStatus(v1alpha1.GroupProjectObservation{RunnersToken: "token"}),
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

	isProjectUpToDateCases := map[string]interface{}{
		"Path":                                      "path",
		"DefaultBranch":                             "Default branch",
		"Description":                               "description",
		"IssuesAccessLevel":                         gitlab.PrivateAccessControl,
		"RepositoryAccessLevel":                     gitlab.PrivateAccessControl,
		"MergeRequestsAccessLevel":                  gitlab.PrivateAccessControl,
		"ForkingAccessLevel":                        gitlab.PrivateAccessControl,
		"BuildsAccessLevel":                         gitlab.PrivateAccessControl,
		"WikiAccessLevel":                           gitlab.PrivateAccessControl,
		"SnippetsAccessLevel":                       gitlab.PrivateAccessControl,
		"PagesAccessLevel":                          gitlab.PrivateAccessControl,
		"ResolveOutdatedDiffDiscussions":            true,
		"ContainerRegistryEnabled":                  true,
		"SharedRunnersEnabled":                      true,
		"Visibility":                                gitlab.PrivateVisibility,
		"PublicBuilds":                              true,
		"OnlyAllowMergeIfPipelineSucceeds":          true,
		"OnlyAllowMergeIfAllDiscussionsAreResolved": true,
		"MergeMethod":                               gitlab.RebaseMerge,
		"RemoveSourceBranchAfterMerge":              true,
		"LFSEnabled":                                true,
		"RequestAccessEnabled":                      true,
		"TagList":                                   []string{"tag-1", "tag-2"},
		"CIConfigPath":                              "CI configPath",
		"CIDefaultGitDepth":                         1,
		"ApprovalsBeforeMerge":                      1,
		"Mirror":                                    true,
		"MirrorUserID":                              1,
		"MirrorTriggerBuilds":                       true,
		"OnlyMirrorProtectedBranches":               true,
		"MirrorOverwritesDivergedBranches":          true,
		"PackagesEnabled":                           true,
		"ServiceDeskEnabled":                        true,
		"AutocloseReferencedIssues":                 true,
		"AllowMergeOnSkippedPipeline":               true,
		"CIForwardDeploymentEnabled":                true,
	}

	f := false
	i := 0
	al := v1alpha1.PublicAccessControl
	tags := []string{"tag-1 new", "tag-2 new"}
	mergeMethod := v1alpha1.FastForwardMerge
	s := "default string"
	visibility := v1alpha1.PublicVisibility

	projectParameters := v1alpha1.GroupProjectParameters{
		Path:                             &s,
		DefaultBranch:                    &s,
		Description:                      &s,
		IssuesAccessLevel:                &al,
		RepositoryAccessLevel:            &al,
		MergeRequestsAccessLevel:         &al,
		ForkingAccessLevel:               &al,
		BuildsAccessLevel:                &al,
		WikiAccessLevel:                  &al,
		SnippetsAccessLevel:              &al,
		PagesAccessLevel:                 &al,
		ResolveOutdatedDiffDiscussions:   &f,
		ContainerRegistryEnabled:         &f,
		SharedRunnersEnabled:             &f,
		Visibility:                       &visibility,
		PublicBuilds:                     &f,
		OnlyAllowMergeIfPipelineSucceeds: &f,
		OnlyAllowMergeIfAllDiscussionsAreResolved: &f,
		MergeMethod:                      &mergeMethod,
		RemoveSourceBranchAfterMerge:     &f,
		LFSEnabled:                       &f,
		RequestAccessEnabled:             &f,
		TagList:                          tags,
		CIConfigPath:                     &s,
		CIDefaultGitDepth:                &i,
		ApprovalsBeforeMerge:             &i,
		Mirror:                           &f,
		MirrorUserID:                     &i,
		MirrorTriggerBuilds:              &f,
		OnlyMirrorProtectedBranches:      &f,
		MirrorOverwritesDivergedBranches: &f,
		PackagesEnabled:                  &f,
		ServiceDeskEnabled:               &f,
		AutocloseReferencedIssues:        &f,
		AllowMergeOnSkippedPipeline:      &f,
		CIForwardDeploymentEnabled:       &f,
	}

	for name, value := range isProjectUpToDateCases {
		argsProjectModifier := []projectModifier{
			withSpec(projectParameters),
			withExternalName("0"),
		}
		wantProjectModifier := []projectModifier{
			withSpec(projectParameters),
			withExternalName("0"),
			withConditions(xpv1.Available()),
		}
		gitlabProject := &gitlab.Project{
			Path:                             s,
			DefaultBranch:                    s,
			Description:                      s,
			IssuesAccessLevel:                gitlab.PublicAccessControl,
			RepositoryAccessLevel:            gitlab.PublicAccessControl,
			MergeRequestsAccessLevel:         gitlab.PublicAccessControl,
			ForkingAccessLevel:               gitlab.PublicAccessControl,
			BuildsAccessLevel:                gitlab.PublicAccessControl,
			WikiAccessLevel:                  gitlab.PublicAccessControl,
			SnippetsAccessLevel:              gitlab.PublicAccessControl,
			PagesAccessLevel:                 gitlab.PublicAccessControl,
			ResolveOutdatedDiffDiscussions:   f,
			ContainerRegistryEnabled:         f,
			SharedRunnersEnabled:             f,
			Visibility:                       gitlab.PublicVisibility,
			PublicBuilds:                     f,
			OnlyAllowMergeIfPipelineSucceeds: f,
			OnlyAllowMergeIfAllDiscussionsAreResolved: f,
			MergeMethod:                      gitlab.FastForwardMerge,
			RemoveSourceBranchAfterMerge:     f,
			LFSEnabled:                       f,
			RequestAccessEnabled:             f,
			TagList:                          tags,
			CIConfigPath:                     s,
			CIDefaultGitDepth:                i,
			ApprovalsBeforeMerge:             i,
			Mirror:                           f,
			MirrorUserID:                     i,
			MirrorTriggerBuilds:              f,
			OnlyMirrorProtectedBranches:      f,
			MirrorOverwritesDivergedBranches: f,
			PackagesEnabled:                  f,
			ServiceDeskEnabled:               f,
			AutocloseReferencedIssues:        f,
			AllowMergeOnSkippedPipeline:      f,
			CIForwardDeploymentEnabled:       f,
		}
		gitlabProject.Name = name
		structValue := reflect.ValueOf(gitlabProject).Elem()
		structFieldValue := structValue.FieldByName(name)
		val := reflect.ValueOf(value)

		structFieldValue.Set(val)
		cases["IsGroupUpToDate"+name] = struct {
			args
			want
		}{
			args: args{
				project: &fake.MockClient{
					MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return gitlabProject, &gitlab.Response{}, nil
					},
				},
				cr: project(argsProjectModifier...),
			},
			want: want{
				cr: project(wantProjectModifier...),
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
			e := &external{kube: tc.kube, client: tc.project}
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
				err: errors.New(errNotGroupProject),
			},
		},
		"SuccessfulCreation": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				project: &fake.MockClient{
					MockCreateProject: func(opt *gitlab.CreateProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{Name: extName, ID: 0}, &gitlab.Response{}, nil
					},
				},
				cr: project(withAnnotations(extNameAnnotation)),
			},
			want: want{
				cr:     project(withExternalName("0")),
				result: managed.ExternalCreation{ExternalNameAssigned: true},
			},
		},
		"FailedCreation": {
			args: args{
				project: &fake.MockClient{
					MockCreateProject: func(opt *gitlab.CreateProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{}, &gitlab.Response{}, errBoom
					},
				},
				cr: project(),
			},
			want: want{
				cr:  project(),
				err: errors.Wrap(errBoom, errCreateFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.project}
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
				err: errors.New(errNotGroupProject),
			},
		},
		"SuccessfulEditProject": {
			args: args{
				project: &fake.MockClient{
					MockEditProject: func(pid interface{}, opt *gitlab.EditProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{}, &gitlab.Response{}, nil
					},
				},
				cr: project(withStatus(v1alpha1.GroupProjectObservation{ID: 1234})),
			},
			want: want{
				cr: project(withStatus(v1alpha1.GroupProjectObservation{ID: 1234})),
			},
		},
		"FailedEdit": {
			args: args{
				project: &fake.MockClient{
					MockEditProject: func(pid interface{}, opt *gitlab.EditProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{}, &gitlab.Response{}, errBoom
					},
				},
				cr: project(withStatus(v1alpha1.GroupProjectObservation{ID: 1234})),
			},
			want: want{
				cr:  project(withStatus(v1alpha1.GroupProjectObservation{ID: 1234})),
				err: errors.Wrap(errBoom, errUpdateFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.project}
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
				err: errors.New(errNotGroupProject),
			},
		},
		"SuccessfulDeletion": {
			args: args{
				project: &fake.MockClient{
					MockDeleteProject: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: project(withExternalName("0")),
			},
			want: want{
				cr:  project(withExternalName("0")),
				err: nil,
			},
		},
		"FailedDeletion": {
			args: args{
				project: &fake.MockClient{
					MockDeleteProject: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: project(),
			},
			want: want{
				cr:  project(),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.project}
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
