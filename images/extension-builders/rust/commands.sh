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
set -ue

export CARGO_TARGET="wasm32-unknown-unknown"

# Ensure location of the build cache directory.
# See https://doc.rust-lang.org/cargo/guide/build-cache.html
export CARGO_TARGET_DIR=${CARGO_TARGET_DIR:-${PWD}/target}

# Keep Cargo cache inside the build directory, unless a user explicitly overrides CARGO_HOME.
# See https://doc.rust-lang.org/cargo/guide/cargo-home.html
export CARGO_HOME=${CARGO_HOME:-${CARGO_TARGET_DIR}/.cache/getenvoy/extension/rust-builder/cargo}

# This is used in macOS, where we build in temporary directories. We only cache minimum as copying is slow.
# See https://doc.rust-lang.org/cargo/guide/cargo-home.html#caching-the-cargo-home-in-c
copy_cargo_home_cache() {
	source_dir=${1}
	dest_dir=${2}

	# See if there is anything to copy
	dirs=""
	for dir in bin registry/index registry/cache git/db; do
		test -d "${source_dir}/$dir" && dirs="$dirs $dir"
	done

	# If any directory existed, copy them to the destination
	if [ -n "$dirs" ]; then
		mkdir -p "${dest_dir}" 2>&- || true
		log_message "     Copying cacheable dirs $dirs from ${source_dir} to ${dest_dir}"
		(
			cd "${source_dir}"
			tar -cpf - $dirs | (
				cd "${dest_dir}"
				tar -xpf -
			)
		)
	fi
}

#########################################################################
# Build Wasm extension and copy *.wasm file to a given location.
# Globals:
#   CARGO_HOME
#   CARGO_TARGET_DIR
#   GETENVOY_GOOS
# Arguments:
#   Path relative to the workspace root to copy *.wasm file to.
#########################################################################
extension_build() {
	# Avoid slow IO problems when the Docker host is macOS and bind-mounted volumes. We do this by building into a
	# temp directory, copying back cacheable contents later.
	if [ "${GETENVOY_GOOS:-}" == "darwin" ]; then
		REAL_CARGO_HOME=${CARGO_HOME}
		export CARGO_HOME=/tmp/$$-cargo
		copy_cargo_home_cache "${REAL_CARGO_HOME}" "${CARGO_HOME}"

		export CARGO_TARGET_DIR=/tmp/$$-build
		# We don't copy revert the updated target dir back to the original location because the copying is slower than
		# recompiling each time.
	fi

	mkdir -p "${CARGO_HOME}" 2>&- || true
	mkdir -p "${CARGO_TARGET_DIR}" 2>&- || true

	cargo build --target "${CARGO_TARGET}"

	local profile="debug"
	local lib_name="extension"
	local file_name="${lib_name}.wasm"
	local cargo_output_file="${CARGO_TARGET_DIR}/${CARGO_TARGET}/${profile}/${file_name}"

	if [[ ! -f "${cargo_output_file}" ]]; then
		error "Cargo didn't build a *.wasm file at expected location: ${cargo_output_file}.

help:  make sure Cargo workspace includes a library crate with name '${lib_name}' and type 'cdylib', e.g.

       wasm/module/Cargo.toml:
       ...
       [lib]
       name = \"${lib_name}\"
       crate-type = [\"cdylib\"]
       ...
"
	fi

	local destination_file="${PWD}/$1"
	log_message "     Copying *.wasm file to '${destination_file}'"
	mkdir -p "$(dirname "${destination_file}")"
	cp "${cargo_output_file}" "${destination_file}"

	if [ "${GETENVOY_GOOS:-}" == "darwin" ]; then
		copy_cargo_home_cache "${CARGO_HOME}" "${REAL_CARGO_HOME}"
		rm -rf /tmp/$$*
	fi
}

extension_test() {
	cargo test
}

extension_clean() {
	cargo clean
}
