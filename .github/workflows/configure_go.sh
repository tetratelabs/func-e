#!/usr/bin/env bash
#
# This configures Go according to go.mod, choosing the a GOROOT based on existing variables.
# For example, if go.mod includes "go 1.17", GOROOT=${GOROOT_1_17} or ${GOROOT_1_17_X64}
#
# Variables such as GOROOT_1_17_X64 are defined by GitHub Actions runners. So, this will not
# result in downloading or installing anything not already there.
#
# Notes:
# * In GitHub, these evaluate to ${RUNNER_TOOL_CACHE}/go/${GO_RELEASE}*/x64
#   * RUNNER_TOOL_CACHE lags Go releases by 1-2 weeks https://github.com/actions/virtual-environments
# * To simulate GitHub for testing, set GITHUB_ENV and GITHUB_PATH to temporary files
#   * Ex. `GITHUB_ENV=/tmp/test-env GITHUB_PATH=/tmp/test-path .github/workflows/configure_go.sh
# * This uses bash because we need indirect variable expansion and GHA runners all have bash.
set -uex pipefail

go_release=$(sed -n 's/^go //gp' go.mod)
echo GO_RELEASE="${go_release}" >> "${GITHUB_ENV}"

# Match last exported GOROOT variable name that includes the version we want. Ex. GOROOT_1_17_X64
goroot_name=$(env|grep "^GOROOT_${go_release//./_}"| sed 's/=.*//g'|sort -n|tail -1)

# Remove this if/else after actions/virtual-environments#4156 is solved
if [ -n "${goroot_name}" ]; then
  go_root=${!goroot_name}
else
  # This works around missing variables on macOS via naming convention.
  # Ex. /Users/runner/hostedtoolcache/go/1.17.1/x64
  go_root=$(ls -d "${RUNNER_TOOL_CACHE}"/go/"${go_release}"*/x64|sort -n|tail -1)
fi

# Ensure go works
go="${go_root}/bin/go"
${go} version >/dev/null

# Setup the GOROOT
echo GOROOT="${go_root}" >> "${GITHUB_ENV}"
echo "${go_root}/bin" >>"${GITHUB_PATH}"
