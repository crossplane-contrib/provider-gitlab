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
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/accesstokens"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/approvalrules"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/badges"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/deploykeys"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/deploytokens"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/hooks"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/integrationcustomissuetracker"
	integrationemailsonpush "github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/integrationemailsonpush"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/integrationexternalwiki"
	integrationgithub "github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/integrationgithub"
	integrationharbor "github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/integrationharbor"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/integrationjenkins"
	integrationmattermost "github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/integrationmattermost"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/integrationmicrosoftteams"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/integrationpipelinesemail"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/integrationredmine"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/integrationtelegram"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/integrationyoutrack"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/members"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/pipelineschedules"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/projects"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/projectsharegroups"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/protectedbranches"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/protectedenvironments"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/protectedtags"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/repositoryfiles"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/runners"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/serviceaccountaccesstokens"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/serviceaccounts"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/projects/variables"
)

// Setup all project controllers
func Setup(mgr ctrl.Manager, o common.ControllerOptions) error {
	for _, setup := range []func(ctrl.Manager, common.ControllerOptions) error{
		projects.SetupProject,
		hooks.SetupHook,
		members.SetupMember,
		deploytokens.SetupDeployToken,
		accesstokens.SetupAccessToken,
		variables.SetupVariable,
		repositoryfiles.SetupRepositoryFile,
		deploykeys.SetupDeployKey,
		pipelineschedules.SetupPipelineSchedule,
		approvalrules.SetupRules,
		runners.SetupRunner,
		protectedbranches.SetupProtectedBranch,
		protectedtags.SetupProtectedTag,
		badges.SetupBadge,
		integrationmattermost.SetupIntegrationMattermost,
		integrationemailsonpush.SetupIntegrationEmailsOnPush,
		integrationharbor.SetupIntegrationHarbor,
		integrationjenkins.SetupIntegrationJenkins,
		integrationmicrosoftteams.SetupIntegrationMicrosoftTeams,
		integrationtelegram.SetupIntegrationTelegram,
		integrationpipelinesemail.SetupIntegrationPipelinesEmail,
		integrationgithub.SetupIntegrationGithub,
		protectedenvironments.SetupProtectedEnvironment,
		projectsharegroups.SetupProjectShareGroup,
		serviceaccounts.SetupServiceAccount,
		serviceaccountaccesstokens.SetupServiceAccountAccessToken,
		integrationcustomissuetracker.SetupIntegrationCustomIssueTracker,
		integrationexternalwiki.SetupIntegrationExternalWiki,
		integrationredmine.SetupIntegrationRedmine,
		integrationyoutrack.SetupIntegrationYouTrack,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupGated creates all Gitlab API controllers with the supplied logger and adds
// them to the supplied manager with CRD gate support for SafeStart.
func SetupGated(mgr ctrl.Manager, o common.ControllerOptions) error {
	for _, setup := range []func(ctrl.Manager, common.ControllerOptions) error{
		projects.SetupProjectGated,
		hooks.SetupHookGated,
		members.SetupMemberGated,
		deploytokens.SetupDeployTokenGated,
		accesstokens.SetupAccessTokenGated,
		variables.SetupVariableGated,
		repositoryfiles.SetupRepositoryFileGated,
		deploykeys.SetupDeployKeyGated,
		pipelineschedules.SetupPipelineScheduleGated,
		approvalrules.SetupRulesGated,
		runners.SetupRunnerGated,
		protectedbranches.SetupProtectedBranchGated,
		protectedtags.SetupProtectedTagGated,
		badges.SetupBadgeGated,
		integrationmattermost.SetupIntegrationMattermostGated,
		integrationemailsonpush.SetupIntegrationEmailsOnPushGated,
		integrationharbor.SetupIntegrationHarborGated,
		integrationjenkins.SetupIntegrationJenkinsGated,
		integrationmicrosoftteams.SetupIntegrationMicrosoftTeamsGated,
		integrationtelegram.SetupIntegrationTelegramGated,
		integrationpipelinesemail.SetupIntegrationPipelinesEmailGated,
		integrationgithub.SetupIntegrationGithubGated,
		protectedenvironments.SetupProtectedEnvironmentGated,
		projectsharegroups.SetupProjectShareGroupGated,
		serviceaccounts.SetupServiceAccountGated,
		serviceaccountaccesstokens.SetupServiceAccountAccessTokenGated,
		integrationcustomissuetracker.SetupIntegrationCustomIssueTrackerGated,
		integrationexternalwiki.SetupIntegrationExternalWikiGated,
		integrationredmine.SetupIntegrationRedmineGated,
		integrationyoutrack.SetupIntegrationYouTrackGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
