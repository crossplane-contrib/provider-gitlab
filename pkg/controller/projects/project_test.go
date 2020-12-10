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

package projects

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects/fake"
)

var (
	path = "some/path/to/repo"

	errBoom           = errors.New("boom")
	extName           = "example-project"
	extNameAnnotation = map[string]string{
		meta.AnnotationKeyExternalName: extName,
	}
)

type args struct {
	project projects.Client
	kube    client.Client
	cr      *v1alpha1.Project
}

type projectModifier func(*v1alpha1.Project)

func withConditions(c ...runtimev1alpha1.Condition) projectModifier {
	return func(r *v1alpha1.Project) { r.Status.ConditionedStatus.Conditions = c }
}

func withPath(p *string) projectModifier {
	return func(r *v1alpha1.Project) { r.Spec.ForProvider.Path = p }
}

func withExternalName(n string) projectModifier {
	return func(r *v1alpha1.Project) { meta.SetExternalName(r, n) }
}

func withStatus(s v1alpha1.ProjectObservation) projectModifier {
	return func(r *v1alpha1.Project) { r.Status.AtProvider = s }
}

func withDefaultValues() projectModifier {
	return func(p *v1alpha1.Project) {
		f := false
		i := 0
		p.Spec.ForProvider = v1alpha1.ProjectParameters{
			ResolveOutdatedDiffDiscussions:            &f,
			ContainerRegistryEnabled:                  &f,
			SharedRunnersEnabled:                      &f,
			PublicBuilds:                              &f,
			OnlyAllowMergeIfPipelineSucceeds:          &f,
			OnlyAllowMergeIfAllDiscussionsAreResolved: &f,
			RemoveSourceBranchAfterMerge:              &f,
			LFSEnabled:                                &f,
			RequestAccessEnabled:                      &f,
			CIDefaultGitDepth:                         &i,
			Mirror:                                    &f,
			MirrorUserID:                              &i,
			MirrorTriggerBuilds:                       &f,
			OnlyMirrorProtectedBranches:               &f,
			MirrorOverwritesDivergedBranches:          &f,
			PackagesEnabled:                           &f,
			ServiceDeskEnabled:                        &f,
			AutocloseReferencedIssues:                 &f,
		}
	}
}

func withAnnotations(a map[string]string) projectModifier {
	return func(p *v1alpha1.Project) { meta.AddAnnotations(p, a) }
}

func project(m ...projectModifier) *v1alpha1.Project {
	cr := &v1alpha1.Project{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.Project
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				project: &fake.MockClient{
					MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{Name: "example-project"}, &gitlab.Response{}, nil
					},
				},
				cr: project(
					withDefaultValues(),
					withExternalName(extName),
				),
			},
			want: want{
				cr: project(
					withDefaultValues(),
					withConditions(runtimev1alpha1.Available()),
					withAnnotations(extNameAnnotation),
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
				project: &fake.MockClient{
					MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{}, &gitlab.Response{}, errBoom
					},
				},
				cr: project(withExternalName(extName)),
			},
			want: want{
				cr:  project(withAnnotations(extNameAnnotation)),
				err: errors.Wrap(errBoom, errGetFailed),
			},
		},
		"LateInitFailedKubeUpdate": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(errBoom),
				},
				project: &fake.MockClient{
					MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{
							Path: path,
						}, &gitlab.Response{}, nil
					},
				},
				cr: project(
					withDefaultValues(),
					withExternalName(extName),
				),
			},
			want: want{
				cr: project(
					withDefaultValues(),
					withPath(&path),
					withAnnotations(extNameAnnotation),
				),
				err: errors.Wrap(errBoom, errKubeUpdateFailed),
			},
		},
		"LateInitSuccess": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				project: &fake.MockClient{
					MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{
							Path:         path,
							RunnersToken: "token",
						}, &gitlab.Response{}, nil
					},
				},
				cr: project(
					withDefaultValues(),
					withExternalName(extName),
				),
			},
			want: want{
				cr: project(
					withDefaultValues(),
					withConditions(runtimev1alpha1.Available()),
					withPath(&path),
					withAnnotations(extNameAnnotation),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
					ConnectionDetails:       managed.ConnectionDetails{"runnersToken": []byte("token")},
				},
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
		cr     *v1alpha1.Project
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
				project: &fake.MockClient{
					MockCreateProject: func(opt *gitlab.CreateProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{Name: extName, PathWithNamespace: extName}, &gitlab.Response{}, nil
					},
				},
				cr: project(),
			},
			want: want{
				cr: project(
					withConditions(runtimev1alpha1.Creating()),
					withExternalName(extName),
				),
				result: managed.ExternalCreation{},
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
				cr: project(
					withConditions(runtimev1alpha1.Creating()),
				),
				err: errors.Wrap(errBoom, errCreateFailed),
			},
		},
		"FailedKubeUpdate": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(errBoom),
				},
				project: &fake.MockClient{
					MockCreateProject: func(opt *gitlab.CreateProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{}, &gitlab.Response{}, nil
					},
				},
				cr: project(
					withExternalName(extName),
				),
			},
			want: want{
				cr: project(
					withConditions(runtimev1alpha1.Creating()),
					withExternalName(""),
				),
				err: errors.Wrap(errBoom, errKubeUpdateFailed),
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
	var (
		newExtName = "new/path/with/namespace"
	)
	type want struct {
		cr     *v1alpha1.Project
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulEditProject": {
			args: args{
				project: &fake.MockClient{
					MockEditProject: func(pid interface{}, opt *gitlab.EditProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{}, &gitlab.Response{}, nil
					},
				},
				cr: project(withStatus(v1alpha1.ProjectObservation{ID: 1234})),
			},
			want: want{
				cr: project(withStatus(v1alpha1.ProjectObservation{ID: 1234})),
			},
		},
		"FailedEdit": {
			args: args{
				project: &fake.MockClient{
					MockEditProject: func(pid interface{}, opt *gitlab.EditProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{}, &gitlab.Response{}, errBoom
					},
				},
				cr: project(withStatus(v1alpha1.ProjectObservation{ID: 1234})),
			},
			want: want{
				cr:  project(withStatus(v1alpha1.ProjectObservation{ID: 1234})),
				err: errors.Wrap(errBoom, errUpdateFailed),
			},
		},
		"SuccessfulUpdateExternalName": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				project: &fake.MockClient{
					MockEditProject: func(pid interface{}, opt *gitlab.EditProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{PathWithNamespace: newExtName}, &gitlab.Response{}, nil
					},
				},
				cr: project(
					withStatus(v1alpha1.ProjectObservation{ID: 1234}),
					withExternalName(extName),
				),
			},
			want: want{
				cr: project(
					withStatus(v1alpha1.ProjectObservation{ID: 1234}),
					withExternalName(newExtName),
				),
			},
		},
		"FailedUpdateExternalName": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(errBoom),
				},
				project: &fake.MockClient{
					MockEditProject: func(pid interface{}, opt *gitlab.EditProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{PathWithNamespace: newExtName}, &gitlab.Response{}, nil
					},
				},
				cr: project(
					withStatus(v1alpha1.ProjectObservation{ID: 1234}),
					withExternalName(extName),
				),
			},
			want: want{
				cr: project(
					withStatus(v1alpha1.ProjectObservation{ID: 1234}),
					withExternalName(newExtName),
				),
				err: errors.Wrap(errBoom, errKubeUpdateFailed),
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
