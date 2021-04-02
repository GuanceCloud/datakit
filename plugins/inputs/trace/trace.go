package trace

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
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
	Source string

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
	TraceID        string
	SpanID         string
	Status         string
	SpanType       string
	EndPoint       string
	Type           string
	Pid            string
	HttpMethod     string
	HttpStatusCode string

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

	PROJECT = "project"
	VERSION = "version"
	ENV     = "env"

	SPAN_SERVICE_APP    = "app"
	SPAN_SERVICE_DB     = "db"
	SPAN_SERVICE_WEB    = "web"
	SPAN_SERVICE_CACHE  = "cache"
	SPAN_SERVICE_CUSTOM = "custom"
)

var (
	log  *logger.Logger
	once sync.Once
)

func BuildLineProto(tAdpt *TraceAdapter) (*dkio.Point, error) {
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["project"] = tAdpt.Project
	tags["operation"] = tAdpt.OperationName
	tags["service"] = tAdpt.ServiceName
	tags["parent_id"] = tAdpt.ParentID
	tags["trace_id"] = tAdpt.TraceID
	tags["span_id"] = tAdpt.SpanID
	tags["version"] = tAdpt.Version
	tags["env"] = tAdpt.Env
	tags["http_method"] = tAdpt.HttpMethod
	tags["http_status_code"] = tAdpt.HttpStatusCode

	if tAdpt.Type != "" {
		tags["type"] = tAdpt.Type
	} else {
		tags["type"] = SPAN_SERVICE_CUSTOM
	}

	for tag, tagV := range tAdpt.Tags {
		tags[tag] = tagV
	}

	tags["status"] = tAdpt.Status

	if tAdpt.EndPoint != "" {
		tags["endpoint"] = tAdpt.EndPoint
	} else {
		tags["endpoint"] = "null"
	}

	if tAdpt.SpanType != "" {
		tags["span_type"] = tAdpt.SpanType
	} else {
		tags["span_type"] = SPAN_TYPE_ENTRY
	}

	fields["duration"] = tAdpt.Duration / 1000
	fields["start"] = tAdpt.Start / 1000
	fields["message"] = tAdpt.Content
	fields["resource"] = tAdpt.Resource

	ts := time.Unix(tAdpt.Start/int64(time.Second), tAdpt.Start%int64(time.Second))

	pt, err := dkio.MakePoint(tAdpt.Source, tags, fields, ts)
	if err != nil {
		GetInstance().Errorf("build metric err: %s", err)
		return nil, err
	}

	return pt, err
}

func MkLineProto(adapterGroup []*TraceAdapter, pluginName string) {
	var pts []*dkio.Point
	for _, tAdpt := range adapterGroup {
		pt, err := BuildLineProto(tAdpt)
		if err != nil {
			continue
		}
		pts = append(pts, pt)

	}

	if err := dkio.Feed(pluginName, dkio.Tracing, pts, &dkio.Option{HighFreq: true}); err != nil {
		GetInstance().Errorf("io feed err: %s", err)
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

//GetInstance 用于获取单例模式对象
func GetInstance() *logger.Logger {
	once.Do(func() {
		log = logger.SLogger("trace")
	})
	return log
}

func GetFromPluginTag(tags map[string]string, tagName string) string {
	return tags[tagName]
}
