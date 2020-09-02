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

package errors

import (
	"fmt"
	"os"
)

// NewShutdownError returns a new ShutdownError.
func NewShutdownError(signal os.Signal) ShutdownError {
	return ShutdownError{signal: signal}
}

// ShutdownError represents an error caused by a shutdown signal.
type ShutdownError struct {
	signal os.Signal
}

// Error returns the error message.
func (e ShutdownError) Error() string {
	return fmt.Sprintf("Shutting down early because a Ctrl-C (%q) was received.", e.signal)
}
