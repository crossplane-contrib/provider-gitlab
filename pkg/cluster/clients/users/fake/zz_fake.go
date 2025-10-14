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

package fake

import (
	gitlab "gitlab.com/gitlab-org/api/client-go"

	users "github.com/crossplane-contrib/provider-gitlab/pkg/cluster/clients/users"
)

var _ users.UserClient = &MockClient{}

type MockClient struct {
	users.UserClient

	MockCreateUserRunner func(opts *gitlab.CreateUserRunnerOptions, options ...gitlab.RequestOptionFunc) (*gitlab.UserRunner, *gitlab.Response, error)
}

func (m *MockClient) CreateUserRunner(opts *gitlab.CreateUserRunnerOptions, options ...gitlab.RequestOptionFunc) (*gitlab.UserRunner, *gitlab.Response, error) {
	return m.MockCreateUserRunner(opts)
}
