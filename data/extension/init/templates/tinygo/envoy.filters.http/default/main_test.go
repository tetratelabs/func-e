package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxytest"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

type headers = [][2]string

func TestHttpHeaders_OnHttpRequestHeaders(t *testing.T) {
	opt := proxytest.NewEmulatorOption().
		WithNewRootContext(newRootContext).
		WithNewHttpContext(newHttpContext)
	host := proxytest.NewHostEmulator(opt)
	defer host.Done() // release the host emulation lock so that other test cases can insert their own host emulation

	host.StartVM() // call OnVMStart -> the metric is initialized

	contextID := host.HttpFilterInitContext() // create http stream

	hs := headers{
		{"key1", "value1"},
		{"key2", "value2"},
	}

	host.HttpFilterPutRequestHeaders(contextID, hs) // call OnHttpRequestHeaders

	logs := host.GetLogs(types.LogLevelInfo)
	require.Greater(t, len(logs), 2)

	require.Equal(t, "key2: value2", logs[len(logs)-1])
	require.Equal(t, "key1: value1", logs[len(logs)-2])
	require.Equal(t, "observing request headers", logs[len(logs)-3])
}
