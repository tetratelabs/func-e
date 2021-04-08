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

package cmd

import (
	"github.com/spf13/cobra"
)

// Run represents a callback function that returns no value.
type Run func(cmd *cobra.Command, args []string)

// RunE represents a callback function that returns an error.
type RunE func(cmd *cobra.Command, args []string) error

// ComposableRunE represents a composable callback function.
type ComposableRunE RunE

// ThenE returns a composite callback function that first calls
// the original one and then a given one.
func (fn ComposableRunE) ThenE(nextFn RunE) ComposableRunE {
	return func(cmd *cobra.Command, args []string) error {
		if err := fn(cmd, args); err != nil {
			return err
		}
		return nextFn(cmd, args)
	}
}

// Then returns a composite callback function that first calls
// the original one and then a given one.
func (fn ComposableRunE) Then(nextFn Run) ComposableRunE {
	return fn.ThenE(func(cmd *cobra.Command, args []string) error {
		nextFn(cmd, args)
		return nil
	})
}

// CallParentPersistentPreRunE returns a handler function that calls
// PersistentPreRun* on the parent command.
//
// By default, cobra.Command calls PersistentPreRunE/PersistentPreRun only on the the first
// command that defines that handler.
// Using this function it is possible to compose more sophisticated behavior, e.g. one
// where a handler function is called on multiple commands.
func CallParentPersistentPreRunE() ComposableRunE {
	return func(cmd *cobra.Command, args []string) error {
		return persistentPreRunE(cmd.Parent(), args)
	}
}

// persistentPreRunE walks command tree up and calls PersistentPreRunE/PersistentPreRun
// on the first command that defines that handler.
//
// This function replicates behavior of the original cobra.Command to make it reusable.
func persistentPreRunE(c *cobra.Command, args []string) error {
	for p := c; p != nil; p = p.Parent() {
		if p.PersistentPreRunE != nil {
			if err := p.PersistentPreRunE(c, args); err != nil {
				return err
			}
			break
		} else if p.PersistentPreRun != nil {
			p.PersistentPreRun(c, args)
			break
		}
	}
	return nil
}
