// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package globals

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/tetratelabs/func-e/experimental/admin"
	"github.com/tetratelabs/func-e/internal/version"
)

// RunOpts support invocations of "func-e run"
type RunOpts struct {
	// EnvoyPath is the exec.Cmd path to "envoy". Defaults to "$DataHome/envoy-versions/$version/bin/envoy"
	EnvoyPath string
	// EnvoyOut is where to write Envoy's stdout.
	EnvoyOut io.Writer
	// EnvoyErr is where to write Envoy's stderr.
	EnvoyErr io.Writer
	// RunDir is the per-run directory for logs. Generated from StateHome + runID.
	RunDir string
	// RuntimeDir is the per-run directory for ephemeral files. Generated from RuntimeDir + runID.
	RuntimeDir string
	// RunID is the unique identifier for this run. Used in RunDir and RuntimeDir paths.
	RunID string
	// StartupHook is an experimental hook that runs after Envoy starts.
	StartupHook admin.StartupHook
}

// GlobalOpts represents options that affect more than one func-e commands.
//
// Fields representing non-hidden flags have values set according to the following rules:
//  1. value that precedes flag parsing, used in tests
//  2. to a value of the command line argument, e.g. `--home-dir`
//  3. optional mapping to an environment variable, e.g. `FUNC_E_HOME` (not all flags are mapped to ENV)
//  4. otherwise, to the default value, e.g. DefaultDataHome
type GlobalOpts struct {
	// RunOpts are inlined to allow tests to override parameters without changing ENV variables or flags
	RunOpts
	// Version is the version of the CLI, used in help statements and HTTP requests via "User-Agent".
	// Override this via "-X main.version=XXX"
	Version string
	// EnvoyVersionsURL is the path to the envoy-versions.json. Defaults to DefaultEnvoyVersionsURL
	EnvoyVersionsURL string
	// EnvoyVersion is the default version of Envoy to run. Defaults to the contents of "$ConfigHome/envoy-version".
	// When that file is missing, it is generated from ".latestVersion" from the EnvoyVersionsURL. Its
	// value can be in full version major.minor.patch format, e.g. 1.18.1 or without patch component,
	// major.minor, e.g. 1.18.
	EnvoyVersion version.PatchVersion
	// ConfigHome is the directory containing configuration files. Defaults to DefaultConfigHome
	ConfigHome string
	// DataHome is the directory containing Envoy binaries. Defaults to DefaultDataHome
	DataHome string
	// StateHome is the directory containing persistent state like logs. Defaults to DefaultStateHome
	StateHome string
	// RuntimeDir is the directory for ephemeral runtime files. Defaults to DefaultRuntimeDir
	RuntimeDir string
	// HomeDir is the deprecated FUNC_E_HOME directory. When set, legacy paths are used.
	HomeDir string
	// Quiet means don't Logf to Out
	Quiet bool
	// Out is where status messages are written. Defaults to os.Stdout
	Out io.Writer
	// The platform to target for the Envoy install.
	Platform version.Platform
	// GetEnvoyVersions returns Envoy release versions from EnvoyVersionsURL.
	GetEnvoyVersions version.GetReleaseVersions
}

// Logf is used for shared functions that log conditionally on GlobalOpts.Quiet
func (o *GlobalOpts) Logf(format string, a ...interface{}) {
	if o.Quiet { // TODO: we may want to do scoped logging via a Context property, if this becomes common.
		return
	}
	// Always add a newline to ensure consistent formatting
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(o.Out, format, a...) //nolint:errcheck
}

// Mkdirs creates XDG Base Directory directories needed by func-e.
// Only creates directories that will actually be used:
// - ConfigHome and DataHome are always created (for version files and binaries)
// - Per-run directories are only created when RunDir is set (intermediate dirs created automatically)
// Permissions follow XDG spec: RuntimeDir uses 0700, others use 0750.
func (o *GlobalOpts) Mkdirs() error {
	// Base directories always needed (for version files and binary installation)
	dirs := []struct {
		path string
		perm os.FileMode
	}{
		{o.ConfigHome, 0o750},
		{o.DataHome, 0o750},
		{o.EnvoyVersionsDir(), 0o750},
	}

	// Per-run directories (only when actually running Envoy)
	// os.MkdirAll creates intermediate directories automatically
	if o.RunDir != "" {
		dirs = append(dirs,
			struct {
				path string
				perm os.FileMode
			}{o.RunDir, 0o750},
			struct {
				path string
				perm os.FileMode
			}{o.RunOpts.RuntimeDir, 0o700}, // Use embedded RunOpts.RuntimeDir, not base RuntimeDir
		)
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d.path, d.perm); err != nil {
			return fmt.Errorf("unable to create directory %q: %w", d.path, err)
		}
	}
	return nil
}

const (
	// DefaultEnvoyVersionsURL is the default value for GlobalOpts.EnvoyVersionsURL
	DefaultEnvoyVersionsURL = "https://archive.tetratelabs.io/envoy/envoy-versions.json"
	// DefaultEnvoyVersionsSchemaURL is the JSON schema used to validate GlobalOpts.EnvoyVersionsURL
	DefaultEnvoyVersionsSchemaURL = "https://archive.tetratelabs.io/release-versions-schema.json"
	// DefaultPlatform is the current platform of the host machine
	DefaultPlatform = version.Platform(runtime.GOOS + "/" + runtime.GOARCH)
)

var (
	// DefaultConfigHome is the default text for GlobalOpts.ConfigHome
	DefaultConfigHome = "${HOME}/.config/func-e"
	// DefaultDataHome is the default text for GlobalOpts.DataHome
	DefaultDataHome = "${HOME}/.local/share/func-e"
	// DefaultStateHome is the default text for GlobalOpts.StateHome
	DefaultStateHome = "${HOME}/.local/state/func-e"
	// DefaultRuntimeDir is the default text for GlobalOpts.RuntimeDir
	DefaultRuntimeDir = "/tmp/func-e-${UID}"
)
