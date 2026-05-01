// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/alecthomas/kong"

	publicapi "github.com/tetratelabs/func-e/api"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/runtime"
	"github.com/tetratelabs/func-e/internal/version"
)

type (
	// CLI corresponds to the top-level `func-e` command.
	CLI struct {
		HomeDir          string `name:"home-dir" env:"FUNC_E_HOME" help:"(deprecated) func-e home directory - use --config-home, --data-home, --state-home or --runtime-dir instead"`
		ConfigHome       string `name:"config-home" env:"FUNC_E_CONFIG_HOME" help:"directory for configuration files" placeholder:"PATH"`
		DataHome         string `name:"data-home" env:"FUNC_E_DATA_HOME" help:"directory for Envoy binaries" placeholder:"PATH"`
		StateHome        string `name:"state-home" env:"FUNC_E_STATE_HOME" help:"directory for logs (used by run command)" placeholder:"PATH"`
		RuntimeDir       string `name:"runtime-dir" env:"FUNC_E_RUNTIME_DIR" help:"directory for temporary files (used by run command)" placeholder:"PATH"`
		RunID            string `name:"run-id" env:"FUNC_E_RUN_ID" help:"custom run identifier for logs/runtime directories (used by run command)" placeholder:"ID"`
		EnvoyVersionsURL string `name:"envoy-versions-url" env:"ENVOY_VERSIONS_URL" help:"URL of Envoy versions JSON" placeholder:"URL"`
		Platform         string `name:"platform" env:"FUNC_E_PLATFORM" help:"the host OS and architecture of Envoy binaries. Ex. darwin/arm64" placeholder:"OS/ARCH"`

		Run      cmdRun      `cmd:"" passthrough:"" help:"Run Envoy with the given [arguments...] until interrupted"`
		Versions cmdVersions `cmd:"" help:"List Envoy versions"`
		Use      cmdUse      `cmd:"" help:"Sets the current [version] used by the \"run\" command"`
		Which    cmdWhich    `cmd:"" help:"Prints the path to the Envoy binary used by the \"run\" command"`
		Version  versionFlag `name:"version" short:"v" help:"Print the version of func-e"`
	}

	versionFlag bool
)

// BeforeReset is called by kong to print the version and exit.
func (v versionFlag) BeforeReset(app *kong.Kong) error {
	fmt.Fprintf(app.Stdout, "%v version %v\n", app.Model.Name, app.Model.Vars()["version"])
	app.Exit(0)
	return nil
}

// ExitError signals that Kong requested process termination with the given
// code (e.g. --help, --version, or a parse error Kong already printed). The
// caller should translate it into a process exit code.
type ExitError struct {
	Code int
}

// Error implements the error interface.
func (e *ExitError) Error() string {
	return fmt.Sprintf("exit %d", e.Code)
}

// DoMain is the main entry point for the CLI, callable from both main.go and tests.
// On clean termination requested by Kong (e.g. --help), it returns an *ExitError
// carrying the requested process exit code.
func DoMain(ctx context.Context, stdout, stderr io.Writer, args []string, o *globals.GlobalOpts, ver string) error {
	if o == nil {
		o = &globals.GlobalOpts{
			Version: ver,
			Out:     stdout,
			RunOpts: globals.RunOpts{HTTPClientFunc: publicapi.DefaultHTTPClient},
		}
	}
	if o.HTTPClientFunc == nil {
		o.HTTPClientFunc = publicapi.DefaultHTTPClient
	}

	cli, parsed, err := parse(stdout, stderr, args, ver)
	if err != nil {
		return err
	}

	if err := initOpts(cli, o, stderr); err != nil {
		return err
	}

	if parsed.Empty() {
		return parsed.PrintUsage(false)
	}

	o.EnvoyOut = stdout
	o.EnvoyErr = stderr
	parsed.BindTo(ctx, (*context.Context)(nil))
	parsed.BindTo(stdout, (*io.Writer)(nil))
	return parsed.Run(o)
}

// ParseFlags parses CLI args and initializes GlobalOpts without running any command.
// This is used by tests that only want to verify flag/env parsing.
func ParseFlags(stdout, stderr io.Writer, args []string, o *globals.GlobalOpts) error {
	cli, _, err := parse(stdout, stderr, args, "test")
	if err != nil {
		return err
	}
	return initOpts(cli, o, stderr)
}

// parse creates a Kong parser and parses the CLI args. When Kong requests an
// exit (--help, --version, or a fatal parse error), the returned error is an
// *ExitError; Kong has already printed any output. We translate Kong's exit
// callback into a panic so Kong's own control flow halts immediately, then
// recover it here as a typed error.
func parse(stdout, stderr io.Writer, args []string, ver string) (cli *CLI, parsed *kong.Context, err error) {
	defer func() {
		if r := recover(); r != nil {
			if ep, ok := r.(exitPanic); ok {
				err = &ExitError{Code: ep.code}
				return
			}
			panic(r)
		}
	}()

	cli = &CLI{}

	// Kong interprets ${...} as variable interpolation. Since our description
	// contains shell variables like ${HOME}, escape them with $$ so kong
	// passes them through literally.
	desc := strings.ReplaceAll(description, "${", "$${")

	parser, perr := kong.New(cli,
		kong.Name("func-e"),
		kong.Description(desc),
		kong.Writers(stdout, stderr),
		kong.Exit(func(code int) { panic(exitPanic{code: code}) }),
		kong.Vars{"version": ver},
	)
	if perr != nil {
		return nil, nil, fmt.Errorf("creating parser: %w", perr)
	}

	parsed, perr = parser.Parse(args)
	if perr != nil {
		// FatalIfErrorf prints the error and calls our exit panic.
		parser.FatalIfErrorf(perr)
	}
	return cli, parsed, nil
}

// exitPanic carries an exit code out of Kong's parse machinery and is recovered
// at the parse() boundary.
type exitPanic struct{ code int }

// initOpts wires parsed CLI flags into GlobalOpts.
// When a CLI flag is explicitly set, the corresponding pre-set value on o is
// cleared so that InitializeGlobalOpts processes the CLI value instead of
// keeping the test default.
func initOpts(cli *CLI, o *globals.GlobalOpts, stderr io.Writer) error {
	if cli.HomeDir != "" {
		fmt.Fprintln(stderr, "WARNING: $FUNC_E_HOME (--home-dir) is deprecated and will be removed in a future version.")
		fmt.Fprintln(stderr, "Please use --config-home, --data-home, --state-home or --runtime-dir instead.")
	}
	for _, p := range []struct {
		cli   string
		clear func()
	}{
		{cli.ConfigHome, func() { o.ConfigHome = "" }},
		{cli.DataHome, func() { o.DataHome = "" }},
		{cli.StateHome, func() { o.StateHome = "" }},
		{cli.RuntimeDir, func() { o.RuntimeDir = "" }},
		{cli.Platform, func() { o.Platform = "" }},
	} {
		if p.cli != "" {
			p.clear()
		}
	}
	if cli.EnvoyVersionsURL != "" {
		// Also clear GetEnvoyVersions so it gets re-initialized for the new URL
		// instead of keeping a fake server bound by tests.
		o.EnvoyVersionsURL = ""
		o.GetEnvoyVersions = nil
	}
	if err := runtime.InitializeGlobalOpts(o, cli.EnvoyVersionsURL, cli.HomeDir, cli.ConfigHome, cli.DataHome, cli.StateHome, cli.RuntimeDir, cli.Platform, cli.RunID); err != nil {
		return NewValidationError(err.Error())
	}
	return nil
}

// IsExit reports whether err signals a Kong-requested exit (--help, --version,
// fatal parse error). The returned code is the requested process exit code.
func IsExit(err error) (int, bool) {
	if ee, ok := errors.AsType[*ExitError](err); ok {
		return ee.Code, true
	}
	return 0, false
}

var description = fmt.Sprintf(`To run Envoy, execute `+"`func-e run -c your_envoy_config.yaml`"+`. This
downloads and installs the latest version of Envoy for you.

To list versions of Envoy you can use, execute `+"`func-e versions -a`"+`. To
choose one, invoke `+"`func-e use %s`"+`. This installs into
`+"`$FUNC_E_DATA_HOME/envoy-versions/%[1]s`"+`, if not already present. You may
also use minor version, such as `+"`func-e use %s`"+`.

You may want to override `+"`$ENVOY_VERSIONS_URL`"+` to supply custom builds or
otherwise control the source of Envoy binaries. When overriding, validate
your JSON first: %s

Directory structure:
  `+"`$FUNC_E_CONFIG_HOME`"+` stores configuration files
    (default: %s)
  `+"`$FUNC_E_DATA_HOME`"+` stores Envoy binaries
    (default: %s)
  `+"`$FUNC_E_STATE_HOME`"+` stores logs
    (default: %s)
  `+"`$FUNC_E_RUNTIME_DIR`"+` stores temporary files
    (default: %s)

Advanced:
`+"`FUNC_E_PLATFORM`"+` overrides the host OS and architecture of Envoy binaries.
This is used when emulating another platform, e.g. x86 on Apple Silicon M1.
Note: Changing the OS value can cause problems as Envoy has dependencies,
such as glibc. This value must be constant within a `+"`$FUNC_E_DATA_HOME`"+`.`,
	version.LastKnownEnvoy,
	version.LastKnownEnvoyMinor,
	globals.DefaultEnvoyVersionsSchemaURL,
	globals.DefaultConfigHome,
	globals.DefaultDataHome,
	globals.DefaultStateHome,
	globals.DefaultRuntimeDir)
