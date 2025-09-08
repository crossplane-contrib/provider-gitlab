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
	commonv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/shared/common/v1alpha1"
)

// RunnerParameters define the desired state of a project Runner.
// A project Runner is a GitLab Runner that is linked to a specific project
// and can execute CI/CD jobs exclusively for that project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#create-a-runner
type RunnerParameters struct {
	// ProjectID is the ID of the project to register the runner to.
	// The runner will be available to execute jobs exclusively for this project.
	// +optional
	// +immutable
	ProjectID *int `json:"projectId,omitempty"`

	// CommonRunnerParameters contains the common runner configuration
	// parameters shared between group and project runners.
	commonv1alpha1.CommonRunnerParameters `json:",inline"`
}

// RunnerObservation represents the observed state of a project Runner.
// This includes the common runner properties as well as project-specific information.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/runners.html#get-runners-details
type RunnerObservation struct {
	// CommonRunnerObservation contains the common observed fields
	// shared between group and project runners.
	commonv1alpha1.CommonRunnerObservation `json:",inline"`

	// Projects contains the list of projects that this runner is associated with.
	// For project runners, this typically contains only the primary project.
	Projects []RunnerProject `json:"projects"`
}

// RunnerProject represents a GitLab project associated with a runner.
// This structure matches the project information returned by the GitLab API
// when retrieving runner details.
type RunnerProject struct {
	// ID is the unique identifier of the project.
	// +optional
	ID int `json:"id"`

	// Name is the name of the project.
	// +optional
	Name string `json:"name"`

	// NameWithNamespace is the full name of the project including its namespace.
	// This follows the format "namespace/project-name".
	// +optional
	NameWithNamespace string `json:"name_with_namespace"`

	// Path is the URL path segment for the project.
	// +optional
	Path string `json:"path"`

	// PathWithNamespace is the full path of the project including its namespace.
	// This follows the format "namespace/project-path" and is used in URLs.
	// +optional
	PathWithNamespace string `json:"path_with_namespace"`
}
