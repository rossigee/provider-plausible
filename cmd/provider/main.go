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

package main

import (
	"os"
	"path/filepath"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	xpcontroller "github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/feature"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"

	"github.com/crossplane-contrib/provider-plausible/apis"
	"github.com/crossplane-contrib/provider-plausible/internal/controller"
	"github.com/crossplane-contrib/provider-plausible/internal/features"
)

func main() {
	var (
		app                = kingpin.New(filepath.Base(os.Args[0]), "Plausible support for Crossplane.").DefaultEnvars()
		debug              = app.Flag("debug", "Run with debug logging.").Short('d').Bool()
		leaderElection     = app.Flag("leader-election", "Use leader election for the controller manager.").Short('l').Default("false").OverrideDefaultFromEnvar("LEADER_ELECTION").Bool()
		leaderElectionNS   = app.Flag("leader-election-namespace", "Namespace to use for leader election.").Default("crossplane-system").OverrideDefaultFromEnvar("LEADER_ELECTION_NAMESPACE").String()
		pollInterval       = app.Flag("poll", "How often individual resources will be checked for drift from the desired state").Short('p').Default("1m").Duration()
		maxReconcileRate   = app.Flag("max-reconcile-rate", "The global maximum rate per second at which resources may checked for drift from the desired state.").Default("10").Int()
		syncPeriod         = app.Flag("sync", "How often all resources will be double-checked for drift from the desired state.").Short('s').Default("1h").Duration()
		enableManagementPolicies = app.Flag("enable-management-policies", "Enable support for management policies.").Default("true").OverrideDefaultFromEnvar("ENABLE_MANAGEMENT_POLICIES").Bool()
	)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	zl := zap.New(zap.UseDevMode(*debug))
	log := logging.NewLogrLogger(zl.WithName("provider-plausible"))
	if *debug {
		// The controller-runtime runs with a no-op logger by default. It is
		// *very* verbose even at info level, so we only provide it a real
		// logger when we're running in debug mode.
		ctrl.SetLogger(zl)
	}

	log.Debug("Starting", "sync-period", syncPeriod.String())

	cfg, err := ctrl.GetConfig()
	if err != nil {
		kingpin.FatalIfError(err, "Cannot get API server rest config")
	}

	mgr, err := ctrl.NewManager(ratelimiter.LimitRESTConfig(cfg, *maxReconcileRate), ctrl.Options{
		Cache: cache.Options{
			SyncPeriod: syncPeriod,
		},
		LeaderElection:             *leaderElection,
		LeaderElectionID:           "crossplane-leader-election-provider-plausible",
		LeaderElectionNamespace:    *leaderElectionNS,
		LeaderElectionResourceLock: "leases",
		LeaseDuration:              func() *time.Duration { d := 60 * time.Second; return &d }(),
		RenewDeadline:              func() *time.Duration { d := 50 * time.Second; return &d }(),
	})
	if err != nil {
		kingpin.FatalIfError(err, "Cannot create controller manager")
	}

	o := xpcontroller.Options{
		Logger:                  log,
		MaxConcurrentReconciles: *maxReconcileRate,
		PollInterval:            *pollInterval,
		GlobalRateLimiter:       ratelimiter.NewGlobal(*maxReconcileRate),
		Features:                &feature.Flags{},
	}

	if *enableManagementPolicies {
		o.Features.Enable(features.EnableAlphaManagementPolicies)
		log.Info("Alpha feature enabled", "flag", features.EnableAlphaManagementPolicies)
	}

	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		kingpin.FatalIfError(err, "Cannot add Plausible APIs to scheme")
	}

	if err := controller.Setup(mgr, o); err != nil {
		kingpin.FatalIfError(err, "Cannot setup Plausible controllers")
	}

	kingpin.FatalIfError(mgr.Start(ctrl.SetupSignalHandler()), "Cannot start controller manager")
}