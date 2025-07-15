// Copyright 2025 Tetrate
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

package e2e

import (
	"context"
	"testing"

	"github.com/tetratelabs/func-e/internal/test/e2e"
)

func TestRun(t *testing.T) {
	e2e.TestRun(context.Background(), t, funcEFactory{})
}

func TestRun_MinimalListener(t *testing.T) {
	e2e.TestRun_MinimalListener(context.Background(), t, funcEFactory{})
}

func TestRun_InvalidConfig(t *testing.T) {
	e2e.TestRun_InvalidConfig(context.Background(), t, funcEFactory{})
}

func TestRun_StaticFile(t *testing.T) {
	e2e.TestRun_StaticFile(context.Background(), t, funcEFactory{})
}
