# Project Setup
PROJECT_NAME := provider-plausible
PROJECT_REPO := github.com/crossplane-contrib/$(PROJECT_NAME)

PLATFORMS ?= linux_amd64 linux_arm64
-include build/makelib/common.mk

# Setup Output
-include build/makelib/output.mk

# Setup Go
NPROCS ?= 1
GO_TEST_PARALLEL := $(shell echo $$(( $(NPROCS) / 2 )))
GO_STATIC_PACKAGES = $(GO_PROJECT)/cmd/provider
GO_LDFLAGS += -X $(GO_PROJECT)/internal/version.Version=$(VERSION)
GO_SUBDIRS += cmd internal apis
GO111MODULE = on
-include build/makelib/golang.mk

# Setup Kubernetes tools
UP_VERSION = v0.28.0
UP_CHANNEL = stable
UPTEST_VERSION = v0.11.1
-include build/makelib/k8s_tools.mk

# Setup Images
IMAGES = provider-plausible
-include build/makelib/imagelight.mk

# Setup XPKG
XPKG_REG_ORGS ?= xpkg.upbound.io/crossplane-contrib
# NOTE: skip promoting on xpkg.upbound.io as channel tags are inferred.
XPKG_REG_ORGS_NO_PROMOTE ?= xpkg.upbound.io/crossplane-contrib
XPKGS = provider-plausible
-include build/makelib/xpkg.mk

# NOTE: we force image building to happen prior to xpkg build so that we ensure
# image is present in daemon.
xpkg.build.provider-plausible: do.build.images

# Setup Package Metadata
export CROSSPLANE_VERSION := $(shell go list -m -f '{{.Version}}' github.com/crossplane/crossplane)
-include build/makelib/local.xpkg.mk
-include build/makelib/controlplane.mk

# Targets

# run `make submodules` after cloning the repository for the first time.
submodules:
	@git submodule sync
	@git submodule update --init --recursive

# NOTE: the build submodule currently overrides XDG_CACHE_HOME in order to
# force the Helm 3 to use the .work/helm directory. This causes Go on Linux
# machines to use that directory as the build cache as well. We should adjust
# this behavior in the build submodule because it is also causing Linux users
# to duplicate their build cache, but for now we just make it easier to identify
# its location in CI so that we cache between builds.
go.cachedir:
	@go env GOCACHE

# Generate deepcopy, managed resource methods, CRDs, RBAC role.
generate: generate.init
	@$(MAKE) go.generate

# Generate deepcopy, managed resource methods, CRDs, RBAC role without
# dependencies.
generate.init:
	@$(INFO) Initializing code generation tools
	@cd $(TOOLS_HOST_DIR) && $(GOHOST) install -tags generate sigs.k8s.io/controller-tools/cmd/controller-gen@v0.13.0
	@cd $(TOOLS_HOST_DIR) && $(GOHOST) install -tags generate github.com/crossplane/crossplane-tools/cmd/angryjet@master

# NOTE: we must ensure up is installed in tool cache prior to build as including the k8s_tools
# machinery prior to the xpkg machinery sets UP to point to tool cache.
build.init: $(UP)

# This is for running out-of-cluster locally, and is for convenience. Running
# this make target will print out the command which was used. For more control,
# try running the binary directly with different arguments.
run: go.build
	@$(INFO) Running Crossplane locally out-of-cluster . . .
	@# To see other arguments that can be provided, run the command with --help instead
	$(GO_OUT_DIR)/provider --debug

# NOTE: we ensure up is installed prior to running platform-specific packaging steps in xpkg.build.
xpkg.build: $(UP)

.PHONY: submodules generate.init run

# Additional targets

# Run tests
test: generate
	@$(INFO) Running tests...
	@$(GO) test -v ./...

# Run tests with coverage
test.cover: generate
	@$(INFO) Running tests with coverage...
	@$(GO) test -v -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html

# Install CRDs into a cluster
install-crds: generate
	kubectl apply -f package/crds

# Uninstall CRDs from a cluster
uninstall-crds:
	kubectl delete -f package/crds