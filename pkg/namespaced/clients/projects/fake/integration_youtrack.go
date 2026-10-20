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

var _ projects.YouTrackClient = &MockYouTrackClient{}

// MockYouTrackClient is a fake implementation of projects.YouTrackClient.
type MockYouTrackClient struct {
	MockGetYouTrackService    func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.YouTrackService, *gitlab.Response, error)
	MockSetYouTrackService    func(pid any, opt *gitlab.SetYouTrackServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.YouTrackService, *gitlab.Response, error)
	MockDeleteYouTrackService func(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// GetYouTrackService calls the underlying MockGetYouTrackService method.
func (c *MockYouTrackClient) GetYouTrackService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.YouTrackService, *gitlab.Response, error) {
	return c.MockGetYouTrackService(pid, options...)
}

// SetYouTrackService calls the underlying MockSetYouTrackService method.
func (c *MockYouTrackClient) SetYouTrackService(pid any, opt *gitlab.SetYouTrackServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.YouTrackService, *gitlab.Response, error) {
	return c.MockSetYouTrackService(pid, opt, options...)
}

// DeleteYouTrackService calls the underlying MockDeleteYouTrackService method.
func (c *MockYouTrackClient) DeleteYouTrackService(pid any, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return c.MockDeleteYouTrackService(pid, options...)
}
