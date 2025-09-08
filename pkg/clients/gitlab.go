/*
Copyright 2019 The Crossplane Authors.

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

package clients

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cleanhttp"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"golang.org/x/oauth2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-gitlab/apis/shared"
	sharedProjectsV1Alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/shared/projects/v1alpha1"
)

// BasicAuth is the expected struct that can be passed in the Config.Token field to add support for BasicAuth AuthMethod
type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Config provides gitlab configurations for the Gitlab client
type Config struct {
	Token              string
	BaseURL            string
	InsecureSkipVerify bool
	AuthMethod         shared.UnifiedAuthType
}

// NewClient creates new Gitlab Client with provided Gitlab Configurations/Credentials.
func NewClient(c Config) *gitlab.Client {
	var cl *gitlab.Client
	var err error
	options := []gitlab.ClientOptionFunc{}
	if c.BaseURL != "" {
		options = append(options, gitlab.WithBaseURL(c.BaseURL))
	}
	if c.InsecureSkipVerify {
		transport := cleanhttp.DefaultPooledTransport()
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{
				MinVersion: tls.VersionTLS12,
			}
		}
		transport.TLSClientConfig.InsecureSkipVerify = true
		httpclient := &http.Client{
			Transport: transport,
		}
		options = append(options, gitlab.WithHTTPClient(httpclient))
	}

	switch c.AuthMethod {
	case shared.UnifiedBasicAuth:
		ba := &BasicAuth{}
		if err = json.Unmarshal([]byte(c.Token), ba); err != nil {
			panic(err)
		}
		cl, err = gitlab.NewBasicAuthClient(ba.Username, ba.Password, options...) //nolint:staticcheck
	case shared.UnifiedJobToken:
		cl, err = gitlab.NewJobClient(c.Token, options...)
	case shared.UnifiedOAuthToken:
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: c.Token})
		cl, err = gitlab.NewAuthSourceClient(gitlab.OAuthTokenSource{TokenSource: ts}, options...)
	case shared.UnifiedPersonalAccessToken:
		cl, err = gitlab.NewClient(c.Token, options...)
	default:
		cl, err = gitlab.NewClient(c.Token, options...)
	}
	if err != nil {
		panic(err)
	}

	return cl
}

// LateInitialize return in if not nil or from.
func LateInitialize[T any](in, from *T) *T {
	if in != nil {
		return in
	}
	return from
}

// LateInitializeFromValue returns in if not nil or a pointer to from.
func LateInitializeFromValue[T any](in *T, from T) *T {
	if in != nil {
		return in
	}
	return &from
}

// LateInitializeFromValue returns from if in is nil and from is not the zero
// value of T. Otherwise it returns in.
func LateInitializeFromValueIfNotZero[T comparable](in *T, from T) *T {
	var zeroValue T
	if in == nil && from != zeroValue {
		return &from
	}
	return in
}

// LateInitializeStringPtr returns `from` if `in` is nil and `from` is non-empty,
// in other cases it returns `in`.
func LateInitializeStringPtr(in *string, from string) *string {
	return LateInitializeFromValueIfNotZero(in, from)
}

// LateInitializeAccessControlValue returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeAccessControlValue(in *sharedProjectsV1Alpha1.AccessControlValue, from gitlab.AccessControlValue) *sharedProjectsV1Alpha1.AccessControlValue {
	if in == nil && from != "" {
		return (*sharedProjectsV1Alpha1.AccessControlValue)(&from)
	}
	return in
}

// LateInitializeVisibilityValue returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeVisibilityValue(in *sharedProjectsV1Alpha1.VisibilityValue, from gitlab.VisibilityValue) *sharedProjectsV1Alpha1.VisibilityValue {
	if in == nil && from != "" {
		return (*sharedProjectsV1Alpha1.VisibilityValue)(&from)
	}
	return in
}

// LateInitializeMergeMethodValue returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeMergeMethodValue(in *sharedProjectsV1Alpha1.MergeMethodValue, from gitlab.MergeMethodValue) *sharedProjectsV1Alpha1.MergeMethodValue {
	if in == nil && from != "" {
		return (*sharedProjectsV1Alpha1.MergeMethodValue)(&from)
	}
	return in
}

// VisibilityValueV1alpha1ToGitlab converts *v1alpha1.VisibilityValue to *gitlab.VisibilityValue
func VisibilityValueV1alpha1ToGitlab(from *sharedProjectsV1Alpha1.VisibilityValue) *gitlab.VisibilityValue {
	return (*gitlab.VisibilityValue)(from)
}

// VisibilityValueStringToGitlab converts string to *gitlab.VisibilityValue
func VisibilityValueStringToGitlab(from string) *gitlab.VisibilityValue {
	return (*gitlab.VisibilityValue)(&from)
}

// AccessControlValueV1alpha1ToGitlab converts *v1alpha1.AccessControlValue to *gitlab.AccessControlValue
func AccessControlValueV1alpha1ToGitlab(from *sharedProjectsV1Alpha1.AccessControlValue) *gitlab.AccessControlValue {
	return (*gitlab.AccessControlValue)(from)
}

// ContainerExpirationPolicyAttributesV1alpha1ToGitlab converts *v1alpha1.ContainerExpirationPolicyAttributes to *gitlab.ContainerExpirationPolicyAttributes
func ContainerExpirationPolicyAttributesV1alpha1ToGitlab(from *sharedProjectsV1Alpha1.ContainerExpirationPolicyAttributes) *gitlab.ContainerExpirationPolicyAttributes {
	return (*gitlab.ContainerExpirationPolicyAttributes)(from)
}

// AccessControlValueStringToGitlab converts string to *gitlab.AccessControlValue
func AccessControlValueStringToGitlab(from string) *gitlab.AccessControlValue {
	return (*gitlab.AccessControlValue)(&from)
}

// MergeMethodV1alpha1ToGitlab converts *v1alpha1.MergeMethodValue to *gitlab.MergeMethodValue
func MergeMethodV1alpha1ToGitlab(from *sharedProjectsV1Alpha1.MergeMethodValue) *gitlab.MergeMethodValue {
	return (*gitlab.MergeMethodValue)(from)
}

// MergeMethodStringToGitlab converts string to *gitlab.MergeMethodValue
func MergeMethodStringToGitlab(from string) *gitlab.MergeMethodValue {
	return (*gitlab.MergeMethodValue)(&from)
}

// StringToPtr converts string to *string
func StringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// IsBoolEqualToBoolPtr compares a *bool with bool
func IsBoolEqualToBoolPtr(bp *bool, b bool) bool {
	if bp != nil {
		if !cmp.Equal(*bp, b) {
			return false
		}
	}
	return true
}

// IsIntEqualToIntPtr compares an *int with int
func IsIntEqualToIntPtr(ip *int, i int) bool {
	if ip != nil {
		if !cmp.Equal(*ip, i) {
			return false
		}
	}
	return true
}

// IsStringEqualToStringPtr compares a *string with string
func IsStringEqualToStringPtr(sp *string, s string) bool {
	if sp != nil {
		if !cmp.Equal(*sp, s) {
			return false
		}
	}
	return true
}

// IsResponseNotFound returns true of Gitlab Response indicates CR was not found
func IsResponseNotFound(res *gitlab.Response) bool {
	if res != nil && res.StatusCode == 404 {
		return true
	}
	return false
}

// TimeToMetaTime returns nil if parameter is nil, otherwise metav1.Time value
func TimeToMetaTime(t *time.Time) *metav1.Time {
	if t == nil {
		return nil
	}
	return &metav1.Time{Time: *t}
}
