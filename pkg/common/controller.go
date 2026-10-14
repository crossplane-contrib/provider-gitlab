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

package common

import (
	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"
)

// ControllerOptions configure every controller this provider sets up.
type ControllerOptions struct {
	controller.Options
}

// NewReconciler returns a managed resource reconciler for the supplied kind
// that uses the provider-wide poll interval, logger, event recorder and
// feature flags. The supplied options are applied after these.
func NewReconciler(mgr ctrl.Manager, o ControllerOptions, name string, of resource.ManagedKind, opts ...managed.ReconcilerOption) *managed.Reconciler {
	ro := []managed.ReconcilerOption{
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}
	if o.Features.Enabled(feature.EnableBetaManagementPolicies) {
		ro = append(ro, managed.WithManagementPolicies())
	}
	return managed.NewReconciler(mgr, of, append(ro, opts...)...)
}
