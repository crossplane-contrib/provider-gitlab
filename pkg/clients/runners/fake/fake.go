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

	runners "github.com/crossplane-contrib/provider-gitlab/pkg/clients/runners"
)

var _ runners.RunnerClient = &MockClient{}

type MockClient struct {
	runners.RunnerClient

	MockGetRunnerDetails               func(rid any, options ...gitlab.RequestOptionFunc) (*gitlab.RunnerDetails, *gitlab.Response, error)
	MockUpdateRunnerDetails            func(rid any, opt *gitlab.UpdateRunnerDetailsOptions, options ...gitlab.RequestOptionFunc) (*gitlab.RunnerDetails, *gitlab.Response, error)
	MockDeleteRegisteredRunnerByID     func(rid int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
	MockResetRunnerAuthenticationToken func(rid int, options ...gitlab.RequestOptionFunc) (*gitlab.RunnerAuthenticationToken, *gitlab.Response, error)
}

func (m *MockClient) GetRunnerDetails(rid any, options ...gitlab.RequestOptionFunc) (*gitlab.RunnerDetails, *gitlab.Response, error) {
	return m.MockGetRunnerDetails(rid)
}

func (m *MockClient) UpdateRunnerDetails(rid any, opt *gitlab.UpdateRunnerDetailsOptions, options ...gitlab.RequestOptionFunc) (*gitlab.RunnerDetails, *gitlab.Response, error) {
	return m.MockUpdateRunnerDetails(rid, opt)
}

func (m *MockClient) DeleteRegisteredRunnerByID(rid int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return m.MockDeleteRegisteredRunnerByID(rid)
}

func (m *MockClient) ResetRunnerAuthenticationToken(rid int, options ...gitlab.RequestOptionFunc) (*gitlab.RunnerAuthenticationToken, *gitlab.Response, error) {
	return m.MockResetRunnerAuthenticationToken(rid)
}
