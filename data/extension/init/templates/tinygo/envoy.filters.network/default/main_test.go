package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxytest"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

func TestNetworkFilter_OnNewConnection(t *testing.T) {
	configuration := `message: this is new connection!`

	opt := proxytest.NewEmulatorOption().
		WithPluginConfiguration([]byte(configuration)).
		WithNewRootContext(newRootContext)

	host := proxytest.NewHostEmulator(opt)
	// Release the host emulation lock so that other test cases can insert their own host emulation.
	defer host.Done()

	// Initialize the plugin and read the config.
	host.StartPlugin()

	// OnNewConnection is called.
	host.InitializeConnection()

	// Retrieve logs emitted to Envoy.
	logs := host.GetLogs(types.LogLevelInfo)
	require.Contains(t, logs, configuration)
}

func TestNetworkFilter_counter(t *testing.T) {
	opt := proxytest.NewEmulatorOption().
		WithNewRootContext(newRootContext)
	host := proxytest.NewHostEmulator(opt)
	// Release the host emulation lock so that other test cases can insert their own host emulation.
	defer host.Done()

	// Initialize the plugin and metric.
	host.StartPlugin()

	// Establish the connection.
	contextID := host.InitializeConnection()

	// Call OnDone on contextID -> increment the connection counter.
	host.CompleteConnection(contextID)

	// Check Envoy logs.
	logs := host.GetLogs(types.LogLevelInfo)
	require.Greater(t, len(logs), 0)
	require.Contains(t, logs, "connection complete!")

	// Check counter.
	actual := counter.Get()
	require.Equal(t, uint64(1), actual)
}
