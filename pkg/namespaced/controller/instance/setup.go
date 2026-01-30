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

package instance

import (
	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/instance/appearance"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/instance/license"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/instance/runners"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/instance/serviceaccounts"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/instance/settings"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/instance/variables"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		settings.SetupApplicationSettings,
		runners.SetupRunner,
		appearance.SetupAppearance,
		serviceaccounts.SetupServiceAccount,
		license.SetupLicense,
		variables.SetupVariable,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupGated creates all Gitlab API controllers with the supplied logger and adds
// them to the supplied manager with CRD gate support for SafeStart.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		settings.SetupApplicationSettingsGated,
		runners.SetupRunnerGated,
		appearance.SetupAppearanceGated,
		serviceaccounts.SetupServiceAccountGated,
		license.SetupLicenseGated,
		variables.SetupVariableGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
