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
	stderrors "errors"
	"strings"

	"github.com/spf13/cobra"
	commonerrors "github.com/tetratelabs/getenvoy/pkg/errors"
)

var (
	// Handlers represents a prioritized list of error handlers.
	Handlers = ErrorHandlers{
		shutdownErrorHandler{},
		defaultErrorHandler{},
	}
)

// defaultErrorHandler represents the default strategy to handle command errors.
type defaultErrorHandler struct{}

func (h defaultErrorHandler) CanHandle(err error) bool {
	return true
}

func (h defaultErrorHandler) Handle(cmd *cobra.Command, err error) {
	if !cmd.SilenceErrors {
		message := err.Error()
		cmd.Println("Error:", message)
		// ensure that an error message is always followed by an empty line
		if !strings.HasSuffix(message, "\n") {
			cmd.Println()
		}
	}
	if !cmd.SilenceUsage {
		cmd.Printf("Run '%v --help' for usage.\n", cmd.CommandPath())
	}
}

// shutdownErrorHandler represents a strategy to handle ShutdownError.
type shutdownErrorHandler struct{}

func (h shutdownErrorHandler) CanHandle(err error) bool {
	return h.asShutdownError(err) != nil
}

func (h shutdownErrorHandler) Handle(cmd *cobra.Command, err error) {
	if serr := h.asShutdownError(err); serr != nil {
		// in case of ShutdownError, we want to avoid any wrapper messages
		cmd.Println("NOTE:", serr.Error())
	}
}

func (h shutdownErrorHandler) asShutdownError(err error) *commonerrors.ShutdownError {
	var serr commonerrors.ShutdownError
	if stderrors.As(err, &serr) {
		return &serr
	}
	return nil
}
