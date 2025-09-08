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
	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/namespaced/groups/accesstokens"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/namespaced/groups/deploytokens"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/namespaced/groups/groups"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/namespaced/groups/members"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/namespaced/groups/runners"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/namespaced/groups/samlgrouplinks"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/namespaced/groups/variables"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		groups.SetupGroup,
		members.SetupMember,
		accesstokens.Setup,
		deploytokens.SetupDeployToken,
		variables.SetupVariable,
		samlgrouplinks.SetupSamlGroupLink,
		runners.SetupRunner,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		groups.SetupGroupGated,
		members.SetupMemberGated,
		accesstokens.SetupGated,
		deploytokens.SetupDeployTokenGated,
		variables.SetupVariableGated,
		samlgrouplinks.SetupSamlGroupLinkGated,
		runners.SetupRunnerGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
