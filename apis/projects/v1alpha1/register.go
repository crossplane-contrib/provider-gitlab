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

package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "projects.gitlab.crossplane.io"
	Version = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// Project type metadata
var (
	ProjectKind             = reflect.TypeOf(Project{}).Name()
	ProjectGroupKind        = schema.GroupKind{Group: Group, Kind: ProjectKind}.String()
	ProjectKindAPIVersion   = ProjectKind + "." + SchemeGroupVersion.String()
	ProjectGroupVersionKind = SchemeGroupVersion.WithKind(ProjectKind)
)

// Hook type metadata
var (
	HookKind             = reflect.TypeOf(Hook{}).Name()
	HookGroupKind        = schema.GroupKind{Group: Group, Kind: HookKind}.String()
	HookKindAPIVersion   = HookKind + "." + SchemeGroupVersion.String()
	HookGroupVersionKind = SchemeGroupVersion.WithKind(HookKind)
)

// Member type metadata
var (
	MemberKind             = reflect.TypeOf(Member{}).Name()
	MemberGroupKind        = schema.GroupKind{Group: Group, Kind: MemberKind}.String()
	MemberKindAPIVersion   = MemberKind + "." + SchemeGroupVersion.String()
	MemberGroupVersionKind = SchemeGroupVersion.WithKind(MemberKind)
)

// Deploy Token type metadata
var (
	DeployTokenKind             = reflect.TypeOf(DeployToken{}).Name()
	DeployTokenGroupKind        = schema.GroupKind{Group: Group, Kind: DeployTokenKind}.String()
	DeployTokenKindAPIVersion   = DeployTokenKind + "." + SchemeGroupVersion.String()
	DeployTokenGroupVersionKind = SchemeGroupVersion.WithKind(DeployTokenKind)
)

// Access Token type metadata
var (
	AccessTokenKind             = reflect.TypeOf(AccessToken{}).Name()
	AccessTokenGroupKind        = schema.GroupKind{Group: Group, Kind: AccessTokenKind}.String()
	AccessTokenKindAPIVersion   = AccessTokenKind + "." + SchemeGroupVersion.String()
	AccessTokenGroupVersionKind = SchemeGroupVersion.WithKind(AccessTokenKind)
)

// Variable type metadata
var (
	VariableKind             = reflect.TypeOf(Variable{}).Name()
	VariableGroupKind        = schema.GroupKind{Group: Group, Kind: VariableKind}.String()
	VariableKindAPIVersion   = VariableKind + "." + SchemeGroupVersion.String()
	VariableGroupVersionKind = SchemeGroupVersion.WithKind(VariableKind)
)

// Deploy Key type metadata
var (
	DeployKeyKind             = reflect.TypeOf(DeployKey{}).Name()
	DeployKeyGroupKind        = schema.GroupKind{Group: Group, Kind: DeployKeyKind}.String()
	DeployKeyKindAPIVersion   = DeployKeyKind + "." + SchemeGroupVersion.String()
	DeployKeyGroupVersionKind = SchemeGroupVersion.WithKind(DeployKeyKind)
)

// Pipeline Sharing type metadata
var (
	PipelineScheduleKind             = reflect.TypeOf(PipelineSchedule{}).Name()
	PipelineScheduleGroupKind        = schema.GroupKind{Group: Group, Kind: PipelineScheduleKind}.String()
	PipelineScheduleKindAPIVersion   = PipelineScheduleKind + "." + SchemeGroupVersion.String()
	PipelineScheduleGroupVersionKind = SchemeGroupVersion.WithKind(PipelineScheduleKind)
)

func init() {
	SchemeBuilder.Register(&Project{}, &ProjectList{})
	SchemeBuilder.Register(&Hook{}, &HookList{})
	SchemeBuilder.Register(&Member{}, &MemberList{})
	SchemeBuilder.Register(&DeployToken{}, &DeployTokenList{})
	SchemeBuilder.Register(&Variable{}, &VariableList{})
	SchemeBuilder.Register(&DeployKey{}, &DeployKeyList{})
	SchemeBuilder.Register(&AccessToken{}, &AccessTokenList{})
	SchemeBuilder.Register(&PipelineSchedule{}, &PipelineScheduleList{})
}
