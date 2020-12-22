package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxytest"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

func TestNetwork_OnNewConnection(t *testing.T) {
	opt := proxytest.NewEmulatorOption().
		WithNewRootContext(newRootContext).
		WithNewStreamContext(newStreamContext)
	host := proxytest.NewHostEmulator(opt)
	defer host.Done() // release the host emulation lock so that other test cases can insert their own host emulation

	host.NetworkFilterInitConnection() // OnNewConnection is called

	logs := host.GetLogs(types.LogLevelInfo) // retrieve logs emitted to Envoy
	assert.Equal(t, logs[0], "new connection!")
}

func TestNetwork_counter(t *testing.T) {
	opt := proxytest.NewEmulatorOption().
		WithNewRootContext(newRootContext).
		WithNewStreamContext(newStreamContext)
	host := proxytest.NewHostEmulator(opt)
	defer host.Done() // release the host emulation lock so that other test cases can insert their own host emulation

	host.StartVM() // init metric

	contextID := host.NetworkFilterInitConnection()
	host.NetworkFilterCompleteConnection(contextID) // call OnDone on contextID -> increment the connection counter

	logs := host.GetLogs(types.LogLevelInfo)
	require.Greater(t, len(logs), 0)

	assert.Equal(t, "connection complete!", logs[len(logs)-1])
	actual := counter.Get()
	assert.Equal(t, uint64(1), actual)
}
