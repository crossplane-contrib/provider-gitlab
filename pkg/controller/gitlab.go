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

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/pkg/controller"

	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/config"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/groups"
	groupsDeployToken "github.com/crossplane-contrib/provider-gitlab/pkg/controller/groups/deploytokens"
	groupsMembers "github.com/crossplane-contrib/provider-gitlab/pkg/controller/groups/members"
	groupsVariables "github.com/crossplane-contrib/provider-gitlab/pkg/controller/groups/variables"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/projects"
	projectsAccessToken "github.com/crossplane-contrib/provider-gitlab/pkg/controller/projects/accesstokens"
	projectsDeployKeys "github.com/crossplane-contrib/provider-gitlab/pkg/controller/projects/deploykeys"
	projectsDeployToken "github.com/crossplane-contrib/provider-gitlab/pkg/controller/projects/deploytokens"
	projectsHooks "github.com/crossplane-contrib/provider-gitlab/pkg/controller/projects/hooks"
	projectsMembers "github.com/crossplane-contrib/provider-gitlab/pkg/controller/projects/members"
	projectsPipelineschedules "github.com/crossplane-contrib/provider-gitlab/pkg/controller/projects/pipelineschedules"
	projectsVariables "github.com/crossplane-contrib/provider-gitlab/pkg/controller/projects/variables"
)

// Setup creates all Gitlab API controllers with the supplied logger and adds
// them to the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		config.Setup,
		groups.SetupGroup,
		groupsMembers.SetupMember,
		groupsDeployToken.SetupDeployToken,
		groupsVariables.SetupVariable,
		projects.SetupProject,
		projectsHooks.SetupHook,
		projectsMembers.SetupMember,
		projectsDeployToken.SetupDeployToken,
		projectsAccessToken.SetupAccessToken,
		projectsVariables.SetupVariable,
		projectsDeployKeys.SetupDeployKey,
		projectsPipelineschedules.SetupPipelineSchedule,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
