// +cluster-scope:delete=1
//go:generate go run generate.go

package instance

import (
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/instance/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
)

// ApplicationSettingsClient defines Gitlab Application Settings service operations
type ApplicationSettingsClient interface {
	GetSettings(options ...gitlab.RequestOptionFunc) (*gitlab.Settings, *gitlab.Response, error)
	UpdateSettings(opt *gitlab.UpdateSettingsOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Settings, *gitlab.Response, error)
}

// NewApplicationSettingsClient returns a new Gitlab Application Settings service
func NewApplicationSettingsClient(cfg common.Config) ApplicationSettingsClient {
	git := common.NewClient(cfg)
	return git.Settings
}

// isDefaultBranchProtectionDefaultsPtrEqualToDefaultsPtr compares a *gitlab.BranchProtectionDefaults with gitlab.BranchProtectionDefaults
func isDefaultBranchProtectionDefaultsPtrEqualToDefaultsPtr(dbpdv *v1alpha1.DefaultBranchProtectionDefaultsOptions, dbpdg *gitlab.BranchProtectionDefaults) bool {
	if dbpdv == nil {
		return true
	}
	if dbpdg == nil {
		return false
	}
	equals := clients.IsComparableEqualToComparablePtr(dbpdv.AllowForcePush, dbpdg.AllowForcePush)
	equals = equals && clients.IsComparableEqualToComparablePtr(dbpdv.DeveloperCanInitialPush, dbpdg.DeveloperCanInitialPush)
	equals = equals && clients.IsComparableSliceEqualToComparableSlicePtr(dbpdv.AllowedToMerge, groupAccessSliceToIntSlice(dbpdg.AllowedToMerge))
	equals = equals && clients.IsComparableSliceEqualToComparableSlicePtr(dbpdv.AllowedToPush, groupAccessSliceToIntSlice(dbpdg.AllowedToPush))
	return equals
}

// gitlabBranchProtectionDefaultsTov1Alpha1BranchProtectionDefaults converts v1alpha1.DefaultBranchProtectionDefaultsOptions to gitlab.BranchProtectionDefaults
func gitlabBranchProtectionDefaultsTov1Alpha1BranchProtectionDefaults(input *gitlab.BranchProtectionDefaults) v1alpha1.BranchProtectionDefaults {
	result := v1alpha1.BranchProtectionDefaults{}
	if input == nil {
		return result
	}
	result.AllowForcePush = input.AllowForcePush
	result.DeveloperCanInitialPush = input.DeveloperCanInitialPush
	if input.AllowedToMerge != nil {
		result.AllowedToMerge = make([]*int, len(input.AllowedToMerge))
		for i, v := range input.AllowedToMerge {
			if v != nil && v.AccessLevel != nil {
				buf := int(*v.AccessLevel)
				result.AllowedToMerge[i] = &buf
			}
		}
	}
	if input.AllowedToPush != nil {
		result.AllowedToPush = make([]*int, len(input.AllowedToPush))
		for i, v := range input.AllowedToPush {
			if v != nil && v.AccessLevel != nil {
				buf := int(*v.AccessLevel)
				result.AllowedToPush[i] = &buf
			}
		}
	}
	return result
}

// gitlabBranchProtectionDefaultsTov1Alpha1BranchProtectionDefaultsOptions converts v1alpha1.DefaultBranchProtectionDefaultsOptions to gitlab.BranchProtectionDefaults
func gitlabBranchProtectionDefaultsTov1Alpha1BranchProtectionDefaultsOptions(input *gitlab.BranchProtectionDefaults) v1alpha1.DefaultBranchProtectionDefaultsOptions {
	result := v1alpha1.DefaultBranchProtectionDefaultsOptions{}
	if input == nil {
		return result
	}
	result.AllowForcePush = &input.AllowForcePush
	result.DeveloperCanInitialPush = &input.DeveloperCanInitialPush
	if input.AllowedToMerge != nil {
		allowedToMerge := make([]int, len(input.AllowedToMerge))
		for i, v := range input.AllowedToMerge {
			if v != nil && v.AccessLevel != nil {
				allowedToMerge[i] = int(*v.AccessLevel)
			}
		}
		result.AllowedToMerge = &allowedToMerge
	}
	if input.AllowedToPush != nil {
		allowedToPush := make([]int, len(input.AllowedToPush))
		for i, v := range input.AllowedToPush {
			if v != nil && v.AccessLevel != nil {
				allowedToPush[i] = int(*v.AccessLevel)
			}
		}
		result.AllowedToPush = &allowedToPush
	}
	return result
}

// v1Alpha1DefaultBranchProtectionDefaultsOptionsPtrToGitlabBranchProtectionDefaults converts v1alpha1.DefaultBranchProtectionDefaultsOptions to gitlab.BranchProtectionDefaults
func v1Alpha1DefaultBranchProtectionDefaultsOptionsPtrToGitlabBranchProtectionDefaults(input *v1alpha1.DefaultBranchProtectionDefaultsOptions) *gitlab.DefaultBranchProtectionDefaultsOptions {
	if input == nil {
		return nil
	}
	result := &gitlab.DefaultBranchProtectionDefaultsOptions{
		AllowForcePush:          input.AllowForcePush,
		DeveloperCanInitialPush: input.DeveloperCanInitialPush,
	}

	if input.AllowedToMerge != nil {
		allowedToMerge := make([]*gitlab.GroupAccessLevel, len(*input.AllowedToMerge))
		for i, v := range *input.AllowedToMerge {
			accessLevel := gitlab.AccessLevelValue(v)
			allowedToMerge[i] = &gitlab.GroupAccessLevel{
				AccessLevel: &accessLevel,
			}
		}
		result.AllowedToMerge = &allowedToMerge
	}

	if input.AllowedToPush != nil {
		allowedToPush := make([]*gitlab.GroupAccessLevel, len(*input.AllowedToPush))
		for i, v := range *input.AllowedToPush {
			accessLevel := gitlab.AccessLevelValue(v)
			allowedToPush[i] = &gitlab.GroupAccessLevel{
				AccessLevel: &accessLevel,
			}
		}
		result.AllowedToPush = &allowedToPush
	}

	return result
}

// groupAccessSliceToIntSlice converts a slice of gitlab.GroupAccessLevel pointers to a slice of ints
func groupAccessSliceToIntSlice(input []*gitlab.GroupAccessLevel) []int {
	result := make([]int, len(input))
	for i, v := range input {
		if v.AccessLevel != nil {
			result[i] = int(*v.AccessLevel)
		}
	}
	return result
}
