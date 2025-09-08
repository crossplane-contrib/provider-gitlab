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

// DeployKeyParameters define desired state of Gitlab Deploy Key.
// https://docs.gitlab.com/ee/api/deploy_keys.html
// At least 1 of [ProjectID, ProjectIDRef, ProjectIDSelector] required.
type DeployKeyParameters struct {
	// The ID or URL-encoded path of the project owned by the authenticated user.
	// +optional
	// +immutable
	ProjectID *int `json:"projectId,omitempty"`

	// New Deploy Key’s title.
	// This property is required.
	// +kubebuilder:example="CI/CD Deploy Key"
	Title string `json:"title"`

	// Can Deploy Key push to the project’s repository.
	// +optional
	CanPush *bool `json:"canPush,omitempty"`

	// Expiration date for the Deploy Key. Does not expire if no value is provided.
	// Expected in ISO 8601 format (2019-03-15T08:00:00Z).
	// +optional
	ExpiresAt *metav1.Time `json:"expiresAt,omitempty"`
}

// DeployKeyObservation represents observed stated of Deploy Key.
// https://docs.gitlab.com/ee/api/deploy_keys.html
type DeployKeyObservation struct {
	ID        *int         `json:"id,omitempty"`
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`
}
