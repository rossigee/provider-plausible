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

// TeamParameters are the configurable fields of a Team.
// Note: Teams are read-only resources that represent existing teams in Plausible.
// This resource is primarily for discovery and reference purposes.
type TeamParameters struct {
	// TeamID is the unique identifier of the team in Plausible.
	// This is used to filter and discover existing teams.
	// +optional
	TeamID *string `json:"teamID,omitempty"`
}

// TeamObservation are the observable fields of a Team.
type TeamObservation struct {
	// ID is the unique identifier of the team.
	ID string `json:"id,omitempty"`

	// Name is the display name of the team.
	Name string `json:"name,omitempty"`

	// APIEnabled indicates whether the Sites API is enabled for this team.
	APIEnabled bool `json:"apiEnabled,omitempty"`

	// CreatedAt is the timestamp when the team was created.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// UpdatedAt is the timestamp when the team was last updated.
	UpdatedAt *metav1.Time `json:"updatedAt,omitempty"`
}

// A TeamSpec defines the desired state of a Team.
type TeamSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       TeamParameters `json:"forProvider"`
}

// A TeamStatus represents the observed state of a Team.
type TeamStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          TeamObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Team is a managed resource that represents a Plausible team (read-only discovery).
// +kubebuilder:printcolumn:name="TEAM-ID",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:printcolumn:name="NAME",type="string",JSONPath=".status.atProvider.name"
// +kubebuilder:printcolumn:name="API-ENABLED",type="boolean",JSONPath=".status.atProvider.apiEnabled"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,plausible}
type Team struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TeamSpec   `json:"spec"`
	Status TeamStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TeamList contains a list of Team
type TeamList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Team `json:"items"`
}

// Team type metadata.
var (
	TeamKind             = reflect.TypeOf(Team{}).Name()
	TeamGroupKind        = schema.GroupKind{Group: Group, Kind: TeamKind}.String()
	TeamKindAPIVersion   = TeamKind + "." + SchemeGroupVersion.String()
	TeamGroupVersionKind = SchemeGroupVersion.WithKind(TeamKind)

	TeamListKind             = reflect.TypeOf(TeamList{}).Name()
	TeamListGroupKind        = schema.GroupKind{Group: Group, Kind: TeamListKind}.String()
	TeamListKindAPIVersion   = TeamListKind + "." + SchemeGroupVersion.String()
	TeamListGroupVersionKind = SchemeGroupVersion.WithKind(TeamListKind)
)