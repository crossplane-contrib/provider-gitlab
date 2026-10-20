/*
Copyright 2026 The Crossplane Authors.

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

package repositoryfiles

import (
	"context"
	"encoding/base64"
	"net/http"
	"testing"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects/fake"
)

const (
	testProjectID = "1234"
	testBranch    = "main"
	testFilePath  = "docs/a.md"
	testCommitID  = "c0ffee"
	// sha256 of "hello".
	helloSHA256 = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
)

var (
	errBoom = errors.New("boom")

	notFound    = &gitlab.Response{Response: &http.Response{StatusCode: http.StatusNotFound, Header: http.Header{"Content-Type": {"application/json"}}}}
	serverError = &gitlab.Response{Response: &http.Response{StatusCode: http.StatusInternalServerError}}

	// kubeWithContent serves the Secret referenced by withContentSecret, which
	// must be looked up in the namespace of the RepositoryFile.
	kubeWithContent = &test.MockClient{
		MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
			if key != (client.ObjectKey{Namespace: "default", Name: "file-content"}) {
				return errors.Errorf("unexpected Secret %s", key)
			}
			s, ok := obj.(*corev1.Secret)
			if !ok {
				return errors.New("unexpected object")
			}
			*s = corev1.Secret{Data: map[string][]byte{"content": []byte("hello")}}
			return nil
		},
	}

	// kubeBoom fails every read, so tests can assert content is not resolved.
	kubeBoom = &test.MockClient{MockGet: test.NewMockGetFn(errBoom)}
)

type modifier func(*v1alpha1.RepositoryFile)

func withContent(c string) modifier {
	return func(cr *v1alpha1.RepositoryFile) { cr.Spec.ForProvider.Content = &c }
}

func withContentSecret() modifier {
	return func(cr *v1alpha1.RepositoryFile) {
		cr.Spec.ForProvider.Content = nil
		cr.Spec.ForProvider.ContentSecretRef = common.TestCreateLocalSecretKeySelector("file-content", "content")
	}
}

func withoutContent() modifier {
	return func(cr *v1alpha1.RepositoryFile) {
		cr.Spec.ForProvider.Content = nil
		cr.Spec.ForProvider.ContentSecretRef = nil
	}
}

func withoutProjectID() modifier {
	return func(cr *v1alpha1.RepositoryFile) { cr.Spec.ForProvider.ProjectID = nil }
}

func withExecuteFilemode(b *bool) modifier {
	return func(cr *v1alpha1.RepositoryFile) { cr.Spec.ForProvider.ExecuteFilemode = b }
}

func withPolicies(p ...v2.ManagementAction) modifier {
	return func(cr *v1alpha1.RepositoryFile) { cr.SetManagementPolicies(p) }
}

func withObservation(o v1alpha1.RepositoryFileObservation) modifier {
	return func(cr *v1alpha1.RepositoryFile) { cr.Status.AtProvider = o }
}

func withConditions(c ...v2.Condition) modifier {
	return func(cr *v1alpha1.RepositoryFile) { cr.Status.SetConditions(c...) }
}

func withDeletionTimestamp() modifier {
	return func(cr *v1alpha1.RepositoryFile) { cr.SetDeletionTimestamp(&metav1.Time{Time: time.Unix(1, 0)}) }
}

func repositoryFile(m ...modifier) *v1alpha1.RepositoryFile {
	cr := &v1alpha1.RepositoryFile{
		ObjectMeta: metav1.ObjectMeta{Name: "file", Namespace: "default"},
		Spec: v1alpha1.RepositoryFileSpec{
			ForProvider: v1alpha1.RepositoryFileParameters{
				ProjectID:       new(testProjectID),
				Branch:          testBranch,
				FilePath:        testFilePath,
				Content:         new("hello"),
				ExecuteFilemode: new(false),
			},
		},
	}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var observed = v1alpha1.RepositoryFileObservation{
	FilePath:     testFilePath,
	BlobID:       "blob",
	LastCommitID: testCommitID,
	SHA256:       helloSHA256,
	Size:         5,
}

func observedFile() *gitlab.File {
	return &gitlab.File{
		FilePath:     testFilePath,
		BlobID:       "blob",
		CommitID:     "head",
		LastCommitID: testCommitID,
		SHA256:       helloSHA256,
		Size:         5,
		Ref:          testBranch,
		Encoding:     "base64",
	}
}

// getFileMetaData returns a mock that asserts it was asked for the test file
// on the test branch.
func getFileMetaData(t *testing.T, f *gitlab.File, res *gitlab.Response, err error) func(pid any, fileName string, opt *gitlab.GetFileMetaDataOptions, options ...gitlab.RequestOptionFunc) (*gitlab.File, *gitlab.Response, error) {
	t.Helper()
	return func(pid any, fileName string, opt *gitlab.GetFileMetaDataOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.File, *gitlab.Response, error) {
		want := &gitlab.GetFileMetaDataOptions{Ref: new(testBranch)}
		if pid != testProjectID || fileName != testFilePath || !cmp.Equal(want, opt) {
			t.Errorf("GetFileMetaData(%v, %q, %+v): unexpected arguments", pid, fileName, opt)
		}
		return f, res, err
	}
}

func TestObserve(t *testing.T) {
	type args struct {
		kube                      client.Client
		client                    projects.RepositoryFileClient
		managementPoliciesEnabled bool
		mg                        resource.Managed
	}
	type want struct {
		cr     *v1alpha1.RepositoryFile
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"NotRepositoryFile": {
			reason: "An unexpected managed resource should be an error.",
			args:   args{mg: nil},
			want:   want{err: errors.New(errNotRepositoryFile)},
		},
		"ProjectIDMissing": {
			reason: "An unresolved projectId should be an error.",
			args: args{
				mg: repositoryFile(withoutProjectID()),
			},
			want: want{
				cr:  repositoryFile(withoutProjectID()),
				err: errors.New(errProjectIDMissing),
			},
		},
		"NotFound": {
			reason: "A 404 means the file does not exist.",
			args: args{
				client: &fake.MockClient{MockGetFileMetaData: getFileMetaData(t, nil, notFound, errBoom)},
				mg:     repositoryFile(),
			},
			want: want{
				cr:     repositoryFile(),
				result: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"GetFailed": {
			reason: "Other GitLab errors should be returned.",
			args: args{
				client: &fake.MockClient{MockGetFileMetaData: getFileMetaData(t, nil, serverError, errBoom)},
				mg:     repositoryFile(),
			},
			want: want{
				cr:  repositoryFile(),
				err: errors.Wrap(errBoom, errGetFailed),
			},
		},
		"UpToDate": {
			reason: "Matching content should be up to date.",
			args: args{
				client: &fake.MockClient{MockGetFileMetaData: getFileMetaData(t, observedFile(), &gitlab.Response{}, nil)},
				mg:     repositoryFile(),
			},
			want: want{
				cr:     repositoryFile(withObservation(observed), withConditions(v2.Available())),
				result: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true},
			},
		},
		"ContentDrift": {
			reason: "Different content should not be up to date.",
			args: args{
				client: &fake.MockClient{MockGetFileMetaData: getFileMetaData(t, observedFile(), &gitlab.Response{}, nil)},
				mg:     repositoryFile(withContent("hello world")),
			},
			want: want{
				cr:     repositoryFile(withContent("hello world"), withObservation(observed), withConditions(v2.Available())),
				result: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false},
			},
		},
		"LateInitFilemode": {
			reason: "An unset filemode should be late initialized from GitLab.",
			args: args{
				client: &fake.MockClient{MockGetFileMetaData: getFileMetaData(t, observedFile(), &gitlab.Response{}, nil)},
				mg:     repositoryFile(withExecuteFilemode(nil)),
			},
			want: want{
				cr:     repositoryFile(withObservation(observed), withConditions(v2.Available())),
				result: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true, ResourceLateInitialized: true},
			},
		},
		"SecretContent": {
			reason: "Content from a Secret should be compared like inline content.",
			args: args{
				kube:   kubeWithContent,
				client: &fake.MockClient{MockGetFileMetaData: getFileMetaData(t, observedFile(), &gitlab.Response{}, nil)},
				mg:     repositoryFile(withContentSecret()),
			},
			want: want{
				cr:     repositoryFile(withContentSecret(), withObservation(observed), withConditions(v2.Available())),
				result: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true},
			},
		},
		"SecretFailed": {
			reason: "A Secret that cannot be read should be an error when updates are allowed.",
			args: args{
				kube:   kubeBoom,
				client: &fake.MockClient{MockGetFileMetaData: getFileMetaData(t, observedFile(), &gitlab.Response{}, nil)},
				mg:     repositoryFile(withContentSecret()),
			},
			want: want{
				cr:  repositoryFile(withContentSecret(), withObservation(observed), withConditions(v2.Available())),
				err: errors.Wrap(errors.Wrap(errors.Wrap(errBoom, common.ErrSecretNotFound), errContentSecretFailed), errContentFailed),
			},
		},
		"UpdateNotAllowed": {
			reason: "Without the Update policy content must not be resolved, so its source may be gone.",
			args: args{
				kube:                      kubeBoom,
				client:                    &fake.MockClient{MockGetFileMetaData: getFileMetaData(t, observedFile(), &gitlab.Response{}, nil)},
				managementPoliciesEnabled: true,
				mg:                        repositoryFile(withContentSecret(), withPolicies(v2.ManagementActionObserve, v2.ManagementActionCreate)),
			},
			want: want{
				cr: repositoryFile(withContentSecret(), withPolicies(v2.ManagementActionObserve, v2.ManagementActionCreate),
					withObservation(observed), withConditions(v2.Available())),
				result: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true},
			},
		},
		"BeingDeleted": {
			reason: "Content must not be resolved during deletion, so a Secret deleted first cannot block Delete.",
			args: args{
				kube:   kubeBoom,
				client: &fake.MockClient{MockGetFileMetaData: getFileMetaData(t, observedFile(), &gitlab.Response{}, nil)},
				mg:     repositoryFile(withContentSecret(), withDeletionTimestamp()),
			},
			want: want{
				cr:     repositoryFile(withContentSecret(), withDeletionTimestamp(), withObservation(observed), withConditions(v2.Available())),
				result: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true},
			},
		},
		"ObserveOnlyWithoutContent": {
			reason: "An observe-only file needs no content.",
			args: args{
				client:                    &fake.MockClient{MockGetFileMetaData: getFileMetaData(t, observedFile(), &gitlab.Response{}, nil)},
				managementPoliciesEnabled: true,
				mg:                        repositoryFile(withoutContent(), withPolicies(v2.ManagementActionObserve)),
			},
			want: want{
				cr: repositoryFile(withoutContent(), withPolicies(v2.ManagementActionObserve),
					withObservation(observed), withConditions(v2.Available())),
				result: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.args.kube, client: tc.args.client, managementPoliciesEnabled: tc.args.managementPoliciesEnabled}
			got, err := e.Observe(context.Background(), tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.result, got); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s", tc.reason, diff)
			}
			if tc.want.cr == nil {
				return
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want cr, +got cr:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type args struct {
		kube   client.Client
		client projects.RepositoryFileClient
		mg     resource.Managed
	}
	type want struct {
		cr  *v1alpha1.RepositoryFile
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"NotRepositoryFile": {
			reason: "An unexpected managed resource should be an error.",
			args:   args{mg: nil},
			want:   want{err: errors.New(errNotRepositoryFile)},
		},
		"ContentMissing": {
			reason: "Creating a file without content should be an error.",
			args:   args{mg: repositoryFile(withoutContent())},
			want: want{
				cr:  repositoryFile(withoutContent()),
				err: errors.Wrap(errors.New(errContentMissing), errContentFailed),
			},
		},
		"Successful": {
			reason: "The file should be created on its branch with base64 content and without start_branch.",
			args: args{
				client: &fake.MockClient{
					MockCreateFile: func(pid any, fileName string, opt *gitlab.CreateFileOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.FileInfo, *gitlab.Response, error) {
						want := &gitlab.CreateFileOptions{
							Branch:          new(testBranch),
							Encoding:        new("base64"),
							Content:         new(base64.StdEncoding.EncodeToString([]byte("hello"))),
							CommitMessage:   new("crossplane: create " + testFilePath),
							ExecuteFilemode: new(false),
						}
						if pid != testProjectID || fileName != testFilePath {
							t.Errorf("CreateFile(%v, %q, ...): unexpected arguments", pid, fileName)
						}
						if diff := cmp.Diff(want, opt); diff != "" {
							t.Errorf("CreateFile(...): -want options, +got options:\n%s", diff)
						}
						return &gitlab.FileInfo{}, &gitlab.Response{}, nil
					},
				},
				mg: repositoryFile(),
			},
			want: want{
				cr: repositoryFile(withConditions(v2.Creating())),
			},
		},
		"Failed": {
			reason: "GitLab errors should be returned.",
			args: args{
				client: &fake.MockClient{
					MockCreateFile: func(_ any, _ string, _ *gitlab.CreateFileOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.FileInfo, *gitlab.Response, error) {
						return nil, serverError, errBoom
					},
				},
				mg: repositoryFile(),
			},
			want: want{
				cr:  repositoryFile(withConditions(v2.Creating())),
				err: errors.Wrap(errBoom, errCreateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.args.kube, client: tc.args.client}
			_, err := e.Create(context.Background(), tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Create(...): -want error, +got error:\n%s", tc.reason, diff)
			}
			if tc.want.cr == nil {
				return
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("\n%s\ne.Create(...): -want cr, +got cr:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type args struct {
		client projects.RepositoryFileClient
		mg     resource.Managed
	}

	cases := map[string]struct {
		reason string
		args   args
		want   error
	}{
		"NotRepositoryFile": {
			reason: "An unexpected managed resource should be an error.",
			args:   args{mg: nil},
			want:   errors.New(errNotRepositoryFile),
		},
		"Successful": {
			reason: "The update should be guarded by the last observed commit.",
			args: args{
				client: &fake.MockClient{
					MockUpdateFile: func(pid any, fileName string, opt *gitlab.UpdateFileOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.FileInfo, *gitlab.Response, error) {
						want := &gitlab.UpdateFileOptions{
							Branch:          new(testBranch),
							Encoding:        new("base64"),
							Content:         new(base64.StdEncoding.EncodeToString([]byte("hello"))),
							CommitMessage:   new("crossplane: update " + testFilePath),
							LastCommitID:    new(testCommitID),
							ExecuteFilemode: new(false),
						}
						if pid != testProjectID || fileName != testFilePath {
							t.Errorf("UpdateFile(%v, %q, ...): unexpected arguments", pid, fileName)
						}
						if diff := cmp.Diff(want, opt); diff != "" {
							t.Errorf("UpdateFile(...): -want options, +got options:\n%s", diff)
						}
						return &gitlab.FileInfo{}, &gitlab.Response{}, nil
					},
				},
				mg: repositoryFile(withObservation(observed)),
			},
		},
		"Failed": {
			reason: "GitLab errors should be returned.",
			args: args{
				client: &fake.MockClient{
					MockUpdateFile: func(_ any, _ string, _ *gitlab.UpdateFileOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.FileInfo, *gitlab.Response, error) {
						return nil, serverError, errBoom
					},
				},
				mg: repositoryFile(withObservation(observed)),
			},
			want: errors.Wrap(errBoom, errUpdateFailed),
		},
		"ContentFailed": {
			reason: "Content that cannot be resolved must be an error rather than an empty commit.",
			args: args{
				client: &fake.MockClient{},
				mg:     repositoryFile(withoutContent(), withObservation(observed)),
			},
			want: errors.Wrap(errors.New(errContentMissing), errContentFailed),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.args.client}
			_, err := e.Update(context.Background(), tc.args.mg)
			if diff := cmp.Diff(tc.want, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Update(...): -want error, +got error:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		client projects.RepositoryFileClient
		mg     resource.Managed
	}
	type want struct {
		cr  *v1alpha1.RepositoryFile
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"NotRepositoryFile": {
			reason: "An unexpected managed resource should be an error.",
			args:   args{mg: nil},
			want:   want{err: errors.New(errNotRepositoryFile)},
		},
		"Successful": {
			reason: "The delete should be guarded by the last observed commit.",
			args: args{
				client: &fake.MockClient{
					MockDeleteFile: func(pid any, fileName string, opt *gitlab.DeleteFileOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						want := &gitlab.DeleteFileOptions{
							Branch:        new(testBranch),
							CommitMessage: new("crossplane: delete " + testFilePath),
							LastCommitID:  new(testCommitID),
						}
						if pid != testProjectID || fileName != testFilePath {
							t.Errorf("DeleteFile(%v, %q, ...): unexpected arguments", pid, fileName)
						}
						if diff := cmp.Diff(want, opt); diff != "" {
							t.Errorf("DeleteFile(...): -want options, +got options:\n%s", diff)
						}
						return &gitlab.Response{}, nil
					},
				},
				mg: repositoryFile(withObservation(observed)),
			},
			want: want{
				cr: repositoryFile(withObservation(observed), withConditions(v2.Deleting())),
			},
		},
		"NotFound": {
			reason: "A 404 (e.g. the branch or project is gone) leaves nothing to delete.",
			args: args{
				client: &fake.MockClient{
					MockDeleteFile: func(_ any, _ string, _ *gitlab.DeleteFileOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return notFound, errBoom
					},
				},
				mg: repositoryFile(),
			},
			want: want{
				cr: repositoryFile(withConditions(v2.Deleting())),
			},
		},
		"Failed": {
			reason: "Other GitLab errors should be returned.",
			args: args{
				client: &fake.MockClient{
					MockDeleteFile: func(_ any, _ string, _ *gitlab.DeleteFileOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return serverError, errBoom
					},
				},
				mg: repositoryFile(),
			},
			want: want{
				cr:  repositoryFile(withConditions(v2.Deleting())),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.args.client}
			_, err := e.Delete(context.Background(), tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Delete(...): -want error, +got error:\n%s", tc.reason, diff)
			}
			if tc.want.cr == nil {
				return
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("\n%s\ne.Delete(...): -want cr, +got cr:\n%s", tc.reason, diff)
			}
		})
	}
}
