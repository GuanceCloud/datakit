package filter

import (
	"regexp"
	"regexp/syntax"
)

const (
	minRegexpFilterCaptures    = 3
	minRegexpFilterInstSavings = 2
)

type RegexpFilter struct {
	RE           *regexp.Regexp
	Expr         string
	CaptureCount int
}

func BuildRegexpFilter(expr string) *RegexpFilter {
	parsed, err := syntax.Parse(expr, syntax.Perl)
	if err != nil {
		return nil
	}
	parsed = parsed.Simplify()

	captureCount := regexpCaptureCount(parsed)
	if captureCount < minRegexpFilterCaptures {
		return nil
	}

	filterTree := stripRegexpCaptures(parsed)
	if filterTree == nil {
		return nil
	}
	filterTree = filterTree.Simplify()

	filterExpr := filterTree.String()
	if filterExpr == "" || filterExpr == expr {
		return nil
	}

	fullProg, err := syntax.Compile(parsed)
	if err != nil {
		return nil
	}
	filterProg, err := syntax.Compile(filterTree)
	if err != nil {
		return nil
	}
	if len(fullProg.Inst)-len(filterProg.Inst) < minRegexpFilterInstSavings {
		return nil
	}

	re, err := regexp.Compile(filterExpr)
	if err != nil {
		return nil
	}

	return &RegexpFilter{
		RE:           re,
		Expr:         filterExpr,
		CaptureCount: captureCount,
	}
}

func regexpCaptureCount(re *syntax.Regexp) int {
	if re == nil {
		return 0
	}

	count := 0
	if re.Op == syntax.OpCapture {
		count++
	}
	for _, sub := range re.Sub {
		count += regexpCaptureCount(sub)
	}
	return count
}

func stripRegexpCaptures(re *syntax.Regexp) *syntax.Regexp {
	if re == nil {
		return nil
	}
	if re.Op == syntax.OpCapture {
		if len(re.Sub) == 0 {
			return nil
		}
		return stripRegexpCaptures(re.Sub[0])
	}

	clone := *re
	if len(re.Sub) > 0 {
		clone.Sub = make([]*syntax.Regexp, len(re.Sub))
		for i, sub := range re.Sub {
			clone.Sub[i] = stripRegexpCaptures(sub)
		}
	}
	return &clone
}
