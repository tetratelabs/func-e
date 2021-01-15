#!/usr/bin/env bash

# Copyright 2020 Tetrate
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

set -e

SCRIPT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}")"  && pwd)
WORKSPACE_DIR="${SCRIPT_DIR}/../../.."

E2E_CACHE_DIR="${E2E_CACHE_DIR:-$HOME/cache/getenvoy}"

# make sure the cache directory is first created on behalf of the current user
mkdir -p "${E2E_CACHE_DIR}"

# TODO: support multiple language
# to speed up `getenvoy extension build|test`, re-use a single cache across all extensions created by e2e tests
export E2E_BUILTIN_TOOLCHAIN_CONTAINER_OPTIONS="${E2E_BUILTIN_TOOLCHAIN_CONTAINER_OPTIONS} -v ${E2E_CACHE_DIR}:/tmp/cache/getenvoy -e CARGO_HOME=/tmp/cache/getenvoy/extension/rust-builder/cargo"

# set HOME directory
export E2E_BUILTIN_TOOLCHAIN_CONTAINER_OPTIONS="${E2E_BUILTIN_TOOLCHAIN_CONTAINER_OPTIONS} -e HOME=/tmp/getenvoy"

# restore executable bit that get lost by Github Actions during artifact upload/download
chmod a+x ${WORKSPACE_DIR}/build/bin/darwin/amd64/*

# run e2e tests on a `getenvoy` binary built by the upstream job
export E2E_GETENVOY_BINARY="${WORKSPACE_DIR}/build/bin/darwin/amd64/getenvoy"

# run e2e tests with '-ginkgo.v' flag to be able to see the progress
${WORKSPACE_DIR}/build/bin/darwin/amd64/e2e -ginkgo.v
