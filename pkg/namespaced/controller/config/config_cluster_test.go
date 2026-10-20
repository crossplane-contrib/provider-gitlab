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

// +cluster-scope:skip-file

package config

import (
	"strings"
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/v1beta1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
)

type recordingGate struct {
	callbacks []func()
}

func (g *recordingGate) Register(callback func(), _ ...schema.GroupVersionKind) {
	g.callbacks = append(g.callbacks, callback)
}

func (g *recordingGate) Set(schema.GroupVersionKind, bool) bool { return false }

func TestSetupGatedRegistersBothControllers(t *testing.T) {
	s := runtime.NewScheme()
	if err := v1beta1.SchemeBuilder.AddToScheme(s); err != nil {
		t.Fatalf("AddToScheme(...) error = %v", err)
	}
	mgr, err := ctrl.NewManager(&rest.Config{Host: "https://127.0.0.1:1"}, ctrl.Options{
		Scheme:  s,
		Metrics: metricsserver.Options{BindAddress: "0"},
	})
	if err != nil {
		t.Fatalf("NewManager(...) error = %v", err)
	}
	g := &recordingGate{}
	o := common.ControllerOptions{Options: controller.Options{Logger: logging.NewNopLogger(), Gate: g}}

	if err := SetupGated(mgr, o); err != nil {
		t.Fatalf("SetupGated(...) error = %v", err)
	}
	// The ProviderConfig gate opens first. The ClusterProviderConfig gate must
	// still register its own controller.
	for _, open := range g.callbacks {
		open()
	}

	// controller-runtime rejects a second controller with a name in use, so
	// each setup fails only if its gate already registered the controller.
	cases := map[string]struct {
		setup func(ctrl.Manager, common.ControllerOptions) error
		want  string
	}{
		"ProviderConfig": {
			setup: SetupNamespaced,
			want:  "controller with name providerconfig/providerconfig.gitlab.m.crossplane.io already exists",
		},
		"ClusterProviderConfig": {
			setup: SetupCluster,
			want:  "controller with name providerconfig/clusterproviderconfig.gitlab.m.crossplane.io already exists",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if err := tc.setup(mgr, o); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("setup after gate opened: error = %v, want error containing %q", err, tc.want)
			}
		})
	}
}
