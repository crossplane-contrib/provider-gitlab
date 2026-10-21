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

package v1beta2

import (
	"os"
	"sort"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const providerConfigCRD = "../../../package/crds/gitlab.m.crossplane.io_providerconfigs.yaml"

// The ProviderConfig CRD converts between its served versions with strategy
// None, which is a no-op: the stored object is reinterpreted under the target
// schema and unknown fields are pruned. That is only safe while the storage
// version can represent every field of each other served version. Otherwise
// the round trip through storage loses data.
//
// The API server performs that round trip whenever it merges managed fields,
// so a lossy conversion makes server-side apply on the non-storage version
// fail as soon as a second field manager owns part of the object:
//
//	failed to prune fields: failed add back owned items: failed to convert
//	merged object at version gitlab.m.crossplane.io/v1beta2:
//	.spec.credentials.secretRef.namespace: field not declared in schema
//
// That is the normal state for a ProviderConfig rendered by a Composition:
// Crossplane applies composed resources server-side, and the provider owns the
// finalizer on the storage version.
//
// Fields this version deliberately ignores (secretRef.namespace, fs, env) must
// therefore stay declared. UseProvicerConfig must not read them;
// TestUseProviderConfigResolvesLocalSecretRef in pkg/common covers that.
func TestStorageVersionCanRepresentEveryServedVersion(t *testing.T) {
	crd := loadCRD(t)

	if s := crd.Spec.Conversion; s != nil && s.Strategy != apiextensionsv1.NoneConverter {
		t.Skipf("conversion strategy is %q, not %q: pruning no longer applies", s.Strategy, apiextensionsv1.NoneConverter)
	}

	storageName, storage := storageVersion(t, crd)

	for _, v := range crd.Spec.Versions {
		if v.Name == storageName || !v.Served {
			continue
		}
		missing := fieldsMissingFrom(storage, v.Schema.OpenAPIV3Schema, "")
		if len(missing) > 0 {
			sort.Strings(missing)
			t.Errorf("storage version %s cannot represent served version %s, these fields would be pruned on conversion: %v",
				storageName, v.Name, missing)
		}
	}
}

// fieldsMissingFrom returns the paths present in want but absent from got.
// Object properties recurse by name and arrays through their item schema, which
// together cover every path a CRD structural schema can address.
func fieldsMissingFrom(got, want *apiextensionsv1.JSONSchemaProps, path string) []string {
	if got == nil || want == nil {
		return nil
	}

	var missing []string
	for name, wantProp := range want.Properties {
		child := path + "." + name
		gotProp, ok := got.Properties[name]
		if !ok {
			// x-kubernetes-preserve-unknown-fields keeps everything below it,
			// so nothing under this path is pruned.
			if ptrBool(got.XPreserveUnknownFields) {
				continue
			}
			missing = append(missing, child)
			continue
		}
		missing = append(missing, fieldsMissingFrom(&gotProp, &wantProp, child)...)
	}

	if want.Items != nil && want.Items.Schema != nil && got.Items != nil && got.Items.Schema != nil {
		missing = append(missing, fieldsMissingFrom(got.Items.Schema, want.Items.Schema, path+"[]")...)
	}

	return missing
}

func storageVersion(t *testing.T, crd *apiextensionsv1.CustomResourceDefinition) (string, *apiextensionsv1.JSONSchemaProps) {
	t.Helper()

	for _, v := range crd.Spec.Versions {
		if v.Storage {
			return v.Name, v.Schema.OpenAPIV3Schema
		}
	}
	t.Fatalf("%s declares no storage version", providerConfigCRD)

	return "", nil
}

func loadCRD(t *testing.T) *apiextensionsv1.CustomResourceDefinition {
	t.Helper()

	f, err := os.Open(providerConfigCRD)
	if err != nil {
		t.Fatalf("os.Open(%q) error = %v", providerConfigCRD, err)
	}
	defer f.Close()

	crd := &apiextensionsv1.CustomResourceDefinition{}
	if err := yaml.NewYAMLOrJSONDecoder(f, 4096).Decode(crd); err != nil {
		t.Fatalf("Decode(%q) error = %v", providerConfigCRD, err)
	}

	return crd
}

func ptrBool(b *bool) bool { return b != nil && *b }
