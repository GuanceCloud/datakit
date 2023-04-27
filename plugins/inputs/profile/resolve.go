// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package profile

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

const (
	eventJSONFile = "event"
)

const (
	Java    Language = "java"
	Python  Language = "python"
	Golang  Language = "golang"
	Ruby    Language = "ruby"
	NodeJS  Language = "nodejs"
	PHP     Language = "php"
	DotNet  Language = "dotnet"
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
	"c#":      DotNet,
	"csharp":  DotNet,
	"golang":  Golang,
	"node":    NodeJS,
	"go":      Golang,
}

// Tags refer to parsed formValue["tags[]"].
type Tags map[string]string

func NewTags(originTags []string) Tags {
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

func (pt Tags) Get(name string, defVal ...string) string {
	if tag, ok := pt[name]; ok {
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

func randomProfileID() string {
	var (
		id  uuid.UUID
		err error
	)
	// generate google uuid
	for i := 0; i < 3; i++ {
		if id, err = uuid.NewRandom(); err == nil {
			return id.String()
		}
	}
	log.Error("call uuid.NewRandom() fail, use our uuid instead: ", err)
	return OurUUID()
}

func OurUUID() string {
	var buf [12]byte
	rand.Read(buf[:]) //nolint:gosec,errcheck
	random := hex.EncodeToString(buf[:8])

	nanos := strconv.FormatInt(time.Now().UnixNano(), 16)
	if len(nanos) > 14 {
		nanos = nanos[len(nanos)-14:]
	}
	host, err := os.Hostname()
	if err == nil && len(host) > 0 {
		host = hex.EncodeToString([]byte(host))
		if len(host) > 8 {
			host = host[len(host)-8:]
		}
	} else {
		host = hex.EncodeToString(buf[8:])
	}
	return fmt.Sprintf("%s-%s-%s-%s-%s", random[:8], random[8:12], random[12:], host, nanos)
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
	case float64:
		return strconv.FormatFloat(baseV, 'g', -1, 64), nil
	case float32:
		return strconv.FormatFloat(float64(baseV), 'g', -1, 64), nil
	case int64:
		return strconv.FormatInt(baseV, 10), nil
	case uint64:
		return strconv.FormatUint(baseV, 10), nil
	case int:
		return strconv.Itoa(baseV), nil
	case uint:
		return strconv.FormatUint(uint64(baseV), 10), nil
	case bool:
		if baseV {
			return "true", nil
		} else {
			return "false", nil
		}
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

func parseMetadata(req *http.Request) (map[string][]string, int64, error) {
	if err := req.ParseMultipartForm(profileMaxSize); err != nil {
		return nil, 0, fmt.Errorf("parse multipart/form-data fail: %w", err)
	}

	filesize := int64(0)
	for _, files := range req.MultipartForm.File {
		for _, f := range files {
			filesize += f.Size
		}
	}

	if req.MultipartForm.Value != nil {
		if _, ok := req.MultipartForm.Value["tags[]"]; ok {
			return req.MultipartForm.Value, filesize, nil
		}
	}
	if eventFiles, ok := req.MultipartForm.File[eventJSONFile]; ok {
		if len(eventFiles) == 1 {
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
			metadata := json2StringMap(events)
			if len(metadata["tags_profiler"]) == 1 {
				metadata["tags[]"] = strings.Split(metadata["tags_profiler"][0], ",")
				delete(metadata, "tags_profiler")
			}
			if _, ok := metadata["tags[]"]; !ok {
				return nil, filesize, fmt.Errorf("the profiling data format not supported, tags[] field missing")
			}
			return metadata, filesize, nil
		}
	}
	return nil, filesize, fmt.Errorf("the profiling data format not supported, check your datadog trace library version")
}

func cache(req *http.Request) (string, int64, error) {
	formValues, filesize, err := parseMetadata(req)
	if err != nil {
		return "", 0, fmt.Errorf("parse multipart form values fail: %w", err)
	}

	profileStart, err := resolveTime(formValues, []string{"recording-start", "start"})
	if err != nil {
		return "", 0, fmt.Errorf("can not resolve profile start time: %w", err)
	}

	profileEnd, err := resolveTime(formValues, []string{"recording-end", "end"})
	if err != nil {
		return "", 0, fmt.Errorf("can not resolve profile end time: %w", err)
	}

	startNS, endNS := profileStart.UnixNano(), profileEnd.UnixNano()

	tags := NewTags(formValues["tags[]"])

	runtime := getForm("runtime", formValues)
	if runtime == "" {
		runtime = tags.Get("runtime")
	}

	pointTags := map[string]string{
		TagHost:        tags.Get("host"),
		TagEndPoint:    req.URL.Path,
		TagService:     tags.Get("service"),
		TagOs:          tags.Get("runtime_os"),
		TagRuntimeArch: tags.Get("runtime_arch"),
		TagEnv:         tags.Get("env"),
		TagVersion:     tags.Get("version"),
		TagLanguage:    resolveLang(formValues, tags).String(),
		TagRuntime:     runtime,
	}
	if pointTags[TagService] == "" {
		pointTags[TagService] = "unnamed-service"
	}

	profileID := randomProfileID()

	pointFields := map[string]interface{}{
		FieldProfileID:  profileID,
		FieldRuntimeID:  tags.Get("runtime-id"),
		FieldPid:        tags.Get("pid"),
		FieldFormat:     getForm("format", formValues),
		FieldLibraryVer: tags.Get("profiler_version"),
		FieldDatakitVer: datakit.Version,
		FieldStart:      startNS, // precision: ns
		FieldEnd:        endNS,
		FieldDuration:   endNS - startNS,
		FieldFileSize:   filesize,
	}

	// user custom tags
	for tName, tVal := range tags {
		if _, ok := pointTags[tName]; ok {
			continue
		}
		if _, ok := pointFields[tName]; ok {
			continue
		}
		// line protocol field name must not contain "."
		tName = strings.ReplaceAll(tName, ".", "_")
		pointFields[tName] = tVal
	}

	// 删除中划线 "runtime-id"
	delete(pointFields, "runtime-id")

	pt, err := point.NewPoint(inputName, pointTags, pointFields, &point.PointOption{
		Time:     time.Unix(0, startNS),
		Category: datakit.Profiling,
		Strict:   false,
	})
	if err != nil {
		return "", 0, fmt.Errorf("build profile point fail: %w", err)
	}

	pointCache.push(profileID, time.Now(), pt)

	return profileID, startNS, nil
}

func sendToIO(profileID string) error {
	pt := pointCache.drop(profileID)

	if pt != nil {
		if err := dkio.Feed(inputName+"/"+pt.Tags()[TagLanguage],
			datakit.Profiling,
			[]*point.Point{pt},
			&dkio.Option{CollectCost: time.Since(pt.Time())}); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("cache中找不到对应的profile数据， profileID = [%s]", profileID)
}
