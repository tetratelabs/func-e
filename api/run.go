// Package api allows go projects to use func-e as a library.
package api

import (
	"github.com/tetratelabs/func-e/internal/cmd"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/version"
	"io"
	"os"
	"runtime"
)

// HomeDir is an absolute path which most importantly contains "versions"
// installed from EnvoyVersionsURL. Defaults to "${HOME}/.func-e"
func HomeDir(homeDir string) RunOption {
	return func(o *runOpts) {
		o.homeDir = homeDir
	}
}

// EnvoyVersionsURL is the path to the envoy-versions.json.
// Defaults to "https://archive.tetratelabs.io/envoy/envoy-versions.json"
func EnvoyVersionsURL(envoyVersionsURL string) RunOption {
	return func(o *runOpts) {
		o.envoyVersionsURL = envoyVersionsURL
	}
}

// EnvoyVersion overrides the version of Envoy to run. Defaults to the
// contents of "$HomeDir/versions/version".
//
// When that file is missing, it is generated from ".latestVersion" from the
// EnvoyVersionsURL. Its value can be in full version major.minor.patch format,
// e.g. 1.18.1 or without patch component, major.minor, e.g. 1.18.
func EnvoyVersion(envoyVersion string) RunOption {
	return func(o *runOpts) {
		o.envoyVersion = envoyVersion
	}
}

// Out is where status messages are written. Defaults to os.Stdout
func Out(out io.Writer) RunOption {
	return func(o *runOpts) {
		o.out = out
	}
}

// RunOption is configuration for Run.
type RunOption func(*runOpts)

type runOpts struct {
	homeDir          string
	envoyVersion     string
	envoyVersionsURL string
	out              io.Writer
}

func Run(args []string, options ...RunOption) error {
	ro := &runOpts{
		homeDir:          globals.DefaultHomeDir,
		envoyVersion:     "", // default to lookup
		envoyVersionsURL: globals.DefaultEnvoyVersionsURL,
		out:              os.Stdout,
	}
	for _, option := range options {
		option(ro)
	}

	o := globals.GlobalOpts{
		HomeDir:          ro.homeDir,
		EnvoyVersion:     version.PatchVersion(ro.envoyVersion),
		EnvoyVersionsURL: ro.envoyVersion,
		Out:              ro.out,
	}

	funcECmd := cmd.NewApp(&o)

	funcERunArgs := []string{"func-e", "--platform", runtime.GOOS + "/" + runtime.GOARCH, "run"}
	funcERunArgs = append(funcERunArgs, args...)
	return funcECmd.Run(funcERunArgs)
}
