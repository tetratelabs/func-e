// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/shirou/gopsutil/v4/process"

	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/test/build"
	"github.com/tetratelabs/func-e/internal/test/e2e"
)

var (
	// funcEPathEnvKey holds the path to funcEBin.
	funcEPathEnvKey = "E2E_FUNC_E_PATH"
	// funcEBin holds a path to a 'func-e' binary under test.
	funcEBin string
)

// readOrBuildFuncEBin reads E2E_FUNC_E_PATH or builds it like `make build` would have.
func readOrBuildFuncEBin() error {
	// Get the directory where this source file is located
	_, thisFile, _, _ := runtime.Caller(0)
	e2eDir := filepath.Dir(thisFile)
	projectRoot := filepath.Dir(e2eDir) // parent of e2e directory
	if funcEBin = os.Getenv(funcEPathEnvKey); funcEBin != "" {
		if !filepath.IsAbs(funcEBin) {
			funcEBin = filepath.Join(projectRoot, funcEBin, "func-e"+moreos.Exe)
		}
	} else {
		fmt.Fprintf(os.Stderr, "%s was not set. Building %s...\n", funcEPathEnvKey, funcEBin)

		// Create the build directory if it doesn't exist
		buildDir := filepath.Join(projectRoot, "build", fmt.Sprintf("func-e_%s_%s", runtime.GOOS, runtime.GOARCH))
		if err := os.MkdirAll(buildDir, 0o750); err != nil {
			return fmt.Errorf("failed to create build directory %s: %w", buildDir, err)
		}
		var err error
		if funcEBin, err = build.GoBuild(filepath.Join(projectRoot, "main.go"), buildDir); err != nil {
			return err
		}
	}

	// Ensure funcEBin is executable
	if err := os.Chmod(funcEBin, 0o750); err != nil {
		return fmt.Errorf("failed to set executable permissions on %s: %w", funcEBin, err)
	}

	fmt.Fprintln(os.Stderr, "using", funcEBin)
	return nil
}

// funcEExec is a temporary adapter for e2e tests except run.
func funcEExec(args ...string) (string, string, error) {
	cmd := exec.CommandContext(context.Background(), funcEBin, args...)
	cmd.SysProcAttr = moreos.ProcessGroupAttr()
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd.Stdout = io.MultiWriter(os.Stdout, stdout) // we want to see full `func-e` output in the test log
	cmd.Stderr = io.MultiWriter(os.Stderr, stderr)
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// funcEFactory implements runtest.FuncEFactory for E2E tests using a compiled func-e binary.
type funcEFactory struct{}

func (funcEFactory) New(_ context.Context, _ *testing.T, stdout, stderr io.Writer) (e2e.FuncE, error) {
	return &funcE{stdout: stdout, stderr: stderr}, nil
}

// funcE implements runtest.FuncE for e2e tests using the compiled binary
type funcE struct {
	cmd            *exec.Cmd
	stdout, stderr io.Writer
}

// OnStart inspects the running func-e process tree to find the Envoy process and its run directory.
func (a *funcE) OnStart(ctx context.Context) (runDir string, envoyPid int32, err error) {
	if a.cmd == nil || a.cmd.Process == nil {
		return "", 0, fmt.Errorf("no active process")
	}
	funcEPid := int32(a.cmd.Process.Pid)
	funcEProc, err := process.NewProcessWithContext(ctx, funcEPid)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get func-e process: %w", err)
	}

	children, err := funcEProc.Children()
	if err != nil {
		return "", 0, fmt.Errorf("failed to get child processes: %w", err)
	}

	if len(children) == 0 {
		return "", 0, fmt.Errorf("no child processes found")
	}

	envoyProc := children[0]
	envoyPidValue := envoyProc.Pid

	// Get command line args to find run directory
	cmdline, err := envoyProc.CmdlineSlice()
	if err != nil {
		return "", 0, fmt.Errorf("failed to get command line of envoy: %w", err)
	}

	// Look for the run directory in the command line
	for i, arg := range cmdline {
		if arg == "--func-e-run-dir" && i+1 < len(cmdline) {
			return cmdline[i+1], envoyPidValue, nil
		}
	}
	return "", 0, fmt.Errorf("failed to find run dir in envoy args: %v", cmdline)
}

// Run invokes `func-e run args...` and blocks until the process exits.
func (a *funcE) Run(ctx context.Context, args []string) error {
	cmdArgs := append([]string{"run"}, args...)
	a.cmd = exec.CommandContext(ctx, funcEBin, cmdArgs...)
	a.cmd.SysProcAttr = moreos.ProcessGroupAttr()
	a.cmd.Stdout = a.stdout
	a.cmd.Stderr = a.stderr
	if err := a.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start func-e run command: %w", err)
	}
	return a.cmd.Wait() // Block until process exits
}

// Interrupt sends an interrupt signal to the running func-e process.
func (a *funcE) Interrupt(_ context.Context) error {
	if a.cmd == nil || a.cmd.Process == nil {
		return fmt.Errorf("no active process to interrupt")
	}
	return moreos.Interrupt(a.cmd.Process) // Only signal, do not wait
}
