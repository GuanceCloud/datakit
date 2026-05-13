package grok

import "strings"

type elasticRunnerKind uint8

const (
	elasticDefault elasticRunnerKind = iota + 1
	elasticIndexSlow
	elasticSearchSlow
)

const (
	elasticDefaultPattern    = `^\[%{TIMESTAMP_ISO8601:time}\]\[%{LOGLEVEL:status}%{SPACE}\]\[%{NOTSPACE:name}%{SPACE}\]%{SPACE}(\[%{HOSTNAME:nodeId}\])?.*`
	elasticSearchSlowPattern = `^\[%{TIMESTAMP_ISO8601:time}\]\[%{LOGLEVEL:status}%{SPACE}\]\[i.s.s.(?:query|fetch)%{SPACE}\] (?:\[%{HOSTNAME:nodeId}\] )?\[%{NOTSPACE:index}\]\[%{INT}\] took\[.*\], took_millis\[%{INT:duration}\].*`
)

type elasticDefaultRunner struct {
	timeIdx   int
	statusIdx int
	nameIdx   int
	nodeIDIdx int
}

func compileElasticDefaultRunner(pattern string, nameIndex map[string]int, storage PatternStorageIface) (*elasticDefaultRunner, bool) {
	if pattern != elasticDefaultPattern || nameIndex == nil || !defaultPatternDefinitionsMatch(storage, "TIMESTAMP_ISO8601", "LOGLEVEL", "SPACE", "NOTSPACE", "HOSTNAME") {
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
	nameIdx, ok := nameIndex["name"]
	if !ok {
		return nil, false
	}
	nodeIDIdx := -1
	if idx, ok := nameIndex["nodeId"]; ok {
		nodeIDIdx = idx
	}
	return &elasticDefaultRunner{timeIdx: timeIdx, statusIdx: statusIdx, nameIdx: nameIdx, nodeIDIdx: nodeIDIdx}, true
}

func (r *elasticDefaultRunner) run(dst []string, content string, trimSpace bool) bool {
	timeVal, statusVal, nameVal, nodeVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignStringAt(dst, r.timeIdx, timeVal, trimSpace)
	assignStringAt(dst, r.statusIdx, statusVal, trimSpace)
	assignStringAt(dst, r.nameIdx, nameVal, trimSpace)
	assignStringAt(dst, r.nodeIDIdx, nodeVal, trimSpace)
	return true
}

func (r *elasticDefaultRunner) runTyped(dst []any, content string, trimSpace bool, kinds []valueKind) bool {
	timeVal, statusVal, nameVal, nodeVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignTypedAt(dst, kinds, r.timeIdx, timeVal, trimSpace)
	assignTypedAt(dst, kinds, r.statusIdx, statusVal, trimSpace)
	assignTypedAt(dst, kinds, r.nameIdx, nameVal, trimSpace)
	assignTypedAt(dst, kinds, r.nodeIDIdx, nodeVal, trimSpace)
	return true
}

func (r *elasticDefaultRunner) match(content string) (timeVal, statusVal, nameVal, nodeVal string, ok bool) {
	timeVal, statusVal, pos, ok := consumeElasticPrefix(content)
	if !ok || pos >= len(content) || content[pos] != '[' {
		return "", "", "", "", false
	}
	headerStart := pos + 1
	end := strings.IndexByte(content[headerStart:], ']')
	if end < 0 {
		return "", "", "", "", false
	}
	headerRaw := content[headerStart : headerStart+end]
	if len(headerRaw) == 0 || headerRaw[len(headerRaw)-1] != ' ' {
		return "", "", "", "", false
	}
	nameVal = strings.TrimSpace(headerRaw)
	pos = headerStart + end + 1
	pos = skipASCIISpaces(content, pos)
	nodeVal, _ = consumeOptionalBracketValue(content, pos)
	return timeVal, statusVal, nameVal, nodeVal, true
}

type elasticSearchSlowRunner struct {
	timeIdx     int
	statusIdx   int
	nodeIDIdx   int
	indexIdx    int
	durationIdx int
}

func compileElasticSearchSlowRunner(pattern string, nameIndex map[string]int, storage PatternStorageIface) (*elasticSearchSlowRunner, bool) {
	if pattern != elasticSearchSlowPattern || nameIndex == nil || !defaultPatternDefinitionsMatch(storage, "TIMESTAMP_ISO8601", "LOGLEVEL", "SPACE", "HOSTNAME", "NOTSPACE", "INT") {
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
	indexIdx, ok := nameIndex["index"]
	if !ok {
		return nil, false
	}
	durationIdx, ok := nameIndex["duration"]
	if !ok {
		return nil, false
	}
	nodeIDIdx := -1
	if idx, ok := nameIndex["nodeId"]; ok {
		nodeIDIdx = idx
	}
	return &elasticSearchSlowRunner{timeIdx: timeIdx, statusIdx: statusIdx, nodeIDIdx: nodeIDIdx, indexIdx: indexIdx, durationIdx: durationIdx}, true
}

func (r *elasticSearchSlowRunner) run(dst []string, content string, trimSpace bool) bool {
	timeVal, statusVal, nodeVal, indexVal, durationVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignStringAt(dst, r.timeIdx, timeVal, trimSpace)
	assignStringAt(dst, r.statusIdx, statusVal, trimSpace)
	assignStringAt(dst, r.nodeIDIdx, nodeVal, trimSpace)
	assignStringAt(dst, r.indexIdx, indexVal, trimSpace)
	assignStringAt(dst, r.durationIdx, durationVal, trimSpace)
	return true
}

func (r *elasticSearchSlowRunner) runTyped(dst []any, content string, trimSpace bool, kinds []valueKind) bool {
	timeVal, statusVal, nodeVal, indexVal, durationVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignTypedAt(dst, kinds, r.timeIdx, timeVal, trimSpace)
	assignTypedAt(dst, kinds, r.statusIdx, statusVal, trimSpace)
	assignTypedAt(dst, kinds, r.nodeIDIdx, nodeVal, trimSpace)
	assignTypedAt(dst, kinds, r.indexIdx, indexVal, trimSpace)
	assignTypedAt(dst, kinds, r.durationIdx, durationVal, trimSpace)
	return true
}

func (r *elasticSearchSlowRunner) match(content string) (timeVal, statusVal, nodeVal, indexVal, durationVal string, ok bool) {
	timeVal, statusVal, pos, ok := consumeElasticPrefix(content)
	if !ok || pos >= len(content) || content[pos] != '[' {
		return "", "", "", "", "", false
	}
	headerStart := pos + 1
	end := strings.IndexByte(content[headerStart:], ']')
	if end < 0 {
		return "", "", "", "", "", false
	}
	header := strings.TrimSpace(content[headerStart : headerStart+end])
	if header != "i.s.s.query" && header != "i.s.s.fetch" {
		return "", "", "", "", "", false
	}
	pos = headerStart + end + 1
	pos = skipASCIISpaces(content, pos)
	first, nextPos, ok := consumeBracketValue(content, pos)
	if !ok {
		return "", "", "", "", "", false
	}
	pos = nextPos
	if pos < len(content) && content[pos] == ' ' {
		nodeVal = first
		pos = skipASCIISpaces(content, pos)
		indexVal, pos, ok = consumeBracketValue(content, pos)
		if !ok {
			return "", "", "", "", "", false
		}
	} else {
		indexVal = first
	}
	_, pos, ok = consumeBracketValue(content, pos)
	if !ok {
		return "", "", "", "", "", false
	}
	key := "took_millis["
	durStart := strings.Index(content[pos:], key)
	if durStart < 0 {
		return "", "", "", "", "", false
	}
	pos += durStart + len(key)
	endPos := pos
	for endPos < len(content) && isASCIIDigit(content[endPos]) {
		endPos++
	}
	if endPos == pos || endPos >= len(content) || content[endPos] != ']' {
		return "", "", "", "", "", false
	}
	durationVal = content[pos:endPos]
	return timeVal, statusVal, nodeVal, indexVal, durationVal, true
}

func consumeElasticPrefix(content string) (timeVal, statusVal string, pos int, ok bool) {
	if len(content) < 2 || content[0] != '[' {
		return "", "", 0, false
	}
	next, ok := consumeTimestampISO8601(content, 1)
	if !ok || next >= len(content) || content[next] != ']' || next+1 >= len(content) || content[next+1] != '[' {
		return "", "", 0, false
	}
	timeVal = content[1:next]
	pos = next + 2
	statusStart := pos
	for pos < len(content) && content[pos] != ']' {
		pos++
	}
	if pos == statusStart {
		return "", "", 0, false
	}
	statusVal = strings.TrimSpace(content[statusStart:pos])
	pos++
	pos = skipASCIISpaces(content, pos)
	return timeVal, statusVal, pos, true
}

func consumeBracketValue(content string, pos int) (string, int, bool) {
	if pos >= len(content) || content[pos] != '[' {
		return "", pos, false
	}
	start := pos + 1
	end := strings.IndexByte(content[start:], ']')
	if end < 0 {
		return "", pos, false
	}
	return content[start : start+end], start + end + 1, true
}

type elasticRunner struct {
	kind        elasticRunnerKind
	timeIdx     int
	statusIdx   int
	nameIdx     int
	nodeIDIdx   int
	indexIdx    int
	durationIdx int
}

func compileElasticRunner(pattern string, nameIndex map[string]int, storage PatternStorageIface) (*elasticRunner, bool) {
	if nameIndex == nil || !strings.HasPrefix(pattern, `^\[%{TIMESTAMP_ISO8601:time}\]\[%{LOGLEVEL:status}`) || !defaultPatternDefinitionsMatch(storage, "TIMESTAMP_ISO8601", "LOGLEVEL", "SPACE", "HOSTNAME", "NOTSPACE", "INT") {
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
	r := &elasticRunner{timeIdx: timeIdx, statusIdx: statusIdx, nodeIDIdx: -1, nameIdx: -1, indexIdx: -1, durationIdx: -1}
	if idx, ok := nameIndex["nodeId"]; ok {
		r.nodeIDIdx = idx
	}
	if idx, ok := nameIndex["name"]; ok {
		r.nameIdx = idx
	}
	if idx, ok := nameIndex["index"]; ok {
		r.indexIdx = idx
	}
	if idx, ok := nameIndex["duration"]; ok {
		r.durationIdx = idx
	}
	switch {
	case strings.Contains(pattern, `[i.i.s.index%{SPACE}\]`):
		r.kind = elasticIndexSlow
	default:
		return nil, false
	}
	return r, true
}

func (r *elasticRunner) run(dst []string, content string, trimSpace bool) bool {
	timeVal, statusVal, nameVal, nodeVal, indexVal, durationVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignStringAt(dst, r.timeIdx, timeVal, trimSpace)
	assignStringAt(dst, r.statusIdx, statusVal, trimSpace)
	assignStringAt(dst, r.nameIdx, nameVal, trimSpace)
	assignStringAt(dst, r.nodeIDIdx, nodeVal, trimSpace)
	assignStringAt(dst, r.indexIdx, indexVal, trimSpace)
	assignStringAt(dst, r.durationIdx, durationVal, trimSpace)
	return true
}

func (r *elasticRunner) runTyped(dst []any, content string, trimSpace bool, kinds []valueKind) bool {
	timeVal, statusVal, nameVal, nodeVal, indexVal, durationVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignTypedAt(dst, kinds, r.timeIdx, timeVal, trimSpace)
	assignTypedAt(dst, kinds, r.statusIdx, statusVal, trimSpace)
	assignTypedAt(dst, kinds, r.nameIdx, nameVal, trimSpace)
	assignTypedAt(dst, kinds, r.nodeIDIdx, nodeVal, trimSpace)
	assignTypedAt(dst, kinds, r.indexIdx, indexVal, trimSpace)
	assignTypedAt(dst, kinds, r.durationIdx, durationVal, trimSpace)
	return true
}

func (r *elasticRunner) match(content string) (timeVal, statusVal, nameVal, nodeVal, indexVal, durationVal string, ok bool) {
	if len(content) < 2 || content[0] != '[' {
		return "", "", "", "", "", "", false
	}
	next, ok := consumeTimestampISO8601(content, 1)
	if !ok || next >= len(content) || content[next] != ']' || next+1 >= len(content) || content[next+1] != '[' {
		return "", "", "", "", "", "", false
	}
	timeVal = content[1:next]
	pos := next + 2
	statusStart := pos
	for pos < len(content) && content[pos] != ']' {
		pos++
	}
	if pos == statusStart {
		return "", "", "", "", "", "", false
	}
	statusVal = strings.TrimSpace(content[statusStart:pos])
	pos++
	if pos >= len(content) || content[pos] != '[' {
		return "", "", "", "", "", "", false
	}
	pos++
	headerStart := pos
	for pos < len(content) && content[pos] != ']' {
		pos++
	}
	if pos == headerStart {
		return "", "", "", "", "", "", false
	}
	headerRaw := content[headerStart:pos]
	headerVal := strings.TrimSpace(headerRaw)
	pos++
	pos = skipASCIISpaces(content, pos)

	switch r.kind {
	case elasticDefault:
		if len(headerRaw) == 0 || headerRaw[len(headerRaw)-1] != ' ' {
			return "", "", "", "", "", "", false
		}
		nameVal = headerVal
		if pos < len(content) && content[pos] == '[' {
			nodeStart := pos + 1
			end := strings.IndexByte(content[nodeStart:], ']')
			if end < 0 {
				return "", "", "", "", "", "", false
			}
			nodeVal = content[nodeStart : nodeStart+end]
		}
		return timeVal, statusVal, nameVal, nodeVal, "", "", true
	case elasticIndexSlow:
		if headerVal != "i.i.s.index" {
			return "", "", "", "", "", "", false
		}
		nodeVal, pos = consumeOptionalBracketValue(content, pos)
		if pos >= len(content) || content[pos] != '[' {
			return "", "", "", "", "", "", false
		}
		itemStart := pos + 1
		end := strings.IndexByte(content[itemStart:], ']')
		if end < 0 {
			return "", "", "", "", "", "", false
		}
		item := content[itemStart : itemStart+end]
		slash := strings.IndexByte(item, '/')
		if slash <= 0 {
			return "", "", "", "", "", "", false
		}
		indexVal = item[:slash]
		pos = itemStart + end + 1
	case elasticSearchSlow:
		if headerVal != "i.s.s.query" && headerVal != "i.s.s.fetch" {
			return "", "", "", "", "", "", false
		}
		nodeVal, pos = consumeOptionalBracketValue(content, pos)
		if pos >= len(content) || content[pos] != '[' {
			return "", "", "", "", "", "", false
		}
		indexStart := pos + 1
		end := strings.IndexByte(content[indexStart:], ']')
		if end < 0 {
			return "", "", "", "", "", "", false
		}
		indexVal = content[indexStart : indexStart+end]
		pos = indexStart + end + 1
		if pos >= len(content) || content[pos] != '[' {
			return "", "", "", "", "", "", false
		}
		end = strings.IndexByte(content[pos+1:], ']')
		if end < 0 {
			return "", "", "", "", "", "", false
		}
		pos = pos + 1 + end + 1
	}
	key := "took_millis["
	durStart := strings.Index(content[pos:], key)
	if durStart < 0 {
		return "", "", "", "", "", "", false
	}
	pos += durStart + len(key)
	end := pos
	for end < len(content) && isASCIIDigit(content[end]) {
		end++
	}
	if end == pos || end >= len(content) || content[end] != ']' {
		return "", "", "", "", "", "", false
	}
	durationVal = content[pos:end]
	return timeVal, statusVal, nameVal, nodeVal, indexVal, durationVal, true
}

func consumeOptionalBracketValue(content string, pos int) (string, int) {
	pos = skipASCIISpaces(content, pos)
	if pos >= len(content) || content[pos] != '[' {
		return "", pos
	}
	start := pos + 1
	end := strings.IndexByte(content[start:], ']')
	if end < 0 {
		return "", pos
	}
	value := content[start : start+end]
	pos = start + end + 1
	pos = skipASCIISpaces(content, pos)
	return value, pos
}

type nginxErrorRunnerKind uint8

const (
	nginxErrorDetailed nginxErrorRunnerKind = iota + 1
	nginxErrorSimple
	nginxErrorBare
)

type nginxErrorRunner struct {
	kind        nginxErrorRunnerKind
	timeIdx     int
	statusIdx   int
	msgIdx      int
	clientIdx   int
	serverIdx   int
	methodIdx   int
	urlIdx      int
	versionIdx  int
	upstreamIdx int
	hostIdx     int
	clientKind  structuredKind
	serverKind  structuredKind
	methodKind  structuredKind
	versionKind structuredKind
	hostKind    structuredKind
}

func compileNginxErrorRunner(pattern string, nameIndex map[string]int, storage PatternStorageIface) (*nginxErrorRunner, bool) {
	if nameIndex == nil || !nginxErrorPatternHasDate(pattern) || !nginxErrorPatternHasLevel(pattern) || !nginxErrorPatternDefinitionsMatch(pattern, storage) {
		return nil, false
	}
	r := &nginxErrorRunner{timeIdx: -1, statusIdx: -1, msgIdx: -1, clientIdx: -1, serverIdx: -1, methodIdx: -1, urlIdx: -1, versionIdx: -1, upstreamIdx: -1, hostIdx: -1}
	var ok bool
	if r.timeIdx, ok = nameIndex["time"]; !ok {
		return nil, false
	}
	if r.statusIdx = lookupFirstNameIndex(nameIndex, "status", "level"); r.statusIdx < 0 {
		return nil, false
	}
	if r.msgIdx, ok = nameIndex["msg"]; !ok {
		return nil, false
	}
	r.clientIdx = lookupFirstNameIndex(nameIndex, "client_ip", "client")
	r.serverIdx = lookupNameIndex(nameIndex, "server")
	r.methodIdx = lookupFirstNameIndex(nameIndex, "http_method", "method")
	r.urlIdx = lookupFirstNameIndex(nameIndex, "http_url", "path")
	r.versionIdx = lookupNameIndex(nameIndex, "http_version")
	r.upstreamIdx = lookupNameIndex(nameIndex, "upstream")
	r.hostIdx = lookupFirstNameIndex(nameIndex, "ip_or_host", "host")
	r.clientKind = structuredNotSpace
	if strings.Contains(pattern, `client: %{IPORHOST:client}`) {
		r.clientKind = structuredIPOrHost
	}
	r.serverKind = structuredNotSpace
	r.methodKind = structuredWord
	r.versionKind = structuredNumber
	r.hostKind = structuredNotSpace
	switch {
	case pattern == `%{date2:time} \[%{LOGLEVEL:status}\] %{GREEDYDATA:msg}`:
		r.kind = nginxErrorBare
	case nginxErrorPatternHasDetailedShape(pattern):
		if strings.Contains(pattern, `upstream:`) || strings.Contains(pattern, `%{UPBLOCK}`) {
			r.kind = nginxErrorDetailed
		} else {
			r.kind = nginxErrorSimple
		}
	default:
		return nil, false
	}
	return r, true
}

func nginxErrorPatternHasDate(pattern string) bool {
	return strings.Contains(pattern, `%{date2:time}`) ||
		strings.Contains(pattern, `%{APPDATE:time}`)
}

func nginxErrorPatternHasLevel(pattern string) bool {
	return strings.Contains(pattern, `%{LOGLEVEL:status}`) ||
		strings.Contains(pattern, `%{LOGLEVEL:level}`)
}

func nginxErrorPatternHasDetailedShape(pattern string) bool {
	return strings.Contains(pattern, `client: %{NOTSPACE:client_ip}`) ||
		strings.Contains(pattern, `client: %{IPORHOST:client}`)
}

func nginxErrorPatternDefinitionsMatch(pattern string, storage PatternStorageIface) bool {
	names := []string{"LOGLEVEL", "GREEDYDATA"}
	switch {
	case strings.Contains(pattern, `%{date2:time}`):
		names = append(names, "date2")
	case strings.Contains(pattern, `%{APPDATE:time}`):
		names = append(names, "APPDATE")
	default:
		return false
	}
	if strings.Contains(pattern, `%{NOTSPACE:client_ip}`) || strings.Contains(pattern, `%{NOTSPACE:server}`) || strings.Contains(pattern, `%{NOTSPACE:host}`) {
		names = append(names, "NOTSPACE")
	}
	if strings.Contains(pattern, `%{IPORHOST:client}`) {
		names = append(names, "IPORHOST")
	}
	if strings.Contains(pattern, `%{WORD:`) {
		names = append(names, "WORD")
	}
	if strings.Contains(pattern, `%{NUMBER:`) {
		names = append(names, "NUMBER")
	}
	if strings.Contains(pattern, `%{UPBLOCK}`) {
		names = append(names, "UPBLOCK")
	}
	return defaultPatternDefinitionsMatch(storage, names...)
}

func lookupFirstNameIndex(nameIndex map[string]int, names ...string) int {
	for _, name := range names {
		if idx, ok := nameIndex[name]; ok {
			return idx
		}
	}
	return -1
}

func (r *nginxErrorRunner) run(dst []string, content string, trimSpace bool) bool {
	timeVal, statusVal, msgVal, clientVal, serverVal, methodVal, urlVal, versionVal, upstreamVal, hostVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignStringAt(dst, r.timeIdx, timeVal, trimSpace)
	assignStringAt(dst, r.statusIdx, statusVal, trimSpace)
	assignStringAt(dst, r.msgIdx, msgVal, trimSpace)
	assignStringAt(dst, r.clientIdx, clientVal, trimSpace)
	assignStringAt(dst, r.serverIdx, serverVal, trimSpace)
	assignStringAt(dst, r.methodIdx, methodVal, trimSpace)
	assignStringAt(dst, r.urlIdx, urlVal, trimSpace)
	assignStringAt(dst, r.versionIdx, versionVal, trimSpace)
	assignStringAt(dst, r.upstreamIdx, upstreamVal, trimSpace)
	assignStringAt(dst, r.hostIdx, hostVal, trimSpace)
	return true
}

func (r *nginxErrorRunner) runTyped(dst []any, content string, trimSpace bool, kinds []valueKind) bool {
	timeVal, statusVal, msgVal, clientVal, serverVal, methodVal, urlVal, versionVal, upstreamVal, hostVal, ok := r.match(content)
	if !ok {
		return false
	}
	assignTypedAt(dst, kinds, r.timeIdx, timeVal, trimSpace)
	assignTypedAt(dst, kinds, r.statusIdx, statusVal, trimSpace)
	assignTypedAt(dst, kinds, r.msgIdx, msgVal, trimSpace)
	assignTypedAt(dst, kinds, r.clientIdx, clientVal, trimSpace)
	assignTypedAt(dst, kinds, r.serverIdx, serverVal, trimSpace)
	assignTypedAt(dst, kinds, r.methodIdx, methodVal, trimSpace)
	assignTypedAt(dst, kinds, r.urlIdx, urlVal, trimSpace)
	assignTypedAt(dst, kinds, r.versionIdx, versionVal, trimSpace)
	assignTypedAt(dst, kinds, r.upstreamIdx, upstreamVal, trimSpace)
	assignTypedAt(dst, kinds, r.hostIdx, hostVal, trimSpace)
	return true
}

func (r *nginxErrorRunner) match(content string) (timeVal, statusVal, msgVal, clientVal, serverVal, methodVal, urlVal, versionVal, upstreamVal, hostVal string, ok bool) {
	next, ok := consumeSlashDateTime(content, 0)
	if !ok || next >= len(content) || content[next] != ' ' {
		return "", "", "", "", "", "", "", "", "", "", false
	}
	timeVal = content[:next]
	pos := next + 1
	if pos >= len(content) || content[pos] != '[' {
		return "", "", "", "", "", "", "", "", "", "", false
	}
	statusStart := pos + 1
	end := strings.IndexByte(content[statusStart:], ']')
	if end < 0 {
		return "", "", "", "", "", "", "", "", "", "", false
	}
	statusVal = content[statusStart : statusStart+end]
	if !structuredSegmentMatchesKind(statusVal, structuredLogLevel) {
		return "", "", "", "", "", "", "", "", "", "", false
	}
	pos = statusStart + end + 1
	if pos >= len(content) || content[pos] != ' ' {
		return "", "", "", "", "", "", "", "", "", "", false
	}
	pos++
	if r.kind == nginxErrorBare {
		if pos >= len(content) {
			return "", "", "", "", "", "", "", "", "", "", false
		}
		msgVal = content[pos:]
		return timeVal, statusVal, msgVal, "", "", "", "", "", "", "", true
	}
	marker := `, client: `
	msgEnd := strings.Index(content[pos:], marker)
	if msgEnd < 0 {
		return "", "", "", "", "", "", "", "", "", "", false
	}
	msgVal = content[pos : pos+msgEnd]
	pos += msgEnd + len(marker)
	clientEnd := strings.Index(content[pos:], `, server: `)
	if clientEnd < 0 {
		return "", "", "", "", "", "", "", "", "", "", false
	}
	clientVal = content[pos : pos+clientEnd]
	if !structuredSegmentMatchesKind(clientVal, r.clientKind) {
		return "", "", "", "", "", "", "", "", "", "", false
	}
	pos += clientEnd + len(`, server: `)
	serverEnd := strings.Index(content[pos:], `, request: "`)
	if serverEnd < 0 {
		return "", "", "", "", "", "", "", "", "", "", false
	}
	serverVal = content[pos : pos+serverEnd]
	if !structuredSegmentMatchesKind(serverVal, r.serverKind) {
		return "", "", "", "", "", "", "", "", "", "", false
	}
	pos += serverEnd + len(`, request: "`)
	httpIdx := strings.Index(content[pos:], ` HTTP/`)
	if httpIdx < 0 {
		return "", "", "", "", "", "", "", "", "", "", false
	}
	request := content[pos : pos+httpIdx]
	space := strings.IndexByte(request, ' ')
	if space < 0 {
		return "", "", "", "", "", "", "", "", "", "", false
	}
	methodVal = request[:space]
	if !structuredSegmentMatchesKind(methodVal, r.methodKind) {
		return "", "", "", "", "", "", "", "", "", "", false
	}
	urlVal = request[space+1:]
	pos += httpIdx + len(` HTTP/`)
	versionEnd := strings.IndexByte(content[pos:], '"')
	if versionEnd < 0 {
		return "", "", "", "", "", "", "", "", "", "", false
	}
	versionVal = content[pos : pos+versionEnd]
	if !structuredSegmentMatchesKind(versionVal, r.versionKind) {
		return "", "", "", "", "", "", "", "", "", "", false
	}
	pos += versionEnd + 1
	if strings.HasPrefix(content[pos:], `, upstream: "`) {
		pos += len(`, upstream: "`)
		upEnd := strings.Index(content[pos:], `", host: "`)
		if upEnd < 0 {
			return "", "", "", "", "", "", "", "", "", "", false
		}
		upstreamVal = content[pos : pos+upEnd]
		pos += upEnd + len(`", host: "`)
	} else if strings.HasPrefix(content[pos:], `, host: "`) {
		pos += len(`, host: "`)
	} else {
		return "", "", "", "", "", "", "", "", "", "", false
	}
	hostEnd := strings.IndexByte(content[pos:], '"')
	if hostEnd < 0 {
		return "", "", "", "", "", "", "", "", "", "", false
	}
	hostVal = content[pos : pos+hostEnd]
	if !structuredSegmentMatchesKind(hostVal, r.hostKind) {
		return "", "", "", "", "", "", "", "", "", "", false
	}
	return timeVal, statusVal, msgVal, clientVal, serverVal, methodVal, urlVal, versionVal, upstreamVal, hostVal, true
}

func consumeSlashDateTime(s string, start int) (int, bool) {
	i := start
	if !consumeNDigits(s, &i, 4) || i >= len(s) {
		return 0, false
	}
	sep := s[i]
	if sep != '/' && sep != '.' {
		return 0, false
	}
	i++
	if !consumeTwoDigitRange(s, &i, 1, 12, true) || i >= len(s) || s[i] != sep {
		return 0, false
	}
	i++
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
