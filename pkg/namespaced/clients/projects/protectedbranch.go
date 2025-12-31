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

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
)

const (
	errProtectedBranchNotFound = "404 Not found"
)

// ProtectedBranchClient defines GitLab Protected Branch service operations
type ProtectedBranchClient interface {
	GetProtectedBranch(pid interface{}, branch string, options ...gitlab.RequestOptionFunc) (*gitlab.ProtectedBranch, *gitlab.Response, error)
	ProtectRepositoryBranches(pid interface{}, opt *gitlab.ProtectRepositoryBranchesOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProtectedBranch, *gitlab.Response, error)
	UnprotectRepositoryBranches(pid interface{}, branch string, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewProtectedBranchClient returns a new GitLab Protected Branch client
func NewProtectedBranchClient(cfg common.Config) ProtectedBranchClient {
	git := common.NewClient(cfg)
	return git.ProtectedBranches
}

// IsErrorProtectedBranchNotFound helper function to test for errProtectedBranchNotFound error.
func IsErrorProtectedBranchNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errProtectedBranchNotFound)
}

// LateInitializeProtectedBranch fills the empty fields in the protected branch spec with the
// values seen in gitlab.ProtectedBranch.
func LateInitializeProtectedBranch(in *v1alpha1.ProtectedBranchParameters, pb *gitlab.ProtectedBranch) { //nolint:gocyclo
	if pb == nil {
		return
	}

	if in.AllowForcePush == nil {
		in.AllowForcePush = &pb.AllowForcePush
	}
	if in.CodeOwnerApprovalRequired == nil {
		in.CodeOwnerApprovalRequired = &pb.CodeOwnerApprovalRequired
	}

	// Late initialize access levels
	if len(in.PushAccessLevels) == 0 && len(pb.PushAccessLevels) > 0 {
		in.PushAccessLevels = make([]*v1alpha1.BranchAccessDescription, len(pb.PushAccessLevels))
		for i, pal := range pb.PushAccessLevels {
			in.PushAccessLevels[i] = &v1alpha1.BranchAccessDescription{
				AccessLevel:            (*v1alpha1.AccessLevelValue)(&pal.AccessLevel),
				AccessLevelDescription: &pal.AccessLevelDescription,
				UserID:                 &pal.UserID,
				GroupID:                &pal.GroupID,
			}
		}
	}

	if len(in.MergeAccessLevels) == 0 && len(pb.MergeAccessLevels) > 0 {
		in.MergeAccessLevels = make([]*v1alpha1.BranchAccessDescription, len(pb.MergeAccessLevels))
		for i, mal := range pb.MergeAccessLevels {
			in.MergeAccessLevels[i] = &v1alpha1.BranchAccessDescription{
				AccessLevel:            (*v1alpha1.AccessLevelValue)(&mal.AccessLevel),
				AccessLevelDescription: &mal.AccessLevelDescription,
				UserID:                 &mal.UserID,
				GroupID:                &mal.GroupID,
			}
		}
	}

	if len(in.UnprotectAccessLevels) == 0 && len(pb.UnprotectAccessLevels) > 0 {
		in.UnprotectAccessLevels = make([]*v1alpha1.BranchAccessDescription, len(pb.UnprotectAccessLevels))
		for i, ual := range pb.UnprotectAccessLevels {
			in.UnprotectAccessLevels[i] = &v1alpha1.BranchAccessDescription{
				AccessLevel:            (*v1alpha1.AccessLevelValue)(&ual.AccessLevel),
				AccessLevelDescription: &ual.AccessLevelDescription,
				UserID:                 &ual.UserID,
				GroupID:                &ual.GroupID,
			}
		}
	}
}

// GenerateProtectedBranchObservation produces a ProtectedBranchObservation from a gitlab.ProtectedBranch
func GenerateProtectedBranchObservation(pb *gitlab.ProtectedBranch) v1alpha1.ProtectedBranchObservation {
	if pb == nil {
		return v1alpha1.ProtectedBranchObservation{}
	}

	o := v1alpha1.ProtectedBranchObservation{
		ID:                        pb.ID,
		AllowForcePush:            pb.AllowForcePush,
		CodeOwnerApprovalRequired: pb.CodeOwnerApprovalRequired,
	}

	// Convert push access levels
	if len(pb.PushAccessLevels) > 0 {
		o.PushAccessLevels = make([]*v1alpha1.BranchAccessDescription, len(pb.PushAccessLevels))
		for i, pal := range pb.PushAccessLevels {
			o.PushAccessLevels[i] = &v1alpha1.BranchAccessDescription{
				AccessLevel:            (*v1alpha1.AccessLevelValue)(&pal.AccessLevel),
				AccessLevelDescription: &pal.AccessLevelDescription,
				UserID:                 &pal.UserID,
				GroupID:                &pal.GroupID,
			}
		}
	}

	// Convert merge access levels
	if len(pb.MergeAccessLevels) > 0 {
		o.MergeAccessLevels = make([]*v1alpha1.BranchAccessDescription, len(pb.MergeAccessLevels))
		for i, mal := range pb.MergeAccessLevels {
			o.MergeAccessLevels[i] = &v1alpha1.BranchAccessDescription{
				AccessLevel:            (*v1alpha1.AccessLevelValue)(&mal.AccessLevel),
				AccessLevelDescription: &mal.AccessLevelDescription,
				UserID:                 &mal.UserID,
				GroupID:                &mal.GroupID,
			}
		}
	}

	// Convert unprotect access levels
	if len(pb.UnprotectAccessLevels) > 0 {
		o.UnprotectAccessLevels = make([]*v1alpha1.BranchAccessDescription, len(pb.UnprotectAccessLevels))
		for i, ual := range pb.UnprotectAccessLevels {
			o.UnprotectAccessLevels[i] = &v1alpha1.BranchAccessDescription{
				AccessLevel:            (*v1alpha1.AccessLevelValue)(&ual.AccessLevel),
				AccessLevelDescription: &ual.AccessLevelDescription,
				UserID:                 &ual.UserID,
				GroupID:                &ual.GroupID,
			}
		}
	}

	return o
}

// GenerateProtectRepositoryBranchesOptions produces *gitlab.ProtectRepositoryBranchesOptions from ProtectedBranchParameters
func GenerateProtectRepositoryBranchesOptions(name string, p *v1alpha1.ProtectedBranchParameters) *gitlab.ProtectRepositoryBranchesOptions {
	opt := &gitlab.ProtectRepositoryBranchesOptions{
		Name: &name,
	}

	if p.AllowForcePush != nil {
		opt.AllowForcePush = p.AllowForcePush
	}
	if p.CodeOwnerApprovalRequired != nil {
		opt.CodeOwnerApprovalRequired = p.CodeOwnerApprovalRequired
	}

	// For GitLab API, access levels are typically set as simple access level values
	// rather than arrays of complex permissions
	if len(p.PushAccessLevels) > 0 && p.PushAccessLevels[0].AccessLevel != nil {
		accessLevel := gitlab.AccessLevelValue(*p.PushAccessLevels[0].AccessLevel)
		opt.PushAccessLevel = &accessLevel
	}

	if len(p.MergeAccessLevels) > 0 && p.MergeAccessLevels[0].AccessLevel != nil {
		accessLevel := gitlab.AccessLevelValue(*p.MergeAccessLevels[0].AccessLevel)
		opt.MergeAccessLevel = &accessLevel
	}

	if len(p.UnprotectAccessLevels) > 0 && p.UnprotectAccessLevels[0].AccessLevel != nil {
		accessLevel := gitlab.AccessLevelValue(*p.UnprotectAccessLevels[0].AccessLevel)
		opt.UnprotectAccessLevel = &accessLevel
	}

	return opt
}

// IsProtectedBranchUpToDate checks whether there is a change in any of the modifiable fields.
func IsProtectedBranchUpToDate(p *v1alpha1.ProtectedBranchParameters, pb *gitlab.ProtectedBranch) bool {
	if pb == nil {
		return false
	}

	if p.AllowForcePush != nil && *p.AllowForcePush != pb.AllowForcePush {
		return false
	}
	if p.CodeOwnerApprovalRequired != nil && *p.CodeOwnerApprovalRequired != pb.CodeOwnerApprovalRequired {
		return false
	}

	// Compare access levels
	if !isAccessLevelsUpToDate(p.PushAccessLevels, pb.PushAccessLevels) {
		return false
	}
	if !isAccessLevelsUpToDate(p.MergeAccessLevels, pb.MergeAccessLevels) {
		return false
	}
	if !isAccessLevelsUpToDate(p.UnprotectAccessLevels, pb.UnprotectAccessLevels) {
		return false
	}

	return true
}

// isAccessLevelsUpToDate compares access levels between spec and GitLab
func isAccessLevelsUpToDate(specLevels []*v1alpha1.BranchAccessDescription, gitlabLevels []*gitlab.BranchAccessDescription) bool { //nolint:gocyclo
	if len(specLevels) != len(gitlabLevels) {
		return false
	}

	// Simple comparison - if lengths match and all access levels match
	for _, specLevel := range specLevels {
		found := false
		for _, gitlabLevel := range gitlabLevels {
			if specLevel.AccessLevel != nil && int64(*specLevel.AccessLevel) == int64(gitlabLevel.AccessLevel) {
				// Check if user and group IDs also match
				userMatch := (specLevel.UserID == nil && gitlabLevel.UserID == 0) || (specLevel.UserID != nil && *specLevel.UserID == gitlabLevel.UserID)
				groupMatch := (specLevel.GroupID == nil && gitlabLevel.GroupID == 0) || (specLevel.GroupID != nil && *specLevel.GroupID == gitlabLevel.GroupID)

				if userMatch && groupMatch {
					found = true
					break
				}
			}
		}
		if !found {
			return false
		}
	}

	return true
}
