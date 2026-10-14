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
// that are shared between instance, group and project service account types.
//
// GitLab API docs:
// https://docs.gitlab.com/api/service_accounts/
type CommonServiceAccountParameters struct {
	// name represents the display name of the service account.
	// +optional
	Name *string `json:"name,omitempty"`
	// username represents the user @ name of the service account.
	// +optional
	Username *string `json:"username,omitempty"`
	// email represents the email of the service account.
	// name, username and email are reconciled in place (rename-in-place); a
	// change to any of them is pushed to GitLab on update.
	// +optional
	Email *string `json:"email,omitempty"`
	// avatar is the base64 encoded image GitLab shows for the service account.
	// GitLab accepts PNG, JPEG, GIF, BMP, WEBP and ICO images up to 200 KB.
	// Setting it requires an administrator token, with the admin_mode scope
	// when Admin Mode is on, also for group and project service accounts. The
	// image is uploaded again whenever the avatar in GitLab differs, including
	// changes made outside of Crossplane. When unset, the avatar is not managed.
	// +optional
	Avatar []byte `json:"avatar,omitempty"`
}

// CommonServiceAccountObservation represents the observed state of a user service account
// that is common between instance and group service accounts.
//
// GitLab API docs:
// https://docs.gitlab.com/api/service_accounts/
type CommonServiceAccountObservation struct {
	// ID is the unique identifier of the service account.
	ID int64 `json:"id"`
	// Name represents the display name of the service account.
	Name string `json:"name"`
	// Username represents the user @ name of the service account.
	Username string `json:"username"`
	// Email represents the email of the service account.
	Email string `json:"email"`
	// Admin represents whether the service account has admin privileges.
	Admin bool `json:"admin"`
	// AvatarURL is the URL of the avatar GitLab shows for the service account.
	AvatarURL string `json:"avatarUrl,omitempty"`
}
