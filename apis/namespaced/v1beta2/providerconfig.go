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

import v2 "github.com/crossplane/crossplane/apis/v2/core/v2"

// The methods below implement resource.ProviderConfig so the usage-accounting
// reconciler can operate on the storage version directly. They are written by
// hand because angryjet only matches the crossplane-runtime
// ProviderConfigStatus, not the core/v2 one embedded here.

// GetCondition of this ProviderConfig.
func (p *ProviderConfig) GetCondition(ct v2.ConditionType) v2.Condition {
	return p.Status.GetCondition(ct)
}

// GetUsers of this ProviderConfig.
func (p *ProviderConfig) GetUsers() int64 {
	return p.Status.Users
}

// SetConditions of this ProviderConfig.
func (p *ProviderConfig) SetConditions(c ...v2.Condition) {
	p.Status.SetConditions(c...)
}

// SetUsers of this ProviderConfig.
func (p *ProviderConfig) SetUsers(i int64) {
	p.Status.Users = i
}
