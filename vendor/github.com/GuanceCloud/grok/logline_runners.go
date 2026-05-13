package grok

import "strings"

const (
	mysqlSimplePattern  = `%{TIMESTAMP_ISO8601:time}\s+%{INT:thread_id}\s+%{WORD:operation}\s+%{GREEDYDATA:raw_query}`
	sqlserverPattern    = `%{TIMESTAMP_ISO8601:time} %{NOTSPACE:origin}\s+%{GREEDYDATA:msg}`
	kafkaBracketPattern = `^\[%{date1:time}\] %{WORD:status} %{DATA:msg} \(%{DATA:name}\)`
	kingbasePattern     = `%{log_date:time}%{SPACE}\[%{INT:process_id}\]%{SPACE}%{status:status}:\s+%{GREEDYDATA:msg}`
	redisPattern        = `%{INT:pid}:%{WORD:role} %{date2:time} %{NOTSPACE:serverity} %{GREEDYDATA:msg}`
	damengPattern       = `%{log_date:time}%{SPACE}\[%{status:status}\]\s+%{GREEDYDATA:msg}`
)

var specializedRunnerAllowedCustomDefinitions = map[string][]string{
	"APPDATE":  {`%{YEAR}[./]%{MONTHNUM}[./]%{MONTHDAY} %{TIME}`},
	"UPBLOCK":  {`(upstream: "%{GREEDYDATA:upstream}", )?`, `(?:upstream: "%{GREEDYDATA:upstream}", )?`},
	"date2":    {`%{YEAR}[./]%{MONTHNUM}[./]%{MONTHDAY} %{TIME}`, `%{MONTHDAY} %{MONTH} %{YEAR}?%{TIME}`},
	"log_date": {`%{YEAR}-%{MONTHNUM}-%{MONTHDAY}%{SPACE}%{HOUR}:%{MINUTE}:%{SECOND}%{SPACE}(?:CST|UTC)`},
	"olf_time": {`%{MONTHDAY}-%{MONTH}-%{YEAR} %{TIME}`},
	"QUEUEID":  {`[0-9A-F]{10,11}`},
	"solrReporter": {
		`[\w.]+`,
		`[A-Za-z0-9_.]+`,
		`(?:[.\w\d]+)`,
	},
	"status": {`(LOG|ERROR|FATAL|PANIC|WARNING|NOTICE|INFO)`},
}

func defaultPatternDefinitionsMatch(storage PatternStorageIface, names ...string) bool {
	if storage == nil {
		return false
	}
	visiting := make(map[string]bool, len(names))
	for _, name := range names {
		if !defaultPatternDefinitionMatches(storage, name, visiting) {
			return false
		}
	}
	return true
}

func defaultPatternDefinitionMatches(storage PatternStorageIface, name string, visiting map[string]bool) bool {
	if visiting[name] {
		return false
	}
	pattern, ok := storage.GetPattern(name)
	if !ok || pattern == nil {
		return false
	}
	if !defaultOrAllowedCustomPatternDefinition(name, pattern.pattern) {
		return false
	}

	visiting[name] = true
	defer delete(visiting, name)

	ok = true
	if err := walkPatternRefs(pattern.pattern, func(_, _ int, ref patternRef) error {
		if !defaultPatternDefinitionMatches(storage, ref.syntax, visiting) {
			ok = false
		}
		return nil
	}); err != nil {
		return false
	}
	return ok
}

func defaultOrAllowedCustomPatternDefinition(name string, pattern string) bool {
	if defaultPattern, ok := defalutPatterns[name]; ok && pattern == defaultPattern {
		return true
	}
	return allowedCustomPatternDefinition(name, pattern)
}

func allowedCustomPatternDefinition(name string, pattern string) bool {
	for _, allowed := range specializedRunnerAllowedCustomDefinitions[name] {
		if pattern == allowed {
			return true
		}
	}
	return false
}

type mysqlSimpleRunner struct {
	timeIdx      int
	threadIdx    int
	operationIdx int
	queryIdx     int
}

func compileMySQLSimpleRunner(pattern string, nameIndex map[string]int, storage PatternStorageIface) (*mysqlSimpleRunner, bool) {
	if pattern != mysqlSimplePattern || nameIndex == nil || !defaultPatternDefinitionsMatch(storage, "TIMESTAMP_ISO8601", "INT", "WORD", "GREEDYDATA") {
		return nil, false
	}
	timeIdx, ok := nameIndex["time"]
	if !ok {
		return nil, false
	}
	threadIdx, ok := nameIndex["thread_id"]
	if !ok {
		return nil, false
	}
	operationIdx, ok := nameIndex["operation"]
	if !ok {
		return nil, false
	}
	queryIdx, ok := nameIndex["raw_query"]
	if !ok {
		return nil, false
	}
	return &mysqlSimpleRunner{
		timeIdx:      timeIdx,
		threadIdx:    threadIdx,
		operationIdx: operationIdx,
		queryIdx:     queryIdx,
	}, true
}

func (r *mysqlSimpleRunner) run(dst []string, content string, trimSpace bool) bool {
	timeVal, threadVal, opVal, queryVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignStringAt(dst, r.timeIdx, timeVal, trimSpace)
	assignStringAt(dst, r.threadIdx, threadVal, trimSpace)
	assignStringAt(dst, r.operationIdx, opVal, trimSpace)
	assignStringAt(dst, r.queryIdx, queryVal, trimSpace)
	return true
}

func (r *mysqlSimpleRunner) runTyped(dst []any, content string, trimSpace bool, kinds []valueKind) bool {
	timeVal, threadVal, opVal, queryVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignTypedAt(dst, kinds, r.timeIdx, timeVal, trimSpace)
	assignTypedAt(dst, kinds, r.threadIdx, threadVal, trimSpace)
	assignTypedAt(dst, kinds, r.operationIdx, opVal, trimSpace)
	assignTypedAt(dst, kinds, r.queryIdx, queryVal, trimSpace)
	return true
}

func (r *mysqlSimpleRunner) match(content string) (timeVal, threadVal, opVal, queryVal string, ok bool) {
	next, ok := consumeTimestampISO8601(content, 0)
	if !ok {
		return "", "", "", "", false
	}
	timeVal = content[:next]
	pos := skipASCIISpaces(content, next)
	threadStart := pos
	for pos < len(content) && isASCIIDigit(content[pos]) {
		pos++
	}
	if pos == threadStart {
		return "", "", "", "", false
	}
	threadVal = content[threadStart:pos]
	pos = skipASCIISpaces(content, pos)
	opStart := pos
	for pos < len(content) && isRabbitAlphaByte(content[pos]) {
		pos++
	}
	if pos == opStart || pos >= len(content) || !isASCIISpace(content[pos]) {
		return "", "", "", "", false
	}
	opVal = content[opStart:pos]
	pos = skipASCIISpaces(content, pos)
	if pos >= len(content) {
		return "", "", "", "", false
	}
	queryVal = content[pos:]
	return timeVal, threadVal, opVal, queryVal, true
}

type sqlServerRunner struct {
	timeIdx   int
	originIdx int
	msgIdx    int
}

func compileSQLServerRunner(pattern string, nameIndex map[string]int, storage PatternStorageIface) (*sqlServerRunner, bool) {
	if pattern != sqlserverPattern || nameIndex == nil || !defaultPatternDefinitionsMatch(storage, "TIMESTAMP_ISO8601", "NOTSPACE", "GREEDYDATA") {
		return nil, false
	}
	timeIdx, ok := nameIndex["time"]
	if !ok {
		return nil, false
	}
	originIdx, ok := nameIndex["origin"]
	if !ok {
		return nil, false
	}
	msgIdx, ok := nameIndex["msg"]
	if !ok {
		return nil, false
	}
	return &sqlServerRunner{timeIdx: timeIdx, originIdx: originIdx, msgIdx: msgIdx}, true
}

func (r *sqlServerRunner) run(dst []string, content string, trimSpace bool) bool {
	timeVal, originVal, msgVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignStringAt(dst, r.timeIdx, timeVal, trimSpace)
	assignStringAt(dst, r.originIdx, originVal, trimSpace)
	assignStringAt(dst, r.msgIdx, msgVal, trimSpace)
	return true
}

func (r *sqlServerRunner) runTyped(dst []any, content string, trimSpace bool, kinds []valueKind) bool {
	timeVal, originVal, msgVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignTypedAt(dst, kinds, r.timeIdx, timeVal, trimSpace)
	assignTypedAt(dst, kinds, r.originIdx, originVal, trimSpace)
	assignTypedAt(dst, kinds, r.msgIdx, msgVal, trimSpace)
	return true
}

func (r *sqlServerRunner) match(content string) (timeVal, originVal, msgVal string, ok bool) {
	next, ok := consumeTimestampISO8601(content, 0)
	if !ok || next >= len(content) || content[next] != ' ' {
		return "", "", "", false
	}
	timeVal = content[:next]
	pos := next + 1
	originStart := pos
	for pos < len(content) && !isASCIISpace(content[pos]) {
		pos++
	}
	if pos == originStart || pos >= len(content) {
		return "", "", "", false
	}
	originVal = content[originStart:pos]
	pos = skipASCIISpaces(content, pos)
	if pos >= len(content) {
		return "", "", "", false
	}
	msgVal = content[pos:]
	return timeVal, originVal, msgVal, true
}

type kafkaBracketRunner struct {
	timeIdx   int
	statusIdx int
	msgIdx    int
	nameIdx   int
}

func compileKafkaBracketRunner(pattern string, nameIndex map[string]int, storage PatternStorageIface) (*kafkaBracketRunner, bool) {
	if pattern != kafkaBracketPattern || nameIndex == nil || !defaultPatternDefinitionsMatch(storage, "date1", "WORD", "DATA") {
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
	msgIdx, ok := nameIndex["msg"]
	if !ok {
		return nil, false
	}
	nameIdx, ok := nameIndex["name"]
	if !ok {
		return nil, false
	}
	return &kafkaBracketRunner{timeIdx: timeIdx, statusIdx: statusIdx, msgIdx: msgIdx, nameIdx: nameIdx}, true
}

func (r *kafkaBracketRunner) run(dst []string, content string, trimSpace bool) bool {
	timeVal, statusVal, msgVal, nameVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignStringAt(dst, r.timeIdx, timeVal, trimSpace)
	assignStringAt(dst, r.statusIdx, statusVal, trimSpace)
	assignStringAt(dst, r.msgIdx, msgVal, trimSpace)
	assignStringAt(dst, r.nameIdx, nameVal, trimSpace)
	return true
}

func (r *kafkaBracketRunner) runTyped(dst []any, content string, trimSpace bool, kinds []valueKind) bool {
	timeVal, statusVal, msgVal, nameVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignTypedAt(dst, kinds, r.timeIdx, timeVal, trimSpace)
	assignTypedAt(dst, kinds, r.statusIdx, statusVal, trimSpace)
	assignTypedAt(dst, kinds, r.msgIdx, msgVal, trimSpace)
	assignTypedAt(dst, kinds, r.nameIdx, nameVal, trimSpace)
	return true
}

func (r *kafkaBracketRunner) match(content string) (timeVal, statusVal, msgVal, nameVal string, ok bool) {
	if len(content) < 3 || content[0] != '[' {
		return "", "", "", "", false
	}
	next, ok := consumeBracketDateTimeMillis(content, 1)
	if !ok || next >= len(content) || content[next] != ']' {
		return "", "", "", "", false
	}
	timeVal = content[1:next]
	pos := next + 1
	if pos >= len(content) || content[pos] != ' ' {
		return "", "", "", "", false
	}
	pos++
	statusStart := pos
	for pos < len(content) && isRabbitAlphaByte(content[pos]) {
		pos++
	}
	if pos == statusStart || pos >= len(content) || content[pos] != ' ' {
		return "", "", "", "", false
	}
	statusVal = content[statusStart:pos]
	pos++
	nameStart := strings.LastIndex(content[pos:], " (")
	if nameStart < 0 || !strings.HasSuffix(content, ")") {
		return "", "", "", "", false
	}
	namePos := pos + nameStart
	msgVal = content[pos:namePos]
	nameVal = content[namePos+2 : len(content)-1]
	if nameVal == "" {
		return "", "", "", "", false
	}
	return timeVal, statusVal, msgVal, nameVal, true
}

type kingbaseRunner struct {
	timeIdx   int
	pidIdx    int
	statusIdx int
	msgIdx    int
}

func compileKingbaseRunner(pattern string, nameIndex map[string]int, storage PatternStorageIface) (*kingbaseRunner, bool) {
	if pattern != kingbasePattern || nameIndex == nil || !defaultPatternDefinitionsMatch(storage, "log_date", "SPACE", "INT", "status", "GREEDYDATA") {
		return nil, false
	}
	timeIdx, ok := nameIndex["time"]
	if !ok {
		return nil, false
	}
	pidIdx, ok := nameIndex["process_id"]
	if !ok {
		return nil, false
	}
	statusIdx, ok := nameIndex["status"]
	if !ok {
		return nil, false
	}
	msgIdx, ok := nameIndex["msg"]
	if !ok {
		return nil, false
	}
	return &kingbaseRunner{timeIdx: timeIdx, pidIdx: pidIdx, statusIdx: statusIdx, msgIdx: msgIdx}, true
}

func (r *kingbaseRunner) run(dst []string, content string, trimSpace bool) bool {
	timeVal, pidVal, statusVal, msgVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignStringAt(dst, r.timeIdx, timeVal, trimSpace)
	assignStringAt(dst, r.pidIdx, pidVal, trimSpace)
	assignStringAt(dst, r.statusIdx, statusVal, trimSpace)
	assignStringAt(dst, r.msgIdx, msgVal, trimSpace)
	return true
}

func (r *kingbaseRunner) runTyped(dst []any, content string, trimSpace bool, kinds []valueKind) bool {
	timeVal, pidVal, statusVal, msgVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignTypedAt(dst, kinds, r.timeIdx, timeVal, trimSpace)
	assignTypedAt(dst, kinds, r.pidIdx, pidVal, trimSpace)
	assignTypedAt(dst, kinds, r.statusIdx, statusVal, trimSpace)
	assignTypedAt(dst, kinds, r.msgIdx, msgVal, trimSpace)
	return true
}

func (r *kingbaseRunner) match(content string) (timeVal, pidVal, statusVal, msgVal string, ok bool) {
	next, ok := consumeDateTimeWithTZ(content, 0)
	if !ok || next >= len(content) || content[next] != ' ' {
		return "", "", "", "", false
	}
	timeVal = content[:next]
	pos := next + 1
	if pos >= len(content) || content[pos] != '[' {
		return "", "", "", "", false
	}
	pos++
	pidStart := pos
	for pos < len(content) && isASCIIDigit(content[pos]) {
		pos++
	}
	if pos == pidStart || !strings.HasPrefix(content[pos:], "] ") {
		return "", "", "", "", false
	}
	pidVal = content[pidStart:pos]
	pos += 2
	statusStart := pos
	for pos < len(content) && isRabbitAlphaByte(content[pos]) {
		pos++
	}
	if pos == statusStart || !strings.HasPrefix(content[pos:], ":") {
		return "", "", "", "", false
	}
	statusVal = content[statusStart:pos]
	pos++
	pos = skipASCIISpaces(content, pos)
	if pos >= len(content) {
		return "", "", "", "", false
	}
	msgVal = content[pos:]
	return timeVal, pidVal, statusVal, msgVal, true
}

type redisRunner struct {
	pidIdx       int
	roleIdx      int
	timeIdx      int
	serverityIdx int
	msgIdx       int
}

func compileRedisRunner(pattern string, nameIndex map[string]int, storage PatternStorageIface) (*redisRunner, bool) {
	if pattern != redisPattern || nameIndex == nil || !defaultPatternDefinitionsMatch(storage, "INT", "WORD", "date2", "NOTSPACE", "GREEDYDATA") {
		return nil, false
	}
	pidIdx, ok := nameIndex["pid"]
	if !ok {
		return nil, false
	}
	roleIdx, ok := nameIndex["role"]
	if !ok {
		return nil, false
	}
	timeIdx, ok := nameIndex["time"]
	if !ok {
		return nil, false
	}
	serverityIdx, ok := nameIndex["serverity"]
	if !ok {
		return nil, false
	}
	msgIdx, ok := nameIndex["msg"]
	if !ok {
		return nil, false
	}
	return &redisRunner{pidIdx: pidIdx, roleIdx: roleIdx, timeIdx: timeIdx, serverityIdx: serverityIdx, msgIdx: msgIdx}, true
}

func (r *redisRunner) run(dst []string, content string, trimSpace bool) bool {
	pidVal, roleVal, timeVal, sevVal, msgVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignStringAt(dst, r.pidIdx, pidVal, trimSpace)
	assignStringAt(dst, r.roleIdx, roleVal, trimSpace)
	assignStringAt(dst, r.timeIdx, timeVal, trimSpace)
	assignStringAt(dst, r.serverityIdx, sevVal, trimSpace)
	assignStringAt(dst, r.msgIdx, msgVal, trimSpace)
	return true
}

func (r *redisRunner) runTyped(dst []any, content string, trimSpace bool, kinds []valueKind) bool {
	pidVal, roleVal, timeVal, sevVal, msgVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignTypedAt(dst, kinds, r.pidIdx, pidVal, trimSpace)
	assignTypedAt(dst, kinds, r.roleIdx, roleVal, trimSpace)
	assignTypedAt(dst, kinds, r.timeIdx, timeVal, trimSpace)
	assignTypedAt(dst, kinds, r.serverityIdx, sevVal, trimSpace)
	assignTypedAt(dst, kinds, r.msgIdx, msgVal, trimSpace)
	return true
}

func (r *redisRunner) match(content string) (pidVal, roleVal, timeVal, sevVal, msgVal string, ok bool) {
	pos := 0
	pidStart := pos
	for pos < len(content) && isASCIIDigit(content[pos]) {
		pos++
	}
	if pos == pidStart || pos >= len(content) || content[pos] != ':' {
		return "", "", "", "", "", false
	}
	pidVal = content[pidStart:pos]
	pos++
	roleStart := pos
	for pos < len(content) && isRabbitAlphaByte(content[pos]) {
		pos++
	}
	if pos == roleStart || pos >= len(content) || content[pos] != ' ' {
		return "", "", "", "", "", false
	}
	roleVal = content[roleStart:pos]
	pos++
	next, ok := consumeRedisDate(content, pos)
	if !ok || next >= len(content) || content[next] != ' ' {
		return "", "", "", "", "", false
	}
	timeVal = content[pos:next]
	pos = next + 1
	sevStart := pos
	for pos < len(content) && !isASCIISpace(content[pos]) {
		pos++
	}
	if pos == sevStart || pos >= len(content) || content[pos] != ' ' {
		return "", "", "", "", "", false
	}
	sevVal = content[sevStart:pos]
	msgVal = content[pos+1:]
	return pidVal, roleVal, timeVal, sevVal, msgVal, true
}

type damengRunner struct {
	timeIdx   int
	statusIdx int
	msgIdx    int
}

func compileDamengRunner(pattern string, nameIndex map[string]int, storage PatternStorageIface) (*damengRunner, bool) {
	if pattern != damengPattern || nameIndex == nil || !defaultPatternDefinitionsMatch(storage, "log_date", "status", "GREEDYDATA") {
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
	msgIdx, ok := nameIndex["msg"]
	if !ok {
		return nil, false
	}
	return &damengRunner{timeIdx: timeIdx, statusIdx: statusIdx, msgIdx: msgIdx}, true
}

func (r *damengRunner) run(dst []string, content string, trimSpace bool) bool {
	timeVal, statusVal, msgVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignStringAt(dst, r.timeIdx, timeVal, trimSpace)
	assignStringAt(dst, r.statusIdx, statusVal, trimSpace)
	assignStringAt(dst, r.msgIdx, msgVal, trimSpace)
	return true
}

func (r *damengRunner) runTyped(dst []any, content string, trimSpace bool, kinds []valueKind) bool {
	timeVal, statusVal, msgVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignTypedAt(dst, kinds, r.timeIdx, timeVal, trimSpace)
	assignTypedAt(dst, kinds, r.statusIdx, statusVal, trimSpace)
	assignTypedAt(dst, kinds, r.msgIdx, msgVal, trimSpace)
	return true
}

func (r *damengRunner) match(content string) (timeVal, statusVal, msgVal string, ok bool) {
	next, ok := consumeDateTimeWithOptionalFraction(content, 0)
	if !ok || next >= len(content) || content[next] != ' ' {
		return "", "", "", false
	}
	timeVal = content[:next]
	pos := next + 1
	if pos >= len(content) || content[pos] != '[' {
		return "", "", "", false
	}
	pos++
	statusStart := pos
	for pos < len(content) && isRabbitAlphaByte(content[pos]) {
		pos++
	}
	if pos == statusStart || !strings.HasPrefix(content[pos:], "]") {
		return "", "", "", false
	}
	statusVal = content[statusStart:pos]
	pos++
	pos = skipASCIISpaces(content, pos)
	if pos >= len(content) {
		return "", "", "", false
	}
	msgVal = content[pos:]
	return timeVal, statusVal, msgVal, true
}

func skipASCIISpaces(s string, pos int) int {
	for pos < len(s) && isASCIISpace(s[pos]) {
		pos++
	}
	return pos
}

func consumeDateTimeWithOptionalFraction(s string, start int) (int, bool) {
	i := start
	if !consumeNDigits(s, &i, 4) || i >= len(s) || s[i] != '-' {
		return 0, false
	}
	i++
	if !consumeOneOrTwoDigits(s, &i) || i >= len(s) || s[i] != '-' {
		return 0, false
	}
	i++
	if !consumeOneOrTwoDigits(s, &i) || i >= len(s) || s[i] != ' ' {
		return 0, false
	}
	i++
	next, ok := consumeTimeOfDay(s, i)
	if !ok {
		return 0, false
	}
	return next, true
}

func consumeDateTimeWithTZ(s string, start int) (int, bool) {
	next, ok := consumeDateTimeWithOptionalFraction(s, start)
	if !ok || next >= len(s) || s[next] != ' ' {
		return 0, false
	}
	if strings.HasPrefix(s[next:], " UTC") || strings.HasPrefix(s[next:], " CST") {
		return next + 4, true
	}
	return 0, false
}

func consumeBracketDateTimeMillis(s string, start int) (int, bool) {
	i := start
	if !consumeNDigits(s, &i, 4) || i >= len(s) || s[i] != '-' {
		return 0, false
	}
	i++
	if !consumeOneOrTwoDigits(s, &i) || i >= len(s) || s[i] != '-' {
		return 0, false
	}
	i++
	if !consumeOneOrTwoDigits(s, &i) || i >= len(s) || s[i] != ' ' {
		return 0, false
	}
	i++
	if !consumeNDigits(s, &i, 2) || i >= len(s) || s[i] != ':' {
		return 0, false
	}
	i++
	if !consumeNDigits(s, &i, 2) || i >= len(s) || s[i] != ':' {
		return 0, false
	}
	i++
	if !consumeNDigits(s, &i, 2) || i >= len(s) || s[i] != ',' {
		return 0, false
	}
	i++
	if !consumeAtLeastOneDigit(s, &i) {
		return 0, false
	}
	return i, true
}

func consumeRedisDate(s string, start int) (int, bool) {
	i := start
	if !consumeOneOrTwoDigits(s, &i) || i >= len(s) || s[i] != ' ' {
		return 0, false
	}
	i++
	monthStart := i
	for i < len(s) && isRabbitAlphaByte(s[i]) {
		i++
	}
	if i == monthStart || !isMonthNameValue(s[monthStart:i]) || i >= len(s) || s[i] != ' ' {
		return 0, false
	}
	i++
	if !consumeNDigits(s, &i, 4) || i >= len(s) || s[i] != ' ' {
		return 0, false
	}
	i++
	next, ok := consumeTimeOfDay(s, i)
	if !ok {
		return 0, false
	}
	return next, true
}
