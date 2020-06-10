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

package binary

import (
	"context"
	"io"
	"os"

	"github.com/tetratelabs/getenvoy/pkg/manifest"
)

// Runner manages the lifecycle of a binary process
type Runner interface {
	Run(key *manifest.Key, args []string) error
	RunPath(path string, args []string) error
	RegisterPreStart(f ...func(Runner) error)
	RegisterPreTermination(f ...func(Runner) error)
	RegisterWait(int)
	RegisterDone()
	SendSignal(signal os.Signal)
	Status() int
	GetPid() (int, error)
	AppendArgs([]string)
	Wait(int)
	WaitWithContext(context.Context, int)
	DebugStore() string
	SetStdout(func(io.Writer) io.Writer)
	SetStderr(func(io.Writer) io.Writer)
}

const (
	// The Runner's child process is represented as a finite state machine
	// The states are ordered and monotonic i.e. starting -> started -> ready -> terminated (0 -> 1 -> 2 -> 3)
	// Any additional states must be added to the iota in the order they are expected to occur

	// StatusStarting indicates the child process is not yet started
	StatusStarting = iota
	// StatusStarted indicates the child process has started but is not yet ready
	// If there is no concept of readiness for the child process then this status is skipped
	StatusStarted
	// StatusReady indicates the child process is ready
	StatusReady
	// StatusTerminated indicates the child process has been shut down
	StatusTerminated
)

// Fetcher retreives the binary from the location and stores it bases on key
// TODO (Liam): make this less Envoy specific (not using manifest.Key) so it can be reused
type Fetcher interface {
	Fetch(key *manifest.Key, binaryLocation string) error
	AlreadyDownloaded(key *manifest.Key) bool
	BinaryStore() string
}

// FetchRunner combines functionality for running and fetching of a binary
type FetchRunner interface {
	Runner
	Fetcher
	FetchAndRun(reference string, args []string) error
}
