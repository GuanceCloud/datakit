package tracing

var (
	AllTraceSpanSet = NewSpanIDSet()
)

type SpanIDSet struct {
	Set map[string]struct{}
}

func NewSpanIDSet() *SpanIDSet {
	return &SpanIDSet{
		Set: make(map[string]struct{}),
	}
}

func (ss *SpanIDSet) Put(id string) {
	ss.Set[id] = struct{}{}
}

func (ss *SpanIDSet) Contains(id string) bool {
	// avoid nil pointer
	if ss.Set == nil {
		return false
	}
	_, ok := ss.Set[id]
	return ok
}

// getSpanParentID 查询 spanID 的顶级父级ID， 并进行路径压缩
func getSpanParentID(spanIDMaps map[string]string, topID string, spanID string) string {
	for {
		pid, ok := spanIDMaps[spanID]
		if !ok {
			return ""
		}
		if pid != "0" && pid != topID {
			spanIDMaps[spanID] = getSpanParentID(spanIDMaps, topID, pid)
		}
		return spanIDMaps[spanID]
	}
}
