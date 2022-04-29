package trace

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestOmitStatusCode(t *testing.T) {
	testcases := make(DatakitTraces, 100)
	for i := 0; i < 100; i++ {
		testcases[i] = randDatakitTrace(t, 10)
	}

	var afterOmitStatusCode DatakitTraces
	for i := range testcases {
		if t, ok := OmitStatusCodeFilterWrapper([]string{"404", "500", "307"})(testcases[i]); !ok {
			afterOmitStatusCode = append(afterOmitStatusCode, t)
		}
	}

	for i := range afterOmitStatusCode {
		for j := range afterOmitStatusCode[i] {
			switch afterOmitStatusCode[i][j].HTTPStatusCode {
			case "404", "500", "307":
				t.Errorf("status code %s should be omitted", afterOmitStatusCode[i][j].HTTPStatusCode)
				t.FailNow()
			}
		}
	}
}

func TestPenetrateError(t *testing.T) {
	testcases := make(DatakitTraces, 100)
	for i := 0; i < 100; i++ {
		testcases[i] = randDatakitTrace(t, 10)
	}

	var afterErrPenetrate DatakitTraces
	for i := range testcases {
		if t, ok := PenetrateErrorTracing(testcases[i]); ok {
			afterErrPenetrate = append(afterErrPenetrate, t)
		}
	}

	for i := range afterErrPenetrate {
		iserr := false
	FOUND_ERR:
		for j := range afterErrPenetrate[i] {
			switch afterErrPenetrate[i][j].Status {
			case STATUS_ERR, STATUS_CRITICAL:
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

func TestCloseResource(t *testing.T) {
	testcases := DatakitTraces{
		randDatakitTraceByService(t, 10, "login", "Allen123", "ddtrace"),
		randDatakitTraceByService(t, 10, "game", "Bravo333", "ddtrace"),
		randDatakitTraceByService(t, 10, "logout", "Clear666", "ddtrace"),
	}
	expected := []func(trace DatakitTrace) bool{
		func(trace DatakitTrace) bool { return trace != nil },
		func(trace DatakitTrace) bool { return trace == nil },
		func(trace DatakitTrace) bool { return trace != nil },
	}

	closer := &CloseResource{}
	closer.UpdateIgnResList(map[string][]string{"game": {".*333"}})

	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() { // nolint:govet,staticcheck
			defer wg.Done()

			for i := range testcases {
				parentialize(testcases[i])

				trace, _ := closer.Close(testcases[i])
				if !expected[i](trace) {
					t.Errorf("close resource %s failed trace:%v", testcases[i][0].Resource, trace)
					t.FailNow() // nolint:govet,staticcheck
				}
			}
		}()
	}
	wg.Wait()
}

func TestKeepRareResource(t *testing.T) {
	var traces DatakitTraces
	for i := 0; i < 10; i++ {
		trace := randDatakitTraceByService(t, 10, "test-rare-resource", "kept", "ddtrace")
		parentialize(trace)
		traces = append(traces, trace)
	}

	keep := &KeepRareResource{}
	keep.UpdateStatus(true, 10*time.Millisecond)

	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() { // nolint:govet,staticcheck
			defer wg.Done()

			var kept DatakitTraces
			for i := range traces {
				time.Sleep(5 * time.Millisecond)
				if t, skip := keep.Keep(traces[i]); skip {
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
		if t, skip := keep.Keep(traces[i]); skip {
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
		dktrace := randDatakitTrace(t, 1)
		parentialize(dktrace)
		origin = append(origin, dktrace)
	}

	sampler := &Sampler{}
	sampler.UpdateArgs(PriorityAuto, 0.15)

	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() { // nolint:govet,staticcheck
			defer wg.Done()

			var sampled DatakitTraces
			for i := range origin {
				if t, _ := sampler.Sample(origin[i]); t != nil {
					sampled = append(sampled, t)
				}
			}
			fmt.Printf("origin traces count: %d sampled traces count: %d\n", len(origin), len(sampled))
		}()
	}
	wg.Wait()
}
