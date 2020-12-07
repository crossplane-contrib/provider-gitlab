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

package fake

import (
	"github.com/xanzy/go-gitlab"

	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
)

var _ projects.Client = &MockClient{}

// MockClient is a fake implementation of projects.Client.
type MockClient struct {
	projects.Client

	MockGetProject    func(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error)
	MockCreateProject func(opt *gitlab.CreateProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error)
	MockEditProject   func(pid interface{}, opt *gitlab.EditProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error)
}

// // GetProject calls the underlying MockGetProject method.
func (c *MockClient) GetProject(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
	return c.MockGetProject(pid, opt)
}

// CreateProject calls the underlying MockCreateProject method
func (c *MockClient) CreateProject(opt *gitlab.CreateProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
	return c.MockCreateProject(opt)
}

// EditProject calls the underlying MockEditProject method
func (c *MockClient) EditProject(pid interface{}, opt *gitlab.EditProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
	return c.MockEditProject(pid, opt)
}
