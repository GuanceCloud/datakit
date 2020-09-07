package trace

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/ftagent/utils"
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
	Source      string
	Duration    int64
	TimestampUs int64
	Content     string

	Class         string
	ServiceName   string
	OperationName string
	ParentID      string
	TraceID       string
	SpanID        string
	IsError       string
	SpanType      string
	EndPoint      string

	Tags map[string]string
}

const (
	US_PER_SECOND   int64 = 1000000
	SPAN_TYPE_ENTRY       = "entry"
	SPAN_TYPE_LOCAL       = "local"
	SPAN_TYPE_EXIT        = "exit"
)

var (
	log  *logger.Logger
	once sync.Once
)

func BuildLineProto(tAdpt *TraceAdapter) ([]byte, error) {
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["__class"] = tAdpt.Class
	tags["__operationName"] = tAdpt.OperationName
	tags["__serviceName"] = tAdpt.ServiceName
	tags["__parentID"] = tAdpt.ParentID
	tags["__traceID"] = tAdpt.TraceID
	tags["__spanID"] = tAdpt.SpanID

	for tag, tagV := range tAdpt.Tags {
		tags[tag] = tagV
	}
	if tAdpt.IsError == "true" {
		tags["__isError"] = "true"
	} else {
		tags["__isError"] = "false"
	}

	if tAdpt.EndPoint != "" {
		tags["__endpoint"] = tAdpt.EndPoint
	} else {
		tags["__endpoint"] = "null"
	}

	if tAdpt.SpanType != "" {
		tags["__spanType"] = tAdpt.SpanType
	} else {
		tags["__spanType"] = SPAN_TYPE_ENTRY
	}

	fields["__duration"] = tAdpt.Duration
	fields["__content"] = tAdpt.Content

	ts := time.Unix(tAdpt.TimestampUs/US_PER_SECOND, (tAdpt.TimestampUs%US_PER_SECOND)*1000)

	pt, err := dkio.MakeMetric(tAdpt.Source, tags, fields, ts)
	if err != nil {
		GetInstance().Errorf("build metric err: %s", err)
		return nil, err
	}

	lineProtoStr := string(pt)
	GetInstance().Debugf(lineProtoStr)
	return pt, err
}

func MkLineProto(adapterGroup []*TraceAdapter, pluginName string) {
	for _, tAdpt := range adapterGroup {
		pt, err := BuildLineProto(tAdpt)
		if err != nil {
			continue
		}

		if err := dkio.NamedFeed(pt, dkio.Logging, pluginName); err != nil {
			GetInstance().Errorf("io feed err: %s", err)
		}
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
		body, err = utils.ReadCompressed(bytes.NewReader(body), true)
		if err != nil {
			return req, err
		}
	}
	req.Body = body
	return req, err
}

//GetInstance 用于获取单例模式对象
func GetInstance() *logger.Logger {
	once.Do(func() {
		log = logger.SLogger("trace")
	})
	return log
}
