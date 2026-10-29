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
	"path/filepath"
	"testing"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/validation"
)

func TestGeneratedCRDs(t *testing.T) {
	files, err := filepath.Glob("../../../package/crds/*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no generated CRDs found")
	}
	for _, filename := range files {
		t.Run(filepath.Base(filename), func(t *testing.T) {
			if errs := validation.ValidateCustomResourceDefinition(context.Background(), Load(t, filename)); len(errs) != 0 {
				t.Fatalf("invalid generated CRD: %v", errs)
			}
		})
	}
}
