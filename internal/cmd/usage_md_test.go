// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/version"
)

const siteMarkdownFile = "../../USAGE.md"

// TestUsageMarkdownMatchesCommands is in the "cmd" package because changes here will drift siteMarkdownFile.
func TestUsageMarkdownMatchesCommands(t *testing.T) {
	expected := generateUsageMarkdown()

	actual, err := os.ReadFile(siteMarkdownFile)
	require.NoError(t, err)
	require.Equal(t, expected, string(actual))
}

func generateUsageMarkdown() string {
	_ = version.LastKnownEnvoy // ensure we reference the version package
	return fmt.Sprintf(`# func-e Overview
%s

# Commands

| Name | Usage |
| ---- | ----- |
| run | Run Envoy with the given [arguments...] until interrupted |
| versions | List Envoy versions |
| use | Sets the current [version] used by the "run" command |
| which | Prints the path to the Envoy binary used by the "run" command |
| --version, -v | Print the version of func-e |

# Environment Variables

| Name | Usage | Default |
| ---- | ----- | ------- |
| FUNC_E_HOME | (deprecated) func-e home directory - use --config-home, --data-home, --state-home or --runtime-dir instead |  |
| FUNC_E_CONFIG_HOME | directory for configuration files | %s |
| FUNC_E_DATA_HOME | directory for Envoy binaries | %s |
| FUNC_E_STATE_HOME | directory for logs (used by run command) | %s |
| FUNC_E_RUNTIME_DIR | directory for temporary files (used by run command) | %s |
| FUNC_E_RUN_ID | custom run identifier for logs/runtime directories (used by run command) | auto-generated timestamp |
| ENVOY_VERSIONS_URL | URL of Envoy versions JSON | %s |
| FUNC_E_PLATFORM | the host OS and architecture of Envoy binaries. Ex. darwin/arm64 | $GOOS/$GOARCH |
`, description,
		globals.DefaultConfigHome,
		globals.DefaultDataHome,
		globals.DefaultStateHome,
		globals.DefaultRuntimeDir,
		globals.DefaultEnvoyVersionsURL)
}
