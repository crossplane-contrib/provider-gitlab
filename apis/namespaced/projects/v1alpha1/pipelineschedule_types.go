/*
Copyright 2023 The Crossplane Authors.

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
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	sharedProjectsV1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/shared/projects/v1alpha1"
)

// PipelineScheduleParameters represents a pipeline schedule.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/pipeline_schedules.html
// At least 1 of [ProjectID, ProjectIDRef, ProjectIDSelector] required.
type PipelineScheduleParameters struct {
	sharedProjectsV1alpha1.PipelineScheduleParameters `json:",inline"`

	// ProjectIDRef is a reference to a project to retrieve its ProjectID.
	// +optional
	// +immutable
	ProjectIDRef *xpv1.NamespacedReference `json:"projectIdRef,omitempty"`

	// ProjectIDSelector selects reference to a project to retrieve its ProjectID.
	// +optional
	// +immutable
	ProjectIDSelector *xpv1.NamespacedSelector `json:"projectIdSelector,omitempty"`
}

// PipelineScheduleSpec defines desired state of Gitlab Pipeline Schedule.
type PipelineScheduleSpec struct {
	xpv2.ManagedResourceSpec `json:","`
	ForProvider              PipelineScheduleParameters `json:"forProvider"`
}

// PipelineScheduleStatus represents observed state of Gitlab Pipeline Schedule.
type PipelineScheduleStatus struct {
	xpv1.ResourceStatus `json:","`
	AtProvider          sharedProjectsV1alpha1.PipelineScheduleObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A PipelineSchedule is a managed resource that represents a Gitlab Pipeline Schedule.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}
type PipelineSchedule struct {
	metav1.TypeMeta   `json:","`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelineScheduleSpec   `json:"spec"`
	Status PipelineScheduleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PipelineScheduleList contains a list of Pipeline Schedule items.
type PipelineScheduleList struct {
	metav1.TypeMeta `json:","`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []PipelineSchedule `json:"items"`
}
