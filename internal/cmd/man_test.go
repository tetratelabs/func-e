// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const siteManpageFile = "../../packaging/nfpm/func-e.8"

func TestManPageMatchesCommands(t *testing.T) {
	actual, err := os.ReadFile(siteManpageFile)
	require.NoError(t, err)
	require.Equal(t, `.nh
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
`+description+`
.EE


.SH GLOBAL OPTIONS
\fB--config-home\fP="": directory for configuration files

.PP
\fB--data-home\fP="": directory for Envoy binaries

.PP
\fB--envoy-versions-url\fP="": URL of Envoy versions JSON

.PP
\fB--home-dir\fP="": (deprecated) func-e home directory - use --config-home, --data-home, --state-home or --runtime-dir instead

.PP
\fB--platform\fP="": the host OS and architecture of Envoy binaries. Ex. darwin/arm64

.PP
\fB--run-id\fP="": custom run identifier for logs/runtime directories (used by run command)

.PP
\fB--runtime-dir\fP="": directory for temporary files (used by run command)

.PP
\fB--state-home\fP="": directory for logs (used by run command)


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
`, string(actual))
}
