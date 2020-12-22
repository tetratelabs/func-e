package main

import (
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

var (
	requestCounterName = "my_http_filter.request_counter"
	counter            proxywasm.MetricCounter
)

func main() {
	proxywasm.SetNewRootContext(newRootContext)
	proxywasm.SetNewHttpContext(newHttpContext)
}

type rootContext struct {
	// we must embed the default context
	proxywasm.DefaultRootContext
	contextID uint32
}

func newRootContext(rootContextID uint32) proxywasm.RootContext {
	return &rootContext{contextID: rootContextID}
}

func (ctx *rootContext) OnVMStart(vmConfigurationSize int) bool {
	counter = proxywasm.DefineCounterMetric(requestCounterName)
	return true
}

type httpContext struct {
	// you must embed the default context
	proxywasm.DefaultHttpContext
	rootContextID, contextID uint32
}

func newHttpContext(rootContextID, contextID uint32) proxywasm.HttpContext {
	return &httpContext{contextID: contextID, rootContextID: rootContextID}
}

func (ctx *httpContext) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
	hs, err := proxywasm.GetHttpRequestHeaders()
	if err != nil {
		proxywasm.LogCriticalf("failed to get request headers: %v", err)
		return types.ActionPause
	}

	proxywasm.LogInfo("observing request headers")
	for _, h := range hs {
		proxywasm.LogInfof("%s: %s", h[0], h[1])
	}

	return types.ActionContinue
}

func (ctx *httpContext) OnHttpResponseHeaders(numHeaders int, endOfStream bool) types.Action {
	if err := proxywasm.SetHttpResponseHeader("additional", "header"); err != nil {
		proxywasm.LogCriticalf("failed to add header: %v", err)
		return types.ActionPause
	}
	return types.ActionContinue
}

func (ctx *httpContext) OnHttpStreamDone() {
	counter.Increment(1)
}
