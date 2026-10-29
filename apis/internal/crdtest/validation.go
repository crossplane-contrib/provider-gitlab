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

package crdtest

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema/cel"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	celconfig "k8s.io/apiserver/pkg/apis/cel"
	"sigs.k8s.io/yaml"
)

// Load applies the defaults and conversion used when the API server creates a CRD.
func Load(t testing.TB, filename string) *apiextensions.CustomResourceDefinition {
	t.Helper()
	data, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		t.Fatal(err)
	}
	var versioned apiextensionsv1.CustomResourceDefinition
	if err := yaml.UnmarshalStrict(data, &versioned); err != nil {
		t.Fatal(err)
	}
	apiextensionsv1.SetObjectDefaults_CustomResourceDefinition(&versioned)
	var crd apiextensions.CustomResourceDefinition
	if err := apiextensionsv1.Convert_v1_CustomResourceDefinition_To_apiextensions_CustomResourceDefinition(&versioned, &crd, nil); err != nil {
		t.Fatal(err)
	}
	for _, version := range crd.Spec.Versions {
		if version.Storage {
			crd.Status.StoredVersions = append(crd.Status.StoredVersions, version.Name)
		}
	}
	return &crd
}

// Validator evaluates transition rules from a generated CRD against typed resources.
type Validator struct {
	schema *schema.Structural
	cel    *cel.Validator
}

// NewValidator loads the v1alpha1 schema, including shared schemas after conversion.
func NewValidator(t testing.TB, filename string) Validator {
	t.Helper()
	crd := Load(t, filename)
	validation := crd.Spec.Validation
	for _, version := range crd.Spec.Versions {
		if version.Name == "v1alpha1" && version.Schema != nil {
			validation = version.Schema
		}
	}
	if validation == nil {
		t.Fatal("CRD has no validation schema")
	}
	structural, err := schema.NewStructural(validation.OpenAPIV3Schema)
	if err != nil {
		t.Fatal(err)
	}
	validator := cel.NewValidator(structural, true, celconfig.PerCallLimit)
	if validator == nil {
		t.Fatal("CRD has no CEL validation rules")
	}
	return Validator{schema: structural, cel: validator}
}

// CheckUpdate requires either acceptance or a rejection containing wantError.
func (v Validator) CheckUpdate(t testing.TB, oldResource, newResource runtime.Object, wantError string) {
	t.Helper()
	oldObject, err := runtime.DefaultUnstructuredConverter.ToUnstructured(oldResource)
	if err != nil {
		t.Fatal(err)
	}
	newObject, err := runtime.DefaultUnstructuredConverter.ToUnstructured(newResource)
	if err != nil {
		t.Fatal(err)
	}
	errs, _ := v.cel.Validate(context.Background(), field.NewPath("resource"), v.schema, newObject, oldObject, celconfig.RuntimeCELCostBudget)
	if wantError == "" {
		if len(errs) != 0 {
			t.Fatalf("update rejected: %v", errs)
		}
		return
	}
	if len(errs) == 0 || !strings.Contains(errs.ToAggregate().Error(), wantError) {
		t.Fatalf("want rejection containing %q, got %v", wantError, errs)
	}
}
