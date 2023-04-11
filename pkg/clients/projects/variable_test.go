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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
)

var (
	variableKey       = "A"
	variableValue     = "B"
	variableType      = gitlab.EnvVariableType
	variableMasked    = true
	variableProtected = false
	variableEnvScope  = "blah/*"
	variableRaw       = false
)

var (
	variableTypeLocal = v1alpha1.VariableType(variableType)
)

func TestVariableToParameters(t *testing.T) {
	type args struct {
		ph gitlab.ProjectVariable
	}

	cases := map[string]struct {
		args args
		want v1alpha1.VariableParameters
	}{
		"Full": {
			args: args{
				ph: gitlab.ProjectVariable{
					Key:              variableKey,
					Value:            variableValue,
					VariableType:     variableType,
					Masked:           variableMasked,
					Protected:        variableProtected,
					EnvironmentScope: variableEnvScope,
					Raw:              variableRaw,
				},
			},
			want: v1alpha1.VariableParameters{
				Key:              variableKey,
				Value:            variableValue,
				VariableType:     &variableTypeLocal,
				Masked:           &variableMasked,
				Protected:        &variableProtected,
				EnvironmentScope: &variableEnvScope,
				Raw:              &variableRaw,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := VariableToParameters(tc.args.ph)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
func TestLateInitializeVariable(t *testing.T) {
	cases := map[string]struct {
		parameters *v1alpha1.VariableParameters
		variable   *gitlab.ProjectVariable
		want       *v1alpha1.VariableParameters
	}{
		"AllOptionalFields": {
			parameters: &v1alpha1.VariableParameters{},
			variable: &gitlab.ProjectVariable{
				VariableType:     variableType,
				Protected:        variableProtected,
				Masked:           variableMasked,
				EnvironmentScope: variableEnvScope,
				Raw:              variableRaw,
			},
			want: &v1alpha1.VariableParameters{
				VariableType:     &variableTypeLocal,
				Protected:        &variableProtected,
				Masked:           &variableMasked,
				EnvironmentScope: &variableEnvScope,
				Raw:              &variableRaw,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeVariable(tc.parameters, tc.variable)
			if diff := cmp.Diff(tc.want, tc.parameters); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateCreateVariableOptions(t *testing.T) {
	type args struct {
		parameters *v1alpha1.VariableParameters
	}
	cases := map[string]struct {
		args args
		want *gitlab.CreateProjectVariableOptions
	}{
		"AllFields": {
			args: args{
				parameters: &v1alpha1.VariableParameters{
					Key:              variableKey,
					Value:            variableValue,
					VariableType:     &variableTypeLocal,
					Masked:           &variableMasked,
					Protected:        &variableProtected,
					EnvironmentScope: &variableEnvScope,
					Raw:              &variableRaw,
				},
			},
			want: &gitlab.CreateProjectVariableOptions{
				Key:              &variableKey,
				Value:            &variableValue,
				VariableType:     &variableType,
				Protected:        &variableProtected,
				Masked:           &variableMasked,
				EnvironmentScope: &variableEnvScope,
				Raw:              &variableRaw,
			},
		},
		"SomeFields": {
			args: args{
				parameters: &v1alpha1.VariableParameters{
					Key:          variableKey,
					Value:        variableValue,
					VariableType: &variableTypeLocal,
				},
			},
			want: &gitlab.CreateProjectVariableOptions{
				Key:          &variableKey,
				Value:        &variableValue,
				VariableType: &variableType,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateCreateVariableOptions(tc.args.parameters)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateUpdateVariableOptions(t *testing.T) {
	type args struct {
		parameters *v1alpha1.VariableParameters
	}
	cases := map[string]struct {
		args args
		want *gitlab.UpdateProjectVariableOptions
	}{
		"AllFields": {
			args: args{
				parameters: &v1alpha1.VariableParameters{
					Value:            variableValue,
					VariableType:     &variableTypeLocal,
					Masked:           &variableMasked,
					Protected:        &variableProtected,
					EnvironmentScope: &variableEnvScope,
					Raw:              &variableRaw,
				},
			},
			want: &gitlab.UpdateProjectVariableOptions{
				Value:            &variableValue,
				VariableType:     &variableType,
				Protected:        &variableProtected,
				Masked:           &variableMasked,
				EnvironmentScope: &variableEnvScope,
				Raw:              &variableRaw,
				Filter:           &gitlab.VariableFilter{EnvironmentScope: variableEnvScope},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateUpdateVariableOptions(tc.args.parameters)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsVariableUpToDate(t *testing.T) {
	type args struct {
		variable *gitlab.ProjectVariable
		p        *v1alpha1.VariableParameters
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				p: &v1alpha1.VariableParameters{
					Key:              variableKey,
					Value:            variableValue,
					VariableType:     &variableTypeLocal,
					Protected:        &variableProtected,
					Masked:           &variableMasked,
					EnvironmentScope: &variableEnvScope,
					Raw:              &variableRaw,
				},
				variable: &gitlab.ProjectVariable{
					Key:              variableKey,
					Value:            variableValue,
					VariableType:     variableType,
					Masked:           variableMasked,
					Protected:        variableProtected,
					EnvironmentScope: variableEnvScope,
					Raw:              variableRaw,
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				p: &v1alpha1.VariableParameters{
					Key:              variableKey,
					Value:            variableValue,
					VariableType:     &variableTypeLocal,
					Protected:        &variableProtected,
					Masked:           &variableMasked,
					EnvironmentScope: &variableEnvScope,
					Raw:              &variableRaw,
				},
				variable: &gitlab.ProjectVariable{
					Key:              variableKey,
					Value:            "RANDOM VALUE",
					VariableType:     variableType,
					Masked:           variableMasked,
					Protected:        variableProtected,
					EnvironmentScope: variableEnvScope,
					Raw:              variableRaw,
				},
			},
			want: false,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsVariableUpToDate(tc.args.p, tc.args.variable)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}

}

func TestGenerateGetVariableOptions(t *testing.T) {
	type args struct {
		p *v1alpha1.VariableParameters
	}
	tests := map[string]struct {
		args args
		want *gitlab.GetProjectVariableOptions
	}{
		"Scope": {
			args: args{
				p: &v1alpha1.VariableParameters{
					EnvironmentScope: &variableEnvScope,
				},
			},
			want: &gitlab.GetProjectVariableOptions{
				Filter: &gitlab.VariableFilter{
					EnvironmentScope: variableEnvScope,
				},
			},
		},
		"NoScope": {
			args: args{
				p: &v1alpha1.VariableParameters{},
			},
			want: nil,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := GenerateGetVariableOptions(tc.args.p)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateRemoveVariableOptions(t *testing.T) {
	type args struct {
		p *v1alpha1.VariableParameters
	}
	tests := map[string]struct {
		args args
		want *gitlab.RemoveProjectVariableOptions
	}{
		"Scope": {
			args: args{
				p: &v1alpha1.VariableParameters{
					EnvironmentScope: &variableEnvScope,
				},
			},
			want: &gitlab.RemoveProjectVariableOptions{
				Filter: &gitlab.VariableFilter{
					EnvironmentScope: variableEnvScope,
				},
			},
		},
		"NoScope": {
			args: args{
				p: &v1alpha1.VariableParameters{},
			},
			want: nil,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := GenerateRemoveVariableOptions(tc.args.p)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
