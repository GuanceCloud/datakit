package grok

import internalfilter "github.com/GuanceCloud/grok/internal/filter"

type matcherSetFilter = internalfilter.MatcherSetFilter
type matcherSetEvalContext = internalfilter.EvalContext

func compileMatcherSetFilter(pf *regexpPrefilter, atomIDs map[string]int) matcherSetFilter {
	if pf == nil {
		return matcherSetFilter{}
	}

	return internalfilter.Compile(internalfilter.Spec{
		AnchoredPrefix: pf.anchoredPrefix,
		LiteralPrefix:  pf.literalPrefix,
		LiteralExact:   pf.literalExact,
		ExactByLen:     pf.exactByLen,
		LiteralSet:     pf.literalSet,
		Required:       pf.required,
	}, atomIDs)
}

func runMatcherSetFilterProgram(filter matcherSetFilter, ctx matcherSetEvalContext) bool {
	return filter.RunProgram(ctx)
}

func runMatcherSetFilterStruct(filter matcherSetFilter, ctx matcherSetEvalContext) bool {
	return filter.RunStruct(ctx)
}

func matcherSetFilterJITEnabled(filter matcherSetFilter) bool {
	return filter.ProgramJITEnabled()
}

func matcherSetFilterJITCandidate(filter matcherSetFilter) bool {
	return filter.ProgramJITCandidate()
}

func matcherSetFilterJITInfo(filter matcherSetFilter) internalfilter.ProgramJITInfo {
	return filter.ProgramJITInfo()
}

func matcherSetFilterProgramRunnerKind(filter matcherSetFilter) string {
	return filter.ProgramRunnerKind()
}
