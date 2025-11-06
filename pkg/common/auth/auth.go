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

package gitlabauth

const (
	// BasicAuth is gitlab's BasicAuth method of authentification that needs a username and a password
	BasicAuth AuthType = "BasicAuth"

	// JobToken is gitlab's JobToken method of authentification
	JobToken AuthType = "JobToken"

	// OAuthToken is gitlab's OAuthToken method of authentification
	OAuthToken AuthType = "OAuthToken"

	// PersonalAccessToken is gitlab's PersonalAccessToken method of authentification.
	PersonalAccessToken AuthType = "PersonalAccessToken"
)

// AuthType represents an authentication type within GitLab.
type AuthType string
