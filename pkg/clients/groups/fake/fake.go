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

	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups"
)

var _ groups.Client = &MockClient{}

// MockClient is a fake implementation of groups.Client.
type MockClient struct {
	groups.Client

	MockGetGroup    func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error)
	MockCreateGroup func(opt *gitlab.CreateGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error)
	MockUpdateGroup func(pid interface{}, opt *gitlab.UpdateGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error)
	MockDeleteGroup func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// GetGroup calls the underlying MockGetGroup method.
func (c *MockClient) GetGroup(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
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
func (c *MockClient) DeleteGroup(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteGroup(pid)
}
