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

extension_build()  {
	set_trap
	tinygo build -o "$1" -scheduler=none -target wasi main.go
}

extension_test()  {
	set_trap
	go test -tags=proxytest -v ./...
}

extension_clean()  {
	rm -rf build
}

set_trap() {
	# This is necessary since the created go caches are with read-only permission,
	# and without this, the host user cannot delete the build directory with "rm -rf".
	trap "chmod -R u+rw ${GOMODCACHE}" EXIT ERR INT TERM
}
