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

	"github.com/urfave/cli/v2"

	"github.com/tetratelabs/func-e/internal/cmd"
	"github.com/tetratelabs/func-e/internal/globals"
)

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Args))
}

// version is the string representation of globals.GlobalOpts
// We can't use debug.ReadBuildInfo because it doesn't get the last known version properly
// See https://github.com/golang/go/issues/37475
var version = "dev"

// run handles all error logging and coding so that no other place needs to.
func run(stdout, stderr io.Writer, args []string) int {
	app := cmd.NewApp(&globals.GlobalOpts{Version: version, Out: stdout})
	app.Writer = stdout
	app.ErrWriter = stderr
	app.Action = func(c *cli.Context) error {
		command := c.Args().First()
		if command == "" { // Show help by default
			return cli.ShowSubcommandHelp(c)
		}
		return cmd.NewValidationError(fmt.Sprintf("unknown command %q", command))
	}
	app.OnUsageError = func(c *cli.Context, err error, isSub bool) error {
		return cmd.NewValidationError(err.Error())
	}
	sigCtx, sigCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer sigCancel()
	if err := app.RunContext(sigCtx, args); err != nil {
		var validationErr *cmd.ValidationError
		if errors.As(err, &validationErr) {
			fmt.Fprintf(stderr, "%s\n", err) //nolint:errcheck
			logUsageError(app.Name, stderr)
		} else {
			fmt.Fprintf(stderr, "error: %s\n", err) //nolint:errcheck
		}
		return 1
	}
	return 0
}

func logUsageError(name string, stderr io.Writer) {
	fmt.Fprintf(stderr, "show usage with: %s help\n", name) //nolint:errcheck
}
