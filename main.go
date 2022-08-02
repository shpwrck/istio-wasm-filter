package main

import (
	"fmt"

	"github.com/tidwall/gjson"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

func main() {
	proxywasm.SetVMContext(&vmContext{})
}

type vmContext struct {
	types.DefaultVMContext
}

func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
	return &pluginContext{}
}

type pluginContext struct {
	types.DefaultPluginContext
	configuration pluginConfiguration
	callBack      func(numHeaders, bodySize, numTrailers int)
}

type pluginConfiguration struct {
	remoteURL       string
	responseMapping map[string]string
}

func (ctx *pluginContext) OnPluginStart(pluginConfigurationSize int) types.OnPluginStartStatus {
	data, err := proxywasm.GetPluginConfiguration()
	if err != nil && err != types.ErrorStatusNotFound {
		proxywasm.LogCriticalf("error reading plugin configuration: %v", err)
		return types.OnPluginStartStatusFailed
	}
	config, err := parsePluginConfiguration(data)
	if err != nil {
		proxywasm.LogCriticalf("error parsing plugin configuration: %v", err)
		return types.OnPluginStartStatusFailed
	}
	ctx.configuration = config
	proxywasm.LogWarnf("Remote URL: %v", config.remoteURL)
	for k, v := range config.responseMapping {
		proxywasm.LogWarnf("Response Mapping [%v] -> [%v]", k, v)
	}

	ctx.callBack = func(numHeaders, bodySize, numTrailers int) {
		responseHeaders, err := proxywasm.GetHttpCallResponseHeaders()
		if err != nil {
			proxywasm.LogCriticalf("failed to get response headers: %v", err)
			return
		}
		for _, h := range responseHeaders {
			proxywasm.LogWarnf("response header: %s: %s", h[0], h[1])
		}

		responseBody, err := proxywasm.GetHttpCallResponseBody(0, 10000)
		if err != nil {
			proxywasm.LogCriticalf("failed to get response body: %v", err)
			return
		}
		jsonData := gjson.ParseBytes(responseBody).Get("headers")
		for key, value := range ctx.configuration.responseMapping {
			proxywasm.RemoveHttpRequestHeader(key)
			proxywasm.AddHttpRequestHeader(value, jsonData.Get(key).String())
		}
		proxywasm.LogWarnf("Body: %v", jsonData.Raw)

		proxywasm.ResumeHttpRequest()
		return
	}

	return types.OnPluginStartStatusOK
}

func parsePluginConfiguration(data []byte) (pluginConfiguration, error) {
	if len(data) == 0 {
		return pluginConfiguration{}, nil
	}

	if !gjson.ValidBytes(data) {
		return pluginConfiguration{}, fmt.Errorf("the plugin configuration is not a valid json: %q", string(data))
	}

	config := &pluginConfiguration{}
	jsonData := gjson.ParseBytes(data)
	remoteURL := jsonData.Get("remoteURL").String()
	responseMap := make(map[string]string)
	jsonData.Get("responseMapping").ForEach(func(key, value gjson.Result) bool {
		responseMap[key.String()] = value.String()
		return true
	})

	config.remoteURL = remoteURL
	config.responseMapping = responseMap

	return *config, nil
}

func (ctx *pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &payloadContext{remoteURL: ctx.configuration.remoteURL, callBack: ctx.callBack}
}

type payloadContext struct {
	types.DefaultHttpContext
	totalRequestBodySize int
	remoteURL            string
	callBack             func(numHeaders, bodySize, numTrailers int)
}

var _ types.HttpContext = (*payloadContext)(nil)

func (ctx *payloadContext) OnHttpRequestHeaders(numHeaders int, _ bool) types.Action {
	originalHeaders, err := proxywasm.GetHttpRequestHeaders()
	if err != nil {
		proxywasm.LogCriticalf("failed to forward request headers: %v", err)
	}

	if _, err := proxywasm.DispatchHttpCall(ctx.remoteURL, originalHeaders, nil, nil, 5000, ctx.callBack); err != nil {
		proxywasm.LogCriticalf("dispatch httpcall failed: %v", ctx.remoteURL)
		proxywasm.LogCriticalf("dispatch httpcall failed: %v", err)
	}

	return types.ActionPause
}
