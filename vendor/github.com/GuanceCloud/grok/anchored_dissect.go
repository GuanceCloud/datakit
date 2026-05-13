package grok

import (
	"strings"

	internalmatch "github.com/GuanceCloud/grok/internal/match"
)

type anchoredDissectRunner struct {
	impl    *internalmatch.AnchoredDissectRunner
	parsers []*structuredParser
}

func compileAnchoredDissectRunner(steps []structuredStep, anchoredStart bool) (*anchoredDissectRunner, bool) {
	if !anchoredStart || len(steps) == 0 || matcherNeedsBacktracking(steps) {
		return nil, false
	}

	ops, parsers, ok := buildAnchoredDissectOps(steps)
	if !ok {
		return nil, false
	}

	return &anchoredDissectRunner{
		impl:    &internalmatch.AnchoredDissectRunner{Ops: ops},
		parsers: parsers,
	}, true
}

func compileAccessDissectRunner(steps []structuredStep) (*anchoredDissectRunner, bool) {
	if len(steps) == 0 || !matchesSimpleAccessShape(steps) {
		return nil, false
	}

	ops, parsers, ok := buildAnchoredDissectOps(steps)
	if !ok {
		return nil, false
	}

	return &anchoredDissectRunner{
		impl:    &internalmatch.AnchoredDissectRunner{Ops: ops},
		parsers: parsers,
	}, true
}

func buildAnchoredDissectOps(steps []structuredStep) ([]internalmatch.AnchoredDissectOp, []*structuredParser, bool) {
	if len(steps) == 0 {
		return nil, nil, false
	}

	ops := make([]internalmatch.AnchoredDissectOp, 0, len(steps))
	parsers := make([]*structuredParser, 0, len(steps))
	for _, step := range steps {
		if step.submatcher != nil || len(step.alternatives) > 0 {
			return nil, nil, false
		}
		if step.optional && !step.deterministicOptional {
			return nil, nil, false
		}
		if step.literal == "" && step.parser == nil {
			return nil, nil, false
		}

		parserIndex := -1
		parserDstIndex := -1
		if step.parser != nil {
			parserIndex = len(parsers)
			parserDstIndex = step.parser.dstIndex
			parsers = append(parsers, step.parser)
		}

		ops = append(ops, internalmatch.AnchoredDissectOp{
			Literal:        step.literal,
			ParserIndex:    parserIndex,
			ParserDstIndex: parserDstIndex,
			CaptureIndex:   step.captureIndex,
			Optional:       step.optional,
			OptPrefix:      step.optPrefix,
			OptSkip:        step.optPrefixSkips,
			TailMinWidth:   step.tailMinWidth,
		})
	}

	return ops, parsers, true
}

func matchesSimpleAccessShape(steps []structuredStep) bool {
	if matchesApacheFixtureAccessShape(steps) {
		return true
	}
	if len(steps) != 11 {
		return false
	}
	return stepHasParserKind(steps[0], structuredIPOrHost, structuredHostName, structuredNotSpace) &&
		stepHasWrappedParserKind(steps[1], " ", " ", structuredNotSpace) &&
		stepHasParserKind(steps[2], structuredNotSpace) &&
		stepHasWrappedParserKind(steps[3], " [", "] \"", structuredHTTPDate) &&
		stepHasParserKind(steps[4], structuredUntilLiteral, structuredGreedyUntilLiteral, structuredWord, structuredNotSpace) &&
		stepHasLiteral(steps[5], " ") &&
		stepHasParserKind(steps[6], structuredUntilLiteral, structuredGreedyUntilLiteral, structuredURIPathParam, structuredURIPath, structuredNotSpace) &&
		stepHasWrappedParserKind(steps[7], " HTTP/", "\" ", structuredNumber, structuredInt, structuredPosInt, structuredNonNegInt) &&
		stepHasParserKind(steps[8], structuredNumber, structuredInt, structuredPosInt, structuredNonNegInt) &&
		stepHasLiteral(steps[9], " ") &&
		stepHasParserKind(steps[10], structuredNumber, structuredInt, structuredPosInt, structuredNonNegInt)
}

func matchesApacheFixtureAccessShape(steps []structuredStep) bool {
	if len(steps) != 8 {
		return false
	}
	return stepHasParserKind(steps[0], structuredUntilLiteral, structuredGreedyUntilLiteral, structuredIPOrHost, structuredHostName, structuredNotSpace) &&
		stepHasWrappedParserKind(steps[1], " - - [", "] \"", structuredHTTPDate) &&
		stepHasParserKind(steps[2], structuredUntilLiteral, structuredGreedyUntilLiteral, structuredWord, structuredNotSpace) &&
		stepHasLiteral(steps[3], " ") &&
		stepHasParserKind(steps[4], structuredUntilLiteral, structuredGreedyUntilLiteral, structuredURIPathParam, structuredURIPath, structuredNotSpace) &&
		stepHasWrappedParserKind(steps[5], " HTTP/", "\" ", structuredNumber, structuredInt, structuredPosInt, structuredNonNegInt) &&
		stepHasParserKind(steps[6], structuredNumber, structuredInt, structuredPosInt, structuredNonNegInt) &&
		stepHasLiteral(steps[7], " ")
}

func stepHasLiteral(step structuredStep, lit string) bool {
	return !step.optional && step.submatcher == nil && len(step.alternatives) == 0 && step.parser == nil && step.literal == lit
}

func stepHasParserKind(step structuredStep, kinds ...structuredKind) bool {
	if step.optional || step.literal != "" || step.submatcher != nil || len(step.alternatives) != 0 || step.parser == nil || step.parser.wrapPrefix != "" || step.parser.wrapSuffix != "" {
		return false
	}
	for _, kind := range kinds {
		if step.parser.kind == kind {
			return true
		}
	}
	return false
}

func stepHasWrappedParserKind(step structuredStep, prefix, suffix string, kinds ...structuredKind) bool {
	if step.optional || step.literal != "" || step.submatcher != nil || len(step.alternatives) != 0 || step.parser == nil {
		return false
	}
	if step.parser.wrapPrefix != prefix || step.parser.wrapSuffix != suffix {
		return false
	}
	for _, kind := range kinds {
		if step.parser.kind == kind {
			return true
		}
	}
	return false
}

func (r *anchoredDissectRunner) run(dst []string, content string, trimSpace bool) (int, bool) {
	return r.runAt(dst, content, 0, trimSpace)
}

func (r *anchoredDissectRunner) runAt(dst []string, content string, pos int, trimSpace bool) (int, bool) {
	if r == nil || r.impl == nil {
		return 0, false
	}
	if pos < 0 || pos > len(content) {
		return 0, false
	}

	for opIdx, op := range r.impl.Ops {
		if len(content)-pos < op.TailMinWidth {
			return 0, false
		}
		if op.Literal != "" {
			if hasPrefixAt(content, pos, op.Literal) {
				if op.CaptureIndex >= 0 {
					dst[op.CaptureIndex] = maybeTrim(content[pos:pos+len(op.Literal)], trimSpace)
				}
				pos += len(op.Literal)
				continue
			}
			if op.Optional {
				if op.OptPrefix != "" && !op.OptSkip && hasPrefixAt(content, pos, op.OptPrefix) {
					return 0, false
				}
				continue
			}
			return 0, false
		}

		parser := r.parsers[op.ParserIndex]
		if opIdx == len(r.impl.Ops)-1 && canConsumeTerminalDotRun(parser) {
			next, value, ok := consumeTerminalDotRun(parser, content, pos)
			if !ok {
				return 0, false
			}
			if op.ParserDstIndex >= 0 {
				dst[op.ParserDstIndex] = maybeTrim(value, trimSpace)
			}
			return next, true
		}
		if op.Optional && shouldSkipOptionalParser(parser, structuredStep{
			parser:                parser,
			optional:              true,
			deterministicOptional: true,
			optPrefix:             op.OptPrefix,
			optPrefixSkips:        op.OptSkip,
		}, content, pos) {
			continue
		}

		next, value, ok := parser.consume(content, pos)
		if !ok {
			return 0, false
		}
		if op.ParserDstIndex >= 0 {
			dst[op.ParserDstIndex] = maybeTrim(value, trimSpace)
		}
		pos = next
	}

	return pos, true
}

func (r *anchoredDissectRunner) runTyped(dst []any, content string, trimSpace bool, kinds []valueKind) (int, bool) {
	return r.runTypedAt(dst, content, 0, trimSpace, kinds)
}

func (r *anchoredDissectRunner) runTypedAt(dst []any, content string, pos int, trimSpace bool, kinds []valueKind) (int, bool) {
	if r == nil || r.impl == nil {
		return 0, false
	}
	if pos < 0 || pos > len(content) {
		return 0, false
	}

	for opIdx, op := range r.impl.Ops {
		if len(content)-pos < op.TailMinWidth {
			return 0, false
		}
		if op.Literal != "" {
			if hasPrefixAt(content, pos, op.Literal) {
				if op.CaptureIndex >= 0 {
					dst[op.CaptureIndex] = castStructuredValue(content[pos:pos+len(op.Literal)], trimSpace, kinds[op.CaptureIndex])
				}
				pos += len(op.Literal)
				continue
			}
			if op.Optional {
				if op.OptPrefix != "" && !op.OptSkip && hasPrefixAt(content, pos, op.OptPrefix) {
					return 0, false
				}
				continue
			}
			return 0, false
		}

		parser := r.parsers[op.ParserIndex]
		if opIdx == len(r.impl.Ops)-1 && canConsumeTerminalDotRun(parser) {
			next, value, ok := consumeTerminalDotRun(parser, content, pos)
			if !ok {
				return 0, false
			}
			if op.ParserDstIndex >= 0 {
				dst[op.ParserDstIndex] = castStructuredValue(value, trimSpace, kinds[op.ParserDstIndex])
			}
			return next, true
		}
		if op.Optional && shouldSkipOptionalParser(parser, structuredStep{
			parser:                parser,
			optional:              true,
			deterministicOptional: true,
			optPrefix:             op.OptPrefix,
			optPrefixSkips:        op.OptSkip,
		}, content, pos) {
			continue
		}

		next, value, ok := parser.consume(content, pos)
		if !ok {
			return 0, false
		}
		if op.ParserDstIndex >= 0 {
			dst[op.ParserDstIndex] = castStructuredValue(value, trimSpace, kinds[op.ParserDstIndex])
		}
		pos = next
	}

	return pos, true
}

func hasPrefixAt(s string, pos int, prefix string) bool {
	return pos >= 0 && len(s)-pos >= len(prefix) && s[pos:pos+len(prefix)] == prefix
}

func canConsumeTerminalDotRun(p *structuredParser) bool {
	return p != nil &&
		p.inner == nil &&
		p.wrapPrefix == "" &&
		p.wrapSuffix == "" &&
		p.nextLiteral == "" &&
		p.kind == structuredGreedyUntilLiteral
}

func consumeTerminalDotRun(p *structuredParser, content string, pos int) (int, string, bool) {
	if pos > len(content) {
		return 0, "", false
	}
	next := len(content)
	if !p.dotAll {
		if idx := strings.IndexByte(content[pos:], '\n'); idx >= 0 {
			next = pos + idx
		}
	}
	if !p.allowEmpty && next == pos {
		return 0, "", false
	}
	return next, content[pos:next], true
}
