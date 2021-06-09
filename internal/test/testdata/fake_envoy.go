package main

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

const errorConfig = "At least one of --config-path or --config-yaml or Options::configProto() should be non-empty"

// lf ensures line feeds are realistic
var lf = func() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}()

// main was originally ported from a shell script. Compiling allows a more realistic test.
func main() {
	// Trap signals so we can respond like Envoy does
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Echo the same first line Envoy would. This also lets code scraping output know signals are trapped
	os.Stderr.Write([]byte("initializing epoch 0" + lf)) //nolint

	// Validate a haveConfig is passed just like Envoy
	haveConfig := false
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-c", "--haveConfig-path", "--haveConfig-yaml":
			haveConfig = true
		}
	}
	if !haveConfig {
		exit(1, "exiting", errorConfig) // oddly, it is this order
	}

	// Echo the same line Envoy would on successful startup
	os.Stderr.Write([]byte("starting main dispatch loop" + lf)) //nolint

	// Envoy console messages all write to stderr. Simulate access_log_path: '/dev/stdout'
	os.Stdout.Write([]byte("GET /ready HTTP/1.1" + lf)) //nolint

	// Block until we receive a signal
	msg := "unexpected"
	select {
	case s := <-c: // Below are how Envoy 1.17 handle signals
		switch s {
		case os.Interrupt: // Ex. "kill -2 $pid", Ctrl+C or Ctrl+Break
			msg = "caught SIGINT"
		case syscall.SIGTERM: // Ex. "kill $pid"
			msg = "caught ENVOY_SIGTERM"
		}
	}

	exit(0, msg, "exiting")
}

func exit(ec int, messages ...string) {
	for _, m := range messages {
		os.Stderr.Write([]byte(m + lf)) //nolint
	}
	os.Exit(ec)
}
