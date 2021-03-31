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

# see #153 as some of the below may be incorporated into make or CI setup or otherwise improved
# TODO: why default to HOME where linux defaults to /tmp?
E2E_CACHE_DIR="${E2E_CACHE_DIR:-$HOME/cache/getenvoy}"

# make sure the cache directory is first created on behalf of the current user
mkdir -p "${E2E_CACHE_DIR}"

# TODO: support multiple language
# to speed up `getenvoy extension build|test`, re-use a single cache across all extensions created by e2e tests
export E2E_BUILTIN_TOOLCHAIN_CONTAINER_OPTIONS="${E2E_BUILTIN_TOOLCHAIN_CONTAINER_OPTIONS} -v ${E2E_CACHE_DIR}:/tmp/cache/getenvoy -e CARGO_HOME=/tmp/cache/getenvoy/extension/rust-builder/cargo"

# set HOME directory (TODO: why? why not GETENVOY_HOME? why not also in linux?)
export E2E_BUILTIN_TOOLCHAIN_CONTAINER_OPTIONS="${E2E_BUILTIN_TOOLCHAIN_CONTAINER_OPTIONS} -e HOME=/tmp/getenvoy"

# run the normal make script.
make e2e
