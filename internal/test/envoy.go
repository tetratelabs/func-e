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

package test

import (
	"bufio"
	"context"
	_ "embed" // Embedding the fakeEnvoySrc is easier than file I/O and ensures it doesn't skew coverage
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Runner allows us to not introduce dependency cycles on envoy.Runtime
type Runner interface {
	Run(ctx context.Context, args []string) error
}

// RequireRun executes Run on the given Runtime and calls shutdown after it started.
func RequireRun(t *testing.T, shutdown func(), r Runner, stderr io.Reader, args ...string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// If there's no shutdown function, shutdown via cancellation. This is similar to ctrl-c
	if shutdown == nil {
		shutdown = cancel
	}

	// Run in a goroutine, and signal when that completes
	ran := make(chan bool)
	var err error
	go func() {
		if e := r.Run(ctx, args); e != nil && err == nil {
			err = e // first error
		}
		ran <- true
	}()

	// Block until we reach an expected line or timeout
	reader := bufio.NewReader(stderr)
	waitFor := "initializing epoch 0"
	if !assert.Eventually(t, func() bool {
		b, e := reader.Peek(512)
		return e != nil && strings.Contains(string(b), waitFor)
	}, 5*time.Second, 100*time.Millisecond) {
		if err == nil { // first error
			err = fmt.Errorf(`timeout waiting for stderr to contain "%s": runner: %s`, waitFor, r)
		}
	}

	// Even if we had an error, we invoke the shutdown at this point to avoid leaking a process
	shutdown()
	<-ran // block until the runner finished
	return err
}
