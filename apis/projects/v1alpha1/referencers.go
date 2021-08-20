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

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// resolve int ptr to string value
func fromPtrValue(v *int) string {
	if v == nil {
		return ""
	}
	return strconv.Itoa(*v)
}

// resolve string value to int pointer
func toPtrValue(v string) *int {
	if v == "" {
		return nil
	}

	r, err := strconv.Atoi(v)
	if err != nil {
		return nil
	}

	return &r
}

// ResolveReferences of this ProjectHook
func (mg *ProjectHook) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// resolve spec.forProvider.projectIdRef
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: fromPtrValue(mg.Spec.ForProvider.ProjectID),
		Reference:    mg.Spec.ForProvider.ProjectIDRef,
		Selector:     mg.Spec.ForProvider.ProjectIDSelector,
		To:           reference.To{Managed: &Project{}, List: &ProjectList{}},
		Extract:      reference.ExternalName(),
	})

	if err != nil {
		return errors.Wrap(err, "spec.forProvider.projectId")
	}

	mg.Spec.ForProvider.ProjectID = toPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.ProjectIDRef = rsp.ResolvedReference

	return nil

}

// ResolveReferences of this Project
func (mg *Project) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// resolve spec.forProvider.namespaceIdRef
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: fromPtrValue(mg.Spec.ForProvider.NamespaceID),
		Reference:    mg.Spec.ForProvider.NamespaceIDRef,
		Selector:     mg.Spec.ForProvider.NamespaceIDSelector,
		To:           reference.To{Managed: &v1alpha1.Group{}, List: &v1alpha1.GroupList{}},
		Extract:      reference.ExternalName(),
	})

	if err != nil {
		return errors.Wrap(err, "spec.forProvider.namespaceId")
	}

	mg.Spec.ForProvider.NamespaceID = toPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.NamespaceIDRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this Deploy Token
func (mg *DeployToken) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// resolve spec.forProvider.projectIdRef
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: fromPtrValue(mg.Spec.ForProvider.ProjectID),
		Reference:    mg.Spec.ForProvider.ProjectIDRef,
		Selector:     mg.Spec.ForProvider.ProjectIDSelector,
		To:           reference.To{Managed: &Project{}, List: &ProjectList{}},
		Extract:      reference.ExternalName(),
	})

	if err != nil {
		return errors.Wrap(err, "spec.forProvider.projectId")
	}

	mg.Spec.ForProvider.ProjectID = toPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.ProjectIDRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this ProjectMember
func (mg *ProjectMember) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// resolve spec.forProvider.projectIdRef
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: fromPtrValue(mg.Spec.ForProvider.ProjectID),
		Reference:    mg.Spec.ForProvider.ProjectIDRef,
		Selector:     mg.Spec.ForProvider.ProjectIDSelector,
		To:           reference.To{Managed: &Project{}, List: &ProjectList{}},
		Extract:      reference.ExternalName(),
	})

	if err != nil {
		return errors.Wrap(err, "spec.forProvider.projectId")
	}

	mg.Spec.ForProvider.ProjectID = toPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.ProjectIDRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this Variable
func (mg *Variable) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// resolve spec.forProvider.projectIdRef
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: fromPtrValue(mg.Spec.ForProvider.ProjectID),
		Reference:    mg.Spec.ForProvider.ProjectIDRef,
		Selector:     mg.Spec.ForProvider.ProjectIDSelector,
		To:           reference.To{Managed: &Project{}, List: &ProjectList{}},
		Extract:      reference.ExternalName(),
	})

	if err != nil {
		return errors.Wrap(err, "spec.forProvider.projectId")
	}

	mg.Spec.ForProvider.ProjectID = toPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.ProjectIDRef = rsp.ResolvedReference

	return nil
}
