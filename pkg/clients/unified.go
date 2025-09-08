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

package clients

import (
	"context"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/cluster/v1beta1"
	namespacedv1beta1 "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/v1beta1"
)

func ResolveProviderConfig(ctx context.Context, crClient client.Client, mg resource.Managed) (*Config, error) {
	switch mg := mg.(type) {
	case resource.LegacyManaged:
		return resolveLegacyProviderConfig(ctx, crClient, mg)
	case resource.ModernManaged:
		return resolveNamespacedProviderConfig(ctx, crClient, mg)
	default:
		return nil, errors.New("unsupported resource type")
	}
}

func resolveLegacyProviderConfig(ctx context.Context, c client.Client, mg resource.LegacyManaged) (*Config, error) {
	pcRef := mg.GetProviderConfigReference()
	if pcRef == nil {
		return nil, errors.New("providerConfigRef is not given")
	}

	pc := &v1beta1.ProviderConfig{}
	if err := c.Get(ctx, types.NamespacedName{Name: pcRef.Name}, pc); err != nil {
		return nil, errors.Wrap(err, "cannot get referenced Provider")
	}

	t := resource.NewLegacyProviderConfigUsageTracker(c, &v1beta1.ProviderConfigUsage{})
	if err := t.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, "cannot track ProviderConfig usage")
	}

	switch s := pc.Spec.Credentials.Source; s {
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
			AuthMethod:         pc.Spec.Credentials.Method,
		}, nil
	default:
		return nil, errors.Errorf("credentials source %s is not currently supported", s)
	}
}

func resolveNamespacedProviderConfig(ctx context.Context, crClient client.Client, mg resource.ModernManaged) (*Config, error) {
	configRef := mg.GetProviderConfigReference()
	if configRef == nil {
		return nil, errors.New("provider config is not set")
	}

	// Try namespaced ProviderConfig first
	pc := &namespacedv1beta1.ProviderConfig{}
	if err := crClient.Get(ctx, types.NamespacedName{Name: configRef.Name, Namespace: mg.GetNamespace()}, pc); err == nil {
		return buildConfigFromNamespacedPC(ctx, crClient, mg, pc)
	}

	// Fallback to ClusterProviderConfig
	cpc := &namespacedv1beta1.ClusterProviderConfig{}
	if err := crClient.Get(ctx, types.NamespacedName{Name: configRef.Name}, cpc); err != nil {
		return nil, errors.Wrap(err, "cannot get provider config")
	}

	return buildConfigFromClusterPC(ctx, crClient, cpc)
}

func buildConfigFromNamespacedPC(ctx context.Context, crClient client.Client, mg resource.ModernManaged, pc *namespacedv1beta1.ProviderConfig) (*Config, error) {
	t := resource.NewProviderConfigUsageTracker(crClient, &namespacedv1beta1.ProviderConfigUsage{})
	if err := t.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, "cannot track ProviderConfig usage")
	}

	switch s := pc.Spec.Credentials.Source; s {
	case xpv1.CredentialsSourceSecret:
		csr := pc.Spec.Credentials.SecretRef
		if csr == nil {
			return nil, errors.New("no credentials secret referenced")
		}
		s := &corev1.Secret{}
		if err := crClient.Get(ctx, types.NamespacedName{Namespace: csr.Namespace, Name: csr.Name}, s); err != nil {
			return nil, errors.Wrap(err, "cannot get credentials secret")
		}
		return &Config{
			BaseURL:    pc.Spec.BaseURL,
			Token:      string(s.Data[csr.Key]),
			AuthMethod: pc.Spec.Credentials.Method,
		}, nil
	default:
		return nil, errors.Errorf("credentials source %s is not currently supported", s)
	}
}

func buildConfigFromClusterPC(ctx context.Context, crClient client.Client, cpc *namespacedv1beta1.ClusterProviderConfig) (*Config, error) {
	switch s := cpc.Spec.Credentials.Source; s {
	case xpv1.CredentialsSourceSecret:
		csr := cpc.Spec.Credentials.SecretRef
		if csr == nil {
			return nil, errors.New("no credentials secret referenced")
		}
		s := &corev1.Secret{}
		if err := crClient.Get(ctx, types.NamespacedName{Namespace: csr.Namespace, Name: csr.Name}, s); err != nil {
			return nil, errors.Wrap(err, "cannot get credentials secret")
		}
		return &Config{
			BaseURL:    cpc.Spec.BaseURL,
			Token:      string(s.Data[csr.Key]),
			AuthMethod: cpc.Spec.Credentials.Method,
		}, nil
	default:
		return nil, errors.Errorf("credentials source %s is not currently supported", s)
	}
}
