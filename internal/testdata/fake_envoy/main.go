// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

// Package main provides a fake Envoy binary for unit testing func-e without downloading real Envoy.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/tetratelabs/func-e/internal"
	"github.com/tetratelabs/func-e/internal/envoy/config"
)

const (
	errorConfig = "At least one of --config-path or --config-yaml or Options::configProto() should be non-empty"
	lf          = "\n" // lf ensures line feeds are realistic
)

type listenerStatus struct {
	Name         string `json:"name"`
	LocalAddress struct {
		SocketAddress struct {
			PortValue int `json:"port_value"`
		} `json:"socket_address"`
	} `json:"local_address"`
}

var listenerStatuses []listenerStatus

// main simulates the behavior of real Envoy for testing purposes:
// - Validates configuration arguments and requires at least one config source
// - Sets up HTTP listeners based on static configurations
// - Starts an admin server with endpoints like /stats, /clusters, /listeners
// - Writes admin address to the specified path if requested
// - Handles both two listener styles used in e2e tests (inline string and static file)
// - Gracefully shuts down on SIGINT or SIGTERM signals
// - Outputs formatted logs to stderr matching Envoy's format
//
// IMPORTANT: Only add logic that simulates what real Envoy does. This code's purpose
// is for unit tests to run without actually downloading Envoy. Do not add artificial
// timeouts or other behaviors that don't match real Envoy.
func main() {
	// Log arguments like Envoy
	fmt.Fprintln(os.Stderr, strings.Join(os.Args, ", "))

	// Initialize epoch
	fmt.Fprintln(os.Stderr, "initializing epoch 0")

	// Parse and validate arguments
	haveConfig, adminAddressPath, configArgs := parseArgs()
	if !haveConfig {
		exit(1, "exiting", errorConfig)
	}

	// Parse listener configs
	cfg, err := config.ParseListeners(configArgs)
	if err != nil {
		exit(1, err.Error())
	}

	// Trap signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup
	var servers []*http.Server
	var listeners []net.Listener

	// Start static listeners
	startStaticListeners(cfg, &wg, &servers, &listeners)

	// Start admin server if configured
	adminAddress, err := config.FindAdminAddress(configArgs)
	if err != nil {
		exit(1, err.Error())
	}
	if adminAddress != "" {
		startAdminServer(adminAddress, adminAddressPath, &wg, &servers, &listeners)
	}

	// Indicate readiness
	fmt.Fprintln(os.Stderr, "starting main dispatch loop")

	// Wait for shutdown signal
	handleShutdown(sigChan, &wg, servers, listeners)
}

// parseArgs processes command-line arguments, collecting config-related flags and detecting admin path.
func parseArgs() (haveConfig bool, adminAddressPath string, configArgs []string) {
	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "run": // Prevent uber bug
			exit(1, "run -- Couldn't find match for argument")
		case "-c", "--config-path", "--config-yaml":
			haveConfig = true
			if i+1 < len(os.Args) {
				configArgs = append(configArgs, os.Args[i], os.Args[i+1])
				i++
			}
		case "--admin-address-path":
			if i+1 < len(os.Args) {
				i++
				adminAddressPath = os.Args[i]
			}
		}
	}
	return haveConfig, adminAddressPath, configArgs
}

// startStaticListeners initializes HTTP servers for static listener configurations.
func startStaticListeners(cfg *config.Config, wg *sync.WaitGroup, servers *[]*http.Server, listeners *[]net.Listener) {
	filterHandlers := map[string]func(http.ResponseWriter, *http.Request){
		internal.StaticFileTypedConfigYaml: staticFileHandler,
		internal.AccessLogTypedConfigYaml:  accessLogHandler,
	}

	for _, l := range cfg.StaticListeners {
		var handler func(http.ResponseWriter, *http.Request)
		for _, f := range l.Filters {
			if h, ok := filterHandlers[f.Config]; ok {
				handler = h
				break
			}
		}
		if handler == nil {
			exit(1, fmt.Sprintf("no handler found for listener '%s'", l.Name))
		}

		ln, err := net.Listen("tcp", l.Address)
		if err != nil {
			exit(1, err.Error())
		}

		addr := ln.Addr().String()
		fmt.Fprintf(os.Stderr, "listener '%s' started on address %s\n", l.Name, addr)

		_, portStr, _ := net.SplitHostPort(addr)
		port := 0
		_, _ = fmt.Sscanf(portStr, "%d", &port)

		// Record listener status with bound port
		listenerStatuses = append(listenerStatuses, listenerStatus{
			Name: l.Name,
			LocalAddress: struct {
				SocketAddress struct {
					PortValue int `json:"port_value"`
				} `json:"socket_address"`
			}{SocketAddress: struct {
				PortValue int `json:"port_value"`
			}{PortValue: port}},
		})

		server := &http.Server{Handler: http.HandlerFunc(handler)}
		*servers = append(*servers, server)
		*listeners = append(*listeners, ln)

		wg.Add(1)
		go func(srv *http.Server, ln net.Listener) {
			defer wg.Done()
			_ = srv.Serve(ln)
		}(server, ln)
	}
}

// startAdminServer sets up the admin HTTP server and writes its address if requested.
func startAdminServer(adminAddress, adminAddressPath string, wg *sync.WaitGroup, servers *[]*http.Server, listeners *[]net.Listener) {
	ln, err := net.Listen("tcp", adminAddress)
	if err != nil {
		exit(1, err.Error())
	}

	addr := ln.Addr().String()
	if adminAddressPath != "" {
		if err := os.WriteFile(adminAddressPath, []byte(addr), 0o600); err != nil {
			exit(1, err.Error())
		}
	}

	serverReady := make(chan struct{})
	adminServer := &http.Server{Handler: http.HandlerFunc(adminEndpoints)}
	*servers = append(*servers, adminServer)
	*listeners = append(*listeners, ln)

	wg.Add(1)
	go func() {
		defer wg.Done()
		close(serverReady)
		_ = adminServer.Serve(ln)
	}()

	<-serverReady
	fmt.Fprintf(os.Stderr, "admin address: %s\n", addr)
}

// handleShutdown waits for a signal and gracefully shuts down all servers and listeners.
func handleShutdown(sigChan <-chan os.Signal, wg *sync.WaitGroup, servers []*http.Server, listeners []net.Listener) {
	s := <-sigChan
	var msg string
	switch s {
	case os.Interrupt:
		msg = "caught SIGINT"
	case syscall.SIGTERM:
		msg = "caught ENVOY_SIGTERM"
	default:
		msg = "caught signal"
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	for _, srv := range servers {
		go func(srv *http.Server) {
			_ = srv.Shutdown(shutdownCtx)
		}(srv)
	}
	for _, ln := range listeners {
		_ = ln.Close()
	}

	wg.Wait()
	exit(0, msg, "exiting")
}

// exit writes messages to stderr and exits with the given code.
func exit(code int, messages ...string) {
	for _, m := range messages {
		fmt.Fprintln(os.Stderr, m)
	}
	os.Exit(code)
}

// adminEndpoints serves endpoints actually used by func-e or its tests.
func adminEndpoints(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/config_dump":
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"configs": []}`))
	case "/ready":
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("LIVE" + lf))
	case "/listeners":
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		b, _ := json.Marshal(struct {
			ListenerStatuses []listenerStatus `json:"listener_statuses"`
		}{ListenerStatuses: listenerStatuses})
		_, _ = w.Write(b)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// accessLogHandler simulates Envoy's access logging behavior, which is used to
// validate the STDOUT of envoy independent of func-e.
func accessLogHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// The corresponding config returns a constant response so we know what the
	// status code will be.
	response := []byte("Hello, World!")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response)

	duration := time.Since(startTime).Milliseconds()
	fmt.Fprintf(os.Stdout, "[%s] \"%s %s %s\" %d - 0 %d %dms - \"-\" \"%s\" \"-\" \"%s\" \"-\"%s",
		startTime.Format("2006-01-02T15:04:05.000-07:00"),
		r.Method,
		r.URL.Path,
		r.Proto,
		200,
		len(response),
		duration,
		r.UserAgent(),
		r.Host,
		lf)
}

// staticFileHandler serves a static file for testing purposes.
func staticFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		data, err := os.ReadFile("response.txt")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("error: " + err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
