// Copyright 2020 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"os"

	"github.com/pkg/errors"
)

//nolint:golint
const (
	E2E_GETENVOY_BINARY                     = "E2E_GETENVOY_BINARY"
	E2E_BUILTIN_TOOLCHAIN_CONTAINER_OPTIONS = "E2E_BUILTIN_TOOLCHAIN_CONTAINER_OPTIONS"
)

var (
	// Env represents environment the e2e tests run in.
	Env env
)

type env struct{}

func (env) GetEnvoyBinary() (string, error) {
	value := os.Getenv(E2E_GETENVOY_BINARY)
	if value == "" {
		return "", errors.Errorf("Mandatory environment variable %s is not set.", E2E_GETENVOY_BINARY)
	}
	return value, nil
}

func (env) GetBuiltinContainerOptions() []string {
	value := os.Getenv(E2E_BUILTIN_TOOLCHAIN_CONTAINER_OPTIONS)
	if value == "" {
		return nil
	}
	return []string{"--toolchain-container-options", value}
}
