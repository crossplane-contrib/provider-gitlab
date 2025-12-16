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

package runners

import (
	"context"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
)

// UpdateVariableFromSecret updates the Variable parameters with the value from the secret.
func UpdateVariableFromSecret(kube client.Client, mg resource.Managed, ctx context.Context, selector *xpv1.LocalSecretKeySelector, params *v1alpha1.CommonVariableParameters) error {
	value, err := common.GetTokenValueFromLocalSecret(ctx, kube, mg, selector)
	if err != nil {
		return err
	}

	// Mask variable if it hasn't already been explicitly configured.
	if params.Masked == nil {
		params.Masked = gitlab.Ptr(true)
	}

	// Make variable raw if it hasn't already been explicitly configured.
	if params.Raw == nil {
		params.Raw = gitlab.Ptr(true)
	}

	params.Value = value
	return nil
}
