package trace

import (
	"fmt"
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
