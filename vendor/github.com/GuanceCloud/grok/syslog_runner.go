package grok

import "strings"

const postfixQueuePattern = `%{SYSLOGTIMESTAMP:timestamp} %{SYSLOGHOST:host} %{WORD:program}\[%{POSINT:pid}\]: %{QUEUEID:queue_id}: %{GREEDYDATA:msg}`

type postfixQueueRunner struct {
	timestampIdx int
	hostIdx      int
	programIdx   int
	pidIdx       int
	queueIdx     int
	msgIdx       int
}

func compilePostfixQueueRunner(pattern string, nameIndex map[string]int, storage PatternStorageIface) (*postfixQueueRunner, bool) {
	if pattern != postfixQueuePattern || nameIndex == nil || !defaultPatternDefinitionsMatch(storage, "SYSLOGTIMESTAMP", "SYSLOGHOST", "WORD", "POSINT", "QUEUEID", "GREEDYDATA") {
		return nil, false
	}
	r := &postfixQueueRunner{
		timestampIdx: lookupNameIndex(nameIndex, "timestamp"),
		hostIdx:      lookupNameIndex(nameIndex, "host"),
		programIdx:   lookupNameIndex(nameIndex, "program"),
		pidIdx:       lookupNameIndex(nameIndex, "pid"),
		queueIdx:     lookupNameIndex(nameIndex, "queue_id"),
		msgIdx:       lookupNameIndex(nameIndex, "msg"),
	}
	if r.timestampIdx < 0 || r.hostIdx < 0 || r.programIdx < 0 || r.pidIdx < 0 || r.queueIdx < 0 || r.msgIdx < 0 {
		return nil, false
	}
	return r, true
}

func (r *postfixQueueRunner) run(dst []string, content string, trimSpace bool) bool {
	timestamp, host, program, pid, queueID, msg, ok := r.match(content)
	if !ok {
		return false
	}
	assignStringAt(dst, r.timestampIdx, timestamp, trimSpace)
	assignStringAt(dst, r.hostIdx, host, trimSpace)
	assignStringAt(dst, r.programIdx, program, trimSpace)
	assignStringAt(dst, r.pidIdx, pid, trimSpace)
	assignStringAt(dst, r.queueIdx, queueID, trimSpace)
	assignStringAt(dst, r.msgIdx, msg, trimSpace)
	return true
}

func (r *postfixQueueRunner) runTyped(dst []any, content string, trimSpace bool, kinds []valueKind) bool {
	timestamp, host, program, pid, queueID, msg, ok := r.match(content)
	if !ok {
		return false
	}
	assignTypedAt(dst, kinds, r.timestampIdx, timestamp, trimSpace)
	assignTypedAt(dst, kinds, r.hostIdx, host, trimSpace)
	assignTypedAt(dst, kinds, r.programIdx, program, trimSpace)
	assignTypedAt(dst, kinds, r.pidIdx, pid, trimSpace)
	assignTypedAt(dst, kinds, r.queueIdx, queueID, trimSpace)
	assignTypedAt(dst, kinds, r.msgIdx, msg, trimSpace)
	return true
}

func (r *postfixQueueRunner) match(content string) (timestamp, host, program, pid, queueID, msg string, ok bool) {
	next, ok := consumeSyslogTimestamp(content, 0)
	if !ok || next >= len(content) || content[next] != ' ' {
		return "", "", "", "", "", "", false
	}
	timestamp = content[:next]
	pos := next + 1

	hostStart := pos
	for pos < len(content) && content[pos] != ' ' {
		pos++
	}
	if pos == hostStart || pos >= len(content) || !structuredSegmentMatchesKind(content[hostStart:pos], structuredIPOrHost) {
		return "", "", "", "", "", "", false
	}
	host = content[hostStart:pos]
	pos++

	programStart := pos
	for pos < len(content) && isWordByte(content[pos]) {
		pos++
	}
	if pos == programStart || pos >= len(content) || content[pos] != '[' {
		return "", "", "", "", "", "", false
	}
	program = content[programStart:pos]
	pos++

	pidStart := pos
	pid, next, ok = slicePosInt(content, pos)
	if !ok || next == pidStart || !strings.HasPrefix(content[next:], "]: ") {
		return "", "", "", "", "", "", false
	}
	pos = next + len("]: ")

	queueStart := pos
	for pos < len(content) && isPostfixQueueIDByte(content[pos]) {
		pos++
	}
	queueID = content[queueStart:pos]
	if !isPostfixQueueID(queueID) || !strings.HasPrefix(content[pos:], ": ") {
		return "", "", "", "", "", "", false
	}
	pos += len(": ")

	msg = content[pos:]
	return timestamp, host, program, pid, queueID, msg, true
}

func consumeSyslogTimestamp(s string, start int) (int, bool) {
	i := start
	monthStart := i
	for i < len(s) && isRabbitAlphaByte(s[i]) {
		i++
	}
	if i == monthStart || !isMonthNameValue(s[monthStart:i]) || i >= len(s) || s[i] != ' ' {
		return 0, false
	}
	for i < len(s) && s[i] == ' ' {
		i++
	}
	if !consumeTwoDigitRange(s, &i, 1, 31, true) || i >= len(s) || s[i] != ' ' {
		return 0, false
	}
	i++
	next, ok := consumeTimeOfDay(s, i)
	if !ok {
		return 0, false
	}
	return next, true
}

func isPostfixQueueID(s string) bool {
	if len(s) != 10 && len(s) != 11 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if !isPostfixQueueIDByte(s[i]) {
			return false
		}
	}
	return true
}

func isPostfixQueueIDByte(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'A' && b <= 'F')
}
