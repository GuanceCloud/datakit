package trace

import (
	"fmt"
	"regexp"
	"testing"
)

func TestSampler(t *testing.T) {
	var origin DatakitTraces
	for i := 0; i < 1000; i++ {
		dktrace := randDatakitTrace(t, 1)
		parentialize(dktrace)
		origin = append(origin, dktrace)
	}

	sampler := &Sampler{
		Priority:           PriorityAuto,
		SamplingRateGlobal: 0.15,
	}
	var sampled DatakitTraces
	for i := range origin {
		if t, _ := sampler.Sample(origin[i]); t != nil {
			sampled = append(sampled, t)
		}
	}

	fmt.Printf("origin traces count: %d sampled traces count: %d\n", len(origin), len(sampled))
}

func TestCloseResource(t *testing.T) {
	testcases := DatakitTraces{
		randDatakitTraceByService(t, 10, "login", "Allen123"),
		randDatakitTraceByService(t, 10, "game", "Bravo333"),
		randDatakitTraceByService(t, 10, "logout", "Clear666"),
	}
	expected := []func(trace DatakitTrace) bool{
		func(trace DatakitTrace) bool { return trace != nil },
		func(trace DatakitTrace) bool { return trace == nil },
		func(trace DatakitTrace) bool { return trace != nil },
	}

	closer := &CloseResource{
		IgnoreResources: map[string][]*regexp.Regexp{"game": {regexp.MustCompile(".*333")}},
	}
	for i := range testcases {
		parentialize(testcases[i])

		trace, _ := closer.Close(testcases[i])
		if !expected[i](trace) {
			t.Errorf("close resource %s failed trace:%v", testcases[i][0].Resource, trace)
		}
	}
}
