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
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
)

// CreateSecretKeySelector creates a SecretKeySelector for testing
func TestCreateSecretKeySelector(name, key string) *v2.SecretKeySelector {
	return &v2.SecretKeySelector{
		Key: key,
		SecretReference: v2.SecretReference{
			Name:      name,
			Namespace: "default",
		},
	}
}

// CreateLocalSecretKeySelector creates a LocalSecretKeySelector for testing
func TestCreateLocalSecretKeySelector(name, key string) *v2.LocalSecretKeySelector {
	return &v2.LocalSecretKeySelector{
		Key: key,
		LocalSecretReference: v2.LocalSecretReference{
			Name: name,
		},
	}
}

// CreateSecretKeySelector creates a SecretKeySelector for testing
func TestCreateSecretReference(name string) *v2.SecretReference {
	return &v2.SecretReference{
		Name:      name,
		Namespace: "default",
	}
}

// CreateLocalSecretKeySelector creates a LocalSecretKeySelector for testing
func TestCreateLocalSecretReference(name string) *v2.LocalSecretReference {
	return &v2.LocalSecretReference{
		Name: name,
	}
}
