/*
Copyright 2024 The Crossplane Authors.

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
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/projects/accesstokens"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/projects/approvalrules"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/projects/deploykeys"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/projects/deploytokens"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/projects/hooks"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/projects/members"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/projects/pipelineschedules"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/projects/projects"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/projects/runners"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/projects/variables"
)

// Setup all project controllers
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		projects.SetupProject,
		hooks.SetupHook,
		members.SetupMember,
		deploytokens.SetupDeployToken,
		accesstokens.SetupAccessToken,
		variables.SetupVariable,
		deploykeys.SetupDeployKey,
		pipelineschedules.SetupPipelineSchedule,
		approvalrules.SetupRules,
		runners.SetupRunner,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
