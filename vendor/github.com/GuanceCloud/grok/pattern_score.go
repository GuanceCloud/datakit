package grok

import (
	"fmt"
	"sort"
	"strings"
)

type patternScoreVerdict string

const (
	patternScoreFastPathPreferred patternScoreVerdict = "fast-path preferred"
	patternScoreRegexpPreferred   patternScoreVerdict = "regexp preferred"
	patternScorePrefilterOnly     patternScoreVerdict = "prefilter only"
)

type patternScore struct {
	Verdict          patternScoreVerdict
	Prefix           string
	Suffix           string
	RequiredLiterals []string
	MinMatchLength   int
	Metrics          patternScoreMetrics
	Reasons          []string
}

type patternScoreMetrics struct {
	LiteralCount                int
	LiteralBytes                int
	RefCount                    int
	PrimitiveRefCount           int
	WideRefCount                int
	GreedyRefCount              int
	OptionalCount               int
	AlternationCount            int
	RepeatCount                 int
	RawRegexCount               int
	UnsupportedRegexCount       int
	MaxDepth                    int
	DeterministicDelimiterCount int
}

type patternScoreSummary struct {
	metrics patternScoreMetrics
	exact   string
}

func compilePatternScoreIR(pattern string, storage PatternStorageIface) (patternScore, error) {
	matchIR, err := compileMatchIR(pattern)
	if err != nil {
		return patternScore{}, err
	}

	cache := make(map[string]patternScoreSummary)
	visiting := make(map[string]bool)
	summary := summarizePatternScore(matchIR.root, storage, cache, visiting, 1)
	prefilter := compilePatternPrefilterFromMatchIR(matchIR.root, storage)

	score := patternScore{
		Prefix:           prefilter.prefix,
		Suffix:           prefilter.suffix,
		RequiredLiterals: finalizePrefilterLiterals(prefilter),
		MinMatchLength:   prefilter.minLength,
		Metrics:          summary.metrics,
	}
	score.Verdict = decidePatternScoreVerdict(score)
	score.Reasons = explainPatternScore(score)
	return score, nil
}

func dumpPatternScore(pattern string, storage PatternStorageIface) (string, error) {
	score, err := compilePatternScoreIR(pattern, storage)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("PatternScore\n")
	fmt.Fprintf(&b, "verdict=%q\n", score.Verdict)
	fmt.Fprintf(&b, "prefix=%q\n", score.Prefix)
	fmt.Fprintf(&b, "suffix=%q\n", score.Suffix)
	fmt.Fprintf(&b, "required_literals=%q\n", score.RequiredLiterals)
	fmt.Fprintf(&b, "min_match_length=%d\n", score.MinMatchLength)
	b.WriteString("metrics\n")
	fmt.Fprintf(&b, "  literal_count=%d\n", score.Metrics.LiteralCount)
	fmt.Fprintf(&b, "  literal_bytes=%d\n", score.Metrics.LiteralBytes)
	fmt.Fprintf(&b, "  ref_count=%d\n", score.Metrics.RefCount)
	fmt.Fprintf(&b, "  primitive_ref_count=%d\n", score.Metrics.PrimitiveRefCount)
	fmt.Fprintf(&b, "  wide_ref_count=%d\n", score.Metrics.WideRefCount)
	fmt.Fprintf(&b, "  greedy_ref_count=%d\n", score.Metrics.GreedyRefCount)
	fmt.Fprintf(&b, "  optional_count=%d\n", score.Metrics.OptionalCount)
	fmt.Fprintf(&b, "  alternation_count=%d\n", score.Metrics.AlternationCount)
	fmt.Fprintf(&b, "  repeat_count=%d\n", score.Metrics.RepeatCount)
	fmt.Fprintf(&b, "  raw_regex_count=%d\n", score.Metrics.RawRegexCount)
	fmt.Fprintf(&b, "  unsupported_regex_count=%d\n", score.Metrics.UnsupportedRegexCount)
	fmt.Fprintf(&b, "  max_depth=%d\n", score.Metrics.MaxDepth)
	fmt.Fprintf(&b, "  deterministic_delimiter_count=%d\n", score.Metrics.DeterministicDelimiterCount)
	b.WriteString("reasons\n")
	for _, reason := range score.Reasons {
		fmt.Fprintf(&b, "  - %s\n", reason)
	}
	return b.String(), nil
}

func summarizePatternScore(node *matchIRNode, storage PatternStorageIface, cache map[string]patternScoreSummary, visiting map[string]bool, depth int) patternScoreSummary {
	summary := patternScoreSummary{}
	if depth > summary.metrics.MaxDepth {
		summary.metrics.MaxDepth = depth
	}
	if node == nil {
		return summary
	}

	switch node.kind {
	case matchIREmpty:
		return summary
	case matchIRLiteral:
		summary.metrics.LiteralCount = 1
		summary.metrics.LiteralBytes = len(node.literal)
		summary.exact = node.literal
		return summary
	case matchIRRef:
		summary.metrics.RefCount = 1
		if kind, ok := structuredPrimitiveKind(node.ref.syntax); ok {
			summary.metrics.PrimitiveRefCount = 1
			if kind == structuredUntilLiteral || kind == structuredGreedyUntilLiteral {
				summary.metrics.WideRefCount = 1
			}
			if kind == structuredGreedyUntilLiteral {
				summary.metrics.GreedyRefCount = 1
			}
			return summary
		}
		if storage == nil || node.ref.syntax == "" {
			return summary
		}
		if child, ok := summarizeRefPattern(node.ref.syntax, storage, cache, visiting, depth+1); ok {
			summary.metrics = mergePatternScoreMetrics(summary.metrics, child.metrics)
			return summary
		}
		return summary
	case matchIRRawRegex:
		summary.metrics.RawRegexCount = 1
		if isUnsupportedRawRegexForScore(node.raw) {
			summary.metrics.UnsupportedRegexCount = 1
		}
		return summary
	case matchIRSequence:
		return summarizeSequencePatternScore(node.children, storage, cache, visiting, depth)
	case matchIRAlternate:
		return summarizeAlternatePatternScore(node.children, storage, cache, visiting, depth)
	case matchIRRepeat:
		return summarizeRepeatPatternScore(node, storage, cache, visiting, depth)
	default:
		return summary
	}
}

func summarizeRefPattern(name string, storage PatternStorageIface, cache map[string]patternScoreSummary, visiting map[string]bool, depth int) (patternScoreSummary, bool) {
	if name == "" || storage == nil {
		return patternScoreSummary{}, false
	}
	if summary, ok := cache[name]; ok {
		return summary, true
	}
	if visiting[name] {
		return patternScoreSummary{}, false
	}

	gp, ok := storage.GetPattern(name)
	if !ok || gp == nil {
		return patternScoreSummary{}, false
	}
	compiled, err := compileMatchIR(gp.pattern)
	if err != nil {
		return patternScoreSummary{}, false
	}

	visiting[name] = true
	summary := summarizePatternScore(compiled.root, storage, cache, visiting, depth)
	delete(visiting, name)
	cache[name] = summary
	return summary, true
}

func summarizeSequencePatternScore(children []*matchIRNode, storage PatternStorageIface, cache map[string]patternScoreSummary, visiting map[string]bool, depth int) patternScoreSummary {
	summary := patternScoreSummary{}
	summary.metrics.MaxDepth = depth

	childSummaries := make([]patternScoreSummary, len(children))
	for i, child := range children {
		childSummary := summarizePatternScore(child, storage, cache, visiting, depth+1)
		childSummaries[i] = childSummary
		summary.metrics = mergePatternScoreMetrics(summary.metrics, childSummary.metrics)
	}

	for i := 1; i < len(childSummaries)-1; i++ {
		if childSummaries[i].exact == "" {
			continue
		}
		if childSummaries[i-1].exact == "" || childSummaries[i+1].exact == "" {
			summary.metrics.DeterministicDelimiterCount++
		}
	}

	var exact strings.Builder
	for _, child := range childSummaries {
		if child.exact == "" {
			summary.exact = ""
			return summary
		}
		exact.WriteString(child.exact)
	}
	summary.exact = exact.String()
	return summary
}

func summarizeAlternatePatternScore(children []*matchIRNode, storage PatternStorageIface, cache map[string]patternScoreSummary, visiting map[string]bool, depth int) patternScoreSummary {
	summary := patternScoreSummary{}
	summary.metrics.MaxDepth = depth
	if len(children) > 0 {
		summary.metrics.AlternationCount = 1
	}

	var exact string
	exactSet := true
	for i, child := range children {
		childSummary := summarizePatternScore(child, storage, cache, visiting, depth+1)
		summary.metrics = mergePatternScoreMetrics(summary.metrics, childSummary.metrics)
		if i == 0 {
			exact = childSummary.exact
			continue
		}
		if exact == "" || exact != childSummary.exact {
			exactSet = false
		}
	}

	if exactSet {
		summary.exact = exact
	}
	return summary
}

func summarizeRepeatPatternScore(node *matchIRNode, storage PatternStorageIface, cache map[string]patternScoreSummary, visiting map[string]bool, depth int) patternScoreSummary {
	summary := patternScoreSummary{}
	summary.metrics.MaxDepth = depth
	summary.metrics.RepeatCount = 1
	if node.min == 0 && node.max == 1 {
		summary.metrics.OptionalCount = 1
	}
	if len(node.children) == 0 {
		return summary
	}

	child := summarizePatternScore(node.children[0], storage, cache, visiting, depth+1)
	summary.metrics = mergePatternScoreMetrics(summary.metrics, child.metrics)

	if node.min == node.max && node.min > 0 && child.exact != "" {
		summary.exact = strings.Repeat(child.exact, node.min)
	}
	return summary
}

func mergePatternScoreMetrics(dst, src patternScoreMetrics) patternScoreMetrics {
	dst.LiteralCount += src.LiteralCount
	dst.LiteralBytes += src.LiteralBytes
	dst.RefCount += src.RefCount
	dst.PrimitiveRefCount += src.PrimitiveRefCount
	dst.WideRefCount += src.WideRefCount
	dst.GreedyRefCount += src.GreedyRefCount
	dst.OptionalCount += src.OptionalCount
	dst.AlternationCount += src.AlternationCount
	dst.RepeatCount += src.RepeatCount
	dst.RawRegexCount += src.RawRegexCount
	dst.UnsupportedRegexCount += src.UnsupportedRegexCount
	dst.DeterministicDelimiterCount += src.DeterministicDelimiterCount
	if src.MaxDepth > dst.MaxDepth {
		dst.MaxDepth = src.MaxDepth
	}
	return dst
}

func isUnsupportedRawRegexForScore(raw string) bool {
	switch raw {
	case "", "^", "$", `\b`, `\B`, `\A`, `\z`, `(?s)`:
		return false
	case `\s`, `\S`, `\d`, `\D`, `\w`, `\W`:
		return false
	default:
		if strings.HasPrefix(raw, "[") && strings.HasSuffix(raw, "]") {
			return false
		}
		return true
	}
}

func decidePatternScoreVerdict(score patternScore) patternScoreVerdict {
	m := score.Metrics
	if m.UnsupportedRegexCount == 0 &&
		m.GreedyRefCount <= 1 &&
		m.WideRefCount <= 1 &&
		m.OptionalCount <= 1 &&
		m.AlternationCount <= 1 &&
		m.MaxDepth <= 5 &&
		m.DeterministicDelimiterCount >= 1 {
		return patternScoreFastPathPreferred
	}
	if score.Prefix != "" || len(score.RequiredLiterals) > 0 || score.MinMatchLength > 0 {
		return patternScorePrefilterOnly
	}
	return patternScoreRegexpPreferred
}

func explainPatternScore(score patternScore) []string {
	m := score.Metrics
	reasons := make([]string, 0, 8)

	if score.Prefix != "" {
		reasons = append(reasons, fmt.Sprintf("has fixed prefix %q", score.Prefix))
	}
	if len(score.RequiredLiterals) > 0 {
		reasons = append(reasons, fmt.Sprintf("has %d required literal fragment(s)", len(score.RequiredLiterals)))
	}
	if score.MinMatchLength > 0 {
		reasons = append(reasons, fmt.Sprintf("minimum match length is %d", score.MinMatchLength))
	}
	if m.DeterministicDelimiterCount > 0 {
		reasons = append(reasons, fmt.Sprintf("has %d deterministic delimiter boundary(ies)", m.DeterministicDelimiterCount))
	}
	if m.GreedyRefCount > 0 {
		reasons = append(reasons, fmt.Sprintf("contains %d greedy ref(s)", m.GreedyRefCount))
	}
	if m.OptionalCount > 0 {
		reasons = append(reasons, fmt.Sprintf("contains %d optional segment(s)", m.OptionalCount))
	}
	if m.AlternationCount > 0 {
		reasons = append(reasons, fmt.Sprintf("contains %d alternation node(s)", m.AlternationCount))
	}
	if m.UnsupportedRegexCount > 0 {
		reasons = append(reasons, fmt.Sprintf("contains %d unsupported raw regex node(s)", m.UnsupportedRegexCount))
	}
	if m.MaxDepth > 0 {
		reasons = append(reasons, fmt.Sprintf("maximum derived depth is %d", m.MaxDepth))
	}

	switch score.Verdict {
	case patternScoreFastPathPreferred:
		reasons = append(reasons, "score favors structured execution but does not change runtime behavior")
	case patternScorePrefilterOnly:
		reasons = append(reasons, "score favors conservative prefiltering while keeping regexp fallback")
	case patternScoreRegexpPreferred:
		reasons = append(reasons, "score favors regexp fallback")
	}
	return reasons
}

func sortedPatternScoreLiterals(literals []string) []string {
	out := append([]string(nil), literals...)
	sort.Strings(out)
	return out
}
