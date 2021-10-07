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
	"sort"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/tetratelabs/func-e/internal/envoy"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/version"
)

// NewUseCmd create a command responsible for downloading and extracting Envoy
func NewUseCmd(o *globals.GlobalOpts) *cli.Command {
	lastKnownEnvoy := getLastKnownEnvoy(o)
	versionsDir := moreos.ReplacePathSeparator("$FUNC_E_HOME/versions/")
	currentVersionWorkingDirFile := moreos.ReplacePathSeparator(envoy.CurrentVersionWorkingDirFile)
	currentVersionHomeDirFile := moreos.ReplacePathSeparator(envoy.CurrentVersionHomeDirFile)

	return &cli.Command{
		Name:      "use",
		Usage:     `Sets the current [version] used by the "run" command`,
		ArgsUsage: "[version]",
		Description: moreos.Sprintf(`The '[version]' is from the "versions -a" command.
The Envoy [version] installs on-demand into `+versionsDir+`[version]
if needed. You may also exclude the patch component of the [version]
to use the latest patch version or to download the binary if it is
not already downloaded.

This updates %s or %s with [version],
depending on which is present.

Example:
$ func-e use %s
$ func-e use %s`, currentVersionWorkingDirFile, currentVersionHomeDirFile, lastKnownEnvoy,
			lastKnownEnvoy[:strings.LastIndex(string(lastKnownEnvoy), ".")]),
		Before: func(c *cli.Context) error {
			return validateVersionArg(c, o)
		},
		Action: func(c *cli.Context) error {
			v := version.Version(c.Args().First())
			latest := v
			if matched := globals.EnvoyStrictMinorVersionPattern.MatchString(string(v)); matched {
				var err error
				if latest, err = o.FuncEVersions.FindLatestPatch(c.Context, v); err != nil {
					if latest, err = getLatestInstalledPatch(o, v); err != nil {
						return err
					}
					o.Logf("couldn't check the latest patch for %q for platform %q using %q instead\n", v, o.Platform, latest)
				}
			}
			if _, err := envoy.InstallIfNeeded(c.Context, o, latest); err != nil {
				return err
			}
			return envoy.WriteCurrentVersion(v, o.HomeDir)
		},
		CustomHelpTemplate: moreos.Sprintf(cli.CommandHelpTemplate),
	}
}

func validateVersionArg(c *cli.Context, o *globals.GlobalOpts) error {
	if c.NArg() == 0 {
		return NewValidationError("missing [version] argument")
	}
	v := c.Args().First()
	if matched := globals.EnvoyMinorVersionPattern.MatchString(v); !matched {
		lastKnownEnvoy := getLastKnownEnvoy(o)
		return NewValidationError("invalid [version] argument: %q should look like %q or %q", v, lastKnownEnvoy,
			lastKnownEnvoy[:strings.LastIndex(string(lastKnownEnvoy), ".")])
	}
	return nil
}

func getLatestInstalledPatch(o *globals.GlobalOpts, minorVersion version.Version) (version.Version, error) {
	rows, err := getInstalledVersions(o.HomeDir)
	if err != nil {
		return "", err
	}
	// Sort so that new release dates appear first and on conflict choosing the higher version.
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].releaseDate == rows[j].releaseDate {
			return rows[i].version > rows[j].version
		}
		return rows[i].releaseDate > rows[j].releaseDate
	})

	strMinorVersion := string(minorVersion)
	splitStrMinorVersion := strings.Split(strMinorVersion, "_debug")

	prefix := splitStrMinorVersion[0] + "."
	hasDebug := strings.HasSuffix(strMinorVersion, "_debug")
	for i := range rows {
		v := string(rows[i].version)
		if strings.HasPrefix(v, prefix) {
			return rows[i].version, nil
		}

		if hasDebug && !strings.HasSuffix(v, "_debug") {
			continue
		} else if !hasDebug && strings.HasSuffix(v, "_debug") {
			continue
		}
	}
	return "", fmt.Errorf("couldn't find the latest patch for %q for platform %q", minorVersion, o.Platform)
}
