// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package metrics

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	EventFile        = "event"
	EventJSONFile    = "event.json"
	ProfFile         = "prof"
	MainFile         = "main"
	MainJFRFile      = "main.jfr"
	MainPprofFile    = "main.pprof"
	AutoFile         = "auto"
	AutoJFRFile      = "auto.jfr"
	MetricFile       = "metrics"
	MetricJSONFile   = "metrics.json"
	AutoPprofFile    = "auto.pprof"
	profileTagsKey   = "tags[]"
	eventFileTagsKey = "tags_profiler"
	SubCustomTagsKey = "sub_custom_tags"
	PprofExt         = ".pprof"
)

const (
	// Deprecated: use Collapsed instead.
	RawFlameGraph Format = "rawflamegraph" // flamegraph collapse
	Collapsed     Format = "collapse"      // flamegraph collapse format see https://github.com/brendangregg/FlameGraph/blob/master/stackcollapse.pl
	JFR           Format = "jfr"           // see https://github.com/openjdk/jmc#core-api-example
	PPROF         Format = "pprof"         // see https://github.com/google/pprof/blob/main/proto/profile.proto
)

const (
	Unknown       Profiler = "unknown"
	DDtrace       Profiler = "ddtrace"
	AsyncProfiler Profiler = "async-profiler"
	PySpy         Profiler = "py-spy"
	Pyroscope     Profiler = "pyroscope"

	// GoPProf golang builtin pprof.
	GoPProf Profiler = "pprof"
)

type Profiler string

type Format string

const (
	Java    Language = "java"
	Python  Language = "python"
	Golang  Language = "golang"
	Ruby    Language = "ruby"
	NodeJS  Language = "nodejs"
	PHP     Language = "php"
	DotNet  Language = "dotnet"
	CPP     Language = "cpp"
	UnKnown Language = "unknown"
)

type Language string

func (l Language) String() string {
	return string(l)
}

// line protocol tags.
const (
	TagEndPoint    = "endpoint"
	TagService     = "service"
	TagEnv         = "env"
	TagVersion     = "version"
	TagHost        = "host"
	TagLanguage    = "language"
	TagRuntime     = "runtime"
	TagOs          = "runtime_os"
	TagRuntimeArch = "runtime_arch"
)

// line protocol fields.
const (
	FieldProfileID  = "profile_id"
	FieldLibraryVer = "library_ver"
	FieldDatakitVer = "datakit_ver"
	FieldStart      = "start"
	FieldEnd        = "end"
	FieldDuration   = "duration"
	FieldFormat     = "format"
	FieldPid        = "pid"
	FieldRuntimeID  = "runtime_id"
	FieldFileSize   = "__file_size"
)

var LangMaps = map[string]Language{
	"java":    Java,
	"jvm":     Java,
	"py":      Python,
	"python":  Python,
	"ruby":    Ruby,
	"node":    NodeJS,
	"nodejs":  NodeJS,
	"node.js": NodeJS,
	"php":     PHP,
	".net":    DotNet,
	"dotnet":  DotNet,
	"c#":      DotNet,
	"csharp":  DotNet,
	"golang":  Golang,
	"go":      Golang,
}

type RFC3339Time time.Time

func NewRFC3339Time(t time.Time) *RFC3339Time {
	return (*RFC3339Time)(&t)
}

var (
	_ json.Marshaler   = (*RFC3339Time)(nil)
	_ json.Unmarshaler = (*RFC3339Time)(nil)
)

func (r *RFC3339Time) Before(t *RFC3339Time) bool {
	if r == nil || t == nil {
		return false
	}
	return time.Time(*r).Before(time.Time(*t))
}

func (r *RFC3339Time) After(t *RFC3339Time) bool {
	if r == nil || t == nil {
		return false
	}
	return time.Time(*r).After(time.Time(*t))
}

func (r *RFC3339Time) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(*r).Format(time.RFC3339Nano) + `"`), nil
}

func (r *RFC3339Time) UnmarshalJSON(bytes []byte) error {
	s := string(bytes[1 : len(bytes)-1])

	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		*r = RFC3339Time(t)
		return nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		*r = RFC3339Time(t)
		return nil
	}
	return fmt.Errorf("unresolvable time format: [%s]", s)
}

type Metadata struct {
	Format        Format            `json:"format"`
	Profiler      Profiler          `json:"profiler"`
	Attachments   []string          `json:"attachments"`
	Language      Language          `json:"language,omitempty"`
	Tags          map[string]string `json:"-"`
	TagsProfiler  string            `json:"tags_profiler"`
	SubCustomTags string            `json:"sub_custom_tags,omitempty"`
	Start         *RFC3339Time      `json:"start"`
	End           *RFC3339Time      `json:"end"`
}

func NewTags(originTags []string) map[string]string {
	tags := make(map[string]string)
	for _, tag := range originTags {
		// 有":"， 用:切割成键值对
		if strings.Index(tag, ":") > 0 {
			pairs := strings.SplitN(tag, ":", 2)
			tags[pairs[0]] = pairs[1]
		} else {
			// 没有":" 整个值做key, value为空
			tags[tag] = ""
		}
	}
	return tags
}

func mixFormValueToTags(tags map[string]string, formValues map[string][]string) map[string]string {
	for k, v := range formValues {
		if tv, ok := tags[k]; !ok || tv == "" {
			switch len(v) {
			case 0:
				tags[k] = ""
			case 1:
				tags[k] = v[0]
			default:
				tags[k] = strings.Join(v, ",")
			}
		}
	}
	return tags
}

func ResolveLanguage(runtimes []string) Language {
	for _, r := range runtimes {
		r = strings.ToLower(r)
		for name, lang := range LangMaps {
			if strings.Contains(r, name) {
				return lang
			}
		}
	}
	return UnKnown
}

func ResolveStartTime(metadata map[string]string) (time.Time, error) {
	return resolveTime(metadata, []string{"recording-start", "start"})
}

func ResolveEndTime(metadata map[string]string) (time.Time, error) {
	return resolveTime(metadata, []string{"recording-end", "end"})
}

func resolveTime(tags map[string]string, fields []string) (time.Time, error) {
	var tm time.Time

	if len(fields) == 0 {
		return tm, fmt.Errorf("form time fields is empty")
	}

	var err error
	for _, field := range fields {
		if timeVal := tags[field]; timeVal != "" {
			if strings.Contains(timeVal, ".") {
				tm, err = time.Parse(time.RFC3339Nano, timeVal)
			} else {
				tm, err = time.Parse(time.RFC3339, timeVal)
			}
			if err == nil {
				return tm, nil
			}
		}
	}
	if err != nil {
		return tm, err
	}

	return tm, errors.New("there is not proper form time field")
}

func ResolveLang(metadata map[string]string) Language {
	var runtimes []string

	aliasNames := []string{
		"language",
		"runtime",
		"family",
	}

	for _, field := range aliasNames {
		if v := metadata[field]; v != "" {
			runtimes = append(runtimes, v)
		}
	}

	return ResolveLanguage(runtimes)
}

func any2String(i interface{}) string {
	switch ix := i.(type) {
	case string:
		return ix
	case []byte:
		return string(ix)
	case float32, float64:
		return strconv.FormatFloat(reflect.ValueOf(i).Float(), 'g', -1, 64)
	case int, int8, int16, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(i).Int(), 10)
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return strconv.FormatUint(reflect.ValueOf(i).Uint(), 10)
	case bool:
		if ix {
			return "true"
		}
		return "false"
	case json.Number:
		return ix.String()
	}
	return fmt.Sprintf("%v", i)
}

func json2FormValues(m map[string]interface{}) map[string][]string {
	formatted := make(map[string][]string, len(m))

	for k, v := range m {
		switch vx := v.(type) {
		case []interface{}:
			for _, elem := range vx {
				formatted[k] = append(formatted[k], any2String(elem))
			}
		default:
			formatted[k] = append(formatted[k], any2String(v))
		}
	}
	return formatted
}

func ParseMetadata(req *http.Request) (map[string]string, int64, error) {
	filesize := int64(0)
	for _, files := range req.MultipartForm.File {
		for _, f := range files {
			filesize += f.Size
		}
	}

	if req.MultipartForm.Value != nil {
		if _, ok := req.MultipartForm.Value[profileTagsKey]; ok {
			tags := NewTags(req.MultipartForm.Value[profileTagsKey])
			delete(req.MultipartForm.Value, profileTagsKey)
			return mixFormValueToTags(tags, req.MultipartForm.Value), filesize, nil
		}
	}

	eventFiles, ok := req.MultipartForm.File[EventFile]
	if !ok {
		eventFiles, ok = req.MultipartForm.File[EventJSONFile]
	}

	if ok && len(eventFiles) > 0 {
		fp, err := eventFiles[0].Open()
		if err != nil {
			return nil, filesize, fmt.Errorf("read event json file fail: %w", err)
		}
		defer func() {
			_ = fp.Close()
		}()

		var events map[string]interface{}
		decoder := json.NewDecoder(fp)
		if err := decoder.Decode(&events); err != nil {
			return nil, filesize, fmt.Errorf("resolve the event file fail: %w", err)
		}
		eventFormValues := json2FormValues(events)
		var tags []string
		if len(eventFormValues[eventFileTagsKey]) > 0 && eventFormValues[eventFileTagsKey][0] != "" {
			tags = strings.Split(eventFormValues[eventFileTagsKey][0], ",")
			delete(eventFormValues, eventFileTagsKey)
		}

		return mixFormValueToTags(NewTags(tags), eventFormValues), filesize, nil
	}
	return nil, filesize, fmt.Errorf("the profiling data format not supported, check your datadog trace library version")
}

func JoinTags(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}

	var sb strings.Builder

	for k, v := range m {
		sb.WriteString(k)
		sb.WriteByte(':')
		sb.WriteString(v)
		sb.WriteByte(',')
	}
	return strings.TrimSuffix(sb.String(), ",")
}
