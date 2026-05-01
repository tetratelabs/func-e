// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const siteManpageFile = "../../packaging/nfpm/func-e.8"

func TestManPageMatchesCommands(t *testing.T) {
	expected := generateManPage()

	actual, err := os.ReadFile(siteManpageFile)
	require.NoError(t, err)
	require.Equal(t, expected, string(actual))
}

func generateManPage() string {
	var b strings.Builder
	b.WriteString(`.nh
.TH func-e 8

.SH NAME
func-e \- Install and run Envoy


.SH SYNOPSIS
func-e

.EX
[--config-home]=[value]
[--data-home]=[value]
[--envoy-versions-url]=[value]
[--home-dir]=[value]
[--platform]=[value]
[--run-id]=[value]
[--runtime-dir]=[value]
[--state-home]=[value]
.EE

.PP
\fBUsage\fP:

.EX
`)
	b.WriteString(description)
	b.WriteString(`
.EE


.SH GLOBAL OPTIONS
`)
	flags := []struct{ name, desc string }{
		{"config-home", "directory for configuration files"},
		{"data-home", "directory for Envoy binaries"},
		{"envoy-versions-url", "URL of Envoy versions JSON"},
		{"home-dir", "(deprecated) func-e home directory - use --config-home, --data-home, --state-home or --runtime-dir instead"},
		{"platform", "the host OS and architecture of Envoy binaries. Ex. darwin/arm64"},
		{"run-id", "custom run identifier for logs/runtime directories (used by run command)"},
		{"runtime-dir", "directory for temporary files (used by run command)"},
		{"state-home", "directory for logs (used by run command)"},
	}
	for i, f := range flags {
		fmt.Fprintf(&b, `\fB--%s\fP="": %s`, f.name, f.desc)
		b.WriteByte('\n')
		if i < len(flags)-1 {
			b.WriteString(`
.PP
`)
		}
	}
	b.WriteString(`

.SH COMMANDS
.SH run
Run Envoy with the given [arguments...] until interrupted

.SH versions
List Envoy versions

.PP
\fB--all, -a\fP: Show all versions including ones not yet installed

.SH use
Sets the current [version] used by the "run" command

.SH which
Prints the path to the Envoy binary used by the "run" command
`)
	return b.String()
}
