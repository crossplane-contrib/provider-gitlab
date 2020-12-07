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
	"fmt"
	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/go-ini/ini"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "github.com/xanzy/go-gitlab"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"

	"github.com/crossplane-contrib/provider-gitlab/apis/v1beta1"
)

// DefaultSection for INI files.
const DefaultSection = ini.DefaultSection

// Config provides gitlab configurations for the Gitlab client
type Config struct {
	Token   string
	BaseURL string
}

// NewClient creates new Gitlab Client with provided Gitlab Configurations/Credentials.
func NewClient(c Config) *gitlab.Client {
	var cl *gitlab.Client
	var err error
	if c.BaseURL != "" {
		cl, err = gitlab.NewClient(c.Token, gitlab.WithBaseURL(c.BaseURL))
	} else {
		cl, err = gitlab.NewClient(c.Token)
	}
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

// UseProviderConfig to produce a config that can be used to authenticate to AWS.
func UseProviderConfig(ctx context.Context, c client.Client, mg resource.Managed) (*Config, error) {
	// pc := &v1beta1.ProviderConfig{}
	pc := &v1beta1.ProviderConfig{}
	if err := c.Get(ctx, types.NamespacedName{Name: mg.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, "cannot get referenced Provider")
	}

	t := resource.NewProviderConfigUsageTracker(c, &v1beta1.ProviderConfigUsage{})
	if err := t.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, "cannot track ProviderConfig usage")
	}

	switch s := pc.Spec.Credentials.Source; s { //nolint:exhaustive
	case runtimev1alpha1.CredentialsSourceSecret:
		csr := pc.Spec.Credentials.SecretRef
		if csr == nil {
			return nil, errors.New("no credentials secret referenced")
		}
		s := &corev1.Secret{}
		if err := c.Get(ctx, types.NamespacedName{Namespace: csr.Namespace, Name: csr.Name}, s); err != nil {
			return nil, errors.Wrap(err, "cannot get credentials secret")
		}
		return UseProviderSecret(ctx, s.Data[csr.Key], DefaultSection)
	default:
		return nil, errors.Errorf("credentials source %s is not currently supported", s)
	}
}

// UseProviderSecret retrieves the gitlab token and base url that are being used
// to authenticate to Gitlab.
// Example:
// [default]
// token = <YOUR_PERSONAL_ACCESS_TOKEN>
// base_url = <YOUR_GITLAB_BASE_URL>
func UseProviderSecret(_ context.Context, data []byte, section string) (*Config, error) {
	config, err := ini.InsensitiveLoad(data)
	if err != nil {
		return &Config{}, errors.Wrap(err, "cannot parse credentials secret")
	}

	iniSection, err := config.GetSection(section)
	if err != nil {
		return &Config{}, errors.Wrap(err, fmt.Sprintf("cannot get %s section in credentials secret", section))
	}

	token := iniSection.Key("token")
	baseURL := iniSection.Key("base_url")

	if token == nil || baseURL == nil {
		return &Config{}, errors.New("returned key can be empty but cannot be nil")
	}

	return &Config{Token: token.Value(), BaseURL: baseURL.Value()}, nil
}

// LateInitializeStringRef returns `from` if `in` is nil and `from` is non-empty,
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

// LateInitializeAccessControlValue returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeVisibilityValue(in *v1alpha1.VisibilityValue, from gitlab.VisibilityValue) *v1alpha1.VisibilityValue {
	if in == nil && from != "" {
		return (*v1alpha1.VisibilityValue)(&from)
	}
	return in
}

// LateInitializeAccessControlValue returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeMergeMethodValue(in *v1alpha1.MergeMethodValue, from gitlab.MergeMethodValue) *v1alpha1.MergeMethodValue {
	if in == nil && from != "" {
		return (*v1alpha1.MergeMethodValue)(&from)
	}
	return in
}

func VisibilityValueV1alpha1ToGitlab(from *v1alpha1.VisibilityValue) *gitlab.VisibilityValue {
	return (*gitlab.VisibilityValue)(from)
}
func VisibilityValueStringToGitlab(from string) *gitlab.VisibilityValue {
	return (*gitlab.VisibilityValue)(&from)
}

func AccessControlValueV1alpha1ToGitlab(from *v1alpha1.AccessControlValue) *gitlab.AccessControlValue {
	return (*gitlab.AccessControlValue)(from)
}
func AccessControlValueStringToGitlab(from string) *gitlab.AccessControlValue {
	return (*gitlab.AccessControlValue)(&from)
}

func MergeMethodV1alpha1ToGitlab(from *v1alpha1.MergeMethodValue) *gitlab.MergeMethodValue {
	return (*gitlab.MergeMethodValue)(from)
}
func MergeMethodStringToGitlab(from string) *gitlab.MergeMethodValue {
	return (*gitlab.MergeMethodValue)(&from)
}

func StringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func IsBoolEqualToBoolPtr(bp *bool, b bool) bool {
	if bp != nil {
		if !cmp.Equal(*bp, b) {
			return false
		}
	}
	return true
}

func IsIntEqualToIntPtr(ip *int, i int) bool {
	if ip != nil {
		if !cmp.Equal(*ip, i) {
			return false
		}
	}
	return true
}
