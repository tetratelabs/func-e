package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxytest"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

func TestHttpFilter_OnHttpRequestHeaders(t *testing.T) {
	configuration := `HELLO=WORLD
ENVOY=ISTIO`
	opt := proxytest.NewEmulatorOption().
		WithNewRootContext(newRootContext).
		WithPluginConfiguration([]byte(configuration))

	host := proxytest.NewHostEmulator(opt)
	// Release the host emulation lock so that other test cases can insert their own host emulation.
	defer host.Done()

	// Call OnPluginStart -> the metric is initialized.
	status := host.StartPlugin()
	// Check the status returned by OnPluginStart is OK.
	require.Equal(t, types.OnPluginStartStatusOK, status)

	// Create http context.
	contextID := host.InitializeHttpContext()

	// Call OnHttpRequestHeaders with the given headers.
	hs := types.Headers{
		{"key1", "value1"},
		{"key2", "value2"},
	}
	action := host.CallOnRequestHeaders(contextID, hs, false)
	// Check the action returned by OnRequestHeaders is Continue.
	require.Equal(t, types.ActionContinue, action)

	// Call OnHttpResponseHeaders.
	action = host.CallOnResponseHeaders(contextID, nil, false)
	// Check the action returned by OnResponseHeaders is Continue.
	require.Equal(t, types.ActionContinue, action)

	// Check Envoy logs.
	logs := host.GetLogs(types.LogLevelInfo)
	require.Contains(t, logs, "header set: ENVOY=ISTIO")
	require.Contains(t, logs, "header set: HELLO=WORLD")
	require.Contains(t, logs, "header set: additional=header")
	require.Contains(t, logs, "key2: value2")
	require.Contains(t, logs, "key1: value1")
	require.Contains(t, logs, "observing request headers")
}
