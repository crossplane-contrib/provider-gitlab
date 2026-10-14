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
	"testing"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource/fake"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type fakeManager struct {
	fake.Manager
}

func (m *fakeManager) GetEventRecorderFor(string) record.EventRecorder {
	return record.NewFakeRecorder(8)
}

func TestNewReconciler(t *testing.T) {
	cases := map[string]struct {
		reason             string
		annotation         string
		managementPolicies bool
		want               reconcile.Result
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
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
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
		})
	}
}
