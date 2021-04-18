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

package getenvoy

import (
	"github.com/tetratelabs/multierror"

	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/debug"
	types "github.com/tetratelabs/getenvoy/pkg/extension/workspace/example/runtime"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/example/runtime/configdir"
	executil "github.com/tetratelabs/getenvoy/pkg/util/exec"
)

// NewRuntime returns a new runtime backed by GetEnvoy.
func NewRuntime() types.Runtime {
	return &runtime{}
}

type runtime struct{}

func (r *runtime) Run(ctx *types.RunContext) (errs error) {
	// create a temporary directory per run
	configDir, err := configdir.NewConfigDir(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if e := configDir.Close(); e != nil {
			errs = multierror.Append(errs, err)
		}
	}()

	// run the example using `getenvoy run`
	runtime, err := envoy.NewRuntime(envoy.RuntimeOption(
		func(r *envoy.Runtime) {
			r.WorkingDir = configDir.GetDir()
			r.IO = ctx.IO
		}).
		AndAll(debug.EnableAll())...,
	)
	if err != nil {
		return err
	}
	args := executil.Args{"-c", configDir.GetBootstrapFile()}.
		Add(ctx.Opts.Envoy.Args...)
	return runtime.FetchAndRun(ctx.Opts.GetEnvoyReference(), args)
}
