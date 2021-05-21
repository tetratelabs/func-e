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
	"bufio"
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/binary/envoy"
)

// RequireRunTerminate executes Run on the given Runtime and terminates it after starting.
func RequireRunTerminate(t *testing.T, terminate func(r *envoy.Runtime), r *envoy.Runtime, args ...string) (err error) {
	if terminate == nil {
		terminate = func(r *envoy.Runtime) {
			fakeInterrupt := r.FakeInterrupt
			if fakeInterrupt != nil {
				fakeInterrupt()
			}
		}
	}

	// tee the error stream so we can look for the "started" line without consuming it.
	stderr := new(bytes.Buffer)
	r.Err = io.MultiWriter(r.Err, stderr)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		err = r.Run(ctx, args)
		cancel()
	}()

	reader := bufio.NewReader(stderr)
	require.Eventually(t, func() bool {
		b, e := reader.Peek(512)
		return e != nil && strings.Contains(string(b), "started\n")
	}, 2*time.Second, 100*time.Millisecond, "never started process")

	terminate(r)

	select { // Await run completion
	case <-time.After(10 * time.Second):
		t.Fatal("Run never completed")
	case <-ctx.Done():
	}
	return //nolint
}
