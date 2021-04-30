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

// DefaultDockerUser allows tests to read-back what would be used in Docker commands.
var DefaultDockerUser = func() string {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s:%s", u.Uid, u.Gid)
}()

// NewToolchain returns a builtin toolchain with a given configuration.
func NewToolchain(name string, cfg *config.ToolchainConfig, workspace model.Workspace) *builtin { //nolint
	dockerPath := cfg.Container.DockerPath
	if dockerPath == "" {
		dockerPath = "docker"
	}
	return &builtin{name: name, dockerPath: dockerPath, dockerUser: DefaultDockerUser, cfg: cfg, workspace: workspace}
}

// builtin represents a builtin toolchain.
type builtin struct {
	name                   string
	dockerPath, dockerUser string
	cfg                    *config.ToolchainConfig
	workspace              model.Workspace
}

func (t *builtin) GetName() string {
	return t.name
}

func (t *builtin) GetBuildOutputWasmFile() string {
	return filepath.Join(t.workspace.GetDir().GetRootDir(), t.cfg.GetBuildOutputWasmFile())
}

func (t *builtin) Build(context types.BuildContext) error {
	args := t.dockerCliArgs(t.cfg.GetBuildContainer())
	// #nosec -> the current design is an argument builder, not arg literals
	cmd := exec.Command(t.dockerPath, append(args, commandBuild, "--output-file", t.cfg.GetBuildOutputWasmFile())...)
	cmd.Dir = t.workspace.GetDir().GetRootDir() // execute DockerPath in ExtensionDir
	return executil.Run(cmd, context.IO)
}

func (t *builtin) Test(context types.TestContext) error {
	args := t.dockerCliArgs(t.cfg.GetTestContainer())
	// #nosec -> the current design is an argument builder, not arg literals
	cmd := exec.Command(t.dockerPath, append(args, commandTest)...)
	cmd.Dir = t.workspace.GetDir().GetRootDir() // execute DockerPath in ExtensionDir
	return executil.Run(cmd, context.IO)
}

func (t *builtin) Clean(context types.CleanContext) error {
	args := t.dockerCliArgs(t.cfg.GetCleanContainer())
	// #nosec -> the current design is an argument builder, not arg literals
	cmd := exec.Command(t.dockerPath, append(args, commandClean)...)
	cmd.Dir = t.workspace.GetDir().GetRootDir() // execute DockerPath in ExtensionDir
	return executil.Run(cmd, context.IO)
}

func (t *builtin) dockerCliArgs(container *config.ContainerConfig) []string {
	extensionDir := t.workspace.GetDir().GetRootDir()
	volume := fmt.Sprintf("%s:/source", extensionDir) // docker doesn't understand '.'
	return append([]string{
		"run",
		"-u", DefaultDockerUser, // to get proper ownership on files created by the container
		"--rm",
		"-e", "GETENVOY_GOOS=" + runtime.GOOS, // Allows builder images to act based on execution env
		"-t", // to get interactive/colored output out of container
		"-v", volume,
		"-w", "/source",
		"--init", // to ensure container will be responsive to SIGTERM signal
	}, append(container.Options, container.Image)...)
}
