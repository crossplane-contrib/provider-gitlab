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

	MockGetHook    func(pid interface{}, hook int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error)
	MockAddHook    func(pid interface{}, opt *gitlab.AddProjectHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error)
	MockEditHook   func(pid interface{}, hook int, opt *gitlab.EditProjectHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error)
	MockDeleteHook func(pid interface{}, hook int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

	MockGetMember    func(pid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error)
	MockAddMember    func(pid interface{}, opt *gitlab.AddProjectMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error)
	MockEditMember   func(pid interface{}, user int, opt *gitlab.EditProjectMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error)
	MockDeleteMember func(pid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

	MockCreateDeployToken     func(pid interface{}, opt *gitlab.CreateProjectDeployTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error)
	MockDeleteDeployToken     func(pid interface{}, deployToken int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
	MockGetProjectDeployToken func(pid interface{}, deployToken int, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error)

	MockGetVariable    func(pid interface{}, key string, opt *gitlab.GetProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error)
	MockCreateVariable func(pid interface{}, opt *gitlab.CreateProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error)
	MockUpdateVariable func(pid interface{}, key string, opt *gitlab.UpdateProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error)
	MockListVariables  func(pid interface{}, opt *gitlab.ListProjectVariablesOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectVariable, *gitlab.Response, error)
	MockRemoveVariable func(pid interface{}, key string, opt *gitlab.RemoveProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

	MockGetProjectAccessToken    func(pid interface{}, id int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectAccessToken, *gitlab.Response, error)
	MockCreateProjectAccessToken func(pid interface{}, opt *gitlab.CreateProjectAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectAccessToken, *gitlab.Response, error)
	MockRevokeProjectAccessToken func(pid interface{}, id int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

	MockAddDeployKey    func(pid interface{}, opt *gitlab.AddDeployKeyOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectDeployKey, *gitlab.Response, error)
	MockDeleteDeployKey func(pid interface{}, deployKey int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
	MockUpdateDeployKey func(pid interface{}, deployKey int, opt *gitlab.UpdateDeployKeyOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectDeployKey, *gitlab.Response, error)
	MockGetDeployKey    func(pid interface{}, deployKey int, options ...*gitlab.RequestOptionFunc) (*gitlab.ProjectDeployKey, *gitlab.Response, error)

	MockGetPipelineSchedule            func(pid interface{}, schedule int, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error)
	MockCreatePipelineSchedule         func(pid interface{}, opt *gitlab.CreatePipelineScheduleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error)
	MockEditPipelineSchedule           func(pid interface{}, schedule int, opt *gitlab.EditPipelineScheduleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error)
	MockDeletePipelineSchedule         func(pid interface{}, schedule int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
	MockCreatePipelineScheduleVariable func(pid interface{}, schedule int, opt *gitlab.CreatePipelineScheduleVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineVariable, *gitlab.Response, error)
	MockEditPipelineScheduleVariable   func(pid interface{}, schedule int, key string, opt *gitlab.EditPipelineScheduleVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineVariable, *gitlab.Response, error)
	MockDeletePipelineScheduleVariable func(pid interface{}, schedule int, key string, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineVariable, *gitlab.Response, error)
}

// GetPipelineSchedule calls the underlying MockGetPipelineSchedule method.
func (c *MockClient) GetPipelineSchedule(pid interface{}, schedule int, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error) {
	return c.MockGetPipelineSchedule(pid, schedule, options...)
}

// CreatePipelineSchedule calls the underlying MockCreatePipelineSchedule method.
func (c *MockClient) CreatePipelineSchedule(pid interface{}, opt *gitlab.CreatePipelineScheduleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error) {
	return c.MockCreatePipelineSchedule(pid, opt)
}

// EditPipelineSchedule calls the underlying MockEditPipelineSchedule method.
func (c *MockClient) EditPipelineSchedule(pid interface{}, schedule int, opt *gitlab.EditPipelineScheduleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error) {
	return c.MockEditPipelineSchedule(pid, schedule, opt)
}

// DeletePipelineSchedule calls the underlying MockDeletePipelineSchedule method.
func (c *MockClient) DeletePipelineSchedule(pid interface{}, schedule int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeletePipelineSchedule(pid, schedule)
}

// CreatePipelineScheduleVariable calls the underlying MockCreatePipelineScheduleVariable method.
func (c *MockClient) CreatePipelineScheduleVariable(pid interface{}, schedule int, opt *gitlab.CreatePipelineScheduleVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineVariable, *gitlab.Response, error) {
	return c.MockCreatePipelineScheduleVariable(pid, schedule, opt)
}

// EditPipelineScheduleVariable calls the underlying MockEditPipelineScheduleVariable method.
func (c *MockClient) EditPipelineScheduleVariable(pid interface{}, schedule int, key string, opt *gitlab.EditPipelineScheduleVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineVariable, *gitlab.Response, error) {
	return c.MockEditPipelineScheduleVariable(pid, schedule, key, opt)
}

// DeletePipelineScheduleVariable calls the underlying MockDeletePipelineScheduleVariable method.
func (c *MockClient) DeletePipelineScheduleVariable(pid interface{}, schedule int, key string, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineVariable, *gitlab.Response, error) {
	return c.MockDeletePipelineScheduleVariable(pid, schedule, key)
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
	return c.MockGetHook(pid, hook)
}

// AddProjectHook calls the underlying MockAddHook method.
// AddProjectHook calls the underlying MockAddHook method.
func (c *MockClient) AddProjectHook(pid interface{}, opt *gitlab.AddProjectHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error) {
	return c.MockAddHook(pid, opt)
}

// EditProjectHook calls the underlying MockEditProjectHook method.
func (c *MockClient) EditProjectHook(pid interface{}, hook int, opt *gitlab.EditProjectHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error) {
	return c.MockEditHook(pid, hook, opt)
}

// DeleteProjectHook calls the underlying MockDeleteProjectHook method.
func (c *MockClient) DeleteProjectHook(pid interface{}, hook int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteHook(pid, hook)
}

// GetProjectMember calls the underlying MockGetMember method.
// GetProjectMember calls the underlying MockGetMember method.
func (c *MockClient) GetProjectMember(pid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
	return c.MockGetMember(pid, user)
}

// AddProjectMember calls the underlying MockAddMember method.
// AddProjectMember calls the underlying MockAddMember method.
func (c *MockClient) AddProjectMember(pid interface{}, opt *gitlab.AddProjectMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
	return c.MockAddMember(pid, opt)
}

// EditProjectMember calls the underlying MockEditMember method.
// EditProjectMember calls the underlying MockEditMember method.
func (c *MockClient) EditProjectMember(pid interface{}, user int, opt *gitlab.EditProjectMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
	return c.MockEditMember(pid, user, opt)
}

// DeleteProjectMember calls the underlying MockDeleteMember method.
// DeleteProjectMember calls the underlying MockDeleteMember method.
func (c *MockClient) DeleteProjectMember(pid interface{}, user int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteMember(pid, user)
}

// CreateProjectDeployToken calls the underlying MockCreateProjectDeployToken method.
func (c *MockClient) CreateProjectDeployToken(pid interface{}, opt *gitlab.CreateProjectDeployTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error) {
	return c.MockCreateDeployToken(pid, opt)
}

// DeleteProjectDeployToken calls the underlying MockDeleteProjectDeployToken method.
func (c *MockClient) DeleteProjectDeployToken(pid interface{}, deployToken int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteDeployToken(pid, deployToken)
}

// GetProjectDeployToken calls the underlying MockGetProjectDeployToken method.
func (c *MockClient) GetProjectDeployToken(pid interface{}, deployToken int, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error) {
	return c.MockGetProjectDeployToken(pid, deployToken)
}

// GetVariable calls the underlying MockGetProjectVariable
func (c *MockClient) GetVariable(pid interface{}, key string, opt *gitlab.GetProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error) {
	return c.MockGetVariable(pid, key, opt)
}

// CreateVariable calls the underlying MockCreateProjectVariable
func (c *MockClient) CreateVariable(pid interface{}, opt *gitlab.CreateProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error) {
	return c.MockCreateVariable(pid, opt)
}

// UpdateVariable calls the underlying MockUpdateProjectVariable
func (c *MockClient) UpdateVariable(pid interface{}, key string, opt *gitlab.UpdateProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error) {
	return c.MockUpdateVariable(pid, key, opt)
}

// RemoveVariable calls the underlying MockRemoveProjectVariable
func (c *MockClient) RemoveVariable(pid interface{}, key string, opt *gitlab.RemoveProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockRemoveVariable(pid, key, opt)
}

// ListVariables calls the underlying MockListVariables
func (c *MockClient) ListVariables(pid interface{}, opt *gitlab.ListProjectVariablesOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectVariable, *gitlab.Response, error) {
	return c.MockListVariables(pid, opt)
}

// GetDeployKey calls the underlying MockGetDeployKey
func (c *MockClient) GetDeployKey(pid interface{}, deployKey int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectDeployKey, *gitlab.Response, error) {
	return c.MockGetDeployKey(pid, deployKey)
}

// AddDeployKey calls the underlying MockAddDeployKey
func (c *MockClient) AddDeployKey(pid interface{}, opt *gitlab.AddDeployKeyOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectDeployKey, *gitlab.Response, error) {
	return c.MockAddDeployKey(pid, opt)
}

// DeleteDeployKey calls the underlying MockDeleteDeployKey
func (c *MockClient) DeleteDeployKey(pid interface{}, deployKey int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteDeployKey(pid, deployKey)
}

// UpdateDeployKey cals the underlying MockUpdateDeployKey
func (c *MockClient) UpdateDeployKey(pid interface{}, deployKey int, opt *gitlab.UpdateDeployKeyOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectDeployKey, *gitlab.Response, error) {
	return c.MockUpdateDeployKey(pid, deployKey, opt)
}

// GetProjectAccessToken calls the underlying MockGetProjectAccessToken method.
func (c *MockClient) GetProjectAccessToken(pid interface{}, id int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectAccessToken, *gitlab.Response, error) {
	return c.MockGetProjectAccessToken(pid, id)
}

// CreateProjectAccessToken calls the underlying MockCreateProjectAccessToken method.
func (c *MockClient) CreateProjectAccessToken(pid interface{}, opt *gitlab.CreateProjectAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectAccessToken, *gitlab.Response, error) {
	return c.MockCreateProjectAccessToken(pid, opt)
}

// RevokeProjectAccessToken calls the underlying MockRevokeProjectAccessToken method.
func (c *MockClient) RevokeProjectAccessToken(pid interface{}, id int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockRevokeProjectAccessToken(pid, id)
}
