#!/bin/sh -ue

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

case $(uname -m) in
  amd64* | x86_64* )
    deb_arch=amd64
    ;;
  arm64* | aarch64* )
    deb_arch=arm64
    ;;
  * )
    >&2 echo "Unsupported hardware: $(uname -m)"
    exit 1;
esac

deb_file=${1:-$(ls dist/func-e_*_linux_${deb_arch}.deb)}

echo installing "${deb_file}"
sudo dpkg -i "${deb_file}"

echo ensuring func-e was installed
test -f /usr/bin/func-e
func-e -version

echo ensuring func-e man page was installed
test -f /usr/local/share/man/man8/func-e.8

echo uninstalling func-e
sudo apt-get remove -yqq func-e

echo ensuring func-e was uninstalled
test -f /usr/bin/func-e && exit 1
exit 0
