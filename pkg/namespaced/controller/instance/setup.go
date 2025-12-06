package instance

import (
	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/controller/instance/settings"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		settings.SetupApplicationSettings,
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
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
