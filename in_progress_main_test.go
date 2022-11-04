package main

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/proxytest"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
	"os"
	"testing"
)

func TestHttpHeaders_OnHttpRequestHeaders(t *testing.T) {
	vmTest(t, func(t *testing.T, vm types.VMContext) {
		opt := proxytest.NewEmulatorOption().WithVMContext(vm)
		host, reset := proxytest.NewHostEmulator(opt)
		defer reset()

		// Initialize http context.
		id := host.InitializeHttpContext()

		// Call OnHttpRequestHeaders.
		hs := [][2]string{{"Org", "Test"}, {"Product", "Test"}}
		action := host.CallOnRequestHeaders(id,
			hs, false)
		require.Equal(t, types.ActionPause, action)

		// Check headers.
		resultHeaders := host.GetCurrentRequestHeaders(id)
		t.Log(resultHeaders)
		var found bool
		for _, val := range resultHeaders {
			if val[0] == "Org" {
				require.Equal(t, "Test", val[1])
				found = true
			}
		}
		require.True(t, found)

		// Call OnHttpStreamDone.
		host.CompleteHttpContext(id)

		// Check Envoy logs.
		logs := host.GetInfoLogs()
		require.Contains(t, logs, fmt.Sprintf("%d finished", id))
		require.Contains(t, logs, "request header --> key2: value2")
		require.Contains(t, logs, "request header --> key1: value1")
	})
}

func vmTest(t *testing.T, f func(*testing.T, types.VMContext)) {
	t.Helper()

	t.Run("go", func(t *testing.T) {
		f(t, &vmContext{})
	})

	t.Run("wasm", func(t *testing.T) {
		wasm, err := os.ReadFile("main.wasm")
		if err != nil {
			t.Skip("wasm not found")
		}
		v, err := proxytest.NewWasmVMContext(wasm)
		require.NoError(t, err)
		defer v.Close()
		f(t, v)
	})
}
