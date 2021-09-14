package main

// only import moreos, as that's what we are testing
import (
	"context"
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

	// Like envoy.GetHomeVersion, $FUNC_E_HOME/versions/$(cat $FUNC_E_HOME/version)/bin/envoy$GOEXE.
	cmd := exec.Command(os.Getenv("ENVOY_PATH"), os.Args[2:]...)
	cmd.SysProcAttr = moreos.ProcessGroupAttr()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Like envoy.Run.
	waitCtx, waitCancel := context.WithCancel(context.Background())
	defer waitCancel()

	sigCtx, stop := signal.NotifyContext(waitCtx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	moreos.Fprintf(os.Stdout, "starting: %s\n", strings.Join(cmd.Args, " ")) //nolint
	if err := cmd.Start(); err != nil {
		moreos.Fprintf(os.Stderr, "error: unable to start Envoy process: %s\n", err)
		os.Exit(1)
	}

	// Wait in a goroutine. We may need to kill the process if a signal occurs first.
	go func() {
		defer waitCancel()
		_ = cmd.Wait() // Envoy logs like "caught SIGINT" or "caught ENVOY_SIGTERM", so we don't repeat logging here.
	}()

	// Block until we receive SIGINT or are canceled because Envoy has died.
	<-sigCtx.Done()

	// Simulate handleShutdown like in envoy.Run.
	_ = moreos.Interrupt(cmd.Process)

	// Block until it exits to ensure file descriptors are closed prior to archival.
	// Allow up to 5 seconds for a clean stop, killing if it can't for any reason.
	select {
	case <-waitCtx.Done(): // cmd.Wait goroutine finished
	case <-time.After(5 * time.Second):
		_ = moreos.EnsureProcessDone(cmd.Process)
	}

	if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() > 0 {
		moreos.Fprintf(os.Stderr, "envoy exited with status: %d\n", cmd.ProcessState.ExitCode())
		os.Exit(1)
	}
	os.Exit(0)
}
