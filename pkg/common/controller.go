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
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/statemetrics"
	ctrl "sigs.k8s.io/controller-runtime"
)

// ControllerOptions configure every controller this provider sets up.
type ControllerOptions struct {
	controller.Options

	// MinPollInterval is the shortest poll interval a managed resource may
	// request with the crossplane.io/poll-interval annotation. Shorter
	// requests are raised to this value.
	MinPollInterval time.Duration
}

// NewReconciler returns a managed resource reconciler for the supplied kind
// that uses the provider-wide poll intervals, logger, event recorder, metric
// recorder and feature flags. The supplied options are applied after these.
func NewReconciler(mgr ctrl.Manager, o ControllerOptions, name string, of resource.ManagedKind, opts ...managed.ReconcilerOption) *managed.Reconciler {
	ro := []managed.ReconcilerOption{
		managed.WithPollInterval(o.PollInterval),
		managed.WithMinPollInterval(o.MinPollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}
	if o.MetricOptions != nil && o.MetricOptions.MRMetrics != nil {
		ro = append(ro, managed.WithMetricRecorder(o.MetricOptions.MRMetrics))
	}
	if o.Features.Enabled(feature.EnableBetaManagementPolicies) {
		ro = append(ro, managed.WithManagementPolicies())
	}
	return managed.NewReconciler(mgr, of, append(ro, opts...)...)
}

// AddMRStateRecorder adds a runnable to mgr that records the exists, ready and
// synced metrics of the managed resources in list. It adds nothing when the
// options carry no MRStateMetrics.
func AddMRStateRecorder(mgr ctrl.Manager, o ControllerOptions, list resource.ManagedList) error {
	if o.MetricOptions == nil || o.MetricOptions.MRStateMetrics == nil {
		return nil
	}
	return mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, list, o.MetricOptions.PollStateMetricInterval))
}
