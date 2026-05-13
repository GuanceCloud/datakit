package grok

import "strings"

const rabbitMQBracketPattern = `%{DATA:time} \[%{LOGLEVEL:status}\] %{GREEDYDATA:msg}`

type rabbitMQRunner struct {
	timeIdx   int
	statusIdx int
	msgIdx    int
}

func compileRabbitMQRunner(pattern string, steps []structuredStep, storage PatternStorageIface) (*rabbitMQRunner, bool) {
	if pattern != rabbitMQBracketPattern || len(steps) != 3 || !defaultPatternDefinitionsMatch(storage, "DATA", "LOGLEVEL", "GREEDYDATA") {
		return nil, false
	}
	if steps[0].parser == nil || steps[1].parser == nil || steps[2].parser == nil {
		return nil, false
	}
	return &rabbitMQRunner{
		timeIdx:   steps[0].parser.dstIndex,
		statusIdx: steps[1].parser.dstIndex,
		msgIdx:    steps[2].parser.dstIndex,
	}, true
}

func (r *rabbitMQRunner) run(dst []string, content string, trimSpace bool) bool {
	timeVal, statusVal, msgVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignStringAt(dst, r.timeIdx, timeVal, trimSpace)
	assignStringAt(dst, r.statusIdx, statusVal, trimSpace)
	assignStringAt(dst, r.msgIdx, msgVal, trimSpace)
	return true
}

func (r *rabbitMQRunner) runTyped(dst []any, content string, trimSpace bool, kinds []valueKind) bool {
	timeVal, statusVal, msgVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignTypedAt(dst, kinds, r.timeIdx, timeVal, trimSpace)
	assignTypedAt(dst, kinds, r.statusIdx, statusVal, trimSpace)
	assignTypedAt(dst, kinds, r.msgIdx, msgVal, trimSpace)
	return true
}

func (r *rabbitMQRunner) match(content string) (timeVal string, statusVal string, msgVal string, ok bool) {
	search := 0
	for {
		rel := strings.Index(content[search:], " [")
		if rel < 0 {
			return "", "", "", false
		}
		pos := search + rel
		statusStart := pos + 2
		statusEnd := statusStart
		for statusEnd < len(content) && isRabbitAlphaByte(content[statusEnd]) {
			statusEnd++
		}
		if statusEnd == statusStart || !strings.HasPrefix(content[statusEnd:], "] ") {
			search = pos + 1
			continue
		}

		status := content[statusStart:statusEnd]
		if !isLogLevelValue(status) {
			search = pos + 1
			continue
		}

		return content[:pos], status, content[statusEnd+2:], true
	}
}

func isRabbitAlphaByte(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}
