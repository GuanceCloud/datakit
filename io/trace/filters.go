package trace

type TraceBlocker interface{}

func BlockTraces(traces DatakitTrace, blockers ...TraceBlocker) DatakitTrace {
	return nil
}
