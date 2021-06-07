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

package projects

import (
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
)

const (
	errProjectHookNotFound = "404 Not found"
)

// ProjectHookClient defines Gitlab ProjectHook service operations
type ProjectHookClient interface {
	GetProjectHook(pid interface{}, hook int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error)
	AddProjectHook(pid interface{}, opt *gitlab.AddProjectHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error)
	EditProjectHook(pid interface{}, hook int, opt *gitlab.EditProjectHookOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error)
	DeleteProjectHook(pid interface{}, hook int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewProjectHookClient returns a new Gitlab Project service
func NewProjectHookClient(cfg clients.Config) ProjectHookClient {
	git := clients.NewClient(cfg)
	return git.Projects
}

// IsErrorProjectHookNotFound helper function to test for errProjectNotFound error.
func IsErrorProjectHookNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errProjectHookNotFound)
}

// LateInitializeProjectHook fills the empty fields in the projecthook spec with the
// values seen in gitlab.ProjectHook.
func LateInitializeProjectHook(in *v1alpha1.ProjectHookParameters, projecthook *gitlab.ProjectHook) { // nolint:gocyclo
	if projecthook == nil {
		return
	}

	if in.ConfidentialNoteEvents == nil {
		in.ConfidentialNoteEvents = &projecthook.ConfidentialNoteEvents
	}
	if in.PushEvents == nil {
		in.PushEvents = &projecthook.PushEvents
	}
	if in.IssuesEvents == nil {
		in.IssuesEvents = &projecthook.IssuesEvents
	}
	in.PushEventsBranchFilter = clients.LateInitializeStringPtr(in.PushEventsBranchFilter, projecthook.PushEventsBranchFilter)
	if in.ConfidentialIssuesEvents == nil {
		in.ConfidentialIssuesEvents = &projecthook.ConfidentialIssuesEvents
	}
	if in.MergeRequestsEvents == nil {
		in.MergeRequestsEvents = &projecthook.MergeRequestsEvents
	}
	if in.TagPushEvents == nil {
		in.TagPushEvents = &projecthook.TagPushEvents
	}
	if in.NoteEvents == nil {
		in.NoteEvents = &projecthook.NoteEvents
	}
	if in.JobEvents == nil {
		in.JobEvents = &projecthook.JobEvents
	}
	if in.PipelineEvents == nil {
		in.PipelineEvents = &projecthook.PipelineEvents
	}
	if in.WikiPageEvents == nil {
		in.WikiPageEvents = &projecthook.WikiPageEvents
	}
	if in.EnableSSLVerification == nil {
		in.EnableSSLVerification = &projecthook.EnableSSLVerification
	}
}

// GenerateProjectHookObservation is used to produce v1alpha1.ProjectHookObservation from
// gitlab.ProjectHook.
func GenerateProjectHookObservation(projecthook *gitlab.ProjectHook) v1alpha1.ProjectHookObservation { // nolint:gocyclo
	if projecthook == nil {
		return v1alpha1.ProjectHookObservation{}
	}

	o := v1alpha1.ProjectHookObservation{
		ID: projecthook.ID,
	}

	if projecthook.CreatedAt != nil {
		o.CreatedAt = &metav1.Time{Time: *projecthook.CreatedAt}
	}
	return o
}

// GenerateCreateProjectHookOptions generates project creation options
func GenerateCreateProjectHookOptions(p *v1alpha1.ProjectHookParameters) *gitlab.AddProjectHookOptions {
	projecthook := &gitlab.AddProjectHookOptions{
		URL:                      p.URL,
		ConfidentialNoteEvents:   p.ConfidentialNoteEvents,
		PushEvents:               p.PushEvents,
		PushEventsBranchFilter:   p.PushEventsBranchFilter,
		IssuesEvents:             p.IssuesEvents,
		ConfidentialIssuesEvents: p.ConfidentialIssuesEvents,
		MergeRequestsEvents:      p.MergeRequestsEvents,
		TagPushEvents:            p.TagPushEvents,
		NoteEvents:               p.NoteEvents,
		JobEvents:                p.JobEvents,
		PipelineEvents:           p.PipelineEvents,
		WikiPageEvents:           p.WikiPageEvents,
		EnableSSLVerification:    p.EnableSSLVerification,
		Token:                    p.Token,
	}

	return projecthook
}

// GenerateEditProjectHookOptions generates project edit options
func GenerateEditProjectHookOptions(p *v1alpha1.ProjectHookParameters) *gitlab.EditProjectHookOptions {
	o := &gitlab.EditProjectHookOptions{
		URL:                      p.URL,
		ConfidentialNoteEvents:   p.ConfidentialNoteEvents,
		PushEvents:               p.PushEvents,
		PushEventsBranchFilter:   p.PushEventsBranchFilter,
		IssuesEvents:             p.IssuesEvents,
		ConfidentialIssuesEvents: p.ConfidentialIssuesEvents,
		MergeRequestsEvents:      p.MergeRequestsEvents,
		TagPushEvents:            p.TagPushEvents,
		NoteEvents:               p.NoteEvents,
		JobEvents:                p.JobEvents,
		PipelineEvents:           p.PipelineEvents,
		WikiPageEvents:           p.WikiPageEvents,
		EnableSSLVerification:    p.EnableSSLVerification,
		Token:                    p.Token,
	}

	return o
}

// IsProjectHookUpToDate checks whether there is a change in any of the modifiable fields.
func IsProjectHookUpToDate(p *v1alpha1.ProjectHookParameters, g *gitlab.ProjectHook) bool { // nolint:gocyclo
	if !cmp.Equal(p.URL, clients.StringToPtr(g.URL)) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.ConfidentialNoteEvents, g.ConfidentialNoteEvents) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.PushEvents, g.PushEvents) {
		return false
	}
	if !cmp.Equal(p.PushEventsBranchFilter, clients.StringToPtr(g.PushEventsBranchFilter)) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.IssuesEvents, g.IssuesEvents) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.ConfidentialIssuesEvents, g.ConfidentialIssuesEvents) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.MergeRequestsEvents, g.MergeRequestsEvents) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.TagPushEvents, g.TagPushEvents) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.NoteEvents, g.NoteEvents) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.JobEvents, g.JobEvents) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.PipelineEvents, g.PipelineEvents) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.WikiPageEvents, g.WikiPageEvents) {
		return false
	}
	if !clients.IsBoolEqualToBoolPtr(p.EnableSSLVerification, g.EnableSSLVerification) {
		return false
	}

	return true
}
