/*
Copyright 2022 The Crossplane Authors.

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

// CommonRunnerParameters contains common configuration parameters for user runners
// that are shared between group and project runner types.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#create-a-runner
type CommonRunnerParameters struct {
	// Description is a human-readable description of the runner.
	// +optional
	Description *string `json:"description,omitempty"`

	// Paused indicates whether the runner is paused and will not accept new jobs.
	// When true, the runner will not pick up new jobs but can complete running jobs.
	// +optional
	Paused *bool `json:"paused,omitempty"`

	// Locked indicates whether the runner is locked to the current project/group.
	// When true, the runner cannot be enabled for other projects/groups.
	// +optional
	Locked *bool `json:"locked,omitempty"`

	// RunUntagged indicates whether the runner should run jobs without tags.
	// When true, the runner will accept jobs that have no tags specified.
	// +optional
	RunUntagged *bool `json:"runUntagged,omitempty"`

	// TagList contains the list of tags associated with the runner.
	// Jobs with matching tags will be eligible to run on this runner.
	// +optional
	TagList *[]string `json:"tagList,omitempty"`

	// AccessLevel specifies the access level of the runner.
	// Valid values: "not_protected", "ref_protected"
	// - "not_protected": Runner can execute jobs for all branches/tags
	// - "ref_protected": Runner can only execute jobs for protected branches/tags
	// +optional
	// +kubebuilder:validation:Enum=not_protected;ref_protected
	AccessLevel *string `json:"accessLevel,omitempty"`

	// MaximumTimeout is the maximum time (in seconds) that a job can run on this runner.
	// If a job exceeds this time limit, it will be terminated.
	// +optional
	MaximumTimeout *int `json:"maximumTimeout,omitempty"`

	// MaintenanceNote is a note that can be set when the runner is in maintenance mode.
	// This is displayed in the GitLab UI when the runner is offline or paused.
	// +optional
	MaintenanceNote *string `json:"maintenanceNote,omitempty"`
}

// CommonRunnerObservation represents the observed state of a user runner
// that is common between group and project runners.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/runners.html#get-runners-details
type CommonRunnerObservation struct {
	// Paused indicates whether the runner is currently paused.
	// Paused runners do not accept new jobs but can complete running jobs.
	// +optional
	Paused bool `json:"paused"`

	// Description is the human-readable description of the runner.
	// +optional
	Description string `json:"description"`

	// ID is the unique identifier of the runner assigned by GitLab.
	ID int `json:"id"`

	// IsShared indicates whether this is a shared runner available to all projects.
	// User runners are never shared, so this is typically false.
	// +optional
	IsShared bool `json:"isShared"`

	// RunnerType indicates the type of runner.
	// For user runners, this is typically "project_type" or "group_type".
	// +optional
	RunnerType string `json:"runnerType"`

	// ContactedAt is the timestamp when the runner last contacted GitLab.
	// This is used to determine if a runner is online or offline.
	// +optional
	ContactedAt *metav1.Time `json:"contactedAt"`

	// MaintenanceNote contains any maintenance note set for the runner.
	// This is displayed in the GitLab UI when the runner is offline or paused.
	// +optional
	MaintenanceNote string `json:"maintenanceNote"`

	// Name is the name of the runner as reported by the runner agent.
	// This is typically set automatically by the GitLab Runner software.
	// +optional
	Name string `json:"name"`

	// Online indicates whether the runner is currently online and available.
	// This is determined by the last contact time and heartbeat interval.
	// +optional
	Online bool `json:"online"`

	// Status represents the current status of the runner.
	// Valid values: "online", "offline", "stale", "never_contacted"
	// +optional
	Status string `json:"status"`

	// TagList contains the list of tags associated with the runner.
	// Jobs with matching tags will be eligible to run on this runner.
	// +optional
	TagList []string `json:"tagList"`

	// RunUntagged indicates whether the runner accepts jobs without tags.
	// +optional
	RunUntagged bool `json:"runUntagged"`

	// Locked indicates whether the runner is locked to its current scope.
	// +optional
	Locked bool `json:"locked"`

	// AccessLevel specifies the access level of the runner.
	// Valid values: "not_protected", "ref_protected"
	// +optional
	AccessLevel string `json:"accessLevel"`

	// MaximumTimeout is the maximum time (in seconds) that jobs can run on this runner.
	// +optional
	MaximumTimeout int `json:"maximumTimeout"`

	// TokenExpiresAt is the timestamp when the runner token expires.
	// Currently GitLab does not provide an expiration for user runner tokens,
	// so this is typically nil for user runners.
	// This field is included for consistency and future extensibility.
	// +optional
	TokenExpiresAt *metav1.Time `json:"tokenExpiresAt"`
}
