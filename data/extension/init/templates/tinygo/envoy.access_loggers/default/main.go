package main

import (
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
)

func main() {
	proxywasm.SetNewRootContext(newAccessLogger)
}

type accessLogger struct {
	// you must embed the default context
	proxywasm.DefaultRootContext
}

func newAccessLogger(contextID uint32) proxywasm.RootContext {
	return &accessLogger{}
}

// override
func (ctx *accessLogger) OnLog() {
	hdr, err := proxywasm.GetHttpRequestHeader(":path")
	if err != nil {
		proxywasm.LogCritical(err.Error())
		return
	}

	proxywasm.LogInfof("OnLog: :path = %s", hdr)
}
