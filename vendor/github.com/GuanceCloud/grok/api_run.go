package grok

// MatchCount returns the number of named capture slots in the compiled pattern.
// It is useful for callers that want to preallocate reusable buffers.
func (g *GrokRegexp) MatchCount() int {
	return g.matchCount()
}

// RunTo matches the content and writes results into dst when capacity allows.
// The returned slice aliases dst's backing array when reuse is possible.
func (g *GrokRegexp) RunTo(content string, trimSpace bool, dst []string) ([]string, error) {
	return g.runTo(content, trimSpace, dst)
}

// RunWithTypeInfoTo matches the content and writes typed results into dst when
// capacity allows. This is the public reusable-buffer entrypoint intended for
// pipeline-style callers that want to avoid per-call allocations.
func (g *GrokRegexp) RunWithTypeInfoTo(content string, trimSpace bool, dst []any) ([]any, error) {
	return g.runWithTypeInfoTo(content, trimSpace, dst)
}
