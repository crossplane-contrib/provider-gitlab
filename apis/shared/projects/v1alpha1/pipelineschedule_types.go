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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PipelineScheduleParameters represents a pipeline schedule.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/pipeline_schedules.html
// At least 1 of [ProjectID, ProjectIDRef, ProjectIDSelector] required.
type PipelineScheduleParameters struct {
	// The ID or URL-encoded path of the project owned by the authenticated user.
	// +optional
	// +immutable
	ProjectID *int `json:"projectId,omitempty"`

	// Description is a description of the pipeline schedule.
	// +kubebuilder:example="Nightly build and test"
	// +required
	Description string `json:"description"`

	// Ref is the branch or tag name that is triggered.
	// +kubebuilder:example="main"
	// +required
	Ref string `json:"ref"`

	// Cron is the cron schedule, for example: 0 1 * * *.
	// +kubebuilder:example="0 2 * * *"
	// +required
	Cron string `json:"cron"`

	// CronTimezone is the time zone supported by ActiveSupport::TimeZone,
	// for example: Pacific Time (US & Canada) (default: UTC).
	// +kubebuilder:example="UTC"
	// +optional
	CronTimezone *string `json:"cronTimezone,omitempty"`

	// Active is the activation of pipeline schedule.
	// If false is set, the pipeline schedule is initially deactivated (default: true).
	// +optional
	Active *bool `json:"active,omitempty"`

	// PipelineVariables is a type of environment variable.
	Variables []PipelineVariable `json:"variables,omitempty"`
}

// PipelineScheduleObservation represents observed stated of Gitlab Pipeline Schedule.
// https://docs.gitlab.com/ee/api/pipeline_schedules.htm
type PipelineScheduleObservation struct {
	ID           *int          `json:"id,omitempty"`
	NextRunAt    *metav1.Time  `json:"nextRunAt,omitempty"`
	CreatedAt    *metav1.Time  `json:"createdAt,omitempty"`
	UpdatedAt    *metav1.Time  `json:"updatedAt,omitempty"`
	Owner        *User         `json:"owner,omitempty"`
	LastPipeline *LastPipeline `json:"lastPipeline,omitempty"`
}

// LastPipeline represents the last pipeline ran by schedule
// this will be returned only for individual schedule get operation
type LastPipeline struct {
	ID     int    `json:"id"`
	SHA    string `json:"sha"`
	Ref    string `json:"ref"`
	Status string `json:"status"`
	WebURL string `json:"webUrl"`
}

// PipelineVariable represents a pipeline variable.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/pipelines.html
type PipelineVariable struct {
	// +kubebuilder:example="DEPLOY_ENV"
	Key string `json:"key"`
	// +kubebuilder:example="production"
	Value string `json:"value"`
	// +kubebuilder:example="env_var"
	// +optional
	VariableType *string `json:"variableType,omitempty"`
}
