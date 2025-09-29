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

package goal

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"

	goalv1beta1 "github.com/rossigee/provider-plausible/apis/goal/v1beta1"
	sitev1beta1 "github.com/rossigee/provider-plausible/apis/site/v1beta1"
	"github.com/rossigee/provider-plausible/internal/clients"
)

// PlausibleService defines the interface for Plausible operations
type PlausibleService interface {
	GetGoal(siteDomain, goalID string) (*clients.Goal, error)
	ListGoals(siteDomain string) ([]clients.Goal, error)
	CreateGoal(siteDomain string, req clients.CreateGoalRequest) (*clients.Goal, error)
	DeleteGoal(goalID string) error
}

// testExternal is a test version of external that takes an interface
type testExternal struct {
	service PlausibleService
	kube    client.Client
}

func (c *testExternal) getSiteDomain(ctx context.Context, cr *goalv1beta1.Goal) (string, error) {
	// If direct domain is specified, use it
	if cr.Spec.ForProvider.SiteDomain != nil && *cr.Spec.ForProvider.SiteDomain != "" {
		return *cr.Spec.ForProvider.SiteDomain, nil
	}

	// If reference is specified, resolve it
	if cr.Spec.ForProvider.SiteDomainRef != nil {
		site := &sitev1beta1.Site{}
		nn := client.ObjectKey{
			Name: cr.Spec.ForProvider.SiteDomainRef.Name,
		}
		if err := c.kube.Get(ctx, nn, site); err != nil {
			return "", errors.Wrap(err, errGetSite)
		}
		return site.Spec.ForProvider.Domain, nil
	}

	// If selector is specified, we don't support it in this simple implementation
	if cr.Spec.ForProvider.SiteDomainSelector != nil {
		return "", errors.New("site domain selector is not yet implemented")
	}

	return "", errors.New("no site domain specified")
}

func (c *testExternal) goalMatches(cr *goalv1beta1.Goal, goal *clients.Goal) bool {
	if cr.Spec.ForProvider.GoalType != goal.GoalType {
		return false
	}

	switch cr.Spec.ForProvider.GoalType {
	case "event":
		return cr.Spec.ForProvider.EventName != nil && *cr.Spec.ForProvider.EventName == goal.EventName
	case "page":
		return cr.Spec.ForProvider.PagePath != nil && *cr.Spec.ForProvider.PagePath == goal.PagePath
	}

	return false
}

func (c *testExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*goalv1beta1.Goal)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotGoal)
	}

	siteDomain, err := c.getSiteDomain(ctx, cr)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	// If we have an external name (goal ID), try to get it
	if meta.GetExternalName(cr) != "" {
		goal, err := c.service.GetGoal(siteDomain, meta.GetExternalName(cr))
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, "failed to get goal")
		}

		if goal == nil {
			return managed.ExternalObservation{
				ResourceExists: false,
			}, nil
		}

		cr.Status.AtProvider = goalv1beta1.GoalObservation{
			ID:        goal.ID,
			GoalType:  goal.GoalType,
			EventName: goal.EventName,
			PagePath:  goal.PagePath,
		}

		cr.SetConditions(xpv1.Available())

		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: true, // Goals cannot be updated
		}, nil
	}

	// If no external name, try to find by matching goal properties
	goals, err := c.service.ListGoals(siteDomain)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to list goals")
	}

	for _, goal := range goals {
		if c.goalMatches(cr, &goal) {
			meta.SetExternalName(cr, goal.ID)

			cr.Status.AtProvider = goalv1beta1.GoalObservation{
				ID:        goal.ID,
				GoalType:  goal.GoalType,
				EventName: goal.EventName,
				PagePath:  goal.PagePath,
			}

			cr.SetConditions(xpv1.Available())

			return managed.ExternalObservation{
				ResourceExists:   true,
				ResourceUpToDate: true,
			}, nil
		}
	}

	return managed.ExternalObservation{
		ResourceExists: false,
	}, nil
}

func (c *testExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*goalv1beta1.Goal)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotGoal)
	}

	cr.SetConditions(xpv1.Creating())

	siteDomain, err := c.getSiteDomain(ctx, cr)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	req := clients.CreateGoalRequest{
		GoalType: cr.Spec.ForProvider.GoalType,
	}

	switch cr.Spec.ForProvider.GoalType {
	case "event":
		if cr.Spec.ForProvider.EventName == nil {
			return managed.ExternalCreation{}, errors.New("event name is required for event goals")
		}
		req.EventName = *cr.Spec.ForProvider.EventName
	case "page":
		if cr.Spec.ForProvider.PagePath == nil {
			return managed.ExternalCreation{}, errors.New("page path is required for page goals")
		}
		req.PagePath = *cr.Spec.ForProvider.PagePath
	}

	goal, err := c.service.CreateGoal(siteDomain, req)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "failed to create goal")
	}

	meta.SetExternalName(cr, goal.ID)

	return managed.ExternalCreation{}, nil
}

func (c *testExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// Goals cannot be updated in Plausible API
	return managed.ExternalUpdate{}, nil
}

func (c *testExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*goalv1beta1.Goal)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotGoal)
	}

	cr.SetConditions(xpv1.Deleting())

	err := c.service.DeleteGoal(meta.GetExternalName(cr))
	if err != nil && !clients.IsNotFound(err) {
		return managed.ExternalDelete{}, errors.Wrap(err, "failed to delete goal")
	}

	return managed.ExternalDelete{}, nil
}

func (c *testExternal) Disconnect(ctx context.Context) error {
	// Nothing to disconnect for Plausible API client
	return nil
}

// Mock service implementation
type mockPlausibleService struct {
	getGoalFn    func(siteDomain, goalID string) (*clients.Goal, error)
	listGoalsFn  func(siteDomain string) ([]clients.Goal, error)
	createGoalFn func(siteDomain string, req clients.CreateGoalRequest) (*clients.Goal, error)
	deleteGoalFn func(goalID string) error
}

func (m *mockPlausibleService) GetGoal(siteDomain, goalID string) (*clients.Goal, error) {
	if m.getGoalFn != nil {
		return m.getGoalFn(siteDomain, goalID)
	}
	return nil, nil
}

func (m *mockPlausibleService) ListGoals(siteDomain string) ([]clients.Goal, error) {
	if m.listGoalsFn != nil {
		return m.listGoalsFn(siteDomain)
	}
	return nil, nil
}

func (m *mockPlausibleService) CreateGoal(siteDomain string, req clients.CreateGoalRequest) (*clients.Goal, error) {
	if m.createGoalFn != nil {
		return m.createGoalFn(siteDomain, req)
	}
	return nil, nil
}

func (m *mockPlausibleService) DeleteGoal(goalID string) error {
	if m.deleteGoalFn != nil {
		return m.deleteGoalFn(goalID)
	}
	return nil
}

func TestObserve(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		o   managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		e    testExternal
		args args
		want want
	}{
		"GoalExistsWithExternalName": {
			e: testExternal{
				service: &mockPlausibleService{
					getGoalFn: func(siteDomain, goalID string) (*clients.Goal, error) {
						return &clients.Goal{
							ID:        "goal-123",
							GoalType:  "event",
							EventName: "signup",
							PagePath:  "",
						}, nil
					},
				},
			},
			args: args{
				mg: func() resource.Managed {
					goal := &goalv1beta1.Goal{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								meta.AnnotationKeyExternalName: "goal-123",
							},
						},
						Spec: goalv1beta1.GoalSpec{
							ForProvider: goalv1beta1.GoalParameters{
								SiteDomain: stringPtr("example.com"),
								GoalType:   "event",
								EventName:  stringPtr("signup"),
							},
						},
					}
					return goal
				}(),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"GoalDoesNotExistWithExternalName": {
			e: testExternal{
				service: &mockPlausibleService{
					getGoalFn: func(siteDomain, goalID string) (*clients.Goal, error) {
						return nil, nil
					},
				},
			},
			args: args{
				mg: func() resource.Managed {
					goal := &goalv1beta1.Goal{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								meta.AnnotationKeyExternalName: "goal-123",
							},
						},
						Spec: goalv1beta1.GoalSpec{
							ForProvider: goalv1beta1.GoalParameters{
								SiteDomain: stringPtr("example.com"),
								GoalType:   "event",
								EventName:  stringPtr("signup"),
							},
						},
					}
					return goal
				}(),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"GoalFoundByMatching": {
			e: testExternal{
				service: &mockPlausibleService{
					listGoalsFn: func(siteDomain string) ([]clients.Goal, error) {
						return []clients.Goal{
							{
								ID:        "goal-123",
								GoalType:  "event",
								EventName: "signup",
								PagePath:  "",
							},
						}, nil
					},
				},
			},
			args: args{
				mg: &goalv1beta1.Goal{
					Spec: goalv1beta1.GoalSpec{
						ForProvider: goalv1beta1.GoalParameters{
							SiteDomain: stringPtr("example.com"),
							GoalType:   "event",
							EventName:  stringPtr("signup"),
						},
					},
				},
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"GoalNotFoundByMatching": {
			e: testExternal{
				service: &mockPlausibleService{
					listGoalsFn: func(siteDomain string) ([]clients.Goal, error) {
						return []clients.Goal{
							{
								ID:        "goal-123",
								GoalType:  "event",
								EventName: "different",
								PagePath:  "",
							},
						}, nil
					},
				},
			},
			args: args{
				mg: &goalv1beta1.Goal{
					Spec: goalv1beta1.GoalSpec{
						ForProvider: goalv1beta1.GoalParameters{
							SiteDomain: stringPtr("example.com"),
							GoalType:   "event",
							EventName:  stringPtr("signup"),
						},
					},
				},
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"GetGoalFailed": {
			e: testExternal{
				service: &mockPlausibleService{
					getGoalFn: func(siteDomain, goalID string) (*clients.Goal, error) {
						return nil, errors.New("api error")
					},
				},
			},
			args: args{
				mg: func() resource.Managed {
					goal := &goalv1beta1.Goal{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								meta.AnnotationKeyExternalName: "goal-123",
							},
						},
						Spec: goalv1beta1.GoalSpec{
							ForProvider: goalv1beta1.GoalParameters{
								SiteDomain: stringPtr("example.com"),
								GoalType:   "event",
								EventName:  stringPtr("signup"),
							},
						},
					}
					return goal
				}(),
			},
			want: want{
				err: errors.New("api error"), // Just check for any error
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			o, err := tc.e.Observe(context.Background(), tc.args.mg)

			if tc.want.err != nil {
				if err == nil {
					t.Errorf("Observe(...): expected error, got nil")
				}
			} else if err != nil {
				t.Errorf("Observe(...): unexpected error: %v", err)
			}

			if err == nil {
				if diff := cmp.Diff(tc.want.o, o); diff != "" {
					t.Errorf("Observe(...): -want, +got:\n%s", diff)
				}
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		c   managed.ExternalCreation
		err error
	}

	cases := map[string]struct {
		e    testExternal
		args args
		want want
	}{
		"EventGoalCreated": {
			e: testExternal{
				service: &mockPlausibleService{
					createGoalFn: func(siteDomain string, req clients.CreateGoalRequest) (*clients.Goal, error) {
						return &clients.Goal{
							ID:        "goal-123",
							GoalType:  "event",
							EventName: req.EventName,
						}, nil
					},
				},
			},
			args: args{
				mg: &goalv1beta1.Goal{
					Spec: goalv1beta1.GoalSpec{
						ForProvider: goalv1beta1.GoalParameters{
							SiteDomain: stringPtr("example.com"),
							GoalType:   "event",
							EventName:  stringPtr("signup"),
						},
					},
				},
			},
			want: want{
				c: managed.ExternalCreation{},
			},
		},
		"PageGoalCreated": {
			e: testExternal{
				service: &mockPlausibleService{
					createGoalFn: func(siteDomain string, req clients.CreateGoalRequest) (*clients.Goal, error) {
						return &clients.Goal{
							ID:       "goal-123",
							GoalType: "page",
							PagePath: req.PagePath,
						}, nil
					},
				},
			},
			args: args{
				mg: &goalv1beta1.Goal{
					Spec: goalv1beta1.GoalSpec{
						ForProvider: goalv1beta1.GoalParameters{
							SiteDomain: stringPtr("example.com"),
							GoalType:   "page",
							PagePath:   stringPtr("/signup"),
						},
					},
				},
			},
			want: want{
				c: managed.ExternalCreation{},
			},
		},
		"EventGoalMissingEventName": {
			e: testExternal{
				service: &mockPlausibleService{},
			},
			args: args{
				mg: &goalv1beta1.Goal{
					Spec: goalv1beta1.GoalSpec{
						ForProvider: goalv1beta1.GoalParameters{
							SiteDomain: stringPtr("example.com"),
							GoalType:   "event",
						},
					},
				},
			},
			want: want{
				err: errors.New("event name is required for event goals"),
			},
		},
		"PageGoalMissingPagePath": {
			e: testExternal{
				service: &mockPlausibleService{},
			},
			args: args{
				mg: &goalv1beta1.Goal{
					Spec: goalv1beta1.GoalSpec{
						ForProvider: goalv1beta1.GoalParameters{
							SiteDomain: stringPtr("example.com"),
							GoalType:   "page",
						},
					},
				},
			},
			want: want{
				err: errors.New("page path is required for page goals"),
			},
		},
		"CreateFailed": {
			e: testExternal{
				service: &mockPlausibleService{
					createGoalFn: func(siteDomain string, req clients.CreateGoalRequest) (*clients.Goal, error) {
						return nil, errors.New("api error")
					},
				},
			},
			args: args{
				mg: &goalv1beta1.Goal{
					Spec: goalv1beta1.GoalSpec{
						ForProvider: goalv1beta1.GoalParameters{
							SiteDomain: stringPtr("example.com"),
							GoalType:   "event",
							EventName:  stringPtr("signup"),
						},
					},
				},
			},
			want: want{
				err: errors.New("api error"), // Just check for any error
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c, err := tc.e.Create(context.Background(), tc.args.mg)

			if tc.want.err != nil {
				if err == nil {
					t.Errorf("Create(...): expected error, got nil")
				}
			} else if err != nil {
				t.Errorf("Create(...): unexpected error: %v", err)
			}

			if err == nil {
				if diff := cmp.Diff(tc.want.c, c); diff != "" {
					t.Errorf("Create(...): -want, +got:\n%s", diff)
				}
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	e := testExternal{service: &mockPlausibleService{}}

	// Goals cannot be updated, should always return empty update
	u, err := e.Update(context.Background(), &goalv1beta1.Goal{})

	if err != nil {
		t.Errorf("Update(...): unexpected error: %v", err)
	}

	expected := managed.ExternalUpdate{}
	if diff := cmp.Diff(expected, u); diff != "" {
		t.Errorf("Update(...): -want, +got:\n%s", diff)
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		d   managed.ExternalDelete
		err error
	}

	cases := map[string]struct {
		e    testExternal
		args args
		want want
	}{
		"Successful": {
			e: testExternal{
				service: &mockPlausibleService{
					deleteGoalFn: func(goalID string) error {
						return nil
					},
				},
			},
			args: args{
				mg: func() resource.Managed {
					goal := &goalv1beta1.Goal{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								meta.AnnotationKeyExternalName: "goal-123",
							},
						},
					}
					return goal
				}(),
			},
			want: want{
				d: managed.ExternalDelete{},
			},
		},
		"DeleteFailed": {
			e: testExternal{
				service: &mockPlausibleService{
					deleteGoalFn: func(goalID string) error {
						return errors.New("api error")
					},
				},
			},
			args: args{
				mg: func() resource.Managed {
					goal := &goalv1beta1.Goal{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								meta.AnnotationKeyExternalName: "goal-123",
							},
						},
					}
					return goal
				}(),
			},
			want: want{
				err: errors.New("api error"), // Just check for any error
			},
		},
		"AlreadyDeleted": {
			e: testExternal{
				service: &mockPlausibleService{
					deleteGoalFn: func(goalID string) error {
						return errors.New("API request failed: status 404")
					},
				},
			},
			args: args{
				mg: func() resource.Managed {
					goal := &goalv1beta1.Goal{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								meta.AnnotationKeyExternalName: "goal-123",
							},
						},
					}
					return goal
				}(),
			},
			want: want{
				d: managed.ExternalDelete{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			d, err := tc.e.Delete(context.Background(), tc.args.mg)

			if tc.want.err != nil {
				if err == nil {
					t.Errorf("Delete(...): expected error, got nil")
				}
			} else if err != nil {
				t.Errorf("Delete(...): unexpected error: %v", err)
			}

			if err == nil {
				if diff := cmp.Diff(tc.want.d, d); diff != "" {
					t.Errorf("Delete(...): -want, +got:\n%s", diff)
				}
			}
		})
	}
}

func TestGetSiteDomain(t *testing.T) {
	cases := map[string]struct {
		goal     *goalv1beta1.Goal
		mockSite *sitev1beta1.Site
		want     string
		wantErr  bool
	}{
		"DirectDomain": {
			goal: &goalv1beta1.Goal{
				Spec: goalv1beta1.GoalSpec{
					ForProvider: goalv1beta1.GoalParameters{
						SiteDomain: stringPtr("example.com"),
					},
				},
			},
			want: "example.com",
		},
		"DomainReference": {
			goal: &goalv1beta1.Goal{
				Spec: goalv1beta1.GoalSpec{
					ForProvider: goalv1beta1.GoalParameters{
						SiteDomainRef: &xpv1.Reference{
							Name: "test-site",
						},
					},
				},
			},
			mockSite: &sitev1beta1.Site{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-site",
				},
				Spec: sitev1beta1.SiteSpec{
					ForProvider: sitev1beta1.SiteParameters{
						Domain: "example.com",
					},
				},
			},
			want: "example.com",
		},
		"SelectorNotSupported": {
			goal: &goalv1beta1.Goal{
				Spec: goalv1beta1.GoalSpec{
					ForProvider: goalv1beta1.GoalParameters{
						SiteDomainSelector: &xpv1.Selector{},
					},
				},
			},
			wantErr: true,
		},
		"NoSiteDomainSpecified": {
			goal: &goalv1beta1.Goal{
				Spec: goalv1beta1.GoalSpec{
					ForProvider: goalv1beta1.GoalParameters{},
				},
			},
			wantErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			kube := &test.MockClient{
				MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
					if tc.mockSite != nil {
						site := obj.(*sitev1beta1.Site)
						tc.mockSite.DeepCopyInto(site)
						return nil
					}
					return errors.New("site not found")
				},
			}

			e := testExternal{
				kube: kube,
			}

			got, err := e.getSiteDomain(context.Background(), tc.goal)

			if tc.wantErr {
				if err == nil {
					t.Errorf("getSiteDomain(...): expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("getSiteDomain(...): unexpected error: %v", err)
				return
			}

			if got != tc.want {
				t.Errorf("getSiteDomain(...): got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestGoalMatches(t *testing.T) {
	cases := map[string]struct {
		cr   *goalv1beta1.Goal
		goal *clients.Goal
		want bool
	}{
		"EventGoalMatches": {
			cr: &goalv1beta1.Goal{
				Spec: goalv1beta1.GoalSpec{
					ForProvider: goalv1beta1.GoalParameters{
						GoalType:  "event",
						EventName: stringPtr("signup"),
					},
				},
			},
			goal: &clients.Goal{
				GoalType:  "event",
				EventName: "signup",
			},
			want: true,
		},
		"EventGoalDoesNotMatch": {
			cr: &goalv1beta1.Goal{
				Spec: goalv1beta1.GoalSpec{
					ForProvider: goalv1beta1.GoalParameters{
						GoalType:  "event",
						EventName: stringPtr("signup"),
					},
				},
			},
			goal: &clients.Goal{
				GoalType:  "event",
				EventName: "login",
			},
			want: false,
		},
		"PageGoalMatches": {
			cr: &goalv1beta1.Goal{
				Spec: goalv1beta1.GoalSpec{
					ForProvider: goalv1beta1.GoalParameters{
						GoalType: "page",
						PagePath: stringPtr("/signup"),
					},
				},
			},
			goal: &clients.Goal{
				GoalType: "page",
				PagePath: "/signup",
			},
			want: true,
		},
		"PageGoalDoesNotMatch": {
			cr: &goalv1beta1.Goal{
				Spec: goalv1beta1.GoalSpec{
					ForProvider: goalv1beta1.GoalParameters{
						GoalType: "page",
						PagePath: stringPtr("/signup"),
					},
				},
			},
			goal: &clients.Goal{
				GoalType: "page",
				PagePath: "/login",
			},
			want: false,
		},
		"TypeMismatch": {
			cr: &goalv1beta1.Goal{
				Spec: goalv1beta1.GoalSpec{
					ForProvider: goalv1beta1.GoalParameters{
						GoalType:  "event",
						EventName: stringPtr("signup"),
					},
				},
			},
			goal: &clients.Goal{
				GoalType: "page",
				PagePath: "/signup",
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := testExternal{}
			got := e.goalMatches(tc.cr, tc.goal)

			if got != tc.want {
				t.Errorf("goalMatches(...): got %v, want %v", got, tc.want)
			}
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}