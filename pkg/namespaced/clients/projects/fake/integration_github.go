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
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"

	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects"
)

var _ projects.GithubClient = &MockGithubClient{}

// MockGithubClient is a fake implementation of projects.GithubClient.
type MockGithubClient struct {
	MockGetGithubService    func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.GithubService, *gitlab.Response, error)
	MockSetGithubService    func(pid any, opt *gitlab.SetGithubServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GithubService, *gitlab.Response, error)
	MockDeleteGithubService func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// GetGithubService calls the underlying MockGetGithubService method.
func (c *MockGithubClient) GetGithubService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.GithubService, *gitlab.Response, error) {
	return c.MockGetGithubService(pid, options...)
}

// SetGithubService calls the underlying MockSetGithubService method.
func (c *MockGithubClient) SetGithubService(pid any, opt *gitlab.SetGithubServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GithubService, *gitlab.Response, error) {
	return c.MockSetGithubService(pid, opt, options...)
}

// DeleteGithubService calls the underlying MockDeleteGithubService method.
func (c *MockGithubClient) DeleteGithubService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteGithubService(pid, options...)
}
