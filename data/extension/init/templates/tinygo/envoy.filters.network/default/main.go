package main

import (
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

var (
	connectionCounterName = "my_network_filter.connection_counter"
	counter               proxywasm.MetricCounter
)

func main() {
	proxywasm.SetNewRootContext(newRootContext)
}

type rootContext struct {
	// You must embed the default context.
	proxywasm.DefaultRootContext
	config string
}

func newRootContext(rootContextID uint32) proxywasm.RootContext {
	return &rootContext{}
}

// Override proxywasm.DefaultRootContext
func (ctx *rootContext) OnPluginStart(configurationSize int) bool {
	counter = proxywasm.DefineCounterMetric(connectionCounterName)

	data, err := proxywasm.GetPluginConfiguration(configurationSize)
	if err != nil && err != types.ErrorStatusNotFound {
		proxywasm.LogCriticalf("failed to load config: %v", err)
		return false
	}
	ctx.config = string(data)
	return true
}

// Override proxywasm.DefaultRootContext
func (ctx *rootContext) NewStreamContext(contextID uint32) proxywasm.StreamContext {
	return &streamContext{newConnectionMessage: ctx.config}
}

type streamContext struct {
	// You must embed the default context.
	proxywasm.DefaultStreamContext
	newConnectionMessage string
}

// Override proxywasm.DefaultStreamContext
func (ctx *streamContext) OnNewConnection() types.Action {
	proxywasm.LogInfo(ctx.newConnectionMessage)
	return types.ActionContinue
}

// Override proxywasm.DefaultStreamContext
func (ctx *streamContext) OnStreamDone() {
	counter.Increment(1)
	proxywasm.LogInfof("connection complete!")
}
