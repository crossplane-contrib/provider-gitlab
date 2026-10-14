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
	"context"
	"strings"
	"testing"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource/fake"
	"github.com/crossplane/crossplane-runtime/v2/pkg/statemetrics"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	namespacedgroupsv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
)

type fakeManager struct {
	fake.Manager
	runnables []manager.Runnable
}

func (m *fakeManager) GetEventRecorderFor(string) record.EventRecorder {
	return record.NewFakeRecorder(8)
}

func (m *fakeManager) Add(r manager.Runnable) error {
	m.runnables = append(m.runnables, r)
	return nil
}

func firstReconciles(t *testing.T) uint64 {
	t.Helper()
	families, err := metrics.Registry.Gather()
	if err != nil {
		t.Fatalf("metrics.Registry.Gather(): %v", err)
	}
	var n uint64
	for _, f := range families {
		if f.GetName() != "crossplane_managed_resource_first_time_to_reconcile_seconds" {
			continue
		}
		for _, m := range f.GetMetric() {
			n += m.GetHistogram().GetSampleCount()
		}
	}
	return n
}

func stateGauges(t *testing.T) map[string]float64 {
	t.Helper()
	families, err := metrics.Registry.Gather()
	if err != nil {
		t.Fatalf("metrics.Registry.Gather(): %v", err)
	}
	got := map[string]float64{}
	for _, f := range families {
		if !strings.HasPrefix(f.GetName(), "crossplane_managed_resource_") {
			continue
		}
		for _, m := range f.GetMetric() {
			for _, l := range m.GetLabel() {
				if l.GetName() == "gvk" {
					got[f.GetName()+"{"+l.GetValue()+"}"] = m.GetGauge().GetValue()
				}
			}
		}
	}
	return got
}

func TestAddMRStateRecorder(t *testing.T) {
	cases := map[string]struct {
		reason        string
		metricOptions func(*statemetrics.MRStateMetrics) *controller.MetricOptions
		wantRunnables int
		wantGauges    map[string]float64
	}{
		"NilMetricOptions": {
			reason:     "Without metric options no recorder should be added.",
			wantGauges: map[string]float64{},
		},
		"MetricOptionsWithoutStateMetrics": {
			reason: "Metric options without MRStateMetrics should add no recorder.",
			metricOptions: func(*statemetrics.MRStateMetrics) *controller.MetricOptions {
				return &controller.MetricOptions{PollStateMetricInterval: time.Millisecond}
			},
			wantGauges: map[string]float64{},
		},
		"StateMetrics": {
			reason: "With MRStateMetrics one recorder should be added that records the supplied list into those metrics.",
			metricOptions: func(sm *statemetrics.MRStateMetrics) *controller.MetricOptions {
				return &controller.MetricOptions{MRStateMetrics: sm, PollStateMetricInterval: time.Millisecond}
			},
			wantRunnables: 1,
			wantGauges: map[string]float64{
				"crossplane_managed_resource_exists{groups.gitlab.m.crossplane.io/v1alpha1, Kind=Variable}": 2,
				"crossplane_managed_resource_ready{groups.gitlab.m.crossplane.io/v1alpha1, Kind=Variable}":  1,
				"crossplane_managed_resource_synced{groups.gitlab.m.crossplane.io/v1alpha1, Kind=Variable}": 1,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			sm := statemetrics.NewMRStateMetrics()
			if err := metrics.Registry.Register(sm); err != nil {
				t.Fatalf("metrics.Registry.Register(...): %v", err)
			}
			t.Cleanup(func() { metrics.Registry.Unregister(sm) })

			s := runtime.NewScheme()
			if err := namespacedgroupsv1alpha1.SchemeBuilder.AddToScheme(s); err != nil {
				t.Fatalf("AddToScheme(...): %v", err)
			}
			ready := namespacedgroupsv1alpha1.Variable{}
			ready.SetConditions(xpv2.Available(), xpv2.ReconcileSuccess())

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			list := &namespacedgroupsv1alpha1.VariableList{}
			mgr := &fakeManager{Manager: fake.Manager{Client: &test.MockClient{
				MockScheme: test.NewMockSchemeFn(s),
				MockList: test.NewMockListFn(nil, func(client.ObjectList) error {
					list.Items = []namespacedgroupsv1alpha1.Variable{ready, {}}
					cancel()
					return nil
				}),
			}}}
			o := ControllerOptions{Options: controller.Options{Logger: logging.NewNopLogger()}}
			if tc.metricOptions != nil {
				o.MetricOptions = tc.metricOptions(sm)
			}

			if err := AddMRStateRecorder(mgr, o, list); err != nil {
				t.Fatalf("\n%s\nAddMRStateRecorder(...): unexpected error: %v", tc.reason, err)
			}
			if diff := cmp.Diff(tc.wantRunnables, len(mgr.runnables)); diff != "" {
				t.Fatalf("\n%s\nrunnables added: -want, +got:\n%s", tc.reason, diff)
			}
			for _, r := range mgr.runnables {
				if err := r.Start(ctx); err != nil {
					t.Fatalf("\n%s\nr.Start(...): unexpected error: %v", tc.reason, err)
				}
			}
			if diff := cmp.Diff(tc.wantGauges, stateGauges(t)); diff != "" {
				t.Errorf("\n%s\nstate gauges: -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestNewReconciler(t *testing.T) {
	cases := map[string]struct {
		reason              string
		annotation          string
		managementPolicies  bool
		metricOptions       func(managed.MetricRecorder) *controller.MetricOptions
		want                reconcile.Result
		wantFirstReconciles uint64
	}{
		"AnnotationBelowMinimum": {
			reason:     "An annotation below MinPollInterval should poll at MinPollInterval.",
			annotation: "2s",
			want:       reconcile.Result{RequeueAfter: 10 * time.Second},
		},
		"AnnotationAboveMinimum": {
			reason:     "An annotation above MinPollInterval should poll at the annotation value.",
			annotation: "45s",
			want:       reconcile.Result{RequeueAfter: 45 * time.Second},
		},
		"NoAnnotation": {
			reason: "A resource without the annotation should poll at PollInterval.",
			want:   reconcile.Result{RequeueAfter: 10 * time.Minute},
		},
		"ManagementPoliciesEnabled": {
			reason:             "With the management policies feature enabled, empty management policies should pause the resource.",
			managementPolicies: true,
			want:               reconcile.Result{},
		},
		"MetricRecorder": {
			reason: "The MRMetrics recorder from the metric options should record the first reconcile.",
			metricOptions: func(mr managed.MetricRecorder) *controller.MetricOptions {
				return &controller.MetricOptions{MRMetrics: mr}
			},
			want:                reconcile.Result{RequeueAfter: 10 * time.Minute},
			wantFirstReconciles: 1,
		},
		"MetricOptionsWithoutRecorder": {
			reason: "Metric options without an MRMetrics recorder should neither record nor panic.",
			metricOptions: func(managed.MetricRecorder) *controller.MetricOptions {
				return &controller.MetricOptions{}
			},
			want: reconcile.Result{RequeueAfter: 10 * time.Minute},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mr := managed.NewMRMetricRecorder()
			if err := metrics.Registry.Register(mr); err != nil {
				t.Fatalf("metrics.Registry.Register(...): %v", err)
			}
			t.Cleanup(func() { metrics.Registry.Unregister(mr) })

			mgr := &fakeManager{Manager: fake.Manager{
				Client: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						if tc.annotation != "" {
							obj.SetAnnotations(map[string]string{meta.AnnotationKeyPollInterval: tc.annotation})
						}
						return nil
					}),
					MockStatusUpdate: test.NewMockSubResourceUpdateFn(nil),
				},
				Scheme: fake.SchemeWith(&fake.ModernManaged{}),
			}}
			features := &feature.Flags{}
			if tc.managementPolicies {
				features.Enable(feature.EnableBetaManagementPolicies)
			}
			o := ControllerOptions{
				Options:         controller.Options{Logger: logging.NewNopLogger(), PollInterval: 10 * time.Minute, Features: features},
				MinPollInterval: 10 * time.Second,
			}
			if tc.metricOptions != nil {
				o.MetricOptions = tc.metricOptions(mr)
			}

			r := NewReconciler(mgr, o, "test", resource.ManagedKind(fake.GVK(&fake.ModernManaged{})),
				managed.WithInitializers(),
				managed.WithFinalizer(resource.FinalizerFns{AddFinalizerFn: func(context.Context, resource.Object) error { return nil }}),
				managed.WithExternalConnector(managed.ExternalConnectorFn(func(context.Context, resource.Managed) (managed.ExternalClient, error) {
					return &managed.ExternalClientFns{
						ObserveFn: func(context.Context, resource.Managed) (managed.ExternalObservation, error) {
							return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
						},
						DisconnectFn: func(context.Context) error { return nil },
					}, nil
				})),
			)

			got, err := r.Reconcile(context.Background(), reconcile.Request{})
			if err != nil {
				t.Fatalf("\n%s\nr.Reconcile(...): unexpected error: %v", tc.reason, err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("\n%s\nr.Reconcile(...): -want, +got:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.wantFirstReconciles, firstReconciles(t)); diff != "" {
				t.Errorf("\n%s\nfirst reconciles recorded: -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}
