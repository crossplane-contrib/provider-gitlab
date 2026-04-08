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

package projects

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	projectsv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
)

func TestRepositoryFileContentSHA256(t *testing.T) {
	plain := "hello world"
	encoded := base64.StdEncoding.EncodeToString([]byte(plain))
	plainSum := fmt.Sprintf("%x", sha256.Sum256([]byte(plain)))

	text := "text"
	base64Encoding := "base64"

	cases := map[string]struct {
		params   *projectsv1alpha1.RepositoryFileParameters
		content  string
		wantHash string
	}{
		"TextContent": {
			params:   &projectsv1alpha1.RepositoryFileParameters{Encoding: &text},
			content:  plain,
			wantHash: plainSum,
		},
		"Base64Content": {
			params:   &projectsv1alpha1.RepositoryFileParameters{Encoding: &base64Encoding},
			content:  encoded,
			wantHash: plainSum,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := RepositoryFileContentSHA256(tc.params, tc.content)
			if diff := cmp.Diff(tc.wantHash, got); diff != "" {
				t.Fatalf("RepositoryFileContentSHA256(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestRepositoryFileReconcileInterval(t *testing.T) {
	oneHour := "1h"
	cases := map[string]struct {
		params      *projectsv1alpha1.RepositoryFileParameters
		defaultPoll time.Duration
		want        time.Duration
		wantErr     bool
	}{
		"Default": {
			params:      &projectsv1alpha1.RepositoryFileParameters{},
			defaultPoll: time.Minute,
			want:        time.Minute,
		},
		"Custom": {
			params:      &projectsv1alpha1.RepositoryFileParameters{ReconcileInterval: &oneHour},
			defaultPoll: time.Minute,
			want:        time.Hour,
		},
		"Invalid": {
			params:      &projectsv1alpha1.RepositoryFileParameters{ReconcileInterval: stringPtr("nope")},
			defaultPoll: time.Minute,
			wantErr:     true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := RepositoryFileReconcileInterval(tc.params, tc.defaultPoll)
			if tc.wantErr {
				if err == nil {
					t.Fatal("RepositoryFileReconcileInterval() expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("RepositoryFileReconcileInterval() error = %v", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Fatalf("RepositoryFileReconcileInterval(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateRepositoryFileOptions(t *testing.T) {
	content := "content"
	branch := "main"
	startBranch := "main"
	authorEmail := "dev@example.org"
	authorName := "Dev"
	createCommit := "create"
	updateCommit := "update"
	deleteCommit := "delete"
	encoding := "base64"
	exec := true
	lastCommitID := "abc123"

	params := &projectsv1alpha1.RepositoryFileParameters{
		Branch:              branch,
		StartBranch:         &startBranch,
		AuthorEmail:         &authorEmail,
		AuthorName:          &authorName,
		CreateCommitMessage: &createCommit,
		UpdateCommitMessage: &updateCommit,
		DeleteCommitMessage: &deleteCommit,
		Encoding:            &encoding,
		ExecuteFilemode:     &exec,
		FilePath:            "README.md",
	}

	create := GenerateCreateFileOptions(params, content)
	update := GenerateUpdateFileOptions(params, content, &lastCommitID)
	deleteOpts := GenerateDeleteFileOptions(params, &lastCommitID)

	if diff := cmp.Diff(&gitlab.CreateFileOptions{
		Branch:          &branch,
		StartBranch:     &startBranch,
		Encoding:        &encoding,
		AuthorEmail:     &authorEmail,
		AuthorName:      &authorName,
		Content:         &content,
		CommitMessage:   &createCommit,
		ExecuteFilemode: &exec,
	}, create); diff != "" {
		t.Fatalf("GenerateCreateFileOptions(): -want, +got:\n%s", diff)
	}

	if diff := cmp.Diff(&gitlab.UpdateFileOptions{
		Branch:          &branch,
		StartBranch:     &startBranch,
		Encoding:        &encoding,
		AuthorEmail:     &authorEmail,
		AuthorName:      &authorName,
		Content:         &content,
		CommitMessage:   &updateCommit,
		LastCommitID:    &lastCommitID,
		ExecuteFilemode: &exec,
	}, update); diff != "" {
		t.Fatalf("GenerateUpdateFileOptions(): -want, +got:\n%s", diff)
	}

	if diff := cmp.Diff(&gitlab.DeleteFileOptions{
		Branch:        &branch,
		StartBranch:   &startBranch,
		AuthorEmail:   &authorEmail,
		AuthorName:    &authorName,
		CommitMessage: &deleteCommit,
		LastCommitID:  &lastCommitID,
	}, deleteOpts); diff != "" {
		t.Fatalf("GenerateDeleteFileOptions(): -want, +got:\n%s", diff)
	}
}

func TestIsRepositoryFileUpToDate(t *testing.T) {
	content := "hello"
	branch := "main"
	sha := RepositoryFileContentSHA256(&projectsv1alpha1.RepositoryFileParameters{}, content)

	cases := map[string]struct {
		params   *projectsv1alpha1.RepositoryFileParameters
		external *gitlab.File
		content  string
		want     bool
	}{
		"Matching": {
			params: &projectsv1alpha1.RepositoryFileParameters{
				FilePath: "README.md",
				Branch:   branch,
			},
			external: &gitlab.File{FilePath: "README.md", Ref: branch, SHA256: sha},
			content:  content,
			want:     true,
		},
		"DifferentSHA": {
			params: &projectsv1alpha1.RepositoryFileParameters{
				FilePath: "README.md",
				Branch:   branch,
			},
			external: &gitlab.File{FilePath: "README.md", Ref: branch, SHA256: "different"},
			content:  content,
			want:     false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsRepositoryFileUpToDate(tc.params, tc.external, tc.content)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Fatalf("IsRepositoryFileUpToDate(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestRepositoryFilePoliciesEqual(t *testing.T) {
	a := xpv1.ManagementPolicies{xpv1.ManagementActionObserve, xpv1.ManagementActionCreate, xpv1.ManagementActionDelete}
	b := xpv1.ManagementPolicies{xpv1.ManagementActionDelete, xpv1.ManagementActionObserve, xpv1.ManagementActionCreate}
	c := xpv1.ManagementPolicies{xpv1.ManagementActionAll}

	if !RepositoryFilePoliciesEqual(a, b) {
		t.Fatal("RepositoryFilePoliciesEqual() = false, want true")
	}
	if RepositoryFilePoliciesEqual(a, c) {
		t.Fatal("RepositoryFilePoliciesEqual() = true, want false")
	}
}

func TestRepositoryFileCreateOnlyPolicies(t *testing.T) {
	want := xpv1.ManagementPolicies{xpv1.ManagementActionObserve, xpv1.ManagementActionCreate, xpv1.ManagementActionDelete}
	if diff := cmp.Diff(want, RepositoryFileCreateOnlyPolicies()); diff != "" {
		t.Fatalf("RepositoryFileCreateOnlyPolicies(): -want, +got:\n%s", diff)
	}
}

func stringPtr(s string) *string {
	return &s
}
