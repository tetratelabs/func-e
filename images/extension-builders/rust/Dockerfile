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
# ghcr.io allows building without using docker.io as it is rate-limited
# .github/workflows/internal-images.yml publishes this
FROM ghcr.io/tetratelabs/getenvoy-internal:rust

RUN rustup target add wasm32-unknown-unknown

# Unset CARGO_HOME and CARGO_TARGET_DIR. This way we will know when a user overrides them.
ENV CARGO_HOME=
ENV CARGO_TARGET_DIR=

COPY ./entrypoint.sh /usr/local/getenvoy/extension/builder/entrypoint.sh
COPY ./rust/commands.sh /usr/local/getenvoy/extension/builder/commands.sh
ENTRYPOINT ["/usr/local/getenvoy/extension/builder/entrypoint.sh"]
CMD ["--help"]
