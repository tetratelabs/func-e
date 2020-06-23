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

// overridable by unit tests
var (
	GetCurrentUser = user.Current
)

// NewToolchain returns a builtin toolchain with a given configuration.
func NewToolchain(name string, cfg *config.ToolchainConfig, workspace model.Workspace) *builtin {
	return &builtin{name: name, cfg: cfg, workspace: workspace}
}

// builtin represents a builtin toolchain.
type builtin struct {
	name      string
	cfg       *config.ToolchainConfig
	workspace model.Workspace
}

func (t *builtin) Build(context types.BuildContext) error {
	args, err := t.dockerCliArgs(t.cfg.GetBuildContainer())
	if err != nil {
		return err
	}
	cmd := exec.Command("docker", args.Add(commandBuild).Add("--output-file", t.cfg.GetBuildOutputWasmFile())...)
	return executil.Run(cmd, context.IO)
}

func (t *builtin) Test(context types.TestContext) error {
	args, err := t.dockerCliArgs(t.cfg.GetTestContainer())
	if err != nil {
		return err
	}
	cmd := exec.Command("docker", args.Add(commandTest)...)
	return executil.Run(cmd, context.IO)
}

func (t *builtin) Clean(context types.CleanContext) error {
	args, err := t.dockerCliArgs(t.cfg.GetCleanContainer())
	if err != nil {
		return err
	}
	cmd := exec.Command("docker", args.Add(commandClean)...)
	return executil.Run(cmd, context.IO)
}

func (t *builtin) dockerCliArgs(container *config.ContainerConfig) (argList, error) {
	user, err := GetCurrentUser()
	if err != nil {
		return nil, err
	}
	return argList{
		"run",
		"-u", fmt.Sprintf("%s:%s", user.Uid, user.Gid), // to get proper ownership on files created by the container
		"--rm",
		"-t", // to get interactive/colored output out of container
		"-v", fmt.Sprintf("%s:%s", t.workspace.GetDir().GetRootDir(), "/source"),
		"-w", "/source",
		"--init", // to ensure container will be responsive to SIGTERM signal
	}.Add(container.Options...).Add(container.Image), nil
}

type argList []string

func (l argList) Add(values ...string) argList {
	return append(l, values...)
}
