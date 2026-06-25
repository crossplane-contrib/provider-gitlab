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

package v1alpha1

import (
	"context"
	"strconv"

	"github.com/crossplane/crossplane-runtime/v2/pkg/reference"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// resolve int64 ptr to string value
func fromPtrValue(v *int64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatInt(*v, 10)
}

// resolve string value to int64 pointer
func toPtrValue(v string) (*int64, error) {
	if v == "" {
		return nil, nil
	}

	r, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// ResolveReferences of this ServiceAccountAccessToken
func (mg *ServiceAccountAccessToken) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPINamespacedResolver(c, mg)

	// resolve spec.forProvider.serviceAccountIdRef
	rsp, err := r.Resolve(ctx, reference.NamespacedResolutionRequest{
		CurrentValue: fromPtrValue(mg.Spec.ForProvider.ServiceAccountID),
		Reference:    mg.Spec.ForProvider.ServiceAccountIDRef,
		Selector:     mg.Spec.ForProvider.ServiceAccountIDSelector,
		To:           reference.To{Managed: &ServiceAccount{}, List: &ServiceAccountList{}},
		Extract:      reference.ExternalName(),
	})

	if err != nil {
		return errors.Wrap(err, "spec.forProvider.serviceAccountId")
	}

	resolvedSAID, err := toPtrValue(rsp.ResolvedValue)
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.serviceAccountId")
	}

	mg.Spec.ForProvider.ServiceAccountID = resolvedSAID
	mg.Spec.ForProvider.ServiceAccountIDRef = rsp.ResolvedReference

	return nil
}
