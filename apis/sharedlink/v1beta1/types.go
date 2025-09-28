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

// SharedLinkParameters are the configurable fields of a SharedLink.
type SharedLinkParameters struct {
	// SiteDomain is the domain of the site this shared link belongs to.
	// This can be specified directly or via a reference/selector.
	// +optional
	SiteDomain *string `json:"siteDomain,omitempty"`

	// SiteDomainRef references a Site resource to retrieve its domain.
	// +optional
	SiteDomainRef *xpv1.Reference `json:"siteDomainRef,omitempty"`

	// SiteDomainSelector selects a Site resource to retrieve its domain.
	// +optional
	SiteDomainSelector *xpv1.Selector `json:"siteDomainSelector,omitempty"`

	// Name is the name of the shared link.
	// This is used to identify and retrieve the shared link.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Password provides optional password protection for the shared link.
	// If set, viewers must enter this password to access the dashboard.
	// +optional
	Password *string `json:"password,omitempty"`
}

// SharedLinkObservation are the observable fields of a SharedLink.
type SharedLinkObservation struct {
	// Name is the name of the shared link.
	Name string `json:"name,omitempty"`

	// URL is the shareable URL for accessing the dashboard.
	URL string `json:"url,omitempty"`

	// HasPassword indicates whether the shared link is password protected.
	HasPassword bool `json:"hasPassword,omitempty"`

	// CreatedAt is the timestamp when the shared link was created.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`
}

// A SharedLinkSpec defines the desired state of a SharedLink.
type SharedLinkSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       SharedLinkParameters `json:"forProvider"`
}

// A SharedLinkStatus represents the observed state of a SharedLink.
type SharedLinkStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          SharedLinkObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A SharedLink is a managed resource that represents a Plausible shared dashboard link.
// +kubebuilder:printcolumn:name="NAME",type="string",JSONPath=".spec.forProvider.name"
// +kubebuilder:printcolumn:name="URL",type="string",JSONPath=".status.atProvider.url"
// +kubebuilder:printcolumn:name="PROTECTED",type="boolean",JSONPath=".status.atProvider.hasPassword"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,plausible}
type SharedLink struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SharedLinkSpec   `json:"spec"`
	Status SharedLinkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SharedLinkList contains a list of SharedLink
type SharedLinkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SharedLink `json:"items"`
}

// SharedLink type metadata.
var (
	SharedLinkKind             = reflect.TypeOf(SharedLink{}).Name()
	SharedLinkGroupKind        = schema.GroupKind{Group: Group, Kind: SharedLinkKind}.String()
	SharedLinkKindAPIVersion   = SharedLinkKind + "." + SchemeGroupVersion.String()
	SharedLinkGroupVersionKind = SchemeGroupVersion.WithKind(SharedLinkKind)

	SharedLinkListKind             = reflect.TypeOf(SharedLinkList{}).Name()
	SharedLinkListGroupKind        = schema.GroupKind{Group: Group, Kind: SharedLinkListKind}.String()
	SharedLinkListKindAPIVersion   = SharedLinkListKind + "." + SchemeGroupVersion.String()
	SharedLinkListGroupVersionKind = SchemeGroupVersion.WithKind(SharedLinkListKind)
)