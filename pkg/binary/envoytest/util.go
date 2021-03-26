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

package envoytest

import (
	"context"
	"fmt"
	"path/filepath"
	"syscall"
	"time"

	"github.com/mholt/archiver"
	"github.com/tetratelabs/log"

	"github.com/tetratelabs/getenvoy/pkg/binary"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
)

// Reference indicates the default Envoy version to be used for testing
var Reference = "standard:1.11.0"

// Fetch retrieves the Envoy indicated by Reference
func Fetch() error {
	key, _ := manifest.NewKey(Reference)
	r, _ := envoy.NewRuntime()
	if !r.AlreadyDownloaded(key) {
		location, err := manifest.Locate(key)
		if err != nil {
			return fmt.Errorf("unable to retrieve manifest from %v: %v", manifest.GetURL(), err)
		}
		if err := r.Fetch(key, location); err != nil {
			return fmt.Errorf("unable to retrieve binary from %v: %v", location, err)
		}
	}
	return nil
}

// Run executes envoy and waits for it to be ready
// It is blocking and will only return once ready (nil) or context timeout is exceeded (error)
func Run(ctx context.Context, r binary.Runner, bootstrap string) error {
	key, _ := manifest.NewKey(Reference)
	args := []string{}
	if bootstrap != "" {
		args = append(args, "-c", bootstrap)
	}
	go func() {
		if err := r.Run(key, args); err != nil {
			log.Errorf("unable to run key %s: %v", key, err)
		}
	}()
	r.WaitWithContext(ctx, binary.StatusReady)
	return ctx.Err()
}

// Kill sends sigint to a running enboy, waits for termination, then unarchives the debug directory.
// It is blocking and will only return once terminated (nil) or context timeout is exceeded (error)
func Kill(ctx context.Context, r binary.Runner) error {
	r.SendSignal(syscall.SIGINT)
	r.WaitWithContext(ctx, binary.StatusTerminated)
	if err := archiver.Unarchive(r.DebugStore()+".tar.gz", filepath.Dir(r.DebugStore())); err != nil {
		return fmt.Errorf("error killing context: %w", err)
	}
	return ctx.Err()
}

// RunKill executes envoy, waits for ready, sends sigint, waits for termination, then unarchives the debug directory.
// It should be used when you just want to cycle through an Envoy lifecycle
// It is blocking and will only return once completed (nil) or context timeout is exceeded (error)
// If timeout passed is 0, it defaults to 3 seconds
func RunKill(r binary.Runner, bootstrap string, timeout time.Duration) error {
	if timeout == 0 {
		timeout = time.Second * 3
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := Run(ctx, r, bootstrap); err != nil {
		return err
	}
	return Kill(ctx, r)
}
