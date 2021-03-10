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
	host.StartPlugin()

	// Create http context.
	contextID := host.InitializeHttpContext()

	// Call OnHttpRequestHeaders with the given headers.
	hs := types.Headers{
		{"key1", "value1"},
		{"key2", "value2"},
	}
	host.CallOnRequestHeaders(contextID, hs, false)

	// Call OnHttpResponseHeaders.
	host.CallOnResponseHeaders(contextID, nil, false)

	// Check Envoy logs.
	logs := host.GetLogs(types.LogLevelInfo)
	require.Contains(t, logs, "header set: ENVOY=ISTIO")
	require.Contains(t, logs, "header set: HELLO=WORLD")
	require.Contains(t, logs, "header set: additional=header")
	require.Contains(t, logs, "key2: value2")
	require.Contains(t, logs, "key1: value1")
	require.Contains(t, logs, "observing request headers")
}
