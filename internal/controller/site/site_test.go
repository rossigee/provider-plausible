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

package site

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"

	v1beta1 "github.com/rossigee/provider-plausible/apis/site/v1beta1"
	"github.com/rossigee/provider-plausible/internal/clients"
)

// PlausibleService defines the interface for Plausible operations
type PlausibleService interface {
	GetSite(siteID string) (*clients.Site, error)
	GetSiteByDomain(domain string) (*clients.Site, error)
	CreateSite(req clients.CreateSiteRequest) (*clients.Site, error)
	UpdateSite(siteID string, newDomain string) (*clients.Site, error)
	DeleteSite(siteID string) error
}

// testExternal is a test version of external that takes an interface
type testExternal struct {
	service PlausibleService
}

func (c *testExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.Site)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotSite)
	}

	// If we have an external name (site ID), try to get by ID
	if meta.GetExternalName(cr) != "" {
		site, err := c.service.GetSite(meta.GetExternalName(cr))
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, "failed to get site by ID")
		}

		if site == nil {
			return managed.ExternalObservation{
				ResourceExists: false,
			}, nil
		}

		cr.Status.AtProvider = v1beta1.SiteObservation{
			ID:     site.ID,
			Domain: site.Domain,
			TeamID: site.TeamID,
		}

		cr.SetConditions(xpv1.Available())

		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: c.isUpToDate(cr, site),
		}, nil
	}

	// If no external name, try to find by domain
	site, err := c.service.GetSiteByDomain(cr.Spec.ForProvider.Domain)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to get site by domain")
	}

	if site == nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	// Set the external name to the site ID
	meta.SetExternalName(cr, site.ID)

	cr.Status.AtProvider = v1beta1.SiteObservation{
		ID:     site.ID,
		Domain: site.Domain,
		TeamID: site.TeamID,
	}

	cr.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: c.isUpToDate(cr, site),
	}, nil
}

func (c *testExternal) isUpToDate(cr *v1beta1.Site, site *clients.Site) bool {
	// Check if domain needs to be updated
	if cr.Spec.ForProvider.NewDomain != nil && *cr.Spec.ForProvider.NewDomain != site.Domain {
		return false
	}

	// Note: Team ID and timezone cannot be updated after creation via API
	return true
}

func (c *testExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.Site)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotSite)
	}

	cr.SetConditions(xpv1.Creating())

	req := clients.CreateSiteRequest{
		Domain: cr.Spec.ForProvider.Domain,
	}

	if cr.Spec.ForProvider.TeamID != nil {
		req.TeamID = *cr.Spec.ForProvider.TeamID
	}

	if cr.Spec.ForProvider.Timezone != nil {
		req.Timezone = *cr.Spec.ForProvider.Timezone
	}

	site, err := c.service.CreateSite(req)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "failed to create site")
	}

	meta.SetExternalName(cr, site.ID)

	return managed.ExternalCreation{}, nil
}

func (c *testExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.Site)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotSite)
	}

	// Only domain can be updated
	if cr.Spec.ForProvider.NewDomain != nil && *cr.Spec.ForProvider.NewDomain != cr.Status.AtProvider.Domain {
		_, err := c.service.UpdateSite(meta.GetExternalName(cr), *cr.Spec.ForProvider.NewDomain)
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "failed to update site domain")
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (c *testExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1beta1.Site)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotSite)
	}

	cr.SetConditions(xpv1.Deleting())

	err := c.service.DeleteSite(meta.GetExternalName(cr))
	if err != nil && !clients.IsNotFound(err) {
		return managed.ExternalDelete{}, errors.Wrap(err, "failed to delete site")
	}

	return managed.ExternalDelete{}, nil
}

func (c *testExternal) Disconnect(ctx context.Context) error {
	return nil
}

// MockPlausibleClient is a mock implementation of the Plausible client
type MockPlausibleClient struct {
	MockGetSite         func(siteID string) (*clients.Site, error)
	MockGetSiteByDomain func(domain string) (*clients.Site, error)
	MockCreateSite      func(req clients.CreateSiteRequest) (*clients.Site, error)
	MockUpdateSite      func(siteID string, newDomain string) (*clients.Site, error)
	MockDeleteSite      func(siteID string) error
}

func (m *MockPlausibleClient) GetSite(siteID string) (*clients.Site, error) {
	return m.MockGetSite(siteID)
}

func (m *MockPlausibleClient) GetSiteByDomain(domain string) (*clients.Site, error) {
	return m.MockGetSiteByDomain(domain)
}

func (m *MockPlausibleClient) CreateSite(req clients.CreateSiteRequest) (*clients.Site, error) {
	return m.MockCreateSite(req)
}

func (m *MockPlausibleClient) UpdateSite(siteID string, newDomain string) (*clients.Site, error) {
	return m.MockUpdateSite(siteID, newDomain)
}

func (m *MockPlausibleClient) DeleteSite(siteID string) error {
	return m.MockDeleteSite(siteID)
}

func (m *MockPlausibleClient) ListSites() ([]clients.Site, error) {
	return nil, nil
}

func TestObserve(t *testing.T) {
	type args struct {
		service PlausibleService
		cr      *v1beta1.Site
	}
	type want struct {
		cr          *v1beta1.Site
		observation managed.ExternalObservation
		err         error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"SiteExists": {
			args: args{
				service: &MockPlausibleClient{
					MockGetSiteByDomain: func(domain string) (*clients.Site, error) {
						if domain != "example.com" {
							return nil, fmt.Errorf("unexpected domain: %s", domain)
						}
						return &clients.Site{
							ID:       "example.com",
							Domain:   "example.com",
							Timezone: "UTC",
						}, nil
					},
				},
				cr: &v1beta1.Site{
					Spec: v1beta1.SiteSpec{
						ForProvider: v1beta1.SiteParameters{
							Domain:   "example.com",
							Timezone: ptr("UTC"),
						},
					},
				},
			},
			want: want{
				cr: &v1beta1.Site{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"crossplane.io/external-name": "example.com",
						},
					},
					Spec: v1beta1.SiteSpec{
						ForProvider: v1beta1.SiteParameters{
							Domain:   "example.com",
							Timezone: ptr("UTC"),
						},
					},
					Status: v1beta1.SiteStatus{
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Available()},
							},
						},
						AtProvider: v1beta1.SiteObservation{
							ID:     "example.com",
							Domain: "example.com",
						},
					},
				},
				observation: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"SiteDoesNotExist": {
			args: args{
				service: &MockPlausibleClient{
					MockGetSiteByDomain: func(domain string) (*clients.Site, error) {
						return nil, nil
					},
				},
				cr: &v1beta1.Site{
					Spec: v1beta1.SiteSpec{
						ForProvider: v1beta1.SiteParameters{
							Domain:   "nonexistent.com",
							Timezone: ptr("UTC"),
						},
					},
				},
			},
			want: want{
				cr: &v1beta1.Site{
					Spec: v1beta1.SiteSpec{
						ForProvider: v1beta1.SiteParameters{
							Domain:   "nonexistent.com",
							Timezone: ptr("UTC"),
						},
					},
				},
				observation: managed.ExternalObservation{
					ResourceExists:   false,
					ResourceUpToDate: false,
				},
			},
		},
		"SiteNeedsUpdate": {
			args: args{
				service: &MockPlausibleClient{
					MockGetSiteByDomain: func(domain string) (*clients.Site, error) {
						return &clients.Site{
							ID:       "example.com",
							Domain:   "example.com",
							Timezone: "UTC",
						}, nil
					},
				},
				cr: &v1beta1.Site{
					Spec: v1beta1.SiteSpec{
						ForProvider: v1beta1.SiteParameters{
							Domain:   "example.com",
							Timezone: ptr("Asia/Bangkok"), // Different timezone
						},
					},
				},
			},
			want: want{
				cr: &v1beta1.Site{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"crossplane.io/external-name": "example.com",
						},
					},
					Spec: v1beta1.SiteSpec{
						ForProvider: v1beta1.SiteParameters{
							Domain:   "example.com",
							Timezone: ptr("Asia/Bangkok"),
						},
					},
					Status: v1beta1.SiteStatus{
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Available()},
							},
						},
						AtProvider: v1beta1.SiteObservation{
							ID:     "example.com",
							Domain: "example.com",
						},
					},
				},
				observation: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &testExternal{service: tc.args.service}
			observation, err := e.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Observe(): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("Observe(): -want cr, +got cr:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.observation, observation); diff != "" {
				t.Errorf("Observe(): -want observation, +got observation:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type args struct {
		service PlausibleService
		cr      *v1beta1.Site
	}
	type want struct {
		cr      *v1beta1.Site
		created managed.ExternalCreation
		err     error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"Successful": {
			args: args{
				service: &MockPlausibleClient{
					MockCreateSite: func(req clients.CreateSiteRequest) (*clients.Site, error) {
						if req.Domain != "new.example.com" {
							return nil, fmt.Errorf("unexpected domain: %s", req.Domain)
						}
						return &clients.Site{
							Domain:   "new.example.com",
							Timezone: "UTC",
						}, nil
					},
				},
				cr: &v1beta1.Site{
					Spec: v1beta1.SiteSpec{
						ForProvider: v1beta1.SiteParameters{
							Domain:   "new.example.com",
							Timezone: ptr("UTC"),
						},
					},
				},
			},
			want: want{
				cr: &v1beta1.Site{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"crossplane.io/external-name": "",
						},
					},
					Spec: v1beta1.SiteSpec{
						ForProvider: v1beta1.SiteParameters{
							Domain:   "new.example.com",
							Timezone: ptr("UTC"),
						},
					},
					Status: v1beta1.SiteStatus{
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Creating()},
							},
						},
					},
				},
				created: managed.ExternalCreation{},
			},
		},
		"CreateFailed": {
			args: args{
				service: &MockPlausibleClient{
					MockCreateSite: func(req clients.CreateSiteRequest) (*clients.Site, error) {
						return nil, errors.New("API error")
					},
				},
				cr: &v1beta1.Site{
					Spec: v1beta1.SiteSpec{
						ForProvider: v1beta1.SiteParameters{
							Domain: "fail.example.com",
						},
					},
				},
			},
			want: want{
				cr: &v1beta1.Site{
					Spec: v1beta1.SiteSpec{
						ForProvider: v1beta1.SiteParameters{
							Domain: "fail.example.com",
						},
					},
					Status: v1beta1.SiteStatus{
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Creating()},
							},
						},
					},
				},
				err: errors.Wrap(errors.New("API error"), "failed to create site"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &testExternal{service: tc.args.service}
			created, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Create(): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("Create(): -want cr, +got cr:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.created, created); diff != "" {
				t.Errorf("Create(): -want created, +got created:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type args struct {
		service PlausibleService
		cr      *v1beta1.Site
	}
	type want struct {
		cr      *v1beta1.Site
		updated managed.ExternalUpdate
		err     error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"UpdateDomain": {
			args: args{
				service: &MockPlausibleClient{
					MockUpdateSite: func(siteID string, newDomain string) (*clients.Site, error) {
						if siteID != "example.com" {
							return nil, fmt.Errorf("unexpected site ID: %s", siteID)
						}
						if newDomain != "new.example.com" {
							return nil, fmt.Errorf("unexpected new domain: %s", newDomain)
						}
						return &clients.Site{
							ID:       "example.com",
							Domain:   "new.example.com",
							Timezone: "UTC",
						}, nil
					},
				},
				cr: &v1beta1.Site{
					Spec: v1beta1.SiteSpec{
						ForProvider: v1beta1.SiteParameters{
							Domain:    "example.com",
							NewDomain: ptr("new.example.com"),
						},
					},
					Status: v1beta1.SiteStatus{
						AtProvider: v1beta1.SiteObservation{
							Domain: "example.com",
						},
					},
				},
			},
			want: want{
				cr: &v1beta1.Site{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"crossplane.io/external-name": "example.com",
						},
					},
					Spec: v1beta1.SiteSpec{
						ForProvider: v1beta1.SiteParameters{
							Domain:    "example.com",
							NewDomain: ptr("new.example.com"),
						},
					},
					Status: v1beta1.SiteStatus{
						AtProvider: v1beta1.SiteObservation{
							Domain: "example.com",
						},
					},
				},
				updated: managed.ExternalUpdate{},
			},
		},
		"UpdateDomainSecond": {
			args: args{
				service: &MockPlausibleClient{
					MockUpdateSite: func(siteID string, newDomain string) (*clients.Site, error) {
						if newDomain != "new.example.com" {
							return nil, fmt.Errorf("unexpected new domain: %s", newDomain)
						}
						return &clients.Site{
							ID:       "old.example.com",
							Domain:   "new.example.com",
							Timezone: "UTC",
						}, nil
					},
				},
				cr: &v1beta1.Site{
					Spec: v1beta1.SiteSpec{
						ForProvider: v1beta1.SiteParameters{
							Domain:    "old.example.com",
							NewDomain: ptr("new.example.com"),
						},
					},
				},
			},
			want: want{
				cr: &v1beta1.Site{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"crossplane.io/external-name": "old.example.com",
						},
					},
					Spec: v1beta1.SiteSpec{
						ForProvider: v1beta1.SiteParameters{
							Domain:    "old.example.com", // Should remain unchanged
							NewDomain: ptr("new.example.com"), // Should remain unchanged
						},
					},
				},
				updated: managed.ExternalUpdate{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Set external name for update
			meta.SetExternalName(tc.args.cr, tc.args.cr.Spec.ForProvider.Domain)
			
			e := &testExternal{service: tc.args.service}
			updated, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Update(): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("Update(): -want cr, +got cr:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.updated, updated); diff != "" {
				t.Errorf("Update(): -want updated, +got updated:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		service PlausibleService
		cr      *v1beta1.Site
	}
	type want struct {
		err error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"Successful": {
			args: args{
				service: &MockPlausibleClient{
					MockDeleteSite: func(siteID string) error {
						if siteID != "example.com" {
							return fmt.Errorf("unexpected site ID: %s", siteID)
						}
						return nil
					},
				},
				cr: &v1beta1.Site{
					Spec: v1beta1.SiteSpec{
						ForProvider: v1beta1.SiteParameters{
							Domain: "example.com",
						},
					},
				},
			},
			want: want{
				err: nil,
			},
		},
		"DeleteFailed": {
			args: args{
				service: &MockPlausibleClient{
					MockDeleteSite: func(siteID string) error {
						return errors.New("API error")
					},
				},
				cr: &v1beta1.Site{
					Spec: v1beta1.SiteSpec{
						ForProvider: v1beta1.SiteParameters{
							Domain: "example.com",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API error"), "failed to delete site"),
			},
		},
		"AlreadyDeleted": {
			args: args{
				service: &MockPlausibleClient{
					MockDeleteSite: func(siteID string) error {
						return fmt.Errorf("API request failed with status 404: Not Found")
					},
				},
				cr: &v1beta1.Site{
					Spec: v1beta1.SiteSpec{
						ForProvider: v1beta1.SiteParameters{
							Domain: "example.com",
						},
					},
				},
			},
			want: want{
				err: nil, // 404 errors are ignored
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Set external name for delete
			meta.SetExternalName(tc.args.cr, tc.args.cr.Spec.ForProvider.Domain)
			
			e := &testExternal{service: tc.args.service}
			_, err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Delete(): -want error, +got error:\n%s", diff)
			}
		})
	}
}

// Helper function
func ptr[T any](v T) *T {
	return &v
}