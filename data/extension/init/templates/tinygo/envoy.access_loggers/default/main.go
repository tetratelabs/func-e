package main

import (
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

func main() {
	proxywasm.SetNewRootContext(newAccessLogger)
}

type accessLogger struct {
	// You must embed the default context.
	proxywasm.DefaultRootContext
	logMessage string
}

func newAccessLogger(contextID uint32) proxywasm.RootContext {
	return &accessLogger{}
}

// Override proxywasm.DefaultRootContext
func (l *accessLogger) OnPluginStart(configurationSize int) bool {
	// Read plugin configuration provided in Envoy configuration
	data, err := proxywasm.GetPluginConfiguration(configurationSize)
	if err != nil && err != types.ErrorStatusNotFound {
		proxywasm.LogCriticalf("failed to load config: %v", err)
		return false
	}
	l.logMessage = string(data)
	return true
}

// Override proxywasm.DefaultRootContext
func (l *accessLogger) OnLog() {
	hdr, err := proxywasm.GetHttpRequestHeader(":path")
	if err != nil {
		proxywasm.LogCritical(err.Error())
		return
	}

	proxywasm.LogInfof("OnLog: :path = %s", hdr)
	proxywasm.LogInfof("message = %s", l.logMessage)

}
