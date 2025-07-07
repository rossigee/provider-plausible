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

package v1alpha1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// GoalParameters are the configurable fields of a Goal.
type GoalParameters struct {
	// SiteDomain is the domain of the site this goal belongs to.
	// This can be specified directly or via a reference/selector.
	// +optional
	SiteDomain *string `json:"siteDomain,omitempty"`

	// SiteDomainRef references a Site resource to retrieve its domain.
	// +optional
	SiteDomainRef *xpv1.Reference `json:"siteDomainRef,omitempty"`

	// SiteDomainSelector selects a Site resource to retrieve its domain.
	// +optional
	SiteDomainSelector *xpv1.Selector `json:"siteDomainSelector,omitempty"`

	// GoalType is the type of goal (e.g., "event", "page").
	// +kubebuilder:validation:Enum=event;page
	// +kubebuilder:validation:Required
	GoalType string `json:"goalType"`

	// EventName is required when GoalType is "event".
	// +optional
	EventName *string `json:"eventName,omitempty"`

	// PagePath is required when GoalType is "page".
	// +optional
	PagePath *string `json:"pagePath,omitempty"`
}

// GoalObservation are the observable fields of a Goal.
type GoalObservation struct {
	// ID is the unique identifier of the goal in Plausible.
	ID string `json:"id,omitempty"`

	// GoalType is the type of the goal.
	GoalType string `json:"goalType,omitempty"`

	// EventName if the goal is an event type.
	EventName string `json:"eventName,omitempty"`

	// PagePath if the goal is a page type.
	PagePath string `json:"pagePath,omitempty"`

	// CreatedAt is the timestamp when the goal was created.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`
}

// A GoalSpec defines the desired state of a Goal.
type GoalSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       GoalParameters `json:"forProvider"`
}

// A GoalStatus represents the observed state of a Goal.
type GoalStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          GoalObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Goal is a managed resource that represents a Plausible goal.
// +kubebuilder:printcolumn:name="TYPE",type="string",JSONPath=".spec.forProvider.goalType"
// +kubebuilder:printcolumn:name="GOAL-ID",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,plausible}
type Goal struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GoalSpec   `json:"spec"`
	Status GoalStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GoalList contains a list of Goal
type GoalList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Goal `json:"items"`
}

// Goal type metadata.
var (
	GoalKind             = reflect.TypeOf(Goal{}).Name()
	GoalGroupKind        = schema.GroupKind{Group: Group, Kind: GoalKind}.String()
	GoalKindAPIVersion   = GoalKind + "." + SchemeGroupVersion.String()
	GoalGroupVersionKind = SchemeGroupVersion.WithKind(GoalKind)

	GoalListKind             = reflect.TypeOf(GoalList{}).Name()
	GoalListGroupKind        = schema.GroupKind{Group: Group, Kind: GoalListKind}.String()
	GoalListKindAPIVersion   = GoalListKind + "." + SchemeGroupVersion.String()
	GoalListGroupVersionKind = SchemeGroupVersion.WithKind(GoalListKind)
)