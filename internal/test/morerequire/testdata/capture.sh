#!/bin/sh

# Copyright 2021 Tetrate
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

# Below are how Envoy 1.17 handle signals
# * `kill -2 pid` or Ctrl-C
trap 'echo >&2 "caught SIGINT"; exit 0' INT
# * `kill pid`
trap 'echo >&2 "caught ENVOY_SIGTERM"; exit 0' TERM
# * `kill -9 pid` cannot be trapped, but Envoy echos "Killed: 9" and exits 137

# Echo a line so we know the traps applied
echo >&2 "started"

# This script allows us to test what was invoked by exec.Command, and any signal handling.
# This is intentionally written in minimal syntax, which works in ash (Alpine) as well as bash.
set -ue

# Echo invocation context to stdout and fake stderr to ensure it is not combined into stdout.
# We prefix output to help distinguish between this script and what's calling it.
bin=$(basename "$0")
if [ "$bin" != "quiet" ]; then
  echo "$bin wd: $PWD"
  echo "$bin bin: $0"
  echo "$bin args: $*"
fi

# If any arg is ${bin}_exit=N, exit with that code. Ex. envoy_exit=3
for arg in "$@"; do
  case $arg in ${bin}_exit=*) exit $(echo "$arg" | cut -d= -f2) ;; esac
done

# sleep and echo the interrupting signal.
while true; do sleep 1; done
