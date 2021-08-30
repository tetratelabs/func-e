# Copyright 2021 Tetrate
# Licensed under the Apache License, Version 2.0 (the "License")
#
# This script uses automatic variables (ex $<) and substitution references $(<:.signed=)
# Please see GNU make's documentation if unfamiliar: https://www.gnu.org/software/make/manual/html_node/
.PHONY: test build e2e dist

# This should be driven by automation and result in N.N.N, not vN.N.N
VERSION   ?= dev
build_dir := build
dist_dir  := dist

# Build the path relating to the current runtime (goos,goarch)
goos   := $(shell go env GOOS)
goarch := $(shell go env GOARCH)
goexe  := $(shell go env GOEXE)
current_binary := $(build_dir)/func-e_$(goos)_$(goarch)/func-e$(goexe)

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
	@printf "$(ansi_format_bright)" test "ok"

# Tests run one at a time, in verbose mode, so that failures are easy to diagnose.
# Note: -failfast helps as it stops at the first error. However, it is not a cacheable flag, so runs won't cache.
e2e: $(current_binary) ## Run all end-to-end tests
	@printf "$(ansi_format_dark)" e2e "running end-to-end tests"
	@E2E_FUNC_E_PATH=$(dir $(current_binary)) go test -parallel 1 -v -failfast ./e2e
	@printf "$(ansi_format_bright)" e2e "ok"

non_windows_platforms := darwin_amd64 darwin_arm64 linux_amd64 linux_arm64
# TODO: arm64 on Windows https://github.com/envoyproxy/envoy/issues/17572
windows_platforms := windows_amd64

$(build_dir)/func-e_%/func-e: main.go
	$(call go-build, $@, $^)

$(dist_dir)/func-e_$(VERSION)_%.tar.gz: $(build_dir)/func-e_%/func-e
	@printf "$(ansi_format_dark)" tar.gz "tarring $@"
	@mkdir -p $(dir $@)
	@tar --strip-components 2 -cpzf $@ $<
	@printf "$(ansi_format_bright)" tar.gz "ok"

$(build_dir)/func-e_%/func-e.exe: main.go
	$(call go-build, $@, $^)

$(dist_dir)/func-e_$(VERSION)_%.zip: $(build_dir)/func-e_%/func-e.exe.signed
	@printf "$(ansi_format_dark)" zip "zipping $@"
	@mkdir -p $(dir $@)
ifeq ($(goos),windows)  # Windows 10 etc use 7zip
	@# '-bso0 -bsp0' makes it quiet, except errors
	@# Wildcards in 7zip will skip the directory. We single-quote to ensure they aren't interpreted by the shell.
	@7z -bso0 -bsp0 a $@ './$(dir $<)/*.exe'
else  # Otherwise, assume zip is available
	@zip -qj $@ $(<:.signed=)
endif
	@printf "$(ansi_format_bright)" zip "ok"

# msi-arch is a macro so we can detect it based on the file naming convention
msi-arch = $(if $(findstring amd64,$1),x64,arm64)
# Default to a dummy version, which is always lower than a real release
msi_version=$(VERSION:dev=0.0.1)

# This builds the Windows installer (MSI) using platform-dependent WIX commands.
$(dist_dir)/func-e_$(VERSION)_%.msi: $(build_dir)/func-e_%/func-e.exe.signed
	@printf "$(ansi_format_dark)" msi "building $@"
	@mkdir -p $(dir $@)
ifeq ($(goos),windows)  # Windows 10 etc use https://wixtoolset.org
	@candle -nologo -arch $(call msi-arch,$@) -dVersion=$(msi_version) -dBin=$(<:.signed=) packaging/msi/func-e.wxs
	@light -nologo func-e.wixobj -o $@ -spdb
	@rm func-e.wixobj
else  # use https://wiki.gnome.org/msitools
	@wixl -a $(call msi-arch,$@) -D Version=$(msi_version) -D Bin=$(<:.signed=) -o $@ packaging/msi/func-e.wxs
endif
	$(call codesign, $@)
	@printf "$(ansi_format_bright)" msi "ok"

# Archives are tar.gz, except in the case of Windows, which uses zip.
archives  := $(non_windows_platforms:%=$(dist_dir)/func-e_$(VERSION)_%.tar.gz) $(windows_platforms:%=$(dist_dir)/func-e_$(VERSION)_%.zip)
packages  := $(windows_platforms:%=$(dist_dir)/func-e_$(VERSION)_%.msi)
checksums := $(dist_dir)/func-e_$(VERSION)_checksums.txt

# Darwin doesn't have sha256sum. See https://github.com/actions/virtual-environments/issues/90
sha256sum := $(if $(findstring darwin,$(goos)),shasum -a 256,sha256sum)
$(checksums): $(archives) $(packages)
	@printf "$(ansi_format_dark)" sha256sum "generating $@"
	@$(sha256sum) $^ > $@
	@printf "$(ansi_format_bright)" sha256sum "ok"

# dist generates the assets that attach to a release
# Ex. https://github.com/tetratelabs/func-e/releases/tag/v$(VERSION)
dist: $(archives) $(packages) $(checksums) ## generates release assets

# this makes a marker file ending in .signed to avoid repeatedly calling codesign
%.signed: %
	$(call codesign, $<)
	@echo > $@

# define macros for multi-platform builds. these parse the filename being built
go-arch = $(if $(findstring amd64,$1),amd64,arm64)
go-os   = $(if $(findstring .exe,$1),windows,$(if $(findstring linux,$1),linux,darwin))
define go-build
	@printf "$(ansi_format_dark)" build "building $1"
	@CGO_ENABLED=0 GOOS=$(call go-os,$1) GOARCH=$(call go-arch,$1) go build \
		-ldflags "-s -w -X main.version=$(VERSION)" \
		-o $1 $2
	@printf "$(ansi_format_bright)" build "ok"
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
WINDOWS_CODESIGN_P12 ?= packaging/msi/func-e.p12
WINDOWS_CODESIGN_PASSWORD ?= func-e-bunch
define codesign
	@printf "$(ansi_format_dark)" codesign "signing $1"
	@osslsigncode sign -h sha256 -pkcs12 ${WINDOWS_CODESIGN_P12} -pass "${WINDOWS_CODESIGN_PASSWORD}" \
	-n "func-e makes running Envoy® easy" -i https://func-e.io -t http://timestamp.digicert.com \
	$(if $(findstring msi,$(1)),-add-msi-dse) -in $1 -out $1-signed
	@mv $1-signed $1
	@printf "$(ansi_format_bright)" codesign "ok"
endef
