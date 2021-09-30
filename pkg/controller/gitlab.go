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

	"github.com/crossplane/crossplane-runtime/pkg/logging"

	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/config"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/groups"
	groupDeployToken "github.com/crossplane-contrib/provider-gitlab/pkg/controller/groups/deploytokens"
	groupmembers "github.com/crossplane-contrib/provider-gitlab/pkg/controller/groups/members"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/projects"
	projectDeployToken "github.com/crossplane-contrib/provider-gitlab/pkg/controller/projects/deploytokens"
	projecthooks "github.com/crossplane-contrib/provider-gitlab/pkg/controller/projects/hooks"
	projectmembers "github.com/crossplane-contrib/provider-gitlab/pkg/controller/projects/members"
	"github.com/crossplane-contrib/provider-gitlab/pkg/controller/projects/variables"
)

// Setup creates all Gitlab API controllers with the supplied logger and adds
// them to the supplied manager.
func Setup(mgr ctrl.Manager, l logging.Logger) error {
	for _, setup := range []func(ctrl.Manager, logging.Logger) error{
		config.Setup,
		groups.SetupGroup,
		groupmembers.SetupMember,
		groupDeployToken.SetupDeployToken,
		projects.SetupProject,
		projecthooks.SetupHook,
		projectmembers.SetupMember,
		projectDeployToken.SetupDeployToken,
		variables.SetupVariable,
	} {
		if err := setup(mgr, l); err != nil {
			return err
		}
	}
	return nil
}
