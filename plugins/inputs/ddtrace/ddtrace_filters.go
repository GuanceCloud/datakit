package ddtrace

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/hashcode"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

// traceFilter will determine whether a trace should be drop or not,
// return nil Trace if the trace data dropped,
// return true will tell runFilterWithBreak to break out loop early.
type traceFilter func(Trace) (Trace, bool)

func runFiltersWithBreak(trace Trace, filters ...traceFilter) Trace {
	if len(trace) == 0 {
		log.Debug("empty trace")

		return nil
	}

	var abort bool
	for _, filter := range filters {
		if trace, abort = filter(trace); abort || trace == nil {
			return trace
		}
	}

	return trace
}

var present = make(map[string]time.Time)

// service resource env.
func rare(trace Trace) (Trace, bool) {
	var rootSpan *Span
	for i := range trace {
		if trace[i].ParentID == 0 {
			rootSpan = trace[i]
			break
		}
	}
	if rootSpan == nil {
		return trace, false
	}

	checksum := hashcode.GenMapHash(map[string]string{
		"service":  rootSpan.Service,
		"resource": rootSpan.Resource,
		"env":      rootSpan.Meta[itrace.ENV],
	})
	if then, ok := present[checksum]; !ok || time.Since(then) >= time.Hour {
		present[checksum] = time.Now()

		return trace, true
	}

	return trace, false
}

func checkResource(trace Trace) (Trace, bool) {
	for i := range trace {
		for k := range ignoreResources {
			if ignoreResources[k].MatchString(trace[i].Resource) {
				return nil, false
			}
		}
	}

	return trace, false
}

func sample(trace Trace) (Trace, bool) {
	for i := range trace {
		if trace[i].ParentID == 0 {
			if priority, ok := trace[i].Metrics["_sampling_priority_v1"]; ok && priority < 1 {
				return nil, false
			}
			break
		}
	}

	return trace, false
}
