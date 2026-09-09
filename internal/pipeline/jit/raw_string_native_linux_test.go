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

func TestNativeRawStringCapabilityAndInput(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	rt, err := openRuntime(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close()
	compiled, err := rt.Compile(`lowercase(message)`, "pipeline-go-1.4.3-datakit")
	if err != nil {
		t.Fatal(err)
	}
	p := compiled.(*nativeProgram)
	defer p.Close()
	capabilities, err := p.Capabilities()
	if err != nil || !capabilities.RawStringValues {
		t.Fatalf("capabilities=%+v error=%v", capabilities, err)
	}
	p.static = nil // Independently decode the indexed mutation result.
	input, err := EncodeFlatPoints([]Point{{Version: 1, Category: "logging", Measurement: "raw", Fields: map[string]any{"message": RawString("A\xffB")}}})
	if err != nil {
		t.Fatal(err)
	}
	batch, err := p.ProcessIndexed(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Records) != 1 || batch.Records[0].Status != TerminalOK {
		t.Fatalf("batch=%+v", batch)
	}
	deltas, err := batch.Records[0].MutationDeltas()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, delta := range deltas {
		for _, op := range delta.Operations {
			if op.Kind == MutationSetField && op.Key == "message" {
				found = true
				if op.Value != "a\ufffdb" {
					t.Fatalf("message=%#v", op.Value)
				}
			}
		}
	}
	if !found {
		t.Fatal("missing message mutation")
	}
}
