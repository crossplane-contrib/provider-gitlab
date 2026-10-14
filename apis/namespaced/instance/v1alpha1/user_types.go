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
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UserState is the administrative state the provider keeps a GitLab user in.
// +kubebuilder:validation:Enum=active;blocked
type UserState string

// User states the provider can converge to.
const (
	UserStateActive  UserState = "active"
	UserStateBlocked UserState = "blocked"
)

// UserParameters define the desired state of a GitLab user.
// The ProviderConfig must authenticate with an instance administrator token.
// When Admin Mode is enabled, the token also needs the admin_mode scope.
//
// GitLab API docs:
// https://docs.gitlab.com/api/users/
// +kubebuilder:validation:XValidation:rule="!has(self.passwordSecretRef) || !((has(self.resetPassword) && self.resetPassword) || (has(self.forceRandomPassword) && self.forceRandomPassword))",message="passwordSecretRef is mutually exclusive with resetPassword and forceRandomPassword"
type UserParameters struct {
	// Username is the unique handle of the user.
	// Changing it renames the user's personal namespace.
	// +kubebuilder:validation:MinLength=1
	Username string `json:"username"`

	// Name is the display name of the user.
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Email is the primary email address of the user.
	// Changes are applied without email reconfirmation.
	// +kubebuilder:validation:MinLength=1
	Email string `json:"email"`

	// PasswordSecretRef selects the secret key that holds the initial password.
	// GitLab never returns passwords, so the value is only sent when the user is created.
	// GitLab rejects a new user unless passwordSecretRef, resetPassword or forceRandomPassword is set.
	// Mutually exclusive with resetPassword and forceRandomPassword.
	// +optional
	PasswordSecretRef *v2.LocalSecretKeySelector `json:"passwordSecretRef,omitempty"`

	// ResetPassword sends the user a password reset link when the user is created.
	// +optional
	ResetPassword *bool `json:"resetPassword,omitempty"`

	// ForceRandomPassword sets a random password when the user is created.
	// Use it for users that only sign in through an external identity provider.
	// +optional
	ForceRandomPassword *bool `json:"forceRandomPassword,omitempty"`

	// SkipConfirmation marks the email address as confirmed when the user is created.
	// +optional
	SkipConfirmation *bool `json:"skipConfirmation,omitempty"`

	// Admin grants instance administrator privileges.
	// +optional
	Admin *bool `json:"admin,omitempty"`

	// CanCreateGroup allows the user to create top level groups.
	// +optional
	CanCreateGroup *bool `json:"canCreateGroup,omitempty"`

	// ProjectsLimit is the number of personal projects the user can create.
	// +kubebuilder:validation:Minimum=0
	// +optional
	ProjectsLimit *int64 `json:"projectsLimit,omitempty"`

	// External limits the user to projects they are explicitly granted access to.
	// +optional
	External *bool `json:"external,omitempty"`

	// Note is an administrator note about the user.
	// +optional
	Note *string `json:"note,omitempty"`

	// State is the administrative state of the user.
	// When unset, the provider does not change the state.
	// +optional
	State *UserState `json:"state,omitempty"`
}

// UserObservation represents the observed state of a GitLab user.
type UserObservation struct {
	// ID is the unique identifier of the user.
	ID int64 `json:"id"`
	// Username is the unique handle of the user.
	Username string `json:"username"`
	// Name is the display name of the user.
	Name string `json:"name"`
	// Email is the primary email address of the user.
	Email string `json:"email"`
	// State is the state GitLab reports, for example active, blocked, deactivated or banned.
	State string `json:"state"`
	// Admin reports whether the user is an instance administrator.
	Admin bool `json:"admin"`
	// CanCreateGroup reports whether the user can create top level groups.
	CanCreateGroup bool `json:"canCreateGroup"`
	// ProjectsLimit is the number of personal projects the user can create.
	ProjectsLimit int64 `json:"projectsLimit"`
	// External reports whether the user is an external user.
	External bool `json:"external"`
	// Note is the administrator note about the user.
	Note string `json:"note"`
	// NamespaceID is the ID of the user's personal namespace.
	NamespaceID int64 `json:"namespaceId"`
}

// A UserSpec defines the desired state of a GitLab user.
type UserSpec struct {
	v2.ManagedResourceSpec `json:",inline"`
	ForProvider            UserParameters `json:"forProvider"`
}

// A UserStatus represents the observed state of a GitLab user.
type UserStatus struct {
	v2.ManagedResourceStatus `json:",inline"`
	AtProvider               UserObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A User is a managed resource that represents a GitLab instance user.
// Deleting it deletes the GitLab user. GitLab removes the user's personal
// projects and moves the user's contributions to the ghost user.
// GitLab refuses to delete a user who is the sole owner of a group.
//
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}
type User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserSpec   `json:"spec"`
	Status UserStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UserList contains a list of User items.
type UserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []User `json:"items"`
}
