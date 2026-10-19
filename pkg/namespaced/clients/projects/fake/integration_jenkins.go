/*
Copyright 2026 The Crossplane Authors.

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

var _ projects.JenkinsClient = &MockJenkinsClient{}

// MockJenkinsClient is a fake implementation of projects.JenkinsClient.
type MockJenkinsClient struct {
	MockGetJenkinsCIService    func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.JenkinsCIService, *gitlab.Response, error)
	MockSetJenkinsCIService    func(pid any, opt *gitlab.SetJenkinsCIServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.JenkinsCIService, *gitlab.Response, error)
	MockDeleteJenkinsCIService func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// GetJenkinsCIService calls the underlying MockGetJenkinsCIService method.
func (c *MockJenkinsClient) GetJenkinsCIService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.JenkinsCIService, *gitlab.Response, error) {
	return c.MockGetJenkinsCIService(pid, options...)
}

// SetJenkinsCIService calls the underlying MockSetJenkinsCIService method.
func (c *MockJenkinsClient) SetJenkinsCIService(pid any, opt *gitlab.SetJenkinsCIServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.JenkinsCIService, *gitlab.Response, error) {
	return c.MockSetJenkinsCIService(pid, opt, options...)
}

// DeleteJenkinsCIService calls the underlying MockDeleteJenkinsCIService method.
func (c *MockJenkinsClient) DeleteJenkinsCIService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteJenkinsCIService(pid, options...)
}
