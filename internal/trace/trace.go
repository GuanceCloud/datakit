// Package trace wrap all APM related protocol converion functions.
package trace

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

//nolint:stylecheck
const (
	CONTAINER_HOST = "container_host"
	ENV            = "env"
	PROJECT        = "project"
	VERSION        = "version"

	STATUS_OK       = "ok"
	STATUS_ERR      = "error"
	STATUS_INFO     = "info"
	STATUS_WARN     = "warning"
	STATUS_CRITICAL = "critical"

	SPAN_SERVICE_APP    = "app"
	SPAN_SERVICE_CACHE  = "cache"
	SPAN_SERVICE_CUSTOM = "custom"
	SPAN_SERVICE_DB     = "db"
	SPAN_SERVICE_WEB    = "web"
	SPAN_TYPE_ENTRY     = "entry"
	SPAN_TYPE_EXIT      = "exit"
	SPAN_TYPE_LOCAL     = "local"

	TAG_CONTAINER_HOST = "container_host"
	TAG_ENDPOINT       = "endpoint"
	TAG_ENV            = "env"
	TAG_HTTP_CODE      = "http_status_code"
	TAG_HTTP_METHOD    = "http_method"
	TAG_OPERATION      = "operation"
	TAG_PROJECT        = "project"
	TAG_SERVICE        = "service"
	TAG_SPAN_STATUS    = "status"
	TAG_SPAN_TYPE      = "span_type"
	TAG_TYPE           = "type"
	TAG_VERSION        = "version"

	FIELD_DURATION = "duration"
	FIELD_MSG      = "message"
	FIELD_PARENTID = "parent_id"
	FIELD_PID      = "pid"
	FIELD_RESOURCE = "resource"
	FIELD_SPANID   = "span_id"
	FIELD_START    = "start"
	FIELD_TRACEID  = "trace_id"
)

var log = logger.DefaultSLogger("trace")

type TraceReqInfo struct {
	Source      string
	Version     string
	ContentType string
	Body        []byte
}

type TraceRepInfo struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type TraceAdapter struct {
	ContainerHost  string
	Content        string
	Duration       int64 // 纳秒单位
	EndPoint       string
	Env            string
	HTTPMethod     string
	HTTPStatusCode string
	OperationName  string
	ParentID       string
	Pid            string
	Project        string
	Resource       string
	ServiceName    string
	Source         string // third part source name
	SpanID         string
	SpanType       string
	Start          int64 // 纳秒单位
	Status         string
	Tags           map[string]string
	TraceID        string // source trace id
	Type           string
	Version        string
}

func BuildLineProto(tAdpt *TraceAdapter) (*dkio.Point, error) {
	var (
		tags   = make(map[string]string)
		fields = make(map[string]interface{})
	)

	tags[TAG_PROJECT] = tAdpt.Project
	tags[TAG_OPERATION] = tAdpt.OperationName
	tags[TAG_SERVICE] = tAdpt.ServiceName
	tags[TAG_VERSION] = tAdpt.Version
	tags[TAG_ENV] = tAdpt.Env
	tags[TAG_HTTP_METHOD] = tAdpt.HTTPMethod
	tags[TAG_HTTP_CODE] = tAdpt.HTTPStatusCode

	if tAdpt.Type != "" {
		tags[TAG_TYPE] = tAdpt.Type
	} else {
		tags[TAG_TYPE] = SPAN_SERVICE_CUSTOM
	}

	for tag, tagV := range tAdpt.Tags {
		tags[tag] = tagV
	}

	tags[TAG_SPAN_STATUS] = tAdpt.Status

	if tAdpt.EndPoint != "" {
		tags[TAG_ENDPOINT] = tAdpt.EndPoint
	} else {
		tags[TAG_ENDPOINT] = "null"
	}

	if tAdpt.SpanType != "" {
		tags[TAG_SPAN_TYPE] = tAdpt.SpanType
	} else {
		tags[TAG_SPAN_TYPE] = SPAN_TYPE_ENTRY
	}

	if tAdpt.ContainerHost != "" {
		tags[TAG_CONTAINER_HOST] = tAdpt.ContainerHost
	}

	if tAdpt.ParentID == "" {
		tAdpt.ParentID = "0"
	}

	fields[FIELD_DURATION] = tAdpt.Duration / int64(time.Microsecond)
	fields[FIELD_START] = tAdpt.Start / int64(time.Microsecond)
	fields[FIELD_MSG] = tAdpt.Content
	fields[FIELD_RESOURCE] = tAdpt.Resource
	fields[FIELD_PARENTID] = tAdpt.ParentID
	fields[FIELD_TRACEID] = tAdpt.TraceID
	fields[FIELD_SPANID] = tAdpt.SpanID

	ts := time.Unix(tAdpt.Start/int64(time.Second), tAdpt.Start%int64(time.Second))
	pt, err := dkio.MakePoint(tAdpt.Source, tags, fields, ts)
	if err != nil {
		log.Errorf("build metric err: %s", err)
	}

	return pt, err
}

func MkLineProto(adapterGroup []*TraceAdapter, pluginName string) {
	var pts []*dkio.Point
	for _, tAdpt := range adapterGroup {
		if pt, err := BuildLineProto(tAdpt); err == nil {
			pts = append(pts, pt)
		}
	}

	if err := dkio.Feed(pluginName, datakit.Tracing, pts, &dkio.Option{HighFreq: true}); err != nil {
		log.Errorf("io feed err: %s", err)
	}
}

func ParseHTTPReq(req *http.Request) (*TraceReqInfo, error) {
	defer req.Body.Close() //nolint:errcheck
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	reqInfo := &TraceReqInfo{
		Source:      req.URL.Query().Get("source"),
		ContentType: req.Header.Get("Content-Type"),
		Version:     req.URL.Query().Get("version"),
		Body:        body,
	}
	if req.Header.Get("Content-Encoding") == "gzip" {
		var rd *gzip.Reader
		if rd, err = gzip.NewReader(bytes.NewBuffer(body)); err == nil {
			if body, err = io.ReadAll(rd); err == nil {
				reqInfo.Body = body
			}
		}
	}

	return reqInfo, err
}
