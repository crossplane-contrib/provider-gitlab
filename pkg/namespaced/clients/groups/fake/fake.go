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
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"

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
	MockListGroups            func(opt *gitlab.ListGroupsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.Group, *gitlab.Response, error)

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
	MockRotateGroupAccessToken func(gid interface{}, accessToken int64, opt *gitlab.RotateGroupAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupAccessToken, *gitlab.Response, error)
	MockRotateSelf             func(opt *gitlab.RotatePersonalAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error)

	MockListServiceAccountPersonalAccessTokens  func(gid any, serviceAccount int64, opt *gitlab.ListServiceAccountPersonalAccessTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.PersonalAccessToken, *gitlab.Response, error)
	MockCreateServiceAccountPersonalAccessToken func(gid any, serviceAccount int64, opt *gitlab.CreateServiceAccountPersonalAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error)
	MockRevokeServiceAccountPersonalAccessToken func(gid any, serviceAccount, token int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
	MockRotateServiceAccountPersonalAccessToken func(gid any, serviceAccount, token int64, opt *gitlab.RotateServiceAccountPersonalAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error)
	MockGetServiceAccountSelf                   func(options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error)
	MockRotateServiceAccountSelf                func(opt *gitlab.RotatePersonalAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error)
	MockRevokeServiceAccountSelf                func(options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

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

	MockGetGroupHook    func(gid interface{}, hook int64, options ...gitlab.RequestOptionFunc) (*gitlab.GroupHook, *gitlab.Response, error)
	MockAddGroupHook    func(gid interface{}, opt *gitlab.AddGroupHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupHook, *gitlab.Response, error)
	MockEditGroupHook   func(gid interface{}, hook int64, opt *gitlab.EditGroupHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupHook, *gitlab.Response, error)
	MockDeleteGroupHook func(gid interface{}, hook int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

	MockGetGroupHarborSettings func(gid any, options ...gitlab.RequestOptionFunc) (*gitlab.HarborIntegration, *gitlab.Response, error)
	MockSetUpGroupHarbor       func(gid any, opt *gitlab.SetUpHarborOptions, options ...gitlab.RequestOptionFunc) (*gitlab.HarborIntegration, *gitlab.Response, error)
	MockDisableGroupHarbor     func(gid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// GetGroupHarborSettings calls the underlying MockGetGroupHarborSettings method.
func (c *MockClient) GetGroupHarborSettings(gid any, options ...gitlab.RequestOptionFunc) (*gitlab.HarborIntegration, *gitlab.Response, error) {
	return c.MockGetGroupHarborSettings(gid, options...)
}

// SetUpGroupHarbor calls the underlying MockSetUpGroupHarbor method.
func (c *MockClient) SetUpGroupHarbor(gid any, opt *gitlab.SetUpHarborOptions, options ...gitlab.RequestOptionFunc) (*gitlab.HarborIntegration, *gitlab.Response, error) {
	return c.MockSetUpGroupHarbor(gid, opt, options...)
}

// DisableGroupHarbor calls the underlying MockDisableGroupHarbor method.
func (c *MockClient) DisableGroupHarbor(gid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDisableGroupHarbor(gid, options...)
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

// ListGroups calls the underlying MockListGroups method.
func (c *MockClient) ListGroups(opt *gitlab.ListGroupsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.Group, *gitlab.Response, error) {
	return c.MockListGroups(opt, options...)
}

// GetGroupMember calls the underlying MockGetMember method.
func (c *MockClient) GetGroupMember(gid interface{}, user int64, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
	return c.MockGetMember(gid, user, options...)
}

// AddGroupMember calls the underlying MockAddMember method.
func (c *MockClient) AddGroupMember(gid interface{}, opt *gitlab.AddGroupMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
	return c.MockAddMember(gid, opt, options...)
}

// EditGroupMember calls the underlying MockEditMember method.
func (c *MockClient) EditGroupMember(gid interface{}, user int64, opt *gitlab.EditGroupMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
	return c.MockEditMember(gid, user, opt, options...)
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

// RotateGroupAccessToken calls the underlying MockRotateGroupAccessToken method.
func (c *MockClient) RotateGroupAccessToken(gid interface{}, accessToken int64, opt *gitlab.RotateGroupAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupAccessToken, *gitlab.Response, error) {
	return c.MockRotateGroupAccessToken(gid, accessToken, opt, options...)
}

// RotateSelf calls the underlying MockRotateSelf method.
func (c *MockClient) RotateSelf(opt *gitlab.RotatePersonalAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return c.MockRotateSelf(opt, options...)
}

// ListServiceAccountPersonalAccessTokens calls the underlying MockListServiceAccountPersonalAccessTokens method.
func (c *MockClient) ListServiceAccountPersonalAccessTokens(gid any, serviceAccount int64, opt *gitlab.ListServiceAccountPersonalAccessTokensOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return c.MockListServiceAccountPersonalAccessTokens(gid, serviceAccount, opt, options...)
}

// CreateServiceAccountPersonalAccessToken calls the underlying MockCreateServiceAccountPersonalAccessToken method.
func (c *MockClient) CreateServiceAccountPersonalAccessToken(gid any, serviceAccount int64, opt *gitlab.CreateServiceAccountPersonalAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return c.MockCreateServiceAccountPersonalAccessToken(gid, serviceAccount, opt, options...)
}

// RevokeServiceAccountPersonalAccessToken calls the underlying MockRevokeServiceAccountPersonalAccessToken method.
func (c *MockClient) RevokeServiceAccountPersonalAccessToken(gid any, serviceAccount, token int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockRevokeServiceAccountPersonalAccessToken(gid, serviceAccount, token, options...)
}

// RotateServiceAccountPersonalAccessToken calls the underlying MockRotateServiceAccountPersonalAccessToken method.
func (c *MockClient) RotateServiceAccountPersonalAccessToken(gid any, serviceAccount, token int64, opt *gitlab.RotateServiceAccountPersonalAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return c.MockRotateServiceAccountPersonalAccessToken(gid, serviceAccount, token, opt, options...)
}

// RotateServiceAccountSelf calls the underlying MockRotateServiceAccountSelf method.
func (c *MockClient) RotateServiceAccountSelf(opt *gitlab.RotatePersonalAccessTokenOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return c.MockRotateServiceAccountSelf(opt, options...)
}

// GetServiceAccountSelf calls the underlying MockGetServiceAccountSelf method.
func (c *MockClient) GetServiceAccountSelf(options ...gitlab.RequestOptionFunc) (*gitlab.PersonalAccessToken, *gitlab.Response, error) {
	return c.MockGetServiceAccountSelf(options...)
}

// RevokeServiceAccountSelf calls the underlying MockRevokeServiceAccountSelf method.
func (c *MockClient) RevokeServiceAccountSelf(options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockRevokeServiceAccountSelf(options...)
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

// GetGroupHook calls the underlying MockGetGroupHook method.
func (c *MockClient) GetGroupHook(gid interface{}, hook int64, options ...gitlab.RequestOptionFunc) (*gitlab.GroupHook, *gitlab.Response, error) {
	return c.MockGetGroupHook(gid, hook)
}

// AddGroupHook calls the underlying MockAddGroupHook method.
func (c *MockClient) AddGroupHook(gid interface{}, opt *gitlab.AddGroupHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupHook, *gitlab.Response, error) {
	return c.MockAddGroupHook(gid, opt)
}

// EditGroupHook calls the underlying MockEditGroupHook method.
func (c *MockClient) EditGroupHook(gid interface{}, hook int64, opt *gitlab.EditGroupHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupHook, *gitlab.Response, error) {
	return c.MockEditGroupHook(gid, hook, opt)
}

// DeleteGroupHook calls the underlying MockDeleteGroupHook method.
func (c *MockClient) DeleteGroupHook(gid interface{}, hook int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteGroupHook(gid, hook)
}
