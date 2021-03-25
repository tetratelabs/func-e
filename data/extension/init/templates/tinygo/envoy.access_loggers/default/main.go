package main

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

func main() {
	proxywasm.SetNewRootContext(newAccessLogger)
}

type accessLogger struct {
	// You'd better embed the default root context
	// so that you don't need to reimplement all the methods by yourself.
	proxywasm.DefaultRootContext
	logMessage string
}

func newAccessLogger(contextID uint32) proxywasm.RootContext {
	return &accessLogger{}
}

// Override proxywasm.DefaultRootContext
func (l *accessLogger) OnPluginStart(configurationSize int) types.OnPluginStartStatus {
	// Read plugin configuration provided in Envoy configuration.
	data, err := proxywasm.GetPluginConfiguration(configurationSize)
	if err != nil && err != types.ErrorStatusNotFound {
		proxywasm.LogCriticalf("failed to load config: %v", err)
		return types.OnPluginStartStatusFailed
	}

	// Ignore comment lines starting with "#" in the configuration.
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
	l.logMessage = strings.Join(lines, "\n")
	return types.OnPluginStartStatusOK
}

// Override proxywasm.DefaultRootContext
func (l *accessLogger) OnLog() {
	hdr, err := proxywasm.GetHttpRequestHeader(":path")
	if err != nil {
		proxywasm.LogCritical(err.Error())
		return
	}

	proxywasm.LogInfof(":path = %s", hdr)
	proxywasm.LogInfof("message = %s", l.logMessage)

}
