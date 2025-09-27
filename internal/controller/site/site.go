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

	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"

	sitev1beta1 "github.com/rossigee/provider-plausible/apis/site/v1beta1"
	"github.com/rossigee/provider-plausible/internal/clients"
)

const (
	errNotSite      = "managed resource is not a Site custom resource"
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	errGetCreds     = "cannot get credentials"

)

// Setup adds a controller that reconciles Site managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(sitev1beta1.SiteGroupKind)


	r := managed.NewReconciler(mgr,
		resource.ManagedKind(sitev1beta1.SiteGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:         mgr.GetClient(),
			usage:        clients.NewProviderConfigUsageTracker(mgr.GetClient()),
			newServiceFn: clients.NewClient,
		}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&sitev1beta1.Site{}).
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
	_, ok := mg.(*sitev1beta1.Site)
	if !ok {
		return nil, errors.New(errNotSite)
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

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*sitev1beta1.Site)
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

		cr.Status.AtProvider = sitev1beta1.SiteObservation{
			ID:     site.ID,
			Domain: site.Domain,
			TeamID: site.TeamID,
		}

		cr.SetConditions(xpv1.Available())
		cr.SetConditions(xpv1.ReconcileSuccess())

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

	cr.Status.AtProvider = sitev1beta1.SiteObservation{
		ID:     site.ID,
		Domain: site.Domain,
		TeamID: site.TeamID,
	}

	cr.SetConditions(xpv1.Available())
	cr.SetConditions(xpv1.ReconcileSuccess())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: c.isUpToDate(cr, site),
	}, nil
}

func (c *external) isUpToDate(cr *sitev1beta1.Site, site *clients.Site) bool {
	// Check if domain needs to be updated
	if cr.Spec.ForProvider.NewDomain != nil && *cr.Spec.ForProvider.NewDomain != site.Domain {
		return false
	}

	// Note: Team ID and timezone cannot be updated after creation via API
	return true
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*sitev1beta1.Site)
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

	// Return connection details for the created site
	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{
			"siteId": []byte(site.ID),
			"domain": []byte(site.Domain),
		},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*sitev1beta1.Site)
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

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*sitev1beta1.Site)
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

func (c *external) Disconnect(ctx context.Context) error {
	// Nothing to disconnect for Plausible API client
	return nil
}