package grok

import "strings"

const (
	jenkinsPattern        = `%{TIMESTAMP_ISO8601:time} \[id=%{GREEDYDATA:id}\]\t%{GREEDYDATA:status}\t`
	jenkinsPatternTabs    = "%{TIMESTAMP_ISO8601:time} \\[id=%{GREEDYDATA:id}\\]\t%{GREEDYDATA:status}\t"
	tomcatCatalinaPattern = `%{olf_time:time} %{LOGLEVEL:status} \[%{NOTSPACE:thread_name}\] %{NOTSPACE:report_source} %{GREEDYDATA:msg}`
)

type jenkinsRunner struct {
	timeIdx   int
	idIdx     int
	statusIdx int
}

func compileJenkinsRunner(pattern string, nameIndex map[string]int, storage PatternStorageIface) (*jenkinsRunner, bool) {
	if (pattern != jenkinsPattern && pattern != jenkinsPatternTabs) || nameIndex == nil || !defaultPatternDefinitionsMatch(storage, "TIMESTAMP_ISO8601", "GREEDYDATA") {
		return nil, false
	}
	timeIdx, ok := nameIndex["time"]
	if !ok {
		return nil, false
	}
	idIdx, ok := nameIndex["id"]
	if !ok {
		return nil, false
	}
	statusIdx, ok := nameIndex["status"]
	if !ok {
		return nil, false
	}
	return &jenkinsRunner{
		timeIdx:   timeIdx,
		idIdx:     idIdx,
		statusIdx: statusIdx,
	}, true
}

func (r *jenkinsRunner) run(dst []string, content string, trimSpace bool) bool {
	timeVal, idVal, statusVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignStringAt(dst, r.timeIdx, timeVal, trimSpace)
	assignStringAt(dst, r.idIdx, idVal, trimSpace)
	assignStringAt(dst, r.statusIdx, statusVal, trimSpace)
	return true
}

func (r *jenkinsRunner) runTyped(dst []any, content string, trimSpace bool, kinds []valueKind) bool {
	timeVal, idVal, statusVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignTypedAt(dst, kinds, r.timeIdx, timeVal, trimSpace)
	assignTypedAt(dst, kinds, r.idIdx, idVal, trimSpace)
	assignTypedAt(dst, kinds, r.statusIdx, statusVal, trimSpace)
	return true
}

func (r *jenkinsRunner) match(content string) (timeVal, idVal, statusVal string, ok bool) {
	next, ok := consumeTimestampISO8601(content, 0)
	if !ok || !strings.HasPrefix(content[next:], " [id=") {
		return "", "", "", false
	}
	timeVal = content[:next]
	pos := next + len(" [id=")
	endID := strings.Index(content[pos:], "]\t")
	if endID < 0 {
		return "", "", "", false
	}
	idVal = content[pos : pos+endID]
	pos += endID + 2
	endStatus := strings.IndexByte(content[pos:], '\t')
	if endStatus < 0 {
		return "", "", "", false
	}
	statusVal = content[pos : pos+endStatus]
	return timeVal, idVal, statusVal, true
}

type tomcatCatalinaRunner struct {
	timeIdx         int
	statusIdx       int
	threadIdx       int
	reportSourceIdx int
	msgIdx          int
}

func compileTomcatCatalinaRunner(pattern string, nameIndex map[string]int, storage PatternStorageIface) (*tomcatCatalinaRunner, bool) {
	if pattern != tomcatCatalinaPattern || nameIndex == nil || !defaultPatternDefinitionsMatch(storage, "olf_time", "LOGLEVEL", "NOTSPACE", "GREEDYDATA") {
		return nil, false
	}
	timeIdx, ok := nameIndex["time"]
	if !ok {
		return nil, false
	}
	statusIdx, ok := nameIndex["status"]
	if !ok {
		return nil, false
	}
	threadIdx, ok := nameIndex["thread_name"]
	if !ok {
		return nil, false
	}
	reportSourceIdx, ok := nameIndex["report_source"]
	if !ok {
		return nil, false
	}
	msgIdx, ok := nameIndex["msg"]
	if !ok {
		return nil, false
	}
	return &tomcatCatalinaRunner{
		timeIdx:         timeIdx,
		statusIdx:       statusIdx,
		threadIdx:       threadIdx,
		reportSourceIdx: reportSourceIdx,
		msgIdx:          msgIdx,
	}, true
}

func (r *tomcatCatalinaRunner) run(dst []string, content string, trimSpace bool) bool {
	timeVal, statusVal, threadVal, reportVal, msgVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignStringAt(dst, r.timeIdx, timeVal, trimSpace)
	assignStringAt(dst, r.statusIdx, statusVal, trimSpace)
	assignStringAt(dst, r.threadIdx, threadVal, trimSpace)
	assignStringAt(dst, r.reportSourceIdx, reportVal, trimSpace)
	assignStringAt(dst, r.msgIdx, msgVal, trimSpace)
	return true
}

func (r *tomcatCatalinaRunner) runTyped(dst []any, content string, trimSpace bool, kinds []valueKind) bool {
	timeVal, statusVal, threadVal, reportVal, msgVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignTypedAt(dst, kinds, r.timeIdx, timeVal, trimSpace)
	assignTypedAt(dst, kinds, r.statusIdx, statusVal, trimSpace)
	assignTypedAt(dst, kinds, r.threadIdx, threadVal, trimSpace)
	assignTypedAt(dst, kinds, r.reportSourceIdx, reportVal, trimSpace)
	assignTypedAt(dst, kinds, r.msgIdx, msgVal, trimSpace)
	return true
}

func (r *tomcatCatalinaRunner) match(content string) (timeVal, statusVal, threadVal, reportVal, msgVal string, ok bool) {
	next, ok := consumeOlfTime(content, 0)
	if !ok || next >= len(content) || content[next] != ' ' {
		return "", "", "", "", "", false
	}
	timeVal = content[:next]
	pos := next + 1

	statusStart := pos
	for pos < len(content) && isRabbitAlphaByte(content[pos]) {
		pos++
	}
	if pos == statusStart || pos >= len(content) || content[pos] != ' ' || !isLogLevelValue(content[statusStart:pos]) {
		return "", "", "", "", "", false
	}
	statusVal = content[statusStart:pos]
	pos++

	if pos >= len(content) || content[pos] != '[' {
		return "", "", "", "", "", false
	}
	pos++
	threadStart := pos
	for pos < len(content) && content[pos] != ']' {
		if isASCIISpace(content[pos]) {
			return "", "", "", "", "", false
		}
		pos++
	}
	if pos == threadStart || !strings.HasPrefix(content[pos:], "] ") {
		return "", "", "", "", "", false
	}
	threadVal = content[threadStart:pos]
	pos += 2

	reportStart := pos
	for pos < len(content) && !isASCIISpace(content[pos]) {
		pos++
	}
	if pos == reportStart || pos >= len(content) || content[pos] != ' ' {
		return "", "", "", "", "", false
	}
	reportVal = content[reportStart:pos]
	msgVal = content[pos+1:]
	return timeVal, statusVal, threadVal, reportVal, msgVal, true
}

func consumeOlfTime(s string, start int) (int, bool) {
	i := start
	if !consumeTwoDigitRange(s, &i, 1, 31, true) || i >= len(s) || s[i] != '-' {
		return 0, false
	}
	i++
	monthStart := i
	for i < len(s) && isRabbitAlphaByte(s[i]) {
		i++
	}
	if i == monthStart || !isMonthNameValue(s[monthStart:i]) || i >= len(s) || s[i] != '-' {
		return 0, false
	}
	i++
	if !consumeOneOrTwoYearChunks(s, &i) || i >= len(s) || s[i] != ' ' {
		return 0, false
	}
	i++
	next, ok := consumeTimeOfDay(s, i)
	if !ok {
		return 0, false
	}
	return next, true
}
