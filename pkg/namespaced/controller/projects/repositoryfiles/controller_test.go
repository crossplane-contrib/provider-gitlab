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
	"net/http"
	"testing"
	"time"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	projectsv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	commonpkg "github.com/crossplane-contrib/provider-gitlab/pkg/common"
	projectclients "github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects/fake"
)

func TestResolveContent(t *testing.T) {
	content := "hello"
	cr := &projectsv1alpha1.RepositoryFile{}
	cr.Namespace = "default"

	e := &external{}
	got, err := e.resolveContent(context.Background(), cr, &projectsv1alpha1.RepositoryFileParameters{Content: &content})
	if err != nil {
		t.Fatalf("resolveContent() error = %v", err)
	}
	if diff := cmp.Diff(content, got); diff != "" {
		t.Fatalf("resolveContent(): -want, +got:\n%s", diff)
	}
}

func TestResolveContentFromSecret(t *testing.T) {
	content := "hello"
	cr := &projectsv1alpha1.RepositoryFile{}
	cr.Namespace = "default"

	e := &external{kube: &test.MockClient{
		MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
			secret := obj.(*corev1.Secret)
			secret.Data = map[string][]byte{"content": []byte(content)}
			return nil
		},
	}}

	got, err := e.resolveContent(context.Background(), cr, &projectsv1alpha1.RepositoryFileParameters{
		ContentSecretRef: commonpkg.TestCreateLocalSecretKeySelector("", "content"),
	})
	if err != nil {
		t.Fatalf("resolveContent() error = %v", err)
	}
	if diff := cmp.Diff(content, got); diff != "" {
		t.Fatalf("resolveContent(): -want, +got:\n%s", diff)
	}
}

func TestObserveCreateOnlyAndInterval(t *testing.T) {
	projectID := "123"
	filePath := "README.md"
	branch := "main"
	content := "hello"
	createOnly := true
	observeAt := metav1.NewTime(time.Now())

	cr := &projectsv1alpha1.RepositoryFile{
		Spec: projectsv1alpha1.RepositoryFileSpec{
			ForProvider: projectsv1alpha1.RepositoryFileParameters{
				ProjectID:         &projectID,
				FilePath:          filePath,
				Branch:            branch,
				Content:           &content,
				CreateOnly:        &createOnly,
				ReconcileInterval: stringPtr("1h"),
			},
		},
		Status: projectsv1alpha1.RepositoryFileStatus{
			AtProvider: projectsv1alpha1.RepositoryFileObservation{
				LastObserveTime: &observeAt,
			},
		},
	}

	e := &external{pollInterval: time.Minute, client: &fake.MockClient{}}
	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("Observe() error = %v", err)
	}
	if diff := cmp.Diff(managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, obs); diff != "" {
		t.Fatalf("Observe(): -want, +got:\n%s", diff)
	}
}

func TestObserveExternalNameMismatch(t *testing.T) {
	projectID := "123"
	content := "hello"
	cr := &projectsv1alpha1.RepositoryFile{
		Spec: projectsv1alpha1.RepositoryFileSpec{
			ForProvider: projectsv1alpha1.RepositoryFileParameters{
				ProjectID: &projectID,
				FilePath:  "README.md",
				Branch:    "main",
				Content:   &content,
			},
		},
	}
	meta.SetExternalName(cr, "DIFFERENT.md")

	e := &external{pollInterval: time.Minute, client: &fake.MockClient{}}
	_, err := e.Observe(context.Background(), cr)
	if err == nil || err.Error() != errExternalNameMismatch {
		t.Fatalf("Observe() error = %v, want %s", err, errExternalNameMismatch)
	}
}

func TestDeleteIgnores404(t *testing.T) {
	projectID := "123"
	content := "hello"
	cr := &projectsv1alpha1.RepositoryFile{
		Spec: projectsv1alpha1.RepositoryFileSpec{
			ForProvider: projectsv1alpha1.RepositoryFileParameters{
				ProjectID: &projectID,
				FilePath:  "README.md",
				Branch:    "main",
				Content:   &content,
			},
		},
	}

	e := &external{client: &fake.MockClient{
		MockDeleteFile: func(pid any, fileName string, opt *gitlab.DeleteFileOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
			return &gitlab.Response{Response: &http.Response{StatusCode: 404}}, context.DeadlineExceeded
		},
	}}

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if got := cr.Status.ConditionedStatus.Conditions[0].Reason; got != xpv1.Deleting().Reason {
		t.Fatalf("Delete() condition reason = %s, want %s", got, xpv1.Deleting().Reason)
	}
	_ = projectclients.RepositoryFileCreateOnly
}

func stringPtr(s string) *string {
	return &s
}
