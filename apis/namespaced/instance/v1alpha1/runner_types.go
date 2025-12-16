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
	// +cluster-scope:delete=1
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
)

// RunnerParameters define the desired state of an instance Runner.
// A instance Runner is a GitLab Runner that is linked to a specific instance
// and can execute CI/CD jobs exclusively for that instance.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#create-a-runner
type RunnerParameters struct {
	// CommonRunnerParameters contains the common runner configuration
	// parameters shared between group and instance runners.
	commonv1alpha1.CommonRunnerParameters `json:",inline"`
}

// RunnerObservation represents the observed state of an instance Runner.
// This includes the common runner properties as well as instance-specific information.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/runners.html#get-runners-details
type RunnerObservation struct {
	// CommonRunnerObservation contains the common observed fields
	// shared between group and instance runners.
	commonv1alpha1.CommonRunnerObservation `json:",inline"`
}

// RunnerSpec defines the desired state of a instance Runner.
// This includes the configuration parameters for creating and managing
// a GitLab Runner linked to a specific instance.
type RunnerSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`
	ForProvider              RunnerParameters `json:"forProvider"`
}

// RunnerStatus represents the observed state of a instance Runner.
// This includes the current status and properties of the runner as
// reported by the GitLab API.
type RunnerStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          RunnerObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Runner is a managed resource that represents a GitLab Runner linked to a instance.
// Instance runners can execute CI/CD jobs exclusively for the associated instance
// and provide dedicated runner resources for a single instance.
//
// IMPORTANT: You MUST specify either writeConnectionSecretToRef or publishConnectionDetailsTo
// to receive the runner token. Without the token, the runner cannot be registered and is unusable.
// The token is required to configure the actual GitLab Runner agent.
//
// Example usage:
//
//	spec:
//	  writeConnectionSecretToRef:
//	    name: my-runner-token
//	    namespace: default
//
// When a Runner is created, it generates a runner token that must be used
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
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}
type Runner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RunnerSpec   `json:"spec"`
	Status RunnerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RunnerList contains a list of project Runner resources.
type RunnerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Runner `json:"items"`
}
