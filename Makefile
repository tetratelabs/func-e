# Copyright 2019 Tetrate
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Make sure we pick up any local overrides.
-include .makerc

goimports := golang.org/x/tools/cmd/goimports@v0.1.5
golangci_lint := github.com/golangci/golangci-lint/cmd/golangci-lint@v1.42.0
goreleaser := github.com/goreleaser/goreleaser@v0.175.0
hugo := github.com/gohugoio/hugo@v0.87.0
licenser := github.com/liamawhite/licenser@v0.6.0

##@ Binary distribution

.PHONY: release
release:
	@echo "--- release ---"
	@go run $(goreleaser) release --rm-dist

GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
BIN := dist/func-e_$(GOOS)_$(GOARCH)
bin $(BIN):
	@echo "--- bin ---"
# skip post hooks on bin so that e2e tests don't need to have osslsigncode installed
	@go run $(goreleaser) build --snapshot --single-target --skip-post-hooks --rm-dist

# Requires `wixl` from msitools https://wiki.gnome.org/msitools (or `brew install msitools`)
# If Windows, you can download from here https://github.com/wixtoolset/wix3/releases
WIN_BIN := dist/func-e_windows_amd64
WIN_BIN_EXE := $(WIN_BIN)/func-e.exe
# Default to a dummy version, but in a release this should be overridden
MSI_VERSION ?= 0.0.1

# Right now, the only arch we support is amd64 because Envoy doesn't yet support arm64 on Windows
# https://github.com/envoyproxy/envoy/issues/17572
# Once that occurs, we will need to set -arch arm64 and bundle accordingly.
$(WIN_BIN_EXE):
	@echo "--- win-bin ---"
	@GOOS=windows GOARCH=amd64 @go run $(goreleaser) build --snapshot --single-target --rm-dist

# Default is self-signed while production should be a Digicert signing key
#
# Ex.
# ```bash
# keytool -genkey -alias func-e -storetype PKCS12 -keyalg RSA -keysize 2048 -storepass func-e-bunch \
# -keystore func-e.p12 -dname "O=func-e,CN=func-e.io" -validity 3650
# ```
WINDOWS_CODESIGN_P12 ?= packaging/msi/func-e.p12
WINDOWS_CODESIGN_PASSWORD ?= func-e-bunch

# This is invoked as a part of goreleaser to make sure the binary is signed on release.
#
# This requires osslsigncode package (apt or brew) or latest windows release from mtrojnar/osslsigncode
.PHONY: sign-win
sign-win: $(WIN_BIN_EXE)
	@osslsigncode sign -h sha256 -pkcs12 ${WINDOWS_CODESIGN_P12} -pass "${WINDOWS_CODESIGN_PASSWORD}" \
	-n "func-e makes running Envoy® easy" -i https://func-e.io -t http://timestamp.digicert.com \
	-in $(WIN_BIN_EXE) -out $(WIN_BIN_EXE)-signed.exe
	@mv $(WIN_BIN_EXE)-signed.exe $(WIN_BIN_EXE)

# Right now, the only arch we support is amd64 because Envoy doesn't yet support arm64 on Windows
# https://github.com/envoyproxy/envoy/issues/17572
# Once that occurs, we will need to set -arch arm64 and bundle accordingly.
.PHONY: msi
msi: $(WIN_BIN_EXE)
ifeq ($(OS),Windows_NT)  # Windows 10 etc use https://wixtoolset.org
	@candle -nologo -arch x64 -dVersion=$(MSI_VERSION) \
	-dBin=$(WIN_BIN)/func-e.exe \
	packaging/msi/func-e.wxs
	@light -nologo func-e.wixobj -o $(WIN_BIN)/func-e.msi -spdb
	@rm func-e.wixobj
else  # use https://wiki.gnome.org/msitools
	@wixl -a x64 -D Version=$(MSI_VERSION) \
	-D Bin=$(WIN_BIN)/func-e.exe \
	-o $(WIN_BIN)/func-e.msi \
	packaging/msi/func-e.wxs
endif
	@osslsigncode sign -h sha256 -pkcs12 ${WINDOWS_CODESIGN_P12} -pass "${WINDOWS_CODESIGN_PASSWORD}" \
	-n "func-e makes running Envoy® easy" -i https://func-e.io -t http://timestamp.digicert.com \
	-add-msi-dse -in $(WIN_BIN)/func-e.msi -out $(WIN_BIN)/func-e.msi-signed.exe
	@mv $(WIN_BIN)/func-e.msi-signed.exe $(WIN_BIN)/func-e.msi

##@ Test website
.PHONY: site
site:
	@echo "--- site ---"
	@git submodule update
	@cd site && go run $(hugo) server --disableFastRender -D

##@ Unit and End-to-End tests

TEST_PACKAGES ?= . ./internal/...
.PHONY: test
test:
	@echo "--- test ---"
	@go test $(TEST_PACKAGES)

# End-to-end (e2e) tests run against the func-e binary, built on-demand with goreleaser.
#
# Tests run one at a time, in verbose mode, so that failures are easy to diagnose.
# Note: -failfast helps as it stops at the first error. However, it is not a cacheable flag, so runs won't cache.
.PHONY: e2e
e2e: $(BIN)
	@echo "--- e2e ---"
	@E2E_FUNC_E_PATH=$(BIN) go test -parallel 1 -v -failfast ./e2e

##@ Code quality and integrity

COVERAGE_PACKAGES ?= $(shell echo $(TEST_PACKAGES)| tr -s " " ",")
.PHONY: coverage
coverage:
	@echo "--- coverage ---"
	@go test -coverprofile=coverage.txt -covermode=atomic --coverpkg $(COVERAGE_PACKAGES) $(TEST_PACKAGES)
	@go tool cover -func coverage.txt

.PHONY: lint
lint: .golangci.yml .goreleaser.yaml ## Run the linters
	@echo "--- lint ---"
	@go run $(licenser) verify -r .
	@go run $(golangci_lint) run --timeout 5m --config .golangci.yml ./...
	@go run $(goreleaser) check -q

# The goimports tool does not arrange imports in 3 blocks if there are already more than three blocks.
# To avoid that, before running it, we collapse all imports in one block, then run the formatter.
.PHONY: format
format: ## Format all Go code
	@echo "--- format ---"
	@go run $(licenser) apply -r "Tetrate"
	@find . -type f -name '*.go' | xargs gofmt -s -w
	@for f in `find . -name '*.go'`; do \
	    awk '/^import \($$/,/^\)$$/{if($$0=="")next}{print}' $$f > /tmp/fmt; \
	    mv /tmp/fmt $$f; \
	    go run $(goimports) -w -local github.com/tetratelabs/func-e $$f; \
	done

# Enforce go version matches what's in go.mod when running `make check` assuming the following:
# * 'go version' returns output like "go version go1.16 darwin/amd64"
# * go.mod contains a line like "go 1.16"
EXPECTED_GO_VERSION_PREFIX := "go version go$(shell sed -ne '/^go /s/.* //gp' go.mod )"
GO_VERSION := $(shell go version)

.PHONY: check
check:  ## CI blocks merge until this passes. If this fails, run "make check" locally and commit the difference.
# case statement because /bin/sh cannot do prefix comparison, awk is awkward and assuming /bin/bash is brittle
	@case "$(GO_VERSION)" in $(EXPECTED_GO_VERSION_PREFIX)* ) ;; * ) \
		echo "Expected 'go version' to start with $(EXPECTED_GO_VERSION_PREFIX), but it didn't: $(GO_VERSION)"; \
		exit 1; \
	esac
	@$(MAKE) lint
	@$(MAKE) format
# this will taint if we are behind from latest binary. printf avoids adding a newline to the file
    @curl -fsSL https://archive.tetratelabs.io/envoy/envoy-versions.json |jq -er .latestVersion|xargs printf "%s" \
         >./internal/version/last_known_envoy.txt
	@go mod tidy
	@if [ ! -z "`git status -s`" ]; then \
		echo "The following differences will fail CI until committed:"; \
		git diff --exit-code; \
	fi

.PHONY: clean
clean: ## Clean all binaries
	@echo "--- $@ ---"
	@rm -rf dist coverage.txt
	@go clean -testcache
	@go run $(golangci_lint) cache clean
