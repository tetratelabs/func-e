package main

// only import moreos, as that's what we are testing
import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/tetratelabs/func-e/internal/moreos"
)

// main simulates ../../.../main.go, but only focuses on sub process control style used by envoy.Run.
// This allows us to write unit tests and identify failures more directly than e2e tests.
//
// Notably, this uses a variable ENVOY_PATH instead of envoy.GetHomeVersion, in order to reduce logic.
//
// In the future, some of this process control structure might move to moreos in order to reduce copy/paste between here
// and internal/envoy/run.go (envoy.Run).
func main() {
	if len(os.Args) < 2 {
		moreos.Fprintf(os.Stderr, "not enough args\n")
		os.Exit(1)
	}

	if os.Args[1] != "run" {
		moreos.Fprintf(os.Stderr, "%s not supported\n", os.Args[1])
		os.Exit(1)
	}

	// This is similar to main.go, except we don't import the validation error
	if err := run(context.Background(), os.Args[2:]); err != nil {
		moreos.Fprintf(os.Stderr, "error: %s\n", err) //nolint
		os.Exit(1)
	}
	os.Exit(0)
}

// simulates envoy.Run with slight adjustments
func run(ctx context.Context, args []string) (err error) { //nolint:gocyclo
	// Like envoy.GetHomeVersion, $FUNC_E_HOME/versions/$(cat $FUNC_E_HOME/version)/bin/envoy$GOEXE.
	cmd := exec.Command(os.Getenv("ENVOY_PATH"), args...)
	cmd.SysProcAttr = moreos.ProcessGroupAttr()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Suppress any error and replace it with the envoy exit status when > 1
	defer func() {
		if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() > 0 {
			if err != nil {
				moreos.Fprintf(cmd.Stdout, "warning: %s\n", err) //nolint
			}
			err = fmt.Errorf("envoy exited with status: %d", cmd.ProcessState.ExitCode())
		}
	}()

	waitCtx, waitCancel := context.WithCancel(ctx)
	defer waitCancel()

	sigCtx, stop := signal.NotifyContext(waitCtx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	moreos.Fprintf(cmd.Stdout, "starting: %s\n", strings.Join(cmd.Args, " ")) //nolint
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to start Envoy process: %w", err)
	}

	// Wait in a goroutine. We may need to kill the process if a signal occurs first.
	go func() {
		defer waitCancel()
		_ = cmd.Wait() // Envoy logs like "caught SIGINT" or "caught ENVOY_SIGTERM", so we don't repeat logging here.
	}()

	// Block until we receive SIGINT or are canceled because Envoy has died.
	<-sigCtx.Done()

	handleShutdown(cmd)

	// Block until it exits to ensure file descriptors are closed prior to archival.
	// Allow up to 5 seconds for a clean stop, killing if it can't for any reason.
	select {
	case <-waitCtx.Done(): // cmd.Wait goroutine finished
	case <-time.After(5 * time.Second):
		_ = moreos.EnsureProcessDone(cmd.Process)
	}
	return
}

// handleShutdown simulates the same named function in envoy.Run, except doesn't run any shutdown hooks.
// This is a copy/paste of envoy.Runtime.interruptEnvoy() with some formatting differences.
func handleShutdown(cmd *exec.Cmd) {
	p := cmd.Process
	moreos.Fprintf(cmd.Stdout, "sending interrupt to envoy (pid=%d)\n", p.Pid) //nolint
	if err := moreos.Interrupt(p); err != nil {
		moreos.Fprintf(cmd.Stdout, "warning: %s\n", err) //nolint
	}
}
