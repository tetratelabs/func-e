// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Run execs the Envoy binary at the path with the args passed.
//
// On success, this blocks and returns nil when either `ctx` is done, or the
// process exits with status zero.
func (r *Runtime) Run(ctx context.Context, args []string) error {
	// We require the admin server, so ensure it exists, and we can read its listener via a file path.
	var err error
	r.adminAddressPath, args, err = ensureAdminAddress(r.logf, r.o.RunDir, args)
	if err != nil {
		return err
	}

	// Append the run directory to args for an easy lookup of where pid files etc are stored.
	// Why? MacOS SIP restricts cross-process env var access: we need a solution that works with both Linux and MacOS.
	args = append(args, "--", "--func-e-run-dir", r.o.RunDir)

	cmd := exec.CommandContext(ctx, r.o.EnvoyPath, args...) // #nosec -> users can run whatever binary they like!
	cmd.Stdout = r.Out
	cmd.SysProcAttr = processGroupAttr()

	// Create a pipe to capture stderr and forward to r.Err
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("unable to create stderr pipe: %w", err)
	}

	r.cmd = cmd

	// Print the binary and run directory to the user for debugging purposes.
	r.logf("starting: %s in run directory %s", r.o.EnvoyPath, r.o.RunDir)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to start Envoy process: %w", err)
	}

	// Warn, but don't fail if we can't write the pid file for some reason
	r.maybeWarn(os.WriteFile(filepath.Join(r.o.RunDir, "envoy.pid"), []byte(strconv.Itoa(cmd.Process.Pid)), 0o600))

	// Start a goroutine to scan stderr for "starting main dispatch loop"
	go r.collectAdminDataOnceRunning(ctx, stderrPipe)

	// Wait for the process to exit.
	err = cmd.Wait()
	if err == nil {
		return nil
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		return nil // e.g. graceful shutdown via API
	}
	return err
}

// collectAdminDataOnceRunning scans stderr for the admin address and waits for Envoy to be fully started
// before collecting config_dump to the run directory.
func (r *Runtime) collectAdminDataOnceRunning(ctx context.Context, cmdStderrPipe io.Reader) {
	scanner := bufio.NewScanner(cmdStderrPipe)
	adminCollected := false

	for scanner.Scan() {
		line := scanner.Text()
		// copy stderr to the output writer
		fmt.Fprintln(r.Err, line) //nolint:errcheck

		// Collect config dump when ready
		if !adminCollected && strings.Contains(line, "starting main dispatch loop") {
			adminCollected = true
			adminAddrBytes, err := os.ReadFile(r.adminAddressPath)
			if err != nil {
				r.logf("failed to read admin address from %s: %v", r.adminAddressPath, err)
				continue
			}
			adminAddress := strings.TrimSpace(string(adminAddrBytes))
			r.adminAddress = adminAddress
			// Use a separate goroutine to avoid blocking stderr scanning
			go func(addr string) {
				if err := collectConfigDump(ctx, addr, r.GetRunDir()); err != nil {
					r.logf("failed to collect config_dump from %s: %v", addr, err)
				} else {
					r.logf("collected config_dump from: %s", addr)
				}
			}(adminAddress)
		}
	}

	if err := scanner.Err(); err != nil {
		r.logf("error scanning stderr: %v", err)
	}
}

// collectConfigDump fetches config_dump from Envoy admin API
func collectConfigDump(ctx context.Context, adminAddress, runDir string) error {
	url := fmt.Sprintf("http://%s/config_dump", adminAddress)
	file := filepath.Join(runDir, "config_dump.json")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return copyURLToFile(ctx, url, file)
}

func copyURLToFile(ctx context.Context, url, fullPath string) error {
	// #nosec -> runDir is allowed to be anywhere
	f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("could not open %q: %w", fullPath, err)
	}
	defer f.Close() //nolint:errcheck

	// #nosec -> adminAddress is written by Envoy and the paths are hard-coded
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("could not create request %v: %w", url, err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not read %v: %w", url, err)
	}
	defer res.Body.Close() //nolint:errcheck

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("received %v from %v", res.StatusCode, url)
	}
	if _, err := io.Copy(f, res.Body); err != nil {
		return fmt.Errorf("could not write response body of %v: %w", url, err)
	}
	return nil
}
