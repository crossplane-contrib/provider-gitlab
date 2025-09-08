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
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"

	commonv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/shared/common/v1alpha1"
)

// RunnerParameters define the desired state of a group Runner.
// A group Runner is a GitLab Runner that is linked to a specific group
// and can execute CI/CD jobs for projects within that group.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#create-a-runner
type RunnerParameters struct {
	// GroupID is the ID of the group to register the runner to.
	// The runner will be available to execute jobs for projects within this group.
	// +optional
	// +immutable
	GroupID *int `json:"groupId,omitempty"`

	// CommonRunnerParameters contains the common runner configuration
	// parameters shared between group and project runners.
	commonv1alpha1.CommonRunnerParameters `json:",inline"`
}

// RunnerObservation represents the observed state of a group Runner.
// This includes the common runner properties as well as group-specific information.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/runners.html#get-runners-details
type RunnerObservation struct {
	// CommonRunnerObservation contains the common observed fields
	// shared between group and project runners.
	commonv1alpha1.CommonRunnerObservation `json:",inline"`

	// Groups contains the list of groups that this runner is associated with.
	// For group runners, this typically contains the primary group and any parent groups.
	// +optional
	Groups []RunnerGroup `json:"groups"`
}

// RunnerGroup represents a GitLab group associated with a runner.
// This structure matches the group information returned by the GitLab API
// when retrieving runner details.
type RunnerGroup struct {
	// ID is the unique identifier of the group.
	// +optional
	ID int `json:"id"`

	// Name is the name of the group.
	// +optional
	Name string `json:"name"`

	// WebURL is the web URL to access the group in the GitLab UI.
	// +optional
	WebURL string `json:"web_url"`
}

// RunnerSpec defines the desired state of a group Runner.
// This includes the configuration parameters for creating and managing
// a GitLab Runner linked to a specific group.
type RunnerSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       RunnerParameters `json:"forProvider"`
}

// RunnerStatus represents the observed state of a group Runner.
// This includes the current status and properties of the runner as
// reported by the GitLab API.
type RunnerStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          RunnerObservation `json:"atProvider,omitempty"`
}
