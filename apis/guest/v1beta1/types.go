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

// GuestParameters are the configurable fields of a Guest.
type GuestParameters struct {
	// SiteDomain is the domain of the site this guest should have access to.
	// This can be specified directly or via a reference/selector.
	// +optional
	SiteDomain *string `json:"siteDomain,omitempty"`

	// SiteDomainRef references a Site resource to retrieve its domain.
	// +optional
	SiteDomainRef *xpv1.Reference `json:"siteDomainRef,omitempty"`

	// SiteDomainSelector selects a Site resource to retrieve its domain.
	// +optional
	SiteDomainSelector *xpv1.Selector `json:"siteDomainSelector,omitempty"`

	// Email is the email address of the guest to invite.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format=email
	Email string `json:"email"`

	// Role defines the access level for the guest.
	// Typically "viewer" for read-only access or "admin" for full access.
	// +kubebuilder:validation:Enum=viewer;admin
	// +kubebuilder:default="viewer"
	Role string `json:"role,omitempty"`
}

// GuestObservation are the observable fields of a Guest.
type GuestObservation struct {
	// Email is the email address of the guest.
	Email string `json:"email,omitempty"`

	// Role is the access level of the guest.
	Role string `json:"role,omitempty"`

	// Status indicates the current status of the guest invitation.
	// Can be "pending", "accepted", or "expired".
	Status string `json:"status,omitempty"`

	// InvitedAt is the timestamp when the guest was invited.
	InvitedAt *metav1.Time `json:"invitedAt,omitempty"`

	// AcceptedAt is the timestamp when the invitation was accepted.
	AcceptedAt *metav1.Time `json:"acceptedAt,omitempty"`
}

// A GuestSpec defines the desired state of a Guest.
type GuestSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       GuestParameters `json:"forProvider"`
}

// A GuestStatus represents the observed state of a Guest.
type GuestStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          GuestObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Guest is a managed resource that represents a Plausible site guest/collaborator.
// +kubebuilder:printcolumn:name="EMAIL",type="string",JSONPath=".spec.forProvider.email"
// +kubebuilder:printcolumn:name="ROLE",type="string",JSONPath=".spec.forProvider.role"
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.atProvider.status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,plausible}
type Guest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GuestSpec   `json:"spec"`
	Status GuestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GuestList contains a list of Guest
type GuestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Guest `json:"items"`
}

// Guest type metadata.
var (
	GuestKind             = reflect.TypeOf(Guest{}).Name()
	GuestGroupKind        = schema.GroupKind{Group: Group, Kind: GuestKind}.String()
	GuestKindAPIVersion   = GuestKind + "." + SchemeGroupVersion.String()
	GuestGroupVersionKind = SchemeGroupVersion.WithKind(GuestKind)

	GuestListKind             = reflect.TypeOf(GuestList{}).Name()
	GuestListGroupKind        = schema.GroupKind{Group: Group, Kind: GuestListKind}.String()
	GuestListKindAPIVersion   = GuestListKind + "." + SchemeGroupVersion.String()
	GuestListGroupVersionKind = SchemeGroupVersion.WithKind(GuestListKind)
)