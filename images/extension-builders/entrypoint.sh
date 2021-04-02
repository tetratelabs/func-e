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

SCRIPT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}")"  && pwd)

# Note: we are using option syntax for --output-file even if this was always required. This is for compatibility with
# older versions of the getenvoy binary.
USAGE="usage: build --output-file path/to/extension.wasm
   or: test
   or: clean

examples:
   # build Wasm extension and copy *.wasm file to a given location
   build --output-file target/extension.wasm
"

usage() {
	echo "$USAGE" >&2
	exit 1
}

log_message() {
	echo "$*" >&2
}

error() {
	log_message "error:" "$*"
	exit 1
}

args_error() {
	log_message "error:" "$*"
	log_message
	usage
}

. "${SCRIPT_DIR}/commands.sh"

#######################################################
# Parse command-line arguments and run 'build' command.
#######################################################
command_build()  {
	local output_file=""

	while [[ $# > 0 ]]; do
		case "$1" in
			--output-file)
				if [[ $# < 2 ]]; then
					args_error "--output-file value is missing"
				fi
				output_file="$2"
				shift
				;;
			*)
				usage
				;;
		esac
		shift
	done

	extension_build "${output_file}"
}

#######################################################
# Parse command-line arguments and run 'test' command.
#######################################################
command_test()  {
	extension_test
}

#######################################################
# Parse command-line arguments and run 'clean' command.
#######################################################
command_clean()  {
	extension_clean
}

case "$1" in
	build)
		shift
		command_build "$@"
		;;
	test)
		shift
		command_test "$@"
		;;
	clean)
		shift
		command_clean "$@"
		;;
	*)
		usage
		;;
esac
