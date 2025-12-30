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

package groups

import (
	gitlab "gitlab.com/gitlab-org/api/client-go"

	commonv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
)

// VariableClient defines Gitlab Variable service operations
type VariableClient interface {
	ListVariables(gid any, opt *gitlab.ListGroupVariablesOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.GroupVariable, *gitlab.Response, error)
	GetVariable(gid any, key string, opt *gitlab.GetGroupVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error)
	CreateVariable(gid any, opt *gitlab.CreateGroupVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error)
	UpdateVariable(gid any, key string, opt *gitlab.UpdateGroupVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error)
	RemoveVariable(gid any, key string, opt *gitlab.RemoveGroupVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewVariableClient returns a new Gitlab Group service
func NewVariableClient(cfg common.Config) VariableClient {
	git := common.NewClient(cfg)
	return git.GroupVariables
}

// GenerateVariableObservation creates VariableObservation from gitlab InstanceVariable
func GenerateVariableObservation(variable *gitlab.GroupVariable) v1alpha1.VariableObservation {
	return v1alpha1.VariableObservation{
		CommonVariableObservation: commonv1alpha1.CommonVariableObservation{
			Key:          variable.Key,
			Description:  variable.Description,
			VariableType: commonv1alpha1.VariableType(variable.VariableType),
			Protected:    variable.Protected,
			Masked:       variable.Masked,
			Raw:          variable.Raw,
		},
		EnvironmentScope: variable.EnvironmentScope,
		Hidden:           variable.Hidden,
	}
}

// LateInitializeVariable fills the empty fields in the groupVariable spec with the
// values seen in gitlab.Variable.
func LateInitializeVariable(in *v1alpha1.VariableParameters, variable *gitlab.GroupVariable) {
	if variable == nil {
		return
	}

	if in.VariableType == nil {
		in.VariableType = (*commonv1alpha1.VariableType)(&variable.VariableType)
	}

	if in.Description == nil {
		in.Description = &variable.Description
	}

	if in.Protected == nil {
		in.Protected = &variable.Protected
	}

	if in.Masked == nil {
		in.Masked = &variable.Masked
	}

	if in.EnvironmentScope == nil {
		in.EnvironmentScope = &variable.EnvironmentScope
	}

	if in.Raw == nil {
		in.Raw = &variable.Raw
	}
}

// GenerateCreateVariableOptions generates group creation options
func GenerateCreateVariableOptions(p *v1alpha1.VariableParameters) *gitlab.CreateGroupVariableOptions {
	variable := &gitlab.CreateGroupVariableOptions{
		Key:              &p.Key,
		Value:            p.Value,
		Description:      p.Description,
		VariableType:     (*gitlab.VariableTypeValue)(p.VariableType),
		Protected:        p.Protected,
		Masked:           p.Masked,
		EnvironmentScope: p.EnvironmentScope,
		Raw:              p.Raw,
	}
	return variable
}

// GenerateUpdateVariableOptions generates group update options
func GenerateUpdateVariableOptions(p *v1alpha1.VariableParameters) *gitlab.UpdateGroupVariableOptions {
	variable := &gitlab.UpdateGroupVariableOptions{
		Value:            p.Value,
		Description:      p.Description,
		VariableType:     (*gitlab.VariableTypeValue)(p.VariableType),
		Protected:        p.Protected,
		Masked:           p.Masked,
		EnvironmentScope: p.EnvironmentScope,
		Raw:              p.Raw,
	}
	return variable
}

// GenerateVariableFilter generates a variable filter that matches the variable parameters' environment scope.
func GenerateVariableFilter(p *v1alpha1.VariableParameters) *gitlab.VariableFilter {
	if p.EnvironmentScope == nil {
		return nil
	}

	return &gitlab.VariableFilter{
		EnvironmentScope: *p.EnvironmentScope,
	}
}

// GenerateGetVariableOptions generates group get options
func GenerateGetVariableOptions(p *v1alpha1.VariableParameters) *gitlab.GetGroupVariableOptions {
	variable := &gitlab.GetGroupVariableOptions{
		Filter: GenerateVariableFilter(p),
	}
	return variable
}

// IsVariableUpToDate checks whether there is a change in any of the modifiable fields.
func IsVariableUpToDate(p *v1alpha1.VariableParameters, g *gitlab.GroupVariable) bool { //nolint:gocyclo
	if p == nil {
		return true
	}
	if g == nil {
		return false
	}

	if p.Key != g.Key {
		return false
	}

	if !clients.IsComparableEqualToComparablePtr(p.Value, g.Value) {
		return false
	}

	if !clients.IsComparableEqualToComparablePtr(p.Description, g.Description) {
		return false
	}

	if !clients.IsComparableEqualToComparablePtr((*string)(p.VariableType), (string)(g.VariableType)) {
		return false
	}

	if !clients.IsComparableEqualToComparablePtr(p.Protected, g.Protected) {
		return false
	}

	if !clients.IsComparableEqualToComparablePtr(p.Masked, g.Masked) {
		return false
	}

	if !clients.IsComparableEqualToComparablePtr(p.Raw, g.Raw) {
		return false
	}

	if !clients.IsComparableEqualToComparablePtr(p.EnvironmentScope, g.EnvironmentScope) {
		return false
	}

	return true
}
