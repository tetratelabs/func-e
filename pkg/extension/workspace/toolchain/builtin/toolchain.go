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

package builtin

import (
	"fmt"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"

	config "github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/toolchain/builtin"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain/types"
	executil "github.com/tetratelabs/getenvoy/pkg/util/exec"
)

// commands supported by the out-of-the-box Docker build container.
const (
	commandBuild = "build"
	commandTest  = "test"
	commandClean = "clean"
)

// GetCurrentUser is overridable for unit tests
var GetCurrentUser = user.Current

// NewToolchain returns a builtin toolchain with a given configuration.
func NewToolchain(name string, cfg *config.ToolchainConfig, workspace model.Workspace) *builtin { //nolint
	return &builtin{name: name, cfg: cfg, workspace: workspace}
}

// builtin represents a builtin toolchain.
type builtin struct {
	name      string
	cfg       *config.ToolchainConfig
	workspace model.Workspace
}

func (t *builtin) GetName() string {
	return t.name
}

func (t *builtin) GetBuildOutputWasmFile() string {
	return filepath.Join(t.workspace.GetDir().GetRootDir(), t.cfg.GetBuildOutputWasmFile())
}

func (t *builtin) Build(context types.BuildContext) error {
	args, err := t.dockerCliArgs(t.cfg.GetBuildContainer())
	if err != nil {
		return err
	}
	// #nosec -> the current design is an argument builder, not arg literals
	cmd := exec.Command("docker", args.Add(commandBuild).Add("--output-file", t.cfg.GetBuildOutputWasmFile())...)
	return executil.Run(cmd, context.IO)
}

func (t *builtin) Test(context types.TestContext) error {
	args, err := t.dockerCliArgs(t.cfg.GetTestContainer())
	if err != nil {
		return err
	}
	// #nosec -> the current design is an argument builder, not arg literals
	cmd := exec.Command("docker", args.Add(commandTest)...)
	return executil.Run(cmd, context.IO)
}

func (t *builtin) Clean(context types.CleanContext) error {
	args, err := t.dockerCliArgs(t.cfg.GetCleanContainer())
	if err != nil {
		return err
	}
	// #nosec -> the current design is an argument builder, not arg literals
	cmd := exec.Command("docker", args.Add(commandClean)...)
	return executil.Run(cmd, context.IO)
}

func (t *builtin) dockerCliArgs(container *config.ContainerConfig) (executil.Args, error) {
	u, err := GetCurrentUser()
	if err != nil {
		return nil, err
	}
	return executil.Args{
		"run",
		"-u", fmt.Sprintf("%s:%s", u.Uid, u.Gid), // to get proper ownership on files created by the container
		"--rm",
		"-e", "GETENVOY_GOOS=" + runtime.GOOS, // Allows builder images to act based on execution env
		"-t", // to get interactive/colored output out of container
		"-v", fmt.Sprintf("%s:%s", t.workspace.GetDir().GetRootDir(), "/source"),
		"-w", "/source",
		"--init", // to ensure container will be responsive to SIGTERM signal
	}.Add(container.Options...).Add(container.Image), nil
}
