package ddtrace

type traceFilter func(Trace) Trace

func filterOutResource(trace Trace) Trace {
	for i := range trace {
		for k := range ignoreResources {
			if ignoreResources[k].MatchString(trace[i].Resource) {
				return nil
			}
		}
	}

	return trace
}

func sample(trace Trace) Trace {
	if len(trace) == 0 {
		return nil
	}

	for i := range trace {
		if trace[i].ParentID == 0 {
			if priority, ok := trace[i].Metrics["_sampling_priority_v1"]; ok && priority < 1 {
				return nil
			}
			break
		}
	}

	return trace

	// for i := range trace {
	// 	// bypass error
	// 	var status string = itrace.STATUS_OK
	// 	if trace[i].Error != 0 {
	// 		status = itrace.STATUS_ERR
	// 	}
	// 	if itrace.SampleIgnoreErrStatus(status) {
	// 		return trace
	// 	}
	// 	// bypass ingnore keys
	// 	if itrace.SampleIgnoreKeys(trace[i].Meta, sampleConf.IgnoreTagsList) {
	// 		return trace
	// 	}
	// 	// bypass tags
	// 	if itrace.SampleIgnoreTags(itrace.MergeTags(trace[i].Meta, ddTags), defIgnoreTags) {
	// 		return trace
	// 	}
	// }
	// // do sample
	// if itrace.DefSampleFunc(trace[0].TraceID, sampleConf.Rate, sampleConf.Scope) {
	// 	return trace
	// } else {
	// 	return nil
	// }
}
