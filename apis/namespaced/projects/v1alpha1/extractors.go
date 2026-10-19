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
	"strconv"

	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reference"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
)

// ProjectID extracts the numeric ID of a Project, also when it is named by full path.
func ProjectID() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		name := meta.GetExternalName(mg)
		if _, err := strconv.ParseInt(name, 10, 64); err == nil {
			return name
		}
		// Assert via any: during code generation the cluster-scoped copy
		// briefly lacks the managed methods.
		p, ok := any(mg).(*Project)
		if !ok || p.Status.AtProvider.ID == 0 {
			return ""
		}
		return strconv.FormatInt(p.Status.AtProvider.ID, 10)
	}
}
