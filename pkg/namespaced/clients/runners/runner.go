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
	"strings"

	gitlab "gitlab.com/gitlab-org/api/client-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	groupsv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	instancev1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/instance/v1alpha1"
	projectsv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
)

const (
	errRunnerNotFound = "404 Runner Not Found"
)

// RunnerClient defines Gitlab Runner service operations
type RunnerClient interface {
	GetRunnerDetails(rid any, options ...gitlab.RequestOptionFunc) (*gitlab.RunnerDetails, *gitlab.Response, error)
	UpdateRunnerDetails(rid any, opt *gitlab.UpdateRunnerDetailsOptions, options ...gitlab.RequestOptionFunc) (*gitlab.RunnerDetails, *gitlab.Response, error)
	DeleteRegisteredRunnerByID(rid int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
	ResetRunnerAuthenticationToken(rid int64, options ...gitlab.RequestOptionFunc) (*gitlab.RunnerAuthenticationToken, *gitlab.Response, error)
}

// NewRunnerClient returns a new Gitlab Runner service
func NewRunnerClient(cfg common.Config) RunnerClient {
	git := common.NewClient(cfg)
	return git.Runners
}

// IsErrorRunnerNotFound helper function to test for errRunnerNotFound error.
func IsErrorRunnerNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errRunnerNotFound)
}

// GenerateInstanceRunnerObservation is used to produce v1alpha1.RunnerObservation from
// gitlab.RunnerDetails.
func GenerateInstanceRunnerObservation(runner *gitlab.RunnerDetails) instancev1alpha1.RunnerObservation {
	if runner == nil {
		return instancev1alpha1.RunnerObservation{}
	}

	commonRunnerObservation := generateCommonRunnerObservation(runner)

	return instancev1alpha1.RunnerObservation{
		CommonRunnerObservation: commonRunnerObservation,
	}
}

// GenerateGroupRunnerObservation is used to produce groupsv1alpha1.RunnerObservation from
// gitlab.RunnerDetails.
func GenerateGroupRunnerObservation(runner *gitlab.RunnerDetails) groupsv1alpha1.RunnerObservation {
	if runner == nil {
		return groupsv1alpha1.RunnerObservation{}
	}

	commonRunnerObservation := generateCommonRunnerObservation(runner)

	// Convert each group to the RunnerGroup type
	// This is necessary because the API returns a different structure than the one we use in our
	// API definition.
	groups := make([]groupsv1alpha1.RunnerGroup, 0, len(runner.Groups))
	for _, group := range runner.Groups {
		groups = append(groups, groupsv1alpha1.RunnerGroup{
			ID:     group.ID,
			Name:   group.Name,
			WebURL: group.WebURL,
		})
	}

	return groupsv1alpha1.RunnerObservation{
		CommonRunnerObservation: commonRunnerObservation,
		Groups:                  groups,
	}
}

// GenerateGroupRunnerObservation is used to produce projectsv1alpha1.RunnerObservation from
// gitlab.RunnerDetails.
func GenerateProjectRunnerObservation(runner *gitlab.RunnerDetails) projectsv1alpha1.RunnerObservation {
	if runner == nil {
		return projectsv1alpha1.RunnerObservation{}
	}

	commonRunnerObservation := generateCommonRunnerObservation(runner)

	// Convert each project to the RunnerProject type
	// This is necessary because the API returns a different structure than the one we use in our
	// API definition.
	projects := make([]projectsv1alpha1.RunnerProject, 0, len(runner.Projects))
	for _, project := range runner.Projects {
		projects = append(projects, projectsv1alpha1.RunnerProject{
			ID:                project.ID,
			Name:              project.Name,
			NameWithNamespace: project.NameWithNamespace,
			Path:              project.Path,
			PathWithNamespace: project.PathWithNamespace,
		})
	}

	runnerObservation := projectsv1alpha1.RunnerObservation{
		CommonRunnerObservation: commonRunnerObservation,
		Projects:                projects,
	}

	return runnerObservation
}

// GenerateObservation is used to produce v1alpha1.RunnerObservation from
// gitlab.Runners.
func generateCommonRunnerObservation(runner *gitlab.RunnerDetails) commonv1alpha1.CommonRunnerObservation {
	if runner == nil {
		return commonv1alpha1.CommonRunnerObservation{}
	}
	runnerObservation := commonv1alpha1.CommonRunnerObservation{
		ID:              runner.ID,
		Description:     runner.Description,
		Paused:          runner.Paused,
		Locked:          runner.Locked,
		TagList:         runner.TagList,
		RunnerType:      runner.RunnerType,
		MaintenanceNote: runner.MaintenanceNote,
		Name:            runner.Name,
		Online:          runner.Online,
		Status:          runner.Status,
		RunUntagged:     runner.RunUntagged,
		AccessLevel:     runner.AccessLevel,
		MaximumTimeout:  runner.MaximumTimeout,
		IsShared:        runner.IsShared,
	}

	if runner.ContactedAt != nil {
		runnerObservation.ContactedAt = &metav1.Time{Time: *runner.ContactedAt}
	}

	return runnerObservation
}

// GenerateEditRunnerOptions generates group edit options
func GenerateEditRunnerOptions(p *commonv1alpha1.CommonRunnerParameters) *gitlab.UpdateRunnerDetailsOptions {
	return &gitlab.UpdateRunnerDetailsOptions{
		Description:     p.Description,
		Paused:          p.Paused,
		TagList:         p.TagList,
		RunUntagged:     p.RunUntagged,
		Locked:          p.Locked,
		AccessLevel:     p.AccessLevel,
		MaintenanceNote: p.MaintenanceNote,
		MaximumTimeout:  p.MaximumTimeout,
	}
}

// IsRunnerUpToDate checks whether the observed state of the runner matches the desired state specified
// in the RunnerParameters. It compares each relevant field and returns false if any discrepancies are found.
func IsRunnerUpToDate(spec *commonv1alpha1.CommonRunnerParameters, observed *gitlab.RunnerDetails) bool {
	if spec == nil {
		return true
	}
	if observed == nil {
		return false
	}
	// Convert observed.MaximumTimeout from int64 to int for comparison
	observedMaxTimeout := observed.MaximumTimeout
	// Use a compact list to keep cyclomatic complexity low
	checks := []bool{
		clients.IsComparableEqualToComparablePtr(spec.Description, observed.Description),
		clients.IsComparableEqualToComparablePtr(spec.Paused, observed.Paused),
		clients.IsComparableEqualToComparablePtr(spec.Locked, observed.Locked),
		clients.IsComparableEqualToComparablePtr(spec.RunUntagged, observed.RunUntagged),
		clients.IsComparableSliceEqualToComparableSlicePtr(spec.TagList, observed.TagList),
		clients.IsComparableEqualToComparablePtr(spec.AccessLevel, observed.AccessLevel),
		clients.IsComparableEqualToComparablePtr(spec.MaximumTimeout, observedMaxTimeout),
		clients.IsComparableEqualToComparablePtr(spec.MaintenanceNote, observed.MaintenanceNote),
	}
	for _, ok := range checks {
		if !ok {
			return false
		}
	}
	return true
}
