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

package cmd_test

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	. "github.com/tetratelabs/getenvoy/pkg/util/cmd"
)

func TestComposableRunEThenE(t *testing.T) {
	expectedErr := errors.New("expected")

	compositeFn := ComposableRunE(func(cmd *cobra.Command, args []string) error {
		return nil
	}).ThenE(func(cmd *cobra.Command, args []string) error {
		return expectedErr
	})

	actualErr := compositeFn(new(cobra.Command), nil)

	// We expect ThenE to be called because the upstream didn't err
	require.Equal(t, expectedErr, actualErr)
}

func TestComposableRunEThenEPriorError(t *testing.T) {
	expectedErr := errors.New("expected")

	compositeFn := ComposableRunE(func(cmd *cobra.Command, args []string) error {
		return expectedErr
	}).ThenE(func(cmd *cobra.Command, args []string) error {
		return errors.New("unexpected")
	})

	actualErr := compositeFn(new(cobra.Command), nil)

	// We expect ThenE to not be called because the upstream erred
	require.Equal(t, expectedErr, actualErr)
}

func TestComposableRunEThen(t *testing.T) {
	nextCalled := false

	compositeFn := ComposableRunE(func(cmd *cobra.Command, args []string) error {
		return nil
	}).Then(func(cmd *cobra.Command, args []string) {
		nextCalled = true
	})

	err := compositeFn(new(cobra.Command), nil)

	// We expect Then to be called because the upstream didn't err
	require.NoError(t, err)
	require.True(t, nextCalled)
}

func TestComposableRunEThenPriorError(t *testing.T) {
	expectedErr := errors.New("expected")
	nextCalled := false

	compositeFn := ComposableRunE(func(cmd *cobra.Command, args []string) error {
		return expectedErr
	}).Then(func(cmd *cobra.Command, args []string) {
		nextCalled = true
	})

	actualErr := compositeFn(new(cobra.Command), nil)

	// We expect Then to not be called because the upstream erred
	require.Equal(t, expectedErr, actualErr)
	require.False(t, nextCalled)
}

func TestCallParentPersistentPreRunE(t *testing.T) {
	parentCalled := false
	root := &cobra.Command{
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			parentCalled = true
		},
	}
	nested := new(cobra.Command)
	root.AddCommand(nested)

	err := CallParentPersistentPreRunE()(nested, nil)

	// We expect the parent to be called
	require.NoError(t, err)
	require.True(t, parentCalled)
}

func TestCallParentPersistentPreRunEParentError(t *testing.T) {
	expectedErr := errors.New("expected")
	root := &cobra.Command{
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return expectedErr
		},
	}
	nested := new(cobra.Command)
	root.AddCommand(nested)

	err := CallParentPersistentPreRunE()(nested, nil)

	// We expect to see the error from the root's PersistentPreRunE
	require.Equal(t, expectedErr, err)
}
