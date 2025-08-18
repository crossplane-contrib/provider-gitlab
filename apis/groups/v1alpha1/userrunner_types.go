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
	commonv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UserRunnerParameters define the desired state of a group UserRunner.
// A group UserRunner is a GitLab Runner that is linked to a specific group
// and can execute CI/CD jobs for projects within that group.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#create-a-runner
type UserRunnerParameters struct {
	// GroupID is the ID of the group to register the runner to.
	// The runner will be available to execute jobs for projects within this group.
	// +optional
	// +immutable
	GroupID *int `json:"groupId,omitempty"`

	// GroupIDRef is a reference to a Group resource to retrieve its ID.
	// This provides a way to reference a group managed by Crossplane.
	// +optional
	// +immutable
	GroupIDRef *xpv1.Reference `json:"groupIdRef,omitempty"`

	// GroupIDSelector selects a reference to a Group resource to retrieve its ID.
	// This provides a way to dynamically select a group based on labels.
	// +optional
	GroupIDSelector *xpv1.Selector `json:"groupIdSelector,omitempty"`

	// CommonUserRunnerParameters contains the common runner configuration
	// parameters shared between group and project runners.
	commonv1alpha1.CommonUserRunnerParameters `json:",inline"`
}

// UserRunnerObservation represents the observed state of a group UserRunner.
// This includes the common runner properties as well as group-specific information.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/runners.html#get-runners-details
type UserRunnerObservation struct {
	// CommonUserRunnerObservation contains the common observed fields
	// shared between group and project runners.
	commonv1alpha1.CommonUserRunnerObservation `json:",inline"`

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

// UserRunnerSpec defines the desired state of a group UserRunner.
// This includes the configuration parameters for creating and managing
// a GitLab Runner linked to a specific group.
type UserRunnerSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       UserRunnerParameters `json:"forProvider"`
}

// UserRunnerStatus represents the observed state of a group UserRunner.
// This includes the current status and properties of the runner as
// reported by the GitLab API.
type UserRunnerStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          UserRunnerObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A UserRunner is a managed resource that represents a GitLab Runner linked to a group.
// Group runners can execute CI/CD jobs for all projects within the associated group
// and provide a way to share runner resources across multiple projects.
//
// IMPORTANT: You MUST specify either writeConnectionSecretToRef or publishConnectionDetailsTo
// to receive the runner token. Without the token, the runner cannot be registered and is unusable.
// The token is required to configure the actual GitLab Runner agent.
//
// Example usage:
//   spec:
//     writeConnectionSecretToRef:
//       name: my-runner-token
//       namespace: default
//
// When a UserRunner is created, it generates a runner token that must be used
// to register the actual GitLab Runner agent with the GitLab instance.
// The runner token is made available through Kubernetes secrets via connection details.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#create-a-runner
// https://docs.gitlab.com/ee/api/runners.html
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gitlab}
type UserRunner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserRunnerSpec   `json:"spec"`
	Status UserRunnerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UserRunnerList contains a list of group UserRunner resources.
type UserRunnerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UserRunner `json:"items"`
}
