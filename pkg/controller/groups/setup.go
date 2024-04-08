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

package groups

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/pkg/controller"

	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/groups/accesstokens"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/groups/deploytokens"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/groups/groups"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/groups/members"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/groups/variables"
)

// Setup all group controllers
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		groups.SetupGroup,
		members.SetupMember,
		accesstokens.SetupAccessToken,
		deploytokens.SetupDeployToken,
		variables.SetupVariable,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
