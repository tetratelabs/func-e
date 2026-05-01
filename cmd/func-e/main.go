// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/tetratelabs/func-e/internal/cmd"
)

func main() {
	args := os.Args
	// Coerce old version flag, so old shell scripts can work.
	if len(args) == 2 && args[1] == "-version" {
		args[1] = "--version"
	}
	os.Exit(run(os.Stdout, os.Stderr, args))
}

// version is the string representation of globals.GlobalOpts
// We can't use debug.ReadBuildInfo because it doesn't get the last known version properly
// See https://github.com/golang/go/issues/37475
var version = "dev"

// run handles all error logging and coding so that no other place needs to.
func run(stdout, stderr io.Writer, args []string) int {
	sigCtx, sigCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer sigCancel()

	err := cmd.DoMain(sigCtx, stdout, stderr, args[1:], nil, version)
	if err == nil {
		return 0
	}
	if code, ok := cmd.IsExit(err); ok {
		// Kong already printed help/usage; just return its requested code.
		return code
	}
	if validationErr, ok := errors.AsType[*cmd.ValidationError](err); ok && validationErr != nil {
		fmt.Fprintf(stderr, "%s\n", err)
		fmt.Fprintf(stderr, "show usage with: func-e --help\n")
	} else {
		fmt.Fprintf(stderr, "error: %s\n", err)
	}
	return 1
}
