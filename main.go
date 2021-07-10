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
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v2"

	cmdutil "github.com/tetratelabs/func-e/internal/cmd"
	"github.com/tetratelabs/func-e/internal/globals"
	versionutil "github.com/tetratelabs/func-e/internal/version"
)

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Args))
}

// version is the string representation of globals.GlobalOpts
var version = "dev"

// run handles all error logging and coding so that no other place needs to.
func run(stdout, stderr io.Writer, args []string) int {
	app := cmdutil.NewApp(&globals.GlobalOpts{Version: versionutil.Version(version), Out: stdout})
	app.Writer = stdout
	app.ErrWriter = stderr
	app.Action = func(c *cli.Context) error {
		command := c.Args().First()
		if command == "" { // Show help by default
			return cli.ShowSubcommandHelp(c)
		}
		return cmdutil.NewValidationError("unknown command %q", command)
	}
	app.OnUsageError = func(c *cli.Context, err error, isSub bool) error {
		return cmdutil.NewValidationError(err.Error())
	}
	if err := app.Run(args); err != nil {
		if _, ok := err.(*cmdutil.ValidationError); ok {
			fmt.Fprintln(stderr, err) //nolint
			logUsageError(app.Name, stderr)
		} else {
			fmt.Fprintln(stderr, "error:", err) //nolint
		}
		return 1
	}
	return 0
}

func logUsageError(name string, stderr io.Writer) {
	fmt.Fprintln(stderr, "show usage with:", name, "help") //nolint
}
