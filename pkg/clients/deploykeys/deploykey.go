/*
Copyright 2020 The Crossplane Authors.

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

package deploykeys

import (
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-gitlab/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
)

const (
	errDeployKeyNotFound = "404 Not found"
)

// Client defines Gitlab DeployKey service operations
type Client interface {
	GetDeployKey(pid interface{}, deployKeyID int, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error)
	AddDeployKey(pid interface{}, opt *gitlab.AddDeployKeyOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error)
	UpdateDeployKey(pid interface{}, deployKeyID int, opt *gitlab.UpdateDeployKeyOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DeployKey, *gitlab.Response, error)
	DeleteDeployKey(pid interface{}, deployKeyID int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewDeployKeyClient returns a new Gitlab Project service
func NewDeployKeyClient(cfg clients.Config) Client {
	git := clients.NewClient(cfg)
	return git.DeployKeys
}

// IsErrorDeployKeyNotFound helper function to test for errProjectNotFound error.
func IsErrorDeployKeyNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errDeployKeyNotFound)
}

// LateInitializeDeployKey fills the empty fields in the deploykey spec with the
// values seen in gitlab.DeployKey.
func LateInitializeDeployKey(in *v1alpha1.DeployKeyParameters, deploykey *gitlab.DeployKey) { // nolint:gocyclo
	if deploykey == nil {
		return
	}

	if in.Key == nil {
		in.Key = &deploykey.Key
	}

	if in.CanPush == nil {
		in.CanPush = deploykey.CanPush
	}
}

// GenerateDeployKeyObservation is used to produce v1alpha1.DeployKeyObservation from
// gitlab.DeployKey.
func GenerateDeployKeyObservation(deploykey *gitlab.DeployKey) v1alpha1.DeployKeyObservation { // nolint:gocyclo
	if deploykey == nil {
		return v1alpha1.DeployKeyObservation{}
	}

	o := v1alpha1.DeployKeyObservation{
		ID: deploykey.ID,
	}

	if deploykey.CreatedAt != nil {
		o.CreatedAt = &metav1.Time{Time: *deploykey.CreatedAt}
	}
	return o
}

// GenerateCreateDeployKeyOptions generates deploy key creation options
func GenerateCreateDeployKeyOptions(p *v1alpha1.DeployKeyParameters) *gitlab.AddDeployKeyOptions {

	deploykey := &gitlab.AddDeployKeyOptions{
		Title:   p.Title,
		Key:     p.Key,
		CanPush: p.CanPush,
	}

	return deploykey
}

// GenerateUpdateDeployKeyOptions generates deploy key edit options
func GenerateUpdateDeployKeyOptions(p *v1alpha1.DeployKeyParameters) *gitlab.UpdateDeployKeyOptions {
	o := &gitlab.UpdateDeployKeyOptions{
		Title:   p.Title,
		CanPush: p.CanPush,
	}

	return o
}

// IsDeployKeyUpToDate checks whether there is a change in any of the modifiable fields.
func IsDeployKeyUpToDate(p *v1alpha1.DeployKeyParameters, g *gitlab.DeployKey) bool { // nolint:gocyclo
	if !cmp.Equal(p.Title, clients.StringToPtr(g.Title)) {
		return false
	}

	if !cmp.Equal(p.CanPush, g.CanPush) {
		return false
	}

	return true
}
