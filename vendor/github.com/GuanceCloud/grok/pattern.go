package grok

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

const (
	GTypeStr    = "str"
	GTypeString = "string"
	GTypeInt    = "int"
	GTypeFloat  = "float"
	GTypeBool   = "bool"
)

type GrokPattern struct {
	pattern      string
	denormalized string
	varbType     map[string]string
}

type patternRef struct {
	raw     string
	syntax  string
	alias   string
	varType string
}

type compiledRegexpMeta struct {
	re            *regexp.Regexp
	filter        *regexpFilter
	subMatchNames SubMatchName
	nameIndex     map[string]int
	prefilter     *regexpPrefilter
	multiFilter   multiPatternFilter
}

var compiledRegexpCache sync.Map

type PatternStorageIface interface {
	GetPattern(string) (*GrokPattern, bool)
	SetPattern(string, *GrokPattern)
}

type PatternStorage []map[string]*GrokPattern

func (p PatternStorage) GetPattern(pattern string) (*GrokPattern, bool) {
	for _, v := range p {
		if gp, ok := v[pattern]; ok {
			return gp, ok
		}
	}
	return nil, false
}

func (p PatternStorage) SetPattern(patternAlias string, gp *GrokPattern) {
	if len(p) > 0 {
		p[len(p)-1][patternAlias] = gp
	}
}

func (g *GrokPattern) Pattern() string {
	return g.pattern
}

func (g *GrokPattern) Denormalized() string {
	return g.denormalized
}

func (g *GrokPattern) TypedVar() map[string]string {
	ret := map[string]string{}
	for k, v := range g.varbType {
		ret[k] = v
	}
	return ret
}

func isWordByte(b byte) bool {
	return b == '_' || ('0' <= b && b <= '9') || ('A' <= b && b <= 'Z') || ('a' <= b && b <= 'z')
}

func isPatternName(s string) bool {
	if len(s) == 0 {
		return false
	}

	for i := 0; i < len(s); i++ {
		if isWordByte(s[i]) {
			continue
		}
		if (s[i] == '.' || s[i] == '-') && i > 0 && i < len(s)-1 && isWordByte(s[i-1]) && isWordByte(s[i+1]) {
			continue
		}
		return false
	}

	return true
}

func isPatternAlias(s string) bool {
	if len(s) == 0 {
		return false
	}

	for i := 0; i < len(s); i++ {
		if isWordByte(s[i]) || s[i] == '.' || s[i] == '-' {
			continue
		}
		return false
	}

	return true
}

func sanitizeAlias(alias string) string {
	buf := make([]byte, len(alias))
	for i := 0; i < len(alias); i++ {
		if isWordByte(alias[i]) {
			buf[i] = alias[i]
		} else {
			buf[i] = '_'
		}
	}
	return string(buf)
}

func parsePatternRef(spec string) (patternRef, bool, error) {
	ref := patternRef{raw: spec}

	firstSep := strings.IndexByte(spec, ':')
	if firstSep == -1 {
		if !isPatternName(spec) {
			return patternRef{}, false, nil
		}
		ref.syntax = spec
		return ref, true, nil
	}

	ref.syntax = spec[:firstSep]
	if !isPatternName(ref.syntax) {
		return patternRef{}, false, nil
	}

	rest := spec[firstSep+1:]
	if len(rest) == 0 {
		return patternRef{}, false, nil
	}

	secondSep := strings.IndexByte(rest, ':')
	if secondSep == -1 {
		if !isPatternAlias(rest) {
			return patternRef{}, false, nil
		}
		ref.alias = sanitizeAlias(rest)
		return ref, true, nil
	}

	alias := rest[:secondSep]
	if !isPatternAlias(alias) {
		return patternRef{}, false, nil
	}
	ref.alias = sanitizeAlias(alias)

	ref.varType = rest[secondSep+1:]
	if len(ref.varType) == 0 || !isPatternAlias(ref.varType) {
		return patternRef{}, false, nil
	}

	switch ref.varType {
	case GTypeString, GTypeStr, GTypeInt, GTypeFloat, GTypeBool:
		return ref, true, nil
	default:
		return patternRef{}, true, fmt.Errorf("invalid varb data type: `%s`", ref.varType)
	}
}

func normalizeAnonymousCaptures(raw string) string {
	if len(raw) < 2 {
		return raw
	}

	for i := 0; i < len(raw); i++ {
		if raw[i] != '\\' || i+1 >= len(raw) {
			continue
		}
		if raw[i+1] >= '1' && raw[i+1] <= '9' {
			return raw
		}
		i++
	}

	var builder strings.Builder
	builder.Grow(len(raw))
	inClass := false
	classStart := false
	inQuote := false

	for i := 0; i < len(raw); i++ {
		switch raw[i] {
		case '\\':
			builder.WriteByte(raw[i])
			if i+1 < len(raw) {
				i++
				builder.WriteByte(raw[i])
				switch raw[i] {
				case 'Q':
					inQuote = true
				case 'E':
					inQuote = false
				}
			}
			if inClass {
				classStart = false
			}
		case '[':
			if inQuote {
				builder.WriteByte(raw[i])
				continue
			}
			if !inClass {
				inClass = true
				classStart = true
			} else {
				classStart = false
			}
			builder.WriteByte(raw[i])
		case ']':
			if inQuote {
				builder.WriteByte(raw[i])
				continue
			}
			if inClass && isPOSIXClassClose(raw, i) {
				builder.WriteByte(raw[i])
				continue
			}
			if inClass && !classStart {
				inClass = false
			}
			classStart = false
			builder.WriteByte(raw[i])
		case '(':
			if inQuote || inClass || (i+1 < len(raw) && raw[i+1] == '?') {
				builder.WriteByte(raw[i])
				if inClass {
					classStart = false
				}
				continue
			}
			builder.WriteString("(?:")
		default:
			builder.WriteByte(raw[i])
			if inClass && !(classStart && raw[i] == '^') {
				classStart = false
			}
		}
	}

	return builder.String()
}

func isPOSIXClassClose(raw string, close int) bool {
	if close < 2 || raw[close-1] != ':' {
		return false
	}
	open := close - 2
	for open >= 0 && raw[open] != '[' {
		open--
	}
	if open < 0 || open+1 >= close {
		return false
	}
	if raw[open+1] == '^' {
		return open+2 < close
	}
	return raw[open+1] == ':'
}

func walkPatternRefs(input string, fn func(start, end int, ref patternRef) error) error {
	for i := 0; i < len(input); {
		start := strings.Index(input[i:], "%{")
		if start == -1 {
			return nil
		}
		start += i

		end := strings.IndexByte(input[start+2:], '}')
		if end == -1 {
			return nil
		}
		end += start + 2

		ref, ok, err := parsePatternRef(input[start+2 : end])
		if err != nil {
			if strings.HasPrefix(err.Error(), "invalid varb data type: ") {
				return fmt.Errorf("pattern: `%%{%s}`: %w", input, err)
			}
			return err
		}
		if !ok {
			i = end + 1
			continue
		}

		if err := fn(start, end+1, ref); err != nil {
			return err
		}
		i = end + 1
	}

	return nil
}

// DenormalizePattern denormalizes the pattern to the regular expression.
func DenormalizePattern(input string, denormalized ...PatternStorageIface) (
	*GrokPattern, error,
) {
	input = normalizeAnonymousCaptures(input)
	gPattern := &GrokPattern{
		varbType: make(map[string]string),
		pattern:  input,
	}

	var builder strings.Builder
	last := 0

	if err := walkPatternRefs(input, func(start, end int, ref patternRef) error {
		if len(denormalized) == 0 {
			return fmt.Errorf("no pattern foud for %%{%s}", ref.syntax)
		}

		gP, ok := denormalized[0].GetPattern(ref.syntax)
		if !ok {
			return fmt.Errorf("no pattern foud for %%{%s}", ref.syntax)
		}

		for key, dtype := range gP.varbType {
			if _, ok := gPattern.varbType[key]; !ok {
				gPattern.varbType[key] = dtype
			}
		}

		if ref.varType != "" {
			switch ref.varType {
			case GTypeString, GTypeStr:
				gPattern.varbType[ref.alias] = GTypeStr
			case GTypeInt:
				gPattern.varbType[ref.alias] = GTypeInt
			case GTypeFloat:
				gPattern.varbType[ref.alias] = GTypeFloat
			case GTypeBool:
				gPattern.varbType[ref.alias] = GTypeBool
			}
		}

		builder.WriteString(input[last:start])
		if ref.alias != "" {
			builder.WriteString("(?P<")
			builder.WriteString(ref.alias)
			builder.WriteString(">")
			builder.WriteString(gP.denormalized)
			builder.WriteString(")")
		} else {
			builder.WriteString("(?:")
			builder.WriteString(gP.denormalized)
			builder.WriteString(")")
		}
		last = end
		return nil
	}); err != nil {
		return nil, err
	}

	if last == 0 {
		gPattern.denormalized = input
		return gPattern, nil
	}

	builder.WriteString(input[last:])
	gPattern.denormalized = builder.String()
	return gPattern, nil
}

func CompilePattern(input string, denomalized PatternStorageIface) (*GrokRegexp, error) {
	gP, err := DenormalizePattern(input, denomalized)
	if err != nil {
		return nil, err
	}
	return compileDenormalizedPattern(gP, denomalized)
}

func CompilePattern2(gP *GrokPattern, denomalized PatternStorageIface) (*GrokRegexp, error) {
	return compileDenormalizedPattern(gP, nil)
}

func compileDenormalizedPattern(gP *GrokPattern, storage PatternStorageIface) (*GrokRegexp, error) {
	meta, err := loadCompiledRegexpMeta(gP.denormalized)
	if err != nil {
		return nil, err
	}
	requiredPrefix, requiredSuffix, requiredLiterals, minMatchLength := compilePatternPrefilterIR(gP.pattern, storage)

	valueKinds := make([]valueKind, len(meta.subMatchNames.name))
	for i, name := range meta.subMatchNames.name {
		switch gP.varbType[name] {
		case GTypeInt:
			valueKinds[i] = valueKindInt
		case GTypeFloat:
			valueKinds[i] = valueKindFloat
		case GTypeBool:
			valueKinds[i] = valueKindBool
		default:
			valueKinds[i] = valueKindRaw
		}
	}

	return &GrokRegexp{
		grokPattern:      gP,
		re:               meta.re,
		filter:           meta.filter,
		subMatchNames:    meta.subMatchNames,
		nameIndex:        meta.nameIndex,
		valueKinds:       valueKinds,
		fastMatcher:      buildFastMatcher(gP, storage, meta),
		prefilter:        meta.prefilter,
		multiFilter:      meta.multiFilter,
		requiredPrefix:   requiredPrefix,
		requiredSuffix:   requiredSuffix,
		requiredLiterals: requiredLiterals,
		minMatchLength:   minMatchLength,
	}, nil
}

func loadCompiledRegexpMeta(denormalized string) (*compiledRegexpMeta, error) {
	if meta, ok := compiledRegexpCache.Load(denormalized); ok {
		return meta.(*compiledRegexpMeta), nil
	}

	re, err := regexp.Compile(denormalized)
	if err != nil {
		return nil, err
	}

	multiFilter, err := compileMultiPatternFilter([]string{denormalized})
	if err != nil {
		return nil, err
	}

	subexpNames := re.SubexpNames()
	subMatchNames := SubMatchName{
		name:        make([]string, 0, len(subexpNames)),
		subexpIndex: make([]int, 0, len(subexpNames)),
		subexpCount: len(subexpNames),
	}
	nameIndex := make(map[string]int, len(subexpNames))

	for i, name := range subexpNames {
		if name == "" {
			continue
		}

		if j, ok := nameIndex[name]; ok {
			subMatchNames.subexpIndex[j] = i
		} else {
			nameIndex[name] = len(subMatchNames.name)
		}

		subMatchNames.name = append(subMatchNames.name, name)
		subMatchNames.subexpIndex = append(subMatchNames.subexpIndex, i)
	}

	meta := &compiledRegexpMeta{
		re:            re,
		filter:        buildRegexpFilter(denormalized),
		subMatchNames: subMatchNames,
		nameIndex:     nameIndex,
		prefilter:     buildRegexpPrefilter(denormalized, re),
		multiFilter:   multiFilter,
	}
	actual, _ := compiledRegexpCache.LoadOrStore(denormalized, meta)
	return actual.(*compiledRegexpMeta), nil
}

func LoadPatternsFromPath(path string) (map[string]string, error) {
	if fi, err := os.Stat(path); err == nil {
		if fi.IsDir() {
			path += "/*"
		}
	} else {
		return nil, fmt.Errorf("invalid path : %s", path)
	}

	// only one error can be raised, when pattern is malformed
	// pattern is hard-coded "/*" so we ignore err
	files, _ := filepath.Glob(path)

	filePatterns := map[string]string{}
	for _, fileName := range files {
		// TODO limit filepath range
		// nolint:gosec
		file, err := os.Open(fileName)
		if err != nil {
			return nil, err
		}

		scanner := bufio.NewScanner(bufio.NewReader(file))

		for scanner.Scan() {
			l := scanner.Text()
			if len(l) > 0 && l[0] != '#' {
				names := strings.SplitN(l, " ", 2)
				if len(names) == 2 {
					filePatterns[names[0]] = names[1]
				}
			}
		}

		_ = file.Close()
	}
	return filePatterns, nil
}

// DenormalizePatternsFromMap denormalize pattern from map,
// will return a valid pattern:value map and an invalid pattern:error map.
func DenormalizePatternsFromMap(m map[string]string, denormalized ...map[string]*GrokPattern) (map[string]*GrokPattern, map[string]string) {
	resolved := make(map[string]*GrokPattern, len(m))
	invalid := make(map[string]string)
	visiting := make(map[string]bool, len(m))
	stack := make([]string, 0, 8)
	storage := make(PatternStorage, 0, len(denormalized)+1)
	storage = append(storage, resolved)
	storage = append(storage, denormalized...)

	var resolve func(string) error
	resolve = func(name string) error {
		if _, ok := resolved[name]; ok {
			return nil
		}
		if visiting[name] {
			start := 0
			for i := range stack {
				if stack[i] == name {
					start = i
					break
				}
			}
			line := strings.Join(append(append([]string{}, stack[start:]...), name), " -> ")
			return fmt.Errorf("circular dependency: pattern %s", line)
		}

		value, ok := m[name]
		if !ok {
			return fmt.Errorf("no pattern found for %%{%s}", name)
		}

		visiting[name] = true
		stack = append(stack, name)
		defer func() {
			stack = stack[:len(stack)-1]
			delete(visiting, name)
		}()

		if err := walkPatternRefs(value, func(_, _ int, ref patternRef) error {
			if _, ok := m[ref.syntax]; ok {
				return resolve(ref.syntax)
			}
			for _, de := range denormalized {
				if _, ok := de[ref.syntax]; ok {
					return nil
				}
			}
			return fmt.Errorf("no pattern found for %%{%s}", ref.syntax)
		}); err != nil {
			return err
		}

		gP, err := DenormalizePattern(value, storage)
		if err != nil {
			return err
		}
		resolved[name] = gP
		return nil
	}

	for name := range m {
		if err := resolve(name); err != nil {
			invalid[name] = err.Error()
		}
	}

	return resolved, invalid
}

func CopyDefalutPatterns() map[string]string {
	ret := map[string]string{}
	for k, v := range defalutPatterns {
		ret[k] = v
	}
	return ret
}

// nolint:lll
var defalutPatterns = map[string]string{
	"USERNAME":             `[a-zA-Z0-9._-]+`,
	"USER":                 `%{USERNAME}`,
	"EMAILLOCALPART":       `[a-zA-Z][a-zA-Z0-9_.+-=:]+`,
	"EMAILADDRESS":         `%{EMAILLOCALPART}@%{HOSTNAME}`,
	"HTTPDUSER":            `%{EMAILADDRESS}|%{USER}`,
	"INT":                  `(?:[+-]?(?:[0-9]+))`,
	"BASE10NUM":            `(?:[+-]?(?:[0-9]+(?:\.[0-9]+)?)|\.[0-9]+)`,
	"NUMBER":               `(?:%{BASE10NUM})`,
	"BASE16NUM":            `(?:0[xX]?[0-9a-fA-F]+)`,
	"POSINT":               `\b(?:[1-9][0-9]*)\b`,
	"NONNEGINT":            `\b(?:[0-9]+)\b`,
	"WORD":                 `\b\w+\b`,
	"NOTSPACE":             `\S+`,
	"SPACE":                `\s*`,
	"DATA":                 `.*?`,
	"GREEDYDATA":           `.*`,
	"GREEDYLINES":          `(?s).*`, // make . match \n
	"QUOTEDSTRING":         `"(?:[^"\\]*(?:\\.[^"\\]*)*)"|\'(?:[^\'\\]*(?:\\.[^\'\\]*)*)\'`,
	"UUID":                 `[A-Fa-f0-9]{8}-(?:[A-Fa-f0-9]{4}-){3}[A-Fa-f0-9]{12}`,
	"MAC":                  `(?:%{CISCOMAC}|%{WINDOWSMAC}|%{COMMONMAC})`,
	"CISCOMAC":             `(?:(?:[A-Fa-f0-9]{4}\.){2}[A-Fa-f0-9]{4})`,
	"WINDOWSMAC":           `(?:(?:[A-Fa-f0-9]{2}-){5}[A-Fa-f0-9]{2})`,
	"COMMONMAC":            `(?:(?:[A-Fa-f0-9]{2}:){5}[A-Fa-f0-9]{2})`,
	"IPV6":                 `(?:(?:(?:[0-9A-Fa-f]{1,4}:){7}(?:[0-9A-Fa-f]{1,4}|:))|(?:(?:[0-9A-Fa-f]{1,4}:){6}(?::[0-9A-Fa-f]{1,4}|(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(?:(?:[0-9A-Fa-f]{1,4}:){5}(?:(?:(?::[0-9A-Fa-f]{1,4}){1,2})|:(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(?:(?:[0-9A-Fa-f]{1,4}:){4}(?:(?:(?::[0-9A-Fa-f]{1,4}){1,3})|(?:(?::[0-9A-Fa-f]{1,4})?:(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(?:(?:[0-9A-Fa-f]{1,4}:){3}(?:(?:(?::[0-9A-Fa-f]{1,4}){1,4})|(?:(?::[0-9A-Fa-f]{1,4}){0,2}:(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(?:(?:[0-9A-Fa-f]{1,4}:){2}(?:(?:(?::[0-9A-Fa-f]{1,4}){1,5})|(?:(?::[0-9A-Fa-f]{1,4}){0,3}:(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(?:(?:[0-9A-Fa-f]{1,4}:){1}(?:(?:(?::[0-9A-Fa-f]{1,4}){1,6})|(?:(?::[0-9A-Fa-f]{1,4}){0,4}:(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(?::(?:(?:(?::[0-9A-Fa-f]{1,4}){1,7})|(?:(?::[0-9A-Fa-f]{1,4}){0,5}:(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))(?:%.+)?`,
	"IPV4":                 `(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)`,
	"IP":                   `(?:%{IPV6}|%{IPV4})`,
	"HOSTNAME":             `\b(?:[0-9A-Za-z][0-9A-Za-z-]{0,62})(?:\.(?:[0-9A-Za-z][0-9A-Za-z-]{0,62}))*(?:\.?|\b)`,
	"HOST":                 `%{HOSTNAME}`,
	"IPORHOST":             `(?:%{IP}|%{HOSTNAME})`,
	"HOSTPORT":             `%{IPORHOST}:%{POSINT}`,
	"PATH":                 `(?:%{UNIXPATH}|%{WINPATH})`,
	"UNIXPATH":             `(?:/[\w_%!$@:.,-]?/?)(?:\S+)?`,
	"TTY":                  `(?:/dev/(?:pts|tty(?:[pq])?)(?:\w+)?/?(?:[0-9]+))`,
	"WINPATH":              `(?:[A-Za-z]:|\\)(?:\\[^\\?*]*)+`,
	"URIPROTO":             `[A-Za-z]+(?:\+[A-Za-z+]+)?`,
	"URIHOST":              `%{IPORHOST}(?::%{POSINT:port})?`,
	"URIPATH":              `(?:/[A-Za-z0-9$.+!*'(){},~:;=@#%_\-]*)+`,
	"URIPARAM":             `\?[A-Za-z0-9$.+!*'|(){},~@#%&/=:;_?\-\[\]<>]*`,
	"URIPATHPARAM":         `%{URIPATH}(?:%{URIPARAM})?`,
	"URI":                  `%{URIPROTO}://(?:%{USER}(?::[^@]*)?@)?(?:%{URIHOST})?(?:%{URIPATHPARAM})?`,
	"MONTH":                `\b(?:Jan(?:uary|uar)?|Feb(?:ruary|ruar)?|M(?:a|ä)?r(?:ch|z)?|Apr(?:il)?|Ma(?:y|i)?|Jun(?:e|i)?|Jul(?:y)?|Aug(?:ust)?|Sep(?:tember)?|O(?:c|k)?t(?:ober)?|Nov(?:ember)?|De(?:c|z)(?:ember)?)\b`,
	"MONTHNUM":             `(?:0?[1-9]|1[0-2])`,
	"MONTHNUM2":            `(?:0[1-9]|1[0-2])`,
	"MONTHDAY":             `(?:(?:0[1-9])|(?:[12][0-9])|(?:3[01])|[1-9])`,
	"DAY":                  `(?:Mon(?:day)?|Tue(?:sday)?|Wed(?:nesday)?|Thu(?:rsday)?|Fri(?:day)?|Sat(?:urday)?|Sun(?:day)?)`,
	"YEAR":                 `(?:\d\d){1,2}`,
	"HOUR":                 `(?:2[0123]|[01]?[0-9])`,
	"MINUTE":               `(?:[0-5][0-9])`,
	"SECOND":               `(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)`,
	"TIME":                 `(?:[^0-9]?)%{HOUR}:%{MINUTE}(?::%{SECOND})(?:[^0-9]?)`,
	"DATE_US":              `%{MONTHNUM}[/-]%{MONTHDAY}[/-]%{YEAR}`,
	"DATE_EU":              `%{MONTHDAY}[./-]%{MONTHNUM}[./-]%{YEAR}`,
	"ISO8601_TIMEZONE":     `(?:Z|[+-]%{HOUR}(?::?%{MINUTE}))`,
	"ISO8601_SECOND":       `(?:%{SECOND}|60)`,
	"TIMESTAMP_ISO8601":    `%{YEAR}-%{MONTHNUM}-%{MONTHDAY}[T ]%{HOUR}:?%{MINUTE}(?::?%{SECOND})?%{ISO8601_TIMEZONE}?`,
	"DATE":                 `%{DATE_US}|%{DATE_EU}`,
	"DATESTAMP":            `%{DATE}[- ]%{TIME}`,
	"TZ":                   `(?:[PMCE][SD]T|UTC)`,
	"DATESTAMP_RFC822":     `%{DAY} %{MONTH} %{MONTHDAY} %{YEAR} %{TIME} %{TZ}`,
	"DATESTAMP_RFC2822":    `%{DAY}, %{MONTHDAY} %{MONTH} %{YEAR} %{TIME} %{ISO8601_TIMEZONE}`,
	"DATESTAMP_OTHER":      `%{DAY} %{MONTH} %{MONTHDAY} %{TIME} %{TZ} %{YEAR}`,
	"DATESTAMP_EVENTLOG":   `%{YEAR}%{MONTHNUM2}%{MONTHDAY}%{HOUR}%{MINUTE}%{SECOND}`,
	"HTTPDERROR_DATE":      `%{DAY} %{MONTH} %{MONTHDAY} %{TIME} %{YEAR}`,
	"SYSLOGTIMESTAMP":      `%{MONTH} +%{MONTHDAY} %{TIME}`,
	"PROG":                 `[\x21-\x5a\x5c\x5e-\x7e]+`,
	"SYSLOGPROG":           `%{PROG:program}(?:\[%{POSINT:pid}\])?`,
	"SYSLOGHOST":           `%{IPORHOST}`,
	"SYSLOGFACILITY":       `<%{NONNEGINT:facility}.%{NONNEGINT:priority}>`,
	"HTTPDATE":             `%{MONTHDAY}/%{MONTH}/%{YEAR}:%{TIME} %{INT}`,
	"QS":                   `%{QUOTEDSTRING}`,
	"SYSLOGBASE":           `%{SYSLOGTIMESTAMP:timestamp} (?:%{SYSLOGFACILITY} )?%{SYSLOGHOST:logsource} %{SYSLOGPROG}:`,
	"COMMONAPACHELOG":      `%{IPORHOST:clientip} %{HTTPDUSER:ident} %{USER:auth} \[%{HTTPDATE:timestamp}\] "(?:%{WORD:verb} %{NOTSPACE:request}(?: HTTP/%{NUMBER:httpversion})?|%{DATA:rawrequest})" %{NUMBER:response} (?:%{NUMBER:bytes}|-)`,
	"COMBINEDAPACHELOG":    `%{COMMONAPACHELOG} %{QS:referrer} %{QS:agent}`,
	"HTTPD20_ERRORLOG":     `\[%{HTTPDERROR_DATE:timestamp}\] \[%{LOGLEVEL:loglevel}\] (?:\[client %{IPORHOST:clientip}\] ){0,1}%{GREEDYDATA:errormsg}`,
	"HTTPD24_ERRORLOG":     `\[%{HTTPDERROR_DATE:timestamp}\] \[%{WORD:module}:%{LOGLEVEL:loglevel}\] \[pid %{POSINT:pid}:tid %{NUMBER:tid}\](?: \(%{POSINT:proxy_errorcode}\)%{DATA:proxy_errormessage}:)?(?: \[client %{IPORHOST:client}:%{POSINT:clientport}\])? %{DATA:errorcode}: %{GREEDYDATA:message}`,
	"HTTPD_ERRORLOG":       `%{HTTPD20_ERRORLOG}|%{HTTPD24_ERRORLOG}`,
	"LOGLEVEL":             `(?:[Aa]lert|ALERT|[Tt]race|TRACE|[Dd]ebug|DEBUG|[Nn]otice|NOTICE|[Ii]nfo|INFO|[Ww]arn?(?:ing)?|WARN?(?:ING)?|[Ee]rr?(?:or)?|ERR?(?:OR)?|[Cc]rit?(?:ical)?|CRIT?(?:ICAL)?|[Ff]atal|FATAL|[Ss]evere|SEVERE|EMERG(?:ENCY)?|[Ee]merg(?:ency)?)`,
	"COMMONENVOYACCESSLOG": `\[%{TIMESTAMP_ISO8601:timestamp}\] \"%{DATA:method} (?:%{URIPATH:uri_path}(?:%{URIPARAM:uri_param})?|%{DATA:}) %{DATA:protocol}\" %{NUMBER:status_code} %{DATA:response_flags} %{NUMBER:bytes_received} %{NUMBER:bytes_sent} %{NUMBER:duration} (?:%{NUMBER:upstream_service_time}|%{DATA:tcp_service_time}) \"%{DATA:forwarded_for}\" \"%{DATA:user_agent}\" \"%{DATA:request_id}\" \"%{DATA:authority}\" \"%{DATA:upstream_service}\"`,
}

var defalutDenormalizedPatterns map[string]*GrokPattern = func() map[string]*GrokPattern {
	patterns := CopyDefalutPatterns()
	dePs, _ := DenormalizePatternsFromMap(patterns)
	return dePs
}()

func CopyDenormalizedDefalutPatterns() map[string]*GrokPattern {
	m := map[string]*GrokPattern{}
	for k, v := range defalutDenormalizedPatterns {
		m[k] = v
	}
	return m
}
