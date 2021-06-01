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
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/tetratelabs/getenvoy/internal/envoy"
	"github.com/tetratelabs/getenvoy/internal/globals"
	"github.com/tetratelabs/getenvoy/internal/version"
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
			rows, err := getInstalledVersions(o.HomeDir)
			if err != nil {
				return err
			}
			if c.Bool("all") {
				if ev, err := envoy.GetEnvoyVersions(o.EnvoyVersionsURL, o.UserAgent); err != nil {
					return err
				} else if err := addAvailableVersions(&rows, ev.Versions, globals.CurrentPlatform); err != nil {
					return err
				}
			}

			if len(rows) == 0 {
				fmt.Fprintln(c.App.Writer, "No Envoy versions, yet") //nolint
				return nil
			}

			// Sort so that new release dates appear first and on conflict choosing the higher version
			sort.Slice(rows, func(i, j int) bool {
				if rows[i].releaseDate == rows[j].releaseDate {
					return rows[i].version > rows[j].version
				}
				return rows[i].releaseDate > rows[j].releaseDate
			})

			// This doesn't use tabwriter because the columns are likely to remain the same width.
			fmt.Fprintln(c.App.Writer, "VERSION\tRELEASE_DATE") //nolint
			for _, vr := range rows {                           //nolint:gocritic
				fmt.Fprintf(c.App.Writer, "%s\t%s\n", vr.version, vr.releaseDate) //nolint
			}
			return nil
		},
	}
}

type versionReleaseDate struct {
	// version ex "1.15.5"
	version string
	// releaseDate ex "2021-05-11"
	releaseDate string
}

func getInstalledVersions(homeDir string) ([]versionReleaseDate, error) {
	var rows []versionReleaseDate
	files, err := os.ReadDir(filepath.Join(homeDir, "versions"))
	if os.IsNotExist(err) {
		return rows, nil
	} else if err != nil {
		return nil, err
	}

	for _, f := range files {
		if i, err := f.Info(); f.IsDir() && err == nil {
			rows = append(rows, versionReleaseDate{f.Name(), i.ModTime().Format("2006-01-02")})
		}
	}
	return rows, nil
}

// addAvailableVersions adds remote Envoy versions valid for this platform to "vr", if they don't already exist
func addAvailableVersions(rows *[]versionReleaseDate, remote map[string]version.EnvoyVersion, p string) error {
	existingVersions := make(map[string]bool)
	for _, v := range *rows { //nolint:gocritic
		existingVersions[v.version] = true
	}

	for k, v := range remote {
		if _, ok := v.Tarballs[p]; ok && !existingVersions[k] {
			if _, err := time.Parse("2006-01-02", v.ReleaseDate); err != nil {
				return fmt.Errorf("invalid releaseDate of version %q for platform %q: %w", k, p, err)
			}
			*rows = append(*rows, versionReleaseDate{k, v.ReleaseDate})
		}
	}
	return nil
}
