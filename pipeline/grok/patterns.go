package grok

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
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

// DenormalizePattern denormalizes the pattern to the regular expression.
func DenormalizePattern(input string, denormalized ...PatternStorageIface) (
	*GrokPattern, error,
) {
	gPattern := &GrokPattern{
		varbType: make(map[string]string),
		pattern:  input,
	}

	pattern := input

	for _, values := range normal.FindAllStringSubmatch(pattern, -1) {
		if !valid.MatchString(values[1]) {
			return nil, fmt.Errorf("invalid pattern `%%{%s}`", values[1])
		}

		names := strings.Split(values[1], ":")
		syntax, alias := names[0], names[0]

		// [a-zA-Z0-9_]
		if len(names) > 1 {
			alias = symbolic.ReplaceAllString(names[1], "_")
		}

		// get the data type of the variable, if any
		if len(names) > 2 {
			switch names[2] {
			case GTypeString, GTypeFloat, GTypeInt, GTypeBool:
				gPattern.varbType[alias] = names[2]
			default:
				return nil, fmt.Errorf("pattern: `%%{%s}`: invalid varb data type: `%s`",
					pattern, names[2])
			}
		}

		if len(denormalized) == 0 {
			return nil, fmt.Errorf("no pattern foud for %%{%s}", syntax)
		}

		gP, ok := denormalized[0].GetPattern(syntax)
		if !ok {
			return nil, fmt.Errorf("no pattern foud for %%{%s}", syntax)
		}

		for key, dtype := range gP.varbType {
			if _, ok := gPattern.varbType[key]; !ok {
				gPattern.varbType[key] = dtype
			}
		}

		var buffer bytes.Buffer
		if len(names) > 1 {
			buffer.WriteString("(?P<")
			buffer.WriteString(alias)
			buffer.WriteString(">")
			buffer.WriteString(gP.denormalized)
			buffer.WriteString(")")
		} else {
			buffer.WriteString("(")
			buffer.WriteString(gP.denormalized)
			buffer.WriteString(")")
		}
		pattern = strings.ReplaceAll(pattern, values[0], buffer.String())
	}

	gPattern.denormalized = pattern
	return gPattern, nil
}

func CompilePattern(input string, denomalized PatternStorageIface) (*GrokRegexp, error) {
	gP, err := DenormalizePattern(input, denomalized)
	if err != nil {
		return nil, err
	}
	re, err := regexp.Compile(gP.denormalized)
	if err != nil {
		return nil, err
	}

	return &GrokRegexp{
		grokPattern: gP,
		re:          re,
	}, nil
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
	patternDeps := map[string]*nodeP{}

	for key, value := range m {
		node := &nodeP{
			cnt:   value,
			cNode: []string{},
		}

		// sub pattern
		for _, key := range normal.FindAllStringSubmatch(value, -1) {
			names := strings.Split(key[1], ":")
			syntax := names[0]

			if _, ok := m[syntax]; ok {
			} else { // 取 denormalized 的
				for _, v := range denormalized {
					if deV, ok := v[syntax]; ok {
						node.cNode = append(node.cNode, syntax)
						patternDeps[syntax] = &nodeP{
							cnt: syntax,
							ptn: deV,
						}
						break
					}
				}
			}
			node.cNode = append(node.cNode, syntax)
		}
		patternDeps[key] = node
	}

	return runTree(patternDeps)
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
	"YEAR":                 `(\d\d){1,2}`,
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
	"HTTPD24_ERRORLOG":     `\[%{HTTPDERROR_DATE:timestamp}\] \[%{WORD:module}:%{LOGLEVEL:loglevel}\] \[pid %{POSINT:pid}:tid %{NUMBER:tid}\]( \(%{POSINT:proxy_errorcode}\)%{DATA:proxy_errormessage}:)?( \[client %{IPORHOST:client}:%{POSINT:clientport}\])? %{DATA:errorcode}: %{GREEDYDATA:message}`,
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
