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

// SiteParameters are the configurable fields of a Site.
type SiteParameters struct {
	// Domain is the domain name of the site in Plausible.
	// This is the primary identifier for the site.
	// +kubebuilder:validation:Required
	Domain string `json:"domain"`

	// NewDomain is used when updating the domain of an existing site.
	// This field is only used during updates and should be left empty during creation.
	// +optional
	NewDomain *string `json:"newDomain,omitempty"`

	// TeamID associates the site with a specific team.
	// If not provided, the site will be associated with the default team.
	// +optional
	TeamID *string `json:"teamID,omitempty"`

	// Timezone for the site. Must be a valid IANA timezone string.
	// If not provided, defaults to UTC.
	// +optional
	Timezone *string `json:"timezone,omitempty"`
}

// SiteObservation are the observable fields of a Site.
type SiteObservation struct {
	// ID is the unique identifier of the site in Plausible.
	ID string `json:"id,omitempty"`

	// Domain is the current domain of the site.
	Domain string `json:"domain,omitempty"`

	// TeamID is the ID of the team the site belongs to.
	TeamID string `json:"teamID,omitempty"`

	// CreatedAt is the timestamp when the site was created.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// UpdatedAt is the timestamp when the site was last updated.
	UpdatedAt *metav1.Time `json:"updatedAt,omitempty"`
}

// A SiteSpec defines the desired state of a Site.
type SiteSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       SiteParameters `json:"forProvider"`
}

// A SiteStatus represents the observed state of a Site.
type SiteStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          SiteObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Site is a managed resource that represents a Plausible Analytics site.
// +kubebuilder:printcolumn:name="DOMAIN",type="string",JSONPath=".spec.forProvider.domain"
// +kubebuilder:printcolumn:name="SITE-ID",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,plausible}
type Site struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SiteSpec   `json:"spec"`
	Status SiteStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SiteList contains a list of Site
type SiteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Site `json:"items"`
}

// Site type metadata.
var (
	SiteKind             = reflect.TypeOf(Site{}).Name()
	SiteGroupKind        = schema.GroupKind{Group: Group, Kind: SiteKind}.String()
	SiteKindAPIVersion   = SiteKind + "." + SchemeGroupVersion.String()
	SiteGroupVersionKind = SchemeGroupVersion.WithKind(SiteKind)

	SiteListKind             = reflect.TypeOf(SiteList{}).Name()
	SiteListGroupKind        = schema.GroupKind{Group: Group, Kind: SiteListKind}.String()
	SiteListKindAPIVersion   = SiteListKind + "." + SchemeGroupVersion.String()
	SiteListGroupVersionKind = SchemeGroupVersion.WithKind(SiteListKind)
)