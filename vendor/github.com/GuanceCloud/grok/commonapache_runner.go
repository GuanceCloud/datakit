package grok

import "strings"

type commonApacheRunner struct {
	clientIdx      int
	identIdx       int
	authIdx        int
	timestampIdx   int
	verbIdx        int
	requestIdx     int
	httpVersionIdx int
	rawRequestIdx  int
	responseIdx    int
	bytesIdx       int
}

func compileCommonApacheRunner(pattern string, nameIndex map[string]int, storage PatternStorageIface) (*commonApacheRunner, bool) {
	if pattern != "%{COMMONAPACHELOG}" || len(nameIndex) == 0 {
		return nil, false
	}
	if !defaultPatternDefinitionsMatch(storage, "COMMONAPACHELOG") {
		return nil, false
	}
	runner := &commonApacheRunner{
		clientIdx:      lookupNameIndex(nameIndex, "clientip"),
		identIdx:       lookupNameIndex(nameIndex, "ident"),
		authIdx:        lookupNameIndex(nameIndex, "auth"),
		timestampIdx:   lookupNameIndex(nameIndex, "timestamp"),
		verbIdx:        lookupNameIndex(nameIndex, "verb"),
		requestIdx:     lookupNameIndex(nameIndex, "request"),
		httpVersionIdx: lookupNameIndex(nameIndex, "httpversion"),
		rawRequestIdx:  lookupNameIndex(nameIndex, "rawrequest"),
		responseIdx:    lookupNameIndex(nameIndex, "response"),
		bytesIdx:       lookupNameIndex(nameIndex, "bytes"),
	}
	if runner.clientIdx < 0 || runner.identIdx < 0 || runner.authIdx < 0 ||
		runner.timestampIdx < 0 || runner.verbIdx < 0 || runner.requestIdx < 0 ||
		runner.httpVersionIdx < 0 || runner.rawRequestIdx < 0 ||
		runner.responseIdx < 0 || runner.bytesIdx < 0 {
		return nil, false
	}
	return runner, true
}

func lookupNameIndex(nameIndex map[string]int, key string) int {
	if idx, ok := nameIndex[key]; ok {
		return idx
	}
	return -1
}

func (r *commonApacheRunner) run(dst []string, content string, trimSpace bool) bool {
	match, ok := r.match(content)
	if !ok {
		return false
	}
	assignStringAt(dst, r.clientIdx, match.client, trimSpace)
	assignStringAt(dst, r.identIdx, match.ident, trimSpace)
	assignStringAt(dst, r.authIdx, match.auth, trimSpace)
	assignStringAt(dst, r.timestampIdx, match.timestamp, trimSpace)
	assignStringAt(dst, r.verbIdx, match.verb, trimSpace)
	assignStringAt(dst, r.requestIdx, match.request, trimSpace)
	assignStringAt(dst, r.httpVersionIdx, match.httpVersion, trimSpace)
	assignStringAt(dst, r.rawRequestIdx, match.rawRequest, trimSpace)
	assignStringAt(dst, r.responseIdx, match.response, trimSpace)
	assignStringAt(dst, r.bytesIdx, match.bytes, trimSpace)
	return true
}

func (r *commonApacheRunner) runTyped(dst []any, content string, trimSpace bool, kinds []valueKind) bool {
	match, ok := r.match(content)
	if !ok {
		return false
	}
	assignTypedAt(dst, kinds, r.clientIdx, match.client, trimSpace)
	assignTypedAt(dst, kinds, r.identIdx, match.ident, trimSpace)
	assignTypedAt(dst, kinds, r.authIdx, match.auth, trimSpace)
	assignTypedAt(dst, kinds, r.timestampIdx, match.timestamp, trimSpace)
	assignTypedAt(dst, kinds, r.verbIdx, match.verb, trimSpace)
	assignTypedAt(dst, kinds, r.requestIdx, match.request, trimSpace)
	assignTypedAt(dst, kinds, r.httpVersionIdx, match.httpVersion, trimSpace)
	assignTypedAt(dst, kinds, r.rawRequestIdx, match.rawRequest, trimSpace)
	assignTypedAt(dst, kinds, r.responseIdx, match.response, trimSpace)
	assignTypedAt(dst, kinds, r.bytesIdx, match.bytes, trimSpace)
	return true
}

type commonApacheMatch struct {
	client      string
	ident       string
	auth        string
	timestamp   string
	verb        string
	request     string
	httpVersion string
	rawRequest  string
	response    string
	bytes       string
}

func (r *commonApacheRunner) match(content string) (commonApacheMatch, bool) {
	var out commonApacheMatch
	pos := 0

	out.client, pos = consumeTokenAt(content, pos)
	if pos <= 0 {
		return commonApacheMatch{}, false
	}
	out.ident, pos = consumeTokenAt(content, pos)
	if pos <= 0 {
		return commonApacheMatch{}, false
	}
	out.auth, pos = consumeTokenAt(content, pos)
	if pos <= 0 || !strings.HasPrefix(content[pos:], "[") {
		return commonApacheMatch{}, false
	}
	pos++

	next, ok := consumeHTTPDate(content, pos)
	if !ok || next == pos || !strings.HasPrefix(content[next:], `] "`) {
		return commonApacheMatch{}, false
	}
	out.timestamp = content[pos:next]
	pos = next + 3

	endQuote := strings.IndexByte(content[pos:], '"')
	if endQuote < 0 {
		return commonApacheMatch{}, false
	}
	requestLine := content[pos : pos+endQuote]
	pos += endQuote + 1
	if pos >= len(content) || content[pos] != ' ' {
		return commonApacheMatch{}, false
	}
	pos++

	if !parseCommonApacheRequestLine(requestLine, &out) {
		return commonApacheMatch{}, false
	}

	out.response, pos, ok = consumeNumericAt(content, pos, structuredNumber)
	if !ok || pos >= len(content) || content[pos] != ' ' {
		return commonApacheMatch{}, false
	}
	pos++

	if pos < len(content) && content[pos] == '-' {
		out.bytes = ""
		pos++
	} else {
		out.bytes, pos, ok = consumeNumericAt(content, pos, structuredNumber)
		if !ok {
			return commonApacheMatch{}, false
		}
	}

	if pos != len(content) {
		return commonApacheMatch{}, false
	}
	return out, true
}

func parseCommonApacheRequestLine(requestLine string, out *commonApacheMatch) bool {
	if requestLine == "" {
		return false
	}
	if requestLine == "-" {
		out.rawRequest = "-"
		return true
	}

	firstSpace := strings.IndexByte(requestLine, ' ')
	if firstSpace <= 0 {
		out.rawRequest = requestLine
		return true
	}
	method := requestLine[:firstSpace]
	if !isApacheWord(method) {
		out.rawRequest = requestLine
		return true
	}

	rest := requestLine[firstSpace+1:]
	if rest == "" {
		out.rawRequest = requestLine
		return true
	}

	httpPos := strings.LastIndex(rest, " HTTP/")
	if httpPos < 0 {
		out.verb = method
		out.request = rest
		return true
	}
	request := rest[:httpPos]
	if request == "" {
		out.rawRequest = requestLine
		return true
	}
	version := rest[httpPos+len(" HTTP/"):]
	if version == "" {
		out.rawRequest = requestLine
		return true
	}
	if _, _, ok := consumeNumericAt(version, 0, structuredNumber); !ok {
		out.rawRequest = requestLine
		return true
	}

	out.verb = method
	out.request = request
	out.httpVersion = version
	return true
}
