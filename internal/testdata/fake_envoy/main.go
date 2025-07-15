// Package main provides a fake Envoy binary for unit testing func-e without downloading real Envoy.
package main

import (
	"bufio"
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

var listenerStatuses []struct {
	Name         string
	LocalAddress struct {
		SocketAddress struct {
			PortValue int `json:"port_value"`
		} `json:"socket_address"`
	} `json:"local_address"`
}

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
	stderr := bufio.NewWriter(os.Stderr)
	defer func() { stderr.Flush() }()
	// Trap signals so we can respond like Envoy does
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	stderr.WriteString(strings.Join(os.Args, ", "))
	stderr.Flush()
	stderr.WriteString("initializing epoch 0" + lf)
	stderr.Flush()

	// Validate a config is passed just like Envoy
	haveConfig := false
	adminAddressPath := ""
	var configargs []string
	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "run": // prevent uber bug
			exit(1, "run -- Couldn't find match for argument")
		case "-c", "--config-path", "--config-yaml":
			haveConfig = true
			if i+1 < len(os.Args) {
				configargs = append(configargs, os.Args[i], os.Args[i+1])
				i++
			}
		case "--admin-address-path":
			i++
			adminAddressPath = os.Args[i]
		}
	}
	if !haveConfig {
		exit(1, "exiting", errorConfig)
	}

	cfg, err := config.ParseListeners(configargs)
	if err != nil {
		exit(1, err.Error())
	}

	var wg sync.WaitGroup
	var servers []*http.Server
	var listeners []net.Listener

	// Handler registry for known filter configs
	var filterHandlers = map[string]func(http.ResponseWriter, *http.Request){
		internal.MinimalTypedConfigYaml:    inlineStringHandler,
		internal.StaticFileTypedConfigYaml: staticFileHandler,
	}
	// Start listeners for static file config if present
	for _, l := range cfg.StaticListeners {
		var handler func(http.ResponseWriter, *http.Request)
		for _, f := range l.Filters {
			h, ok := filterHandlers[f.Config]
			if ok {
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
		var port int
		fmt.Sscanf(portStr, "%d", &port)
		// Add to listenerStatuses for admin endpoint using the real bound port
		ls := struct {
			Name         string
			LocalAddress struct {
				SocketAddress struct {
					PortValue int `json:"port_value"`
				} `json:"socket_address"`
			} `json:"local_address"`
		}{Name: l.Name}
		ls.LocalAddress.SocketAddress.PortValue = port
		listenerStatuses = append(listenerStatuses, ls)
		server := &http.Server{Handler: http.HandlerFunc(handler)}
		servers = append(servers, server)
		listeners = append(listeners, ln)
		wg.Add(1)
		go func(srv *http.Server, l net.Listener) {
			defer wg.Done()
			srv.Serve(l)
		}(server, ln)
	}

	// Use configparse to detect admin server info
	adminAddress, err := config.FindAdminAddress(configargs)
	if err != nil {
		exit(1, err.Error())
	}
	if adminAddress != "" {
		ln, err := net.Listen("tcp", adminAddress)
		if err != nil {
			exit(1, err.Error())
		}
		adminAddr := ln.Addr().String()
		if adminAddressPath != "" {
			os.WriteFile(adminAddressPath, []byte(adminAddr), 0o600)
		}

		serverReady := make(chan struct{})
		adminServer := &http.Server{Handler: http.HandlerFunc(adminEndpoints)}
		servers = append(servers, adminServer)
		listeners = append(listeners, ln)
		wg.Add(1)
		go func() {
			defer wg.Done()
			close(serverReady)
			adminServer.Serve(ln)
		}()
		<-serverReady

		fmt.Fprintf(os.Stderr, "admin address: %s\n", adminAddr)
	}

	// Print readiness line after admin server is started (or always if no admin)
	fmt.Fprintf(os.Stderr, "starting main dispatch loop\n")

	// Wait for signals like real Envoy
	var msg string
	s := <-c
	switch s {
	case os.Interrupt: // Ex. "kill -2 $pid", Ctrl+C or Ctrl+Break
		msg = "caught SIGINT"
	case syscall.SIGTERM: // Ex. "kill $pid"
		msg = "caught ENVOY_SIGTERM"
	default:
		msg = "caught signal"
	}
	// Gracefully shutdown all servers
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownCancel()
	for _, srv := range servers {
		go srv.Shutdown(shutdownCtx)
	}
	for _, l := range listeners {
		l.Close()
	}
	wg.Wait()
	exit(0, msg, "exiting")
}

func exit(ec int, messages ...string) {
	for _, m := range messages {
		os.Stderr.Write([]byte(m + lf)) //nolint
	}
	os.Exit(ec)
}

func adminEndpoints(w http.ResponseWriter, r *http.Request) {
	// Envoy console messages all write to stderr. Simulate access_log_path: '/dev/stdout'
	os.Stdout.Write([]byte(fmt.Sprintf("GET %s HTTP/1.1%s", r.RequestURI, lf)))

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	switch r.URL.Path {
	case "/stats":
		w.WriteHeader(200)
		w.Write([]byte(`{"stats": []}`))
	case "/clusters":
		w.WriteHeader(200)
		w.Write([]byte(`{"cluster_statuses": []}`))
	case "/certs":
		w.WriteHeader(200)
		w.Write([]byte(`{"certificates": []}`))
	case "/runtime":
		w.WriteHeader(200)
		w.Write([]byte(`{"entries": []}`))
	case "/", "/server_info":
		w.WriteHeader(200)
		w.Write([]byte(`{"version": "fake-envoy/1.0.0","state": "LIVE"}`))
	case "/config_dump":
		w.WriteHeader(200)
		w.Write([]byte(`{"configs": []}`))
	case "/contention":
		w.WriteHeader(200)
		w.Write([]byte(`{"contention": []}`))
	case "/listeners":
		w.WriteHeader(200)
		b, _ := json.Marshal(struct {
			ListenerStatuses interface{} `json:"listener_statuses"`
		}{ListenerStatuses: listenerStatuses})
		w.Write(b)
	case "/memory":
		w.WriteHeader(200)
		w.Write([]byte(`{"allocated": 1048576}`))
	case "/ready":
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		w.WriteHeader(200)
		w.Write([]byte("LIVE" + lf))
	default:
		w.WriteHeader(404)
	}
}

func inlineStringHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("Hello, World!"))
}

func staticFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		data, err := os.ReadFile("response.txt")
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("error: " + err.Error()))
			return
		}
		w.WriteHeader(200)
		w.Write(data)
	} else {
		w.WriteHeader(404)
	}
}
