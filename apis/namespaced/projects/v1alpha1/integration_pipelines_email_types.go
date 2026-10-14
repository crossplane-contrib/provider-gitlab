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

	"github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
)

// IntegrationPipelinesEmailParameters defines the desired state of a GitLab Project Pipeline Status Emails Integration.
type IntegrationPipelinesEmailParameters struct {
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

	// Recipients are the email addresses that receive pipeline status emails.
	// GitLab accepts at most 30 recipients.
	// +required
	// +listType=set
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=30
	// +kubebuilder:validation:items:Pattern=`^[^@\s,<>]+@[^@\s,<>]+$`
	Recipients []string `json:"recipients"`

	// NotifyOnlyBrokenPipelines sends notifications for broken pipelines only.
	// +optional
	NotifyOnlyBrokenPipelines *bool `json:"notifyOnlyBrokenPipelines,omitempty"`

	// BranchesToBeNotified selects the branches to send notifications for.
	// GitLab defaults to default.
	// +kubebuilder:validation:Enum:=all;default;protected;default_and_protected
	// +optional
	BranchesToBeNotified *string `json:"branchesToBeNotified,omitempty"`

	// PipelineEvents enables notifications for pipeline events.
	// +optional
	PipelineEvents *bool `json:"pipelineEvents,omitempty"`
}

// IntegrationPipelinesEmailObservation represents the observed state of a GitLab Project Pipeline Status Emails Integration.
type IntegrationPipelinesEmailObservation struct {
	v1alpha1.CommonIntegrationObservation `json:",inline"`

	// Recipients are the email addresses that receive pipeline status emails.
	Recipients []string `json:"recipients,omitempty"`
	// NotifyOnlyBrokenPipelines indicates whether notifications are sent for broken pipelines only.
	NotifyOnlyBrokenPipelines bool `json:"notifyOnlyBrokenPipelines,omitempty"`
	// BranchesToBeNotified is the set of branches notifications are sent for.
	BranchesToBeNotified string `json:"branchesToBeNotified,omitempty"`
}

// A IntegrationPipelinesEmailSpec defines the desired state of a GitLab Project Pipeline Status Emails Integration.
type IntegrationPipelinesEmailSpec struct {
	v2.ManagedResourceSpec `json:",inline"`
	// ForProvider represents the desired state of the pipeline status emails integration.
	ForProvider IntegrationPipelinesEmailParameters `json:"forProvider"`
}

// A IntegrationPipelinesEmailStatus represents the observed state of a GitLab Project Pipeline Status Emails Integration.
type IntegrationPipelinesEmailStatus struct {
	v2.ManagedResourceStatus `json:",inline"`
	// AtProvider represents the observed state of the pipeline status emails integration.
	AtProvider IntegrationPipelinesEmailObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A IntegrationPipelinesEmail is a managed resource that represents a GitLab Project Pipeline Status Emails Integration.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="PROJECT",type="string",JSONPath=".spec.forProvider.projectId"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}
type IntegrationPipelinesEmail struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              IntegrationPipelinesEmailSpec   `json:"spec"`
	Status            IntegrationPipelinesEmailStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IntegrationPipelinesEmailList contains a list of IntegrationPipelinesEmail items.
type IntegrationPipelinesEmailList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IntegrationPipelinesEmail `json:"items"`
}
