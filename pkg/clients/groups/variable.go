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
	"strings"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/xanzy/go-gitlab"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
)

const (
	errVariableNotFound = "404 Variable Not Found"
)

// VariableClient defines Gitlab Variable service operations
type VariableClient interface {
	ListVariables(gid interface{}, opt *gitlab.ListGroupVariablesOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.GroupVariable, *gitlab.Response, error)
	GetVariable(gid interface{}, key string, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error)
	CreateVariable(gid interface{}, opt *gitlab.CreateGroupVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error)
	UpdateVariable(gid interface{}, key string, opt *gitlab.UpdateGroupVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupVariable, *gitlab.Response, error)
	RemoveVariable(gid interface{}, key string, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewVariableClient returns a new Gitlab Group service
func NewVariableClient(cfg clients.Config) VariableClient {
	git := clients.NewClient(cfg)
	return git.GroupVariables
}

// IsErrorVariableNotFound helper function to test for errGroupNotFound error.
func IsErrorVariableNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errVariableNotFound)
}

// LateInitializeVariable fills the empty fields in the groupVariable spec with the
// values seen in gitlab.Variable.
func LateInitializeVariable(in *v1alpha1.VariableParameters, variable *gitlab.GroupVariable) { // nolint:gocyclo
	if variable == nil {
		return
	}

	if in.VariableType == nil {
		in.VariableType = (*v1alpha1.VariableType)(&variable.VariableType)
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

// VariableToParameters coonverts a GitLab API representation of a
// Group Variable back into our local VariableParameters format
func VariableToParameters(in gitlab.GroupVariable) v1alpha1.VariableParameters {
	return v1alpha1.VariableParameters{
		Key:              in.Key,
		Value:            in.Value,
		VariableType:     (*v1alpha1.VariableType)(&in.VariableType),
		Protected:        &in.Protected,
		Masked:           &in.Masked,
		EnvironmentScope: &in.EnvironmentScope,
		Raw:              &in.Raw,
	}
}

// GenerateCreateVariableOptions generates group creation options
func GenerateCreateVariableOptions(p *v1alpha1.VariableParameters) *gitlab.CreateGroupVariableOptions {
	variable := &gitlab.CreateGroupVariableOptions{
		Key:              &p.Key,
		Value:            &p.Value,
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
		Value:            &p.Value,
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

// IsVariableUpToDate checks whether there is a change in any of the modifiable fields.
func IsVariableUpToDate(p *v1alpha1.VariableParameters, g *gitlab.GroupVariable) bool {
	if p == nil {
		return true
	}

	return cmp.Equal(*p,
		VariableToParameters(*g),
		cmpopts.EquateEmpty(),
		cmpopts.IgnoreTypes(&xpv1.Reference{}, &xpv1.Selector{}, []xpv1.Reference{}, &xpv1.SecretKeySelector{}),
		cmpopts.IgnoreFields(v1alpha1.VariableParameters{}, "GroupID"),
	)
}
