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

ENVOY = standard:1.11.1
HUB ?= docker.io/getenvoy
TAG ?= dev

deps:
	go mod download

codegen:
	go generate ./pkg/...

build: deps codegen
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o getenvoy ./cmd/getenvoy/main.go

docker: build
	docker build -t $(HUB)/getenvoy:$(TAG) --build-arg reference=$(ENVOY) .

release.dryrun:
	goreleaser release --skip-publish --snapshot --rm-dist
