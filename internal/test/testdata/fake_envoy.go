package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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

	// Validate a config is passed just like Envoy
	haveConfig := false
	adminAddressPath := ""
	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-c", "--config-path", "--config-yaml":
			haveConfig = true
		case "--admin-address-path":
			i++
			adminAddressPath = os.Args[i]
		}
	}
	if !haveConfig {
		exit(1, "exiting", errorConfig) // oddly, it is this order
	}

	// Start a fake admin listener that write the same sort of response Envoy's /ready would, but on all endpoints.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Envoy console messages all write to stderr. Simulate access_log_path: '/dev/stdout'
		os.Stdout.Write([]byte(fmt.Sprintf("GET %s HTTP/1.1%s", r.RequestURI, lf))) //nolint

		w.Header().Add("Content-Type", "text/plain; charset=UTF-8")
		w.WriteHeader(200)
		w.Write([]byte("LIVE" + lf)) //nolint
	}))
	defer ts.Close()
	adminAddress := ts.Listener.Addr().String() // ex. 127.0.0.1:55438

	// We don't echo the admin address intentionally as it makes tests complicated as they
	// would have to use regex to address the random port value.
	if adminAddressPath != "" {
		os.WriteFile(adminAddressPath, []byte(adminAddress), 0600)
	}

	// Echo the same line Envoy would on successful startup
	os.Stderr.Write([]byte("starting main dispatch loop" + lf)) //nolint

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
