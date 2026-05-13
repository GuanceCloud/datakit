package grok

const solrTailPattern = `%{TIMESTAMP_ISO8601:time}%{SPACE}%{LOGLEVEL:status}%{SPACE}\(%{NOTSPACE:thread}\)%{SPACE}\[%{SPACE}%{NOTSPACE}?\]%{SPACE}%{solrReporter:reporter}.*`

type solrRunner struct {
	timeIdx     int
	statusIdx   int
	threadIdx   int
	reporterIdx int
}

func compileSolrRunner(pattern string, steps []structuredStep, storage PatternStorageIface) (*solrRunner, bool) {
	if pattern != solrTailPattern || len(steps) == 0 || !defaultPatternDefinitionsMatch(storage, "TIMESTAMP_ISO8601", "SPACE", "LOGLEVEL", "NOTSPACE", "solrReporter") {
		return nil, false
	}
	var timeIdx, statusIdx, threadIdx, reporterIdx = -1, -1, -1, -1
	for _, step := range steps {
		if step.parser == nil || step.parser.dstIndex < 0 {
			continue
		}
		switch {
		case timeIdx < 0 && step.parser.kind == structuredTimestampISO8601:
			timeIdx = step.parser.dstIndex
		case statusIdx < 0 && step.parser.kind == structuredLogLevel:
			statusIdx = step.parser.dstIndex
		case threadIdx < 0 && step.parser.kind == structuredNotSpace && step.parser.wrapPrefix == "(" && step.parser.wrapSuffix == ")":
			threadIdx = step.parser.dstIndex
		case reporterIdx < 0:
			reporterIdx = step.parser.dstIndex
		}
	}
	if timeIdx < 0 || statusIdx < 0 || threadIdx < 0 || reporterIdx < 0 {
		return nil, false
	}
	return &solrRunner{
		timeIdx:     timeIdx,
		statusIdx:   statusIdx,
		threadIdx:   threadIdx,
		reporterIdx: reporterIdx,
	}, true
}

func (r *solrRunner) run(dst []string, content string, trimSpace bool) bool {
	timeVal, statusVal, threadVal, reporterVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignStringAt(dst, r.timeIdx, timeVal, trimSpace)
	assignStringAt(dst, r.statusIdx, statusVal, trimSpace)
	assignStringAt(dst, r.threadIdx, threadVal, trimSpace)
	assignStringAt(dst, r.reporterIdx, reporterVal, trimSpace)
	return true
}

func (r *solrRunner) runTyped(dst []any, content string, trimSpace bool, kinds []valueKind) bool {
	timeVal, statusVal, threadVal, reporterVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignTypedAt(dst, kinds, r.timeIdx, timeVal, trimSpace)
	assignTypedAt(dst, kinds, r.statusIdx, statusVal, trimSpace)
	assignTypedAt(dst, kinds, r.threadIdx, threadVal, trimSpace)
	assignTypedAt(dst, kinds, r.reporterIdx, reporterVal, trimSpace)
	return true
}

func (r *solrRunner) match(content string) (timeVal, statusVal, threadVal, reporterVal string, ok bool) {
	next, ok := consumeTimestampISO8601(content, 0)
	if !ok || next == 0 {
		return "", "", "", "", false
	}
	timeVal = content[:next]
	pos := next

	pos = consumeRegexpSpaceRun(content, pos)
	statusStart := pos
	for pos < len(content) && isRabbitAlphaByte(content[pos]) {
		pos++
	}
	if pos == statusStart || !isLogLevelValue(content[statusStart:pos]) {
		return "", "", "", "", false
	}
	statusVal = content[statusStart:pos]

	pos = consumeRegexpSpaceRun(content, pos)
	if pos >= len(content) || content[pos] != '(' {
		return "", "", "", "", false
	}
	pos++
	threadStart := pos
	for pos < len(content) && !isASCIISpace(content[pos]) && content[pos] != ')' {
		pos++
	}
	if pos == threadStart || pos >= len(content) || content[pos] != ')' {
		return "", "", "", "", false
	}
	threadVal = content[threadStart:pos]
	pos++

	pos = consumeRegexpSpaceRun(content, pos)
	if pos >= len(content) || content[pos] != '[' {
		return "", "", "", "", false
	}
	pos++
	pos = consumeRegexpSpaceRun(content, pos)
	for pos < len(content) && !isASCIISpace(content[pos]) && content[pos] != ']' {
		pos++
	}
	if pos >= len(content) || content[pos] != ']' {
		return "", "", "", "", false
	}
	pos++

	pos = consumeRegexpSpaceRun(content, pos)
	start := pos
	for pos < len(content) && isSolrReporterByte(content[pos]) {
		pos++
	}
	if pos == start {
		return "", "", "", "", false
	}
	reporterVal = content[start:pos]
	return timeVal, statusVal, threadVal, reporterVal, true
}

func consumeRegexpSpaceRun(content string, pos int) int {
	for pos < len(content) && isRegexpASCIISpace(content[pos]) {
		pos++
	}
	return pos
}

func isSolrReporterByte(b byte) bool {
	return isWordByte(b) || b == '.'
}
