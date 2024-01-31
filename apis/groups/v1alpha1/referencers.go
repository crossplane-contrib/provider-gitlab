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
func toPtrValue(v string) (*int, error) {
	if v == "" {
		return nil, nil
	}

	r, err := strconv.Atoi(v)
	return &r, err
}

// ResolveReferences of this Variable
func (mg *Variable) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// resolve spec.forProvider.projectIdRef
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: fromPtrValue(mg.Spec.ForProvider.GroupID),
		Reference:    mg.Spec.ForProvider.GroupIDRef,
		Selector:     mg.Spec.ForProvider.GroupIDSelector,
		To:           reference.To{Managed: &Group{}, List: &GroupList{}},
		Extract:      reference.ExternalName(),
	})

	if err != nil {
		return errors.Wrap(err, "spec.forProvider.groupId")
	}

	resolvedID, err := toPtrValue(rsp.ResolvedValue)
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.groupId")
	}

	mg.Spec.ForProvider.GroupID = resolvedID
	mg.Spec.ForProvider.GroupIDRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this Member
func (mg *Member) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// resolve spec.forProvider.projectIdRef
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: fromPtrValue(mg.Spec.ForProvider.GroupID),
		Reference:    mg.Spec.ForProvider.GroupIDRef,
		Selector:     mg.Spec.ForProvider.GroupIDSelector,
		To:           reference.To{Managed: &Group{}, List: &GroupList{}},
		Extract:      reference.ExternalName(),
	})

	if err != nil {
		return errors.Wrap(err, "spec.forProvider.groupId")
	}

	resolvedID, err := toPtrValue(rsp.ResolvedValue)
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.groupId")
	}

	mg.Spec.ForProvider.GroupID = resolvedID
	mg.Spec.ForProvider.GroupIDRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this Deploy Token
func (mg *DeployToken) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// resolve spec.forProvider.projectIdRef
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: fromPtrValue(mg.Spec.ForProvider.GroupID),
		Reference:    mg.Spec.ForProvider.GroupIDRef,
		Selector:     mg.Spec.ForProvider.GroupIDSelector,
		To:           reference.To{Managed: &Group{}, List: &GroupList{}},
		Extract:      reference.ExternalName(),
	})

	if err != nil {
		return errors.Wrap(err, "spec.forProvider.groupId")
	}

	resolvedID, err := toPtrValue(rsp.ResolvedValue)
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.groupId")
	}

	mg.Spec.ForProvider.GroupID = resolvedID
	mg.Spec.ForProvider.GroupIDRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this Group.
func (mg *Group) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	var rsp reference.ResolutionResponse
	var err error

	var idstrp *string
	if mg.Spec.ForProvider.ParentID != nil {
		str := strconv.Itoa(*mg.Spec.ForProvider.ParentID)
		idstrp = &str
	}

	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(idstrp),
		Extract:      reference.ExternalName(),
		Reference:    mg.Spec.ForProvider.ParentIDRef,
		Selector:     mg.Spec.ForProvider.ParentIDSelector,
		To: reference.To{
			List:    &GroupList{},
			Managed: &Group{},
		},
	})
	if err != nil {
		return errors.Wrap(err, "mg.Spec.ForProvider.ParentID")
	}

	id, err := toPtrValue(rsp.ResolvedValue)
	if err != nil {
		return errors.Wrap(err, "mg.Spec.ForProvider.ParentID")
	}

	mg.Spec.ForProvider.ParentID = id
	mg.Spec.ForProvider.ParentIDRef = rsp.ResolvedReference

	for i3 := 0; i3 < len(mg.Spec.ForProvider.SharedWithGroups); i3++ {
		idstr := strconv.Itoa(*mg.Spec.ForProvider.SharedWithGroups[i3].GroupID)
		rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
			CurrentValue: reference.FromPtrValue(&idstr),
			Extract:      reference.ExternalName(),
			Reference:    mg.Spec.ForProvider.SharedWithGroups[i3].GroupIDRef,
			Selector:     mg.Spec.ForProvider.SharedWithGroups[i3].GroupIDSelector,
			To: reference.To{
				List:    &GroupList{},
				Managed: &Group{},
			},
		})
		if err != nil {
			return errors.Wrap(err, "mg.Spec.ForProvider.SharedWithGroups[i3].GroupID")
		}

		id, err := toPtrValue(rsp.ResolvedValue)
		if err != nil {
			return errors.Wrap(err, "mg.Spec.ForProvider.SharedWithGroups[i3].GroupID")
		}
		mg.Spec.ForProvider.SharedWithGroups[i3].GroupID = id
		mg.Spec.ForProvider.SharedWithGroups[i3].GroupIDRef = rsp.ResolvedReference

	}

	return nil
}
