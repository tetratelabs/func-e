package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxytest"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

func TestHttpHeaders_OnHttpRequestHeaders(t *testing.T) {
	configuration := `HELLO=WORLD
ENVOY=ISTIO`

	opt := proxytest.NewEmulatorOption().
		WithNewRootContext(newRootContext).
		WithPluginConfiguration([]byte(configuration))
	host := proxytest.NewHostEmulator(opt)
	defer host.Done() // release the host emulation lock so that other test cases can insert their own host emulation

	host.StartPlugin() // call OnPluginStart -> the metric is initialized

	contextID := host.InitializeHttpContext() // create http stream

	hs := types.Headers{
		{"key1", "value1"},
		{"key2", "value2"},
	}

	host.CallOnRequestHeaders(contextID, hs, false)   // call OnHttpRequestHeaders
	host.CallOnResponseHeaders(contextID, nil, false) // call OnHttpRequestHeaders

	logs := host.GetLogs(types.LogLevelInfo)
	require.Greater(t, len(logs), 1)
	require.Equal(t, "additional header: ENVOY=ISTIO", logs[len(logs)-1])
	require.Equal(t, "additional header: HELLO=WORLD", logs[len(logs)-2])
	require.Equal(t, "key2: value2", logs[len(logs)-3])
	require.Equal(t, "key1: value1", logs[len(logs)-4])
	require.Equal(t, "observing request headers", logs[len(logs)-5])
}
