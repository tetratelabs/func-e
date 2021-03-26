# Auto generated binary variables helper managed by https://github.com/bwplotka/bingo v0.4.0. DO NOT EDIT.
# All tools are designed to be build inside $GOBIN.
BINGO_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
GOPATH ?= $(shell go env GOPATH)
GOBIN  ?= $(firstword $(subst :, ,${GOPATH}))/bin
GO     ?= $(shell which go)

# Below generated variables ensure that every time a tool under each variable is invoked, the correct version
# will be used; reinstalling only if needed.
# For example for goimports variable:
#
# In your main Makefile (for non array binaries):
#
#include .bingo/Variables.mk # Assuming -dir was set to .bingo .
#
#command: $(GOIMPORTS)
#	@echo "Running goimports"
#	@$(GOIMPORTS) <flags/args..>
#
GOIMPORTS := $(GOBIN)/goimports-v0.1.0
$(GOIMPORTS): $(BINGO_DIR)/goimports.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/goimports-v0.1.0"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=goimports.mod -o=$(GOBIN)/goimports-v0.1.0 "golang.org/x/tools/cmd/goimports"

GOLANGCI_LINT := $(GOBIN)/golangci-lint-v1.38.0
$(GOLANGCI_LINT): $(BINGO_DIR)/golangci-lint.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/golangci-lint-v1.38.0"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=golangci-lint.mod -o=$(GOBIN)/golangci-lint-v1.38.0 "github.com/golangci/golangci-lint/cmd/golangci-lint"

LICENSER := $(GOBIN)/licenser-v0.6.0
$(LICENSER): $(BINGO_DIR)/licenser.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/licenser-v0.6.0"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=licenser.mod -o=$(GOBIN)/licenser-v0.6.0 "github.com/liamawhite/licenser"

SHFMT := $(GOBIN)/shfmt-v3.2.4
$(SHFMT): $(BINGO_DIR)/shfmt.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/shfmt-v3.2.4"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=shfmt.mod -o=$(GOBIN)/shfmt-v3.2.4 "mvdan.cc/sh/v3/cmd/shfmt"

