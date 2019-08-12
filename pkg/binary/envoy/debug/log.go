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

// +build !linux

package debug

import (
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/log"
)

// EnableEnvoyLogCollection is a preset option that registers collection of Envoy Access Logs
//
// This is not supported on non-Linux platforms as Envoy is not able to access /dev/stdout unless
// *exec.Cmd.Stdout == os.Stdout. This makes it impossible to add the MultiWriter and start Envoy.
//
// The reason for this is not fully understood and not worth investing time yet just for macOS.
var EnableEnvoyLogCollection = func(r *envoy.Runtime) {
	log.Errorf("Log collection is not supported on this Operating System")
}
