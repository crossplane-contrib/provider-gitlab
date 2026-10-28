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

package license

import (
	"context"
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
)

func observeOnly(cr interface{ SetManagementPolicies(v2.ManagementPolicies) }) {
	cr.SetManagementPolicies(v2.ManagementPolicies{v2.ManagementActionObserve})
}

func TestObserveOnlyAdoptsLicense(t *testing.T) {
	getLicense := func(...gitlab.RequestOptionFunc) (*gitlab.License, *gitlab.Response, error) {
		return &gitlab.License{ID: 7}, nil, nil
	}

	cases := map[string]struct {
		externalName string
		wantExists   bool
		wantLateInit bool
		wantName     string
	}{
		"NoExternalNameAdoptsCurrent": {wantExists: true, wantLateInit: true, wantName: "7"},
		"MatchingExternalName":        {externalName: "7", wantExists: true, wantName: "7"},
		"OtherExternalName":           {externalName: "8", wantExists: false, wantName: "8"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cr := license()
			observeOnly(cr)
			if tc.externalName != "" {
				meta.SetExternalName(cr, tc.externalName)
			}
			e := &external{client: &mockClient{MockGetLicense: getLicense}}

			got, err := e.Observe(context.Background(), cr)
			if err != nil {
				t.Fatalf("Observe() error = %v", err)
			}
			if got.ResourceExists != tc.wantExists || got.ResourceLateInitialized != tc.wantLateInit {
				t.Errorf("Observe() exists=%v lateInit=%v, want %v %v", got.ResourceExists, got.ResourceLateInitialized, tc.wantExists, tc.wantLateInit)
			}
			if diff := cmp.Diff(tc.wantName, meta.GetExternalName(cr)); diff != "" {
				t.Errorf("external name: -want +got\n%s", diff)
			}
		})
	}
}
