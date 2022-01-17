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

	SPAN_TYPE_ENTRY = "entry"
	SPAN_TYPE_EXIT  = "exit"
	SPAN_TYPE_LOCAL = "local"

	SPAN_SERVICE_APP    = "app"
	SPAN_SERVICE_CACHE  = "cache"
	SPAN_SERVICE_CUSTOM = "custom"
	SPAN_SERVICE_DB     = "db"
	SPAN_SERVICE_WEB    = "web"

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

var log = logger.DefaultSLogger("dktrace")

type DatakitSpan struct {
	TraceID        string
	ParentID       string
	SpanID         string
	SpanType       string
	Service        string
	Resource       string
	Operation      string
	Source         string // third part source name
	ContainerHost  string
	EndPoint       string
	Env            string
	HTTPMethod     string
	HTTPStatusCode string
	Pid            string
	Start          int64 // nano sec
	Duration       int64 // nano sec
	Status         string
	Type           string
	Tags           map[string]string
	Content        string
	Project        string
	Version        string
}

type DatakitTrace []*DatakitSpan

func FindIntIDSpanType(spanID, parentID int64, spanIDs, parentIDs map[int64]bool) string {
	if parentID != 0 {
		if spanIDs[parentID] {
			if parentIDs[spanID] {
				return SPAN_TYPE_LOCAL
			} else {
				return SPAN_TYPE_EXIT
			}
		}
	}

	return SPAN_TYPE_ENTRY
}

func FindStringIDSpanType(spanID, parentID string, spanIDs, parentIDs map[string]bool) string {
	if parentID != "" && parentID != "0" {
		if spanIDs[parentID] {
			if parentIDs[spanID] {
				return SPAN_TYPE_LOCAL
			} else {
				return SPAN_TYPE_EXIT
			}
		}
	}

	return SPAN_TYPE_ENTRY
}

func BuildLineProto(tAdapter *DatakitSpan) (*dkio.Point, error) {
	var (
		tags   = make(map[string]string)
		fields = make(map[string]interface{})
	)

	tags[TAG_PROJECT] = tAdapter.Project
	tags[TAG_OPERATION] = tAdapter.Operation
	tags[TAG_SERVICE] = tAdapter.Service
	tags[TAG_VERSION] = tAdapter.Version
	tags[TAG_ENV] = tAdapter.Env
	tags[TAG_HTTP_METHOD] = tAdapter.HTTPMethod
	tags[TAG_HTTP_CODE] = tAdapter.HTTPStatusCode

	if tAdapter.Type != "" {
		tags[TAG_TYPE] = tAdapter.Type
	} else {
		tags[TAG_TYPE] = SPAN_SERVICE_CUSTOM
	}

	for k, v := range tAdapter.Tags {
		tags[k] = v
	}

	tags[TAG_SPAN_STATUS] = tAdapter.Status

	if tAdapter.EndPoint != "" {
		tags[TAG_ENDPOINT] = tAdapter.EndPoint
	} else {
		tags[TAG_ENDPOINT] = "null"
	}

	if tAdapter.SpanType != "" {
		tags[TAG_SPAN_TYPE] = tAdapter.SpanType
	} else {
		tags[TAG_SPAN_TYPE] = SPAN_TYPE_ENTRY
	}

	if tAdapter.ContainerHost != "" {
		tags[TAG_CONTAINER_HOST] = tAdapter.ContainerHost
	}

	if tAdapter.ParentID == "" {
		tAdapter.ParentID = "0"
	}

	fields[FIELD_DURATION] = tAdapter.Duration / int64(time.Microsecond)
	fields[FIELD_START] = tAdapter.Start / int64(time.Microsecond)
	fields[FIELD_MSG] = tAdapter.Content
	fields[FIELD_RESOURCE] = tAdapter.Resource
	fields[FIELD_PARENTID] = tAdapter.ParentID
	fields[FIELD_TRACEID] = tAdapter.TraceID
	fields[FIELD_SPANID] = tAdapter.SpanID

	ts := time.Unix(tAdapter.Start/int64(time.Second), tAdapter.Start%int64(time.Second))
	pt, err := dkio.MakePoint(tAdapter.Source, tags, fields, ts)
	if err != nil {
		log.Errorf("build metric err: %s", err)
	}

	return pt, err
}

func MkLineProto(dktrace DatakitTrace, inputName string) {
	var pts []*dkio.Point
	for _, dkspan := range dktrace {
		if pt, err := BuildLineProto(dkspan); err == nil {
			pts = append(pts, pt)
		}
	}

	if err := dkio.Feed(inputName, datakit.Tracing, pts, &dkio.Option{HighFreq: true}); err != nil {
		log.Errorf("io feed err: %s", err)
	}
}

type TraceReqInfo struct {
	Source      string
	Version     string
	ContentType string
	Body        []byte
}

func ParseTraceInfo(req *http.Request) (*TraceReqInfo, error) {
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
