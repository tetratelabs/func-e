# Auto generated binary variables helper managed by https://github.com/bwplotka/bingo v0.5.1. DO NOT EDIT.
# All tools are designed to be build inside $GOBIN.
BINGO_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
GOPATH ?= $(shell go env GOPATH)
# We assume GOPATH has a single entry and those who don't must override GOBIN
# While substituting path separators is easy, element separators is tricky due to c:\ in Windows.
GOBIN  ?= $(subst \,/,${GOPATH})/bin
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
GOIMPORTS := $(GOBIN)/goimports-v0.1.5
$(GOIMPORTS): $(BINGO_DIR)/goimports.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/goimports-v0.1.5"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=goimports.mod -o=$(GOBIN)/goimports-v0.1.5 "golang.org/x/tools/cmd/goimports"

GOLANGCI_LINT := $(GOBIN)/golangci-lint-v1.42.0
$(GOLANGCI_LINT): $(BINGO_DIR)/golangci-lint.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/golangci-lint-v1.42.0"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=golangci-lint.mod -o=$(GOBIN)/golangci-lint-v1.42.0 "github.com/golangci/golangci-lint/cmd/golangci-lint"

GORELEASER := $(GOBIN)/goreleaser-v0.174.2
$(GORELEASER): $(BINGO_DIR)/goreleaser.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/goreleaser-v0.174.2"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=goreleaser.mod -o=$(GOBIN)/goreleaser-v0.174.2 "github.com/goreleaser/goreleaser"

HUGO := $(GOBIN)/hugo-v0.86.1
$(HUGO): $(BINGO_DIR)/hugo.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/hugo-v0.86.1"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=hugo.mod -o=$(GOBIN)/hugo-v0.86.1 -tags=extended "github.com/gohugoio/hugo"

LICENSER := $(GOBIN)/licenser-v0.6.0
$(LICENSER): $(BINGO_DIR)/licenser.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/licenser-v0.6.0"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=licenser.mod -o=$(GOBIN)/licenser-v0.6.0 "github.com/liamawhite/licenser"

