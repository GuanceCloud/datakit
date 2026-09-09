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
	"context"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"
)

type datetimeVector struct {
	name, precision, format, timezone, want string
	input                                   any
	source                                  string
}

func TestDatetimePipelineGoUpstreamAndBoundaryCorpus(t *testing.T) {
	const childKey = "DATAKIT_JIT_DATETIME_UPSTREAM_CHILD"
	if os.Getenv(childKey) != "1" {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^"+t.Name()+"$", "-test.count=1")
		cmd.Env = append(os.Environ(), childKey+"=1", "TZ=CST-8")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("datetime child: %v\n%s", err, output)
		}
		return
	}
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	time.Local = time.FixedZone("CST", 8*60*60)
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 64)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	// Complete pipeline-go v1.4.3 fn_datetime_test.go::TestDateTime table.
	vectors := []datetimeVector{
		{"ANSIC-s", "s", "ANSIC", "", "Tue Nov 30 14:25:18 2021", "1638253518", ""},
		{"ANSIC-ms", "ms", "ANSIC", "", "Tue Nov 30 14:25:18 2021", "1638253518000", ""},
		{"UnixDate-s", "s", "UnixDate", "", "Tue Nov 30 14:25:18 CST 2021", "1638253518", ""},
		{"UnixDate-ms", "ms", "UnixDate", "", "Tue Nov 30 14:25:18 CST 2021", "1638253518999", ""},
		{"RubyDate-ms", "ms", "RubyDate", "", "Tue Nov 30 14:25:18 +0800 2021", "1638253518999", ""},
		{"RubyDate-s", "s", "RubyDate", "", "Tue Nov 30 14:25:18 +0800 2021", "1638253518", ""},
		{"RFC822-ms", "ms", "RFC822", "", "30 Nov 21 14:25 CST", "1638253518999", ""},
		{"RFC822-s", "s", "RFC822", "", "30 Nov 21 14:25 CST", "1638253518", ""},
		{"RFC822Z-ms", "ms", "RFC822Z", "", "30 Nov 21 14:25 +0800", "1638253518999", ""},
		{"RFC822Z-s", "s", "RFC822Z", "", "30 Nov 21 14:25 +0800", "1638253518", ""},
		{"RFC850-ms", "ms", "RFC850", "", "Tuesday, 30-Nov-21 14:25:18 CST", "1638253518999", ""},
		{"RFC850-s", "s", "RFC850", "", "Tuesday, 30-Nov-21 14:25:18 CST", "1638253518", ""},
		{"RFC1123-ms", "ms", "RFC1123", "", "Tue, 30 Nov 2021 14:25:18 CST", "1638253518999", ""},
		{"RFC1123-s", "s", "RFC1123", "", "Tue, 30 Nov 2021 14:25:18 CST", "1638253518", ""},
		{"RFC1123Z-ms", "ms", "RFC1123Z", "", "Tue, 30 Nov 2021 14:25:18 +0800", "1638253518999", ""},
		{"RFC1123Z-s", "s", "RFC1123Z", "", "Tue, 30 Nov 2021 14:25:18 +0800", "1638253518", ""},
		{"RFC3339-s", "s", "RFC3339", "", "2021-01-18T17:03:25+08:00", "1610960605", ""},
		{"RFC3339-ms", "ms", "RFC3339", "", "2021-01-18T17:03:25+08:00", "1610960605000", ""},
		{"RFC3339Nano-s", "s", "RFC3339Nano", "", "2021-01-18T17:03:25+08:00", "1610960605", ""},
		{"RFC3339Nano-ms", "ms", "RFC3339Nano", "", "2021-01-18T17:03:25.001+08:00", "1610960605001", ""},
		{"Kitchen-ms", "ms", "Kitchen", "", "5:03PM", "1610960605001", ""},
		{"Kitchen-s", "s", "Kitchen", "", "5:03PM", "1610960605", ""},
		{"udef-ms", "ms", "%Y-%m-%d", "", "2021-01-18", "1610960605000", ""},
		{"udef-us", "us", "%Y-%m-%d %H:%M:%S", "", "2021-01-18 17:03:25", "1610960605000000", ""},
		{"udef-ns", "ns", "%Y-%m-%d %H:%M:%S", "", "2021-01-18 17:03:25", "1610960605000000000", ""},
		{"udef-Tokyo", "ns", "%Y-%m-%d %H:%M:%S", "Asia/Tokyo", "2021-01-18 18:03:25", "1610960605000000000", `datetime(value,"ns","%Y-%m-%d %H:%M:%S",tz="Asia/Tokyo"); add_key(after,true)`},
		{"udef-UTC", "ns", "%Y-%m-%d %H:%M:%S", "UTC", "2021-01-18 09:03:25", int64(1610960605000000000), `datetime(value,"ns",fmt="%Y-%m-%d %H:%M:%S",tz="UTC"); add_key(after,true)`},
		// pipeline-go recognizes only ASCII "us"; the micro sign is an unknown
		// precision and therefore leaves this already-nanosecond value unchanged.
		{"micro-sign-is-ns", "µs", "%Y-%m-%d %H:%M:%S", "UTC", "2021-01-18 09:03:25", int64(1610960605000000000), ""},
		{"unknown-is-ns", "unknown", "%Y-%m-%d %H:%M:%S", "UTC", "2021-01-18 09:03:25", int64(1610960605000000000), ""},
		{"invalid-string-is-zero", "s", "%Y-%m-%d %H:%M:%S", "UTC", "1970-01-01 00:00:00", "invalid", ""},
	}
	for _, size := range []int{1, 2, 4, 8, 10, 128} {
		t.Run("batch-"+itoa(size), func(t *testing.T) {
			for _, vector := range vectors {
				source := vector.source
				if source == "" {
					source = "datetime(value," + strconv.Quote(vector.precision) + "," + strconv.Quote(vector.format)
					if vector.timezone != "" {
						source += "," + strconv.Quote(vector.timezone)
					}
					source += "); add_key(after,true)"
				}
				check := runner.Check(source)
				if check.Route != RouteJITNative || check.Capabilities.ExecutionMode != "native" {
					t.Fatalf("%s route=%+v", vector.name, check)
				}
				points := make([]Point, size)
				for i := range points {
					points[i] = Point{Version: 1, Category: "logging", Measurement: vector.name,
						Fields: map[string]any{"value": vector.input, "sequence": int64(i)}}
				}
				assertDatetimeBatch(t, runner, source, vector, points)
			}
		})
	}

	upstreamScripts := []struct {
		name, source, message, want string
	}{
		{
			"upstream-composition-ANSIC",
			`json(_,a.timestamp); datetime(a.timestamp,"s","ANSIC"); add_key(after,true)`,
			`{"a":{"timestamp":"1638253518","second":2},"age":47}`,
			"Tue Nov 30 14:25:18 2021",
		},
		{
			"upstream-composition-named-UTC",
			`json(_,a.timestamp); datetime(a.timestamp,"ns",fmt="%Y-%m-%d %H:%M:%S",tz="UTC"); add_key(after,true)`,
			`{"a":{"timestamp":1610960605000000000,"second":2},"age":47}`,
			"2021-01-18 09:03:25",
		},
	}
	for _, tc := range upstreamScripts {
		check := runner.Check(tc.source)
		if check.Route != RouteJITNative || check.Capabilities.ExecutionMode != "native" {
			t.Fatalf("%s route=%+v", tc.name, check)
		}
		points := make([]Point, 10)
		for i := range points {
			points[i] = Point{Version: 1, Category: "logging", Measurement: tc.name,
				Fields: map[string]any{"message": tc.message, "sequence": int64(i)}}
		}
		input, err := EncodeFlatPoints(points)
		if err != nil {
			t.Fatal(err)
		}
		batch, err := runner.Process(tc.source, input)
		if err != nil || len(batch.Records) != len(points) {
			t.Fatalf("%s records=%d error=%v", tc.name, len(batch.Records), err)
		}
		for i, record := range batch.Records {
			if record.Status != TerminalOK {
				t.Fatalf("%s record %d terminal=%d error=%s", tc.name, i, record.Status, record.Error)
			}
			applyHostCompatRecord(t, batch, i, &points[i])
			if points[i].Fields["a.timestamp"] != tc.want || points[i].Fields["after"] != true ||
				points[i].Fields["sequence"] != int64(i) {
				t.Fatalf("%s record %d output=%#v", tc.name, i, points[i])
			}
		}
	}

	const missingSource = `datetime(value,"s","RFC3339"); add_key(after,true)`
	missing := []Point{{Version: 1, Category: "logging", Measurement: "missing", Fields: map[string]any{"sequence": int64(0)}}}
	assertDatetimeBatch(t, runner, missingSource, datetimeVector{name: "missing"}, missing)

	const invalidZone = `datetime(value,"s","RFC3339","Invalid/Zone"); add_key(after,true)`
	check := runner.Check(invalidZone)
	if check.Route != RouteJITNative {
		t.Fatalf("invalid timezone route=%+v", check)
	}
	points := []Point{{Version: 1, Category: "logging", Measurement: "invalid-zone", Fields: map[string]any{"value": int64(0), "sequence": int64(0)}}}
	input, err := EncodeFlatPoints(points)
	if err != nil {
		t.Fatal(err)
	}
	batch, err := runner.Process(invalidZone, input)
	if err != nil || len(batch.Records) != 1 || batch.Records[0].Status != TerminalError {
		t.Fatalf("invalid timezone batch=%+v error=%v", batch, err)
	}
	applyHostCompatRecord(t, batch, 0, &points[0])
	if _, ok := points[0].Fields["after"]; ok || points[0].Fields["value"] != int64(0) {
		t.Fatalf("invalid timezone committed wrong prefix: %#v", points[0])
	}
}

func assertDatetimeBatch(t *testing.T, runner *Runner, source string, vector datetimeVector, points []Point) {
	t.Helper()
	input, err := EncodeFlatPoints(points)
	if err != nil {
		t.Fatal(err)
	}
	batch, err := runner.Process(source, input)
	if err != nil || len(batch.Records) != len(points) {
		t.Fatalf("%s records=%d error=%v", vector.name, len(batch.Records), err)
	}
	for i, record := range batch.Records {
		if record.Status != TerminalOK {
			t.Fatalf("%s record %d terminal=%d error=%s", vector.name, i, record.Status, record.Error)
		}
		applyHostCompatRecord(t, batch, i, &points[i])
		if vector.want != "" && points[i].Fields["value"] != vector.want {
			t.Fatalf("%s record %d value=%#v want=%q", vector.name, i, points[i].Fields["value"], vector.want)
		}
		if vector.want == "" {
			if _, ok := points[i].Fields["value"]; ok {
				t.Fatalf("%s record %d created missing value: %#v", vector.name, i, points[i])
			}
		}
		if points[i].Fields["after"] != true || points[i].Fields["sequence"] != int64(i) {
			t.Fatalf("%s record %d lost input/continuation: %#v", vector.name, i, points[i])
		}
	}
}
