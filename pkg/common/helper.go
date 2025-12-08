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

package common

import (
	"context"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ErrSecretNotFound    = "Cannot find referenced secret"
	ErrSecretKeyNotFound = "Cannot find key in referenced secret"
	ErrSecretSelectorNil = "Secret selector is nil"
)

func GetTokenValueFromSecret(ctx context.Context, client client.Client, m resource.Managed, selector *xpv1.SecretKeySelector) (*string, error) {
	if selector == nil {
		return nil, errors.Errorf(ErrSecretSelectorNil)
	}

	secret := &corev1.Secret{}
	if err := client.Get(ctx, types.NamespacedName{Name: selector.SecretReference.Name, Namespace: selector.SecretReference.Namespace}, secret); err != nil {
		return nil, errors.Wrap(err, ErrSecretNotFound)
	}

	value := secret.Data[selector.Key]
	if value == nil {
		return nil, errors.Errorf(ErrSecretKeyNotFound)
	}

	data := string(value)
	return &data, nil
}

func GetTokenValueFromLocalSecret(ctx context.Context, client client.Client, m resource.Managed, l *xpv1.LocalSecretKeySelector) (*string, error) {
	if l == nil {
		return nil, errors.Errorf(ErrSecretSelectorNil)
	}

	return GetTokenValueFromSecret(ctx, client, m, &xpv1.SecretKeySelector{
		Key: l.Key,
		SecretReference: xpv1.SecretReference{
			Name:      l.Name,
			Namespace: m.GetNamespace(),
		},
	})
}

// ResolvePublicJobsSetting determines the effective publicJobs value
// prioritizing publicJobs over the deprecated publicBuilds field.
// Returns the resolved value and whether the deprecated publicBuilds field was used.
func ResolvePublicJobsSetting(publicBuilds, publicJobs *bool) (*bool, bool) {
	if publicJobs != nil {
		// New field takes precedence
		return publicJobs, false
	}
	if publicBuilds != nil {
		// Deprecated field is used
		return publicBuilds, true
	}
	return nil, false
}
