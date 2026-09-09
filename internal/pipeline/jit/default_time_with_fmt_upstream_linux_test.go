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
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"
)

type defaultTimeWithFmtVector struct {
	name, value, layout, timezone string
	wantUnchanged                 bool
	source                        string
}

func TestDefaultTimeWithFmtPipelineGoUpstreamAndBoundaryCorpus(t *testing.T) {
	const childKey = "DATAKIT_JIT_DEFAULT_TIME_FMT_CHILD"
	if os.Getenv(childKey) != "1" {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^"+t.Name()+"$", "-test.count=1")
		cmd.Env = append(os.Environ(), childKey+"=1", "TZ=Etc/GMT-8")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("default_time_with_fmt child: %v\n%s", err, output)
		}
		return
	}
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	time.Local = time.FixedZone("CST", 8*60*60)
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 16)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	// The first five vectors are the complete input corpus from pipeline-go
	// v1.4.3 fn_default_time_with_fmt_test.go::TestDefaultTimeWithFmt.
	vectors := []defaultTimeWithFmtVector{
		{name: "upstream-offset-jst", value: "02/Dec/2021:12:55:34 +0900", layout: "02/Jan/2006:15:04:05 -0700", timezone: "Asia/Tokyo"},
		{name: "upstream-offset-cst", value: "02/Dec/2021:11:55:34 +0800", layout: "02/Jan/2006:15:04:05 -0700", timezone: "Asia/Tokyo"},
		{name: "upstream-shanghai", value: "02/Dec/2021:11:55:34", layout: "02/Jan/2006:15:04:05", timezone: "Asia/Shanghai"},
		{name: "upstream-local", value: "02/Dec/2021:11:55:34", layout: "02/Jan/2006:15:04:05"},
		{name: "upstream-tokyo", value: "02/Dec/2021:12:55:34", layout: "02/Jan/2006:15:04:05", timezone: "Asia/Tokyo"},
		{name: "utc-nanoseconds", value: "2024-01-02T03:04:05.123456789Z", layout: "2006-01-02T15:04:05.999999999Z07:00", timezone: "UTC"},
		{name: "date-only", value: "20211202", layout: "20060102", timezone: "UTC"},
		{name: "dst-overlap", value: "2021-11-07 01:30:00", layout: "2006-01-02 15:04:05", timezone: "America/New_York"},
		{name: "dst-gap", value: "2021-03-14 02:30:00", layout: "2006-01-02 15:04:05", timezone: "America/New_York"},
		{name: "invalid-value", value: "not-a-time", layout: "2006-01-02", timezone: "UTC", wantUnchanged: true},
		{name: "invalid-zone-with-offset", value: "2024-01-02T03:04:05Z", layout: "2006-01-02T15:04:05Z07:00", timezone: "Invalid/Zone", wantUnchanged: true},
		{name: "extra-argument-ignored", value: "20211202", layout: "20060102", timezone: "UTC", source: `default_time_with_fmt(time,"20060102","UTC",123); add_key(after,true)`},
	}
	for _, size := range []int{1, 2, 4, 8, 10, 128} {
		t.Run("batch-"+itoa(size), func(t *testing.T) {
			for _, vector := range vectors {
				source := vector.source
				if source == "" {
					source = "default_time_with_fmt(time," + strconv.Quote(vector.layout)
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
						Fields: map[string]any{"time": vector.value, "sequence": int64(i)}}
				}
				assertDefaultTimeWithFmtBatch(t, runner, source, vector, points)
			}
		})
	}

	const missingSource = `default_time_with_fmt(time,"20060102","UTC"); add_key(after,true)`
	points := []Point{{Version: 1, Category: "logging", Measurement: "missing", Fields: map[string]any{"sequence": int64(0)}}}
	assertDefaultTimeWithFmtBatch(t, runner, missingSource,
		defaultTimeWithFmtVector{name: "missing", wantUnchanged: true}, points)
}

func assertDefaultTimeWithFmtBatch(t *testing.T, runner *Runner, source string, vector defaultTimeWithFmtVector, points []Point) {
	t.Helper()
	var want any
	if len(points) != 0 {
		want = points[0].Fields["time"]
	}
	if !vector.wantUnchanged && vector.value != "" {
		location := time.Local
		var err error
		if vector.timezone != "" {
			location, err = time.LoadLocation(vector.timezone)
		}
		if err == nil {
			var parsed time.Time
			parsed, err = time.ParseInLocation(vector.layout, vector.value, location)
			if err == nil {
				want = parsed.UnixNano()
			}
		}
		if err != nil {
			t.Fatalf("invalid oracle vector %s: %v", vector.name, err)
		}
	}
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
		if got := points[i].Fields["time"]; fmt.Sprintf("%#v", got) != fmt.Sprintf("%#v", want) {
			t.Fatalf("%s record %d time=%#v want=%#v", vector.name, i, got, want)
		}
		if points[i].Fields["after"] != true || points[i].Fields["sequence"] != int64(i) {
			t.Fatalf("%s record %d lost continuation/input: %#v", vector.name, i, points[i])
		}
	}
}
