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

	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
)

// UpdateStringFromSecret fetches a sensitive setting value from a Kubernetes Secret
// and assigns it to the provided parameter pointer.
func UpdateStringFromSecret(mg resource.Managed, ctx context.Context, kube client.Client, selector *v2.LocalSecretKeySelector, param **string) error {
	value, err := common.GetTokenValueFromLocalSecret(ctx, kube, mg, selector)
	if err != nil {
		return err
	}
	*param = value
	return nil
}
