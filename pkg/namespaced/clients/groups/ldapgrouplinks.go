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

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
)

const (
	errorLdapGroupLinkNotFound = "404 Not Found"
)

// LdapGroupLinkClient defines Gitlab LDAP Group Link Service Operations
type LdapGroupLinkClient interface {
	ListGroupLDAPLinks(gid interface{}, options ...gitlab.RequestOptionFunc) ([]*gitlab.LDAPGroupLink, *gitlab.Response, error)
	AddGroupLDAPLink(gid interface{}, opt *gitlab.AddGroupLDAPLinkOptions, options ...gitlab.RequestOptionFunc) (*gitlab.LDAPGroupLink, *gitlab.Response, error)
	DeleteGroupLDAPLinkWithCNOrFilter(gid interface{}, opts *gitlab.DeleteGroupLDAPLinkWithCNOrFilterOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// IsErrorLdapGroupLinkNotFound helper function to test for errLdapGroupLinkNotFound error.
func IsErrorLdapGroupLinkNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errorLdapGroupLinkNotFound)
}

// NewLdapGroupLinkClient returns a new GitLab Group Service
func NewLdapGroupLinkClient(cfg common.Config) LdapGroupLinkClient {
	git := common.NewClient(cfg)
	return git.Groups
}

// GenerateAddLdapGroupLinkOptions is used to produce Options for LdapGroupLink creation
func GenerateAddLdapGroupLinkOptions(p *v1alpha1.LdapGroupLinkParameters) *gitlab.AddGroupLDAPLinkOptions {
	ldapGroupLink := &gitlab.AddGroupLDAPLinkOptions{
		CN:          &p.CN,
		GroupAccess: (*gitlab.AccessLevelValue)(&p.GroupAccess),
		Provider:    &p.LdapProvider,
	}

	return ldapGroupLink
}

// GenerateAddLdapGroupLinkObservation is used to produce v1alpha1.LdapGroupLinkbObservation
func GenerateAddLdapGroupLinkObservation(ldapGroupLink *gitlab.LDAPGroupLink) v1alpha1.LdapGroupLinkObservation {
	if ldapGroupLink == nil {
		return v1alpha1.LdapGroupLinkObservation{}
	}

	output := v1alpha1.LdapGroupLinkObservation{
		CN: ldapGroupLink.CN,
	}

	return output
}

// GenerateDeleteGroupLDAPLinkWithCNOrFilterOptions is used to produce Options for LdapGroupLink deletion
func GenerateDeleteGroupLDAPLinkWithCNOrFilterOptions(p *v1alpha1.LdapGroupLinkParameters) *gitlab.DeleteGroupLDAPLinkWithCNOrFilterOptions {
	ldapGroupLink := &gitlab.DeleteGroupLDAPLinkWithCNOrFilterOptions{
		CN:       &p.CN,
		Provider: &p.LdapProvider,
	}

	return ldapGroupLink
}
