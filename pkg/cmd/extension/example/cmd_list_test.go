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

package example_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/test/cmd"
	. "github.com/tetratelabs/getenvoy/pkg/test/morerequire"
	cmdutil "github.com/tetratelabs/getenvoy/pkg/util/cmd"
)

func TestGetEnvoyExtensionExamplesListFailsOutsideWorkspaceDirectory(t *testing.T) {
	// Change to a non-workspace dir
	dir, revertWd := RequireChDir(t, relativeRustWorkspaceDirWithOneExample+"/..")
	defer revertWd()

	// Run "getenvoy extension examples list"
	c, stdout, stderr := cmd.NewRootCommand()
	c.SetArgs([]string{"extension", "examples", "list"})
	err := cmdutil.Execute(c)

	// Verify the command failed with the expected error
	expectedErr := "there is no extension directory at or above: " + dir
	require.EqualError(t, err, expectedErr, `expected an error running [%v]`, c)
	require.Empty(t, stdout.String(), `expected no stdout running [%v]`, c)
	expectedStderr := fmt.Sprintf("Error: %s\n\nRun 'getenvoy extension examples list --help' for usage.\n", expectedErr)
	require.Equal(t, expectedStderr, stderr.String(), `expected stderr running [%v]`, c)
}

func TestGetEnvoyExtensionExamplesListNone(t *testing.T) {
	// "getenvoy extension examples list" must be in a valid workspace directory
	_, revertWd := RequireChDir(t, relativeTinyGoWorkspaceDirWithNoExample)
	defer revertWd()

	// Run "getenvoy extension examples list"
	c, stdout, stderr := cmd.NewRootCommand()
	c.SetArgs([]string{"extension", "examples", "list"})
	err := cmdutil.Execute(c)

	// Verify lack of examples on list is a warning. not an error.
	require.NoError(t, err, `expected no error running [%v]`, c)
	require.Empty(t, stdout.String(), `expected no stdout running [%v]`, c)
	require.Equal(t, `Extension has no example setups.

Use "getenvoy extension examples add --help" for more information on how to add one.
`, stderr.String(), `unexpected stderr running [%v]`, c)
}

func TestGetEnvoyExtensionExamplesListOne(t *testing.T) {
	// "getenvoy extension examples list" must be in a valid workspace directory
	_, revertWd := RequireChDir(t, relativeRustWorkspaceDirWithOneExample)
	defer revertWd()

	// Run "getenvoy extension examples list"
	c, stdout, stderr := cmd.NewRootCommand()
	c.SetArgs([]string{"extension", "examples", "list"})
	err := cmdutil.Execute(c)

	// Verify the simple name of each example ended up in stdout
	require.NoError(t, err, `expected no error running [%v]`, c)
	require.Equal(t, `EXAMPLE
default
`, stdout.String(), `unexpected stdout running [%v]`, c)
	require.Empty(t, stderr, `expected no stderr running [%v]`, c)
}

func TestGetEnvoyExtensionExamplesListTwo(t *testing.T) {
	// "getenvoy extension examples list" must be in a valid workspace directory
	_, revertWd := RequireChDir(t, relativeWorkspaceDirWithTwoExamples)
	defer revertWd()

	// Run "getenvoy extension examples list"
	c, stdout, stderr := cmd.NewRootCommand()
	c.SetArgs([]string{"extension", "examples", "list"})
	err := cmdutil.Execute(c)

	// Verify the simple name of each example ended up in stdout
	require.NoError(t, err, `expected no error running [%v]`, c)
	require.Equal(t, `EXAMPLE
another
default
`, stdout.String(), `unexpected stdout running [%v]`, c)
	require.Empty(t, stderr, `expected no stderr running [%v]`, c)
}
