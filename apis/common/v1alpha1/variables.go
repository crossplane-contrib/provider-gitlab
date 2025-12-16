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

package v1alpha1

// VariableType indicates the type of the GitLab CI variable.
type VariableType string

// List of variable type values.
const (
	VariableTypeEnvVar VariableType = "env_var"
	VariableTypeFile   VariableType = "file"
)

// CommonVariableParameters represents the desired state of a GitLab Variable.
type CommonVariableParameters struct {
	// Key of a variable.
	// If the key already exists, the existing variable will be updated instead of creating a new one.
	// +kubebuilder:validation:Pattern:=^[a-zA-Z0-9\_]+$
	// +kubebuilder:validation:MaxLength:=255
	// +immutable
	Key string `json:"key"`

	// Value of a variable. Mutually exclusive with ValueSecretRef.
	// +kubebuilder:validation:MaxLength:=10000
	// +optional
	Value *string `json:"value,omitempty"`

	// The description of the variable. Maximum of 255 characters.
	// +kubebuilder:validation:MaxLength:=255
	// +optional
	Description *string `json:"description,omitempty"`

	// Masked enables or disables variable masking.
	// +optional
	Masked *bool `json:"masked,omitempty"`

	// Protected enables or disables variable protection.
	// +optional
	Protected *bool `json:"protected,omitempty"`

	// Raw disables variable expansion of the variable.
	// +optional
	Raw *bool `json:"raw,omitempty"`

	// VariableType is the type of a variable.
	// +kubebuilder:validation:Enum:=env_var;file
	// +optional
	VariableType *VariableType `json:"variableType,omitempty"`
}

// CommonVariableObservation represents the observed state of a GitLab Variable.
type CommonVariableObservation struct {
	// Key of a variable.
	Key string `json:"key,omitempty"`
	// The description of the variable. Maximum of 255 characters.
	Description string `json:"description,omitempty"`
	// Masked enables or disables variable masking.
	Masked bool `json:"masked,omitempty"`
	// Protected enables or disables variable protection.
	Protected bool `json:"protected,omitempty"`
	// Raw disables variable expansion of the variable.
	Raw bool `json:"raw,omitempty"`
	// VariableType is the type of a variable.
	VariableType VariableType `json:"variableType,omitempty"`
}
