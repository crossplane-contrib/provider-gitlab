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

	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/groups"
)

var _ groups.Client = &MockClient{}

// MockClient is a fake implementation of groups.Client.
type MockClient struct {
	groups.Client

	MockGetGroup              func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error)
	MockCreateGroup           func(opt *gitlab.CreateGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error)
	MockUpdateGroup           func(pid interface{}, opt *gitlab.UpdateGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error)
	MockDeleteGroup           func(pid interface{}, opt *gitlab.DeleteGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
	MockShareGroupWithGroup   func(gid interface{}, opt *gitlab.ShareGroupWithGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error)
	MockUnshareGroupFromGroup func(gid interface{}, groupID int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

	MockGetMember    func(gid interface{}, user int64, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error)
	MockAddMember    func(gid interface{}, opt *gitlab.AddGroupMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error)
	MockEditMember   func(gid interface{}, user int64, opt *gitlab.EditGroupMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error)
	MockRemoveMember func(gid interface{}, user int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

	MockGetGroupDeployToken    func(gid interface{}, deployToken int64, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error)
	MockCreateGroupDeployToken func(gid interface{}, opt *gitlab.CreateGroupDeployTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error)
	MockDeleteGroupDeployToken func(gid interface{}, deployToken int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

	MockGetGroupAccessToken    func(gid interface{}, accessToken int64, options ...gitlab.RequestOptionFunc) (*gitlab.GroupAccessToken, *gitlab.Response, error)
	MockCreateGroupAccessToken func(gid interface{}, opt *gitlab.CreateGroupAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupAccessToken, *gitlab.Response, error)
	MockRevokeGroupAccessToken func(gid interface{}, accessToken int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

	MockListGroupLDAPLinks                func(pid interface{}, options ...gitlab.RequestOptionFunc) ([]*gitlab.LDAPGroupLink, *gitlab.Response, error)
	MockAddGroupLDAPLink                  func(pid interface{}, opt *gitlab.AddGroupLDAPLinkOptions, options ...gitlab.RequestOptionFunc) (*gitlab.LDAPGroupLink, *gitlab.Response, error)
	MockDeleteGroupLDAPLinkWithCNOrFilter func(pid interface{}, opts *gitlab.DeleteGroupLDAPLinkWithCNOrFilterOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

	MockGetGroupSAMLLink    func(pid interface{}, samlGroupName string, options ...gitlab.RequestOptionFunc) (*gitlab.SAMLGroupLink, *gitlab.Response, error)
	MockAddGroupSAMLLink    func(pid interface{}, opt *gitlab.AddGroupSAMLLinkOptions, options ...gitlab.RequestOptionFunc) (*gitlab.SAMLGroupLink, *gitlab.Response, error)
	MockDeleteGroupSAMLLink func(pid interface{}, samlGroupName string, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

	MockListGroupVariables  func(gid interface{}, opt *gitlab.ListGroupVariablesOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.GroupVariable, *gitlab.Response, error)
	MockGetGroupVariable    func(gid interface{}, key string, opt *gitlab.GetGroupVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error)
	MockCreateGroupVariable func(gid interface{}, opt *gitlab.CreateGroupVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error)
	MockUpdateGroupVariable func(gid interface{}, key string, opt *gitlab.UpdateGroupVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error)
	MockRemoveGroupVariable func(gid interface{}, key string, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

	MockListUsers func(opt *gitlab.ListUsersOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.User, *gitlab.Response, error)
}

// GetGroup calls the underlying MockGetGroup method.
func (c *MockClient) GetGroup(pid interface{}, opt *gitlab.GetGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
	return c.MockGetGroup(pid)
}

// CreateGroup calls the underlying MockCreateGroup method
func (c *MockClient) CreateGroup(opt *gitlab.CreateGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
	return c.MockCreateGroup(opt)
}

// UpdateGroup calls the underlying MockUpdateGroup method
func (c *MockClient) UpdateGroup(pid interface{}, opt *gitlab.UpdateGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
	return c.MockUpdateGroup(pid, opt)
}

// DeleteGroup calls the underlying MockDeleteGroup method
func (c *MockClient) DeleteGroup(pid interface{}, opt *gitlab.DeleteGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteGroup(pid, opt)
}

// ShareGroupWithGroup calls the underlying MockShareGroupWithGroup method
func (c *MockClient) ShareGroupWithGroup(gid interface{}, opt *gitlab.ShareGroupWithGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
	return c.MockShareGroupWithGroup(gid, opt, options...)
}

// UnshareGroupFromGroup calls the underlying MockUnshareGroupFromGroup method
func (c *MockClient) UnshareGroupFromGroup(gid interface{}, groupID int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockUnshareGroupFromGroup(gid, groupID, options...)
}

// GetGroupMember calls the underlying MockGetMember method.
func (c *MockClient) GetGroupMember(gid interface{}, user int64, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
	return c.MockGetMember(gid, user)
}

// AddGroupMember calls the underlying MockAddMember method.
func (c *MockClient) AddGroupMember(gid interface{}, opt *gitlab.AddGroupMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
	return c.MockAddMember(gid, opt)
}

// EditGroupMember calls the underlying MockEditMember method.
func (c *MockClient) EditGroupMember(gid interface{}, user int64, opt *gitlab.EditGroupMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
	return c.MockEditMember(gid, user, opt)
}

// RemoveGroupMember calls the underlying MockRemoveMember method.
func (c *MockClient) RemoveGroupMember(gid interface{}, user int64, opt *gitlab.RemoveGroupMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockRemoveMember(gid, user)
}

// GetGroupDeployToken calls the underlying MockGetGroupDeployToken method.
func (c *MockClient) GetGroupDeployToken(gid interface{}, deployToken int64, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error) {
	return c.MockGetGroupDeployToken(gid, deployToken)
}

// CreateGroupDeployToken calls the underlying MockCreateGroupDeployToken method.
func (c *MockClient) CreateGroupDeployToken(gid interface{}, opt *gitlab.CreateGroupDeployTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployToken, *gitlab.Response, error) {
	return c.MockCreateGroupDeployToken(gid, opt)
}

// DeleteGroupDeployToken calls the underlying MockDeleteGroupDeployToken method.
func (c *MockClient) DeleteGroupDeployToken(gid interface{}, deployToken int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteGroupDeployToken(gid, deployToken)
}

// GetGroupAccessToken calls the underlying MockGetGroupDeployToken method.
func (c *MockClient) GetGroupAccessToken(gid interface{}, deployToken int64, options ...gitlab.RequestOptionFunc) (*gitlab.GroupAccessToken, *gitlab.Response, error) {
	return c.MockGetGroupAccessToken(gid, deployToken)
}

// CreateGroupAccessToken calls the underlying MockCreateGroupDeployToken method.
func (c *MockClient) CreateGroupAccessToken(gid interface{}, opt *gitlab.CreateGroupAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupAccessToken, *gitlab.Response, error) {
	return c.MockCreateGroupAccessToken(gid, opt)
}

// RevokeGroupAccessToken calls the underlying MockDeleteGroupDeployToken method.
func (c *MockClient) RevokeGroupAccessToken(gid interface{}, deployToken int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockRevokeGroupAccessToken(gid, deployToken)
}

// ListVariables calls the underlying MockListGroupVariables method.
func (c *MockClient) ListVariables(gid interface{}, opt *gitlab.ListGroupVariablesOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.GroupVariable, *gitlab.Response, error) {
	return c.MockListGroupVariables(gid, opt)
}

// GetVariable calls the underlying MockGetGrouptVariable method.
func (c *MockClient) GetVariable(gid interface{}, key string, opt *gitlab.GetGroupVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error) {
	return c.MockGetGroupVariable(gid, key, opt)
}

// CreateVariable calls the underlying MockCreateGroupVariable method.
func (c *MockClient) CreateVariable(gid interface{}, opt *gitlab.CreateGroupVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error) {
	return c.MockCreateGroupVariable(gid, opt)
}

// UpdateVariable calls the underlying MockUpdateGroupVariable method.
func (c *MockClient) UpdateVariable(gid interface{}, key string, opt *gitlab.UpdateGroupVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error) {
	return c.MockUpdateGroupVariable(gid, key, opt)
}

// RemoveVariable calls the underlying MockRemoveGroupVariable method.
func (c *MockClient) RemoveVariable(gid interface{}, key string, opt *gitlab.RemoveGroupVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockRemoveGroupVariable(gid, key)
}

// ListUsers calls the underlying MockListUsers method.
func (c *MockClient) ListUsers(opt *gitlab.ListUsersOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.User, *gitlab.Response, error) {
	return c.MockListUsers(opt)
}

// ListGroupLDAPLinks calls the underlying MockListGroupLDAPLink method.
func (c *MockClient) ListGroupLDAPLinks(pid interface{}, options ...gitlab.RequestOptionFunc) ([]*gitlab.LDAPGroupLink, *gitlab.Response, error) {
	return c.MockListGroupLDAPLinks(pid)
}

// AddGroupLDAPLink call the underlying MockAddGroupLDAPLink method.
func (c *MockClient) AddGroupLDAPLink(pid interface{}, opt *gitlab.AddGroupLDAPLinkOptions, options ...gitlab.RequestOptionFunc) (*gitlab.LDAPGroupLink, *gitlab.Response, error) {
	return c.MockAddGroupLDAPLink(pid, opt)
}

// DeleteGroupLDAPLink calls the underlying MockDeleteGroupLDAPLink method.
func (c *MockClient) DeleteGroupLDAPLinkWithCNOrFilter(pid interface{}, opts *gitlab.DeleteGroupLDAPLinkWithCNOrFilterOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteGroupLDAPLinkWithCNOrFilter(pid, opts)
}

// GetGroupSAMLLink calls the underlying MockGetGroupSAMLLink method.
func (c *MockClient) GetGroupSAMLLink(pid interface{}, samlGroupName string, options ...gitlab.RequestOptionFunc) (*gitlab.SAMLGroupLink, *gitlab.Response, error) {
	return c.MockGetGroupSAMLLink(pid, samlGroupName)
}

// AddGroupSAMLLink call the underlying MockAddGroupSAMLLink method.
func (c *MockClient) AddGroupSAMLLink(pid interface{}, opt *gitlab.AddGroupSAMLLinkOptions, options ...gitlab.RequestOptionFunc) (*gitlab.SAMLGroupLink, *gitlab.Response, error) {
	return c.MockAddGroupSAMLLink(pid, opt)
}

// DeleteGroupSAMLLink calls the underlying MockDeleteGroupSAMLLink method.
func (c *MockClient) DeleteGroupSAMLLink(pid interface{}, samlGroupName string, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteGroupSAMLLink(pid, samlGroupName)
}
