package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxytest"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

func TestNetworkFilter_OnNewConnection(t *testing.T) {
	configuration := `# this is comment line, and should be ignored.
message: this is new connection!`
	opt := proxytest.NewEmulatorOption().
		WithPluginConfiguration([]byte(configuration)).
		WithNewRootContext(newRootContext)

	host := proxytest.NewHostEmulator(opt)
	// Release the host emulation lock so that other test cases can insert their own host emulation.
	defer host.Done()

	// Initialize the plugin and read the config.
	status := host.StartPlugin()
	// Check the status returned by OnPluginStart is OK.
	require.Equal(t, types.OnPluginStartStatusOK, status)

	// OnNewConnection is called.
	host.InitializeConnection()

	// Retrieve logs emitted to Envoy.
	logs := host.GetLogs(types.LogLevelInfo)
	require.Contains(t, logs, "message: this is new connection!")
}

func TestNetworkFilter_counter(t *testing.T) {
	opt := proxytest.NewEmulatorOption().
		WithNewRootContext(newRootContext)

	host := proxytest.NewHostEmulator(opt)
	// Release the host emulation lock so that other test cases can insert their own host emulation.
	defer host.Done()

	// Initialize the plugin and metric.
	status := host.StartPlugin()
	// Check the status returned by OnPluginStart is OK.
	require.Equal(t, types.OnPluginStartStatusOK, status)

	// Establish the connection.
	contextID, action := host.InitializeConnection()
	// Check the status returned by OnNewConnection is ActionContinue.
	require.Equal(t, types.ActionContinue, action)

	// Call OnDone on contextID -> increment the connection counter.
	host.CompleteConnection(contextID)

	// Check Envoy logs.
	logs := host.GetLogs(types.LogLevelInfo)
	require.Greater(t, len(logs), 0)
	require.Contains(t, logs, "connection complete!")

	// Check counter.
	value, err := host.GetCounterMetric("my_network_filter.connection_counter")
	require.NoError(t, err)
	require.Equal(t, uint64(1), value)
}
