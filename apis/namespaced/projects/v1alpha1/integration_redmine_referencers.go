/*
Copyright 2026 The Crossplane Authors.

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

	"github.com/crossplane/crossplane-runtime/v2/pkg/reference"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ResolveReferences of this IntegrationRedmine
func (mg *IntegrationRedmine) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPINamespacedResolver(c, mg)

	rsp, err := r.Resolve(ctx, reference.NamespacedResolutionRequest{
		CurrentValue: fromPtrValue(mg.Spec.ForProvider.ProjectID),
		Reference:    mg.Spec.ForProvider.ProjectIDRef,
		Selector:     mg.Spec.ForProvider.ProjectIDSelector,
		To:           reference.To{Managed: &Project{}, List: &ProjectList{}},
		Extract:      reference.ExternalName(),
	})

	if err != nil {
		return errors.Wrap(err, "spec.forProvider.projectId")
	}

	resolvedID, err := toPtrValue(rsp.ResolvedValue)
	if err != nil {
		return errors.Wrap(err, "spec.forProvider")
	}
	mg.Spec.ForProvider.ProjectID = resolvedID
	mg.Spec.ForProvider.ProjectIDRef = rsp.ResolvedReference

	return nil
}
