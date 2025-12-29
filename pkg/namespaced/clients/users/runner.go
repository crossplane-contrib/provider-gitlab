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

package users

import (
	gitlab "gitlab.com/gitlab-org/api/client-go"

	commonv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	groupsv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	instancev1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/instance/v1alpha1"
	projectsv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
)

var (
	groupRunnerType    string = "group_type"
	projectRunnerType  string = "project_type"
	instanceRunnerType string = "instance_type"
)

// RunnerClient defines Gitlab User service operations
type RunnerClient interface {
	CreateUserRunner(opts *gitlab.CreateUserRunnerOptions, options ...gitlab.RequestOptionFunc) (*gitlab.UserRunner, *gitlab.Response, error)
}

// NewRunnerClient returns a new Gitlab User service
func NewRunnerClient(cfg common.Config) RunnerClient {
	git := common.NewClient(cfg)
	return git.Users
}

// GenerateInstanceRunnerOptions generates user runner creation options for a runner linked to the instance
func GenerateInstanceRunnerOptions(p *instancev1alpha1.RunnerParameters) *gitlab.CreateUserRunnerOptions {
	opts := generateCommonRunnerOptions(&p.CommonRunnerParameters)

	opts.RunnerType = &instanceRunnerType

	return opts
}

// GenerateGroupRunnerOptions generates user runner creation options for a runner linked to a group
func GenerateGroupRunnerOptions(p *groupsv1alpha1.RunnerParameters) *gitlab.CreateUserRunnerOptions {
	opts := generateCommonRunnerOptions(&p.CommonRunnerParameters)

<<<<<<< HEAD
	opts.GroupID = p.GroupID
=======
	if p.GroupID != nil {
		groupID := int64(*p.GroupID)
		opts.GroupID = &groupID
	}

>>>>>>> 77c306d (feat: migrate CRD types from *int to *int64)
	opts.RunnerType = &groupRunnerType

	return opts
}

// GenerateProjectRunnerOptions generates user runner creation options for a runner linked to a project
func GenerateProjectRunnerOptions(p *projectsv1alpha1.RunnerParameters) *gitlab.CreateUserRunnerOptions {
	opts := generateCommonRunnerOptions(&p.CommonRunnerParameters)

<<<<<<< HEAD
	opts.ProjectID = p.ProjectID
=======
	if p.ProjectID != nil {
		projectID := int64(*p.ProjectID)
		opts.ProjectID = &projectID
	}

>>>>>>> 77c306d (feat: migrate CRD types from *int to *int64)
	opts.RunnerType = &projectRunnerType

	return opts
}

// generateCommonRunnerOptions generates user runner creation options common to all runner types
func generateCommonRunnerOptions(p *commonv1alpha1.CommonRunnerParameters) *gitlab.CreateUserRunnerOptions {
	opts := &gitlab.CreateUserRunnerOptions{
		Description:     p.Description,
		Paused:          p.Paused,
		Locked:          p.Locked,
		RunUntagged:     p.RunUntagged,
		TagList:         p.TagList,
		AccessLevel:     p.AccessLevel,
		MaintenanceNote: p.MaintenanceNote,
		MaximumTimeout:  p.MaximumTimeout,
	}
	if p.MaximumTimeout != nil {
		maxTimeout := int64(*p.MaximumTimeout)
		opts.MaximumTimeout = &maxTimeout
	}
	return opts
}
