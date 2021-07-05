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
	"text/tabwriter"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/tetratelabs/func-e/internal/envoy"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/version"
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

			// tolerate errors determining current version, as that can be due to initial or out-of-band setup
			currentVersion, currentVersionSource, _ := envoy.CurrentVersion(o.HomeDir)

			if c.Bool("all") {
				if ev, err := envoy.FuncEVersions(c.Context, o.EnvoyVersionsURL, globals.CurrentPlatform, version.FuncE); err != nil {
					return err
				} else if err := addAvailableVersions(&rows, ev.Versions, globals.CurrentPlatform); err != nil {
					return err
				}
			}

			// Sort so that new release dates appear first and on conflict choosing the higher version
			sort.Slice(rows, func(i, j int) bool {
				if rows[i].releaseDate == rows[j].releaseDate {
					return rows[i].version > rows[j].version
				}
				return rows[i].releaseDate > rows[j].releaseDate
			})

			// We use a tab writer to ensure we can format the current version
			w := tabwriter.NewWriter(c.App.Writer, 0, 0, 1, ' ', tabwriter.AlignRight)
			for _, vr := range rows { //nolint:gocritic
				if vr.version == currentVersion {
					fmt.Fprintf(w, "* %s %s (set by %s)\n", vr.version, vr.releaseDate, currentVersionSource) //nolint
				} else {
					fmt.Fprintf(w, "  %s %s\n", vr.version, vr.releaseDate) //nolint
				}
			}
			return w.Flush()
		},
	}
}

type versionReleaseDate struct {
	version     version.Version
	releaseDate version.ReleaseDate
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
			rows = append(rows, versionReleaseDate{
				version.Version(f.Name()),
				version.ReleaseDate(i.ModTime().Format("2006-01-02")),
			})
		}
	}
	return rows, nil
}

// addAvailableVersions adds remote Envoy versions valid for this platform to "rows", if they don't already exist
func addAvailableVersions(rows *[]versionReleaseDate, remote map[version.Version]version.Release, p version.Platform) error {
	existingVersions := make(map[version.Version]bool)
	for _, v := range *rows { //nolint:gocritic
		existingVersions[v.version] = true
	}

	for k, v := range remote {
		if _, ok := v.Tarballs[p]; ok && !existingVersions[k] {
			if _, err := time.Parse("2006-01-02", string(v.ReleaseDate)); err != nil {
				return fmt.Errorf("invalid releaseDate of version %q for platform %q: %w", k, p, err)
			}
			*rows = append(*rows, versionReleaseDate{k, v.ReleaseDate})
		}
	}
	return nil
}
