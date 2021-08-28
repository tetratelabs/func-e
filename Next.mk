# Copyright 2021 Tetrate
# Licensed under the Apache License, Version 2.0 (the "License")
.PHONY: test build e2e

# This should be driven by automation and result in N.N.N, not vN.N.N
VERSION ?= dev

# Temporarily match convention of where goreleaser put binaries
build_dir         := dist
# Build the path relating to the current runtime (goos,goarch)
current_binary    := $(build_dir)/func-e_$(shell go env GOOS)_$(shell go env GOARCH)/func-e$(shell go env GOEXE)
# Right now, the only arch we support is amd64 because Envoy doesn't yet support arm64 on Windows
# https://github.com/envoyproxy/envoy/issues/17572
# Once that occurs, we will need to set -arch arm64 and bundle accordingly.
windows_binary    := $(build_dir)/func-e_windows_amd64/func-e.exe

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
	@go test . ./internal/...
	@printf "$(ansi_format_bright)" test "✔"

# Tests run one at a time, in verbose mode, so that failures are easy to diagnose.
# Note: -failfast helps as it stops at the first error. However, it is not a cacheable flag, so runs won't cache.
e2e: $(current_binary) ## Run all end-to-end tests
	@printf "$(ansi_format_dark)" e2e "running end-to-end tests"
	@E2E_FUNC_E_PATH=$(dir $(current_binary)) go test -parallel 1 -v -failfast ./e2e
	@printf "$(ansi_format_bright)" e2e "✔"

# Use patterns to set appropriate variables needed for multi-platform builds
$(build_dir)/func-e_darwin_%/func-e: override goos=darwin
$(build_dir)/func-e_linux_%/func-e: override goos=linux
$(build_dir)/func-e_%_amd64/func-e: override goarch=amd64
$(build_dir)/func-e_%_arm64/func-e: override goarch=arm64
$(build_dir)/func-e_%/func-e:
	$(call go-build, $@)

$(windows_binary): override goos=windows
$(windows_binary): override goarch=amd64
$(windows_binary):
	$(call go-build, $@)

define go-build
	@printf "$(ansi_format_dark)" build "building $1"
	@CGO_ENABLED=0 GOOS=$(goos) GOARCH=$(goarch) go build \
		-ldflags "-s -w -X main.version=$(VERSION)" \
		-o $1 main.go
	@printf "$(ansi_format_bright)" build "✔"
endef
