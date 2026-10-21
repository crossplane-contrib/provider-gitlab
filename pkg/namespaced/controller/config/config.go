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

package config

import (
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/providerconfig"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/v1beta1"
	// +cluster-scope:delete=1
	pcstorage "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/v1beta2"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
)

// Setup adds a controller that reconciles ProviderConfigs by accounting for
// their current usage.
// +cluster-scope:delete=11
//
// The reconciler reads and writes the ProviderConfig in its storage version
// (pcstorage). Round-tripping it through the served-but-deprecated v1beta1
// rewrites spec fields that v1beta2 does not have, so every finalizer update
// bumps metadata.generation, which retriggers the watch in a hot loop.
//
// ProviderConfigUsage exists only in namespaced v1beta1: v1beta2 did not fork
// Usage or ClusterProviderConfig, so usage accounting stays on v1beta1. When
// v1beta1 is eventually removed, the Usage references must move to the
// then-current version, otherwise the gate never fires and this controller
// silently never starts.
func SetupNamespaced(mgr ctrl.Manager, o common.ControllerOptions) error {
	name := providerconfig.ControllerName(pcstorage.ProviderConfigGroupKind)

	of := resource.ProviderConfigKinds{
		Config:    pcstorage.ProviderConfigGroupVersionKind,
		Usage:     v1beta1.ProviderConfigUsageGroupVersionKind,
		UsageList: v1beta1.ProviderConfigUsageListGroupVersionKind,
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&pcstorage.ProviderConfig{}).
		Watches(&v1beta1.ProviderConfigUsage{}, &resource.EnqueueRequestForProviderConfig{}).
		Complete(providerconfig.NewReconciler(mgr, of,
			providerconfig.WithLogger(o.Logger.WithValues("controller", name)),
			providerconfig.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func SetupNamespacedGated(mgr ctrl.Manager, o common.ControllerOptions) error {
	o.Gate.Register(func() {
		if err := SetupNamespaced(mgr, o); err != nil {
			mgr.GetLogger().Error(err, "unable to setup reconcilers", "gvk", pcstorage.ProviderConfigGroupVersionKind.String())
		}
	}, pcstorage.ProviderConfigGroupVersionKind, v1beta1.ProviderConfigUsageGroupVersionKind)
	return nil
}
