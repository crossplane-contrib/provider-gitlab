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

// Package v1beta2 contains the core resources of the gitlab provider. It is the
// storage version of the namespaced ProviderConfig and exposes a cleaned-up
// credentials surface: the credentials secret is referenced with a
// LocalSecretKeySelector (always resolved in the ProviderConfig's own
// namespace), and the filesystem/environment sources that cannot be scoped to a
// namespace are not offered.
// +kubebuilder:object:generate=true
// +groupName=gitlab.m.crossplane.io
// +versionName=v1beta2
package v1beta2
