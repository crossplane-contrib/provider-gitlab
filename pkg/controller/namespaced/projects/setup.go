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
	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/namespaced/projects/accesstokens"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/namespaced/projects/approvalrules"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/namespaced/projects/deploykeys"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/namespaced/projects/deploytokens"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/namespaced/projects/hooks"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/namespaced/projects/members"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/namespaced/projects/pipelineschedules"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/namespaced/projects/projects"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/namespaced/projects/runners"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/namespaced/projects/variables"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		projects.SetupProject,
		accesstokens.SetupAccessToken,
		approvalrules.SetupRules,
		deploykeys.SetupDeployKey,
		deploytokens.SetupDeployToken,
		hooks.SetupHook,
		members.SetupMember,
		pipelineschedules.SetupPipelineSchedule,
		runners.SetupRunner,
		variables.SetupVariable,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		projects.SetupProjectGated,
		accesstokens.SetupAccessTokenGated,
		approvalrules.SetupRulesGated,
		deploykeys.SetupDeployKeyGated,
		deploytokens.SetupDeployTokenGated,
		hooks.SetupHookGated,
		members.SetupMemberGated,
		pipelineschedules.SetupPipelineScheduleGated,
		runners.SetupRunnerGated,
		variables.SetupVariableGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
