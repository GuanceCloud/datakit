package grok

import (
	"fmt"
	"strings"
)

type patternDecisionObservation struct {
	Pattern                  string
	Score                    patternScore
	StructuredCompiled       bool
	StructuredHasWork        bool
	StructuredHeuristicAllow bool
	StructuredStats          structuredStats
	Diverges                 bool
}

func observePatternDecision(pattern string, storage PatternStorageIface) (patternDecisionObservation, error) {
	observation := patternDecisionObservation{Pattern: pattern}

	score, err := compilePatternScoreIR(pattern, storage)
	if err != nil {
		return observation, err
	}
	observation.Score = score

	gP, err := DenormalizePattern(pattern, storage)
	if err != nil {
		return observation, err
	}
	meta, err := loadCompiledRegexpMeta(gP.denormalized)
	if err != nil {
		return observation, err
	}
	steps, ok := compileStructuredSteps(pattern, storage, meta, 0)
	if !ok || len(steps) == 0 {
		observation.Diverges = score.Verdict == patternScoreFastPathPreferred
		return observation, nil
	}

	observation.StructuredCompiled = true
	observation.StructuredHasWork = hasStructuredMatcherWork(steps)
	observation.StructuredStats = collectStructuredStats(steps)
	observation.StructuredHeuristicAllow = observation.StructuredHasWork && shouldUseStructuredMatcher(steps)
	observation.Diverges = observation.StructuredHeuristicAllow != scorePrefersStructured(score)
	return observation, nil
}

func dumpPatternDecisionObservation(pattern string, storage PatternStorageIface) (string, error) {
	observation, err := observePatternDecision(pattern, storage)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("PatternDecisionObservation\n")
	fmt.Fprintf(&b, "pattern=%q\n", pattern)
	fmt.Fprintf(&b, "score_verdict=%q\n", observation.Score.Verdict)
	fmt.Fprintf(&b, "structured_compiled=%t\n", observation.StructuredCompiled)
	fmt.Fprintf(&b, "structured_has_work=%t\n", observation.StructuredHasWork)
	fmt.Fprintf(&b, "heuristic_allows_fast_path=%t\n", observation.StructuredHeuristicAllow)
	fmt.Fprintf(&b, "diverges=%t\n", observation.Diverges)
	b.WriteString("structured_stats\n")
	fmt.Fprintf(&b, "  parser_count=%d\n", observation.StructuredStats.parserCount)
	fmt.Fprintf(&b, "  optional_count=%d\n", observation.StructuredStats.optionalCount)
	fmt.Fprintf(&b, "  alternative_count=%d\n", observation.StructuredStats.alternativeCount)
	fmt.Fprintf(&b, "  submatcher_count=%d\n", observation.StructuredStats.submatcherCount)
	fmt.Fprintf(&b, "  greedy_parser_count=%d\n", observation.StructuredStats.greedyParserCount)
	fmt.Fprintf(&b, "  data_parser_count=%d\n", observation.StructuredStats.dataParserCount)
	fmt.Fprintf(&b, "  newline_literal_count=%d\n", observation.StructuredStats.newlineLiteralCount)
	b.WriteString("score_reasons\n")
	for _, reason := range observation.Score.Reasons {
		fmt.Fprintf(&b, "  - %s\n", reason)
	}
	return b.String(), nil
}

func scorePrefersStructured(score patternScore) bool {
	return score.Verdict == patternScoreFastPathPreferred
}
