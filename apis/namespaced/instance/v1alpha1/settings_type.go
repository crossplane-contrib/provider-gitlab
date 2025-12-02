//go:generate go run generate.go

// +kubebuilder:object:generate=true
package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	// +cluster-scope:delete=1
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DefaultBranchProtectionDefaultsOptions defines the desired state of default branch protection levels
type DefaultBranchProtectionDefaultsOptions struct {
	AllowedToPush           *[]int `json:"allowedToPush,omitempty"`
	AllowForcePush          *bool  `json:"allowForcePush,omitempty"`
	AllowedToMerge          *[]int `json:"allowedToMerge,omitempty"`
	DeveloperCanInitialPush *bool  `json:"developerCanInitialPush,omitempty"`
}

// BranchProtectionDefaults defines the desired state of default branch protection levels
type BranchProtectionDefaults struct {
	AllowedToPush           []*int `json:"allowed_to_push,omitempty"`
	AllowForcePush          bool   `json:"allow_force_push,omitempty"`
	AllowedToMerge          []*int `json:"allowed_to_merge,omitempty"`
	DeveloperCanInitialPush bool   `json:"developer_can_initial_push,omitempty"`
}

// A ApplicationSettingSpec defines the desired state of a Gitlab Project CI
// ApplicationSettings.
type ApplicationSettingSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`
	ForProvider              ApplicationSettingsParameters `json:"forProvider"`
}

// A ApplicationSettingsStatus represents the observed state of a Gitlab Project CI
// ApplicationSettings.
type ApplicationSettingsStatus struct {
	xpv1.ResourceStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// A ApplicationSettings is a managed resource that represents a Gitlab instance Settings.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,gitlab}
type ApplicationSettings struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSettingSpec    `json:"spec"`
	Status ApplicationSettingsStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApplicationSettingsList contains a list of ApplicationSettings items.
type ApplicationSettingsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApplicationSettings `json:"items"`
}
