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
	KubernetesGroup = "groups.gitlab.crossplane.io"
	Version         = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: KubernetesGroup, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// GroupGitLab type metadata
var (
	GroupKind                       = reflect.TypeOf(Group{}).Name()
	GroupKubernetesGroupKind        = schema.GroupKind{Group: KubernetesGroup, Kind: GroupKind}.String()
	GroupKindAPIVersion             = GroupKind + "." + SchemeGroupVersion.String()
	GroupKubernetesGroupVersionKind = SchemeGroupVersion.WithKind(GroupKind)
)

// MemberGitLab type metadata
var (
	MemberKind                       = reflect.TypeOf(Member{}).Name()
	MemberKubernetesGroupKind        = schema.GroupKind{Group: KubernetesGroup, Kind: MemberKind}.String()
	MemberKindAPIVersion             = MemberKind + "." + SchemeGroupVersion.String()
	MemberKubernetesGroupVersionKind = SchemeGroupVersion.WithKind(MemberKind)
)

// Deploy Token type metadata
var (
	DeployTokenKind             = reflect.TypeOf(DeployToken{}).Name()
	DeployTokenGroupKind        = schema.GroupKind{Group: KubernetesGroup, Kind: DeployTokenKind}.String()
	DeployTokenKindAPIVersion   = DeployTokenKind + "." + SchemeGroupVersion.String()
	DeployTokenGroupVersionKind = SchemeGroupVersion.WithKind(DeployTokenKind)
)

// Access Token type metadata
var (
	AccessTokenKind             = reflect.TypeOf(AccessToken{}).Name()
	AccessTokenGroupKind        = schema.GroupKind{Group: KubernetesGroup, Kind: AccessTokenKind}.String()
	AccessTokenKindAPIVersion   = AccessTokenKind + "." + SchemeGroupVersion.String()
	AccessTokenGroupVersionKind = SchemeGroupVersion.WithKind(AccessTokenKind)
)

// Variable type metadata
var (
	VariableKind             = reflect.TypeOf(Variable{}).Name()
	VariableGroupKind        = schema.GroupKind{Group: KubernetesGroup, Kind: VariableKind}.String()
	VariableKindAPIVersion   = VariableKind + "." + SchemeGroupVersion.String()
	VariableGroupVersionKind = SchemeGroupVersion.WithKind(VariableKind)
)

func init() {
	SchemeBuilder.Register(&Group{}, &GroupList{})
	SchemeBuilder.Register(&Member{}, &MemberList{})
	SchemeBuilder.Register(&AccessToken{}, &AccessTokenList{})
	SchemeBuilder.Register(&DeployToken{}, &DeployTokenList{})
	SchemeBuilder.Register(&Variable{}, &VariableList{})
}
