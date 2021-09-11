// Copyright 2019 Tetrate
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

// This file needs to be here since we need to consume packages some of func-e project packages.
// When building this program, we need to run the "go build" from within the func-e root project
// directory (the place where the func-e go.mod resides).

package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/version"
)

func main() {
	envoy := filepath.Join(os.Getenv("FUNC_E_HOME"), "versions", string(version.LastKnownEnvoy), "envoy")
	cmd := exec.CommandContext(context.Background(), envoy+moreos.Exe, "-c")
	cmd.SysProcAttr = moreos.ProcessGroupAttr()
	cmd.Stderr = os.Stderr // Forward the stderr output.
	cmd.Start()
	cmd.Wait()
}
