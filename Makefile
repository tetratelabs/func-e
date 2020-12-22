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

ENVOY = standard:1.11.1
HUB ?= docker.io/getenvoy
GETENVOY_TAG ?= dev
BUILDERS_LANGS := rust tinygo
BUILDERS_TAG ?= latest
EXTRA_TAG ?=

USE_DOCKER_BUILDKIT_CACHE ?= yes
ifneq ($(filter-out yes on true 1,$(USE_DOCKER_BUILDKIT_CACHE)),)
  USE_DOCKER_BUILDKIT_CACHE=
endif

BUILD_DIR ?= build
BIN_DIR ?= $(BUILD_DIR)/bin
COVERAGE_DIR ?= $(BUILD_DIR)/coverage
COVERAGE_PROFILE := $(COVERAGE_DIR)/coverage.out
COVERAGE_REPORT := $(COVERAGE_DIR)/coverage.html

GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

GO_LD_FLAGS := -ldflags="-s -w -X github.com/tetratelabs/getenvoy/pkg/version.version=$(GETENVOY_TAG)"

TEST_PKG_LIST ?= ./pkg/...
GO_TEST_OPTS ?=
GO_TEST_EXTRA_OPTS ?=

# TODO(yskopets): include all packages into test run once blocking issues have been resolved, including
# * https://github.com/tetratelabs/getenvoy/issues/87 `go test -race` fails
# * https://github.com/tetratelabs/getenvoy/issues/88 `go test ./...` fails on Mac
# * https://github.com/tetratelabs/getenvoy/issues/89 `go test github.com/tetratelabs/getenvoy/pkg/binary/envoy/controlplane` removes `/tmp` dir
COVERAGE_PKG_LIST ?= $(shell go list ./pkg/... | grep -v -e github.com/tetratelabs/getenvoy/pkg/binary/envoy/controlplane -e github.com/tetratelabs/getenvoy/pkg/binary/envoy/debug)
GO_COVERAGE_OPTS ?= -covermode=atomic -coverpkg=./...
GO_COVERAGE_EXTRA_OPTS ?=

E2E_OPTS ?= -ginkgo.v
E2E_EXTRA_OPTS ?=

GOOSES := linux darwin
GOARCHS := amd64

GETENVOY_OUT_PATH = $(BIN_DIR)/$(1)/$(2)/getenvoy

define GEN_GETENVOY_BUILD_TARGET
.PHONY: $(call GETENVOY_OUT_PATH,$(1),$(2))
$(call GETENVOY_OUT_PATH,$(1),$(2)): generate
	CGO_ENABLED=0 GOOS=$(1) GOARCH=$(2) go build $(GO_LD_FLAGS) -o $(call GETENVOY_OUT_PATH,$(1),$(2)) ./cmd/getenvoy/main.go
endef
$(foreach os,$(GOOSES),$(foreach arch,$(GOARCHS),$(eval $(call GEN_GETENVOY_BUILD_TARGET,$(os),$(arch)))))

E2E_OUT_PATH = $(BIN_DIR)/$(1)/$(2)/e2e

define GEN_E2E_BUILD_TARGET
.PHONY: $(call E2E_OUT_PATH,$(1),$(2))
$(call E2E_OUT_PATH,$(1),$(2)):
	CGO_ENABLED=0 GOOS=$(1) GOARCH=$(2) go test -c -o $(call E2E_OUT_PATH,$(1),$(2)) ./test/e2e
endef
$(foreach os,$(GOOSES),$(foreach arch,$(GOARCHS),$(eval $(call GEN_E2E_BUILD_TARGET,$(os),$(arch)))))

.PHONY: init
init: generate

.PHONY: deps
deps:
	go mod download

.PHONY: generate
generate: deps
	go generate ./pkg/...

.PHONY: build
build: $(call GETENVOY_OUT_PATH,$(GOOS),$(GOARCH))

.PHONY: docker
docker: $(call GETENVOY_OUT_PATH,linux,amd64)
	docker build -t $(HUB)/getenvoy:$(GETENVOY_TAG) --build-arg reference=$(ENVOY) --build-arg getenvoy_binary=$(call GETENVOY_OUT_PATH,linux,amd64) .

.PHONY: release.dryrun
release.dryrun:
	goreleaser release --skip-publish --snapshot --rm-dist

.PHONY: test
test: generate
	go test $(GO_TEST_OPTS) $(GO_TEST_EXTRA_OPTS) $(TEST_PKG_LIST)

.PHONY: e2e
e2e: $(call GETENVOY_OUT_PATH,$(GOOS),$(GOARCH)) $(call E2E_OUT_PATH,$(GOOS),$(GOARCH))
	E2E_GETENVOY_BINARY=$(PWD)/$(call GETENVOY_OUT_PATH,$(GOOS),$(GOARCH)) $(call E2E_OUT_PATH,$(GOOS),$(GOARCH)) $(GO_TEST_OPTS) $(GO_TEST_EXTRA_OPTS) $(E2E_OPTS) $(E2E_EXTRA_OPTS)

.PHONY: bin
bin: $(foreach os,$(GOOSES), bin/$(os))

define GEN_BIN_GOOS_TARGET
.PHONY: bin/$(1)
bin/$(1): $(foreach arch,$(GOARCHS), bin/$(1)/$(arch))
endef
$(foreach os,$(GOOSES),$(eval $(call GEN_BIN_GOOS_TARGET,$(os))))

define GEN_BIN_GOOS_GOARCH_TARGET
.PHONY: bin/$(1)/$(2)
bin/$(1)/$(2): $(call GETENVOY_OUT_PATH,$(1),$(2)) $(call E2E_OUT_PATH,$(1),$(2))
endef
$(foreach os,$(GOOSES),$(foreach arch,$(GOARCHS),$(eval $(call GEN_BIN_GOOS_GOARCH_TARGET,$(os),$(arch)))))

.PHONY: coverage
coverage: generate
	mkdir -p "$(shell dirname "$(COVERAGE_PROFILE)")"
	go test $(GO_COVERAGE_OPTS) $(GO_COVERAGE_EXTRA_OPTS) -coverprofile="$(COVERAGE_PROFILE)" $(COVERAGE_PKG_LIST)
	go tool cover -html="$(COVERAGE_PROFILE)" -o "$(COVERAGE_REPORT)"

EXTENSION_BUILDER_IMAGE = getenvoy/extension-$(1)-builder:$(2)
EXTENSION_BUILDER_IMAGE_LATEST_VERSION = $(shell git log --pretty=format:'%H' -n 1 images/extension-builders)

define GEN_BUILD_EXTENSION_BUILDER_IMAGE_TARGET
.PHONY: builder/$(1)
builder/$(1):
	$(if $(USE_DOCKER_BUILDKIT_CACHE),DOCKER_BUILDKIT=1,)                                                                           \
	docker build                                                                                                                    \
	$(if $(USE_DOCKER_BUILDKIT_CACHE),--build-arg BUILDKIT_INLINE_CACHE=1,)                                                         \
	$(if $(USE_DOCKER_BUILDKIT_CACHE),--cache-from $(call EXTENSION_BUILDER_IMAGE,$(1),$(EXTENSION_BUILDER_IMAGE_LATEST_VERSION)),) \
	-t $(call EXTENSION_BUILDER_IMAGE,$(1),$(BUILDERS_TAG))                                                                         \
	-f images/extension-builders/$(1)/Dockerfile images/extension-builders
endef
$(foreach lang,$(BUILDERS_LANGS),$(eval $(call GEN_BUILD_EXTENSION_BUILDER_IMAGE_TARGET,$(lang))))

.PHONY: builders
builders: $(foreach lang,$(BUILDERS_LANGS), builder/$(lang))

define GEN_PUSH_EXTENSION_BUILDER_IMAGE_TARGET
.PHONY: push/builder/$(1)
push/builder/$(1):
	docker push $(call EXTENSION_BUILDER_IMAGE,$(1),$(BUILDERS_TAG))
endef
$(foreach lang,$(BUILDERS_LANGS),$(eval $(call GEN_PUSH_EXTENSION_BUILDER_IMAGE_TARGET,$(lang))))

.PHONY: builders.push
builders.push: $(foreach lang,$(BUILDERS_LANGS), push/builder/$(lang))

define GEN_TAG_EXTENSION_BUILDER_IMAGE_TARGET
.PHONY: tag/builder/$(1)
tag/builder/$(1):
	docker tag $(call EXTENSION_BUILDER_IMAGE,$(1),$(BUILDERS_TAG)) $(call EXTENSION_BUILDER_IMAGE,$(1),$(EXTRA_TAG))
endef
$(foreach lang,$(BUILDERS_LANGS),$(eval $(call GEN_TAG_EXTENSION_BUILDER_IMAGE_TARGET,$(lang))))

.PHONY: builders.tag
builders.tag: $(foreach lang,$(BUILDERS_LANGS), tag/builder/$(lang))

define GEN_PULL_EXTENSION_BUILDER_IMAGE_TARGET
.PHONY: pull/builder/$(1)
pull/builder/$(1):
	docker pull $(call EXTENSION_BUILDER_IMAGE,$(1),$(BUILDERS_TAG))
endef
$(foreach lang,$(BUILDERS_LANGS),$(eval $(call GEN_PULL_EXTENSION_BUILDER_IMAGE_TARGET,$(lang))))

.PHONY: builders.pull
builders.pull: $(foreach lang,$(BUILDERS_LANGS), pull/builder/$(lang))
