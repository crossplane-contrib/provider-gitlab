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

	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
)

var _ projects.Client = &MockClient{}

// MockClient is a fake implementation of projects.Client.
type MockClient struct {
	projects.Client

	MockGetProject    func(pid any, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error)
	MockCreateProject func(opt *gitlab.CreateProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error)
	MockEditProject   func(pid any, opt *gitlab.EditProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error)
	MockDeleteProject func(pid any, opt *gitlab.DeleteProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

	MockGetHook    func(pid any, hook int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error)
	MockAddHook    func(pid any, opt *gitlab.AddProjectHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error)
	MockEditHook   func(pid any, hook int, opt *gitlab.EditProjectHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error)
	MockDeleteHook func(pid any, hook int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

	MockGetMember    func(pid any, user int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error)
	MockAddMember    func(pid any, opt *gitlab.AddProjectMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error)
	MockEditMember   func(pid any, user int, opt *gitlab.EditProjectMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error)
	MockDeleteMember func(pid any, user int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

	MockCreateDeployToken     func(pid any, opt *gitlab.CreateProjectDeployTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error)
	MockDeleteDeployToken     func(pid any, deployToken int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
	MockGetProjectDeployToken func(pid any, deployToken int, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error)

	MockGetVariable    func(pid any, key string, opt *gitlab.GetProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error)
	MockCreateVariable func(pid any, opt *gitlab.CreateProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error)
	MockUpdateVariable func(pid any, key string, opt *gitlab.UpdateProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error)
	MockListVariables  func(pid any, opt *gitlab.ListProjectVariablesOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectVariable, *gitlab.Response, error)
	MockRemoveVariable func(pid any, key string, opt *gitlab.RemoveProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

	MockGetProjectAccessToken    func(pid any, id int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectAccessToken, *gitlab.Response, error)
	MockCreateProjectAccessToken func(pid any, opt *gitlab.CreateProjectAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectAccessToken, *gitlab.Response, error)
	MockRevokeProjectAccessToken func(pid any, id int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

	MockAddDeployKey    func(pid any, opt *gitlab.AddDeployKeyOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectDeployKey, *gitlab.Response, error)
	MockDeleteDeployKey func(pid any, deployKey int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
	MockUpdateDeployKey func(pid any, deployKey int, opt *gitlab.UpdateDeployKeyOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectDeployKey, *gitlab.Response, error)
	MockGetDeployKey    func(pid any, deployKey int, options ...*gitlab.RequestOptionFunc) (*gitlab.ProjectDeployKey, *gitlab.Response, error)

	MockGetPipelineSchedule            func(pid any, schedule int, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error)
	MockCreatePipelineSchedule         func(pid any, opt *gitlab.CreatePipelineScheduleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error)
	MockEditPipelineSchedule           func(pid any, schedule int, opt *gitlab.EditPipelineScheduleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error)
	MockDeletePipelineSchedule         func(pid any, schedule int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
	MockCreatePipelineScheduleVariable func(pid any, schedule int, opt *gitlab.CreatePipelineScheduleVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineVariable, *gitlab.Response, error)
	MockEditPipelineScheduleVariable   func(pid any, schedule int, key string, opt *gitlab.EditPipelineScheduleVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineVariable, *gitlab.Response, error)
	MockDeletePipelineScheduleVariable func(pid any, schedule int, key string, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineVariable, *gitlab.Response, error)

	MockListUsers func(opt *gitlab.ListUsersOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.User, *gitlab.Response, error)

	MockGetProjectPushRules func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error)
	MockEditProjectPushRule func(pid any, opt *gitlab.EditProjectPushRuleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error)

	MockGetProjectApprovalRule    func(pid any, ruleID int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectApprovalRule, *gitlab.Response, error)
	MockCreateProjectApprovalRule func(pid any, opt *gitlab.CreateProjectLevelRuleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectApprovalRule, *gitlab.Response, error)
	MockUpdateProjectApprovalRule func(pid any, approvalRule int, opt *gitlab.UpdateProjectLevelRuleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectApprovalRule, *gitlab.Response, error)
	MockDeleteProjectApprovalRule func(pid any, approvalRule int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// GetPipelineSchedule calls the underlying MockGetPipelineSchedule method.
func (c *MockClient) GetPipelineSchedule(pid any, schedule int, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error) {
	return c.MockGetPipelineSchedule(pid, schedule, options...)
}

// CreatePipelineSchedule calls the underlying MockCreatePipelineSchedule method.
func (c *MockClient) CreatePipelineSchedule(pid any, opt *gitlab.CreatePipelineScheduleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error) {
	return c.MockCreatePipelineSchedule(pid, opt)
}

// EditPipelineSchedule calls the underlying MockEditPipelineSchedule method.
func (c *MockClient) EditPipelineSchedule(pid any, schedule int, opt *gitlab.EditPipelineScheduleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error) {
	return c.MockEditPipelineSchedule(pid, schedule, opt)
}

// DeletePipelineSchedule calls the underlying MockDeletePipelineSchedule method.
func (c *MockClient) DeletePipelineSchedule(pid any, schedule int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeletePipelineSchedule(pid, schedule)
}

// CreatePipelineScheduleVariable calls the underlying MockCreatePipelineScheduleVariable method.
func (c *MockClient) CreatePipelineScheduleVariable(pid any, schedule int, opt *gitlab.CreatePipelineScheduleVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineVariable, *gitlab.Response, error) {
	return c.MockCreatePipelineScheduleVariable(pid, schedule, opt)
}

// EditPipelineScheduleVariable calls the underlying MockEditPipelineScheduleVariable method.
func (c *MockClient) EditPipelineScheduleVariable(pid any, schedule int, key string, opt *gitlab.EditPipelineScheduleVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineVariable, *gitlab.Response, error) {
	return c.MockEditPipelineScheduleVariable(pid, schedule, key, opt)
}

// DeletePipelineScheduleVariable calls the underlying MockDeletePipelineScheduleVariable method.
func (c *MockClient) DeletePipelineScheduleVariable(pid any, schedule int, key string, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineVariable, *gitlab.Response, error) {
	return c.MockDeletePipelineScheduleVariable(pid, schedule, key)
}

// GetProject calls the underlying MockGetProject method.
func (c *MockClient) GetProject(pid any, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
	return c.MockGetProject(pid, opt)
}

// CreateProject calls the underlying MockCreateProject method
func (c *MockClient) CreateProject(opt *gitlab.CreateProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
	return c.MockCreateProject(opt)
}

// EditProject calls the underlying MockEditProject method
func (c *MockClient) EditProject(pid any, opt *gitlab.EditProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
	return c.MockEditProject(pid, opt)
}

// DeleteProject calls the underlying MockDeleteProject method
func (c *MockClient) DeleteProject(pid any, opt *gitlab.DeleteProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteProject(pid, opt)
}

// GetProjectHook calls the underlying MockGetProjectHook method.
func (c *MockClient) GetProjectHook(pid any, hook int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error) {
	return c.MockGetHook(pid, hook)
}

// AddProjectHook calls the underlying MockAddHook method.
// AddProjectHook calls the underlying MockAddHook method.
func (c *MockClient) AddProjectHook(pid any, opt *gitlab.AddProjectHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error) {
	return c.MockAddHook(pid, opt)
}

// EditProjectHook calls the underlying MockEditProjectHook method.
func (c *MockClient) EditProjectHook(pid any, hook int, opt *gitlab.EditProjectHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error) {
	return c.MockEditHook(pid, hook, opt)
}

// DeleteProjectHook calls the underlying MockDeleteProjectHook method.
func (c *MockClient) DeleteProjectHook(pid any, hook int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteHook(pid, hook)
}

// GetProjectMember calls the underlying MockGetMember method.
// GetProjectMember calls the underlying MockGetMember method.
func (c *MockClient) GetProjectMember(pid any, user int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
	return c.MockGetMember(pid, user)
}

// AddProjectMember calls the underlying MockAddMember method.
// AddProjectMember calls the underlying MockAddMember method.
func (c *MockClient) AddProjectMember(pid any, opt *gitlab.AddProjectMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
	return c.MockAddMember(pid, opt)
}

// EditProjectMember calls the underlying MockEditMember method.
// EditProjectMember calls the underlying MockEditMember method.
func (c *MockClient) EditProjectMember(pid any, user int, opt *gitlab.EditProjectMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectMember, *gitlab.Response, error) {
	return c.MockEditMember(pid, user, opt)
}

// DeleteProjectMember calls the underlying MockDeleteMember method.
// DeleteProjectMember calls the underlying MockDeleteMember method.
func (c *MockClient) DeleteProjectMember(pid any, user int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteMember(pid, user)
}

// CreateProjectDeployToken calls the underlying MockCreateProjectDeployToken method.
func (c *MockClient) CreateProjectDeployToken(pid any, opt *gitlab.CreateProjectDeployTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error) {
	return c.MockCreateDeployToken(pid, opt)
}

// DeleteProjectDeployToken calls the underlying MockDeleteProjectDeployToken method.
func (c *MockClient) DeleteProjectDeployToken(pid any, deployToken int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteDeployToken(pid, deployToken)
}

// GetProjectDeployToken calls the underlying MockGetProjectDeployToken method.
func (c *MockClient) GetProjectDeployToken(pid any, deployToken int, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error) {
	return c.MockGetProjectDeployToken(pid, deployToken)
}

// GetVariable calls the underlying MockGetProjectVariable
func (c *MockClient) GetVariable(pid any, key string, opt *gitlab.GetProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error) {
	return c.MockGetVariable(pid, key, opt)
}

// CreateVariable calls the underlying MockCreateProjectVariable
func (c *MockClient) CreateVariable(pid any, opt *gitlab.CreateProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error) {
	return c.MockCreateVariable(pid, opt)
}

// UpdateVariable calls the underlying MockUpdateProjectVariable
func (c *MockClient) UpdateVariable(pid any, key string, opt *gitlab.UpdateProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectVariable, *gitlab.Response, error) {
	return c.MockUpdateVariable(pid, key, opt)
}

// RemoveVariable calls the underlying MockRemoveProjectVariable
func (c *MockClient) RemoveVariable(pid any, key string, opt *gitlab.RemoveProjectVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockRemoveVariable(pid, key, opt)
}

// ListVariables calls the underlying MockListVariables
func (c *MockClient) ListVariables(pid any, opt *gitlab.ListProjectVariablesOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectVariable, *gitlab.Response, error) {
	return c.MockListVariables(pid, opt)
}

// GetDeployKey calls the underlying MockGetDeployKey
func (c *MockClient) GetDeployKey(pid any, deployKey int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectDeployKey, *gitlab.Response, error) {
	return c.MockGetDeployKey(pid, deployKey)
}

// AddDeployKey calls the underlying MockAddDeployKey
func (c *MockClient) AddDeployKey(pid any, opt *gitlab.AddDeployKeyOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectDeployKey, *gitlab.Response, error) {
	return c.MockAddDeployKey(pid, opt)
}

// DeleteDeployKey calls the underlying MockDeleteDeployKey
func (c *MockClient) DeleteDeployKey(pid any, deployKey int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteDeployKey(pid, deployKey)
}

// UpdateDeployKey cals the underlying MockUpdateDeployKey
func (c *MockClient) UpdateDeployKey(pid any, deployKey int, opt *gitlab.UpdateDeployKeyOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectDeployKey, *gitlab.Response, error) {
	return c.MockUpdateDeployKey(pid, deployKey, opt)
}

// GetProjectAccessToken calls the underlying MockGetProjectAccessToken method.
func (c *MockClient) GetProjectAccessToken(pid any, id int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectAccessToken, *gitlab.Response, error) {
	return c.MockGetProjectAccessToken(pid, id)
}

// CreateProjectAccessToken calls the underlying MockCreateProjectAccessToken method.
func (c *MockClient) CreateProjectAccessToken(pid any, opt *gitlab.CreateProjectAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectAccessToken, *gitlab.Response, error) {
	return c.MockCreateProjectAccessToken(pid, opt)
}

// RevokeProjectAccessToken calls the underlying MockRevokeProjectAccessToken method.
func (c *MockClient) RevokeProjectAccessToken(pid any, id int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockRevokeProjectAccessToken(pid, id)
}

// ListUsers calls the underlying MockListUsers method.
func (c *MockClient) ListUsers(opt *gitlab.ListUsersOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.User, *gitlab.Response, error) {
	return c.MockListUsers(opt)
}

// GetProjectPushRules calls the underlying MockGetProjectPushRules method.
func (c *MockClient) GetProjectPushRules(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
	return c.MockGetProjectPushRules(pid, options...)
}

// EditProjectPushRule calls the underlying MockEditProjectPushRule method.
func (c *MockClient) EditProjectPushRule(pid any, opt *gitlab.EditProjectPushRuleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectPushRules, *gitlab.Response, error) {
	return c.MockEditProjectPushRule(pid, opt, options...)
}

func (c *MockClient) GetProjectApprovalRule(pid any, ruleID int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectApprovalRule, *gitlab.Response, error) {
	return c.MockGetProjectApprovalRule(pid, ruleID, options...)
}

func (c *MockClient) CreateProjectApprovalRule(pid any, opt *gitlab.CreateProjectLevelRuleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectApprovalRule, *gitlab.Response, error) {
	return c.MockCreateProjectApprovalRule(pid, opt, options...)
}

func (c *MockClient) UpdateProjectApprovalRule(pid any, approvalRule int, opt *gitlab.UpdateProjectLevelRuleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectApprovalRule, *gitlab.Response, error) {
	return c.MockUpdateProjectApprovalRule(pid, approvalRule, opt, options...)
}

func (c *MockClient) DeleteProjectApprovalRule(pid any, approvalRule int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteProjectApprovalRule(pid, approvalRule, options...)
}
