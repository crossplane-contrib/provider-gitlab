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

package v1alpha1

import (
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HookParameters defines the desired state of a Gitlab Project Hook.
type HookParameters struct {
	// URL is the hook URL.
	URL *string `json:"url"`

	// ConfidentialNoteEvents triggers hook on confidential issues events.
	// +optional
	ConfidentialNoteEvents *bool `json:"confidentialNoteEvents,omitempty"`

	// ProjectID is the ID of the project.
	// +optional
	// +immutable
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="projectId is immutable"
	ProjectID *int64 `json:"projectId,omitempty"`

	// ProjectIDRef is a reference to a project to retrieve its projectId
	// +optional
	// +immutable
	ProjectIDRef *v2.NamespacedReference `json:"projectIdRef,omitempty"`

	// ProjectIDSelector selects reference to a project to retrieve its projectId.
	// +optional
	ProjectIDSelector *v2.NamespacedSelector `json:"projectIdSelector,omitempty"`

	// PushEvents triggers hook on push events.
	// +optional
	PushEvents *bool `json:"pushEvents,omitempty"`

	// PushEventsBranchFilter triggers hook on push events for matching branches only.
	// +optional
	PushEventsBranchFilter *string `json:"pushEventsBranch_filter,omitempty"`

	// IssuesEvents triggers hook on issues events.
	// +optional
	IssuesEvents *bool `json:"issuesEvents,omitempty"`

	// ConfidentialIssuesEvents triggers hook on confidential issues events.
	// +optional
	ConfidentialIssuesEvents *bool `json:"confidentialIssuesEvents,omitempty"`

	// MergeRequestsEvents triggers hook on merge requests events.
	// +optional
	MergeRequestsEvents *bool `json:"mergeRequestsEvents,omitempty"`

	// TagPushEvents triggers hook on tag push events.
	// +optional
	TagPushEvents *bool `json:"tagPushEvents,omitempty"`

	// NoteEvents triggers hook on note events.
	// +optional
	NoteEvents *bool `json:"noteEvents,omitempty"`

	// JobEvents triggers hook on job events.
	// +optional
	JobEvents *bool `json:"jobEvents,omitempty"`

	// PipelineEvents triggers hook on pipeline events.
	// +optional
	PipelineEvents *bool `json:"pipelineEvents,omitempty"`

	// WikiPageEvents triggers hook on wiki events.
	// +optional
	WikiPageEvents *bool `json:"wikiPageEvents,omitempty"`

	// EnableSSLVerification enables SSL verification when triggering the hook.
	// +optional
	EnableSSLVerification *bool `json:"enableSslVerification,omitempty"`

	// Token is the secret token to validate received payloads.
	Token *Token `json:"token"`
}

type Token struct {
	SecretRef *v2.LocalSecretKeySelector `json:"secretRef"`
}

// HookObservation represents a project hook.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/projects.html#list-project-hooks
type HookObservation struct {
	// ID of the project hook at gitlab
	ID int64 `json:"id,omitempty"`

	// CreatedAt specifies the time the project hook was created
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// TokenHash is a SHA-256 digest of the secret token last pushed to GitLab.
	// GitLab never returns the token, so this fingerprint is used to detect
	// rotation of the referenced secret and trigger an update. It is never the
	// raw token.
	// +optional
	TokenHash string `json:"tokenHash,omitempty"`
}

// A HookSpec defines the desired state of a Gitlab Project Hook.
type HookSpec struct {
	v2.ManagedResourceSpec `json:",inline"`
	ForProvider            HookParameters `json:"forProvider"`
}

// A HookStatus represents the observed state of a Gitlab Project Hook.
type HookStatus struct {
	v2.ManagedResourceStatus `json:",inline"`
	AtProvider               HookObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Hook is a managed resource that represents a Gitlab Project Hook
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}
type Hook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HookSpec   `json:"spec"`
	Status HookStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HookList contains a list of Project Hook items
type HookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Hook `json:"items"`
}
