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

# Ensure location of the build directory.
export CARGO_TARGET_DIR="${CARGO_TARGET_DIR:-${GETENVOY_WORKSPACE_DIR}/target}"

# Keep Cargo cache inside the build directory, unless a user explicitly
# overrides CARGO_HOME.
export CARGO_HOME="${CARGO_HOME:-${CARGO_TARGET_DIR}/.cache/getenvoy/extension/rust-builder/cargo}"

#########################################################################
# Build Wasm extension and copy *.wasm file to a given location.
# Globals:
#   CARGO_TARGET_DIR
#   GETENVOY_WORKSPACE_DIR
# Arguments:
#   Path relative to the workspace root to copy *.wasm file to.
#########################################################################
extension_build()  {
	local target="wasm32-unknown-unknown"

	cargo build --target "${target}"

	local profile="debug"
	local lib_name="extension"
	local file_name="${lib_name}.wasm"
	local cargo_output_file="${CARGO_TARGET_DIR}/${target}/${profile}/${file_name}"

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

	local destination_file="$1"
	if [[ -n "${destination_file}" ]]; then
		log_message "     Copying *.wasm file to '${destination_file}'"

		destination_file="${GETENVOY_WORKSPACE_DIR}/${destination_file}"
		local tmp_file="${destination_file}.tmp"
		mkdir -p "$(dirname "${tmp_file}")"
		cp "${cargo_output_file}" "${tmp_file}"
		mv "${tmp_file}" "${destination_file}"
	fi
}

extension_test()  {
	cargo test
}

extension_clean()  {
	cargo clean
}
