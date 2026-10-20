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

// +cluster-scope:skip-file

package config

import (
	"os"
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
)

type gvkRecordingGate struct {
	gvks []schema.GroupVersionKind
}

func (g *gvkRecordingGate) Register(_ func(), gvks ...schema.GroupVersionKind) {
	g.gvks = append(g.gvks, gvks...)
}

func (g *gvkRecordingGate) Set(schema.GroupVersionKind, bool) bool { return false }

// The ProviderConfig reconciler must operate on the CRD storage version.
// Writing through any other served version rewrites the stored spec on every
// update, bumps metadata.generation and makes the controller reconcile itself
// in a hot loop.
func TestProviderConfigControllerUsesStorageVersion(t *testing.T) {
	f, err := os.Open("../../../../package/crds/gitlab.m.crossplane.io_providerconfigs.yaml")
	if err != nil {
		t.Fatalf("os.Open(...) error = %v", err)
	}
	defer f.Close()

	crd := &apiextensionsv1.CustomResourceDefinition{}
	if err := yaml.NewYAMLOrJSONDecoder(f, 4096).Decode(crd); err != nil {
		t.Fatalf("Decode(...) error = %v", err)
	}
	var storage string
	for _, v := range crd.Spec.Versions {
		if v.Storage {
			storage = v.Name
		}
	}

	g := &gvkRecordingGate{}
	o := common.ControllerOptions{Options: controller.Options{Logger: logging.NewNopLogger(), Gate: g}}
	if err := SetupNamespacedGated(nil, o); err != nil {
		t.Fatalf("SetupNamespacedGated(...) error = %v", err)
	}

	var got []string
	for _, gvk := range g.gvks {
		if gvk.Kind == crd.Spec.Names.Kind {
			got = append(got, gvk.Version)
		}
	}
	if len(got) != 1 || got[0] != storage {
		t.Errorf("ProviderConfig controller gated on versions %v, want [%s] (CRD storage version)", got, storage)
	}
}
