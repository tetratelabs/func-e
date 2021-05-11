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

trap 'echo >&2 "SIGTERM caught!"; exit 0' TERM
trap 'echo >&2 "SIGINT caught!"; exit 0' INT

# This script allows us to test what was invoked by exec.Command, and any signal handling.
# This is intentionally written in minimal syntax, which works in ash (Alpine) as well as bash.
set -ue

# Echo invocation context to stdout and fake stderr to ensure it is not combined into stdout.
# We prefix output to help distinguish between multiple copies of this script.
bin=$(basename "$0")
echo "$bin wd: $PWD"
echo "$bin bin: $0"
echo "$bin args: $*"
echo >&2 "$bin stderr"

# If any arg is ${bin}_exit=N, exit with that code. Ex. envoy_exit=3
for arg in "$@"; do
	case $arg in ${bin}_exit=*) exit $(echo "$arg" | cut -d= -f2) ;; esac
done

# Depending on the script name, sleep and echo the interrupting signal.
if [ "$bin" = "envoy" ] || [ "$bin" = "sleep" ]; then
	while true; do sleep 1; done
fi
