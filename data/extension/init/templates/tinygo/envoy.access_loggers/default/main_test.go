package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxytest"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

func TestHelloWorld_OnLog(t *testing.T) {
	opt := proxytest.NewEmulatorOption().
		WithNewRootContext(newAccessLogger)
	host := proxytest.NewHostEmulator(opt)
	defer host.Done() // release the host emulation lock so that other test cases can insert their own host emulation

	host.CallOnLogForAccessLogger(types.Headers{
		{":path", "/this/is/path"},
	}, nil) // call OnLog

	logs := host.GetLogs(types.LogLevelInfo)
	require.Greater(t, len(logs), 0)
	msg := logs[len(logs)-1]

	require.Contains(t, msg, "/this/is/path")
}
