package grok

import (
	"regexp"
	"regexp/syntax"
	"sort"
	"strings"
)

type regexpPrefilter struct {
	anchoredPrefix string
	literalPrefix  string
	literalSet     []string
	literalExact   bool
	exactByLen     map[int][]string
	required       []string
}

func buildRegexpPrefilter(expr string, re *regexp.Regexp) *regexpPrefilter {
	if re == nil {
		return nil
	}

	pf := &regexpPrefilter{}
	parsed, err := syntax.Parse(expr, syntax.Perl)
	if err != nil {
		return nil
	}
	parsed = parsed.Simplify()

	if prefix, anchored := regexpLiteralPrefix(parsed); prefix != "" {
		if anchored {
			pf.anchoredPrefix = prefix
		} else {
			pf.literalPrefix = prefix
		}
	} else if !hasFoldCaseLiteral(parsed) {
		prefix, _ := re.LiteralPrefix()
		if hasRegexpStartAnchor(expr) && prefix != "" {
			pf.anchoredPrefix = prefix
		} else if prefix != "" {
			pf.literalPrefix = prefix
		}
	}

	if literals, exact := exactLiteralSet(parsed); len(literals) > 0 {
		pf.literalSet = literals
		pf.literalExact = exact
		if exact {
			pf.exactByLen = buildLiteralLengthBuckets(literals)
		}
		if exact && len(literals) == 1 {
			pf.anchoredPrefix = literals[0]
		}
	}

	if required := regexpRequiredLiterals(parsed, 4); len(required) > 0 {
		pf.required = required
	}

	if pf.anchoredPrefix == "" && pf.literalPrefix == "" && len(pf.literalSet) == 0 && len(pf.required) == 0 {
		return nil
	}
	return pf
}

func hasFoldCaseLiteral(re *syntax.Regexp) bool {
	if re == nil {
		return false
	}
	if re.Op == syntax.OpLiteral && re.Flags&syntax.FoldCase != 0 {
		return true
	}
	for _, sub := range re.Sub {
		if hasFoldCaseLiteral(sub) {
			return true
		}
	}
	return false
}

func (pf *regexpPrefilter) rejects(content string) bool {
	if pf == nil {
		return false
	}
	if pf.anchoredPrefix != "" && !strings.HasPrefix(content, pf.anchoredPrefix) {
		return true
	}
	if pf.literalPrefix != "" && !strings.Contains(content, pf.literalPrefix) {
		return true
	}
	if len(pf.literalSet) > 0 {
		if pf.literalExact {
			return !pf.matchesExactLiteral(content)
		}
		for _, lit := range pf.literalSet {
			if strings.Contains(content, lit) {
				return requiredLiteralRejects(content, pf.anchoredPrefix, pf.required)
			}
		}
		return true
	}
	return requiredLiteralRejects(content, pf.anchoredPrefix, pf.required)
}

func (pf *regexpPrefilter) matchesExactLiteral(content string) bool {
	if pf == nil {
		return false
	}
	if len(pf.exactByLen) == 0 {
		for _, lit := range pf.literalSet {
			if content == lit {
				return true
			}
		}
		return false
	}
	for _, lit := range pf.exactByLen[len(content)] {
		if content == lit {
			return true
		}
	}
	return false
}

func buildLiteralLengthBuckets(literals []string) map[int][]string {
	if len(literals) == 0 {
		return nil
	}
	buckets := make(map[int][]string, len(literals))
	for _, lit := range literals {
		n := len(lit)
		buckets[n] = append(buckets[n], lit)
	}
	return buckets
}

func hasRegexpStartAnchor(expr string) bool {
	return strings.HasPrefix(expr, "^")
}

func exactLiteralSet(parsed *syntax.Regexp) ([]string, bool) {
	literals, ok := exactLiteralSetFromRegexp(parsed, 32)
	if !ok || len(literals) == 0 {
		return nil, false
	}
	return literals, hasExactTextAnchors(parsed)
}

func exactLiteralSetFromRegexp(re *syntax.Regexp, limit int) ([]string, bool) {
	if re == nil {
		return nil, false
	}
	if re.Flags&syntax.FoldCase != 0 {
		return nil, false
	}

	switch re.Op {
	case syntax.OpLiteral:
		return []string{string(re.Rune)}, true
	case syntax.OpEmptyMatch:
		return []string{""}, true
	case syntax.OpBeginText, syntax.OpEndText, syntax.OpBeginLine, syntax.OpEndLine:
		return []string{""}, true
	case syntax.OpCapture:
		return exactLiteralSetFromRegexp(re.Sub[0], limit)
	case syntax.OpConcat:
		out := []string{""}
		for _, sub := range re.Sub {
			parts, ok := exactLiteralSetFromRegexp(sub, limit)
			if !ok {
				return nil, false
			}
			next := make([]string, 0, len(out)*len(parts))
			for _, prefix := range out {
				for _, part := range parts {
					next = append(next, prefix+part)
					if len(next) > limit {
						return nil, false
					}
				}
			}
			out = next
		}
		return out, true
	case syntax.OpAlternate:
		out := make([]string, 0, len(re.Sub))
		for _, sub := range re.Sub {
			parts, ok := exactLiteralSetFromRegexp(sub, limit-len(out))
			if !ok {
				return nil, false
			}
			out = append(out, parts...)
			if len(out) > limit {
				return nil, false
			}
		}
		return out, true
	case syntax.OpCharClass:
		out := make([]string, 0, len(re.Rune)/2)
		for i := 0; i+1 < len(re.Rune); i += 2 {
			lo, hi := re.Rune[i], re.Rune[i+1]
			if hi < lo {
				return nil, false
			}
			if hi-lo > 8 {
				return nil, false
			}
			for r := lo; r <= hi; r++ {
				out = append(out, string(r))
				if len(out) > limit {
					return nil, false
				}
			}
		}
		return out, len(out) > 0
	default:
		return nil, false
	}
}

func hasExactTextAnchors(re *syntax.Regexp) bool {
	if re == nil || re.Op != syntax.OpConcat || len(re.Sub) < 2 {
		return false
	}
	return isRegexpStartAnchor(re.Sub[0]) && isRegexpEndAnchor(re.Sub[len(re.Sub)-1])
}

func isRegexpStartAnchor(re *syntax.Regexp) bool {
	return re != nil && re.Op == syntax.OpBeginText
}

func isRegexpEndAnchor(re *syntax.Regexp) bool {
	return re != nil && re.Op == syntax.OpEndText
}

func regexpLiteralPrefix(re *syntax.Regexp) (string, bool) {
	if re == nil {
		return "", false
	}
	switch re.Op {
	case syntax.OpLiteral:
		if re.Flags&syntax.FoldCase != 0 {
			return "", false
		}
		return string(re.Rune), false
	case syntax.OpCapture:
		return regexpLiteralPrefix(re.Sub[0])
	case syntax.OpConcat:
		start := 0
		anchored := false
		if len(re.Sub) > 0 && isRegexpStartAnchor(re.Sub[0]) {
			anchored = true
			start = 1
		}
		var b strings.Builder
		for i := start; i < len(re.Sub); i++ {
			part, ok := regexpLiteralChunk(re.Sub[i])
			if !ok {
				break
			}
			b.WriteString(part)
		}
		return b.String(), anchored
	default:
		return "", false
	}
}

func regexpLiteralChunk(re *syntax.Regexp) (string, bool) {
	if re == nil {
		return "", false
	}
	if re.Flags&syntax.FoldCase != 0 {
		return "", false
	}
	switch re.Op {
	case syntax.OpLiteral:
		return string(re.Rune), true
	case syntax.OpCapture:
		return regexpLiteralChunk(re.Sub[0])
	case syntax.OpEmptyMatch, syntax.OpBeginText, syntax.OpEndText, syntax.OpBeginLine, syntax.OpEndLine:
		return "", true
	default:
		return "", false
	}
}

func regexpRequiredLiterals(parsed *syntax.Regexp, maxAtoms int) []string {
	chunks, ok := regexpRequiredLiteralChunks(parsed, 32)
	if !ok || len(chunks) == 0 {
		return nil
	}
	return selectRequiredLiteralAtoms(chunks, maxAtoms)
}

func regexpRequiredLiteralChunks(re *syntax.Regexp, limit int) ([]string, bool) {
	if re == nil {
		return nil, true
	}
	if limit <= 0 {
		return nil, false
	}

	switch re.Op {
	case syntax.OpLiteral:
		if re.Flags&syntax.FoldCase != 0 {
			return nil, true
		}
		return []string{string(re.Rune)}, true
	case syntax.OpEmptyMatch, syntax.OpBeginText, syntax.OpEndText, syntax.OpBeginLine, syntax.OpEndLine:
		return nil, true
	case syntax.OpCapture:
		return regexpRequiredLiteralChunks(re.Sub[0], limit)
	case syntax.OpConcat:
		out := make([]string, 0, len(re.Sub))
		for _, sub := range re.Sub {
			parts, ok := regexpRequiredLiteralChunks(sub, limit-len(out))
			if !ok {
				return nil, false
			}
			out = append(out, parts...)
			if len(out) > limit {
				return nil, false
			}
		}
		return out, true
	case syntax.OpAlternate:
		if len(re.Sub) == 0 {
			return nil, true
		}
		out, ok := regexpRequiredLiteralChunks(re.Sub[0], limit)
		if !ok {
			return nil, false
		}
		for i := 1; i < len(re.Sub); i++ {
			other, ok := regexpRequiredLiteralChunks(re.Sub[i], limit)
			if !ok {
				return nil, false
			}
			out = commonLiteralSequence(out, other)
			if len(out) == 0 {
				return nil, true
			}
		}
		return out, true
	case syntax.OpQuest, syntax.OpStar:
		return nil, true
	case syntax.OpPlus:
		return regexpRequiredLiteralChunks(re.Sub[0], limit)
	case syntax.OpRepeat:
		if re.Min <= 0 {
			return nil, true
		}
		return regexpRequiredLiteralChunks(re.Sub[0], limit)
	default:
		return nil, true
	}
}

func commonLiteralSequence(left, right []string) []string {
	if len(left) == 0 || len(right) == 0 {
		return nil
	}

	out := make([]string, 0, minInt(len(left), len(right)))
	start := 0
	for _, lit := range left {
		for ; start < len(right); start++ {
			if right[start] != lit {
				continue
			}
			out = append(out, lit)
			start++
			break
		}
	}
	return out
}

func containsLiteralsInOrder(content string, literals []string) bool {
	for _, lit := range literals {
		if lit == "" {
			continue
		}
		idx := strings.Index(content, lit)
		if idx < 0 {
			return false
		}
		content = content[idx+len(lit):]
	}
	return true
}

func requiredLiteralRejects(content string, anchoredPrefix string, required []string) bool {
	if len(required) == 0 {
		return false
	}
	if anchoredPrefix != "" && required[0] == anchoredPrefix {
		if !strings.HasPrefix(content, anchoredPrefix) {
			return true
		}
		content = content[len(anchoredPrefix):]
		required = required[1:]
	}
	return !containsLiteralsInOrder(content, required)
}

type literalCandidate struct {
	idx int
	lit string
}

func selectRequiredLiteralAtoms(chunks []string, maxAtoms int) []string {
	if len(chunks) == 0 || maxAtoms <= 0 {
		return nil
	}

	candidates := make([]literalCandidate, 0, len(chunks))
	for i, lit := range chunks {
		if len(lit) < 2 {
			continue
		}
		if len(candidates) > 0 && candidates[len(candidates)-1].lit == lit {
			continue
		}
		candidates = append(candidates, literalCandidate{idx: i, lit: lit})
	}
	if len(candidates) == 0 {
		return nil
	}
	if len(candidates) > maxAtoms {
		sort.SliceStable(candidates, func(i, j int) bool {
			if len(candidates[i].lit) != len(candidates[j].lit) {
				return len(candidates[i].lit) > len(candidates[j].lit)
			}
			return candidates[i].idx < candidates[j].idx
		})
		candidates = candidates[:maxAtoms]
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].idx < candidates[j].idx
		})
	}

	out := make([]string, len(candidates))
	for i, candidate := range candidates {
		out[i] = candidate.lit
	}
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
