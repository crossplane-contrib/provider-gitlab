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

	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects/fake"
)

var (
	path              = "some/path/to/repo"
	unexpecedItem     resource.Managed
	errBoom           = errors.New("boom")
	projectID         = int64(1234)
	extName           = strconv.FormatInt(projectID, 10)
	extNameAnnotation = map[string]string{meta.AnnotationKeyExternalName: extName}
	timeNow           = time.Now()
)

type args struct {
	project projects.Client
	kube    client.Client
	cr      resource.Managed
}

type projectModifier func(*v1alpha1.Project)

func withConditions(c ...v2.Condition) projectModifier {
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

func withMergeTrainSettings(enabled, skipAllowed, pipelines *bool) projectModifier {
	return func(r *v1alpha1.Project) {
		r.Spec.ForProvider.MergeTrainsEnabled = enabled
		r.Spec.ForProvider.MergeTrainsSkipTrainAllowed = skipAllowed
		r.Spec.ForProvider.MergePipelinesEnabled = pipelines
	}
}

func withClientDefaultValues() projectModifier {
	return func(p *v1alpha1.Project) {
		f := false
		i64 := int64(0)
		p.Spec.ForProvider = v1alpha1.ProjectParameters{
			AllowMergeOnSkippedPipeline:               &f,
			CIForwardDeploymentEnabled:                &f,
			NamespaceID:                               &i64,
			EmailsDisabled:                            &f,
			ResolveOutdatedDiffDiscussions:            &f,
			ContainerRegistryEnabled:                  &f, //nolint:staticcheck
			SharedRunnersEnabled:                      &f,
			PublicBuilds:                              &f, //nolint:staticcheck
			OnlyAllowMergeIfPipelineSucceeds:          &f,
			OnlyAllowMergeIfAllDiscussionsAreResolved: &f,
			RemoveSourceBranchAfterMerge:              &f,
			LFSEnabled:                                &f,
			RequestAccessEnabled:                      &f,
			PrintingMergeRequestLinkEnabled:           &f,
			BuildTimeout:                              &i64,
			CIDefaultGitDepth:                         &i64,
			AutoDevopsEnabled:                         &f,
			ApprovalsBeforeMerge:                      &i64, //nolint:staticcheck
			MergeTrainsEnabled:                        &f,
			MergeTrainsSkipTrainAllowed:               &f,
			MergePipelinesEnabled:                     &f,
			Mirror:                                    &f,
			MirrorUserID:                              &i64,
			MirrorTriggerBuilds:                       &f,
			OnlyMirrorProtectedBranches:               &f,
			MirrorOverwritesDivergedBranches:          &f,
			InitializeWithReadme:                      &f,
			TemplateProjectID:                         &i64,
			UseCustomTemplate:                         &f,
			GroupWithProjectTemplatesID:               &i64,
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
					withConditions(v2.Available()),
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
					withConditions(v2.Available()),
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
							ID:                         1,
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
					withConditions(v2.Available()),
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
						return &gitlab.ProjectPushRules{ID: 1}, nil, nil
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
						MaxFileSize:                ptr.To(int64(0)),
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
						MaxFileSize:                ptr.To(int64(0)),
						MemberCheck:                ptr.To(false),
						PreventSecrets:             ptr.To(false),
						RejectUnsignedCommits:      ptr.To(false),
						RejectNonDCOCommits:        ptr.To(false),
					}),
					withConditions(v2.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: false,
					ConnectionDetails:       managed.ConnectionDetails{"runnersToken": []byte("")},
				},
			},
		},
		// Regression test: ImportURLSecretRef with credentials embedded in the URL.
		// GitLab ALWAYS strips userinfo from ImportURL in API responses.
		// "https://TOKEN:@github.com/org/repo.git" becomes "https://github.com/org/repo.git".
		// The controller must sanitize the spec URL before comparing so that the
		// credential-bearing URL does not look different from GitLab's response
		"ImportURLSecretRefUpToDate": {
			args: args{
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						*obj.(*corev1.Secret) = corev1.Secret{
							Data: map[string][]byte{
								// Credential-bearing URL as stored in the Secret
								"url": []byte("https://GITHUB_TOKEN:@github.com/example/repo.git"),
							},
						}
						return nil
					}),
				},
				project: &fake.MockClient{
					MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						// GitLab API always returns the sanitized, credential-free URL
						return &gitlab.Project{
							ImportURL: "https://github.com/example/repo.git",
						}, &gitlab.Response{}, nil
					},
					MockGetProjectPushRules: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errBoom
					},
				},
				cr: project(
					withClientDefaultValues(),
					withExternalName(extName),
					func(p *v1alpha1.Project) {
						p.Spec.ForProvider.ImportURLSecretRef = &v2.LocalSecretKeySelector{
							Key:                  "url",
							LocalSecretReference: v2.LocalSecretReference{Name: "import-url-secret"},
						}
					},
				),
			},
			want: want{
				cr: project(
					withClientDefaultValues(),
					withExternalName(extName),
					withConditions(v2.Available()),
					withStatus(v1alpha1.ProjectObservation{}),
					func(p *v1alpha1.Project) {
						p.Spec.ForProvider.ImportURLSecretRef = &v2.LocalSecretKeySelector{
							Key:                  "url",
							LocalSecretReference: v2.LocalSecretReference{Name: "import-url-secret"},
						}
					},
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,  // sanitized URL matches GitLab's response
					ResourceLateInitialized: false, // no spurious late-init from secret resolution
					ConnectionDetails:       managed.ConnectionDetails{"runnersToken": []byte("")},
				},
			},
		},
		// Regression test: credential rotation only (same host/path, different token).
		// Since GitLab never returns credentials we cannot detect a pure token change
		// through the API. The sanitized URLs will match, so the resource is considered
		// up-to-date. This is expected behaviour; the alternative would be an infinite
		// update loop on every reconcile.
		"ImportURLSecretRefCredentialRotationIsUpToDate": {
			args: args{
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						*obj.(*corev1.Secret) = corev1.Secret{
							Data: map[string][]byte{
								// Token was rotated; base URL is the same
								"url": []byte("https://NEW_TOKEN:@github.com/example/repo.git"),
							},
						}
						return nil
					}),
				},
				project: &fake.MockClient{
					MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						// GitLab returns the same sanitized URL regardless of the token used
						return &gitlab.Project{
							ImportURL: "https://github.com/example/repo.git",
						}, &gitlab.Response{}, nil
					},
					MockGetProjectPushRules: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errBoom
					},
				},
				cr: project(
					withClientDefaultValues(),
					withExternalName(extName),
					func(p *v1alpha1.Project) {
						p.Spec.ForProvider.ImportURLSecretRef = &v2.LocalSecretKeySelector{
							Key:                  "url",
							LocalSecretReference: v2.LocalSecretReference{Name: "import-url-secret"},
						}
					},
				),
			},
			want: want{
				cr: project(
					withClientDefaultValues(),
					withExternalName(extName),
					withConditions(v2.Available()),
					withStatus(v1alpha1.ProjectObservation{}),
					func(p *v1alpha1.Project) {
						p.Spec.ForProvider.ImportURLSecretRef = &v2.LocalSecretKeySelector{
							Key:                  "url",
							LocalSecretReference: v2.LocalSecretReference{Name: "import-url-secret"},
						}
					},
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true, // sanitized URLs are identical; credential difference is undetectable via API
					ResourceLateInitialized: false,
					ConnectionDetails:       managed.ConnectionDetails{"runnersToken": []byte("")},
				},
			},
		},
		// Regression test: when the underlying repository URL changes (different host
		// or path, not merely credentials), the resource must be marked not-up-to-date
		// so exactly one corrective update is issued. ResourceLateInitialized must still
		// not be set spuriously.
		"ImportURLSecretRefNotUpToDate": {
			args: args{
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						*obj.(*corev1.Secret) = corev1.Secret{
							Data: map[string][]byte{
								// Secret now points to a completely different repository
								"url": []byte("https://TOKEN:@github.com/example/new-repo.git"),
							},
						}
						return nil
					}),
				},
				project: &fake.MockClient{
					MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						// GitLab still mirrors the old repository
						return &gitlab.Project{
							ImportURL: "https://github.com/example/old-repo.git",
						}, &gitlab.Response{}, nil
					},
					MockGetProjectPushRules: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errBoom
					},
				},
				cr: project(
					withClientDefaultValues(),
					withExternalName(extName),
					func(p *v1alpha1.Project) {
						p.Spec.ForProvider.ImportURLSecretRef = &v2.LocalSecretKeySelector{
							Key:                  "url",
							LocalSecretReference: v2.LocalSecretReference{Name: "import-url-secret"},
						}
					},
				),
			},
			want: want{
				cr: project(
					withClientDefaultValues(),
					withExternalName(extName),
					withConditions(v2.Available()),
					withStatus(v1alpha1.ProjectObservation{}),
					func(p *v1alpha1.Project) {
						p.Spec.ForProvider.ImportURLSecretRef = &v2.LocalSecretKeySelector{
							Key:                  "url",
							LocalSecretReference: v2.LocalSecretReference{Name: "import-url-secret"},
						}
					},
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false, // repository URL changed → one corrective update required
					ResourceLateInitialized: false, // no spurious late-init
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
						return &gitlab.Project{
							Path:         path,
							RunnersToken: "token",
						}, &gitlab.Response{}, nil
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
					withConditions(v2.Available()),
					withPath(&path),
					withExternalName(extName),
					// Push rules should NOT be late-initialized when not specified in spec
					withStatus(v1alpha1.ProjectObservation{}),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true, // Both spec and GitLab have no/empty rules
					ResourceLateInitialized: true, // Only path should be late-initialized
					ConnectionDetails:       managed.ConnectionDetails{"runnersToken": []byte("token")},
				},
			},
		},
		"LateInitMergeTrainSettings": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				project: &fake.MockClient{
					MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{
							MergeTrainsEnabled:          true,
							MergeTrainsSkipTrainAllowed: true,
							MergePipelinesEnabled:       true,
							RunnersToken:                "token",
						}, &gitlab.Response{}, nil
					},
					MockGetProjectPushRules: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
						return &gitlab.ProjectPushRules{}, nil, nil
					},
				},
				cr: project(
					withClientDefaultValues(),
					withMergeTrainSettings(nil, nil, nil),
					withExternalName(extName),
				),
			},
			want: want{
				cr: project(
					withClientDefaultValues(),
					withMergeTrainSettings(ptr.To(true), ptr.To(true), ptr.To(true)),
					withConditions(v2.Available()),
					withExternalName(extName),
					withStatus(v1alpha1.ProjectObservation{
						MergeTrainsEnabled:          true,
						MergeTrainsSkipTrainAllowed: true,
						MergePipelinesEnabled:       true,
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
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
						return &gitlab.ProjectPushRules{ID: 1}, nil, nil
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
						MaxFileSize:                ptr.To(int64(0)),
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
						MaxFileSize:                ptr.To(int64(0)),
						MemberCheck:                ptr.To(false),
						PreventSecrets:             ptr.To(false),
						RejectUnsignedCommits:      ptr.To(false),
						RejectNonDCOCommits:        ptr.To(false),
					}),
					withConditions(v2.Available()),
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
						MaxFileSize:                ptr.To(int64(0)),
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
						MaxFileSize:                ptr.To(int64(0)),
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
		"IssuesAccessLevel":                         gitlab.EnabledAccessControl,
		"RepositoryAccessLevel":                     gitlab.EnabledAccessControl,
		"MergeRequestsAccessLevel":                  gitlab.EnabledAccessControl,
		"MergeTrainsEnabled":                        true,
		"MergeTrainsSkipTrainAllowed":               true,
		"MergePipelinesEnabled":                     true,
		"ForkingAccessLevel":                        gitlab.EnabledAccessControl,
		"BuildsAccessLevel":                         gitlab.EnabledAccessControl,
		"WikiAccessLevel":                           gitlab.EnabledAccessControl,
		"SnippetsAccessLevel":                       gitlab.EnabledAccessControl,
		"PagesAccessLevel":                          gitlab.EnabledAccessControl,
		"AnalyticsAccessLevel":                      gitlab.DisabledAccessControl,
		"ReleasesAccessLevel":                       gitlab.DisabledAccessControl,
		"EnvironmentsAccessLevel":                   gitlab.DisabledAccessControl,
		"FeatureFlagsAccessLevel":                   gitlab.DisabledAccessControl,
		"InfrastructureAccessLevel":                 gitlab.DisabledAccessControl,
		"MonitorAccessLevel":                        gitlab.DisabledAccessControl,
		"PackageRegistryAccessLevel":                gitlab.DisabledAccessControl,
		"SecurityAndComplianceAccessLevel":          gitlab.DisabledAccessControl,
		"ModelExperimentsAccessLevel":               gitlab.DisabledAccessControl,
		"ModelRegistryAccessLevel":                  gitlab.DisabledAccessControl,
		"ResolveOutdatedDiffDiscussions":            true,
		"ContainerRegistryEnabled":                  true,
		"ContainerRegistryAccessLevel":              gitlab.EnabledAccessControl,
		"SharedRunnersEnabled":                      true,
		"Visibility":                                gitlab.PrivateVisibility,
		"PublicJobs":                                true,
		"OnlyAllowMergeIfPipelineSucceeds":          true,
		"OnlyAllowMergeIfAllDiscussionsAreResolved": true,
		"MergeMethod":                               gitlab.RebaseMerge,
		"RemoveSourceBranchAfterMerge":              true,
		"LFSEnabled":                                true,
		"RequestAccessEnabled":                      true,
		"TagList":                                   []string{"tag-1", "tag-2"},
		"Topics":                                    []string{"tag-1", "tag-2"},
		"CIConfigPath":                              "CI configPath",
		"CIDefaultGitDepth":                         int64(1),
		"Mirror":                                    true,
		"MirrorUserID":                              int64(1),
		"MirrorTriggerBuilds":                       true,
		"OnlyMirrorProtectedBranches":               true,
		"MirrorOverwritesDivergedBranches":          true,
		"PackagesEnabled":                           false,
		"ServiceDeskEnabled":                        true,
		"AutocloseReferencedIssues":                 true,
		"AllowMergeOnSkippedPipeline":               true,
		"CIForwardDeploymentEnabled":                true,
		"Archived":                                  true,
	}

	f := false
	i64 := int64(0)
	al := v1alpha1.PublicAccessControl
	tags := []string{"tag-1 new", "tag-2 new"}
	topics := []string{"tag-1 new", "tag-2 new"}
	mergeMethod := v1alpha1.FastForwardMerge
	s := "default string"
	visibility := v1alpha1.PublicVisibility

	projectParameters := v1alpha1.ProjectParameters{
		Name:                                      &s,
		Path:                                      &s,
		DefaultBranch:                             &s,
		Description:                               &s,
		IssuesAccessLevel:                         &al,
		RepositoryAccessLevel:                     &al,
		MergeRequestsAccessLevel:                  &al,
		MergeTrainsEnabled:                        &f,
		MergeTrainsSkipTrainAllowed:               &f,
		MergePipelinesEnabled:                     &f,
		ApprovalsBeforeMerge:                      ptr.To(int64(0)), //nolint:staticcheck
		ForkingAccessLevel:                        &al,
		BuildsAccessLevel:                         &al,
		WikiAccessLevel:                           &al,
		SnippetsAccessLevel:                       &al,
		PagesAccessLevel:                          &al,
		AnalyticsAccessLevel:                      &al,
		ReleasesAccessLevel:                       &al,
		EnvironmentsAccessLevel:                   &al,
		FeatureFlagsAccessLevel:                   &al,
		InfrastructureAccessLevel:                 &al,
		MonitorAccessLevel:                        &al,
		PackageRegistryAccessLevel:                &al,
		SecurityAndComplianceAccessLevel:          &al,
		ModelExperimentsAccessLevel:               &al,
		ModelRegistryAccessLevel:                  &al,
		ResolveOutdatedDiffDiscussions:            &f,
		ContainerRegistryEnabled:                  &f, //nolint:staticcheck
		ContainerRegistryAccessLevel:              &al,
		SharedRunnersEnabled:                      &f,
		Visibility:                                &visibility,
		PublicBuilds:                              &f, //nolint:staticcheck
		OnlyAllowMergeIfPipelineSucceeds:          &f,
		OnlyAllowMergeIfAllDiscussionsAreResolved: &f,
		MergeMethod:                               &mergeMethod,
		RemoveSourceBranchAfterMerge:              &f,
		LFSEnabled:                                &f,
		RequestAccessEnabled:                      &f,
		TagList:                                   tags, //nolint:staticcheck
		Topics:                                    topics,
		CIConfigPath:                              &s,
		CIDefaultGitDepth:                         &i64,
		Mirror:                                    &f,
		MirrorUserID:                              &i64,
		MirrorTriggerBuilds:                       &f,
		OnlyMirrorProtectedBranches:               &f,
		MirrorOverwritesDivergedBranches:          &f,
		PackagesEnabled:                           &f,
		ServiceDeskEnabled:                        &f,
		AutocloseReferencedIssues:                 &f,
		AllowMergeOnSkippedPipeline:               &f,
		CIForwardDeploymentEnabled:                &f,
		Archived:                                  &f,
		PushRules: &v1alpha1.PushRules{
			AuthorEmailRegex:           ptr.To(""),
			BranchNameRegex:            ptr.To(""),
			CommitCommitterCheck:       ptr.To(false),
			CommitCommitterNameCheck:   ptr.To(false),
			CommitMessageNegativeRegex: ptr.To(""),
			CommitMessageRegex:         ptr.To(""),
			DenyDeleteTag:              ptr.To(true),
			FileNameRegex:              ptr.To(""),
			MaxFileSize:                ptr.To(int64(0)),
			MemberCheck:                ptr.To(false),
			PreventSecrets:             ptr.To(false),
			RejectUnsignedCommits:      ptr.To(false),
			RejectNonDCOCommits:        ptr.To(false),
		},
	}

	// The unchanged baseline must be up to date. Otherwise every case below
	// passes regardless of whether the comparison it targets exists.
	const baseline = "Baseline"
	isProjectUpToDateCases[baseline] = nil

	for name, value := range isProjectUpToDateCases {
		argsProjectModifier := []projectModifier{
			withSpec(projectParameters),
			withExternalName("0"),
		}
		gitlabProject := &gitlab.Project{
			Name:                                      s,
			Path:                                      s,
			DefaultBranch:                             s,
			Description:                               s,
			IssuesAccessLevel:                         gitlab.PublicAccessControl,
			RepositoryAccessLevel:                     gitlab.PublicAccessControl,
			MergeRequestsAccessLevel:                  gitlab.PublicAccessControl,
			MergeTrainsEnabled:                        f,
			MergeTrainsSkipTrainAllowed:               f,
			MergePipelinesEnabled:                     f,
			ForkingAccessLevel:                        gitlab.PublicAccessControl,
			BuildsAccessLevel:                         gitlab.PublicAccessControl,
			WikiAccessLevel:                           gitlab.PublicAccessControl,
			SnippetsAccessLevel:                       gitlab.PublicAccessControl,
			PagesAccessLevel:                          gitlab.PublicAccessControl,
			AnalyticsAccessLevel:                      gitlab.PublicAccessControl,
			ReleasesAccessLevel:                       gitlab.PublicAccessControl,
			EnvironmentsAccessLevel:                   gitlab.PublicAccessControl,
			FeatureFlagsAccessLevel:                   gitlab.PublicAccessControl,
			InfrastructureAccessLevel:                 gitlab.PublicAccessControl,
			MonitorAccessLevel:                        gitlab.PublicAccessControl,
			PackageRegistryAccessLevel:                gitlab.PublicAccessControl,
			SecurityAndComplianceAccessLevel:          gitlab.PublicAccessControl,
			ModelExperimentsAccessLevel:               gitlab.PublicAccessControl,
			ModelRegistryAccessLevel:                  gitlab.PublicAccessControl,
			ResolveOutdatedDiffDiscussions:            f,
			ContainerRegistryEnabled:                  f, //nolint:staticcheck
			ContainerRegistryAccessLevel:              gitlab.PublicAccessControl,
			SharedRunnersEnabled:                      f,
			Visibility:                                gitlab.PublicVisibility,
			PublicBuilds:                              f, //nolint:staticcheck
			OnlyAllowMergeIfPipelineSucceeds:          f,
			OnlyAllowMergeIfAllDiscussionsAreResolved: f,
			MergeMethod:                               gitlab.FastForwardMerge,
			RemoveSourceBranchAfterMerge:              f,
			LFSEnabled:                                f,
			RequestAccessEnabled:                      f,
			TagList:                                   tags, //nolint:staticcheck
			Topics:                                    topics,
			CIConfigPath:                              s,
			CIDefaultGitDepth:                         i64,
			ApprovalsBeforeMerge:                      i64, //nolint:staticcheck
			Mirror:                                    f,
			MirrorUserID:                              i64,
			MirrorTriggerBuilds:                       f,
			OnlyMirrorProtectedBranches:               f,
			MirrorOverwritesDivergedBranches:          f,
			// packageRegistryAccessLevel public takes precedence over the
			// spec's packagesEnabled false and implies packages are enabled.
			PackagesEnabled:             true, //nolint:staticcheck
			ServiceDeskEnabled:          f,
			AutocloseReferencedIssues:   f,
			AllowMergeOnSkippedPipeline: f,
			CIForwardDeploymentEnabled:  f,
		}
		if name != baseline {
			reflect.ValueOf(gitlabProject).Elem().FieldByName(name).Set(reflect.ValueOf(value))
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
				MaxFileSize:                ptr.To(int64(0)),
				MemberCheck:                ptr.To(false),
				PreventSecrets:             ptr.To(false),
				RejectUnsignedCommits:      ptr.To(false),
				RejectNonDCOCommits:        ptr.To(false),
			}),
			withConditions(v2.Available()),
			// Some of the changed fields are also reported in the observation.
			withStatus(v1alpha1.ProjectObservation{
				Public:                      gitlabProject.PublicJobs,
				IssuesAccessLevel:           v1alpha1.AccessControlValue(gitlabProject.IssuesAccessLevel),
				BuildsAccessLevel:           v1alpha1.AccessControlValue(gitlabProject.BuildsAccessLevel),
				MergeRequestsAccessLevel:    v1alpha1.AccessControlValue(gitlabProject.MergeRequestsAccessLevel),
				SnippetsAccessLevel:         v1alpha1.AccessControlValue(gitlabProject.SnippetsAccessLevel),
				WikiAccessLevel:             v1alpha1.AccessControlValue(gitlabProject.WikiAccessLevel),
				MergeTrainsEnabled:          gitlabProject.MergeTrainsEnabled,
				MergeTrainsSkipTrainAllowed: gitlabProject.MergeTrainsSkipTrainAllowed,
				MergePipelinesEnabled:       gitlabProject.MergePipelinesEnabled,
				Archived:                    gitlabProject.Archived,
			}),
		}
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
							ID:            1,
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
					ResourceUpToDate:        name == baseline,
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
		cacheExternalPushRules *v1alpha1.PushRules
		cachePushRulesUpToDate bool
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
		"SuccessfulEditPushRules": {
			args: args{
				project: &fake.MockClient{
					MockEditProject: func(pid interface{}, opt *gitlab.EditProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{}, &gitlab.Response{}, nil
					},
					MockEditProjectPushRule: func(pid interface{}, opt *gitlab.EditProjectPushRuleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
						return &gitlab.ProjectPushRules{}, &gitlab.Response{}, nil
					},
				},
				cr: project(
					withStatus(v1alpha1.ProjectObservation{ID: 1234}),
					withProjectPushRules(&v1alpha1.PushRules{DenyDeleteTag: ptr.To(true)}),
				),
			},
			cacheExternalPushRules: &v1alpha1.PushRules{DenyDeleteTag: ptr.To(false)},
			cachePushRulesUpToDate: false,
			want: want{
				cr: project(
					withStatus(v1alpha1.ProjectObservation{ID: 1234}),
					withProjectPushRules(&v1alpha1.PushRules{DenyDeleteTag: ptr.To(true)}),
				),
			},
		},
		"SuccessfulAddPushRules": {
			args: args{
				project: &fake.MockClient{
					MockEditProject: func(pid interface{}, opt *gitlab.EditProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{}, &gitlab.Response{}, nil
					},
					MockAddProjectPushRule: func(pid interface{}, opt *gitlab.AddProjectPushRuleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
						return &gitlab.ProjectPushRules{}, &gitlab.Response{}, nil
					},
				},
				cr: project(
					withStatus(v1alpha1.ProjectObservation{ID: 1234}),
					withProjectPushRules(&v1alpha1.PushRules{DenyDeleteTag: ptr.To(true)}),
				),
			},
			cacheExternalPushRules: nil,
			cachePushRulesUpToDate: false,
			want: want{
				cr: project(
					withStatus(v1alpha1.ProjectObservation{ID: 1234}),
					withProjectPushRules(&v1alpha1.PushRules{DenyDeleteTag: ptr.To(true)}),
				),
			},
		},
		"FailedEditPushRules": {
			args: args{
				project: &fake.MockClient{
					MockEditProject: func(pid interface{}, opt *gitlab.EditProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{}, &gitlab.Response{}, nil
					},
					MockEditProjectPushRule: func(pid interface{}, opt *gitlab.EditProjectPushRuleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
						return &gitlab.ProjectPushRules{}, &gitlab.Response{}, errBoom
					},
				},
				cr: project(
					withStatus(v1alpha1.ProjectObservation{ID: 1234}),
					withProjectPushRules(&v1alpha1.PushRules{DenyDeleteTag: ptr.To(true)}),
				),
			},
			cacheExternalPushRules: &v1alpha1.PushRules{DenyDeleteTag: ptr.To(false)},
			cachePushRulesUpToDate: false,
			want: want{
				cr: project(
					withStatus(v1alpha1.ProjectObservation{ID: 1234}),
					withProjectPushRules(&v1alpha1.PushRules{DenyDeleteTag: ptr.To(true)}),
				),
				err: errors.Wrap(errBoom, errUpdatePushRulesFailed),
			},
		},
		"FailedAddPushRules": {
			args: args{
				project: &fake.MockClient{
					MockEditProject: func(pid interface{}, opt *gitlab.EditProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{}, &gitlab.Response{}, nil
					},
					MockAddProjectPushRule: func(pid interface{}, opt *gitlab.AddProjectPushRuleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
						return &gitlab.ProjectPushRules{}, &gitlab.Response{}, errBoom
					},
				},
				cr: project(
					withStatus(v1alpha1.ProjectObservation{ID: 1234}),
					withProjectPushRules(&v1alpha1.PushRules{DenyDeleteTag: ptr.To(true)}),
				),
			},
			cacheExternalPushRules: nil,
			cachePushRulesUpToDate: false,
			want: want{
				cr: project(
					withStatus(v1alpha1.ProjectObservation{ID: 1234}),
					withProjectPushRules(&v1alpha1.PushRules{DenyDeleteTag: ptr.To(true)}),
				),
				err: errors.Wrap(errBoom, errUpdatePushRulesFailed),
			},
		},
		"PushRulesUpToDateSkipped": {
			args: args{
				project: &fake.MockClient{
					MockEditProject: func(pid interface{}, opt *gitlab.EditProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
						return &gitlab.Project{}, &gitlab.Response{}, nil
					},
					// No push rule mocks needed - they should not be called
				},
				cr: project(
					withStatus(v1alpha1.ProjectObservation{ID: 1234}),
					withProjectPushRules(&v1alpha1.PushRules{DenyDeleteTag: ptr.To(true)}),
				),
			},
			cacheExternalPushRules: &v1alpha1.PushRules{DenyDeleteTag: ptr.To(true)},
			cachePushRulesUpToDate: true,
			want: want{
				cr: project(
					withStatus(v1alpha1.ProjectObservation{ID: 1234}),
					withProjectPushRules(&v1alpha1.PushRules{DenyDeleteTag: ptr.To(true)}),
				),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.project}
			e.cache.externalPushRules = tc.cacheExternalPushRules
			e.cache.isPushRulesUpToDate = tc.cachePushRulesUpToDate
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
					withPermanentlyRemove(new(true)),
					withPath(new("project")),
					withStatus(v1alpha1.ProjectObservation{PathWithNamespace: "path/to/project"}),
				),
			},
			want: want{
				cr: project(
					withExternalName("0"),
					withPermanentlyRemove(new(true)),
					withPath(new("project")),
					withStatus(v1alpha1.ProjectObservation{PathWithNamespace: "path/to/project"}),
				),
				calls: []deleteProjectCalls{
					{Pid: "0", Opt: &gitlab.DeleteProjectOptions{}},
					{Pid: "0", Opt: &gitlab.DeleteProjectOptions{PermanentlyRemove: new(true), FullPath: new("path/to/project")}},
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

// TestSanitizeImportURL verifies that sanitizeImportURL strips credentials
// from a URL the same way the GitLab API does when returning ImportURL in project
// responses. This is critical to prevent an infinite update loop when an
// ImportURLSecretRef is used: GitLab never echoes back the token/password, so
// the comparison must be done on the sanitized URL.
func TestSanitizeImportURL(t *testing.T) {
	cases := map[string]struct {
		input string
		want  string
	}{
		// GitLab pattern: TOKEN only (empty password), as in "https://TOKEN:@host/repo.git"
		"TokenWithEmptyPassword": {
			input: "https://GITHUB_TOKEN:@github.com/example/repo.git",
			want:  "https://github.com/example/repo.git",
		},
		// Username + password present
		"UsernameAndPassword": {
			input: "https://user:password@gitlab.com/group/repo.git",
			want:  "https://gitlab.com/group/repo.git",
		},
		// Username only (no colon)
		"UsernameOnly": {
			input: "https://user@gitlab.com/group/repo.git",
			want:  "https://gitlab.com/group/repo.git",
		},
		// No credentials present – must be returned unchanged
		"NoCredentials": {
			input: "https://github.com/example/repo.git",
			want:  "https://github.com/example/repo.git",
		},
		// SSH URL – no userinfo in the RFC 3986 sense; must pass through unchanged
		"SSHUrl": {
			input: "git@github.com:example/repo.git",
			want:  "git@github.com:example/repo.git",
		},
		// Unparseable string – must be returned as-is rather than panicking
		"Unparseable": {
			input: "://not a url",
			want:  "://not a url",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := sanitizeImportURL(tc.input)
			if got != tc.want {
				t.Errorf("sanitizeImportURL(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// TestLateInitializeSkipsMirrorFields guards against stamping mirror_user on
// non-mirrored projects: unset mirror fields must stay nil after late-init.
func TestLateInitializeSkipsMirrorFields(t *testing.T) {
	e := &external{client: &fake.MockClient{
		MockGetProjectPushRules: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
			return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, nil
		},
	}}
	cr := &v1alpha1.Project{}
	glProject := &gitlab.Project{
		Mirror:                           true,
		MirrorUserID:                     5,
		MirrorTriggerBuilds:              true,
		OnlyMirrorProtectedBranches:      true,
		MirrorOverwritesDivergedBranches: true,
	}

	if err := e.lateInitialize(context.Background(), cr, glProject); err != nil {
		t.Fatalf("lateInitialize returned error: %v", err)
	}

	p := cr.Spec.ForProvider
	if p.Mirror != nil || p.MirrorUserID != nil || p.MirrorTriggerBuilds != nil ||
		p.OnlyMirrorProtectedBranches != nil || p.MirrorOverwritesDivergedBranches != nil {
		t.Errorf("mirror fields must stay nil, got %+v", p)
	}
}

// TestLateInitializeProjectFeatureAccessLevels verifies that each feature
// access level is late-initialized from its own observed field, and that
// packageRegistryAccessLevel is not, because it takes precedence over
// packagesEnabled. The observed values are distinct per field so that reading
// one field from another fails.
func TestLateInitializeProjectFeatureAccessLevels(t *testing.T) {
	e := &external{client: &fake.MockClient{
		MockGetProjectPushRules: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
			return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errBoom
		},
	}}
	cr := &v1alpha1.Project{}
	glProject := &gitlab.Project{
		AnalyticsAccessLevel:             "analytics",
		EnvironmentsAccessLevel:          "environments",
		FeatureFlagsAccessLevel:          "feature-flags",
		InfrastructureAccessLevel:        "infrastructure",
		ModelExperimentsAccessLevel:      "model-experiments",
		ModelRegistryAccessLevel:         "model-registry",
		MonitorAccessLevel:               "monitor",
		PackageRegistryAccessLevel:       "package-registry",
		ReleasesAccessLevel:              "releases",
		SecurityAndComplianceAccessLevel: "security-and-compliance",
	}

	if err := e.lateInitialize(context.Background(), cr, glProject); err != nil {
		t.Fatalf("lateInitialize(...): %v", err)
	}

	level := func(s string) *v1alpha1.AccessControlValue { return ptr.To(v1alpha1.AccessControlValue(s)) }
	p := cr.Spec.ForProvider
	want := map[string]*v1alpha1.AccessControlValue{
		"AnalyticsAccessLevel":             level("analytics"),
		"EnvironmentsAccessLevel":          level("environments"),
		"FeatureFlagsAccessLevel":          level("feature-flags"),
		"InfrastructureAccessLevel":        level("infrastructure"),
		"ModelExperimentsAccessLevel":      level("model-experiments"),
		"ModelRegistryAccessLevel":         level("model-registry"),
		"MonitorAccessLevel":               level("monitor"),
		"PackageRegistryAccessLevel":       nil,
		"ReleasesAccessLevel":              level("releases"),
		"SecurityAndComplianceAccessLevel": level("security-and-compliance"),
	}
	got := map[string]*v1alpha1.AccessControlValue{
		"AnalyticsAccessLevel":             p.AnalyticsAccessLevel,
		"EnvironmentsAccessLevel":          p.EnvironmentsAccessLevel,
		"FeatureFlagsAccessLevel":          p.FeatureFlagsAccessLevel,
		"InfrastructureAccessLevel":        p.InfrastructureAccessLevel,
		"ModelExperimentsAccessLevel":      p.ModelExperimentsAccessLevel,
		"ModelRegistryAccessLevel":         p.ModelRegistryAccessLevel,
		"MonitorAccessLevel":               p.MonitorAccessLevel,
		"PackageRegistryAccessLevel":       p.PackageRegistryAccessLevel,
		"ReleasesAccessLevel":              p.ReleasesAccessLevel,
		"SecurityAndComplianceAccessLevel": p.SecurityAndComplianceAccessLevel,
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("lateInitialize(...): -want, +got:\n%s", diff)
	}
}

// packageRegistryGitLab is a fake GitLab project that stores the package
// registry settings the way GitLab's REST API does on an update: the access
// level is written directly, and only if it did not change does a changed
// packages_enabled reset it from the project visibility (private here). See
// ProjectFeaturesCompatibility and Project#set_package_registry_access_level.
type packageRegistryGitLab struct {
	level           gitlab.AccessControlValue
	packagesEnabled bool
	edits           int
}

func (g *packageRegistryGitLab) project() *gitlab.Project {
	return &gitlab.Project{
		Visibility:                 gitlab.PrivateVisibility,
		PackageRegistryAccessLevel: g.level,
		PackagesEnabled:            g.packagesEnabled, //nolint:staticcheck
	}
}

func (g *packageRegistryGitLab) edit(o *gitlab.EditProjectOptions) {
	g.edits++
	levelChanged := o.PackageRegistryAccessLevel != nil && *o.PackageRegistryAccessLevel != g.level
	if o.PackageRegistryAccessLevel != nil {
		g.level = *o.PackageRegistryAccessLevel
	}
	if o.PackagesEnabled != nil && *o.PackagesEnabled != g.packagesEnabled { //nolint:staticcheck
		g.packagesEnabled = *o.PackagesEnabled //nolint:staticcheck
		if !levelChanged {
			g.level = gitlab.DisabledAccessControl
			if g.packagesEnabled {
				g.level = gitlab.PrivateAccessControl
			}
		}
	}
}

// TestPackageRegistryConverges runs Observe and Update against a fake GitLab
// until the project is up to date, and checks that the package registry
// settings end where the spec asks for, within the expected number of edits.
func TestPackageRegistryConverges(t *testing.T) {
	cases := map[string]struct {
		reason       string
		spec         v1alpha1.ProjectParameters
		start        packageRegistryGitLab
		wantLevel    gitlab.AccessControlValue
		wantPackages bool
		wantEdits    int
	}{
		"DisableLevel": {
			reason:       "Disabling the level must also disable packages, which gate package access.",
			spec:         v1alpha1.ProjectParameters{PackageRegistryAccessLevel: ptr.To(v1alpha1.DisabledAccessControl)},
			start:        packageRegistryGitLab{level: gitlab.EnabledAccessControl, packagesEnabled: true},
			wantLevel:    gitlab.DisabledAccessControl,
			wantPackages: false,
			wantEdits:    1,
		},
		"EnableLevel": {
			reason:       "Enabling the level must also enable packages.",
			spec:         v1alpha1.ProjectParameters{PackageRegistryAccessLevel: ptr.To(v1alpha1.EnabledAccessControl)},
			start:        packageRegistryGitLab{level: gitlab.DisabledAccessControl, packagesEnabled: false},
			wantLevel:    gitlab.EnabledAccessControl,
			wantPackages: true,
			wantEdits:    1,
		},
		"RepairContradictoryState": {
			reason:       "If only packages_enabled changes, GitLab resets the level from the visibility; the next update restores it.",
			spec:         v1alpha1.ProjectParameters{PackageRegistryAccessLevel: ptr.To(v1alpha1.EnabledAccessControl)},
			start:        packageRegistryGitLab{level: gitlab.EnabledAccessControl, packagesEnabled: false},
			wantLevel:    gitlab.EnabledAccessControl,
			wantPackages: true,
			wantEdits:    2,
		},
		"PackagesEnabledWithoutLevel": {
			reason:       "Without a level, packagesEnabled still works as before.",
			spec:         v1alpha1.ProjectParameters{PackagesEnabled: ptr.To(false)},
			start:        packageRegistryGitLab{level: gitlab.PrivateAccessControl, packagesEnabled: true},
			wantLevel:    gitlab.DisabledAccessControl,
			wantPackages: false,
			wantEdits:    1,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			gl := tc.start
			e := &external{client: &fake.MockClient{
				MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
					return gl.project(), &gitlab.Response{}, nil
				},
				MockEditProject: func(pid interface{}, opt *gitlab.EditProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
					gl.edit(opt)
					return gl.project(), &gitlab.Response{}, nil
				},
				MockGetProjectPushRules: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
					return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errBoom
				},
			}}
			cr := project(withExternalName(extName), withSpec(tc.spec))

			upToDate := false
			for range 4 {
				obs, err := e.Observe(context.Background(), cr)
				if err != nil {
					t.Fatalf("Observe(...): %v", err)
				}
				if upToDate = obs.ResourceUpToDate; upToDate {
					break
				}
				if _, err := e.Update(context.Background(), cr); err != nil {
					t.Fatalf("Update(...): %v", err)
				}
			}

			if !upToDate {
				t.Fatalf("%s\nproject never became up to date: level %q, packages_enabled %t after %d edits", tc.reason, gl.level, gl.packagesEnabled, gl.edits)
			}
			got := struct {
				Level    gitlab.AccessControlValue
				Packages bool
				Edits    int
			}{gl.level, gl.packagesEnabled, gl.edits}
			want := struct {
				Level    gitlab.AccessControlValue
				Packages bool
				Edits    int
			}{tc.wantLevel, tc.wantPackages, tc.wantEdits}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("%s\nfinal GitLab state: -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}

// TestPackageRegistryLevelRemoval verifies that removing
// packageRegistryAccessLevel stops managing the package registry instead of
// re-applying a packagesEnabled value that late-initialization copied while
// the level was set.
func TestPackageRegistryLevelRemoval(t *testing.T) {
	gl := packageRegistryGitLab{level: gitlab.EnabledAccessControl, packagesEnabled: true}
	e := &external{client: &fake.MockClient{
		MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
			return gl.project(), &gitlab.Response{}, nil
		},
		MockEditProject: func(pid interface{}, opt *gitlab.EditProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
			gl.edit(opt)
			return gl.project(), &gitlab.Response{}, nil
		},
		MockGetProjectPushRules: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
			return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errBoom
		},
	}}
	cr := project(withExternalName(extName), withSpec(v1alpha1.ProjectParameters{
		PackageRegistryAccessLevel: ptr.To(v1alpha1.EnabledAccessControl),
	}))
	reconcile := func(step string) {
		t.Helper()
		for range 4 {
			obs, err := e.Observe(context.Background(), cr)
			if err != nil {
				t.Fatalf("%s: Observe(...): %v", step, err)
			}
			if obs.ResourceUpToDate {
				return
			}
			if _, err := e.Update(context.Background(), cr); err != nil {
				t.Fatalf("%s: Update(...): %v", step, err)
			}
		}
		t.Fatalf("%s: project never became up to date", step)
	}

	reconcile("level enabled")
	cr.Spec.ForProvider.PackageRegistryAccessLevel = ptr.To(v1alpha1.DisabledAccessControl)
	reconcile("level disabled")
	editsBefore := gl.edits
	cr.Spec.ForProvider.PackageRegistryAccessLevel = nil
	reconcile("level removed")

	got := struct {
		Level         gitlab.AccessControlValue
		Packages      bool
		EditsOnRemove int
	}{gl.level, gl.packagesEnabled, gl.edits - editsBefore}
	want := struct {
		Level         gitlab.AccessControlValue
		Packages      bool
		EditsOnRemove int
	}{gitlab.DisabledAccessControl, false, 0}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("after removing the level: -want, +got:\n%s", diff)
	}
}

// archiveGitLab is a fake GitLab project that records edit, archive, and
// unarchive calls in order. With AncestorArchived set, it rejects archive and
// unarchive calls the way GitLab does for a project in an archived group.
type archiveGitLab struct {
	Archived         bool
	AncestorArchived bool
	Description      string
	Calls            []string
}

var errAncestorArchived = errors.New("Cannot unarchive project since one of the ancestors is archived.")

func (g *archiveGitLab) client() *fake.MockClient {
	setArchived := func(call string, archived bool) (*gitlab.Project, *gitlab.Response, error) {
		g.Calls = append(g.Calls, call)
		if g.AncestorArchived {
			return nil, &gitlab.Response{}, errAncestorArchived
		}
		g.Archived = archived
		return &gitlab.Project{}, &gitlab.Response{}, nil
	}
	return &fake.MockClient{
		MockGetProject: func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
			return &gitlab.Project{Archived: g.Archived, Description: g.Description}, &gitlab.Response{}, nil
		},
		MockEditProject: func(pid interface{}, opt *gitlab.EditProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
			g.Calls = append(g.Calls, "edit")
			if opt.Description != nil {
				g.Description = *opt.Description
			}
			return &gitlab.Project{}, &gitlab.Response{}, nil
		},
		MockArchiveProject: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
			return setArchived("archive", true)
		},
		MockUnarchiveProject: func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
			return setArchived("unarchive", false)
		},
		MockGetProjectPushRules: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
			return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errBoom
		},
	}
}

// TestArchiveConverges runs Observe and Update against a fake GitLab until the
// project is up to date or Update fails, and checks the archive state, the
// applied edit, and the order of the GitLab calls.
func TestArchiveConverges(t *testing.T) {
	cases := map[string]struct {
		reason  string
		spec    v1alpha1.ProjectParameters
		start   archiveGitLab
		want    archiveGitLab
		wantErr error
	}{
		"ArchiveAfterEdit": {
			reason: "Archiving must run after the edit.",
			spec:   v1alpha1.ProjectParameters{Archived: new(true), Description: new("new")},
			start:  archiveGitLab{Description: "old"},
			want:   archiveGitLab{Archived: true, Description: "new", Calls: []string{"edit", "archive"}},
		},
		"UnarchiveBeforeEdit": {
			reason: "Unarchiving must run before the edit.",
			spec:   v1alpha1.ProjectParameters{Archived: new(false), Description: new("new")},
			start:  archiveGitLab{Archived: true, Description: "old"},
			want:   archiveGitLab{Description: "new", Calls: []string{"unarchive", "edit"}},
		},
		"EditStaysArchived": {
			reason: "A project that stays archived must not be unarchived to apply an edit.",
			spec:   v1alpha1.ProjectParameters{Archived: new(true), Description: new("new")},
			start:  archiveGitLab{Archived: true, Description: "old"},
			want:   archiveGitLab{Archived: true, Description: "new", Calls: []string{"edit"}},
		},
		"UnsetLeavesArchiveState": {
			reason: "Without archived in the spec, the archive state must not change.",
			spec:   v1alpha1.ProjectParameters{Description: new("new")},
			start:  archiveGitLab{Archived: true, Description: "old"},
			want:   archiveGitLab{Archived: true, Description: "new", Calls: []string{"edit"}},
		},
		"AncestorArchived": {
			reason:  "If GitLab rejects the unarchive, Update must fail before it edits.",
			spec:    v1alpha1.ProjectParameters{Archived: new(false), Description: new("new")},
			start:   archiveGitLab{Archived: true, AncestorArchived: true, Description: "old"},
			want:    archiveGitLab{Archived: true, AncestorArchived: true, Description: "old", Calls: []string{"unarchive"}},
			wantErr: errors.Wrap(errAncestorArchived, errUnarchiveFailed),
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			gl := tc.start
			e := &external{client: gl.client()}
			cr := project(withExternalName(extName), withSpec(tc.spec))

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
				t.Errorf("%s\nproject never became up to date: %+v", tc.reason, gl)
			}
			if diff := cmp.Diff(tc.want, gl); diff != "" {
				t.Errorf("%s\nfinal GitLab state: -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}
