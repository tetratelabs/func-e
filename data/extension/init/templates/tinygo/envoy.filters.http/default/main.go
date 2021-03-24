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
	// You'd better embed the default root context
	// so that you don't need to reimplement all the methods by yourself.
	proxywasm.DefaultRootContext
	contextID         uint32
	additionalHeaders map[string]string
}

func newRootContext(rootContextID uint32) proxywasm.RootContext {
	return &rootContext{contextID: rootContextID, additionalHeaders: map[string]string{"additional": "header"}}
}

// Override proxywasm.DefaultRootContext
func (ctx *rootContext) OnPluginStart(configurationSize int) types.OnPluginStartStatus {
	// Initialize the counter.
	counter = proxywasm.DefineCounterMetric(requestCounterName)

	// Read plugin configuration provided in Envoy configuration.
	data, err := proxywasm.GetPluginConfiguration(configurationSize)
	if err != nil && err != types.ErrorStatusNotFound {
		proxywasm.LogCriticalf("failed to load config: %v", err)
		return types.OnPluginStartStatusFailed
	}

	// Read the configuration.
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		// Ignore comment lines starting with "#".
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		// Each line in the configuration is in the "KEY=VALUE" format.
		if tokens := strings.Split(scanner.Text(), "="); len(tokens) == 2 {
			ctx.additionalHeaders[tokens[0]] = tokens[1]
		}
	}
	return types.OnPluginStartStatusOK
}

// Override proxywasm.DefaultRootContext
func (ctx *rootContext) NewHttpContext(uint32) proxywasm.HttpContext {
	return &httpContext{additionalHeaders: ctx.additionalHeaders}
}

type httpContext struct {
	// You'd better embed the default http context
	// so that you don't need to reimplement all the methods by yourself.
	proxywasm.DefaultHttpContext
	additionalHeaders map[string]string
}

// Override proxywasm.DefaultHttpContext
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

// Override proxywasm.DefaultHttpContext
func (ctx *httpContext) OnHttpResponseHeaders(numHeaders int, endOfStream bool) types.Action {
	// Set additional headers in the response.
	for key, value := range ctx.additionalHeaders {
		if err := proxywasm.SetHttpResponseHeader(key, value); err != nil {
			proxywasm.LogCriticalf("failed to add header: %v", err)
			return types.ActionPause
		}
		proxywasm.LogInfof("header set: %s=%s", key, value)
	}
	return types.ActionContinue
}

// Override proxywasm.DefaultHttpContext
func (ctx *httpContext) OnHttpStreamDone() {
	counter.Increment(1)
}
