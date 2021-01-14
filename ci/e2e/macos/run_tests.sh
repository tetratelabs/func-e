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

# to speed up `getenvoy extension build|test`, re-use a single cache across all extensions created by e2e tests
export E2E_BUILTIN_TOOLCHAIN_CONTAINER_OPTIONS="${E2E_BUILTIN_TOOLCHAIN_CONTAINER_OPTIONS} -v ${E2E_CACHE_DIR}:/tmp/cache/getenvoy -e CARGO_HOME=/tmp/cache/getenvoy/extension/rust-builder/cargo"

# set HOME directory
export E2E_BUILTIN_TOOLCHAIN_CONTAINER_OPTIONS="${E2E_BUILTIN_TOOLCHAIN_CONTAINER_OPTIONS} -e HOME=/tmp/getenvoy"

forward_ssh_agent=false
case "${E2E_ALLOW_PRIVATE_DEPENDENCIES}" in
	yes | on | true | 1) forward_ssh_agent=true ;;
esac

if [[ "${forward_ssh_agent}" == "true" ]]; then
	# unfortunately, older versions of 'Docker for Mac' (that can be installed in CI environment)
	# do not support SSH agent forwarding.
	# that is why we have to take care of it manually and work around the limitation that it's not
	# possible to mount a Unix socket from a Mac host into a container

	# setup SSH key that will be used by build containers to fetch private dependencies
	mkdir -p $HOME/.ssh/
	echo "${E2E_GITHUB_MACHINE_USER_KEY}" | base64 -D > $HOME/.ssh/id_rsa_e2e_github_machine_user
	chmod 600 $HOME/.ssh/id_rsa_e2e_github_machine_user

	# create a wrapper script around the original container entrypoint to setup SSH agent inside the container
	echo '#!/usr/bin/env bash
set -e
# use an SSH agent to manage the keys (works better than a plain SSH key in case of Cargo)
eval $(ssh-agent -s)
# always kill that SSH agent in the end
trap "ssh-agent -k" EXIT
# load a single key of a GitHub "machine user" that has access to all private repositories needed by e2e tests
ssh-add $HOME/.ssh/id_rsa
# sanity check
ssh-add -l

# call the original entrypoint
/usr/local/getenvoy/extension/builder/entrypoint.sh "$@"
' > /tmp/entrypoint-wrapper.sh
	chmod a+x /tmp/entrypoint-wrapper.sh

	# mount SSH key into extension build containers
	export E2E_BUILTIN_TOOLCHAIN_CONTAINER_OPTIONS="${E2E_BUILTIN_TOOLCHAIN_CONTAINER_OPTIONS} -v $HOME/.ssh/id_rsa_e2e_github_machine_user:/tmp/getenvoy/.ssh/id_rsa"
	# mount the entrypoint wrapper script
	export E2E_BUILTIN_TOOLCHAIN_CONTAINER_OPTIONS="${E2E_BUILTIN_TOOLCHAIN_CONTAINER_OPTIONS} -v /tmp/entrypoint-wrapper.sh:/tmp/getenvoy/entrypoint-wrapper.sh"
	# substitute entrypoint with a wrapper script
	export E2E_BUILTIN_TOOLCHAIN_CONTAINER_OPTIONS="${E2E_BUILTIN_TOOLCHAIN_CONTAINER_OPTIONS} --entrypoint /tmp/getenvoy/entrypoint-wrapper.sh"
fi

# restore executable bit that get lost by Github Actions during artifact upload/download
chmod a+x ${WORKSPACE_DIR}/build/bin/darwin/amd64/*

# start other containers required in e2e tests
docker-compose up -d

# run e2e tests on a `getenvoy` binary built by the upstream job
export E2E_GETENVOY_BINARY="${WORKSPACE_DIR}/build/bin/darwin/amd64/getenvoy"

# run e2e tests with '-ginkgo.v' flag to be able to see the progress
${WORKSPACE_DIR}/build/bin/darwin/amd64/e2e -ginkgo.v
