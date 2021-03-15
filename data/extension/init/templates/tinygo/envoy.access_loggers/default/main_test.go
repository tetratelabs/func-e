package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxytest"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

func TestAccessLogger_OnLog(t *testing.T) {
	configuration := `this is my log message`
	opt := proxytest.NewEmulatorOption().
		WithNewRootContext(newAccessLogger).
		WithPluginConfiguration([]byte(configuration))

	host := proxytest.NewHostEmulator(opt)
	// Release the host emulation lock so that other test cases can insert their own host emulation.
	defer host.Done()

	// Call OnPluginStart -> the message field of root context is configured.
	status := host.StartPlugin()
	// Check the status returned by OnPluginStart is OK.
	require.Equal(t, types.OnPluginStartStatusOK, status)

	// Call OnLog with the given headers.
	host.CallOnLogForAccessLogger(types.Headers{
		{":path", "/this/is/path"},
	}, nil)

	// Check the Envoy logs.
	logs := host.GetLogs(types.LogLevelInfo)
	require.Contains(t, logs, ":path = /this/is/path")
	require.Contains(t, logs, "message = this is my log message")
}
