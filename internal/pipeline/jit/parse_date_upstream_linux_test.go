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
	"time"
)

type parseDateVector struct {
	name, source string
	fields       map[string]any
	want         func(time.Time) int64
	wantError    bool
}

func TestParseDatePipelineGoUpstreamAndBoundaryCorpus(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 32)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	utc8 := time.FixedZone("UTC+8", 8*60*60)
	current := func(hour, minute, second, nanos int, zone *time.Location) func(time.Time) int64 {
		return func(now time.Time) int64 {
			return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, second, nanos, zone).UnixNano()
		}
	}
	fixed := func(year int, month time.Month, day, hour, minute, second, nanos int, zone *time.Location) func(time.Time) int64 {
		return func(time.Time) int64 {
			return time.Date(year, month, day, hour, minute, second, nanos, zone).UnixNano()
		}
	}
	clockFields := func(values ...any) map[string]any {
		fields := map[string]any{}
		for i, key := range []string{"year", "month", "day", "hour", "min", "sec", "msec", "usec", "nsec", "tz"} {
			if i < len(values) && values[i] != nil {
				fields[key] = values[i]
			}
		}
		return fields
	}
	vectors := []parseDateVector{
		{"upstream-positional-ms", `parse_date("result","","","",hour,min,sec,msec,"","","+8"); add_key(after,true)`, clockFields(nil, nil, nil, "16", "40", "03", "290"), current(16, 40, 3, 290_000_000, utc8), false},
		{"upstream-named-ms", `parse_date(key="result",hh=hour,mm=min,ss=sec,ms=msec,zone="+8"); add_key(after,true)`, clockFields(nil, nil, nil, "16", "40", "03", "290"), current(16, 40, 3, 290_000_000, utc8), false},
		{"upstream-full-numeric", `parse_date(key="result",yy=year,MM=month,dd=day,hh=hour,mm=min,ss=sec,ms=msec,zone="+8"); add_key(after,true)`, clockFields("2020", "12", "12", "16", "40", "03", "290"), fixed(2020, 12, 12, 16, 40, 3, 290_000_000, utc8), false},
		{"upstream-month-name", `parse_date(key="result",yy=year,MM=month,dd=day,hh=hour,mm=min,ss=sec,zone=tz); add_key(after,true)`, clockFields("2021", "Sep", "6", "16", "40", "03", nil, nil, nil, "CST"), fixed(2021, 9, 6, 16, 40, 3, 0, utc8), false},
		{"upstream-month-name-spacing-variant", `parse_date(key="result",yy=year,MM=month,dd=day,hh=hour,mm=min,ss=sec,zone=tz); add_key(after,true)`, clockFields("2021", "Sep", "6", "16", "40", "03", nil, nil, nil, "CST"), fixed(2021, 9, 6, 16, 40, 3, 0, utc8), false},
		{"upstream-partial-year-00", `parse_date(key="result",yy=year,MM=month,dd=day,hh=hour,mm=min,ss=sec,zone=tz); add_key(after,true)`, clockFields("00", "Sep", "6", "16", "40", "03", nil, nil, nil, "CST"), fixed(2000, 9, 6, 16, 40, 3, 0, utc8), false},
		{"upstream-partial-year-09", `parse_date(key="result",yy=year,MM=month,dd=day,hh=hour,mm=min,ss=sec,zone=tz); add_key(after,true)`, clockFields("09", "Sep", "6", "16", "40", "03", nil, nil, nil, "CST"), fixed(2009, 9, 6, 16, 40, 3, 0, utc8), false},
		{"upstream-us", `parse_date("result","","","",hour,min,sec,"",usec,"","+8"); add_key(after,true)`, clockFields(nil, nil, nil, "16", "40", "03", nil, "290290"), current(16, 40, 3, 290_290_000, utc8), false},
		{"upstream-ns", `parse_date("result","","","",hour,min,sec,"","",nsec,"+8"); add_key(after,true)`, clockFields(nil, nil, nil, "16", "40", "03", nil, nil, "290290330"), current(16, 40, 3, 290_290_330, utc8), false},
		{"upstream-ns-named", `parse_date(key="result",hh=hour,mm=min,ss=sec,ns=nsec,zone="CST"); add_key(after,true)`, clockFields(nil, nil, nil, "16", "40", "03", nil, nil, "290290330"), current(16, 40, 3, 290_290_330, utc8), false},
		{"upstream-missing-second", `parse_date(key="result",hh=hour,mm=min,zone="CST"); add_key(after,true)`, clockFields(nil, nil, nil, "16", "40"), current(16, 40, 0, 0, utc8), false},
		{"upstream-empty-zone-is-UTC", `parse_date("result","","","",hour,min,sec,msec,"","",""); add_key(after,true)`, clockFields(nil, nil, nil, "16", "40", "03", "290"), current(16, 40, 3, 290_000_000, time.UTC), false},
		{"upstream-invalid-hour-high", `parse_date(key="result",yy=year,MM=month,dd=day,hh=hour,mm=min,ss=sec,zone=tz); add_key(after,true)`, clockFields("2021", "Sep", "6", "25", "40", "03", nil, nil, nil, "CST"), nil, true},
		{"upstream-invalid-hour-negative", `parse_date(key="result",yy=year,MM=month,dd=day,hh=hour,mm=min,ss=sec,zone=tz); add_key(after,true)`, clockFields("2021", "Sep", "6", "-2", "40", "03", nil, nil, nil, "CST"), nil, true},
		{"upstream-invalid-minute-61", `parse_date(key="result",yy=year,MM=month,dd=day,hh=hour,mm=min,ss=sec,zone=tz); add_key(after,true)`, clockFields("2021", "Sep", "6", "12", "61", "03", nil, nil, nil, "CST"), nil, true},
		{"upstream-invalid-minute-60", `parse_date(key="result",yy=year,MM=month,dd=day,hh=hour,mm=min,ss=sec,zone=tz); add_key(after,true)`, clockFields("2021", "Sep", "6", "12", "60", "03", nil, nil, nil, "CST"), nil, true},
		{"upstream-invalid-second-high", `parse_date(key="result",yy=year,MM=month,dd=day,hh=hour,mm=min,ss=sec,zone=tz); add_key(after,true)`, clockFields("2021", "Sep", "6", "12", "11", "61", nil, nil, nil, "CST"), nil, true},
		{"upstream-invalid-second-negative", `parse_date(key="result",yy=year,MM=month,dd=day,hh=hour,mm=min,ss=sec,zone=tz); add_key(after,true)`, clockFields("2021", "Sep", "6", "12", "12", "-1", nil, nil, nil, "CST"), nil, true},
		{"boundary-normalized-Feb-31", `parse_date(key="result",yy="2021",MM="2",dd="31",zone="UTC"); add_key(after,true)`, nil, func(now time.Time) int64 {
			return time.Date(2021, 3, 3, now.Hour(), now.Minute(), 0, 0, time.UTC).UnixNano()
		}, false},
		{"boundary-invalid-month", `parse_date(key="result",yy="2021",MM="13",dd="1",zone="UTC"); add_key(after,true)`, nil, nil, true},
		{"boundary-invalid-zone", `parse_date(key="result",yy="2021",MM="1",dd="1",zone="Invalid/Zone"); add_key(after,true)`, nil, nil, true},
		{"boundary-non-string-component-defaults", `parse_date(key="result",yy=year,MM="1",dd="2",hh="3",zone="UTC"); add_key(after,true)`, map[string]any{"year": int64(1999)}, func(now time.Time) int64 {
			return time.Date(now.Year(), 1, 2, 3, now.Minute(), 0, 0, time.UTC).UnixNano()
		}, false},
	}

	for _, size := range []int{1, 2, 4, 8, 10, 128} {
		t.Run("batch-"+itoa(size), func(t *testing.T) {
			for _, vector := range vectors {
				check := runner.Check(vector.source)
				if check.Route != RouteJITNative || check.Capabilities.ExecutionMode != "native" {
					t.Fatalf("%s route=%+v", vector.name, check)
				}
				before := time.Now()
				points := make([]Point, size)
				for i := range points {
					fields := map[string]any{"sequence": int64(i)}
					for key, value := range vector.fields {
						fields[key] = value
					}
					points[i] = Point{Version: 1, Category: "logging", Measurement: vector.name, Fields: fields}
				}
				input, err := EncodeFlatPoints(points)
				if err != nil {
					t.Fatal(err)
				}
				batch, err := runner.Process(vector.source, input)
				after := time.Now()
				if err != nil || len(batch.Records) != size {
					t.Fatalf("%s records=%d error=%v", vector.name, len(batch.Records), err)
				}
				for i, record := range batch.Records {
					if vector.wantError {
						if record.Status != TerminalError {
							t.Fatalf("%s record %d terminal=%d error=%s", vector.name, i, record.Status, record.Error)
						}
					} else if record.Status != TerminalOK {
						t.Fatalf("%s record %d terminal=%d error=%s", vector.name, i, record.Status, record.Error)
					}
					applyHostCompatRecord(t, batch, i, &points[i])
					if vector.wantError {
						if _, ok := points[i].Fields["result"]; ok {
							t.Fatalf("%s record %d committed result: %#v", vector.name, i, points[i])
						}
						if _, ok := points[i].Fields["after"]; ok {
							t.Fatalf("%s record %d continued after error: %#v", vector.name, i, points[i])
						}
						continue
					}
					got := points[i].Fields["result"]
					wantBefore, wantAfter := vector.want(before), vector.want(after)
					if got != wantBefore && got != wantAfter {
						t.Fatalf("%s record %d result=%#v want %d or %d", vector.name, i, got, wantBefore, wantAfter)
					}
					if points[i].Fields["after"] != true || points[i].Fields["sequence"] != int64(i) {
						t.Fatalf("%s record %d lost input/continuation: %#v", vector.name, i, points[i])
					}
				}
			}
		})
	}

	const composed = `grok(_,"%{INT:year}-%{INT:month}-%{INT:day} %{INT:hour}:%{INT:min}:%{INT:sec}\\.%{INT:msec}"); parse_date(key="time",yy=year,MM=month,dd=day,hh=hour,mm=min,ss=sec,ms=msec,zone="+8"); add_key(after,true)`
	check := runner.Check(composed)
	if check.Route != RouteJITNative || check.Capabilities.ExecutionMode != "native" {
		t.Fatalf("composed success route=%+v", check)
	}
	points := make([]Point, 10)
	for i := range points {
		points[i] = Point{Version: 1, Category: "logging", Measurement: "parse-date-composed",
			Fields: map[string]any{"message": "2020-12-12 16:40:03.290", "sequence": int64(i)}}
	}
	input, err := EncodeFlatPoints(points)
	if err != nil {
		t.Fatal(err)
	}
	batch, err := runner.Process(composed, input)
	if err != nil || len(batch.Records) != len(points) {
		t.Fatalf("composed success records=%d error=%v", len(batch.Records), err)
	}
	wantComposed := time.Date(2020, 12, 12, 16, 40, 3, 290_000_000, utc8).UnixNano()
	for i, record := range batch.Records {
		if record.Status != TerminalOK {
			t.Fatalf("composed success record %d terminal=%d error=%s", i, record.Status, record.Error)
		}
		applyHostCompatRecord(t, batch, i, &points[i])
		if points[i].Fields["time"] != wantComposed || points[i].Fields["after"] != true {
			t.Fatalf("composed success record %d output=%#v", i, points[i])
		}
	}

	const composedError = `grok(_,"%{NOTSPACE:wd}\\s+%{NOTSPACE:month}\\s+%{INT:day}\\s+%{INT:hour}:%{INT:min}:%{INT:sec}\\s+%{NOTSPACE:tz}\\s+%{INT:year}"); parse_date(key="time",yy=year,MM=month,dd=day,hh=hour,mm=min,ss=sec,zone=tz); add_key(after,true)`
	check = runner.Check(composedError)
	if check.Route != RouteJITNative {
		t.Fatalf("composed error route=%+v", check)
	}
	point := Point{Version: 1, Category: "logging", Measurement: "parse-date-composed-error",
		Fields: map[string]any{"message": "Mon Sep  6 12:61:03 CST 2021", "sequence": int64(0)}}
	input, err = EncodeFlatPoints([]Point{point})
	if err != nil {
		t.Fatal(err)
	}
	batch, err = runner.Process(composedError, input)
	if err != nil || len(batch.Records) != 1 || batch.Records[0].Status != TerminalError {
		t.Fatalf("composed error batch=%+v error=%v", batch, err)
	}
	applyHostCompatRecord(t, batch, 0, &point)
	if _, ok := point.Fields["time"]; ok {
		t.Fatalf("composed error committed time: %#v", point)
	}
	if _, ok := point.Fields["after"]; ok {
		t.Fatalf("composed error continued: %#v", point)
	}
	for _, key := range []string{"year", "month", "day", "hour", "min", "sec", "tz"} {
		if _, ok := point.Fields[key]; !ok {
			t.Fatalf("composed error lost grok prefix %q: %#v", key, point)
		}
	}
}
