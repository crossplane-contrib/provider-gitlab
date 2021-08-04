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

// ProjectHook type metadata
var (
	ProjectHookKind             = reflect.TypeOf(ProjectHook{}).Name()
	ProjectHookGroupKind        = schema.GroupKind{Group: Group, Kind: ProjectHookKind}.String()
	ProjectHookKindAPIVersion   = ProjectHookKind + "." + SchemeGroupVersion.String()
	ProjectHookGroupVersionKind = SchemeGroupVersion.WithKind(ProjectHookKind)
)

// ProjectMember type metadata
var (
	ProjectMemberKind             = reflect.TypeOf(ProjectMember{}).Name()
	ProjectMemberGroupKind        = schema.GroupKind{Group: Group, Kind: ProjectMemberKind}.String()
	ProjectMemberKindAPIVersion   = ProjectMemberKind + "." + SchemeGroupVersion.String()
	ProjectMemberGroupVersionKind = SchemeGroupVersion.WithKind(ProjectMemberKind)
)

// Project Deploy Token type metadata
var (
	ProjectDeployTokenKind             = reflect.TypeOf(ProjectDeployToken{}).Name()
	ProjectDeployTokenGroupKind        = schema.GroupKind{Group: Group, Kind: ProjectDeployTokenKind}.String()
	ProjectDeployTokenKindAPIVersion   = ProjectDeployTokenKind + "." + SchemeGroupVersion.String()
	ProjectDeployTokenGroupVersionKind = SchemeGroupVersion.WithKind(ProjectDeployTokenKind)
)

func init() {
	SchemeBuilder.Register(&Project{}, &ProjectList{})
	SchemeBuilder.Register(&ProjectHook{}, &ProjectHookList{})
	SchemeBuilder.Register(&ProjectMember{}, &ProjectMemberList{})
	SchemeBuilder.Register(&ProjectDeployToken{}, &ProjectDeployTokenList{})
}
