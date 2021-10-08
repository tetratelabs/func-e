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

DEB_FILE=${DEB_FILE-"dist/func-e_dev_linux_amd64.deb"}

echo installing "${DEB_FILE}"
dpkg -i "${DEB_FILE}"

echo ensuring func-e was installed
func-e -version

echo uninstalling func-e
apt-get remove -yqq func-e

echo ensuring func-e was uninstalled
func-e -version && exit 1

exit 0
