// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/tetratelabs/func-e/internal/globals"
)

const siteMarkdownFile = "../../USAGE.md"

// TestUsageMarkdownMatchesCommands is in the "cmd" package because changes here will drift siteMarkdownFile.
func TestUsageMarkdownMatchesCommands(t *testing.T) {
	// Use a custom markdown template
	old := cli.MarkdownDocTemplate
	defer func() { cli.MarkdownDocTemplate = old }()
	cli.MarkdownDocTemplate = `# func-e Overview
{{ .App.UsageText }}

# Commands

| Name | Usage |
| ---- | ----- |
{{range $index, $cmd := .App.VisibleCommands}}{{if $index}}
{{end}}| {{$cmd.Name}} | {{$cmd.Usage}} |{{end}}
| --version, -v | Print the version of func-e |

# Environment Variables

| Name | Usage | Default |
| ---- | ----- | ------- |
{{range $index, $option := .App.VisibleFlags}}{{if $index}}
{{end}}| {{index $option.EnvVars 0}} | {{$option.Usage}} | {{$option.DefaultText}} |{{end}}
`
	a := NewApp(&globals.GlobalOpts{})
	expected, err := a.ToMarkdown()
	require.NoError(t, err)

	actual, err := os.ReadFile(siteMarkdownFile)
	require.NoError(t, err)
	require.Equal(t, expected, string(actual))
}
