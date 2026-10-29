/*
Copyright 2026 The Crossplane Authors.

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

package license

import (
	"context"
	"strings"
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpfake "github.com/crossplane/crossplane-runtime/v2/pkg/resource/fake"
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/instance/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
)

type policyTestManager struct{ xpfake.Manager }

func (m *policyTestManager) GetEventRecorderFor(string) record.EventRecorder {
	return record.NewFakeRecorder(32)
}

func TestLicensePolicyReconcile(t *testing.T) {
	cases := map[string]struct {
		policies     v2.ManagementPolicies
		externalName string
		wantName     string
		wantID       int64
		wantReady    corev1.ConditionStatus
		wantSynced   corev1.ConditionStatus
		wantErr      string
		replace      bool
		wantDeletes  int
	}{
		"ObserveCurrent":       {policies: v2.ManagementPolicies{v2.ManagementActionObserve}, wantID: 8, wantReady: corev1.ConditionTrue, wantSynced: corev1.ConditionTrue, replace: true},
		"LateInitializePinsID": {policies: v2.ManagementPolicies{v2.ManagementActionObserve, v2.ManagementActionLateInitialize}, wantName: "7", wantID: 7, wantReady: corev1.ConditionFalse, wantSynced: corev1.ConditionTrue, replace: true},
		"ObserveExplicitID":    {policies: v2.ManagementPolicies{v2.ManagementActionObserve}, externalName: "7", wantName: "7", wantID: 7, wantReady: corev1.ConditionFalse, wantSynced: corev1.ConditionFalse, replace: true},
		"DeleteRequiresID":     {policies: v2.ManagementPolicies{v2.ManagementActionObserve, v2.ManagementActionDelete}, wantErr: errObserveDeleteNeedsID},
		"DeleteExplicitID":     {policies: v2.ManagementPolicies{v2.ManagementActionObserve, v2.ManagementActionDelete}, externalName: "7", wantName: "7", wantID: 7, wantReady: corev1.ConditionTrue, wantSynced: corev1.ConditionTrue, wantDeletes: 1},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			if err := v1alpha1.SchemeBuilder.AddToScheme(scheme); err != nil {
				t.Fatal(err)
			}
			cr := license(withNamespace("default"), withExternalName(tc.externalName))
			cr.Name = "license-policy-test"
			cr.SetManagementPolicies(tc.policies)
			kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cr).WithStatusSubresource(cr).Build()
			mgr := &policyTestManager{Manager: xpfake.Manager{Client: kube, Scheme: scheme}}
			currentID := int64(7)
			deletes := 0
			ext := &external{managementPoliciesEnabled: true, client: &mockClient{
				MockGetLicense: func(...gitlab.RequestOptionFunc) (*gitlab.License, *gitlab.Response, error) {
					return &gitlab.License{ID: currentID}, nil, nil
				},
				MockDeleteLicense: func(id int64, _ ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
					if id != 7 {
						t.Errorf("deleted license %d, want 7", id)
					}
					deletes++
					return nil, nil
				},
			}}
			features := &feature.Flags{}
			features.Enable(feature.EnableBetaManagementPolicies)
			r := common.NewReconciler(mgr, common.ControllerOptions{Options: controller.Options{Features: features, Logger: logging.NewNopLogger()}}, "license-policy-test", resource.ManagedKind(v1alpha1.LicenseGroupVersionKind),
				managed.WithInitializers(),
				managed.WithExternalConnector(managed.ExternalConnectorFn(func(context.Context, resource.Managed) (managed.ExternalClient, error) { return ext, nil })),
			)
			key := types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}
			reconcileOnce := func() {
				t.Helper()
				if _, err := r.Reconcile(context.Background(), reconcile.Request{NamespacedName: key}); err != nil {
					t.Fatal(err)
				}
			}
			for range 2 {
				reconcileOnce()
			}
			if tc.replace {
				currentID = 8
				reconcileOnce()
			}
			got := &v1alpha1.License{}
			if err := kube.Get(context.Background(), key, got); err != nil {
				t.Fatal(err)
			}
			if meta.GetExternalName(got) != tc.wantName || got.Status.AtProvider.ID != tc.wantID {
				t.Fatalf("name=%q ID=%d, want name=%q ID=%d", meta.GetExternalName(got), got.Status.AtProvider.ID, tc.wantName, tc.wantID)
			}
			synced := got.GetCondition(v2.TypeSynced)
			if tc.wantErr != "" {
				if synced.Status != corev1.ConditionFalse || !strings.Contains(synced.Message, tc.wantErr) {
					t.Fatalf("want reconciliation error %q, got %+v", tc.wantErr, synced)
				}
			} else {
				if synced.Status != tc.wantSynced || got.GetCondition(v2.TypeReady).Status != tc.wantReady {
					t.Fatalf("want Ready=%s and Synced=%s, got %+v", tc.wantReady, tc.wantSynced, got.Status.Conditions)
				}
			}
			if err := kube.Delete(context.Background(), got); err != nil {
				t.Fatal(err)
			}
			reconcileOnce()
			if deletes != tc.wantDeletes {
				t.Errorf("DeleteLicense called %d times, want %d", deletes, tc.wantDeletes)
			}
			if err := kube.Get(context.Background(), key, &v1alpha1.License{}); !apierrors.IsNotFound(err) {
				t.Fatalf("managed resource was not finalized: %v", err)
			}
		})
	}
}

func TestObservePoliciesDisabled(t *testing.T) {
	cr := license()
	cr.SetManagementPolicies(v2.ManagementPolicies{v2.ManagementActionObserve})
	e := &external{}
	got, err := e.Observe(context.Background(), cr)
	if err != nil || got.ResourceExists || meta.GetExternalName(cr) != "" {
		t.Fatalf("disabled policy feature must keep the create path: got=%+v err=%v", got, err)
	}
}
