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

// CommonServiceAccountParameters contains common configuration parameters for user service accounts
// that are shared between instance and group service account types.
//
// GitLab API docs:
// https://docs.gitlab.com/api/service_accounts/
type CommonServiceAccountParameters struct {
	// name represents the display name of the service account.
	// +optional
	Name string `json:"name"`
	// username represents the user @ name of the service account.
	// +optional
	Username string `json:"username"`
	// email represents the email of the service account.
	// +optional
	Email string `json:"email"`
}

// CommonServiceAccountObservation represents the observed state of a user service account
// that is common between group and project service accounts.
//
// GitLab API docs:
// https://docs.gitlab.com/api/service_accounts/
type CommonServiceAccountObservation struct {
	// ID is the unique identifier of the service account.
	ID int `json:"id,omitempty"`
	// Name represents the display name of the service account.
	Name string `json:"name,omitempty"`
	// Username represents the user @ name of the service account.
	Username string `json:"username,omitempty"`
}
