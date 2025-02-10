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

	"github.com/xanzy/go-gitlab"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
)

const (
	errorSamlGroupLinkNotFound = "Linked SAML group link not found"
)

// SamlGroupLinkClient defines Gitlab Saml Group Link Service Operations
type SamlGroupLinkClient interface {
	GetGroupSAMLLink(gid interface{}, samlGroupName string, options ...gitlab.RequestOptionFunc) (*gitlab.SAMLGroupLink, *gitlab.Response, error)
	AddGroupSAMLLink(gid interface{}, opt *gitlab.AddGroupSAMLLinkOptions, options ...gitlab.RequestOptionFunc) (*gitlab.SAMLGroupLink, *gitlab.Response, error)
	DeleteGroupSAMLLink(gid interface{}, samlGroupName string, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// IsErrorSamlGroupLinkNotFound helper function to test for errSamlGroupLinkNotFound error.
func IsErrorSamlGroupLinkNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errorSamlGroupLinkNotFound)
}

// NewSamlGroupLinkClient returns a new Giltab Group Service
func NewSamlGroupLinkClient(cfg clients.Config) SamlGroupLinkClient {
	git := clients.NewClient(cfg)
	return git.Groups
}

// GenerateAddSamlGroupLinkOptions is used to produce Options for SamlGroupLink creation
func GenerateAddSamlGroupLinkOptions(p *v1alpha1.SamlGroupLinkParameters) *gitlab.AddGroupSAMLLinkOptions {
	samlGroupName := &gitlab.AddGroupSAMLLinkOptions{
		SAMLGroupName: p.Name,
		AccessLevel:   (*gitlab.AccessLevelValue)(&p.AccessLevel),
		MemberRoleID:  p.MemberRoleID,
	}

	return samlGroupName
}

// GenerateAddSamlGroupLinkObservation is used to produce v1alpha1.SamlGroupLinkbObservation
func GenerateAddSamlGroupLinkObservation(samlGroupLink *gitlab.SAMLGroupLink) v1alpha1.SamlGroupLinkObservation {
	if samlGroupLink == nil {
		return v1alpha1.SamlGroupLinkObservation{}
	}

	output := v1alpha1.SamlGroupLinkObservation{
		Name: samlGroupLink.Name,
	}

	return output
}
