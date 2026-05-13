package grok

import internalbackend "github.com/GuanceCloud/grok/internal/backend"

type multiPatternFilter = internalbackend.Filter
type multiPatternFilterBackend = internalbackend.Backend

var defaultMultiPatternFilterBackend = internalbackend.Default()

func compileMultiPatternFilter(patterns []string) (multiPatternFilter, error) {
	return internalbackend.CompileMultiPatternFilter(patterns)
}

func loadMultiPatternFilter(database []byte) (multiPatternFilter, error) {
	return internalbackend.LoadMultiPatternFilter(database)
}
