// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package cmd

// NewValidationError generates an error with a given format string.
// As noted on ValidationError, this is used by main.go to tell the difference between a validation failure vs a runtime
// error. We don't want to clutter output with help suggestions if Envoy failed due to a runtime concern.
func NewValidationError(format string) error {
	return &ValidationError{format}
}

// ValidationError is a marker of a validation error vs an execution one.
type ValidationError struct {
	string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return e.string
}
