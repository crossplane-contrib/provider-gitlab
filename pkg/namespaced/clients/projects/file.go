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
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

const (
	errFileNotFound = "404 Not found"
)

// FileClient defines Gitlab File service operations
type FileClient interface {
	GetFile(pid any, fileName string, opt *gitlab.GetFileOptions, options ...gitlab.RequestOptionFunc) (*gitlab.File, *gitlab.Response, error)
	CreateFile(pid any, fileName string, opt *gitlab.CreateFileOptions, options ...gitlab.RequestOptionFunc) (*gitlab.FileInfo, *gitlab.Response, error)
	UpdateFile(pid any, fileName string, opt *gitlab.UpdateFileOptions, options ...gitlab.RequestOptionFunc) (*gitlab.FileInfo, *gitlab.Response, error)
	DeleteFile(pid any, fileName string, opt *gitlab.DeleteFileOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// CommitsClient defines Gitlab Commit service operations
type CommitClient interface {
	GetCommit(pid any, sha string, opt *gitlab.GetCommitOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Commit, *gitlab.Response, error)
}

// NewFileClient returns a new Gitlab File service
func NewFileClient(cfg common.Config) FileClient {
	git := common.NewClient(cfg)
	return git.RepositoryFiles
}

// NewCommitsClient returns a new Gitlab Commit service
func NewCommitsClient(cfg common.Config) CommitClient {
	git := common.NewClient(cfg)
	return git.Commits
}

// IsErrorFileNotFound helper function to test for errProjectNotFound error.
func IsErrorFileNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errFileNotFound)
}

// LateInitializeFile fills the empty fields in the file spec with the
// values seen in gitlab.File.
func LateInitializeFile(in *v1alpha1.FileParameters, file *gitlab.File, commit *gitlab.Commit) {
	if file == nil {
		return
	}

	if in.Encoding == nil {
		in.Encoding = clients.StringToPtr("text")
	}

	if in.ExecuteFilemode == nil {
		in.ExecuteFilemode = &file.ExecuteFilemode
	}

	if in.AuthorEmail == nil {
		in.AuthorEmail = &commit.AuthorEmail
	}

	if in.AuthorName == nil {
		in.AuthorName = &commit.AuthorName
	}

	if in.StartBranch == nil {
		in.StartBranch = clients.StringToPtr("")
	}
}

// GenerateFileObservation is used to produce v1alpha1.FileObservation from
// gitlab.File.
func GenerateFileObservation(file *gitlab.File) v1alpha1.FileObservation {
	if file == nil {
		return v1alpha1.FileObservation{}
	}

	o := v1alpha1.FileObservation{
		BlobID:          file.BlobID,
		CommitID:        file.CommitID,
		Content:         file.Content,
		ContentSHA256:   file.SHA256,
		Encoding:        file.Encoding,
		ExecuteFilemode: file.ExecuteFilemode,
		FileName:        file.FileName,
		FilePath:        file.FilePath,
		LastCommitId:    file.LastCommitID,
		Ref:             file.Ref,
		Size:            file.Size,
	}

	return o
}

// GenerateGetFileOptions generates project get options
func GenerateGetFileOptions(p *v1alpha1.FileParameters) *gitlab.GetFileOptions {

	o := &gitlab.GetFileOptions{
		Ref: p.Branch,
	}

	return o
}

// GenerateCreateFileOptions generates project creation options
func GenerateCreateFileOptions(p *v1alpha1.FileParameters) *gitlab.CreateFileOptions {

	o := &gitlab.CreateFileOptions{
		Branch:          p.Branch,
		StartBranch:     p.StartBranch,
		Encoding:        p.Encoding,
		AuthorEmail:     p.AuthorEmail,
		AuthorName:      p.AuthorName,
		Content:         p.Content,
		CommitMessage:   p.CommitMessage,
		ExecuteFilemode: p.ExecuteFilemode,
	}

	return o
}

// GenerateUpdateFileOptions generates project update options
func GenerateUpdateFileOptions(p *v1alpha1.FileParameters) *gitlab.UpdateFileOptions {
	cm := "Update file " + *p.FilePath

	o := &gitlab.UpdateFileOptions{
		Branch:          p.Branch,
		Encoding:        p.Encoding,
		AuthorEmail:     p.AuthorEmail,
		AuthorName:      p.AuthorName,
		Content:         p.Content,
		CommitMessage:   clients.StringToPtr(cm),
		ExecuteFilemode: p.ExecuteFilemode,
	}

	return o
}

// GenerateDeleteFileOptions generates project delete options
func GenerateDeleteFileOptions(p *v1alpha1.FileParameters) *gitlab.DeleteFileOptions {
	cm := "Delete file " + *p.FilePath

	o := &gitlab.DeleteFileOptions{
		Branch:        p.Branch,
		AuthorEmail:   p.AuthorEmail,
		AuthorName:    p.AuthorName,
		CommitMessage: &cm,
	}

	return o
}

// GenerateGetCommitOptions generates project get options
func GenerateGetCommitOptions() *gitlab.GetCommitOptions {
	b := false
	o := &gitlab.GetCommitOptions{
		Stats: &b,
	}

	return o
}

// Calculate SHA256 checksum on file content and return as hexadecimal string
func getFileSHA256(s *string, encoding *string) string {
	var cont string
	if !cmp.Equal(encoding, clients.StringToPtr("text")) {
		cont = base64Decode(*s)
	} else {
		cont = *s
	}
	sum := sha256.Sum256([]byte(cont))
	hex := fmt.Sprintf("%x", sum)
	return hex
}

// Return Base64 decoded string
func base64Decode(enc string) string {
	dec, _ := base64.StdEncoding.DecodeString(enc)
	s := string(dec)
	return s
}

// IsFileUpToDate checks whether there is a change in any of the modifiable fields.
func IsFileUpToDate(p *v1alpha1.FileParameters, g *gitlab.File) bool {
	sha256 := getFileSHA256(p.Content, p.Encoding)

	if !cmp.Equal(sha256, g.SHA256) {
		return false
	}

	if !cmp.Equal(p.Branch, clients.StringToPtr(g.Ref)) {
		return false
	}

	if !clients.IsBoolEqualToBoolPtr(p.ExecuteFilemode, g.ExecuteFilemode) {
		return false
	}
	return true
}
