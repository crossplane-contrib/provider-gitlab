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

var _ projects.EmailsOnPushClient = &MockEmailsOnPushClient{}

// MockEmailsOnPushClient is a fake implementation of projects.EmailsOnPushClient.
type MockEmailsOnPushClient struct {
	MockGetEmailsOnPushService    func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.EmailsOnPushService, *gitlab.Response, error)
	MockSetEmailsOnPushService    func(pid any, opt *gitlab.SetEmailsOnPushServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.EmailsOnPushService, *gitlab.Response, error)
	MockDeleteEmailsOnPushService func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// GetEmailsOnPushService calls the underlying MockGetEmailsOnPushService method.
func (c *MockEmailsOnPushClient) GetEmailsOnPushService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.EmailsOnPushService, *gitlab.Response, error) {
	return c.MockGetEmailsOnPushService(pid, options...)
}

// SetEmailsOnPushService calls the underlying MockSetEmailsOnPushService method.
func (c *MockEmailsOnPushClient) SetEmailsOnPushService(pid any, opt *gitlab.SetEmailsOnPushServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.EmailsOnPushService, *gitlab.Response, error) {
	return c.MockSetEmailsOnPushService(pid, opt, options...)
}

// DeleteEmailsOnPushService calls the underlying MockDeleteEmailsOnPushService method.
func (c *MockEmailsOnPushClient) DeleteEmailsOnPushService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteEmailsOnPushService(pid, options...)
}
