# Copyright 2021 Tetrate
# Licensed under the Apache License, Version 2.0 (the "License")
#
# This script uses automatic variables (ex $<, $(@D)) and substitution references $(<:.signed=)
# Please see GNU make's documentation if unfamiliar: https://www.gnu.org/software/make/manual/html_node/
.PHONY: test build e2e dist clean format lint check site

# Make functions strip spaces and use commas to separate parameters. The below variables escape these characters.
comma := ,
space :=
space +=

# Include versions of tools we build on-demand
include Tools.mk

# This should be driven by automation and result in N.N.N, not vN.N.N
VERSION ?= dev

# This selects the goroot to use in the following priority order:
# 1. ${GOROOT}          - Ex actions/setup-go
# 2. ${GOROOT_1_17_X64} - Ex GitHub Actions runner
# 3. $(go env GOROOT)   - Implicit from the go binary in the path
#
# There may be multiple GOROOT variables, so pick the one matching go.mod.
go_release          := $(shell sed -ne 's/^go //gp' go.mod)
# https://github.com/actions/runner/blob/master/src/Runner.Common/Constants.cs
github_runner_arch  := $(if $(findstring $(shell uname -m),x86_64),X64,ARM64)
github_goroot_name  := GOROOT_$(subst .,_,$(go_release))_$(github_runner_arch)
github_goroot_val   := $(value $(github_goroot_name))
goroot_path         := $(shell go env GOROOT 2>/dev/null)
goroot              := $(firstword $(GOROOT) $(github_goroot_val) $(goroot_path))

ifndef goroot
$(error could not determine GOROOT)
endif

# Ensure POSIX-style GOROOT even in Windows, to support PATH updates in bash.
ifdef COMSPEC
goroot := $(shell cygpath $(goroot))
endif

# We must ensure `go` executes with GOROOT and PATH variables exported:
# * GOROOT ensures versions don't conflict with /usr/local/go or c:\Go
# * PATH ensures tools like golint can fork and execute the correct go binary.
#
# We may be using a very old version of Make (ex. 3.81 on macOS). This means we
# can't re-set GOROOT or PATH via 'export' or use '.ONESHELL' to persist
# variables across lines. Hence, we set variables on one-line.
go := export PATH="$(goroot)/bin:$${PATH}" && export GOROOT="$(goroot)" && go

# Set variables corresponding to the selected goroot and the current host.
goarch := $(shell $(go) env GOARCH)
goexe  := $(shell $(go) env GOEXE)
goos   := $(shell $(go) env GOOS)

# Build the path to the func-e binary for the current runtime (goos,goarch)
current_binary_path := build/func-e_$(goos)_$(goarch)
current_binary      := $(current_binary_path)/func-e$(goexe)

# ANSI escape codes. f_ means foreground, b_ background.
# See https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_(Select_Graphic_Rendition)_parameters
b_black            := $(shell printf "\33[40m")
f_white            := $(shell printf "\33[97m")
f_gray             := $(shell printf "\33[37m")
f_dark_gray        := $(shell printf "\33[90m")
f_bright_magenta   := $(shell printf "\33[95m")
b_bright_magenta   := $(shell printf "\33[105m")
ansi_reset         := $(shell printf "\33[0m")
ansi_func_e        := $(b_black)$(f_white)func-$(b_bright_magenta)e$(ansi_reset)
ansi_format_dark   := $(f_gray)$(f_bright_magenta)%-10s$(ansi_reset) $(f_dark_gray)%s$(ansi_reset)\n
ansi_format_bright := $(f_white)$(f_bright_magenta)%-10s$(ansi_reset) $(f_white)$(b_bright_magenta)%s$(ansi_reset)\n

# This formats help statements in ANSI colors. To hide a target from help, don't comment it with a trailing '##'.
help: ## Describe how to use each target
	@printf "$(ansi_func_e)$(f_white)\n"
	@awk 'BEGIN {FS = ":.*?## "} /^[0-9a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "$(ansi_format_dark)", $$1, $$2}' $(MAKEFILE_LIST)

build: $(current_binary) ## Build the func-e binary

test: ## Run all unit tests
	@printf "$(ansi_format_dark)" test "running unit tests"
	@$(go) test $(main_packages)
	@printf "$(ansi_format_bright)" test "ok"

coverpkg = $(subst $(space),$(comma),$(main_packages))
coverage: ## Generate test coverage
	@printf "$(ansi_format_dark)" coverage "running unit tests with coverage"
	@$(go) test -coverprofile=coverage.txt -covermode=atomic --coverpkg=$(coverpkg) $(main_packages)
	@$(go) tool cover -func coverage.txt
	@printf "$(ansi_format_bright)" coverage "ok"

# Tests run one at a time, in verbose mode, so that failures are easy to diagnose.
# Note: -failfast helps as it stops at the first error. However, it is not a cacheable flag, so runs won't cache.
export E2E_FUNC_E_PATH ?= $(current_binary_path)
e2e: $(E2E_FUNC_E_PATH)/func-e$(goexe) ## Run all end-to-end tests
	@printf "$(ansi_format_dark)" e2e "running end-to-end tests"
	@$(go) test -parallel 1 -v -failfast ./e2e
	@printf "$(ansi_format_bright)" e2e "ok"

linux_platforms       := linux_amd64 linux_arm64
non_windows_platforms := darwin_amd64 darwin_arm64 $(linux_platforms)
# TODO: arm64 on Windows https://github.com/envoyproxy/envoy/issues/17572
windows_platforms     := windows_amd64

# Make 3.81 doesn't support '**' globbing: Set explicitly instead of recursion.
all_sources   := $(wildcard *.go */*.go */*/*.go */*/*/*.go */*/*/*.go */*/*/*/*.go)
all_testdata  := $(wildcard testdata/* */testdata/* */*/testdata/* */*/*/testdata/*)
all_testutil  := $(wildcard internal/test/* internal/test/*/* internal/test/*/*/*)
# main_sources compose the binary, so exclude tests, test utilities and linters
main_sources  := $(wildcard $(filter-out %_test.go $(all_testdata) $(all_testutil) $(wildcard lint/*), $(all_sources)))
# main_packages collect the unique main source directories (sort will dedupe).
# Paths need to all start with ./, so we do that manually vs foreach which strips it.
main_packages := $(sort $(foreach f,$(dir $(main_sources)),$(if $(findstring ./,$(f)),./,./$(f))))

build/func-e_%/func-e: $(main_sources)
	$(call go-build,$@,$<)

dist/func-e_$(VERSION)_%.tar.gz: build/func-e_%/func-e
	@printf "$(ansi_format_dark)" tar.gz "tarring $@"
	@mkdir -p $(@D)
	@tar -C $(<D) -cpzf $@ $(<F)
	@printf "$(ansi_format_bright)" tar.gz "ok"

build/func-e_%/func-e.exe: $(main_sources)
	$(call go-build,$@,$<)

dist/func-e_$(VERSION)_%.zip: build/func-e_%/func-e.exe.signed
	@printf "$(ansi_format_dark)" zip "zipping $@"
	@mkdir -p $(@D)
ifeq ($(goos),windows)  # Windows 10 etc use 7zip
	@# '-bso0 -bsp0' makes it quiet, except errors
	@# Wildcards in 7zip will skip the directory. We single-quote to ensure they aren't interpreted by the shell.
	@7z -bso0 -bsp0 a $@ './$(dir $<)/*.exe'
else  # Otherwise, assume zip is available
	@zip -qj $@ $(<:.signed=)
endif
	@printf "$(ansi_format_bright)" zip "ok"

# Default to a dummy version, which is always lower than a real release
nfpm_version=v$(VERSION:dev=0.0.1)

# It is not precise to put the func-e binary here, but it is easier because the arch pattern matches
# whereas in RPM it won't.
# Note: we are only generating this because the file isn't parameterized.
# See https://github.com/goreleaser/nfpm/issues/362
build/func-e_linux_%/nfpm.yaml: packaging/nfpm/nfpm.yaml build/func-e_linux_%/func-e
	@mkdir -p $(@D)
	@sed -e 's/amd64/$(*)/g' -e 's/v0.0.1/$(nfpm_version)/g' $< > $@

# We can't use a pattern (%) rule because in RPM amd64 -> x86_64, arm64 -> aarch64
rpm_x86_64  := dist/func-e_$(VERSION)_linux_x86_64.rpm
rpm_aarch64 := dist/func-e_$(VERSION)_linux_aarch64.rpm
rpms        := $(rpm_x86_64) $(rpm_aarch64)

man_page := packaging/nfpm/func-e.8

$(rpm_x86_64): build/func-e_linux_amd64/nfpm.yaml $(man_page)
	$(call nfpm-pkg,$<,"rpm",$@)

$(rpm_aarch64): build/func-e_linux_arm64/nfpm.yaml $(man_page)
	$(call nfpm-pkg,$<,"rpm",$@)

# Debian architectures map goarch for amd64 and arm64
dist/func-e_$(VERSION)_linux_%.deb: build/func-e_linux_%/nfpm.yaml $(man_page)
	$(call nfpm-pkg,$<,"deb",$@)

# msi-arch is a macro so we can detect it based on the file naming convention
msi-arch     = $(if $(findstring amd64,$1),x64,arm64)
# Default to a dummy version, which is always lower than a real release
msi_version := $(VERSION:dev=0.0.1)

# This builds the Windows installer (MSI) using platform-dependent WIX commands.
dist/func-e_$(VERSION)_%.msi: build/func-e_%/func-e.exe.signed
	@printf "$(ansi_format_dark)" msi "building $@"
	@mkdir -p $(@D)
ifeq ($(goos),windows)  # Windows 10 etc use https://wixtoolset.org
	@candle -nologo -arch $(call msi-arch,$@) -dVersion=$(msi_version) -dBin=$(<:.signed=) packaging/msi/func-e.wxs
	@light -nologo func-e.wixobj -o $@ -spdb
	@rm func-e.wixobj
else  # use https://wiki.gnome.org/msitools
	@wixl -a $(call msi-arch,$@) -D Version=$(msi_version) -D Bin=$(<:.signed=) -o $@ packaging/msi/func-e.wxs
endif
	$(call codesign,$@)
	@printf "$(ansi_format_bright)" msi "ok"

# Archives are tar.gz, except in the case of Windows, which uses zip.
archives  := $(non_windows_platforms:%=dist/func-e_$(VERSION)_%.tar.gz) $(windows_platforms:%=dist/func-e_$(VERSION)_%.zip)
packages  := $(windows_platforms:%=dist/func-e_$(VERSION)_%.msi) $(linux_platforms:%=dist/func-e_$(VERSION)_%.deb) $(rpms)
checksums := dist/func-e_$(VERSION)_checksums.txt

# Darwin doesn't have sha256sum. See https://github.com/actions/virtual-environments/issues/90
sha256sum := $(if $(findstring darwin,$(goos)),shasum -a 256,sha256sum)
$(checksums): $(archives) $(packages)
	@printf "$(ansi_format_dark)" sha256sum "generating $@"
	@$(sha256sum) $^ > $@
	@printf "$(ansi_format_bright)" sha256sum "ok"

# dist generates the assets that attach to a release
# Ex. https://github.com/tetratelabs/func-e/releases/tag/v$(VERSION)
dist: $(archives) $(packages) $(checksums) ## Generate release assets

clean: ## Ensure a clean build
	@printf "$(ansi_format_dark)" clean "deleting temporary files"
	@rm -rf dist build coverage.txt
	@$(go) clean -testcache
	@printf "$(ansi_format_bright)" clean "ok"

# format is a PHONY target, so always runs. This allows skipping when sources didn't change.
build/format: go.mod $(all_sources)
	@$(go) mod tidy
	@$(go) run $(licenser) apply -r "Tetrate"
	@$(go)fmt -s -w $(all_sources)
	@# Workaround inconsistent goimports grouping with awk until golang/go#20818 or incu6us/goimports-reviser#50
	@for f in $(all_sources); do \
	    awk '/^import \($$/,/^\)$$/{if($$0=="")next}{print}' $$f > /tmp/fmt; \
	    mv /tmp/fmt $$f; \
	done
	@# -local ensures consistent ordering of our module in imports
	@$(go) run $(goimports) -local $$(sed -ne 's/^module //gp' go.mod) -w $(all_sources)
	@mkdir -p $(@D) && touch $@

format:
	@printf "$(ansi_format_dark)" format "formatting project files"
	@$(MAKE) build/format
	@printf "$(ansi_format_bright)" format "ok"

# lint is a PHONY target, so always runs. This allows skipping when sources didn't change.
build/lint: .golangci.yml $(all_sources)
	@$(go) run $(golangci_lint) run --timeout 5m --config $< ./...
	@$(go) test ./lint/...
	@mkdir -p $(@D) && touch $@

lint:
	@printf "$(ansi_format_dark)" lint "Running linters"
	@$(MAKE) build/lint
	@printf "$(ansi_format_bright)" lint "ok"

# CI blocks merge until this passes. If this fails, run "make check" locally and commit the difference.
# This formats code before running lint, as it is annoying to tell people to format first!
check: ## Verify contents of last commit
	@$(MAKE) lint
	@$(MAKE) format
	@# Make sure the check-in is clean
	@if [ ! -z "`git status -s`" ]; then \
		echo "The following differences will fail CI until committed:"; \
		git diff --exit-code; \
	fi

site: ## Serve website content
	@git submodule update
	@cd site && $(go) run $(hugo) server --minify --disableFastRender --baseURL localhost:1313 --cleanDestinationDir -D

# this makes a marker file ending in .signed to avoid repeatedly calling codesign
%.signed: %
	$(call codesign,$<)
	@touch $@

# define macros for multi-platform builds. these parse the filename being built
go-arch = $(if $(findstring amd64,$1),amd64,arm64)
go-os   = $(if $(findstring .exe,$1),windows,$(if $(findstring linux,$1),linux,darwin))
define go-build
	@printf "$(ansi_format_dark)" build "building $1"
	@# $(go:go=) removes the trailing 'go', so we can insert cross-build variables
	@$(go:go=) CGO_ENABLED=0 GOOS=$(call go-os,$1) GOARCH=$(call go-arch,$1) go build \
		-ldflags "-s -w -X main.version=$(VERSION)" \
		-o $1 $2
	@printf "$(ansi_format_bright)" build "ok"
endef

define nfpm-pkg
	@printf "$(ansi_format_dark)" nfpm "packaging $3"
	@mkdir -p $(dir $3)
	@$(go) run $(nfpm) pkg -f $1 --packager $2 --target $3
	@printf "$(ansi_format_bright)" nfpm "ok"
endef

# This requires osslsigncode package (apt or brew) or latest windows release from mtrojnar/osslsigncode
#
# Default is self-signed while production should be a Digicert signing key
#
# Ex.
# ```bash
# keytool -genkey -alias func-e -storetype PKCS12 -keyalg RSA -keysize 2048 -storepass func-e-bunch \
# -keystore func-e.p12 -dname "O=func-e,CN=func-e.io" -validity 3650
# ```
WINDOWS_CODESIGN_P12      ?= packaging/msi/func-e.p12
WINDOWS_CODESIGN_PASSWORD ?= func-e-bunch
define codesign
	@printf "$(ansi_format_dark)" codesign "signing $1"
	@osslsigncode sign -h sha256 -pkcs12 ${WINDOWS_CODESIGN_P12} -pass "${WINDOWS_CODESIGN_PASSWORD}" \
	-n "func-e makes running Envoy® easy" -i https://func-e.io -t http://timestamp.digicert.com \
	$(if $(findstring msi,$(1)),-add-msi-dse) -in $1 -out $1-signed
	@mv $1-signed $1
	@printf "$(ansi_format_bright)" codesign "ok"
endef
