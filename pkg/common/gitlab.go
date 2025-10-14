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

package common

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"golang.org/x/oauth2"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	legacyV1Beta1 "github.com/crossplane-contrib/provider-gitlab/apis/cluster/v1beta1"
	namespacedV1Beta1 "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/v1beta1"
	auth "github.com/crossplane-contrib/provider-gitlab/pkg/common/auth"
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
	AuthMethod         auth.AuthType
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
	case auth.BasicAuth:
		ba := &BasicAuth{}
		if err = json.Unmarshal([]byte(c.Token), ba); err != nil {
			panic(err)
		}
		cl, err = gitlab.NewBasicAuthClient(ba.Username, ba.Password, options...) //nolint:staticcheck
	case auth.JobToken:
		cl, err = gitlab.NewJobClient(c.Token, options...)
	case auth.OAuthToken:
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: c.Token})
		cl, err = gitlab.NewAuthSourceClient(gitlab.OAuthTokenSource{TokenSource: ts}, options...)
	case auth.PersonalAccessToken:
		cl, err = gitlab.NewClient(c.Token, options...)
	default:
		cl, err = gitlab.NewClient(c.Token, options...)
	}
	if err != nil {
		panic(err)
	}

	return cl
}

// GetConfig constructs a Config that can be used to authenticate to Gitlab
// API by the Gitlab Go client
func GetConfig(ctx context.Context, c client.Client, mg resource.Managed) (*Config, error) {
	switch mgC := mg.(type) {
	case resource.LegacyManaged:
		switch {
		case mgC.GetProviderConfigReference() != nil:
			return UseLegacyProviderConfig(ctx, c, mgC)
		default:
			return nil, errors.New("providerConfigRef is not given")
		}
	case resource.ModernManaged:
		switch {
		case mgC.GetProviderConfigReference() != nil:
			return UseProvicerConfig(ctx, c, mgC)
		default:
			return nil, errors.New("providerConfigRef is not given")
		}
	default:
		return nil, errors.New("unknown managed resource type")
	}
}

// UseProviderConfig to produce a config that can be used to authenticate to Gitlab.
func UseLegacyProviderConfig(ctx context.Context, c client.Client, mg resource.LegacyManaged) (*Config, error) {
	pc := &legacyV1Beta1.ProviderConfig{}
	if err := c.Get(ctx, types.NamespacedName{Name: mg.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, "cannot get referenced Provider")
	}

	t := resource.NewLegacyProviderConfigUsageTracker(c, &legacyV1Beta1.ProviderConfigUsage{})
	if err := t.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, "cannot track ProviderConfig usage")
	}

	switch s := pc.Spec.Credentials.Source; s {
	case xpv1.CredentialsSourceSecret:
		if pc.Spec.Credentials.SecretRef == nil {
			return nil, errors.New("no credentials secret referenced")
		}

		token, err := GetTokenValueFromSecret(ctx, c, mg, pc.Spec.Credentials.SecretRef)
		if err != nil {
			return nil, err
		}

		return &Config{
			BaseURL:            pc.Spec.BaseURL,
			Token:              *token,
			InsecureSkipVerify: ptr.Deref(pc.Spec.InsecureSkipVerify, false),
			AuthMethod:         pc.Spec.Credentials.Method,
		}, nil
	default:
		return nil, errors.Errorf("credentials source %s is not currently supported", s)
	}
}

func UseProvicerConfig(ctx context.Context, c client.Client, mg resource.ModernManaged) (*Config, error) {
	pcRef := mg.GetProviderConfigReference()

	switch pcRef.Kind {
	case "ClusterProviderConfig":
		cpc := &namespacedV1Beta1.ClusterProviderConfig{}
		if err := c.Get(ctx, types.NamespacedName{Name: pcRef.Name}, cpc); err != nil {
			return nil, errors.Wrap(err, "cannot get referenced ClusterProviderConfig")
		}
		return buildConfigFromSpec(ctx, c, mg, cpc.Spec)
	default: // "ProviderConfig" or empty (default)
		pc := &namespacedV1Beta1.ProviderConfig{}
		if err := c.Get(ctx, types.NamespacedName{Name: pcRef.Name, Namespace: mg.GetNamespace()}, pc); err != nil {
			return nil, errors.Wrap(err, "cannot get referenced ProviderConfig")
		}
		return buildConfigFromSpec(ctx, c, mg, pc.Spec)
	}
}

func buildConfigFromSpec(ctx context.Context, c client.Client, mg resource.ModernManaged, spec namespacedV1Beta1.ProviderConfigSpec) (*Config, error) {
	t := resource.NewProviderConfigUsageTracker(c, &namespacedV1Beta1.ProviderConfigUsage{})
	if err := t.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, "cannot track ProviderConfig usage")
	}

	switch s := spec.Credentials.Source; s {
	case xpv1.CredentialsSourceSecret:
		if spec.Credentials.SecretRef == nil {
			return nil, errors.New("no credentials secret referenced")
		}

		token, err := GetTokenValueFromSecret(ctx, c, mg, spec.Credentials.SecretRef)
		if err != nil {
			return nil, err
		}

		return &Config{
			BaseURL:            spec.BaseURL,
			Token:              *token,
			InsecureSkipVerify: ptr.Deref(spec.InsecureSkipVerify, false),
			AuthMethod:         spec.Credentials.Method,
		}, nil
	default:
		return nil, errors.Errorf("credentials source %s is not currently supported", s)
	}
}
