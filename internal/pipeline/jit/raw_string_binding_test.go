// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jit

import (
	"context"
	"testing"
)

type rawBindingProgram struct {
	checkProgram
	closed bool
	calls  int
}

func (p *rawBindingProgram) Close() error { p.closed = true; return nil }
func (p *rawBindingProgram) ProcessIndexed(input []byte) (Batch, error) {
	return p.ProcessIndexedContext(context.Background(), input)
}
func (p *rawBindingProgram) ProcessIndexedContext(context.Context, []byte) (Batch, error) {
	p.calls++
	return Batch{}, nil
}

func TestRawStringBindingPinsGeneration(t *testing.T) {
	for _, oldRaw := range []bool{false, true} {
		t.Run(map[bool]string{false: "upgrade", true: "downgrade"}[oldRaw], func(t *testing.T) {
			old := &rawBindingProgram{checkProgram: checkProgram{capabilities: ProgramCapabilities{DescriptorPresent: true, RawStringValues: oldRaw}}}
			next := &rawBindingProgram{checkProgram: checkProgram{capabilities: ProgramCapabilities{DescriptorPresent: true, RawStringValues: !oldRaw}}}
			runtime := &checkRuntime{program: old}
			runner := &Runner{runtime: runtime, profile: "test", max: 8, cache: make(map[[32]byte]*cacheEntry)}
			defer runner.Close()
			boundOld, err := runner.BindSource("source")
			if err != nil {
				t.Fatal(err)
			}
			defer boundOld.Close()
			runner.Invalidate("source")
			runtime.program = next
			boundNew, err := runner.BindSource("source")
			if err != nil {
				t.Fatal(err)
			}
			defer boundNew.Close()
			if old.closed {
				t.Fatal("old program closed before its binding drained")
			}
			for _, tc := range []struct {
				bound *BoundModules
				raw   bool
			}{{boundOld, oldRaw}, {boundNew, !oldRaw}} {
				projection, err := tc.bound.Projection("source")
				if err != nil || projection.AllowsRawStringValues() != tc.raw {
					t.Fatalf("projection=%+v error=%v", projection, err)
				}
				if _, err := tc.bound.Process("source", nil); err != nil {
					t.Fatal(err)
				}
			}
			if old.calls != 1 || next.calls != 1 {
				t.Fatalf("execution crossed generations: old=%d new=%d", old.calls, next.calls)
			}
			boundOld.Close()
			if !old.closed || next.closed {
				t.Fatal("incorrect generation cleanup")
			}
			if _, err := boundOld.Projection("source"); err == nil {
				t.Fatal("closed binding still negotiates")
			}
			if _, err := boundNew.Projection("different"); err == nil {
				t.Fatal("mismatched source accepted")
			}
		})
	}
}
