// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && jitbench && linux && (amd64 || arm64)

package jit

import (
	"fmt"
	"os"
	"testing"
)

// Measures one host-to-native tick with no expiration, including locks/FFI.
// Setup is excluded; Go allocation counts do not include Rust allocations.
func BenchmarkNativeCacheTickScan(b *testing.B) {
	for _, size := range []int{0, 1000, 10000} {
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			path := os.Getenv("PLATYPUS_JIT_RUNTIME")
			if path == "" {
				b.Fatal("requires real PLATYPUS_JIT_RUNTIME")
			}
			rt, err := openRuntime(path, nil)
			if err != nil {
				b.Fatal(err)
			}
			defer rt.Close()
			rt.(*nativeRuntime).manualCacheTicks = true
			compiled, err := rt.Compile(`cache_set(message,"value",1000000000)`, "pipeline-go-1.4.3-datakit")
			if err != nil {
				b.Fatal(err)
			}
			p := compiled.(*nativeProgram)
			defer p.Close()
			if err := p.enableCacheTicks(1_000_000_000); err != nil {
				b.Fatal(err)
			}
			for start := 0; start < size; start += 1000 {
				points := make([]Point, min(1000, size-start))
				for i := range points {
					points[i] = Point{Version: 1, Category: "logging", Measurement: "tick-bench", Fields: map[string]any{"message": fmt.Sprint(start + i)}}
				}
				input, err := EncodeFlatPoints(points)
				if err != nil {
					b.Fatal(err)
				}
				batch, err := p.ProcessIndexed(input)
				if err != nil {
					b.Fatal(err)
				}
				if len(batch.Records) != len(points) {
					b.Fatal("record count")
				}
				for _, record := range batch.Records {
					if record.Status != TerminalOK {
						b.Fatalf("seed failed: %+v", record)
					}
				}
			}
			// Explicit iteration cap prevents the benchmark expiring its fixture.
			if b.N >= 1_000_000_000 {
				b.Fatal("iteration count exceeds fixture TTL")
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := p.advanceCacheTick(); err != nil {
					b.Fatal(err)
				}
			}
			b.StopTimer()
		})
	}
}

func BenchmarkNativeCacheTickExpire(b *testing.B) {
	for _, size := range []int{1000, 10000} {
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			rt, err := openRuntime(os.Getenv("PLATYPUS_JIT_RUNTIME"), nil)
			if err != nil {
				b.Fatal(err)
			}
			defer rt.Close()
			rt.(*nativeRuntime).manualCacheTicks = true
			compiled, err := rt.Compile(`if seed { cache_set(message,"value",1) }; add_key(result,cache_get(message))`, "pipeline-go-1.4.3-datakit")
			if err != nil {
				b.Fatal(err)
			}
			p := compiled.(*nativeProgram)
			defer p.Close()
			p.static = nil
			if err := p.enableCacheTicks(1_000_000_000); err != nil {
				b.Fatal(err)
			}
			var inputs [][]byte
			for start := 0; start < size; start += 1000 {
				points := make([]Point, min(1000, size-start))
				for i := range points {
					points[i] = Point{Version: 1, Category: "logging", Measurement: "expire", Fields: map[string]any{"message": fmt.Sprint(start + i), "seed": true}}
				}
				input, err := EncodeFlatPoints(points)
				if err != nil {
					b.Fatal(err)
				}
				inputs = append(inputs, input)
			}
			probe, err := EncodeFlatPoints([]Point{{Version: 1, Category: "logging", Measurement: "probe", Fields: map[string]any{"message": "0", "seed": false}}})
			if err != nil {
				b.Fatal(err)
			}
			check := func(input []byte, want any) {
				batch, err := p.ProcessIndexed(input)
				if err != nil {
					b.Fatal(err)
				}
				if len(batch.Records) == 0 {
					b.Fatal("empty result")
				}
				for _, record := range batch.Records {
					if record.Status != TerminalOK {
						b.Fatalf("bad record: %+v", record)
					}
					deltas, err := record.MutationDeltas()
					if err != nil {
						b.Fatal(err)
					}
					point := Point{}
					for _, delta := range deltas {
						if err := delta.Apply(&point); err != nil {
							b.Fatal(err)
						}
					}
					if point.Fields["result"] != want {
						b.Fatalf("got=%v want=%v", point.Fields["result"], want)
					}
				}
			}
			b.ReportAllocs()
			b.ResetTimer()
			b.StopTimer()
			for i := 0; i < b.N; i++ {
				for _, input := range inputs {
					check(input, "value")
				}
				b.StartTimer()
				err := p.advanceCacheTick()
				b.StopTimer()
				if err != nil {
					b.Fatal(err)
				}
				check(probe, nil)
			}
		})
	}
}
