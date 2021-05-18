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

# bingo manages go binaries needed for building the project
include .bingo/Variables.mk

GETENVOY_TAG ?= dev

BUILD_DIR ?= build
BIN_DIR ?= $(BUILD_DIR)/bin

GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

GO_LD_FLAGS := -ldflags="-s -w -X github.com/tetratelabs/getenvoy/pkg/version.version=$(GETENVOY_TAG)"

TEST_PKG_LIST ?= $(shell go list ./... | grep -v github.com/tetratelabs/getenvoy/test/e2e)
GO_TEST_OPTS ?=
GO_TEST_EXTRA_OPTS ?=

E2E_PKG_LIST ?= ./test/e2e
# Run only one test at a time, in verbose mode, so that failures are easy to diagnose.
# Note: -failfast helps as it stops at the first error. However, it is not a cacheable flag, so runs won't cache.
E2E_OPTS ?= -parallel 1 -v -failfast
E2E_EXTRA_OPTS ?=

GOOSES := linux darwin
GOARCHS := amd64

GETENVOY_OUT_PATH = $(BIN_DIR)/$(1)/$(2)/getenvoy

define GEN_GETENVOY_BUILD_TARGET
.PHONY: $(call GETENVOY_OUT_PATH,$(1),$(2))
$(call GETENVOY_OUT_PATH,$(1),$(2)):
	CGO_ENABLED=0 GOOS=$(1) GOARCH=$(2) go build $(GO_LD_FLAGS) -o $(call GETENVOY_OUT_PATH,$(1),$(2)) ./cmd/getenvoy/main.go
endef
$(foreach os,$(GOOSES),$(foreach arch,$(GOARCHS),$(eval $(call GEN_GETENVOY_BUILD_TARGET,$(os),$(arch)))))

.PHONY: deps
deps:
	go mod download

.PHONY: build
build: $(call GETENVOY_OUT_PATH,$(GOOS),$(GOARCH))

.PHONY: release.dryrun
release.dryrun:
	goreleaser release --skip-publish --snapshot --rm-dist

.PHONY: test
test:
	go test $(GO_TEST_OPTS) $(GO_TEST_EXTRA_OPTS) $(TEST_PKG_LIST)

.PHONY: e2e
e2e: $(call GETENVOY_OUT_PATH,$(GOOS),$(GOARCH))
	go test $(E2E_OPTS) $(E2E_EXTRA_OPTS) $(E2E_PKG_LIST)

.PHONY: bin
bin: $(foreach os,$(GOOSES), bin/$(os))

define GEN_BIN_GOOS_TARGET
.PHONY: bin/$(1)
bin/$(1): $(foreach arch,$(GOARCHS), bin/$(1)/$(arch))
endef
$(foreach os,$(GOOSES),$(eval $(call GEN_BIN_GOOS_TARGET,$(os))))

define GEN_BIN_GOOS_GOARCH_TARGET
.PHONY: bin/$(1)/$(2)
bin/$(1)/$(2): $(call GETENVOY_OUT_PATH,$(1),$(2))
endef
$(foreach os,$(GOOSES),$(foreach arch,$(GOARCHS),$(eval $(call GEN_BIN_GOOS_GOARCH_TARGET,$(os),$(arch)))))

##@ Code quality and integrity

COVERAGE_PACKAGES ?= $(shell echo $(TEST_PACKAGES)| tr -s " " ",")
.PHONY: coverage
coverage:
	@echo "--- coverage ---"
	@go test -coverprofile=coverage.txt -covermode=atomic --coverpkg $(COVERAGE_PACKAGES) $(TEST_PACKAGES)
	@go tool cover -func coverage.txt

LINT_OPTS ?= --timeout 5m
.PHONY: lint
lint: $(GOLANGCI_LINT) $(LICENSER) $(GORELEASER) .golangci.yml .goreleaser.yaml ## Run the linters
	@echo "--- lint ---"
	@$(LICENSER) verify -r .
	@$(GOLANGCI_LINT) run $(LINT_OPTS) --config .golangci.yml ./...
# TODO: this is chatty until https://github.com/goreleaser/goreleaser/issues/2226
	@$(GORELEASER) check

# The goimports tool does not arrange imports in 3 blocks if there are already more than three blocks.
# To avoid that, before running it, we collapse all imports in one block, then run the formatter.
.PHONY: format
format: $(GOIMPORTS) ## Format all Go code
	@echo "--- format ---"
	@$(LICENSER) apply -r "Tetrate"
	@find . -type f -name '*.go' | xargs gofmt -s -w
	@for f in `find . -name '*.go'`; do \
	    awk '/^import \($$/,/^\)$$/{if($$0=="")next}{print}' $$f > /tmp/fmt; \
	    mv /tmp/fmt $$f; \
	    $(GOIMPORTS) -w -local github.com/tetratelabs/getenvoy $$f; \
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
	@go mod tidy
	@if [ ! -z "`git status -s`" ]; then \
		echo "The following differences will fail CI until committed:"; \
		git diff; \
		exit 1; \
	fi

.PHONY: clean
clean: $(GOLANGCI_LINT) ## Clean all binaries
	@echo "--- $@ ---"
	@rm -rf build coverage.txt
	@go clean -testcache
	@$(GOLANGCI_LINT) cache clean
