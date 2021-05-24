// Copyright 2021 Tetrate
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

package cmd

import "fmt"

// NewValidationError generates an error with a given format string.
// As noted on ValidationError, this is used by main.go to tell the difference between a validation failure vs a runtime
// error. We don't want to clutter output with help suggestions if Envoy failed due to a runtime concern.
func NewValidationError(format string, a ...interface{}) error {
	return &ValidationError{fmt.Sprintf(format, a...)}
}

// ValidationError is a marker of a validation error vs an execution one.
type ValidationError struct {
	string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return e.string
}
