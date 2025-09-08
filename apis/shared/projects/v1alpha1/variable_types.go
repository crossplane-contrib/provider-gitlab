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
//
// GitLab API docs: https://docs.gitlab.com/ee/api/project_level_variables.html
const (
	VariableTypeEnvVar VariableType = "env_var"
	VariableTypeFile   VariableType = "file"
)

// VariableParameters define the desired state of a Gitlab CI Variable
// https://docs.gitlab.com/ee/api/project_level_variables.html
type VariableParameters struct {
	// ProjectID is the ID of the project to create the variable on.
	// +optional
	// +immutable
	ProjectID *int `json:"projectId,omitempty"`

	// Key for the variable.
	// +kubebuilder:validation:Pattern:=^[a-zA-Z0-9\_]+$
	// +kubebuilder:validation:MaxLength:=255
	// +kubebuilder:example="AWS_ROLE_ARN"
	// +immutable
	Key string `json:"key"`

	// Value for the variable. Mutually exclusive with ValueSecretRef.
	// +kubebuilder:example="arn:aws:iam::999999999:role/my-deploy-role"
	// +optional
	Value *string `json:"value,omitempty"`

	// Masked enables or disables variable masking.
	// +optional
	Masked *bool `json:"masked,omitempty"`

	// Protected enables or disables variable protection.
	// +optional
	Protected *bool `json:"protected,omitempty"`

	// Raw disables variable expansion of the variable.
	// +optional
	Raw *bool `json:"raw,omitempty"`

	// VariableType is the type of the variable.
	// +kubebuilder:validation:Enum=env_var;file
	// +kubebuilder:example="file"
	// +optional
	VariableType *VariableType `json:"variableType,omitempty"`

	// EnvironmentScope indicates the environment scope
	// that this variable is applied to.
	// +optional
	EnvironmentScope *string `json:"environmentScope,omitempty"`
}
