/*
Copyright 2023 The Crossplane Authors.

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

package v1beta1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
)

// CustomPropertyParameters are the configurable fields of a CustomProperty.
type CustomPropertyParameters struct {
	// SiteDomain is the domain of the site this custom property belongs to.
	// This can be specified directly or via a reference/selector.
	// +optional
	SiteDomain *string `json:"siteDomain,omitempty"`

	// SiteDomainRef references a Site resource to retrieve its domain.
	// +optional
	SiteDomainRef *xpv1.Reference `json:"siteDomainRef,omitempty"`

	// SiteDomainSelector selects a Site resource to retrieve its domain.
	// +optional
	SiteDomainSelector *xpv1.Selector `json:"siteDomainSelector,omitempty"`

	// Key is the name/key of the custom property.
	// This is used to identify the custom property in analytics.
	// +kubebuilder:validation:Required
	Key string `json:"key"`

	// Description provides a human-readable description of the custom property.
	// +optional
	Description *string `json:"description,omitempty"`
}

// CustomPropertyObservation are the observable fields of a CustomProperty.
type CustomPropertyObservation struct {
	// Key is the key of the custom property.
	Key string `json:"key,omitempty"`

	// Description is the description of the custom property.
	Description string `json:"description,omitempty"`

	// IsEnabled indicates whether the custom property is enabled.
	IsEnabled bool `json:"isEnabled,omitempty"`

	// CreatedAt is the timestamp when the custom property was created.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`
}

// A CustomPropertySpec defines the desired state of a CustomProperty.
type CustomPropertySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       CustomPropertyParameters `json:"forProvider"`
}

// A CustomPropertyStatus represents the observed state of a CustomProperty.
type CustomPropertyStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          CustomPropertyObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A CustomProperty is a managed resource that represents a Plausible custom property.
// +kubebuilder:printcolumn:name="KEY",type="string",JSONPath=".spec.forProvider.key"
// +kubebuilder:printcolumn:name="DESCRIPTION",type="string",JSONPath=".status.atProvider.description"
// +kubebuilder:printcolumn:name="ENABLED",type="boolean",JSONPath=".status.atProvider.isEnabled"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,plausible}
type CustomProperty struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomPropertySpec   `json:"spec"`
	Status CustomPropertyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CustomPropertyList contains a list of CustomProperty
type CustomPropertyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomProperty `json:"items"`
}

// CustomProperty type metadata.
var (
	CustomPropertyKind             = reflect.TypeOf(CustomProperty{}).Name()
	CustomPropertyGroupKind        = schema.GroupKind{Group: Group, Kind: CustomPropertyKind}.String()
	CustomPropertyKindAPIVersion   = CustomPropertyKind + "." + SchemeGroupVersion.String()
	CustomPropertyGroupVersionKind = SchemeGroupVersion.WithKind(CustomPropertyKind)

	CustomPropertyListKind             = reflect.TypeOf(CustomPropertyList{}).Name()
	CustomPropertyListGroupKind        = schema.GroupKind{Group: Group, Kind: CustomPropertyListKind}.String()
	CustomPropertyListKindAPIVersion   = CustomPropertyListKind + "." + SchemeGroupVersion.String()
	CustomPropertyListGroupVersionKind = SchemeGroupVersion.WithKind(CustomPropertyListKind)
)