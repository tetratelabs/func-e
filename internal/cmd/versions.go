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

package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/urfave/cli/v2"

	"github.com/tetratelabs/getenvoy/internal/envoy"
	"github.com/tetratelabs/getenvoy/internal/globals"
)

// NewVersionsCmd returns command that lists available Envoy versions for the current platform.
func NewVersionsCmd(o *globals.GlobalOpts) *cli.Command {
	return &cli.Command{
		Name:  "versions",
		Usage: "List Envoy versions",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "Show all versions including ones not yet installed",
			}},
		Action: func(c *cli.Context) error {
			vd := map[string]string{}
			if err := addInstalledVersions(vd, o.HomeDir); err != nil {
				return err
			}
			if c.Bool("all") {
				if ev, err := envoy.GetEnvoyVersions(o.EnvoyVersionsURL, o.UserAgent); err != nil {
					return err
				} else if err := envoy.AddVersions(vd, ev.Versions, globals.CurrentPlatform); err != nil {
					return err
				}
			}
			if len(vd) == 0 {
				fmt.Fprintln(c.App.Writer, "No Envoy versions, yet") //nolint
			} else {
				printVersions(vd, c.App.Writer)
			}
			return nil
		},
	}
}

// addInstalledVersions adds installed Envoy versions
func addInstalledVersions(vd map[string]string, envoyHome string) error {
	files, err := os.ReadDir(filepath.Join(envoyHome, "versions"))
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	for _, f := range files {
		if i, err := f.Info(); f.IsDir() && err == nil {
			vd[f.Name()] = i.ModTime().Format("2006-01-02")
		}
	}
	return nil
}

// printVersions retrieves the Envoy versions from the passed location and writes it to the passed writer
func printVersions(vd map[string]string, w io.Writer) {
	// Build a list of Envoy versions with release date for this platform
	type versionReleaseDate struct{ version, releaseDate string }

	rows := make([]versionReleaseDate, 0, len(vd))
	for v, d := range vd {
		rows = append(rows, versionReleaseDate{v, d})
	}

	// Sort so that new release dates appear first and on conflict choosing the higher version
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].releaseDate == rows[j].releaseDate {
			return rows[i].version > rows[j].version
		}
		return rows[i].releaseDate > rows[j].releaseDate
	})

	// This doesn't use tabwriter because the columns are likely to remain the same width for the foreseeable future.
	fmt.Fprintln(w, "VERSION\tRELEASE_DATE") //nolint
	for _, vr := range rows {                //nolint:gocritic
		fmt.Fprintf(w, "%s\t%s\n", vr.version, vr.releaseDate) //nolint
	}
}
