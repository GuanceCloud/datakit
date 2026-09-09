// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jit

import (
	"os"
	"testing"

	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
)

func TestCIDRPipelineGoUpstreamAndBoundaryCorpus(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 8)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	cases := []struct {
		name   string
		ip     any
		prefix any
		want   bool
	}{
		{"upstream-ipv4-contains", "192.0.2.233", "192.0.2.1/24", true},
		{"upstream-ipv4-not-contains", "192.0.2.233", "192.0.1.1/24", false},
		{"ipv4-host-exact", "192.0.2.233", "192.0.2.233/32", true},
		{"ipv4-host-miss", "192.0.2.234", "192.0.2.233/32", false},
		{"ipv4-zero-prefix", "203.0.113.9", "0.0.0.0/0", true},
		{"ipv6-contains", "2001:db8::7", "2001:db8::1/64", true},
		{"ipv6-not-contains", "2001:db9::7", "2001:db8::1/64", false},
		{"ipv6-exact", "2001:db8::7", "2001:db8::7/128", true},
		{"ipv4-mapped-not-ipv4", "::ffff:192.0.2.1", "192.0.2.0/24", false},
		{"invalid-ip", "999.0.0.1", "0.0.0.0/0", false},
		{"invalid-prefix", "192.0.2.1", "192.0.2.0/33", false},
		{"prefix-with-space", "192.0.2.1", " 192.0.2.0/24", false},
		{"missing-ip", nil, "192.0.2.0/24", false},
		{"missing-prefix", "192.0.2.1", nil, false},
		{"integer-ip", int64(192), "192.0.2.0/24", false},
		{"integer-prefix", "192.0.2.1", int64(24), false},
	}
	for i := range cases {
		ip, ipOK := cases[i].ip.(string)
		prefix, prefixOK := cases[i].prefix.(string)
		if ipOK && prefixOK {
			got, err := funcs.CIDRContains(ip, prefix)
			if got != cases[i].want {
				t.Fatalf("invalid Go oracle fixture %s: got=%v error=%v want=%v", cases[i].name, got, err, cases[i].want)
			}
		}
	}
	const source = `add_key(result,cidr(ip,prefix)); add_key(after,true)`
	check := runner.Check(source)
	if check.Route != RouteJITNative || check.Capabilities.ExecutionMode != "native" {
		t.Fatalf("cidr route=%+v", check)
	}
	for _, size := range []int{1, 2, 4, 8, 10, 128} {
		t.Run("batch-"+itoa(size), func(t *testing.T) {
			points := make([]Point, size)
			want := make([]bool, size)
			for i := range points {
				tc := cases[i%len(cases)]
				fields := map[string]any{"sequence": int64(i)}
				if tc.ip != nil {
					fields["ip"] = tc.ip
				}
				if tc.prefix != nil {
					fields["prefix"] = tc.prefix
				}
				points[i] = Point{Version: 1, Category: "logging", Measurement: tc.name, Fields: fields}
				want[i] = tc.want
			}
			input, err := EncodeFlatPoints(points)
			if err != nil {
				t.Fatal(err)
			}
			batch, err := runner.Process(source, input)
			if err != nil || len(batch.Records) != size {
				t.Fatalf("records=%d error=%v", len(batch.Records), err)
			}
			for i, record := range batch.Records {
				if record.Status != TerminalOK {
					t.Fatalf("record %d terminal=%d error=%s", i, record.Status, record.Error)
				}
				applyHostCompatRecord(t, batch, i, &points[i])
				if points[i].Fields["result"] != want[i] || points[i].Fields["after"] != true || points[i].Fields["sequence"] != int64(i) {
					t.Fatalf("record %d output=%#v want result=%v", i, points[i], want[i])
				}
			}
		})
	}

	// Preserve the upstream literal-argument form in addition to dynamic fields.
	for _, literal := range []struct {
		source string
		want   any
	}{
		{`if cidr("192.0.2.233","192.0.2.1/24") { add_key(result,true) }; add_key(after,true)`, true},
		{`if cidr("192.0.2.233","192.0.1.1/24") { add_key(result,true) }; add_key(after,true)`, nil},
	} {
		input, err := EncodeFlatPoints([]Point{{Version: 1, Category: "logging", Measurement: "literal", Fields: map[string]any{"sentinel": int64(1)}}})
		if err != nil {
			t.Fatal(err)
		}
		batch, err := runner.Process(literal.source, input)
		if err != nil || len(batch.Records) != 1 || batch.Records[0].Status != TerminalOK {
			t.Fatalf("literal source failed: batch=%+v error=%v", batch, err)
		}
		point := Point{Version: 1, Category: "logging", Measurement: "literal", Fields: map[string]any{"sentinel": int64(1)}}
		applyHostCompatRecord(t, batch, 0, &point)
		if point.Fields["result"] != literal.want || point.Fields["after"] != true || point.Fields["sentinel"] != int64(1) {
			t.Fatalf("literal output=%#v want result=%#v", point, literal.want)
		}
	}
}
