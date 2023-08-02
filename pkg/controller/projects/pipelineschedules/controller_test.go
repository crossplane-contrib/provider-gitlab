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

package pipelineschedules

import (
	"context"
	"net/http"
	"strconv"
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects/fake"
)

var (
	s                = ""
	f                = false
	errorMessage     = "restult: -expected, +actual: \n%s"
	id               = 1234
	standardID       = 0
	extName          = strconv.Itoa(id)
	projectID        = "123456"
	standardPsParams = v1alpha1.PipelineScheduleParameters{
		ProjectID:    &projectID,
		Description:  s,
		Ref:          s,
		Cron:         s,
		CronTimezone: &s,
		Active:       &f,
	}
	psParams = v1alpha1.PipelineScheduleParameters{
		ProjectID:   &projectID,
		Description: s,
		Ref:         s,
		Cron:        s,
		Active:      &f,
	}
	pv1 = &v1alpha1.PipelineVariable{
		Key:          "testKey1",
		Value:        "testValue1",
		VariableType: &s,
	}
	pv2 = &v1alpha1.PipelineVariable{
		Key:          "testKey2",
		Value:        "testValue2",
		VariableType: &s,
	}
	pv2Update = &v1alpha1.PipelineVariable{
		Key:          "testKey2",
		Value:        "testValue2Update",
		VariableType: &s,
	}
	gPvArr = []*gitlab.PipelineVariable{
		{
			Key:          "testKey1",
			Value:        "testValue1",
			VariableType: "",
		},
		{
			Key:          "testKey2",
			Value:        "testValue2",
			VariableType: "",
		},
	}
)

type args struct {
	cr     resource.Managed
	kube   client.Client
	client projects.PipelineScheduleClient
}

type psModifier func(*v1alpha1.PipelineSchedule)

func withExternalName(s string) psModifier {
	return func(ps *v1alpha1.PipelineSchedule) { meta.SetExternalName(ps, s) }
}

func withParams(p v1alpha1.PipelineScheduleParameters) psModifier {
	return func(ps *v1alpha1.PipelineSchedule) { ps.Spec.ForProvider = p }
}

func withID(s int) psModifier {
	return func(ps *v1alpha1.PipelineSchedule) { ps.Status.AtProvider.ID = &s }
}

func withConditions(c xpv1.Condition) psModifier {
	return func(ps *v1alpha1.PipelineSchedule) { ps.Status.SetConditions(c) }
}

func withVariables(varr ...*v1alpha1.PipelineVariable) psModifier {
	return func(ps *v1alpha1.PipelineSchedule) {
		ps.Spec.ForProvider.Variables = make([]v1alpha1.PipelineVariable, len(varr))
		for i, v := range varr {
			ps.Spec.ForProvider.Variables[i] = *v
		}
	}
}

func withProjectID() psModifier {
	return func(ps *v1alpha1.PipelineSchedule) { ps.Spec.ForProvider.ProjectID = &extName }
}

func buildPs(m ...psModifier) *v1alpha1.PipelineSchedule {
	ps := &v1alpha1.PipelineSchedule{}
	for _, psm := range m {
		psm(ps)
	}
	return ps
}

func TestObserve(t *testing.T) {
	type expected struct {
		cr     resource.Managed
		result managed.ExternalObservation
		err    error
	}

	tcs := map[string]struct {
		args
		expected
	}{
		"NoExternalName": {
			args: args{
				cr: &v1alpha1.PipelineSchedule{},
			},
			expected: expected{
				cr:     &v1alpha1.PipelineSchedule{},
				err:    nil,
				result: managed.ExternalObservation{},
			},
		},
		"NoProjectID": {
			args: args{
				cr: buildPs(withExternalName(extName)),
			},
			expected: expected{
				cr:     buildPs(withExternalName(extName)),
				result: managed.ExternalObservation{},
				err:    errors.New(errNoProjectID),
			},
		},
		"GetFail400": {
			args: args{
				cr: buildPs(
					withExternalName(extName),
					withProjectID(),
				),
				client: &fake.MockClient{
					MockGetPipelineSchedule: func(pid interface{}, schedule int, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error) {
						return &gitlab.PipelineSchedule{}, nil, errors.New(errNotPipelineSchedule)
					},
				},
			},
			expected: expected{
				cr: buildPs(
					withExternalName(extName),
					withProjectID(),
				),
				result: managed.ExternalObservation{},
				err:    errors.Wrap(errors.New(errNotPipelineSchedule), errGetPipelineSchedule),
			},
		},
		"GetFail404": {
			args: args{
				cr: buildPs(
					withExternalName(extName),
					withProjectID(),
				),
				client: &fake.MockClient{
					MockGetPipelineSchedule: func(pid interface{}, schedule int, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error) {
						return &gitlab.PipelineSchedule{}, &gitlab.Response{
							Response: &http.Response{StatusCode: 404},
						}, errors.New(errNotPipelineSchedule)
					},
				},
			},
			expected: expected{
				cr: buildPs(
					withExternalName(extName),
					withProjectID(),
				),
				result: managed.ExternalObservation{},
				err:    nil,
			},
		},
		"SuccessLateInitializedFalse": {
			args: args{
				client: &fake.MockClient{
					MockGetPipelineSchedule: func(pid interface{}, schedule int, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error) {
						return &gitlab.PipelineSchedule{Variables: gPvArr}, nil, nil
					},
				},
				cr: buildPs(
					withExternalName(extName),
					withProjectID(),
					withParams(standardPsParams),
					withVariables(pv1, pv2),
				),
			},
			expected: expected{
				cr: buildPs(
					withExternalName(extName),
					withProjectID(),
					withParams(standardPsParams),
					withID(standardID),
					withConditions(xpv1.Available()),
					withVariables(pv1, pv2),
				),
				err: nil,
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: false,
				},
			},
		},
		"SuccessLateInitializedTrue": {
			args: args{
				client: &fake.MockClient{
					MockGetPipelineSchedule: func(pid interface{}, schedule int, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error) {
						return &gitlab.PipelineSchedule{Variables: gPvArr}, nil, nil
					},
				},
				cr: buildPs(
					withExternalName(extName),
					withParams(psParams),
				),
			},
			expected: expected{
				cr: buildPs(
					withExternalName(extName),
					withID(standardID),
					withParams(psParams),
					withConditions(xpv1.Available()),
					withVariables(pv1, pv2),
				),
				err: nil,
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
				},
			},
		},
		"SuccessUpToDateTrue": {
			args: args{
				client: &fake.MockClient{
					MockGetPipelineSchedule: func(pid interface{}, schedule int, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error) {
						return &gitlab.PipelineSchedule{}, nil, nil
					},
				},
				cr: buildPs(
					withParams(standardPsParams),
					withExternalName(extName),
				),
			},
			expected: expected{
				cr: buildPs(
					withParams(standardPsParams),
					withExternalName(extName),
					withID(standardID),
					withConditions(xpv1.Available()),
				),
				err: nil,
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: false,
				},
			},
		},
		"SuccessUpToDateFalse": {
			args: args{
				client: &fake.MockClient{
					MockGetPipelineSchedule: func(pid interface{}, schedule int, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error) {
						return &gitlab.PipelineSchedule{Cron: "cron"}, nil, nil
					},
				},
				cr: buildPs(
					withExternalName(extName),
					withParams(psParams),
				),
			},
			expected: expected{
				cr: buildPs(
					withExternalName(extName),
					withID(standardID),
					withConditions(xpv1.Available()),
					withParams(psParams),
				),
				err: nil,
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
			},
		},
	}

	for tn, tc := range tcs {
		t.Run(tn, func(t *testing.T) {
			victim := &external{kube: tc.kube, client: tc.client}
			result, err := victim.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.expected.err, err, test.EquateErrors()); diff != "" {
				t.Errorf(errorMessage, diff)
			}

			if diff := cmp.Diff(tc.expected.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf(errorMessage, diff)
			}

			if diff := cmp.Diff(tc.expected.result, result); diff != "" {
				t.Errorf(errorMessage, diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type expected struct {
		cr     resource.Managed
		err    error
		result managed.ExternalCreation
	}

	tcs := map[string]struct {
		args
		expected
	}{
		"NoProjectID": {
			args: args{
				cr: buildPs(withExternalName(extName)),
			},
			expected: expected{
				cr:     buildPs(withExternalName(extName)),
				result: managed.ExternalCreation{},
				err:    errors.New(errNoProjectID),
			},
		},
		"CreateSuccess": {
			args: args{
				client: &fake.MockClient{
					MockCreatePipelineSchedule: func(pid interface{}, opt *gitlab.CreatePipelineScheduleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error) {
						return &gitlab.PipelineSchedule{
							ID: id,
						}, nil, nil
					},
					MockCreatePipelineScheduleVariable: func(pid interface{}, schedule int, opt *gitlab.CreatePipelineScheduleVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineVariable, *gitlab.Response, error) {
						return nil, nil, nil
					},
				},
				cr: buildPs(
					withProjectID(),
					withVariables(pv1),
				),
			},
			expected: expected{
				cr: buildPs(
					withProjectID(),
					withExternalName(extName),
					withVariables(pv1),
				),
				err:    nil,
				result: managed.ExternalCreation{},
			},
		},
	}

	for tn, tc := range tcs {
		t.Run(tn, func(t *testing.T) {
			victim := &external{kube: tc.kube, client: tc.client}
			result, err := victim.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.expected.err, err, test.EquateErrors()); diff != "" {
				t.Errorf(errorMessage, diff)
			}

			if diff := cmp.Diff(tc.expected.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf(errorMessage, diff)
			}

			if diff := cmp.Diff(tc.expected.result, result); diff != "" {
				t.Errorf(errorMessage, diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type expected struct {
		cr     resource.Managed
		result managed.ExternalUpdate
		err    error
	}

	tcs := map[string]struct {
		args
		expected
	}{
		"NoExternalName": {
			args: args{
				cr: buildPs(),
			},
			expected: expected{
				cr:     buildPs(),
				err:    errors.New(errExternalNameMissing),
				result: managed.ExternalUpdate{},
			},
		},
		"NoProjectID": {
			args: args{
				cr: buildPs(withExternalName(extName)),
			},
			expected: expected{
				cr:     buildPs(withExternalName(extName)),
				result: managed.ExternalUpdate{},
				err:    errors.New(errNoProjectID),
			},
		},
		"UpdateSuccess": {
			args: args{
				client: &fake.MockClient{
					MockEditPipelineSchedule: func(pid interface{}, schedule int, opt *gitlab.EditPipelineScheduleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error) {
						return &gitlab.PipelineSchedule{}, nil, nil
					},
					MockGetPipelineSchedule: func(pid interface{}, schedule int, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error) {
						return &gitlab.PipelineSchedule{}, nil, nil
					},
				},
				cr: buildPs(
					withParams(standardPsParams),
					withExternalName(extName),
				),
			},
			expected: expected{
				cr: buildPs(
					withParams(standardPsParams),
					withExternalName(extName),
				),
				result: managed.ExternalUpdate{},
				err:    nil,
			},
		},
		"VariablesCreateSuccess": {
			args: args{
				client: &fake.MockClient{
					MockEditPipelineSchedule: func(pid interface{}, schedule int, opt *gitlab.EditPipelineScheduleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error) {
						return nil, nil, nil
					},
					MockGetPipelineSchedule: func(pid interface{}, schedule int, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error) {
						return &gitlab.PipelineSchedule{}, nil, nil
					},
					MockCreatePipelineScheduleVariable: func(pid interface{}, schedule int, opt *gitlab.CreatePipelineScheduleVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineVariable, *gitlab.Response, error) {
						return nil, nil, nil
					},
				},
				cr: buildPs(
					withExternalName(extName),
					withProjectID(),
					withVariables(pv1),
				),
			},
			expected: expected{
				cr: buildPs(
					withExternalName(extName),
					withProjectID(),
					withVariables(pv1),
				),
				result: managed.ExternalUpdate{},
				err:    nil,
			},
		},
		"VariablesUpdateSuccess": {
			args: args{
				client: &fake.MockClient{
					MockEditPipelineSchedule: func(pid interface{}, schedule int, opt *gitlab.EditPipelineScheduleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error) {
						return nil, nil, nil
					},
					MockGetPipelineSchedule: func(pid interface{}, schedule int, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error) {
						return &gitlab.PipelineSchedule{Variables: gPvArr}, nil, nil
					},
					MockEditPipelineScheduleVariable: func(pid interface{}, schedule int, key string, opt *gitlab.EditPipelineScheduleVariableOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineVariable, *gitlab.Response, error) {
						return nil, nil, nil
					},
				},
				cr: buildPs(
					withExternalName(extName),
					withProjectID(),
					withVariables(pv1, pv2Update),
				),
			},
			expected: expected{
				cr: buildPs(
					withExternalName(extName),
					withProjectID(),
					withVariables(pv1, pv2Update),
				),
				result: managed.ExternalUpdate{},
				err:    nil,
			},
		},
		"VariablesDeleteSuccess": {
			args: args{
				client: &fake.MockClient{
					MockEditPipelineSchedule: func(pid interface{}, schedule int, opt *gitlab.EditPipelineScheduleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error) {
						return nil, nil, nil
					},
					MockGetPipelineSchedule: func(pid interface{}, schedule int, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineSchedule, *gitlab.Response, error) {
						return &gitlab.PipelineSchedule{Variables: gPvArr}, nil, nil
					},
					MockDeletePipelineScheduleVariable: func(pid interface{}, schedule int, key string, options ...gitlab.RequestOptionFunc) (*gitlab.PipelineVariable, *gitlab.Response, error) {
						return nil, nil, nil
					},
				},
				cr: buildPs(
					withExternalName(extName),
					withProjectID(),
					withVariables(pv1),
				),
			},
			expected: expected{
				cr: buildPs(
					withExternalName(extName),
					withProjectID(),
					withVariables(pv1),
				),
				result: managed.ExternalUpdate{},
				err:    nil,
			},
		},
	}

	for tn, tc := range tcs {
		t.Run(tn, func(t *testing.T) {
			victim := &external{kube: tc.kube, client: tc.client}
			result, err := victim.Update(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.expected.err, err, test.EquateErrors()); diff != "" {
				t.Errorf(errorMessage, diff)
			}
			if diff := cmp.Diff(tc.expected.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf(errorMessage, diff)
			}
			if diff := cmp.Diff(tc.expected.result, result); diff != "" {
				t.Errorf(errorMessage, diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type expected struct {
		cr  resource.Managed
		err error
	}

	tcs := map[string]struct {
		args
		expected
	}{
		"NoProjectID": {
			args: args{
				cr: buildPs(),
			},
			expected: expected{
				cr:  buildPs(),
				err: errors.New(errNoProjectID),
			},
		},
		"SuccessDelete": {
			args: args{
				cr: buildPs(withExternalName(extName), withProjectID()),
				client: &fake.MockClient{
					MockDeletePipelineSchedule: func(pid interface{}, schedule int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return nil, nil
					},
				},
			},
			expected: expected{
				cr:  buildPs(withExternalName(extName), withProjectID()),
				err: nil,
			},
		},
	}

	for tn, tc := range tcs {
		t.Run(tn, func(t *testing.T) {
			victim := &external{kube: tc.kube, client: tc.client}
			err := victim.Delete(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.expected.err, err, test.EquateErrors()); diff != "" {
				t.Errorf(errorMessage, diff)
			}
			if diff := cmp.Diff(tc.expected.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf(errorMessage, diff)
			}
		})
	}
}
