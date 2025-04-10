// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pyroscope

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/profile/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/rum"
)

type Profiler string

type Format string

type Language string

func (l Language) String() string {
	return string(l)
}

type RFC3339Time time.Time

var (
	_ json.Marshaler   = (*RFC3339Time)(nil)
	_ json.Unmarshaler = (*RFC3339Time)(nil)
)

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
	Format        metrics.Format       `json:"format"`
	Profiler      metrics.Profiler     `json:"profiler"`
	Attachments   []string             `json:"attachments"`
	Language      metrics.Language     `json:"language,omitempty"`
	TagsProfiler  string               `json:"tags_profiler"`
	SubCustomTags string               `json:"sub_custom_tags,omitempty"`
	Start         *metrics.RFC3339Time `json:"start"`
	End           *metrics.RFC3339Time `json:"end"`
}

func resolveLanguage(runtimes []string) metrics.Language {
	for _, r := range runtimes {
		if lang, ok := metrics.LangMaps[r]; ok {
			return lang
		}
	}
	for _, r := range runtimes {
		for name, lang := range metrics.LangMaps {
			if strings.Contains(r, name) {
				return lang
			}
		}
	}
	return metrics.UnKnown
}

func ResolveLanguage(metadata map[string]string) metrics.Language {
	runtimes := []string{strings.ToLower(strings.TrimSuffix(metadata["spyName"], "spy"))}
	return resolveLanguage(runtimes)
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

func QueryToPBForms(values url.Values) map[string]*rum.ValuesSlice {
	queries := make(map[string]*rum.ValuesSlice, len(values))
	for k, v := range values {
		queries[k] = &rum.ValuesSlice{
			Values: v,
		}
	}
	return queries
}

// ParseTags parse tags from url query.
func ParseTags(formValues map[string]*rum.ValuesSlice) map[string]string {
	tags := make(map[string]string, len(formValues))
	for k, v := range formValues {
		if v == nil {
			tags[k] = ""
		} else {
			tags[k] = strings.Join(v.Values, ",")
		}
	}

	if name := tags["name"]; name != "" {
		// name=go-pyroscope-demo{__session_id__=718f0dfa2700ee16,env=demo,host=SpaceX.local,service=go-pyroscope-demo,version=0.0.1}
		if idx := strings.IndexByte(name, '{'); idx > 0 && strings.HasSuffix(name, "}") && idx < len(name)-1 {
			tags["name"] = name[:idx]
			for _, kv := range strings.Split(name[idx+1:len(name)-1], ",") {
				tagK, tagV, _ := strings.Cut(kv, "=")
				if _, ok := tags[tagK]; !ok {
					tags[tagK] = tagV
				}
			}
		}
		if tags["service"] == "" && tags["name"] != "" {
			tags["service"] = tags["name"]
			delete(tags, "name")
		}
		if _, ok := tags["runtime-id"]; !ok && tags["runtime_id"] != "" {
			tags["runtime-id"] = tags["runtime_id"]
		}
	}

	return tags
}
