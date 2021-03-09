package main

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

var (
	requestCounterName = "my_http_filter.request_counter"
	counter            proxywasm.MetricCounter
)

func main() {
	proxywasm.SetNewRootContext(newRootContext)
}

type rootContext struct {
	// we must embed the default context
	proxywasm.DefaultRootContext
	contextID         uint32
	additionalHeaders map[string]string
}

func newRootContext(rootContextID uint32) proxywasm.RootContext {
	return &rootContext{contextID: rootContextID, additionalHeaders: map[string]string{"additional": "header"}}
}

func (ctx *rootContext) OnPluginStart(configurationSize int) bool {
	counter = proxywasm.DefineCounterMetric(requestCounterName)

	// Read plugin configuration provided in Envoy configuration
	data, err := proxywasm.GetPluginConfiguration(configurationSize)
	if err != nil && err != types.ErrorStatusNotFound {
		proxywasm.LogCriticalf("failed to load config: %v", err)
		return false
	}

	// each line in the configuration is in `KEY=VALUE` format
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		tokens := strings.Split(scanner.Text(), "=")
		ctx.additionalHeaders[tokens[0]] = tokens[1]
	}
	return true
}

type httpContext struct {
	// you must embed the default context
	proxywasm.DefaultHttpContext
	additionalHeaders map[string]string
}

func (ctx *rootContext) NewHttpContext(uint32) proxywasm.HttpContext {
	return &httpContext{additionalHeaders: ctx.additionalHeaders}
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
	for key, value := range ctx.additionalHeaders {
		if err := proxywasm.SetHttpResponseHeader(key, value); err != nil {
			proxywasm.LogCriticalf("failed to add header: %v", err)
			return types.ActionPause
		}
		proxywasm.LogInfof("header set: %s %s", key, value)
	}
	return types.ActionContinue
}

func (ctx *httpContext) OnHttpStreamDone() {
	counter.Increment(1)
}
