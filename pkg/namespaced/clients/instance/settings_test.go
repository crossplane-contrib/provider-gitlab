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

package instance

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/instance/v1alpha1"
)

func TestIsDefaultBranchProtectionDefaultsPtrEqualToDefaultsPtr(t *testing.T) {
	type args struct {
		dbpdv *v1alpha1.DefaultBranchProtectionDefaultsOptions
		dbpdg *gitlab.BranchProtectionDefaults
	}
	cases := map[string]struct {
		args args
		want bool
	}{
		"BothNil": {
			args: args{
				dbpdv: nil,
				dbpdg: nil,
			},
			want: true,
		},
		"SpecNil": {
			args: args{
				dbpdv: nil,
				dbpdg: &gitlab.BranchProtectionDefaults{},
			},
			want: true,
		},
		"GitlabNil": {
			args: args{
				dbpdv: &v1alpha1.DefaultBranchProtectionDefaultsOptions{},
				dbpdg: nil,
			},
			want: false,
		},
		"Equal": {
			args: args{
				dbpdv: &v1alpha1.DefaultBranchProtectionDefaultsOptions{
					AllowForcePush:          boolPtr(true),
					DeveloperCanInitialPush: boolPtr(false),
					AllowedToMerge:          &[]int{30, 40},
					AllowedToPush:           &[]int{30},
				},
				dbpdg: &gitlab.BranchProtectionDefaults{
					AllowForcePush:          true,
					DeveloperCanInitialPush: false,
					AllowedToMerge:          []*gitlab.GroupAccessLevel{{AccessLevel: accessLevelPtr(30)}, {AccessLevel: accessLevelPtr(40)}},
					AllowedToPush:           []*gitlab.GroupAccessLevel{{AccessLevel: accessLevelPtr(30)}},
				},
			},
			want: true,
		},
		"NotEqualAllowForcePush": {
			args: args{
				dbpdv: &v1alpha1.DefaultBranchProtectionDefaultsOptions{
					AllowForcePush: boolPtr(true),
				},
				dbpdg: &gitlab.BranchProtectionDefaults{
					AllowForcePush: false,
				},
			},
			want: false,
		},
		"NotEqualDeveloperCanInitialPush": {
			args: args{
				dbpdv: &v1alpha1.DefaultBranchProtectionDefaultsOptions{
					DeveloperCanInitialPush: boolPtr(true),
				},
				dbpdg: &gitlab.BranchProtectionDefaults{
					DeveloperCanInitialPush: false,
				},
			},
			want: false,
		},
		"NotEqualAllowedToMerge": {
			args: args{
				dbpdv: &v1alpha1.DefaultBranchProtectionDefaultsOptions{
					AllowedToMerge: &[]int{30},
				},
				dbpdg: &gitlab.BranchProtectionDefaults{
					AllowedToMerge: []*gitlab.GroupAccessLevel{{AccessLevel: accessLevelPtr(40)}},
				},
			},
			want: false,
		},
		"NotEqualAllowedToPush": {
			args: args{
				dbpdv: &v1alpha1.DefaultBranchProtectionDefaultsOptions{
					AllowedToPush: &[]int{30},
				},
				dbpdg: &gitlab.BranchProtectionDefaults{
					AllowedToPush: []*gitlab.GroupAccessLevel{{AccessLevel: accessLevelPtr(40)}},
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := isDefaultBranchProtectionDefaultsPtrEqualToDefaultsPtr(tc.args.dbpdv, tc.args.dbpdg)
			if got != tc.want {
				t.Errorf("isDefaultBranchProtectionDefaultsPtrEqualToDefaultsPtr() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGitlabBranchProtectionDefaultsTov1Alpha1BranchProtectionDefaults(t *testing.T) {
	type args struct {
		input *gitlab.BranchProtectionDefaults
	}
	cases := map[string]struct {
		args args
		want v1alpha1.BranchProtectionDefaults
	}{
		"NilInput": {
			args: args{
				input: nil,
			},
			want: v1alpha1.BranchProtectionDefaults{},
		},
		"FullInput": {
			args: args{
				input: &gitlab.BranchProtectionDefaults{
					AllowForcePush:          true,
					DeveloperCanInitialPush: false,
					AllowedToMerge:          []*gitlab.GroupAccessLevel{{AccessLevel: accessLevelPtr(30)}},
					AllowedToPush:           []*gitlab.GroupAccessLevel{{AccessLevel: accessLevelPtr(40)}},
				},
			},
			want: v1alpha1.BranchProtectionDefaults{
				AllowForcePush:          true,
				DeveloperCanInitialPush: false,
				AllowedToMerge:          []*int{intPtr(30)},
				AllowedToPush:           []*int{intPtr(40)},
			},
		},
		"InputWithNilAccessLevel": {
			args: args{
				input: &gitlab.BranchProtectionDefaults{
					AllowedToMerge: []*gitlab.GroupAccessLevel{{AccessLevel: nil}},
				},
			},
			want: v1alpha1.BranchProtectionDefaults{
				AllowedToMerge: []*int{nil},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := gitlabBranchProtectionDefaultsTov1Alpha1BranchProtectionDefaults(tc.args.input)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("gitlabBranchProtectionDefaultsTov1Alpha1BranchProtectionDefaults() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGitlabBranchProtectionDefaultsTov1Alpha1BranchProtectionDefaultsOptions(t *testing.T) {
	type args struct {
		input *gitlab.BranchProtectionDefaults
	}
	cases := map[string]struct {
		args args
		want v1alpha1.DefaultBranchProtectionDefaultsOptions
	}{
		"NilInput": {
			args: args{
				input: nil,
			},
			want: v1alpha1.DefaultBranchProtectionDefaultsOptions{},
		},
		"FullInput": {
			args: args{
				input: &gitlab.BranchProtectionDefaults{
					AllowForcePush:          true,
					DeveloperCanInitialPush: false,
					AllowedToMerge:          []*gitlab.GroupAccessLevel{{AccessLevel: accessLevelPtr(30)}},
					AllowedToPush:           []*gitlab.GroupAccessLevel{{AccessLevel: accessLevelPtr(40)}},
				},
			},
			want: v1alpha1.DefaultBranchProtectionDefaultsOptions{
				AllowForcePush:          boolPtr(true),
				DeveloperCanInitialPush: boolPtr(false),
				AllowedToMerge:          &[]int{30},
				AllowedToPush:           &[]int{40},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := gitlabBranchProtectionDefaultsTov1Alpha1BranchProtectionDefaultsOptions(tc.args.input)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("gitlabBranchProtectionDefaultsTov1Alpha1BranchProtectionDefaultsOptions() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestV1Alpha1DefaultBranchProtectionDefaultsOptionsPtrToGitlabBranchProtectionDefaults(t *testing.T) {
	type args struct {
		input *v1alpha1.DefaultBranchProtectionDefaultsOptions
	}
	cases := map[string]struct {
		args args
		want *gitlab.DefaultBranchProtectionDefaultsOptions
	}{
		"NilInput": {
			args: args{
				input: nil,
			},
			want: nil,
		},
		"FullInput": {
			args: args{
				input: &v1alpha1.DefaultBranchProtectionDefaultsOptions{
					AllowForcePush:          boolPtr(true),
					DeveloperCanInitialPush: boolPtr(false),
					AllowedToMerge:          &[]int{30},
					AllowedToPush:           &[]int{40},
				},
			},
			want: &gitlab.DefaultBranchProtectionDefaultsOptions{
				AllowForcePush:          boolPtr(true),
				DeveloperCanInitialPush: boolPtr(false),
				AllowedToMerge:          &[]*gitlab.GroupAccessLevel{{AccessLevel: accessLevelPtr(30)}},
				AllowedToPush:           &[]*gitlab.GroupAccessLevel{{AccessLevel: accessLevelPtr(40)}},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := v1Alpha1DefaultBranchProtectionDefaultsOptionsPtrToGitlabBranchProtectionDefaults(tc.args.input)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("v1Alpha1DefaultBranchProtectionDefaultsOptionsPtrToGitlabBranchProtectionDefaults() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGroupAccessSliceToIntSlice(t *testing.T) {
	type args struct {
		input []*gitlab.GroupAccessLevel
	}
	cases := map[string]struct {
		args args
		want []int
	}{
		"NilInput": {
			args: args{
				input: nil,
			},
			want: []int{},
		},
		"EmptyInput": {
			args: args{
				input: []*gitlab.GroupAccessLevel{},
			},
			want: []int{},
		},
		"FullInput": {
			args: args{
				input: []*gitlab.GroupAccessLevel{
					{AccessLevel: accessLevelPtr(30)},
					{AccessLevel: accessLevelPtr(40)},
				},
			},
			want: []int{30, 40},
		},
		"InputWithNilAccessLevel": {
			args: args{
				input: []*gitlab.GroupAccessLevel{
					{AccessLevel: nil},
				},
			},
			want: []int{0},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := groupAccessSliceToIntSlice(tc.args.input)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("groupAccessSliceToIntSlice() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func boolPtr(b bool) *bool { return &b }
func intPtr(i int) *int    { return &i }

func accessLevelPtr(i int) *gitlab.AccessLevelValue {
	v := gitlab.AccessLevelValue(i)
	return &v
}
