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

package projects

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
	"k8s.io/utils/ptr"

	commonv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
)

// stubProjectSAClient is a minimal ServiceAccountClient for exercising
// GetProjectServiceAccount pagination.
type stubProjectSAClient struct {
	list func(pid any, opt *gitlab.ListProjectServiceAccountsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectServiceAccount, *gitlab.Response, error)
}

func (s *stubProjectSAClient) ListProjectServiceAccounts(pid any, opt *gitlab.ListProjectServiceAccountsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectServiceAccount, *gitlab.Response, error) {
	return s.list(pid, opt, options...)
}

func (s *stubProjectSAClient) CreateProjectServiceAccount(_ any, _ *gitlab.CreateProjectServiceAccountOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.ProjectServiceAccount, *gitlab.Response, error) {
	return nil, nil, nil
}

func (s *stubProjectSAClient) UpdateProjectServiceAccount(_ any, _ int64, _ *gitlab.UpdateProjectServiceAccountOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.ProjectServiceAccount, *gitlab.Response, error) {
	return nil, nil, nil
}

func (s *stubProjectSAClient) DeleteProjectServiceAccount(_ any, _ int64, _ *gitlab.DeleteProjectServiceAccountOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return nil, nil
}

func TestGenerateServiceAccountObservation(t *testing.T) {
	got := GenerateServiceAccountObservation(&gitlab.ProjectServiceAccount{
		ID: 7, Name: "n", Username: "u", Email: "e@example.com",
	})
	want := v1alpha1.ServiceAccountObservation{
		CommonServiceAccountObservation: commonv1alpha1.CommonServiceAccountObservation{
			ID: 7, Name: "n", Username: "u", Email: "e@example.com",
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("observation: -want, +got:\n%s", diff)
	}
	if diff := cmp.Diff(v1alpha1.ServiceAccountObservation{}, GenerateServiceAccountObservation(nil)); diff != "" {
		t.Errorf("nil: -want, +got:\n%s", diff)
	}
}

func TestGenerateServiceAccountOptions(t *testing.T) {
	p := &v1alpha1.ServiceAccountParameters{}
	p.Name = ptr.To("n")
	p.Username = ptr.To("u")
	p.Email = ptr.To("e@example.com")

	c := GenerateServiceAccountCreateOptions(p)
	if *c.Name != "n" || *c.Username != "u" || *c.Email != "e@example.com" {
		t.Errorf("create options mismatch: %+v", c)
	}
	u := GenerateUpdateServiceAccountOptions(p)
	if *u.Name != "n" || *u.Username != "u" || *u.Email != "e@example.com" {
		t.Errorf("update options mismatch: %+v", u)
	}
}

func TestIsServiceAccountUpToDate(t *testing.T) {
	p := &v1alpha1.ServiceAccountParameters{}
	p.Name = ptr.To("n")
	p.Username = ptr.To("u")
	p.Email = ptr.To("e@example.com")

	cases := map[string]struct {
		p    *v1alpha1.ServiceAccountParameters
		a    *gitlab.ProjectServiceAccount
		want bool
	}{
		"NilParams":  {p: nil, a: &gitlab.ProjectServiceAccount{}, want: true},
		"NilAccount": {p: p, a: nil, want: false},
		"UpToDate":   {p: p, a: &gitlab.ProjectServiceAccount{Name: "n", Username: "u", Email: "e@example.com"}, want: true},
		"NameDrift":  {p: p, a: &gitlab.ProjectServiceAccount{Name: "x", Username: "u", Email: "e@example.com"}, want: false},
		"EmailDrift": {p: p, a: &gitlab.ProjectServiceAccount{Name: "n", Username: "u", Email: "z@example.com"}, want: false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := IsServiceAccountUpToDate(tc.p, tc.a); got != tc.want {
				t.Errorf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestGetProjectServiceAccount(t *testing.T) {
	errBoom := errors.New("boom")
	cases := map[string]struct {
		list    func(pid any, opt *gitlab.ListProjectServiceAccountsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectServiceAccount, *gitlab.Response, error)
		id      int64
		wantID  int64
		wantNil bool
		wantErr bool
	}{
		"FoundFirstPage": {
			list: func(_ any, _ *gitlab.ListProjectServiceAccountsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectServiceAccount, *gitlab.Response, error) {
				return []*gitlab.ProjectServiceAccount{{ID: 1}, {ID: 2}}, &gitlab.Response{NextPage: 0}, nil
			},
			id: 2, wantID: 2,
		},
		"FoundSecondPage": {
			list: func() func(any, *gitlab.ListProjectServiceAccountsOptions, ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectServiceAccount, *gitlab.Response, error) {
				calls := 0
				return func(_ any, _ *gitlab.ListProjectServiceAccountsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectServiceAccount, *gitlab.Response, error) {
					calls++
					if calls == 1 {
						return []*gitlab.ProjectServiceAccount{{ID: 1}}, &gitlab.Response{NextPage: 2}, nil
					}
					return []*gitlab.ProjectServiceAccount{{ID: 42}}, &gitlab.Response{NextPage: 0}, nil
				}
			}(),
			id: 42, wantID: 42,
		},
		"NotFound": {
			list: func(_ any, _ *gitlab.ListProjectServiceAccountsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectServiceAccount, *gitlab.Response, error) {
				return []*gitlab.ProjectServiceAccount{{ID: 1}}, &gitlab.Response{NextPage: 0}, nil
			},
			id: 99, wantNil: true,
		},
		"Error": {
			list: func(_ any, _ *gitlab.ListProjectServiceAccountsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectServiceAccount, *gitlab.Response, error) {
				return nil, nil, errBoom
			},
			id: 1, wantErr: true,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, _, err := GetProjectServiceAccount(&stubProjectSAClient{list: tc.list}, 1, tc.id)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %+v", got)
				}
				return
			}
			if got == nil || got.ID != tc.wantID {
				t.Fatalf("want id %d, got %+v", tc.wantID, got)
			}
		})
	}
}
