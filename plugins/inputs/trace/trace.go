package trace

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

type TraceDecoder interface {
	Decode(octets []byte) error
}

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
type ZipkinTracer struct {
	TraceReqInfo
}

type TraceAdapter struct {
	Source string // third part source name

	//纳秒单位
	Duration int64

	//纳秒单位
	Start   int64
	Content string

	Project        string
	Version        string
	Env            string
	ServiceName    string
	OperationName  string
	Resource       string
	ParentID       string
	TraceID        string // source trace id
	SpanID         string
	Status         string
	SpanType       string
	EndPoint       string
	Type           string
	Pid            string
	HttpMethod     string
	HttpStatusCode string
	ContainerHost  string

	Tags map[string]string
}

const (
	SPAN_TYPE_ENTRY = "entry"
	SPAN_TYPE_LOCAL = "local"
	SPAN_TYPE_EXIT  = "exit"

	STATUS_OK       = "ok"
	STATUS_ERR      = "error"
	STATUS_INFO     = "info"
	STATUS_WARN     = "warning"
	STATUS_CRITICAL = "critical"

	PROJECT        = "project"
	VERSION        = "version"
	ENV            = "env"
	CONTAINER_HOST = "container_host"

	SPAN_SERVICE_APP    = "app"
	SPAN_SERVICE_DB     = "db"
	SPAN_SERVICE_WEB    = "web"
	SPAN_SERVICE_CACHE  = "cache"
	SPAN_SERVICE_CUSTOM = "custom"

	TAG_PROJECT        = "project"
	TAG_OPERATION      = "operation"
	TAG_SERVICE        = "service"
	TAG_VERSION        = "version"
	TAG_ENV            = "env"
	TAG_HTTP_METHOD    = "http_method"
	TAG_HTTP_CODE      = "http_status_code"
	TAG_TYPE           = "type"
	TAG_ENDPOINT       = "endpoint"
	TAG_SPAN_STATUS    = "status"
	TAG_SPAN_TYPE      = "span_type"
	TAG_CONTAINER_HOST = "container_host"

	FIELD_PARENTID = "parent_id"
	FIELD_TRACEID  = "trace_id"
	FIELD_SPANID   = "span_id"
	FIELD_DURATION = "duration"
	FIELD_START    = "start"
	FIELD_MSG      = "message"
	FIELD_RESOURCE = "resource"
	FIELD_PID      = "pid"
)

var (
	log = logger.DefaultSLogger("trace")
)

func BuildLineProto(tAdpt *TraceAdapter) (*dkio.Point, error) {
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags[TAG_PROJECT] = tAdpt.Project
	tags[TAG_OPERATION] = tAdpt.OperationName
	tags[TAG_SERVICE] = tAdpt.ServiceName
	tags[TAG_VERSION] = tAdpt.Version
	tags[TAG_ENV] = tAdpt.Env
	tags[TAG_HTTP_METHOD] = tAdpt.HttpMethod
	tags[TAG_HTTP_CODE] = tAdpt.HttpStatusCode

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

	fields[FIELD_DURATION] = tAdpt.Duration / 1000
	fields[FIELD_START] = tAdpt.Start / 1000
	fields[FIELD_MSG] = tAdpt.Content
	fields[FIELD_RESOURCE] = tAdpt.Resource

	fields[FIELD_PARENTID] = tAdpt.ParentID
	fields[FIELD_TRACEID] = tAdpt.TraceID
	fields[FIELD_SPANID] = tAdpt.SpanID

	ts := time.Unix(tAdpt.Start/int64(time.Second), tAdpt.Start%int64(time.Second))

	pt, err := dkio.MakePoint(tAdpt.Source, tags, fields, ts)
	if err != nil {
		log.Errorf("build metric err: %s", err)
		return nil, err
	}

	return pt, err
}

func MkLineProto(adapterGroup []*TraceAdapter, pluginName string) {
	var pts []*dkio.Point
	for _, tAdpt := range adapterGroup {
		// run sample

		pt, err := BuildLineProto(tAdpt)
		if err != nil {
			continue
		}
		pts = append(pts, pt)

	}

	if err := dkio.Feed(pluginName, datakit.Tracing, pts, &dkio.Option{HighFreq: true}); err != nil {
		log.Errorf("io feed err: %s", err)
	}
}

func ParseHttpReq(r *http.Request) (*TraceReqInfo, error) {
	var body []byte
	var err error
	req := &TraceReqInfo{}

	req.Source = r.URL.Query().Get("source")
	req.Version = r.URL.Query().Get("version")
	req.ContentType = r.Header.Get("Content-Type")
	contentEncoding := r.Header.Get("Content-Encoding")

	body, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if contentEncoding == "gzip" {
		body, err = ReadCompressed(bytes.NewReader(body), true)
		if err != nil {
			return req, err
		}
	}
	req.Body = body
	return req, err
}

func ReadCompressed(body *bytes.Reader, isGzip bool) ([]byte, error) {
	var data []byte
	var err error

	if isGzip {
		reader, err := gzip.NewReader(body)
		if err != nil {
			return nil, err
		}

		data, err = ioutil.ReadAll(reader)
		if err != nil {
			return nil, err
		}

	} else {
		data, err = ioutil.ReadAll(body)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

func GetFromPluginTag(tags map[string]string, tagName string) string {
	return tags[tagName]
}
