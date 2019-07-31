#!/bin/bash

# Copyright 2019 Tetrate
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

# This script is used to simulate a process that never terminates unless it receives SIGINT, SIGTERM or SIGKILL
terminate () {
    echo "received $(($? - 128))"
    exit 0
}

trap terminate SIGINT SIGTERM SIGKILL

while true; do
    sleep 0.1
done
