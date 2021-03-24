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

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// DeployKeyParameters define the desired state of a Gitlab deploy key
type DeployKeyParameters struct {
	// ProjectID is the ID of the project to create the deploy key in.
	// +optional
	// +immutable
	ProjectID *int `json:"projectId,omitempty"`

	// ProjectIDRef is a reference to a project to retrieve its projectId
	// +optional
	// +immutable
	ProjectIDRef *xpv1.Reference `json:"projectIdRef,omitempty"`

	// ProjectIDSelector selects reference to a project to retrieve its projectId.
	// +optional
	ProjectIDSelector *xpv1.Selector `json:"projectIdSelector,omitempty"`

	// Title is the new deploy key’s title
	Title *string `json:"title,omitempty"`

	// Key is the new public deploy key. If left empty, a random key will be generated
	// and the private part written to writeConnectionSecretToRef
	// +optional
	Key *string `json:"key,omitempty"`

	// CanPush defines whether the deploy key can push to the project’s repository
	// +optional
	CanPush *bool `json:"canPush,omitempty"`
}

// DeployKeyObservation represents a deploy key.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/deploy_keys.html
type DeployKeyObservation struct {
	// ID of the deploy key at gitlab
	ID int `json:"id,omitempty"`

	// CreatedAt specifies the time the deploy key was created
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`
}

// A DeployKeySpec defines the desired state of a Gitlab DeployKey.
type DeployKeySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       DeployKeyParameters `json:"forProvider"`
}

// A DeployKeyStatus represents the observed state of a Gitlab DeployKey.
type DeployKeyStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          DeployKeyObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A DeployKey is a managed resource that represents a Gitlab deploy key
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gitlab}
type DeployKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeployKeySpec   `json:"spec"`
	Status DeployKeyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DeployKeyList contains a list of DeployKey items
type DeployKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DeployKey `json:"items"`
}

// Bool is a helper routine that allocates a new bool value
// to store v and returns a pointer to it.
func Bool(v bool) *bool {
	p := new(bool)
	*p = v
	return p
}

// Int is a helper routine that allocates a new int32 value
// to store v and returns a pointer to it, but unlike Int32
// its argument value is an int.
func Int(v int) *int {
	p := new(int)
	*p = v
	return p
}

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string {
	p := new(string)
	*p = v
	return p
}

// StringSlice is a helper routine that allocates a new []string value
// to store v and returns a pointer to it.
func StringSlice(v []string) *[]string {
	p := new([]string)
	*p = v
	return p
}
