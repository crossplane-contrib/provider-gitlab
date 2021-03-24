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

	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/deploykeys"
)

var _ deploykeys.Client = &MockClient{}

// MockClient is a fake implementation of deploykeys.Client.
type MockClient struct {
	deploykeys.Client

	MockGetDeployKey    func(pid interface{}, deployKeyID int, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error)
	MockAddDeployKey    func(pid interface{}, opt *gitlab.AddDeployKeyOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error)
	MockUpdateDeployKey func(pid interface{}, deployKeyID int, opt *gitlab.UpdateDeployKeyOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error)
	MockDeleteDeployKey func(pid interface{}, deployKeyID int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// GetDeployKey calls the underlying MockGetDeployKey method.
func (c *MockClient) GetDeployKey(pid interface{}, deployKeyID int, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error) {
	return c.MockGetDeployKey(pid, deployKeyID)
}

// AddDeployKey calls the underlying MockAddDeployKey method.
func (c *MockClient) AddDeployKey(pid interface{}, opt *gitlab.AddDeployKeyOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error) {
	return c.MockAddDeployKey(pid, opt)
}

// UpdateDeployKey calls the underlying MockEditDeployKey method.
func (c *MockClient) UpdateDeployKey(pid interface{}, deployKeyID int, opt *gitlab.UpdateDeployKeyOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error) {
	return c.MockUpdateDeployKey(pid, deployKeyID, opt)
}

// DeleteDeployKey calls the underlying MockDeleteDeployKey method.
func (c *MockClient) DeleteDeployKey(pid interface{}, deployKeyID int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteDeployKey(pid, deployKeyID)
}
