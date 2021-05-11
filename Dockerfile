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

FROM gcr.io/distroless/cc

ARG getenvoy_binary=getenvoy
COPY ${getenvoy_binary} /

# TODO: delete this when we build in default version
ARG reference
ENV ENVOY_REFERENCE=$reference
RUN ["/getenvoy", "fetch", "${ENVOY_REFERENCE}"]
ENTRYPOINT ["/getenvoy", "run", "${ENVOY_REFERENCE}"]

