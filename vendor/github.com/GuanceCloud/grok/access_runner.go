package grok

import "strings"

type accessRunnerMode uint8

const (
	accessRunnerModeSimple accessRunnerMode = iota + 1
	accessRunnerModeApacheFixture
)

type accessLogRunner struct {
	mode        accessRunnerMode
	clientIdx   int
	identIdx    int
	authIdx     int
	timeIdx     int
	methodIdx   int
	urlIdx      int
	versionIdx  int
	statusIdx   int
	bytesIdx    int
	clientKind  structuredKind
	identKind   structuredKind
	authKind    structuredKind
	methodKind  structuredKind
	urlKind     structuredKind
	statusKind  structuredKind
	bytesKind   structuredKind
	versionKind structuredKind
}

func compileAccessRunner(steps []structuredStep) (*accessLogRunner, bool) {
	if matchesApacheFixtureAccessShape(steps) {
		return &accessLogRunner{
			mode:        accessRunnerModeApacheFixture,
			clientIdx:   steps[0].parser.dstIndex,
			clientKind:  steps[0].parser.kind,
			timeIdx:     steps[1].parser.dstIndex,
			methodIdx:   steps[2].parser.dstIndex,
			methodKind:  steps[2].parser.kind,
			urlIdx:      steps[4].parser.dstIndex,
			urlKind:     steps[4].parser.kind,
			versionIdx:  steps[5].parser.dstIndex,
			versionKind: steps[5].parser.kind,
			statusIdx:   steps[6].parser.dstIndex,
			statusKind:  steps[6].parser.kind,
			identIdx:    -1,
			authIdx:     -1,
			bytesIdx:    -1,
		}, true
	}

	if matchesSimpleAccessShape(steps) {
		return &accessLogRunner{
			mode:        accessRunnerModeSimple,
			clientIdx:   steps[0].parser.dstIndex,
			clientKind:  steps[0].parser.kind,
			identIdx:    steps[1].parser.dstIndex,
			identKind:   steps[1].parser.kind,
			authIdx:     steps[2].parser.dstIndex,
			authKind:    steps[2].parser.kind,
			timeIdx:     steps[3].parser.dstIndex,
			methodIdx:   steps[4].parser.dstIndex,
			methodKind:  steps[4].parser.kind,
			urlIdx:      steps[6].parser.dstIndex,
			urlKind:     steps[6].parser.kind,
			versionIdx:  steps[7].parser.dstIndex,
			versionKind: steps[7].parser.kind,
			statusIdx:   steps[8].parser.dstIndex,
			statusKind:  steps[8].parser.kind,
			bytesIdx:    steps[10].parser.dstIndex,
			bytesKind:   steps[10].parser.kind,
		}, true
	}

	return nil, false
}

func (r *accessLogRunner) run(dst []string, content string, trimSpace bool) bool {
	match, ok := r.match(content)
	if !ok {
		return false
	}
	r.assignStrings(dst, match, trimSpace)
	return true
}

func (r *accessLogRunner) runTyped(dst []any, content string, trimSpace bool, kinds []valueKind) bool {
	match, ok := r.match(content)
	if !ok {
		return false
	}
	r.assignTyped(dst, match, trimSpace, kinds)
	return true
}

type accessMatch struct {
	client  string
	ident   string
	auth    string
	time    string
	method  string
	url     string
	version string
	status  string
	bytes   string
}

func (r *accessLogRunner) match(content string) (accessMatch, bool) {
	switch r.mode {
	case accessRunnerModeSimple:
		return r.matchSimple(content)
	case accessRunnerModeApacheFixture:
		return r.matchApacheFixture(content)
	default:
		return accessMatch{}, false
	}
}

func (r *accessLogRunner) matchSimple(content string) (accessMatch, bool) {
	var out accessMatch
	pos := 0

	nextSpace := strings.IndexByte(content, ' ')
	if nextSpace <= 0 {
		return accessMatch{}, false
	}
	out.client = content[:nextSpace]
	if !structuredSegmentMatchesKind(out.client, r.clientKind) {
		return accessMatch{}, false
	}
	pos = nextSpace + 1

	out.ident, pos = consumeAccessTokenAt(content, pos, r.identKind)
	if pos <= 0 {
		return accessMatch{}, false
	}
	out.auth, pos = consumeAccessTokenAt(content, pos, r.authKind)
	if pos <= 0 {
		return accessMatch{}, false
	}

	if !strings.HasPrefix(content[pos:], "[") {
		return accessMatch{}, false
	}
	pos++
	next, ok := consumeHTTPDate(content, pos)
	if !ok || next == pos {
		return accessMatch{}, false
	}
	out.time = content[pos:next]
	pos = next
	if !strings.HasPrefix(content[pos:], "] \"") {
		return accessMatch{}, false
	}
	pos += len(`] "`)

	nextSpace = strings.IndexByte(content[pos:], ' ')
	if nextSpace <= 0 {
		return accessMatch{}, false
	}
	out.method = content[pos : pos+nextSpace]
	if !structuredSegmentMatchesKind(out.method, r.methodKind) {
		return accessMatch{}, false
	}
	pos += nextSpace + 1

	relHTTP := lastLiteralIndex(content[pos:], ` HTTP/`)
	if relHTTP <= 0 {
		return accessMatch{}, false
	}
	out.url = content[pos : pos+relHTTP]
	if !structuredSegmentMatchesKind(out.url, r.urlKind) {
		return accessMatch{}, false
	}
	pos += relHTTP + len(` HTTP/`)

	out.version, pos, ok = consumeNumericAt(content, pos, r.versionKind)
	if !ok || !strings.HasPrefix(content[pos:], `" `) {
		return accessMatch{}, false
	}
	pos += 2

	out.status, pos, ok = consumeNumericAt(content, pos, r.statusKind)
	if !ok || pos >= len(content) || content[pos] != ' ' {
		return accessMatch{}, false
	}
	pos++

	out.bytes, pos, ok = consumeNumericAt(content, pos, r.bytesKind)
	if !ok {
		return accessMatch{}, false
	}
	return out, true
}

func (r *accessLogRunner) matchApacheFixture(content string) (accessMatch, bool) {
	var out accessMatch
	rel := strings.Index(content, ` - - [`)
	if rel <= 0 {
		return accessMatch{}, false
	}
	out.client = content[:rel]
	if !structuredSegmentMatchesKind(out.client, r.clientKind) {
		return accessMatch{}, false
	}
	pos := rel + len(` - - [`)

	next, ok := consumeHTTPDate(content, pos)
	if !ok || next == pos {
		return accessMatch{}, false
	}
	out.time = content[pos:next]
	pos = next
	if !strings.HasPrefix(content[pos:], `] "`) {
		return accessMatch{}, false
	}
	pos += len(`] "`)

	nextSpace := strings.IndexByte(content[pos:], ' ')
	if nextSpace <= 0 {
		return accessMatch{}, false
	}
	out.method = content[pos : pos+nextSpace]
	if !structuredSegmentMatchesKind(out.method, r.methodKind) {
		return accessMatch{}, false
	}
	pos += nextSpace + 1

	relHTTP := lastLiteralIndex(content[pos:], ` HTTP/`)
	if relHTTP <= 0 {
		return accessMatch{}, false
	}
	out.url = content[pos : pos+relHTTP]
	if !structuredSegmentMatchesKind(out.url, r.urlKind) {
		return accessMatch{}, false
	}
	pos += relHTTP + len(` HTTP/`)

	out.version, pos, ok = consumeNumericAt(content, pos, r.versionKind)
	if !ok || !strings.HasPrefix(content[pos:], `" `) {
		return accessMatch{}, false
	}
	pos += 2

	out.status, pos, ok = consumeNumericAt(content, pos, r.statusKind)
	if !ok || pos >= len(content) || content[pos] != ' ' {
		return accessMatch{}, false
	}
	return out, true
}

func consumeAccessTokenAt(content string, pos int, kind structuredKind) (string, int) {
	if pos >= len(content) {
		return "", -1
	}
	end := pos
	for end < len(content) && content[end] != ' ' {
		end++
	}
	if end == pos || end >= len(content) {
		return "", -1
	}
	token := content[pos:end]
	if !structuredSegmentMatchesKind(token, kind) {
		return "", -1
	}
	return token, end + 1
}

func consumeTokenAt(content string, pos int) (string, int) {
	return consumeAccessTokenAt(content, pos, structuredNotSpace)
}

func structuredSegmentMatchesKind(segment string, kind structuredKind) bool {
	if segment == "" {
		return false
	}
	parser := structuredParser{kind: kind}
	next, _, ok := parser.consume(segment, 0)
	return ok && next == len(segment)
}

func consumeNumericAt(content string, pos int, kind structuredKind) (string, int, bool) {
	var (
		value string
		next  int
		ok    bool
	)
	switch kind {
	case structuredInt:
		value, next, ok = sliceInt(content, pos)
	case structuredPosInt:
		value, next, ok = slicePosInt(content, pos)
	case structuredNonNegInt:
		value, next, ok = sliceNonNegInt(content, pos)
	default:
		value, next, ok = sliceNumber(content, pos)
	}
	return value, next, ok
}

func (r *accessLogRunner) assignStrings(dst []string, match accessMatch, trimSpace bool) {
	assignStringAt(dst, r.clientIdx, match.client, trimSpace)
	assignStringAt(dst, r.identIdx, match.ident, trimSpace)
	assignStringAt(dst, r.authIdx, match.auth, trimSpace)
	assignStringAt(dst, r.timeIdx, match.time, trimSpace)
	assignStringAt(dst, r.methodIdx, match.method, trimSpace)
	assignStringAt(dst, r.urlIdx, match.url, trimSpace)
	assignStringAt(dst, r.versionIdx, match.version, trimSpace)
	assignStringAt(dst, r.statusIdx, match.status, trimSpace)
	assignStringAt(dst, r.bytesIdx, match.bytes, trimSpace)
}

func (r *accessLogRunner) assignTyped(dst []any, match accessMatch, trimSpace bool, kinds []valueKind) {
	assignTypedAt(dst, kinds, r.clientIdx, match.client, trimSpace)
	assignTypedAt(dst, kinds, r.identIdx, match.ident, trimSpace)
	assignTypedAt(dst, kinds, r.authIdx, match.auth, trimSpace)
	assignTypedAt(dst, kinds, r.timeIdx, match.time, trimSpace)
	assignTypedAt(dst, kinds, r.methodIdx, match.method, trimSpace)
	assignTypedAt(dst, kinds, r.urlIdx, match.url, trimSpace)
	assignTypedAt(dst, kinds, r.versionIdx, match.version, trimSpace)
	assignTypedAt(dst, kinds, r.statusIdx, match.status, trimSpace)
	assignTypedAt(dst, kinds, r.bytesIdx, match.bytes, trimSpace)
}

func assignStringAt(dst []string, idx int, value string, trimSpace bool) {
	if idx < 0 {
		return
	}
	dst[idx] = maybeTrim(value, trimSpace)
}

func assignTypedAt(dst []any, kinds []valueKind, idx int, value string, trimSpace bool) {
	if idx < 0 {
		return
	}
	dst[idx] = castStructuredValue(value, trimSpace, kinds[idx])
}
