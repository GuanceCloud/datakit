package grok

import (
	"regexp"

	internalfilter "github.com/GuanceCloud/grok/internal/filter"
)

type regexpFilter struct {
	re           *regexp.Regexp
	expr         string
	captureCount int
}

func buildRegexpFilter(expr string) *regexpFilter {
	filter := internalfilter.BuildRegexpFilter(expr)
	if filter == nil {
		return nil
	}
	return &regexpFilter{
		re:           filter.RE,
		expr:         filter.Expr,
		captureCount: filter.CaptureCount,
	}
}
