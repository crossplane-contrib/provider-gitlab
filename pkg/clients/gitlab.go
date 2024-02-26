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
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/pkg/errors"
	gitlab "github.com/xanzy/go-gitlab"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/apis/v1beta1"
)

// Config provides gitlab configurations for the Gitlab client
type Config struct {
	Token              string
	BaseURL            string
	InsecureSkipVerify bool
}

// NewClient creates new Gitlab Client with provided Gitlab Configurations/Credentials.
func NewClient(c Config) *gitlab.Client {
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
	cl, err := gitlab.NewClient(c.Token, options...)
	if err != nil {
		panic(err)
	}
	return cl
}

// GetConfig constructs a Config that can be used to authenticate to Gitlab
// API by the Gitlab Go client
func GetConfig(ctx context.Context, c client.Client, mg resource.Managed) (*Config, error) {
	switch {
	case mg.GetProviderConfigReference() != nil:
		return UseProviderConfig(ctx, c, mg)
	default:
		return nil, errors.New("providerConfigRef is not given")
	}
}

// UseProviderConfig to produce a config that can be used to authenticate to Gitlab.
func UseProviderConfig(ctx context.Context, c client.Client, mg resource.Managed) (*Config, error) {
	pc := &v1beta1.ProviderConfig{}
	if err := c.Get(ctx, types.NamespacedName{Name: mg.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, "cannot get referenced Provider")
	}

	t := resource.NewProviderConfigUsageTracker(c, &v1beta1.ProviderConfigUsage{})
	if err := t.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, "cannot track ProviderConfig usage")
	}

	switch s := pc.Spec.Credentials.Source; s { //nolint:exhaustive
	case xpv1.CredentialsSourceSecret:
		csr := pc.Spec.Credentials.SecretRef
		if csr == nil {
			return nil, errors.New("no credentials secret referenced")
		}
		s := &corev1.Secret{}
		if err := c.Get(ctx, types.NamespacedName{Namespace: csr.Namespace, Name: csr.Name}, s); err != nil {
			return nil, errors.Wrap(err, "cannot get credentials secret")
		}
		return &Config{
			BaseURL:            pc.Spec.BaseURL,
			Token:              string(s.Data[csr.Key]),
			InsecureSkipVerify: ptr.Deref(pc.Spec.InsecureSkipVerify, false),
		}, nil
	default:
		return nil, errors.Errorf("credentials source %s is not currently supported", s)
	}
}

// LateInitializeStringPtr returns `from` if `in` is nil and `from` is non-empty,
// in other cases it returns `in`.
func LateInitializeStringPtr(in *string, from string) *string {
	if in == nil && from != "" {
		return &from
	}
	return in
}

// LateInitializeAccessControlValue returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeAccessControlValue(in *v1alpha1.AccessControlValue, from gitlab.AccessControlValue) *v1alpha1.AccessControlValue {
	if in == nil && from != "" {
		return (*v1alpha1.AccessControlValue)(&from)
	}
	return in
}

// LateInitializeVisibilityValue returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeVisibilityValue(in *v1alpha1.VisibilityValue, from gitlab.VisibilityValue) *v1alpha1.VisibilityValue {
	if in == nil && from != "" {
		return (*v1alpha1.VisibilityValue)(&from)
	}
	return in
}

// LateInitializeMergeMethodValue returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeMergeMethodValue(in *v1alpha1.MergeMethodValue, from gitlab.MergeMethodValue) *v1alpha1.MergeMethodValue {
	if in == nil && from != "" {
		return (*v1alpha1.MergeMethodValue)(&from)
	}
	return in
}

// VisibilityValueV1alpha1ToGitlab converts *v1alpha1.VisibilityValue to *gitlab.VisibilityValue
func VisibilityValueV1alpha1ToGitlab(from *v1alpha1.VisibilityValue) *gitlab.VisibilityValue {
	return (*gitlab.VisibilityValue)(from)
}

// VisibilityValueStringToGitlab converts string to *gitlab.VisibilityValue
func VisibilityValueStringToGitlab(from string) *gitlab.VisibilityValue {
	return (*gitlab.VisibilityValue)(&from)
}

// AccessControlValueV1alpha1ToGitlab converts *v1alpha1.AccessControlValue to *gitlab.AccessControlValue
func AccessControlValueV1alpha1ToGitlab(from *v1alpha1.AccessControlValue) *gitlab.AccessControlValue {
	return (*gitlab.AccessControlValue)(from)
}

// ContainerExpirationPolicyAttributesV1alpha1ToGitlab converts *v1alpha1.ContainerExpirationPolicyAttributes to *gitlab.ContainerExpirationPolicyAttributes
func ContainerExpirationPolicyAttributesV1alpha1ToGitlab(from *v1alpha1.ContainerExpirationPolicyAttributes) *gitlab.ContainerExpirationPolicyAttributes {
	return (*gitlab.ContainerExpirationPolicyAttributes)(from)
}

// AccessControlValueStringToGitlab converts string to *gitlab.AccessControlValue
func AccessControlValueStringToGitlab(from string) *gitlab.AccessControlValue {
	return (*gitlab.AccessControlValue)(&from)
}

// MergeMethodV1alpha1ToGitlab converts *v1alpha1.MergeMethodValue to *gitlab.MergeMethodValue
func MergeMethodV1alpha1ToGitlab(from *v1alpha1.MergeMethodValue) *gitlab.MergeMethodValue {
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
