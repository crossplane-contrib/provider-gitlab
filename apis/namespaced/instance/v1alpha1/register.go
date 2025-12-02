package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "instance.gitlab.m.crossplane.io"
	Version = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// ApplicationSettings type metadata
var (
	ApplicationSettingsKind             = reflect.TypeOf(ApplicationSettings{}).Name()
	ApplicationSettingsGroupKind        = schema.GroupKind{Group: Group, Kind: ApplicationSettingsKind}.String()
	ApplicationSettingsKindAPIVersion   = ApplicationSettingsKind + "." + SchemeGroupVersion.String()
	ApplicationSettingsGroupVersionKind = SchemeGroupVersion.WithKind(ApplicationSettingsKind)

	ApplicationSettingsListKind             = reflect.TypeOf(ApplicationSettingsList{}).Name()
	ApplicationSettingsListGroupKind        = schema.GroupKind{Group: Group, Kind: ApplicationSettingsListKind}.String()
	ApplicationSettingsListKindAPIVersion   = ApplicationSettingsListKind + "." + SchemeGroupVersion.String()
	ApplicationSettingsListGroupVersionKind = SchemeGroupVersion.WithKind(ApplicationSettingsListKind)
)

func init() {
	SchemeBuilder.Register(&ApplicationSettings{}, &ApplicationSettingsList{})
}
