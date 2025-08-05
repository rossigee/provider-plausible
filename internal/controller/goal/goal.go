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

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	goalv1alpha1 "github.com/rossigee/provider-plausible/apis/goal/v1alpha1"
	sitev1alpha1 "github.com/rossigee/provider-plausible/apis/site/v1alpha1"
	"github.com/rossigee/provider-plausible/apis/v1beta1"
	"github.com/rossigee/provider-plausible/internal/clients"
)

const (
	errNotGoal      = "managed resource is not a Goal custom resource"
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	errGetCreds     = "cannot get credentials"

	errNewClient = "cannot create new Service"
	errGetSite   = "cannot get referenced Site"
)

// Setup adds a controller that reconciles Goal managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(goalv1alpha1.GoalGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	// TODO: Add support for alpha management policies
	// if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
	// 	cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1beta1.ProviderConfigUsageGroupVersionKind))
	// }

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(goalv1alpha1.GoalGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:         mgr.GetClient(),
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &v1beta1.ProviderConfigUsage{}),
			newServiceFn: clients.NewClient,
		}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&goalv1alpha1.Goal{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube         client.Client
	usage        resource.Tracker
	newServiceFn func(config clients.Config) *clients.Client
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*goalv1alpha1.Goal)
	if !ok {
		return nil, errors.New(errNotGoal)
	}

	cfg, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}

	svc := c.newServiceFn(*cfg)

	return &external{service: svc, kube: c.kube}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	service *clients.Client
	kube    client.Client
}

func (c *external) getSiteDomain(ctx context.Context, cr *goalv1alpha1.Goal) (string, error) {
	// If direct domain is specified, use it
	if cr.Spec.ForProvider.SiteDomain != nil && *cr.Spec.ForProvider.SiteDomain != "" {
		return *cr.Spec.ForProvider.SiteDomain, nil
	}

	// If reference is specified, resolve it
	if cr.Spec.ForProvider.SiteDomainRef != nil {
		site := &sitev1alpha1.Site{}
		nn := types.NamespacedName{
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

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*goalv1alpha1.Goal)
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

		cr.Status.AtProvider = goalv1alpha1.GoalObservation{
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

			cr.Status.AtProvider = goalv1alpha1.GoalObservation{
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

func (c *external) goalMatches(cr *goalv1alpha1.Goal, goal *clients.Goal) bool {
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

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*goalv1alpha1.Goal)
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

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// Goals cannot be updated in Plausible API
	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*goalv1alpha1.Goal)
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

func (c *external) Disconnect(ctx context.Context) error {
	// Nothing to disconnect for Plausible API client
	return nil
}