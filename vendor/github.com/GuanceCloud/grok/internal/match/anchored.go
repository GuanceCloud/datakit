package match

import "strings"

type AnchoredDissectRunner struct {
	Ops []AnchoredDissectOp
}

type AnchoredDissectOp struct {
	Literal        string
	ParserIndex    int
	ParserDstIndex int
	CaptureIndex   int
	Optional       bool
	OptPrefix      string
	OptSkip        bool
	TailMinWidth   int
}

func (r *AnchoredDissectRunner) RunStrings(
	dst []string,
	content string,
	trimSpace bool,
	consume func(parserIndex int, content string, pos int) (next int, value string, ok bool),
	shouldSkipOptional func(parserIndex int, optPrefix string, optSkip bool, content string, pos int) bool,
	assignLiteral func(dst []string, captureIndex int, value string, trimSpace bool),
	assignParsed func(dst []string, dstIndex int, value string, trimSpace bool),
) (int, bool) {
	if r == nil {
		return 0, false
	}

	pos := 0
	for _, op := range r.Ops {
		if len(content)-pos < op.TailMinWidth {
			return 0, false
		}
		if op.Literal != "" {
			if strings.HasPrefix(content[pos:], op.Literal) {
				if op.CaptureIndex >= 0 {
					assignLiteral(dst, op.CaptureIndex, content[pos:pos+len(op.Literal)], trimSpace)
				}
				pos += len(op.Literal)
				continue
			}
			if op.Optional {
				if op.OptPrefix != "" && !op.OptSkip && strings.HasPrefix(content[pos:], op.OptPrefix) {
					return 0, false
				}
				continue
			}
			return 0, false
		}

		if op.Optional && shouldSkipOptional(op.ParserIndex, op.OptPrefix, op.OptSkip, content, pos) {
			continue
		}

		next, value, ok := consume(op.ParserIndex, content, pos)
		if !ok {
			if op.Optional {
				return 0, false
			}
			return 0, false
		}
		if op.ParserDstIndex >= 0 {
			assignParsed(dst, op.ParserDstIndex, value, trimSpace)
		}
		pos = next
	}

	return pos, true
}

func (r *AnchoredDissectRunner) RunTyped(
	dst []any,
	content string,
	trimSpace bool,
	consume func(parserIndex int, content string, pos int) (next int, value string, ok bool),
	shouldSkipOptional func(parserIndex int, optPrefix string, optSkip bool, content string, pos int) bool,
	assignLiteral func(dst []any, captureIndex int, value string, trimSpace bool),
	assignParsed func(dst []any, dstIndex int, value string, trimSpace bool),
) (int, bool) {
	if r == nil {
		return 0, false
	}

	pos := 0
	for _, op := range r.Ops {
		if len(content)-pos < op.TailMinWidth {
			return 0, false
		}
		if op.Literal != "" {
			if strings.HasPrefix(content[pos:], op.Literal) {
				if op.CaptureIndex >= 0 {
					assignLiteral(dst, op.CaptureIndex, content[pos:pos+len(op.Literal)], trimSpace)
				}
				pos += len(op.Literal)
				continue
			}
			if op.Optional {
				if op.OptPrefix != "" && !op.OptSkip && strings.HasPrefix(content[pos:], op.OptPrefix) {
					return 0, false
				}
				continue
			}
			return 0, false
		}

		if op.Optional && shouldSkipOptional(op.ParserIndex, op.OptPrefix, op.OptSkip, content, pos) {
			continue
		}

		next, value, ok := consume(op.ParserIndex, content, pos)
		if !ok {
			if op.Optional {
				return 0, false
			}
			return 0, false
		}
		if op.ParserDstIndex >= 0 {
			assignParsed(dst, op.ParserDstIndex, value, trimSpace)
		}
		pos = next
	}

	return pos, true
}
