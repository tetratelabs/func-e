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

#
# Builder image for Envoy Wasm extensions written in Rust.
#
FROM rust:1.44

RUN rustup target add wasm32-unknown-unknown

# Unset CARGO_HOME. This way we will be able to determine when a user
# provides an override value.
ENV CARGO_HOME=

COPY ./entrypoint.sh /usr/local/getenvoy/extension/builder/entrypoint.sh
COPY ./commands.sh /usr/local/getenvoy/extension/builder/commands.sh
ENTRYPOINT ["/usr/local/getenvoy/extension/builder/entrypoint.sh"]
CMD ["--help"]
