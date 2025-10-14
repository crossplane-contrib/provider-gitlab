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
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
)

// CreateSecretKeySelector creates a SecretKeySelector for testing
func TestCreateSecretKeySelector(name, key string) *xpv1.SecretKeySelector {
	return &xpv1.SecretKeySelector{
		Key: key,
		SecretReference: xpv1.SecretReference{
			Name:      name,
			Namespace: "default",
		},
	}
}

// CreateLocalSecretKeySelector creates a LocalSecretKeySelector for testing
func TestCreateLocalSecretKeySelector(name, key string) *xpv1.LocalSecretKeySelector {
	return &xpv1.LocalSecretKeySelector{
		Key: key,
		LocalSecretReference: xpv1.LocalSecretReference{
			Name: name,
		},
	}
}

// CreateSecretKeySelector creates a SecretKeySelector for testing
func TestCreateSecretReference(name string) *xpv1.SecretReference {
	return &xpv1.SecretReference{
		Name:      name,
		Namespace: "default",
	}
}

// CreateLocalSecretKeySelector creates a LocalSecretKeySelector for testing
func TestCreateLocalSecretReference(name string) *xpv1.LocalSecretReference {
	return &xpv1.LocalSecretReference{
		Name: name,
	}
}
