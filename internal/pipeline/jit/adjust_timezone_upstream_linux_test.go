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
	"fmt"
	"os"
	"testing"
	"time"
)

func pipelineGoDetectTimezone(logTS, nowTS int64) int64 {
	const hour = int64(time.Hour)
	const tolerance = int64(2 * time.Minute)
	logTS -= (logTS/hour - nowTS/hour) * hour
	if logTS-nowTS > tolerance {
		logTS -= hour
	} else if logTS-nowTS <= -hour+tolerance {
		logTS += hour
	}
	return logTS
}

func TestAdjustTimezonePipelineGoUpstreamAndBoundaryCorpus(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	host, recorder := newObservedPipelineGoHost()
	runner, err := NewRunnerWithHost(path, "pipeline-go-1.4.3-datakit", 16, host)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	cases := []struct {
		name   string
		delta  time.Duration
		source string
	}{
		{"upstream-hour-minus2-minute-minus2", -2*time.Hour - 2*time.Minute - 2*time.Second - 15*time.Nanosecond, `adjust_timezone(value); add_key(after,true)`},
		{"upstream-hour-minus2-same-minute", -2*time.Hour + 5*time.Nanosecond, `adjust_timezone(value); add_key(after,true)`},
		{"upstream-positive-tolerance-edge", -2*time.Hour + 2*time.Minute + time.Second, `adjust_timezone(value); add_key(after,true)`},
		{"upstream-negative-hour-edge", -2*time.Hour + 2*time.Minute, `adjust_timezone(value); add_key(after,true)`},
		{"attribute-path", -8*time.Hour + 10*time.Minute, `adjust_timezone(a.value); add_key(after,true)`},
	}
	for _, size := range []int{1, 2, 4, 8, 10, 128} {
		t.Run("batch-"+itoa(size), func(t *testing.T) {
			for _, tc := range cases {
				check := runner.Check(tc.source)
				if check.Route != RouteJITNative || check.Capabilities.ExecutionMode != "native" {
					t.Fatalf("%s route=%+v", tc.name, check)
				}
				before := time.Now().UnixNano()
				inputTimestamp := before + int64(tc.delta)
				key := "value"
				if tc.name == "attribute-path" {
					key = "a.value"
				}
				points := make([]Point, size)
				for i := range points {
					points[i] = Point{Version: 1, Category: "logging", Measurement: tc.name,
						Fields: map[string]any{key: inputTimestamp, "sequence": int64(i)}}
				}
				input, err := EncodeFlatPoints(points)
				if err != nil {
					t.Fatal(err)
				}
				batch, err := runner.Process(tc.source, input)
				after := time.Now().UnixNano()
				if err != nil || len(batch.Records) != size {
					t.Fatalf("%s records=%d error=%v", tc.name, len(batch.Records), err)
				}
				for i, record := range batch.Records {
					if record.Status != TerminalOK {
						t.Fatalf("%s record %d terminal=%d error=%s", tc.name, i, record.Status, record.Error)
					}
					applyHostCompatRecord(t, batch, i, &points[i])
					got := points[i].Fields[key]
					wantBefore := pipelineGoDetectTimezone(inputTimestamp, before)
					wantAfter := pipelineGoDetectTimezone(inputTimestamp, after)
					if got != wantBefore && got != wantAfter {
						t.Fatalf("%s record %d value=%#v want %d or %d", tc.name, i, got, wantBefore, wantAfter)
					}
					if points[i].Fields["after"] != true || points[i].Fields["sequence"] != int64(i) {
						t.Fatalf("%s record %d lost input/continuation: %#v", tc.name, i, points[i])
					}
				}
			}
		})
	}

	// Retain the upstream composition, not merely the three component
	// functions in isolation. ANSIC has no offset, so default_time parses the
	// UTC-rendered wall clock in Local before adjust_timezone rebases its hour.
	t.Run("upstream-json-default-time-composition", func(t *testing.T) {
		const source = `json(_,time); default_time(time); adjust_timezone(time); add_key(after,true)`
		recorder.reset()
		check := runner.Check(source)
		if check.Route != RouteJITNative || check.Capabilities.ExecutionMode != "native" ||
			len(check.Capabilities.HostCalls) != 0 {
			t.Fatalf("route=%+v", check)
		}
		wall := time.Now().Add(10 * time.Minute).UTC().Truncate(time.Second)
		formatted := wall.Format(time.ANSIC)
		parsed, err := time.ParseInLocation(time.ANSIC, formatted, time.Local)
		if err != nil {
			t.Fatal(err)
		}
		point := Point{Version: 1, Category: "logging", Measurement: "upstream-composition",
			Fields: map[string]any{"message": fmt.Sprintf(`{"time":%q}`, formatted), "sequence": int64(0)}}
		input, err := EncodeFlatPoints([]Point{point})
		if err != nil {
			t.Fatal(err)
		}
		before := time.Now().UnixNano()
		batch, err := runner.Process(source, input)
		after := time.Now().UnixNano()
		if err != nil || len(batch.Records) != 1 || batch.Records[0].Status != TerminalOK {
			t.Fatalf("batch=%+v error=%v", batch, err)
		}
		applyHostCompatRecord(t, batch, 0, &point)
		got := point.Fields["time"]
		wantBefore := pipelineGoDetectTimezone(parsed.UnixNano(), before)
		wantAfter := pipelineGoDetectTimezone(parsed.UnixNano(), after)
		if got != wantBefore && got != wantAfter {
			t.Fatalf("time=%#v want %d or %d", got, wantBefore, wantAfter)
		}
		if point.Fields["after"] != true || point.Fields["sequence"] != int64(0) {
			t.Fatalf("lost continuation/input: %#v", point)
		}
		requireHostCompatCalls(t, recorder, hostCompatTimestampOp)
	})

	for _, tc := range []struct {
		name   string
		value  any
		status TerminalStatus
	}{
		{"missing", nil, TerminalOK},
		{"string-type-error", "123", TerminalError},
		{"float-type-error", float64(123), TerminalError},
		{"bool-type-error", true, TerminalError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const source = `adjust_timezone(value); add_key(after,true)`
			check := runner.Check(source)
			if check.Route != RouteJITNative {
				t.Fatalf("route=%+v", check)
			}
			fields := map[string]any{"sequence": int64(0)}
			if tc.value != nil {
				fields["value"] = tc.value
			}
			point := Point{Version: 1, Category: "logging", Measurement: tc.name, Fields: fields}
			input, err := EncodeFlatPoints([]Point{point})
			if err != nil {
				t.Fatal(err)
			}
			batch, err := runner.Process(source, input)
			if err != nil || len(batch.Records) != 1 || batch.Records[0].Status != tc.status {
				t.Fatalf("batch=%+v error=%v", batch, err)
			}
			applyHostCompatRecord(t, batch, 0, &point)
			if tc.status == TerminalOK {
				if point.Fields["after"] != true {
					t.Fatalf("missing did not continue: %#v", point)
				}
			} else if _, ok := point.Fields["after"]; ok {
				t.Fatalf("type error continued: %#v", point)
			}
		})
	}

	// Pipeline-go GetKey reads a same-named local before the Point and writes
	// the adjusted local value back to the Point destination.
	const localSource = `value=timestamp()-28800000000000; adjust_timezone(value); add_key(after,true)`
	check := runner.Check(localSource)
	if check.Route != RouteJITNative {
		t.Fatalf("local shadow route=%+v", check)
	}
	point := Point{Version: 1, Category: "logging", Measurement: "local-shadow",
		Fields: map[string]any{"value": int64(7), "sequence": int64(0)}}
	input, err := EncodeFlatPoints([]Point{point})
	if err != nil {
		t.Fatal(err)
	}
	batch, err := runner.Process(localSource, input)
	if err != nil || len(batch.Records) != 1 || batch.Records[0].Status != TerminalOK {
		t.Fatalf("local shadow batch=%+v error=%v", batch, err)
	}
	applyHostCompatRecord(t, batch, 0, &point)
	if point.Fields["value"] == int64(7) || point.Fields["after"] != true {
		t.Fatalf("local shadow did not write local result: %#v", point)
	}
}
