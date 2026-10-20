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
	"reflect"

	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
)

// IntegrationTelegramParameters defines the desired state of a GitLab Project Telegram Integration.
type IntegrationTelegramParameters struct {
	// ProjectID is the ID of the project.
	// +optional
	// +immutable
	ProjectID *int64 `json:"projectId,omitempty"`

	// ProjectIDRef is a reference to a project to retrieve its projectId.
	// +optional
	// +immutable
	ProjectIDRef *v2.NamespacedReference `json:"projectIdRef,omitempty"`

	// ProjectIDSelector selects reference to a project to retrieve its projectId.
	// +optional
	ProjectIDSelector *v2.NamespacedSelector `json:"projectIdSelector,omitempty"`

	// TokenSecretRef is a reference to a secret that contains the Telegram bot token.
	// WARNING: This field is NOT reconciled as the GitLab API does not return it as it is a write-only field.
	// +required
	TokenSecretRef v2.LocalSecretKeySelector `json:"tokenSecretRef"`

	// Room is the unique identifier of the target chat or the username of the target channel (in the format @channelusername).
	// +required
	Room string `json:"room"`

	// NotifyOnlyBrokenPipelines sends notifications for broken pipelines only.
	// +optional
	NotifyOnlyBrokenPipelines *bool `json:"notifyOnlyBrokenPipelines,omitempty"`

	// BranchesToBeNotified selects the branches to send notifications for. GitLab defaults to default.
	// +kubebuilder:validation:Enum:=all;default;protected;default_and_protected
	// +optional
	BranchesToBeNotified *string `json:"branchesToBeNotified,omitempty"`

	// PushEvents enables notifications for push events.
	// +optional
	PushEvents *bool `json:"pushEvents,omitempty"`

	// IssuesEvents enables notifications for issue events.
	// +optional
	IssuesEvents *bool `json:"issuesEvents,omitempty"`

	// ConfidentialIssuesEvents enables notifications for confidential issue events.
	// +optional
	ConfidentialIssuesEvents *bool `json:"confidentialIssuesEvents,omitempty"`

	// MergeRequestsEvents enables notifications for merge request events.
	// +optional
	MergeRequestsEvents *bool `json:"mergeRequestsEvents,omitempty"`

	// TagPushEvents enables notifications for tag push events.
	// +optional
	TagPushEvents *bool `json:"tagPushEvents,omitempty"`

	// NoteEvents enables notifications for note events.
	// +optional
	NoteEvents *bool `json:"noteEvents,omitempty"`

	// ConfidentialNoteEvents enables notifications for confidential note events.
	// +optional
	ConfidentialNoteEvents *bool `json:"confidentialNoteEvents,omitempty"`

	// PipelineEvents enables notifications for pipeline events.
	// +optional
	PipelineEvents *bool `json:"pipelineEvents,omitempty"`

	// WikiPageEvents enables notifications for wiki page events.
	// +optional
	WikiPageEvents *bool `json:"wikiPageEvents,omitempty"`
}

// IntegrationTelegramObservation represents the observed state of a GitLab Project Telegram Integration.
type IntegrationTelegramObservation struct {
	v1alpha1.CommonIntegrationObservation `json:",inline"`

	// Room is the target chat or channel.
	Room string `json:"room,omitempty"`
	// NotifyOnlyBrokenPipelines indicates whether notifications are sent for broken pipelines only.
	NotifyOnlyBrokenPipelines bool `json:"notifyOnlyBrokenPipelines"`
	// BranchesToBeNotified are the branches notifications are sent for.
	BranchesToBeNotified string `json:"branchesToBeNotified,omitempty"`
}

// A IntegrationTelegramSpec defines the desired state of a GitLab Project Telegram Integration.
type IntegrationTelegramSpec struct {
	v2.ManagedResourceSpec `json:",inline"`
	// ForProvider represents the desired state of the Telegram integration.
	ForProvider IntegrationTelegramParameters `json:"forProvider"`
}

// A IntegrationTelegramStatus represents the observed state of a GitLab Project Telegram Integration.
type IntegrationTelegramStatus struct {
	v2.ManagedResourceStatus `json:",inline"`
	// AtProvider represents the observed state of the Telegram integration.
	AtProvider IntegrationTelegramObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A IntegrationTelegram is a managed resource that represents a GitLab Project Telegram Integration.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="PROJECT",type="string",JSONPath=".spec.forProvider.projectId"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}
type IntegrationTelegram struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              IntegrationTelegramSpec   `json:"spec"`
	Status            IntegrationTelegramStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IntegrationTelegramList contains a list of IntegrationTelegram items.
type IntegrationTelegramList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IntegrationTelegram `json:"items"`
}

// IntegrationTelegram type metadata
var (
	IntegrationTelegramKind             = reflect.TypeOf(IntegrationTelegram{}).Name()
	IntegrationTelegramGroupKind        = schema.GroupKind{Group: Group, Kind: IntegrationTelegramKind}.String()
	IntegrationTelegramKindAPIVersion   = IntegrationTelegramKind + "." + SchemeGroupVersion.String()
	IntegrationTelegramGroupVersionKind = SchemeGroupVersion.WithKind(IntegrationTelegramKind)
)
