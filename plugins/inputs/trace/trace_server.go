package trace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/ftagent/utils"
)

type Reply struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}
type TraceDecoder interface {
	Decode(octets []byte) error
}

type TraceReqInfo struct {
	Source      string
	Version     string
	ContentType string
}

type ZipkinTracer struct {
	TraceReqInfo
}

type JaegerTracer struct {
	TraceReqInfo
}

type TraceAdapter struct {
	source string

	duration    int64
	timestampUs int64
	content     string

	class         string
	serviceName   string
	operationName string
	parentID      string
	traceID       string
	spanID        string
	isError       string
	spanType      string
	endPoint      string
}

const (
	US_PER_SECOND   int64 = 1000000
	SPAN_TYPE_ENTRY       = "entry"
	SPAN_TYPE_LOCAL       = "local"
)

func (tAdpt *TraceAdapter) mkLineProto() {
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["__class"] = tAdpt.class
	tags["__operationName"] = tAdpt.operationName
	tags["__serviceName"] = tAdpt.serviceName
	tags["__parentID"] = tAdpt.parentID
	tags["__traceID"] = tAdpt.traceID
	tags["__spanID"] = tAdpt.spanID

	for tag, tagV := range gTags {
		tags[tag] = tagV
	}
	if tAdpt.isError == "true" {
		tags["__isError"] = "true"
	} else {
		tags["__isError"] = "false"
	}

	if tAdpt.endPoint != "" {
		tags["__endpoint"] = tAdpt.endPoint
	} else {
		tags["__endpoint"] = "null"
	}

	if tAdpt.spanType != "" {
		tags["__spanType"] = tAdpt.spanType
	} else {
		tags["__spanType"] = SPAN_TYPE_ENTRY
	}

	fields["__duration"] = tAdpt.duration
	fields["__content"] = tAdpt.content

	ts := time.Unix(tAdpt.timestampUs/US_PER_SECOND, (tAdpt.timestampUs%US_PER_SECOND)*1000)

	pt, err := io.MakeMetric(tAdpt.source, tags, fields, ts)
	if err != nil {
		log.Errorf("build metric err: %s", err)
		return
	}

	log.Debugf(string(pt))

	if err := io.NamedFeed(pt, io.Logging, "tracing"); err != nil {
		log.Errorf("io feed err: %s", err)
	}
}

func (t *TraceReqInfo) Decode(octets []byte) error {
	var decoder TraceDecoder
	source := strings.ToLower(t.Source)

	switch source {
	case "zipkin":
		decoder = &ZipkinTracer{*t}
	case "jaeger":
		decoder = &JaegerTracer{*t}
	default:
		return fmt.Errorf("Unsupported trace source %s", t.Source)
	}

	return decoder.Decode(octets)
}

func Handle(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Stack crash: %v", r)
			log.Errorf("Stack info :%s", string(debug.Stack()))
		}
	}()

	if err := handleTrace(w, r); err != nil {
		log.Errorf("%v", err)
	}
}

func handleTrace(w http.ResponseWriter, r *http.Request) error {
	source := r.URL.Query().Get("source")
	version := r.URL.Query().Get("version")
	contentType := r.Header.Get("Content-Type")
	contentEncoding := r.Header.Get("Content-Encoding")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		JsonReply(w, http.StatusBadRequest, "Read body err: %s", err)
		return err
	}
	defer r.Body.Close()

	if contentEncoding == "gzip" {
		body, err = utils.ReadCompressed(bytes.NewReader(body), true)
		if err != nil {
			JsonReply(w, http.StatusBadRequest, "Uncompress body err: %s", err)
			return err
		}
	}

	tInfo := TraceReqInfo{source, version, contentType}
	err = tInfo.Decode(body)
	if err != nil {
		JsonReply(w, http.StatusBadRequest, "Parse trace err: %s", err)
		return err
	}

	JsonReply(w, http.StatusOK, "ok")
	return nil
}

func JsonReply(w http.ResponseWriter, code int, strfmt string, args ...interface{}) {
	msg := fmt.Sprintf(strfmt, args...)
	w.WriteHeader(code)

	r, err := json.Marshal(Reply{
		Code: code,
		Msg:  msg,
	})
	if err == nil {
		w.Write(r)
	}
}
