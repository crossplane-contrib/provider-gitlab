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

package approvalrules

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/cluster/projects/v1alpha1"
	sharedProjectsV1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/shared/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects/fake"
	shared "github.com/crossplane-contrib/provider-gitlab/pkg/controller/shared/projects/approvalrules"
)

var (
	unexpecedItem                 resource.Managed
	errBoom                       = errors.New("boom")
	projectID                     = 0
	approvalsRequired             = 1
	users                         = []*gitlab.BasicUser{{ID: 123, Username: "abc"}, {ID: 456, Username: "testUser"}}
	groups                        = []*gitlab.Group{{ID: 99}}
	protectedBranches             = []*gitlab.ProtectedBranch{{ID: 1}, {ID: 2}}
	name                          = "name"
	ruleType                      = "any_approver"
	appliesToAllProtectedBranches = true
)

type args struct {
	projectApprovalRule projects.ApprovalRulesClient
	kube                client.Client
	cr                  resource.Managed
}

type projectModifier func(*v1alpha1.ApprovalRule)

func withConditions(c ...xpv1.Condition) projectModifier {
	return func(cr *v1alpha1.ApprovalRule) { cr.Status.Conditions = c }
}

func withProjectID() projectModifier {
	return func(r *v1alpha1.ApprovalRule) { r.Spec.ForProvider.ProjectID = &projectID }
}

func withSpec(s sharedProjectsV1alpha1.ApprovalRuleParameters) projectModifier {
	return func(r *v1alpha1.ApprovalRule) { r.Spec.ForProvider.ApprovalRuleParameters = s }
}

func withExternalName(approvalRuleId string) projectModifier {
	return func(r *v1alpha1.ApprovalRule) { meta.SetExternalName(r, approvalRuleId) }
}

func projectApprovalRule(m ...projectModifier) *v1alpha1.ApprovalRule {
	cr := &v1alpha1.ApprovalRule{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestConnect(t *testing.T) {
	type want struct {
		cr     resource.Managed
		result managed.ExternalClient
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(shared.ErrNotApprovalRule),
			},
		},
		"ProviderConfigRefNotGivenError": {
			args: args{
				cr: projectApprovalRule(),
			},
			want: want{
				cr:  projectApprovalRule(),
				err: errors.New("providerConfigRef is not given"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &connector{newGitlabClientFn: nil}
			o, err := c.Connect(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {

	type want struct {
		cr     resource.Managed
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(shared.ErrNotApprovalRule),
			},
		},
		"ExternalNameMissing": {
			args: args{
				cr: projectApprovalRule(),
			},
			want: want{
				cr: projectApprovalRule(),
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"ExternalNameNotInt": {
			args: args{
				cr: projectApprovalRule(
					withExternalName("abc"),
				),
			},
			want: want{
				cr:     projectApprovalRule(withExternalName("abc")),
				result: managed.ExternalObservation{},
				err:    errors.New(shared.ErrIDNotInt),
			},
		},
		"ErrProjectIDMissing": {
			args: args{
				cr: projectApprovalRule(withExternalName("123")),
			},
			want: want{
				cr:     projectApprovalRule(withExternalName("123")),
				result: managed.ExternalObservation{},
				err:    errors.New(shared.ErrProjectIDMissing),
			},
		},
		"ErrGet404": {
			args: args{
				projectApprovalRule: &fake.MockClient{
					MockGetProjectApprovalRule: func(pid any, ruleID int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectApprovalRule, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errBoom
					},
				},
				cr: projectApprovalRule(
					withProjectID(),
					withSpec(sharedProjectsV1alpha1.ApprovalRuleParameters{
						ApprovalsRequired: &approvalsRequired,
						ProjectID:         &projectID,
						Name:              &name,
					})),
			},
			want: want{
				cr: projectApprovalRule(
					withProjectID(),
					withSpec(sharedProjectsV1alpha1.ApprovalRuleParameters{
						ApprovalsRequired: &approvalsRequired,
						Name:              &name,
						ProjectID:         &projectID,
					})),
				result: managed.ExternalObservation{ResourceExists: false},
				err:    nil,
			},
		},
		"ErrGet": {
			args: args{
				projectApprovalRule: &fake.MockClient{
					MockGetProjectApprovalRule: func(pid any, ruleID int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectApprovalRule, *gitlab.Response, error) {
						return nil, nil, errBoom
					},
				},
				cr: projectApprovalRule(withProjectID(), withExternalName("1")),
			},
			want: want{
				cr:     projectApprovalRule(withProjectID(), withExternalName("1")),
				result: managed.ExternalObservation{},
				err:    errors.Wrap(errBoom, errors.New(shared.ErrObserveFailed).Error()),
			},
		},
		"SuccessfulAvailable": {
			args: args{
				projectApprovalRule: &fake.MockClient{
					MockGetProjectApprovalRule: func(pid any, ruleID int, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectApprovalRule, *gitlab.Response, error) {
						return &gitlab.ProjectApprovalRule{
							ApprovalsRequired: approvalsRequired,
							Name:              name,
						}, &gitlab.Response{}, nil
					},
				},
				cr: projectApprovalRule(
					withProjectID(),
					withExternalName("123"),
					withSpec(sharedProjectsV1alpha1.ApprovalRuleParameters{
						ApprovalsRequired: &approvalsRequired,
						Name:              &name,
						ProjectID:         &projectID,
					}),
				),
			},
			want: want{
				cr: projectApprovalRule(
					withConditions(xpv1.Available()),
					withProjectID(),
					withExternalName("123"),
					withSpec(sharedProjectsV1alpha1.ApprovalRuleParameters{
						ApprovalsRequired: &approvalsRequired,
						Name:              &name,
						ProjectID:         &projectID,
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: false,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &shared.External{Client: tc.projectApprovalRule, Kube: tc.kube}
			o, err := e.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type want struct {
		cr     resource.Managed
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(shared.ErrNotApprovalRule),
			},
		},
		"SuccessfulCreation": {
			args: args{
				projectApprovalRule: &fake.MockClient{
					MockCreateProjectApprovalRule: func(pid any, opt *gitlab.CreateProjectLevelRuleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectApprovalRule, *gitlab.Response, error) {
						return &gitlab.ProjectApprovalRule{
							ID:                            projectID,
							Name:                          name,
							RuleType:                      ruleType,
							ApprovalsRequired:             approvalsRequired,
							Users:                         users,
							Groups:                        groups,
							ProtectedBranches:             protectedBranches,
							AppliesToAllProtectedBranches: appliesToAllProtectedBranches,
						}, &gitlab.Response{}, nil
					},
				},
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(nil)},

				cr: projectApprovalRule(
					withSpec(sharedProjectsV1alpha1.ApprovalRuleParameters{ProjectID: &projectID}),
				),
			},
			want: want{
				cr: projectApprovalRule(
					withProjectID(),
					withExternalName(fmt.Sprintf("%d", projectID)),
					withSpec(sharedProjectsV1alpha1.ApprovalRuleParameters{ProjectID: &projectID}),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalCreation{},
			},
		},
		"FailedCreation": {
			args: args{
				projectApprovalRule: &fake.MockClient{
					MockCreateProjectApprovalRule: func(pid any, opt *gitlab.CreateProjectLevelRuleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectApprovalRule, *gitlab.Response, error) {
						return &gitlab.ProjectApprovalRule{}, &gitlab.Response{}, errBoom
					},
				},
				cr: projectApprovalRule(
					withSpec(sharedProjectsV1alpha1.ApprovalRuleParameters{ProjectID: &projectID}),
				),
			},
			want: want{
				cr: projectApprovalRule(
					withProjectID(),
					withSpec(sharedProjectsV1alpha1.ApprovalRuleParameters{ProjectID: &projectID}),
					withConditions(xpv1.Creating()),
				),
				err: errors.Wrap(errBoom, shared.ErrCreateFailed),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &shared.External{Client: tc.projectApprovalRule, Kube: tc.kube}
			o, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}

}

func TestUpdate(t *testing.T) {
	type want struct {
		cr     resource.Managed
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(shared.ErrNotApprovalRule),
			},
		},
		"SuccessfulUpdate": {
			args: args{
				projectApprovalRule: &fake.MockClient{
					MockUpdateProjectApprovalRule: func(pid any, approvalRule int, opt *gitlab.UpdateProjectLevelRuleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectApprovalRule, *gitlab.Response, error) {
						return &gitlab.ProjectApprovalRule{
							ID:                            projectID,
							Name:                          name,
							RuleType:                      ruleType,
							ApprovalsRequired:             approvalsRequired,
							Users:                         users,
							Groups:                        groups,
							ProtectedBranches:             protectedBranches,
							AppliesToAllProtectedBranches: appliesToAllProtectedBranches,
						}, &gitlab.Response{}, nil
					},
				},
				cr: projectApprovalRule(
					withProjectID(),
					withExternalName(fmt.Sprintf("%d", projectID)),
					withSpec(sharedProjectsV1alpha1.ApprovalRuleParameters{
						ApprovalsRequired: &approvalsRequired,
						Name:              &name,
						ProjectID:         &projectID,
					}),
				),
			},
			want: want{
				cr: projectApprovalRule(
					withProjectID(),
					withExternalName(fmt.Sprintf("%d", projectID)),
					withSpec(sharedProjectsV1alpha1.ApprovalRuleParameters{
						ApprovalsRequired: &approvalsRequired,
						Name:              &name,
						ProjectID:         &projectID,
					}),
				),
			},
		},
		"FailedUpdate": {
			args: args{
				projectApprovalRule: &fake.MockClient{
					MockUpdateProjectApprovalRule: func(pid any, approvalRule int, opt *gitlab.UpdateProjectLevelRuleOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProjectApprovalRule, *gitlab.Response, error) {
						return &gitlab.ProjectApprovalRule{}, &gitlab.Response{}, errBoom
					},
				},
				cr: projectApprovalRule(withProjectID(),
					withExternalName(fmt.Sprintf("%d", projectID)),
				),
			},
			want: want{
				cr: projectApprovalRule(withProjectID(),
					withExternalName(fmt.Sprintf("%d", projectID)),
				),
				err: errors.Wrap(errBoom, errors.New(shared.ErrUpdateFailed).Error()),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &shared.External{Client: tc.projectApprovalRule, Kube: tc.kube}
			o, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type want struct {
		cr  resource.Managed
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(shared.ErrNotApprovalRule),
			},
		},
		"SuccessfulDeletion": {
			args: args{
				projectApprovalRule: &fake.MockClient{
					MockDeleteProjectApprovalRule: func(pid any, approvalRule int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, nil
					},
				},
				cr: projectApprovalRule(
					withProjectID(),
					withExternalName(fmt.Sprintf("%d", projectID)),
					withSpec(sharedProjectsV1alpha1.ApprovalRuleParameters{
						ApprovalsRequired: &approvalsRequired,
						Name:              &name,
						ProjectID:         &projectID,
					})),
			},
			want: want{
				cr: projectApprovalRule(
					withExternalName(fmt.Sprintf("%d", projectID)),
					withSpec(sharedProjectsV1alpha1.ApprovalRuleParameters{
						ApprovalsRequired: &approvalsRequired,
						Name:              &name,
						ProjectID:         &projectID,
					})),
				err: nil,
			},
		},
		"FailedDeletion": {
			args: args{
				projectApprovalRule: &fake.MockClient{
					MockDeleteProjectApprovalRule: func(pid any, approvalRule int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
						return &gitlab.Response{}, errBoom
					},
				},
				cr: projectApprovalRule(
					withExternalName(fmt.Sprintf("%d", projectID)),
					withSpec(sharedProjectsV1alpha1.ApprovalRuleParameters{ProjectID: &projectID})),
			},
			want: want{
				cr: projectApprovalRule(
					withExternalName(fmt.Sprintf("%d", projectID)),
					withSpec(sharedProjectsV1alpha1.ApprovalRuleParameters{ProjectID: &projectID})),
				err: errors.Wrap(errBoom, errors.New(shared.ErrDeleteFailed).Error()),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &shared.External{Client: tc.projectApprovalRule, Kube: tc.kube}
			_, err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}

}
