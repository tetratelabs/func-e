// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type FuncEFactory interface {
	New(ctx context.Context, t *testing.T, stdout, stderr io.Writer) (FuncE, error)
}

// FuncE abstracts func-e, so that the same tests can run for library calls and a compiled func-e binary.
type FuncE interface {
	// Run starts func-e with the given arguments and block until completion
	// The implementation should stop func-e if the context is canceled.
	// The returned error might be a process exit or context cancellation.
	Run(ctx context.Context, args []string) error
	// Interrupt signals the running func-e process to terminate gracefully
	Interrupt(context.Context) error
	// OnStart is called when Envoy starts (after "starting main dispatch loop" is detected)
	OnStart(ctx context.Context) (runDir string, envoyPid int32, err error)
}

// AdminClient represents a client for Envoy's Admin API.
// See: https://github.com/envoyproxy/envoy/blob/main/source/server/admin/admin.cc
type AdminClient struct {
	baseURL string
}

type listenersResponse struct {
	ListenerStatuses []listenerStatus `json:"listener_statuses"`
}

type listenerStatus struct {
	Name         string       `json:"name"`
	LocalAddress localAddress `json:"local_address"`
}

type localAddress struct {
	SocketAddress socketAddress `json:"socket_address"`
}

type socketAddress struct {
	PortValue int `json:"port_value"`
}

// IsReady checks if Envoy is ready to serve traffic by calling the /ready endpoint.
// See: https://github.com/envoyproxy/envoy/blob/main/source/server/admin/admin.cc#L65
func (c *AdminClient) IsReady(ctx context.Context) bool {
	_, err := httpGet(ctx, c.baseURL+"/ready")
	return err == nil
}

// GetListenerBaseURL returns the base URL for a specific listener by name.
// See: https://github.com/envoyproxy/envoy/blob/main/source/server/admin/admin.cc#L75
func (c *AdminClient) GetListenerBaseURL(ctx context.Context, name string) (string, error) {
	var lr listenersResponse
	if err := c.getJSON(ctx, "/listeners", &lr); err != nil {
		return "", err
	}
	for _, ls := range lr.ListenerStatuses {
		if ls.Name == name {
			return fmt.Sprintf("http://127.0.0.1:%d", ls.LocalAddress.SocketAddress.PortValue), nil
		}
	}
	return "", fmt.Errorf("didn't find %s listener", name)
}

func (c *AdminClient) getJSON(ctx context.Context, path string, v interface{}) error {
	body, err := httpGet(ctx, c.baseURL+path+"?format=json")
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// requireAdminReady ensures the Envoy admin client is ready
func requireAdminReady(ctx context.Context, t *testing.T, runDir string) *AdminClient {
	adminAddressPath := filepath.Join(runDir, "admin-address.txt")

	adminAddress, err := os.ReadFile(adminAddressPath)
	require.NoError(t, err)

	adminClient, err := newAdminClient(string(adminAddress))
	require.NoError(t, err)

	require.True(t, adminClient.IsReady(ctx))

	return adminClient
}

// newAdminClient returns a new client for Envoy Admin API.
func newAdminClient(address string) (*AdminClient, error) {
	host, port, err := net.SplitHostPort(strings.TrimSpace(address))
	if err != nil {
		return nil, err
	}
	return &AdminClient{baseURL: fmt.Sprintf("http://%s:%s", host, port)}, nil
}

func httpGet(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
