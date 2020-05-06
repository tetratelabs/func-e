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
TAG ?= dev

BUILD_DIR ?= build
COVERAGE_DIR ?= $(BUILD_DIR)/coverage
COVERAGE_PROFILE := $(COVERAGE_DIR)/coverage.out
COVERAGE_REPORT := $(COVERAGE_DIR)/coverage.html

TEST_PKG_LIST ?= ./...
GO_TEST_OPTS ?=
GO_TEST_EXTRA_OPTS ?=

# TODO(yskopets): include all packages into test run once blocking issues have been resolved, including
# * https://github.com/tetratelabs/getenvoy/issues/87 `go test -race` fails
# * https://github.com/tetratelabs/getenvoy/issues/88 `go test ./...` fails on Mac
# * https://github.com/tetratelabs/getenvoy/issues/89 `go test github.com/tetratelabs/getenvoy/pkg/binary/envoy/controlplane` removes `/tmp` dir
COVERAGE_PKG_LIST ?= $(shell go list ./... | grep -v -e github.com/tetratelabs/getenvoy/pkg/binary/envoy/controlplane -e github.com/tetratelabs/getenvoy/pkg/binary/envoy/debug)
GO_COVERAGE_OPTS ?= -covermode=atomic -coverpkg=./...
GO_COVERAGE_EXTRA_OPTS ?=

.PHONY: init
init: generate

.PHONY: deps
deps:
	go mod download

.PHONY: generate
generate: deps
	go generate ./pkg/...

.PHONY: build
build: generate
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o getenvoy ./cmd/getenvoy/main.go

.PHONY: docker
docker: build
	docker build -t $(HUB)/getenvoy:$(TAG) --build-arg reference=$(ENVOY) .

.PHONY: release.dryrun
release.dryrun:
	goreleaser release --skip-publish --snapshot --rm-dist

.PHONY: test
test:
	go test $(GO_TEST_OPTS) $(GO_TEST_EXTRA_OPTS) $(TEST_PKG_LIST)

.PHONY: coverage
coverage:
	mkdir -p "$(shell dirname "$(COVERAGE_PROFILE)")"
	go test $(GO_COVERAGE_OPTS) $(GO_COVERAGE_EXTRA_OPTS) -coverprofile="$(COVERAGE_PROFILE)" $(COVERAGE_PKG_LIST)
	go tool cover -html="$(COVERAGE_PROFILE)" -o "$(COVERAGE_REPORT)"
