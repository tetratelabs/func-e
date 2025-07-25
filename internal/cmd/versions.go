// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

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
			},
		},
		Action: func(c *cli.Context) error {
			rows, err := getInstalledVersions(o.HomeDir)
			if err != nil {
				return err
			}

			currentVersion, currentVersionSource, err := envoy.CurrentVersion(o.HomeDir)
			if err != nil {
				return err
			}

			if c.Bool("all") {
				if evs, err := o.GetEnvoyVersions(c.Context); err != nil {
					return err
				} else if err := addAvailableVersions(&rows, evs.Versions, o.Platform); err != nil {
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
				// TODO: handle when currentVersion is a MinorVersion
				pv, ok := currentVersion.(version.PatchVersion)
				if ok && vr.version == pv {
					fmt.Fprintf(w, "* %s %s (set by %s)\n", vr.version, vr.releaseDate, currentVersionSource) //nolint:errcheck
				} else {
					fmt.Fprintf(w, "  %s %s\n", vr.version, vr.releaseDate) //nolint:errcheck
				}
			}
			return w.Flush()
		},
		CustomHelpTemplate: cli.CommandHelpTemplate,
	}
}

type versionReleaseDate struct {
	version     version.PatchVersion
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
		pv := version.NewPatchVersion(f.Name())
		if i, err := f.Info(); f.IsDir() && pv != "" && err == nil {
			rows = append(rows, versionReleaseDate{
				pv,
				version.ReleaseDate(i.ModTime().Format("2006-01-02")),
			})
		}
	}
	return rows, nil
}

// addAvailableVersions adds remote Envoy versions valid for this platform to "rows", if they don't already exist
func addAvailableVersions(rows *[]versionReleaseDate, remote map[version.PatchVersion]version.Release, p version.Platform) error {
	existingVersions := make(map[version.PatchVersion]bool)
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
