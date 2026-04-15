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
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	projectsv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	commonpkg "github.com/crossplane-contrib/provider-gitlab/pkg/common"
	projectclients "github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects/fake"
)

const testContent = "hello"

func TestResolveContent(t *testing.T) {
	content := testContent
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
	content := testContent
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

func TestObserveCreateOnlyIgnoresDrift(t *testing.T) {
	projectID := "123"
	content := testContent
	createOnly := true
	cr := &projectsv1alpha1.RepositoryFile{
		Spec: projectsv1alpha1.RepositoryFileSpec{
			ForProvider: projectsv1alpha1.RepositoryFileParameters{
				ProjectID:  &projectID,
				FilePath:   "README.md",
				Branch:     "main",
				Content:    &content,
				CreateOnly: &createOnly,
			},
		},
	}

	e := &external{client: &fake.MockClient{
		MockGetFile: func(pid any, fileName string, opt *gitlab.GetFileOptions, options ...gitlab.RequestOptionFunc) (*gitlab.File, *gitlab.Response, error) {
			return &gitlab.File{FilePath: "README.md", Ref: "main", SHA256: "different"}, &gitlab.Response{}, nil
		},
	}}
	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("Observe() error = %v", err)
	}
	if diff := cmp.Diff(true, obs.ResourceUpToDate); diff != "" {
		t.Fatalf("Observe() upToDate: -want, +got:\n%s", diff)
	}
}

func TestCreateOnlyInitializer(t *testing.T) {
	createOnly := true
	cr := &projectsv1alpha1.RepositoryFile{}
	cr.Spec.ForProvider.CreateOnly = &createOnly

	updated := false
	init := createOnlyInitializer{client: &test.MockClient{
		MockUpdate: func(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
			updated = true
			return nil
		},
	}}

	if err := init.Initialize(context.Background(), resource.Managed(cr)); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	if !updated {
		t.Fatal("Initialize() did not persist policy")
	}
	if diff := cmp.Diff(projectclients.RepositoryFileCreateOnlyPolicies(), cr.GetManagementPolicies()); diff != "" {
		t.Fatalf("Initialize(): -want, +got:\n%s", diff)
	}
}

func TestCreateOnlyInitializerConflict(t *testing.T) {
	createOnly := true
	cr := &projectsv1alpha1.RepositoryFile{}
	cr.Spec.ForProvider.CreateOnly = &createOnly
	cr.SetManagementPolicies(xpv1.ManagementPolicies{xpv1.ManagementActionAll})

	init := createOnlyInitializer{client: &test.MockClient{}}
	err := init.Initialize(context.Background(), resource.Managed(cr))
	if err == nil || err.Error() != errCreateOnlyPolicyConflict {
		t.Fatalf("Initialize() error = %v, want %s", err, errCreateOnlyPolicyConflict)
	}
}

func TestObserveExternalNameMismatch(t *testing.T) {
	projectID := "123"
	content := testContent
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
	content := testContent
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
}
