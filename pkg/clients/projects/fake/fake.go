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
	MockDeleteProject func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

	MockGetProjectHook    func(pid interface{}, hook int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error)
	MockAddProjectHook    func(pid interface{}, opt *gitlab.AddProjectHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error)
	MockEditProjectHook   func(pid interface{}, hook int, opt *gitlab.EditProjectHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error)
	MockDeleteProjectHook func(pid interface{}, hook int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

	MockGetProjectMember    func(pid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error)
	MockAddProjectMember    func(pid interface{}, opt *gitlab.AddProjectMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error)
	MockEditProjectMember   func(pid interface{}, user int, opt *gitlab.EditProjectMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error)
	MockDeleteProjectMember func(pid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

	MockListProjectDeployTokens  func(pid interface{}, opt *gitlab.ListProjectDeployTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.DeployToken, *gitlab.Response, error)
	MockCreateProjectDeployToken func(pid interface{}, opt *gitlab.CreateProjectDeployTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error)
	MockDeleteProjectDeployToken func(pid interface{}, deployToken int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// GetProject calls the underlying MockGetProject method.
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

// DeleteProject calls the underlying MockDeleteProject method
func (c *MockClient) DeleteProject(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteProject(pid)
}

// GetProjectHook calls the underlying MockGetProjectHook method.
func (c *MockClient) GetProjectHook(pid interface{}, hook int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error) {
	return c.MockGetProjectHook(pid, hook)
}

// AddProjectHook calls the underlying MockAddProjectHook method.
func (c *MockClient) AddProjectHook(pid interface{}, opt *gitlab.AddProjectHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error) {
	return c.MockAddProjectHook(pid, opt)
}

// EditProjectHook calls the underlying MockEditProjectHook method.
func (c *MockClient) EditProjectHook(pid interface{}, hook int, opt *gitlab.EditProjectHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error) {
	return c.MockEditProjectHook(pid, hook, opt)
}

// DeleteProjectHook calls the underlying MockDeleteProjectHook method.
func (c *MockClient) DeleteProjectHook(pid interface{}, hook int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteProjectHook(pid, hook)
}

// GetProjectMember calls the underlying MockGetProjectMember method.
func (c *MockClient) GetProjectMember(pid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
	return c.MockGetProjectMember(pid, user)
}

// AddProjectMember calls the underlying MockAddProjectMember method.
func (c *MockClient) AddProjectMember(pid interface{}, opt *gitlab.AddProjectMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
	return c.MockAddProjectMember(pid, opt)
}

// EditProjectMember calls the underlying MockEditProjectMember method.
func (c *MockClient) EditProjectMember(pid interface{}, user int, opt *gitlab.EditProjectMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
	return c.MockEditProjectMember(pid, user, opt)
}

// DeleteProjectMember calls the underlying MockDeleteProjectMember method.
func (c *MockClient) DeleteProjectMember(pid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteProjectMember(pid, user)
}

// ListProjectDeployTokens calls the underlying MockListProjectDeployTokens method.
func (c *MockClient) ListProjectDeployTokens(pid interface{}, opt *gitlab.ListProjectDeployTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.DeployToken, *gitlab.Response, error) {
	return c.MockListProjectDeployTokens(pid, opt)
}

// CreateProjectDeployToken calls the underlying MockCreateProjectDeployToken method.
func (c *MockClient) CreateProjectDeployToken(pid interface{}, opt *gitlab.CreateProjectDeployTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error) {
	return c.MockCreateProjectDeployToken(pid, opt)
}

// DeleteProjectDeployToken calls the underlying MockDeleteProjectDeployToken method.
func (c *MockClient) DeleteProjectDeployToken(pid interface{}, deployToken int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteProjectDeployToken(pid, deployToken)
}
