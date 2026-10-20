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

package v1alpha1

import (
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RepositoryFileEncoding describes how the content supplied to a
// RepositoryFile is encoded. It never affects what is stored in the
// repository: the provider always decodes the supplied content and commits
// the resulting bytes.
// +kubebuilder:validation:Enum=text;base64
type RepositoryFileEncoding string

const (
	// RepositoryFileEncodingText means the content is used as-is.
	RepositoryFileEncodingText RepositoryFileEncoding = "text"

	// RepositoryFileEncodingBase64 means the content is standard base64
	// (RFC 4648) and is decoded before it is committed. Use it for binary
	// files supplied inline.
	RepositoryFileEncodingBase64 RepositoryFileEncoding = "base64"
)

// RepositoryFileParameters define the desired state of a file in a GitLab
// repository. A file is identified by projectId, branch and filePath, all of
// which are immutable.
//
// Which actions Crossplane takes on the file is controlled by
// spec.managementPolicies, not by fields in here. For example
// ["Observe", "Create"] commits the file once and then never overwrites or
// deletes it, and ["Observe"] tracks a file without ever committing.
//
// https://docs.gitlab.com/api/repository_files/
// +kubebuilder:validation:XValidation:rule="!has(oldSelf.projectId) || has(self.projectId)",message="projectId cannot be removed once set"
type RepositoryFileParameters struct {
	// ProjectID is the ID or URL-encoded path of the project.
	// +optional
	// +immutable
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="projectId is immutable"
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1.Project
	// +crossplane:generate:reference:refFieldName=ProjectIDRef
	// +crossplane:generate:reference:selectorFieldName=ProjectIDSelector
	ProjectID *string `json:"projectId,omitempty"`

	// ProjectIDRef is a reference to a project to retrieve its projectId.
	// +optional
	// +immutable
	ProjectIDRef *v2.NamespacedReference `json:"projectIdRef,omitempty"`

	// ProjectIDSelector selects reference to a project to retrieve its projectId.
	// +optional
	// +immutable
	ProjectIDSelector *v2.NamespacedSelector `json:"projectIdSelector,omitempty"`

	// Branch is the branch the file lives on. It must already exist unless
	// the repository is empty; the provider does not create branches.
	// +immutable
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="branch is immutable"
	Branch string `json:"branch"`

	// FilePath is the path of the file in the repository, e.g. "docs/README.md".
	// +immutable
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="filePath is immutable"
	FilePath string `json:"filePath"`

	// Content is the file content, encoded as described by Encoding.
	// Mutually exclusive with ContentSecretRef. One of the two is required
	// unless managementPolicies allow neither Create nor Update.
	// +optional
	Content *string `json:"content,omitempty"`

	// ContentSecretRef references a Secret key holding the file content,
	// encoded as described by Encoding. Mutually exclusive with Content.
	// Note that the content is committed to the repository, so it is only
	// as secret as the repository itself.
	// +optional
	ContentSecretRef *v2.LocalSecretKeySelector `json:"contentSecretRef,omitempty"`

	// Encoding describes how Content or the referenced Secret value is
	// encoded.
	// +optional
	// +kubebuilder:default=text
	Encoding *RepositoryFileEncoding `json:"encoding,omitempty"`

	// ExecuteFilemode enables the executable flag on the file.
	// +optional
	ExecuteFilemode *bool `json:"executeFilemode,omitempty"`

	// CreateCommitMessage is the commit message used when creating the file.
	// Defaults to "crossplane: create <filePath>".
	// +optional
	CreateCommitMessage *string `json:"createCommitMessage,omitempty"`

	// UpdateCommitMessage is the commit message used when updating the file.
	// Defaults to "crossplane: update <filePath>".
	// +optional
	UpdateCommitMessage *string `json:"updateCommitMessage,omitempty"`

	// DeleteCommitMessage is the commit message used when deleting the file.
	// Defaults to "crossplane: delete <filePath>".
	// +optional
	DeleteCommitMessage *string `json:"deleteCommitMessage,omitempty"`

	// AuthorEmail is the commit author email.
	// +optional
	AuthorEmail *string `json:"authorEmail,omitempty"`

	// AuthorName is the commit author name.
	// +optional
	AuthorName *string `json:"authorName,omitempty"`
}

// RepositoryFileObservation represents the observed state of a GitLab Repository File.
type RepositoryFileObservation struct {
	// FilePath is the observed file path.
	FilePath string `json:"filePath,omitempty"`

	// BlobID is the blob SHA of the file.
	BlobID string `json:"blobId,omitempty"`

	// LastCommitID is the last commit SHA that modified this file, as seen by
	// the most recent observation. It is sent with updates and deletes so
	// GitLab rejects them if the file changed after that observation. This
	// only guards a single reconcile: the next one observes the new commit
	// and, if updates are allowed, enforces the desired content again.
	LastCommitID string `json:"lastCommitId,omitempty"`

	// SHA256 is the SHA-256 hash of the file content.
	SHA256 string `json:"sha256,omitempty"`

	// Size is the file size in bytes.
	Size int64 `json:"size,omitempty"`
}

// A RepositoryFileSpec defines the desired state of a GitLab Repository File.
type RepositoryFileSpec struct {
	v2.ManagedResourceSpec `json:",inline"`
	ForProvider            RepositoryFileParameters `json:"forProvider"`
}

// A RepositoryFileStatus represents the observed state of a GitLab Repository File.
type RepositoryFileStatus struct {
	v2.ManagedResourceStatus `json:",inline"`
	AtProvider               RepositoryFileObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A RepositoryFile is a managed resource that represents a file in a GitLab repository.
// +kubebuilder:printcolumn:name="PROJECT",type="string",JSONPath=".spec.forProvider.projectId"
// +kubebuilder:printcolumn:name="FILE",type="string",JSONPath=".spec.forProvider.filePath"
// +kubebuilder:printcolumn:name="BRANCH",type="string",JSONPath=".spec.forProvider.branch"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}
type RepositoryFile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:XValidation:rule="!(has(self.forProvider.content) && has(self.forProvider.contentSecretRef))",message="spec.forProvider.content and spec.forProvider.contentSecretRef are mutually exclusive"
	// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.content) || has(self.forProvider.contentSecretRef)",message="spec.forProvider.content or spec.forProvider.contentSecretRef is required when managementPolicies allow Create or Update"
	Spec   RepositoryFileSpec   `json:"spec"`
	Status RepositoryFileStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RepositoryFileList contains a list of RepositoryFile items.
type RepositoryFileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RepositoryFile `json:"items"`
}
