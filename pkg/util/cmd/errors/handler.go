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
	"github.com/spf13/cobra"
)

// ErrorHandler represents a strategy to handle command errors.
type ErrorHandler interface {
	// CanHandle returns true if this strategy is applicable to a given error.
	CanHandle(err error) bool
	// Handle handles a given command error.
	Handle(cmd *cobra.Command, err error)
}

// ErrorHandlers represents a prioritized list of error handlers.
type ErrorHandlers []ErrorHandler

// HandlerFor returns an ErrorHandler for a given error or nil
// if no applicable strategy found.
func (hs ErrorHandlers) HandlerFor(err error) ErrorHandler {
	for _, h := range hs {
		if h.CanHandle(err) {
			return h
		}
	}
	return nil
}
