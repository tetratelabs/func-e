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

	cmdutil "github.com/tetratelabs/getenvoy/pkg/cmd"
	"github.com/tetratelabs/getenvoy/pkg/globals"
)

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Args))
}

// run handles all error logging and coding so that no other place needs to.
func run(stdout, stderr io.Writer, args []string) int {
	app := cmdutil.NewApp(&globals.GlobalOpts{Out: stdout})
	app.SetArgs(args[1:])
	app.SetOut(stdout)
	app.SetErr(stderr)
	if err := app.Execute(); err != nil {
		if _, ok := err.(*cmdutil.ValidationError); ok {
			logUsageError(app.Name(), err, stderr)
		} else {
			fmt.Fprintln(stderr, "error:", err) //nolint
		}
		return 1
	}
	return 0
}

func logUsageError(name string, err error, stderr io.Writer) {
	fmt.Fprintln(stderr, err)                            //nolint
	fmt.Fprintln(stderr, "show usage with:", name, "-h") //nolint
}
