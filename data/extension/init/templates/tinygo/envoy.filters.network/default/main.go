package main

import (
	"bufio"
	"bytes"
	"strings"

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
	// You'd better embed the default root context
	// so that you don't need to reimplement all the methods by yourself.
	proxywasm.DefaultRootContext
	config string
}

func newRootContext(rootContextID uint32) proxywasm.RootContext {
	return &rootContext{}
}

// Override proxywasm.DefaultRootContext
func (ctx *rootContext) OnPluginStart(configurationSize int) types.OnPluginStartStatus {
	counter = proxywasm.DefineCounterMetric(connectionCounterName)

	data, err := proxywasm.GetPluginConfiguration(configurationSize)
	if err != nil && err != types.ErrorStatusNotFound {
		proxywasm.LogCriticalf("failed to load config: %v", err)
		return types.OnPluginStartStatusFailed
	}

	// Ignore comment lines starting with "#" in the extension.txt.
	// Note that we recommend to use json as the configuration format,
	// however, some languages (e.g. TinyGo) does not support ready-to-use json library as of now.
	// As a temporary alternative, we use ".txt" format for the plugin configuration.
	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}
	ctx.config = strings.Join(lines, "\n")
	return types.OnPluginStartStatusOK
}

// Override proxywasm.DefaultRootContext
func (ctx *rootContext) NewStreamContext(contextID uint32) proxywasm.StreamContext {
	return &streamContext{newConnectionMessage: ctx.config}
}

type streamContext struct {
	// You'd better embed the default stream context
	// so that you don't need to reimplement all the methods by yourself.
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
