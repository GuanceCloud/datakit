// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package metrics

import (
	"encoding/json"
	"os"
	"testing"
)

func TestResolveMetricsJSONFile(t *testing.T) {
	f, err := os.Open("testdata/metrics.json")
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	metering, err := parseMetricsJSONFile(f)
	if err != nil {
		t.Fatal(err)
	}

	for name, number := range metering {
		t.Logf("[%s]: [%s]", name, number)
	}
}

func TestResolveStartTime(t *testing.T) {
	n1 := json.Number(`123`)
	n2 := json.Number(`3e6`)
	n3 := json.Number(`3.1415`)
	n4 := json.Number(`1e-2`)
	n5 := json.Number(`1e9`)

	x := resolveJSONNumber(n1)
	t.Logf("%T, %v", x, x)

	x = resolveJSONNumber(n2)
	t.Logf("%T, %v", x, x)

	x = resolveJSONNumber(n3)
	t.Logf("%T, %v", x, x)

	x = resolveJSONNumber(n4)
	t.Logf("%T, %v", x, x)

	x = resolveJSONNumber(n5)
	t.Logf("%T, %v", x, x)
}

func TestPprofSummary(t *testing.T) {
	/*
		   pprof_test.go:64: metric name: samples, value: 607, unit: count
		   pprof_test.go:64: metric name: cpu, value: 6070000000, unit: nanoseconds
		f, err := os.Open("testdata/cpu.pprof")

		    pprof_test.go:67: metric name: contentions, value: 1089, unit: count
		   pprof_test.go:67: metric name: delay, value: 1136108787243, unit: nanoseconds
		f, err := os.Open("testdata/delta-block.pprof")

		    pprof_test.go:70: metric name: alloc_objects, value: 535, unit: count
		   pprof_test.go:70: metric name: alloc_space, value: 16699978, unit: bytes
		   pprof_test.go:70: metric name: inuse_objects, value: 55422, unit: count
		   pprof_test.go:70: metric name: inuse_space, value: 18585974, unit: bytes
		f, err := os.Open("testdata/delta-heap.pprof")

		    pprof_test.go:77: metric name: contentions, value: 570, unit: count
		   pprof_test.go:77: metric name: delay, value: 1603409, unit: nanoseconds
		f, err := os.Open("testdata/delta-mutex.pprof")

		    pprof_test.go:81: metric name: goroutines, value: 25, unit: count
	*/

	/**
	  pprof_test.go:86: metric name: cpu-time, value: 7990188678, unit: nanoseconds
	  pprof_test.go:86: metric name: wall-time, value: 132095573542, unit: nanoseconds
	  pprof_test.go:86: metric name: exception-samples, value: 0, unit: count
	  pprof_test.go:86: metric name: lock-acquire, value: 0, unit: count
	  pprof_test.go:86: metric name: lock-acquire-wait, value: 0, unit: nanoseconds
	  pprof_test.go:86: metric name: lock-release, value: 0, unit: count
	  pprof_test.go:86: metric name: alloc-samples, value: 30720, unit: count
	  pprof_test.go:86: metric name: cpu-samples, value: 10661, unit: count
	  pprof_test.go:86: metric name: alloc-space, value: 151819986, unit: bytes
	  pprof_test.go:86: metric name: heap-space, value: 24367668, unit: bytes
	  pprof_test.go:86: metric name: lock-release-hold, value: 0, unit: nanoseconds
	*/

	f, err := os.Open("testdata/python.pprof")
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	summaries, err := pprofSummary(f)
	if err != nil {
		t.Fatal(err)
	}

	for metricType, quantity := range summaries {
		t.Logf("metric name: %s, value: %d, unit: %s", metricType, quantity.Value, quantity.Unit)
	}
}
