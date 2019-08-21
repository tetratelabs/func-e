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
		location, err := manifest.Locate(key, manifest.DefaultURL)
		if err != nil {
			return fmt.Errorf("unable to retrieve manifest from %v: %v", manifest.DefaultURL, err)
		}
		if err := r.Fetch(key, location); err != nil {
			return fmt.Errorf("unable to retrieve binary from %v: %v", location, err)
		}
	}
	return nil
}

// Run runs, waits for ready, sends sigint, waits for termination, then unarchives the debug directory.
// It is blocking and will only return once completed or context timeout is exceeded
func Run(r binary.Runner, key *manifest.Key, bootstrap string) {
	go r.Run(key, []string{"-c", bootstrap})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	r.WaitWithContext(ctx, binary.StatusReady)
	r.SendSignal(syscall.SIGINT)
	r.WaitWithContext(ctx, binary.StatusTerminated)
	archiver.Unarchive(r.DebugStore()+".tar.gz", filepath.Dir(r.DebugStore()))
}
