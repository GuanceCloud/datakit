// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package profile

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
	eventJSONFile           = "event"
	eventJSONFileWithSuffix = "event.json"
	profileTagsKey          = "tags[]"
	eventFileTagsKey        = "tags_profiler"
	subCustomTagsKey        = "sub_custom_tags"
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

var langMaps = map[string]Language{
	"java":    Java,
	"jvm":     Java,
	"python":  Python,
	"ruby":    Ruby,
	"node.js": NodeJS,
	"nodejs":  NodeJS,
	"php":     PHP,
	"dotnet":  DotNet,
	"c#":      DotNet,
	"csharp":  DotNet,
	"golang":  Golang,
	"node":    NodeJS,
	"go":      Golang,
}

// Tags refer to parsed formValue["tags[]"].
type Tags map[string]string

type rfc3339Time time.Time

func newRFC3339Time(t time.Time) *rfc3339Time {
	return (*rfc3339Time)(&t)
}

var (
	_ json.Marshaler   = (*rfc3339Time)(nil)
	_ json.Unmarshaler = (*rfc3339Time)(nil)
)

func (r *rfc3339Time) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(*r).Format(time.RFC3339Nano) + `"`), nil
}

func (r *rfc3339Time) UnmarshalJSON(bytes []byte) error {
	s := string(bytes[1 : len(bytes)-1])

	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		*r = rfc3339Time(t)
		return nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		*r = rfc3339Time(t)
		return nil
	}
	return fmt.Errorf("unresolvable time format: [%s]", s)
}

type Metadata struct {
	Format        Format       `json:"format"`
	Profiler      Profiler     `json:"profiler"`
	Attachments   []string     `json:"attachments"`
	Language      Language     `json:"language,omitempty"`
	TagsProfiler  string       `json:"tags_profiler"`
	SubCustomTags string       `json:"sub_custom_tags,omitempty"`
	Start         *rfc3339Time `json:"start"`
	End           *rfc3339Time `json:"end"`
}

func newTags(originTags []string) Tags {
	pt := make(Tags)
	for _, tag := range originTags {
		// 有":"， 用:切割成键值对
		if strings.Index(tag, ":") > 0 {
			pairs := strings.SplitN(tag, ":", 2)
			pt[pairs[0]] = pairs[1]
		} else {
			// 没有":" 整个值做key, value为空
			pt[tag] = ""
		}
	}
	return pt
}

func (t Tags) Get(name string, defVal ...string) string {
	if tag, ok := t[name]; ok {
		return tag
	}
	if len(defVal) > 0 {
		return defVal[0]
	}
	return ""
}

func ResolveLanguage(runtimes []string) Language {
	for _, r := range runtimes {
		r = strings.ToLower(r)
		for name, lang := range langMaps {
			if strings.Contains(r, name) {
				return lang
			}
		}
	}
	return UnKnown
}

func resolveStartTime(formValue map[string][]string) (time.Time, error) {
	return resolveTime(formValue, []string{"recording-start", "start"})
}

func resolveEndTime(formValue map[string][]string) (time.Time, error) {
	return resolveTime(formValue, []string{"recording-end", "end"})
}

func resolveTime(formValue map[string][]string, formFields []string) (time.Time, error) {
	var tm time.Time

	if len(formFields) == 0 {
		return tm, fmt.Errorf("form time fields is empty")
	}

	var err error
	for _, field := range formFields {
		if timeVal := formValue[field]; len(timeVal) > 0 {
			tVal := timeVal[0]
			if strings.Contains(tVal, ".") {
				tm, err = time.Parse(time.RFC3339Nano, tVal)
			} else {
				tm, err = time.Parse(time.RFC3339, tVal)
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

func getForm(field string, formValues map[string][]string) string {
	if val := formValues[field]; len(val) > 0 {
		return val[0]
	}
	return ""
}

func resolveLang(formValue map[string][]string, pt Tags) Language {
	var runtimes []string

	if v := pt.Get("language"); v != "" {
		runtimes = append(runtimes, v)
	}

	if v := pt.Get("runtime"); v != "" {
		runtimes = append(runtimes, v)
	}

	formKeys := []string{
		"runtime",
		"language",
		"family",
	}

	for _, field := range formKeys {
		if v := getForm(field, formValue); v != "" {
			runtimes = append(runtimes, v)
		}
	}

	return ResolveLanguage(runtimes)
}

func interface2String(i interface{}) (string, error) {
	switch baseV := i.(type) {
	case string:
		return baseV, nil
	case float32, float64:
		return strconv.FormatFloat(reflect.ValueOf(i).Float(), 'g', -1, 64), nil
	case int, int8, int16, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(i).Int(), 10), nil
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return strconv.FormatUint(reflect.ValueOf(i).Uint(), 10), nil
	case bool:
		if baseV {
			return "true", nil
		}
		return "false", nil
	}
	return "", fmt.Errorf("not suppoerted interface type: %T", i)
}

func json2StringMap(m map[string]interface{}) map[string][]string {
	formatted := make(map[string][]string, len(m))

	for k, v := range m {
		switch baseV := v.(type) {
		case []interface{}:
			for _, elem := range baseV {
				if elemStr, err := interface2String(elem); err == nil {
					formatted[k] = append(formatted[k], elemStr)
				}
			}
		default:
			if vStr, err := interface2String(v); err == nil {
				formatted[k] = append(formatted[k], vStr)
			}
		}
	}
	return formatted
}

type resolvedMetadata struct {
	formValue map[string][]string
	tags      Tags
}

func parseMetadata(req *http.Request) (*resolvedMetadata, int64, error) {
	filesize := int64(0)
	for _, files := range req.MultipartForm.File {
		for _, f := range files {
			filesize += f.Size
		}
	}

	if req.MultipartForm.Value != nil {
		if _, ok := req.MultipartForm.Value[profileTagsKey]; ok {
			return &resolvedMetadata{
				formValue: req.MultipartForm.Value,
				tags:      newTags(req.MultipartForm.Value[profileTagsKey]),
			}, filesize, nil
		}
	}

	eventFiles, ok := req.MultipartForm.File[eventJSONFile]
	if !ok {
		eventFiles, ok = req.MultipartForm.File[eventJSONFileWithSuffix]
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
		eventFormValues := json2StringMap(events)
		if len(eventFormValues[eventFileTagsKey]) > 0 && eventFormValues[eventFileTagsKey][0] != "" {
			eventFormValues[profileTagsKey] = strings.Split(eventFormValues[eventFileTagsKey][0], ",")
		}

		return &resolvedMetadata{
			formValue: eventFormValues,
			tags:      newTags(eventFormValues[profileTagsKey]),
		}, filesize, nil
	}
	return nil, filesize, fmt.Errorf("the profiling data format not supported, check your datadog trace library version")
}
