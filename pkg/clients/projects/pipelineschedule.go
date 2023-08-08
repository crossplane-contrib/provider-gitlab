/*
Copyright 2023 The Crossplane Authors.

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

package projects

import "github.com/xanzy/go-gitlab"

// PipelineScheduleClient is an interface for Gitlab PipelineScheduleService.
type PipelineScheduleClient interface {
	GetPipelineSchedule(pid interface{}, schedule int, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error)
	CreatePipelineSchedule(pid interface{}, opt *gitlab.CreatePipelineScheduleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error)
	EditPipelineSchedule(pid interface{}, schedule int, opt *gitlab.EditPipelineScheduleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error)
	DeletePipelineSchedule(pid interface{}, schedule int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)

	CreatePipelineScheduleVariable(pid interface{}, schedule int, opt *gitlab.CreatePipelineScheduleVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineVariable, *gitlab.Response, error)
	DeletePipelineScheduleVariable(pid interface{}, schedule int, key string, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineVariable, *gitlab.Response, error)
	EditPipelineScheduleVariable(pid interface{}, schedule int, key string, opt *gitlab.EditPipelineScheduleVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineVariable, *gitlab.Response, error)
}
