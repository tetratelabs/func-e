// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	docs "github.com/urfave/cli-docs/v3"

	"github.com/tetratelabs/func-e/internal/globals"
)

const siteMarkdownFile = "../../USAGE.md"

// TestUsageMarkdownMatchesCommands is in the "cmd" package because changes here will drift siteMarkdownFile.
func TestUsageMarkdownMatchesCommands(t *testing.T) {
	// Use a custom markdown template
	old := docs.MarkdownDocTemplate
	defer func() { docs.MarkdownDocTemplate = old }()
	docs.MarkdownDocTemplate = `# func-e Overview
{{ .Command.UsageText }}

# Commands

| Name | Usage |
| ---- | ----- |
{{range $index, $cmd := .Command.Commands}}{{if $index}}
{{end}}| {{$cmd.Name}} | {{$cmd.Usage}} |{{end}}
| --version, -v | Print the version of func-e |

# Environment Variables

| Name | Usage | Default |
| ---- | ----- | ------- |
{{range $index, $option := .Command.VisibleFlags}}{{if $index}}
{{end}}| {{index $option.GetEnvVars 0}} | {{$option.GetUsage}} | {{$option.GetDefaultText}} |{{end}}
`
	a := NewApp(&globals.GlobalOpts{})
	expected, err := docs.ToMarkdown(a)
	require.NoError(t, err)

	actual, err := os.ReadFile(siteMarkdownFile)
	require.NoError(t, err)
	require.Equal(t, expected, string(actual))
}
