// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package trace

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
)

func TestCloseResource(t *testing.T) {
	testcases := DatakitTraces{
		randDatakitTrace(t, 10, randResource("Allen.123"), randSource("ddtrace")),
		randDatakitTrace(t, 10, randService("game"), randResource("Bravo333"), randSource("ddtrace")),
		randDatakitTrace(t, 10, randService("hesi"), randResource("GET /nacos/v1/ns/instance/list"), randSource("ddtrace")),
		randDatakitTrace(t, 10, randService("logout"), randResource("Clear666"), randSource("ddtrace")),
	}
	expected := []func(trace DatakitTrace) bool{
		func(trace DatakitTrace) bool { return trace == nil },
		func(trace DatakitTrace) bool { return trace == nil },
		func(trace DatakitTrace) bool { return trace == nil },
		func(trace DatakitTrace) bool { return trace != nil },
	}

	closer := &CloseResource{}
	closer.UpdateIgnResList(map[string][]string{
		"*": {"Allen.*", ".*333", "GET /nacos/v1/.*"},
	})

	log := logger.DefaultSLogger("filters-test")
	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() { // nolint:govet,staticcheck
			defer wg.Done()

			for i := range testcases {
				parentialize(testcases[i])

				trace, _ := closer.Close(log, testcases[i])
				if !expected[i](trace) {
					t.Errorf("close resource %s failed trace:%v", testcases[i][0].GetFiledToString(FieldResource), trace)
					t.FailNow() // nolint:govet,staticcheck
				}
			}
		}()
	}
	wg.Wait()
}

func TestRespectUserRule(t *testing.T) {
	testcases := make(DatakitTraces, 100)
	for i := 0; i < 100; i++ {
		trace := randDatakitTrace(t, 10, randService(_services...), randResource(_resources...), randPriority(_all_priorities...))
		parentialize(trace)
		testcases[i] = trace
	}

	var (
		keep, auto DatakitTraces
		log        = logger.DefaultSLogger("filters-test")
	)
	for i := range testcases {
		if trace, ok := RespectUserRule(log, testcases[i]); ok {
			if trace != nil {
				keep = append(keep, trace)
			}
		} else {
			auto = append(auto, trace)
		}
	}
	t.Logf("keep len=%d", len(keep))
	for i := range keep {
		for j := range keep[i] {
			priority := keep[i][j].GetFiledToInt(FieldPriority)
			if priority != PriorityUserKeep && priority != PriorityRuleSamplerKeep {
				t.Errorf("unexpected priority %d found", priority)
				t.FailNow()
			}
			break //nolint
		}
	}
	t.Logf("auto len=%d", len(auto))
	for i := range auto {
		for j := range auto[i] {
			priority := auto[i][j].GetFiledToInt(FieldPriority)
			if priority != PriorityAutoKeep && priority != PriorityAutoReject {
				t.Errorf("unexpected priority %d found", priority)
				t.FailNow()
			}
			break //nolint
		}
	}
}

func TestOmitStatusCode(t *testing.T) {
	testcases := make(DatakitTraces, 100)
	for i := 0; i < 100; i++ {
		testcases[i] = randDatakitTrace(t, 10, randService(_services...), randResource(_resources...), randHTTPStatusCode(_http_status_codes...))
	}

	var (
		afterOmitStatusCode DatakitTraces
		log                 = logger.DefaultSLogger("filters-test")
	)
	for i := range testcases {
		if trace, ok := OmitHTTPStatusCodeFilterWrapper([]string{"404", "500", "307"})(log, testcases[i]); !ok {
			afterOmitStatusCode = append(afterOmitStatusCode, trace)
		}
	}

	for i := range afterOmitStatusCode {
		for j := range afterOmitStatusCode[i] {
			switch afterOmitStatusCode[i][j].GetFiledToString(TagHttpStatusCode) {
			case "404", "500", "307":
				t.Errorf("status code %s should be omitted", afterOmitStatusCode[i][j].GetFiledToString(TagHttpStatusCode))
				t.FailNow()
			}
		}
	}
}

func TestPenetrateError(t *testing.T) {
	testcases := make(DatakitTraces, 100)
	for i := 0; i < 100; i++ {
		testcases[i] = randDatakitTrace(t, 10, randService(_services...), randResource(_resources...))
	}

	var (
		afterErrPenetrate DatakitTraces
		log               = logger.DefaultSLogger("filters-test")
	)
	for i := range testcases {
		if t, ok := PenetrateErrorTracing(log, testcases[i]); ok {
			afterErrPenetrate = append(afterErrPenetrate, t)
		}
	}

	for i := range afterErrPenetrate {
		iserr := false
	FOUND_ERR:
		for j := range afterErrPenetrate[i] {
			switch afterErrPenetrate[i][j].GetTag(TagSpanStatus) {
			case StatusErr, StatusCritical:
				iserr = true
				break FOUND_ERR
			}
		}
		if !iserr {
			t.Error("error status not found")
			t.FailNow()
		}
	}
}

func TestKeepRareResource(t *testing.T) {
	var traces DatakitTraces
	for i := 0; i < 10; i++ {
		trace := randDatakitTrace(t, 10, randService("test-rare-resourc"), randResource("kept"), randSource("ddtrace"))
		parentialize(trace)
		traces = append(traces, trace)
	}

	keep := &KeepRareResource{}
	keep.UpdateStatus(true, 10*time.Millisecond)

	log := logger.DefaultSLogger("filters-test")
	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() { // nolint:govet,staticcheck
			defer wg.Done()

			var kept DatakitTraces
			for i := range traces {
				time.Sleep(5 * time.Millisecond)
				if t, skip := keep.Keep(log, traces[i]); skip {
					kept = append(kept, t)
				}
			}
			if len(kept) >= len(traces) {
				t.Errorf("wrong length kept send: %d kept: %d", len(traces), len(kept))
				t.FailNow() // nolint:govet,staticcheck
			}
		}()
	}
	wg.Wait()

	var kept DatakitTraces
	for i := range traces {
		time.Sleep(15 * time.Millisecond)
		if t, skip := keep.Keep(log, traces[i]); skip {
			kept = append(kept, t)
		}
	}
	if len(kept) != len(traces) {
		t.Errorf("wrong length kept sec send: %d kept: %d", len(traces), len(kept))
		t.FailNow()
	}
}

func TestSampler(t *testing.T) {
	var origin DatakitTraces
	for i := 0; i < 1000; i++ {
		dktrace := randDatakitTrace(t, 1, randService(_services...), randResource(_resources...), randPriority(PriorityAutoKeep))
		parentialize(dktrace)
		origin = append(origin, dktrace)
	}

	sampler := &Sampler{SamplingRateGlobal: 0.15}

	log := logger.DefaultSLogger("filters-test")
	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() { // nolint:govet,staticcheck
			defer wg.Done()

			var sampled DatakitTraces
			for i := range origin {
				if t, _ := sampler.Sample(log, origin[i]); t != nil {
					sampled = append(sampled, t)
				}
			}
			fmt.Printf("origin traces count: %d sampled traces count: %d\n", len(origin), len(sampled))
		}()
	}
	wg.Wait()
}
