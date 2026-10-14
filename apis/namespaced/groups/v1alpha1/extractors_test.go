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
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"k8s.io/utils/ptr"
)

func TestGroupID(t *testing.T) {
	group := func(externalName string, id *int64) *Group {
		g := &Group{}
		meta.SetExternalName(g, externalName)
		g.Status.AtProvider.ID = id
		return g
	}

	cases := map[string]struct {
		mg   resource.Managed
		want string
	}{
		"IDExternalName": {
			mg:   group("1234", nil),
			want: "1234",
		},
		"PathExternalName": {
			mg:   group("path/to/example-group", ptr.To(int64(1234))),
			want: "1234",
		},
		"PathExternalNameNotObserved": {
			mg:   group("path/to/example-group", nil),
			want: "",
		},
		"NoExternalName": {
			mg:   &Group{},
			want: "",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if diff := cmp.Diff(tc.want, GroupID()(tc.mg)); diff != "" {
				t.Errorf("GroupID() diff (-want +got):\n%s", diff)
			}
		})
	}
}
