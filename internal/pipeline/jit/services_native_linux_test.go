// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package jit

import (
	"os"
	"testing"
)

func TestNativeInstanceServicesValidationAndClose(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	for _, config := range []ServiceConfig{
		{Version: 2},
		{Version: 1, HTTP: HTTPServiceConfig{CIDRWhitelist: []string{"invalid"}}},
	} {
		if runner, err := NewRunnerWithServices(path, "pipeline-go-1.4.3-datakit", 4, nil, config); err == nil {
			_ = runner.Close()
			t.Fatal("invalid candidate accepted")
		}
	}
	runner, err := NewRunnerWithServices(path, "pipeline-go-1.4.3-datakit", 4, NewPipelineGoHost(nil), ServiceConfig{Version: 1})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = runner.Close() })
	native := runner.runtime.(*nativeRuntime)
	if native.services == 0 {
		t.Fatal("missing instance handle")
	}
	if err := native.configureServices(ServiceConfig{Version: 1}); err == nil {
		t.Fatal("snapshot mutated after construction")
	}
	if result := runner.Check("add_key(ok,true)"); result.Route != RouteJITNative {
		t.Fatalf("single script: %+v", result)
	}
	program, err := native.CompileModules("main.p", map[string]string{"main.p": `use("child.p")`, "child.p": `add_key(child,true)`})
	if err != nil {
		t.Fatal(err)
	}
	if err := program.Close(); err != nil {
		t.Fatal(err)
	}
	if err := runner.Close(); err != nil {
		t.Fatal(err)
	}
	if native.services != 0 || native.handle != 0 {
		t.Fatal("runtime retained handles after close")
	}
	if err := runner.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestInstanceServicesMissingABIFailsExplicitly(t *testing.T) {
	native := &nativeRuntime{}
	if err := native.configureServices(ServiceConfig{Version: 1}); err == nil {
		t.Fatal("missing ABI silently ignored")
	}
}
