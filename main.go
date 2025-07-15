// Copyright 2019 Tetrate
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
	"github.com/tetratelabs/func-e/internal/moreos"
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
			moreos.Fprintf(stderr, "%s\n", err)
			logUsageError(app.Name, stderr)
		} else {
			moreos.Fprintf(stderr, "error: %s\n", err)
		}
		return 1
	}
	return 0
}

func logUsageError(name string, stderr io.Writer) {
	moreos.Fprintf(stderr, "show usage with: %s help\n", name)
}
