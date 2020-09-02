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

# Docker for Mac 2.0.0.3-ce-mac81,31259 (the last version of 'Docker for Mac' that can be installed in CI environment)
E2E_MACOS_DOCKER_CASK_VERSION="${E2E_MACOS_DOCKER_CASK_VERSION:-8ce4e89d10716666743b28c5a46cd54af59a9cc2}"

# install Docker for Mac
brew cask install https://raw.githubusercontent.com/Homebrew/homebrew-cask/${E2E_MACOS_DOCKER_CASK_VERSION}/Casks/docker.rb

# follow instructions from:
#   https://github.com/microsoft/azure-pipelines-image-generation/issues/738#issuecomment-496211237
#   https://github.com/microsoft/azure-pipelines-image-generation/issues/738#issuecomment-522301481
sudo /Applications/Docker.app/Contents/MacOS/Docker --quit-after-install --unattended
nohup /Applications/Docker.app/Contents/MacOS/Docker --unattended > /dev/stdout &
while ! docker info 2> /dev/null; do
	sleep 5
	echo "Waiting for docker service to be in the running state"
done

# sanity check
docker run --rm -t busybox date
