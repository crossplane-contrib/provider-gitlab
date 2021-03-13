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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

// A ProjectHookParameters defines the desired state of a Gitlab ProjectHook.
type ProjectHookParameters struct {
	// URL is the hook URL.
	URL *string `json:"url"`

	// ConfidentialNoteEvents triggers hook on confidential issues events.
	// +optional
	ConfidentialNoteEvents *bool `json:"confidentialNoteEvents,omitempty"`

	// ProjectID is the ID of the project.
	// +immutable
	ProjectID *int `json:"projectId"`

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
	// +optional
	Token *string `json:"token,omitempty"`
}

// ProjectHookObservation represents a project hook.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/projects.html#list-project-hooks
type ProjectHookObservation struct {
	// ID of the project hook at gitlab
	ID int `json:"id,omitempty"`

	// CreatedAt specifies the time the project hook was created
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`
}

// A ProjectHookSpec defines the desired state of a Gitlab ProjectHook.
type ProjectHookSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  ProjectHookParameters `json:"forProvider"`
}

// A ProjectHookStatus represents the observed state of a Gitlab ProjectHook.
type ProjectHookStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     ProjectHookObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A ProjectHook is a managed resource that represents a Gitlab ProjectHook
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gitlab}
type ProjectHook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectHookSpec   `json:"spec"`
	Status ProjectHookStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProjectHookList contains a list of ProjectHook items
type ProjectHookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProjectHook `json:"items"`
}
