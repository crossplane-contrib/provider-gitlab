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

package projects

import (
	"context"
	"net/http"
	"reflect"
	"strconv"
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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
	timeNow           = time.Now()
)

type args struct {
	project projects.Client
	kube    client.Client
	cr      resource.Managed
}

type projectModifier func(*v1alpha1.Project)

func withConditions(c ...xpv1.Condition) projectModifier {
	return func(r *v1alpha1.Project) { r.Status.ConditionedStatus.Conditions = c }
}

func withPath(p *string) projectModifier {
	return func(r *v1alpha1.Project) { r.Spec.ForProvider.Path = p }
}

func withExternalName(projectID string) projectModifier {
	return func(r *v1alpha1.Project) { meta.SetExternalName(r, projectID) }
}

func withStatus(s v1alpha1.ProjectObservation) projectModifier {
	return func(r *v1alpha1.Project) { r.Status.AtProvider = s }
}

func withSpec(s v1alpha1.ProjectParameters) projectModifier {
	return func(r *v1alpha1.Project) { r.Spec.ForProvider = s }
}

func withProjectPushRules(pr *v1alpha1.PushRules) projectModifier {
	return func(r *v1alpha1.Project) { r.Spec.ForProvider.PushRules = pr }
}

func withClientDefaultValues() projectModifier {
	return func(p *v1alpha1.Project) {
		f := false
		i := 0
		p.Spec.ForProvider = v1alpha1.ProjectParameters{
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
	return func(p *v1alpha1.Project) { meta.AddAnnotations(p, a) }
}

func withMirrorUserIDNil() projectModifier {
	return func(p *v1alpha1.Project) { p.Spec.ForProvider.MirrorUserID = nil }
}

func project(m ...projectModifier) *v1alpha1.Project {
	cr := &v1alpha1.Project{}
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
				err: errors.New(errNotProject),
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
				err: errors.New(errNotProject),
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
				err: errors.New(errNotProject),
			},
		},
		"FailedGetRequest": {
			args: args{
				project: &fake.MockClient{
					MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 400}}, errBoom
					},
				},
				cr: project(withExternalName(extName)),
			},
			want: want{
				cr:     project(withExternalName(extName)),
				result: managed.ExternalObservation{ResourceExists: false},
				err:    errors.Wrap(errBoom, errGetFailed),
			},
		},
		"ErrGet404": {
			args: args{
				project: &fake.MockClient{
					MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errBoom
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
		"PushRulesNotAvailable404": {
			args: args{
				project: &fake.MockClient{
					MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{Name: "example-project"}, &gitlab.Response{}, nil
					},
					MockGetProjectPushRules: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
						// Simulate GitLab Community Edition where push rules feature is not available
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errBoom
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
					withStatus(v1alpha1.ProjectObservation{}),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: false,
					ConnectionDetails:       managed.ConnectionDetails{"runnersToken": []byte("")},
				},
			},
		},
		"PushRulesAvailableButNotConfigured": {
			args: args{
				project: &fake.MockClient{
					MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{Name: "example-project"}, &gitlab.Response{}, nil
					},
					MockGetProjectPushRules: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
						// Simulate GitLab Premium/Enterprise where push rules feature is available but not configured
						// This should return 200 with empty/default values, not 404
						return &gitlab.ProjectPushRules{
							// All default values (empty strings, false booleans, 0 integers)
							AuthorEmailRegex:           "",
							BranchNameRegex:            "",
							CommitCommitterCheck:       false,
							CommitCommitterNameCheck:   false,
							CommitMessageNegativeRegex: "",
							CommitMessageRegex:         "",
							DenyDeleteTag:              false,
							FileNameRegex:              "",
							MaxFileSize:                0,
							MemberCheck:                false,
							PreventSecrets:             false,
							RejectUnsignedCommits:      false,
							RejectNonDCOCommits:        false,
						}, &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
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
					// No push rules in spec - should NOT be late-initialized
					withStatus(v1alpha1.ProjectObservation{}),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,  // True because both are effectively empty
					ResourceLateInitialized: false, // No late initialization should happen
					ConnectionDetails:       managed.ConnectionDetails{"runnersToken": []byte("")},
				},
			},
		},
		"PushRulesConfiguredInGitLabButRemovedFromSpec": {
			args: args{
				project: &fake.MockClient{
					MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{Name: "example-project"}, &gitlab.Response{}, nil
					},
					MockGetProjectPushRules: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
						// Simulate GitLab where push rules are configured but removed from spec
						return &gitlab.ProjectPushRules{
							AuthorEmailRegex:           ".*@company.com",
							BranchNameRegex:            "^(feature|hotfix)/.*",
							CommitCommitterCheck:       true,
							CommitCommitterNameCheck:   true,
							CommitMessageNegativeRegex: "",
							CommitMessageRegex:         "^(feat|fix|docs):",
							DenyDeleteTag:              true,
							FileNameRegex:              "\\.(tmp|log)$",
							MaxFileSize:                100,
							MemberCheck:                true,
							PreventSecrets:             true,
							RejectUnsignedCommits:      false,
							RejectNonDCOCommits:        false,
						}, &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
					},
				},
				cr: project(
					withClientDefaultValues(),
					withExternalName(extName),
					// No push rules specified in spec - should clear the ones in GitLab
				),
			},
			want: want{
				cr: project(
					withClientDefaultValues(),
					withExternalName(extName),
					withConditions(xpv1.Available()),
					withStatus(v1alpha1.ProjectObservation{}),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false, // Should be false because GitLab has rules but spec doesn't
					ResourceLateInitialized: false,
					ConnectionDetails:       managed.ConnectionDetails{"runnersToken": []byte("")},
				},
			},
		},
		"SuccessfulAvailable": {
			args: args{
				project: &fake.MockClient{
					MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{Name: "example-project"}, &gitlab.Response{}, nil
					},
					MockGetProjectPushRules: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
						return &gitlab.ProjectPushRules{}, nil, nil
					},
				},
				cr: project(
					withClientDefaultValues(),
					withExternalName(extName),
					withProjectPushRules(&v1alpha1.PushRules{
						AuthorEmailRegex:           ptr.To(""),
						BranchNameRegex:            ptr.To(""),
						CommitCommitterCheck:       ptr.To(false),
						CommitCommitterNameCheck:   ptr.To(false),
						CommitMessageNegativeRegex: ptr.To(""),
						CommitMessageRegex:         ptr.To(""),
						DenyDeleteTag:              ptr.To(false),
						FileNameRegex:              ptr.To(""),
						MaxFileSize:                ptr.To(0),
						MemberCheck:                ptr.To(false),
						PreventSecrets:             ptr.To(false),
						RejectUnsignedCommits:      ptr.To(false),
						RejectNonDCOCommits:        ptr.To(false),
					}),
				),
			},
			want: want{
				cr: project(
					withClientDefaultValues(),
					withExternalName(extName),
					withProjectPushRules(&v1alpha1.PushRules{
						AuthorEmailRegex:           ptr.To(""),
						BranchNameRegex:            ptr.To(""),
						CommitCommitterCheck:       ptr.To(false),
						CommitCommitterNameCheck:   ptr.To(false),
						CommitMessageNegativeRegex: ptr.To(""),
						CommitMessageRegex:         ptr.To(""),
						DenyDeleteTag:              ptr.To(false),
						FileNameRegex:              ptr.To(""),
						MaxFileSize:                ptr.To(0),
						MemberCheck:                ptr.To(false),
						PreventSecrets:             ptr.To(false),
						RejectUnsignedCommits:      ptr.To(false),
						RejectNonDCOCommits:        ptr.To(false),
					}),
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
					MockGetProjectPushRules: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
						return &gitlab.ProjectPushRules{}, nil, nil
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
					// Push rules should NOT be late-initialized when not specified in spec
					withStatus(v1alpha1.ProjectObservation{}),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true, // Both spec and GitLab have no/empty rules
					ResourceLateInitialized: true, // Only the path should be late-initialized
					ConnectionDetails:       managed.ConnectionDetails{"runnersToken": []byte("token")},
				},
			},
		},
		"LateInitSuccessMirrorUserIdZero": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				project: &fake.MockClient{
					MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{MirrorUserID: 0}, &gitlab.Response{}, nil
					},
					MockGetProjectPushRules: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
						return &gitlab.ProjectPushRules{}, nil, nil
					},
				},
				cr: project(
					withClientDefaultValues(),
					withMirrorUserIDNil(),
					withProjectPushRules(&v1alpha1.PushRules{
						AuthorEmailRegex:           ptr.To(""),
						BranchNameRegex:            ptr.To(""),
						CommitCommitterCheck:       ptr.To(false),
						CommitCommitterNameCheck:   ptr.To(false),
						CommitMessageNegativeRegex: ptr.To(""),
						CommitMessageRegex:         ptr.To(""),
						DenyDeleteTag:              ptr.To(false),
						FileNameRegex:              ptr.To(""),
						MaxFileSize:                ptr.To(0),
						MemberCheck:                ptr.To(false),
						PreventSecrets:             ptr.To(false),
						RejectUnsignedCommits:      ptr.To(false),
						RejectNonDCOCommits:        ptr.To(false),
					},
					),
					withExternalName(extName),
				),
			},
			want: want{
				cr: project(
					withClientDefaultValues(),
					withMirrorUserIDNil(),
					withExternalName(extName),
					withProjectPushRules(&v1alpha1.PushRules{
						AuthorEmailRegex:           ptr.To(""),
						BranchNameRegex:            ptr.To(""),
						CommitCommitterCheck:       ptr.To(false),
						CommitCommitterNameCheck:   ptr.To(false),
						CommitMessageNegativeRegex: ptr.To(""),
						CommitMessageRegex:         ptr.To(""),
						DenyDeleteTag:              ptr.To(false),
						FileNameRegex:              ptr.To(""),
						MaxFileSize:                ptr.To(0),
						MemberCheck:                ptr.To(false),
						PreventSecrets:             ptr.To(false),
						RejectUnsignedCommits:      ptr.To(false),
						RejectNonDCOCommits:        ptr.To(false),
					}),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: false,
					ConnectionDetails:       managed.ConnectionDetails{"runnersToken": {}},
				},
			},
		},
		"DeleteOnPendingDeletion": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				project: &fake.MockClient{
					MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{
							MarkedForDeletionOn: &gitlab.ISOTime{},
						}, &gitlab.Response{}, nil
					},
					MockGetProjectPushRules: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
						return &gitlab.ProjectPushRules{}, nil, nil
					},
				},
				cr: project(
					withClientDefaultValues(),
					func(p *v1alpha1.Project) {
						p.DeletionTimestamp = &v1.Time{Time: timeNow}
						p.Spec.ForProvider.RemoveFinalizerOnPendingDeletion = ptr.To(true)
					},
					withExternalName(extName),
					withProjectPushRules(&v1alpha1.PushRules{
						AuthorEmailRegex:           ptr.To(""),
						BranchNameRegex:            ptr.To(""),
						CommitCommitterCheck:       ptr.To(false),
						CommitCommitterNameCheck:   ptr.To(false),
						CommitMessageNegativeRegex: ptr.To(""),
						CommitMessageRegex:         ptr.To(""),
						DenyDeleteTag:              ptr.To(false),
						FileNameRegex:              ptr.To(""),
						MaxFileSize:                ptr.To(0),
						MemberCheck:                ptr.To(false),
						PreventSecrets:             ptr.To(false),
						RejectUnsignedCommits:      ptr.To(false),
						RejectNonDCOCommits:        ptr.To(false),
					}),
				),
			},
			want: want{
				cr: project(
					withClientDefaultValues(),
					withExternalName(extName),
					withProjectPushRules(&v1alpha1.PushRules{
						AuthorEmailRegex:           ptr.To(""),
						BranchNameRegex:            ptr.To(""),
						CommitCommitterCheck:       ptr.To(false),
						CommitCommitterNameCheck:   ptr.To(false),
						CommitMessageNegativeRegex: ptr.To(""),
						CommitMessageRegex:         ptr.To(""),
						DenyDeleteTag:              ptr.To(false),
						FileNameRegex:              ptr.To(""),
						MaxFileSize:                ptr.To(0),
						MemberCheck:                ptr.To(false),
						PreventSecrets:             ptr.To(false),
						RejectUnsignedCommits:      ptr.To(false),
						RejectNonDCOCommits:        ptr.To(false),
					}),
					func(p *v1alpha1.Project) {
						p.DeletionTimestamp = &v1.Time{Time: timeNow}
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

	isProjectUpToDateCases := map[string]interface{}{
		"Name":                                      "name",
		"Path":                                      "path",
		"DefaultBranch":                             "Default branch",
		"Description":                               "description",
		"IssuesAccessLevel":                         gitlab.PublicAccessControl,
		"RepositoryAccessLevel":                     gitlab.PublicAccessControl,
		"MergeRequestsAccessLevel":                  gitlab.PublicAccessControl,
		"ForkingAccessLevel":                        gitlab.PublicAccessControl,
		"BuildsAccessLevel":                         gitlab.PublicAccessControl,
		"WikiAccessLevel":                           gitlab.PublicAccessControl,
		"SnippetsAccessLevel":                       gitlab.PublicAccessControl,
		"PagesAccessLevel":                          gitlab.PublicAccessControl,
		"ResolveOutdatedDiffDiscussions":            true,
		"ContainerRegistryEnabled":                  true,
		"ContainerRegistryAccessLevel":              gitlab.EnabledAccessControl,
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
		"Topics":                                    []string{"tag-1", "tag-2"},
		"CIConfigPath":                              "CI configPath",
		"CIDefaultGitDepth":                         1,
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
	topics := []string{"tag-1 new", "tag-2 new"}
	mergeMethod := v1alpha1.FastForwardMerge
	s := "default string"
	visibility := v1alpha1.PublicVisibility

	projectParameters := v1alpha1.ProjectParameters{
		Name:                             &s,
		Path:                             &s,
		DefaultBranch:                    &s,
		Description:                      &s,
		IssuesAccessLevel:                &al,
		RepositoryAccessLevel:            &al,
		MergeRequestsAccessLevel:         &al,
		ApprovalsBeforeMerge:             ptr.To(0),
		ForkingAccessLevel:               &al,
		BuildsAccessLevel:                &al,
		WikiAccessLevel:                  &al,
		SnippetsAccessLevel:              &al,
		PagesAccessLevel:                 &al,
		ResolveOutdatedDiffDiscussions:   &f,
		ContainerRegistryEnabled:         &f,
		ContainerRegistryAccessLevel:     &al,
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
		Topics:                           topics,
		CIConfigPath:                     &s,
		CIDefaultGitDepth:                &i,
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
		PushRules: &v1alpha1.PushRules{
			AuthorEmailRegex:           ptr.To(""),
			BranchNameRegex:            ptr.To(""),
			CommitCommitterCheck:       ptr.To(false),
			CommitCommitterNameCheck:   ptr.To(false),
			CommitMessageNegativeRegex: ptr.To(""),
			CommitMessageRegex:         ptr.To(""),
			DenyDeleteTag:              ptr.To(true),
			FileNameRegex:              ptr.To(""),
			MaxFileSize:                ptr.To(0),
			MemberCheck:                ptr.To(false),
			PreventSecrets:             ptr.To(false),
			RejectUnsignedCommits:      ptr.To(false),
			RejectNonDCOCommits:        ptr.To(false),
		},
	}

	for name, value := range isProjectUpToDateCases {
		argsProjectModifier := []projectModifier{
			withSpec(projectParameters),
			withExternalName("0"),
		}
		wantProjectModifier := []projectModifier{
			withSpec(projectParameters),
			withExternalName("0"),
			withProjectPushRules(&v1alpha1.PushRules{
				AuthorEmailRegex:           ptr.To(""),
				BranchNameRegex:            ptr.To(""),
				CommitCommitterCheck:       ptr.To(false),
				CommitCommitterNameCheck:   ptr.To(false),
				CommitMessageNegativeRegex: ptr.To(""),
				CommitMessageRegex:         ptr.To(""),
				DenyDeleteTag:              ptr.To(true),
				FileNameRegex:              ptr.To(""),
				MaxFileSize:                ptr.To(0),
				MemberCheck:                ptr.To(false),
				PreventSecrets:             ptr.To(false),
				RejectUnsignedCommits:      ptr.To(false),
				RejectNonDCOCommits:        ptr.To(false),
			}),
			withConditions(xpv1.Available()),
			withStatus(v1alpha1.ProjectObservation{
				IssuesAccessLevel:        al,
				BuildsAccessLevel:        al,
				MergeRequestsAccessLevel: al,
				SnippetsAccessLevel:      al,
				WikiAccessLevel:          al,
			}),
		}
		gitlabProject := &gitlab.Project{
			Name:                             s,
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
			Topics:                           topics,
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
		cases["IsProjectUpToDate"+name] = struct {
			args
			want
		}{
			args: args{
				project: &fake.MockClient{
					MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return gitlabProject, &gitlab.Response{}, nil
					},
					MockGetProjectPushRules: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
						return &gitlab.ProjectPushRules{
							DenyDeleteTag: true,
						}, nil, nil
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
				err: errors.New(errNotProject),
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
				err: errors.New(errNotProject),
			},
		},
		"SuccessfulEditProject": {
			args: args{
				project: &fake.MockClient{
					MockEditProject: func(pid interface{}, opt *gitlab.EditProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{}, &gitlab.Response{}, nil
					},
					MockEditProjectPushRule: func(pid interface{}, opt *gitlab.EditProjectPushRuleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
						return &gitlab.ProjectPushRules{}, &gitlab.Response{}, nil
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
					MockEditProjectPushRule: func(pid interface{}, opt *gitlab.EditProjectPushRuleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
						return &gitlab.ProjectPushRules{}, &gitlab.Response{}, nil
					},
				},
				cr: project(withStatus(v1alpha1.ProjectObservation{ID: 1234})),
			},
			want: want{
				cr:  project(withStatus(v1alpha1.ProjectObservation{ID: 1234})),
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

func withPermanentlyRemove(b *bool) projectModifier {
	return func(r *v1alpha1.Project) { r.Spec.ForProvider.PermanentlyRemove = b }
}

func TestDelete(t *testing.T) {
	type deleteProjectCalls struct {
		Pid interface{}
		Opt *gitlab.DeleteProjectOptions
	}
	var recordedCalls []deleteProjectCalls
	type want struct {
		cr    resource.Managed
		calls []deleteProjectCalls
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
				err: errors.New(errNotProject),
			},
		},
		"SuccessfulDeletion": {
			args: args{
				project: &fake.MockClient{
					MockDeleteProject: func(pid interface{}, opt *gitlab.DeleteProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
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
					MockDeleteProject: func(pid interface{}, opt *gitlab.DeleteProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return nil, errBoom
					},
				},
				cr: project(),
			},
			want: want{
				cr:  project(),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
		"SuccessfulPermanentlyDeletion": {
			args: args{
				project: &fake.MockClient{
					MockDeleteProject: func(pid interface{}, opt *gitlab.DeleteProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						recordedCalls = append(recordedCalls, deleteProjectCalls{Pid: pid, Opt: opt})
						return &gitlab.Response{}, nil
					},
				},
				cr: project(
					withExternalName("0"),
					withPermanentlyRemove(gitlab.Ptr(true)),
					withPath(gitlab.Ptr("project")),
					withStatus(v1alpha1.ProjectObservation{PathWithNamespace: "path/to/project"}),
				),
			},
			want: want{
				cr: project(
					withExternalName("0"),
					withPermanentlyRemove(gitlab.Ptr(true)),
					withPath(gitlab.Ptr("project")),
					withStatus(v1alpha1.ProjectObservation{PathWithNamespace: "path/to/project"}),
				),
				calls: []deleteProjectCalls{
					{Pid: "0", Opt: &gitlab.DeleteProjectOptions{}},
					{Pid: "0", Opt: &gitlab.DeleteProjectOptions{PermanentlyRemove: gitlab.Ptr(true), FullPath: gitlab.Ptr("path/to/project")}},
				},
				err: nil,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			recordedCalls = nil
			e := &external{kube: tc.kube, client: tc.project}
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
