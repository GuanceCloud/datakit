// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package opentelemetry http method

package opentelemetry

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/opentelemetry/collector"
	collectormetricpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	pbContentType   = "application/x-protobuf"
	jsonContentType = "application/json"
)

// handler collector.
type otlpHTTPCollector struct {
	storage      *collector.SpansStorage
	Enable       bool `toml:"enable"`
	HTTPStatusOK int  `toml:"http_status_ok"`

	ExpectedHeaders map[string]string // 用于检测是否包含特定的 header
}

// apiOtlpCollector :trace.
func (o *otlpHTTPCollector) apiOtlpTrace(w http.ResponseWriter, r *http.Request) {
	if o.storage == nil {
		l.Error("storage is nil")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !o.checkHeaders(r) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := collectortracepb.ExportTraceServiceResponse{}
	rawResponse, err := proto.Marshal(&response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rawRequest, err := readRequest(r)
	if err != nil {
		l.Errorf("readRequest err=%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	request, err := unmarshalTraceRequest(rawRequest, r.Header.Get("content-type"))
	if err != nil {
		l.Errorf("unmarshalMetricsRequest err=%v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	writeReply(w, rawResponse, o.HTTPStatusOK, r.Header.Get("content-type"), nil) // 先将信息返回到客户端 然后再处理spans
	if len(request.ResourceSpans) > 0 {
		o.storage.AddSpans(request.ResourceSpans)
	}
}

func (o *otlpHTTPCollector) apiOtlpMetric(w http.ResponseWriter, r *http.Request) {
	if o.storage == nil {
		l.Error("storage is nil")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := collectormetricpb.ExportMetricsServiceResponse{}
	rawResponse, err := proto.Marshal(&response)
	if err != nil {
		l.Errorf("proto marshal error=%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rawRequest, err := readRequest(r)
	if err != nil {
		l.Errorf("readRequest err=%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	request, err := unmarshalMetricsRequest(rawRequest, r.Header.Get("content-type"))
	if err != nil {
		l.Errorf("unmarshalMetricsRequest err=%v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	writeReply(w, rawResponse, o.HTTPStatusOK, r.Header.Get("content-type"), nil) // 先将信息返回到客户端 然后再处理spans
	orms := o.storage.ToDatakitMetric(request.ResourceMetrics)
	o.storage.AddMetric(orms)
}

func (o *otlpHTTPCollector) checkHeaders(r *http.Request) bool {
	for k, v := range o.ExpectedHeaders {
		got := r.Header.Get(k)
		if got != v {
			return false
		}
	}
	return true
}

func unmarshalTraceRequest(rawRequest []byte, contentType string) (*collectortracepb.ExportTraceServiceRequest, error) {
	request := &collectortracepb.ExportTraceServiceRequest{}
	var err error
	switch contentType {
	case pbContentType:
		err = proto.Unmarshal(rawRequest, request)
	case jsonContentType:
		err = protojson.Unmarshal(rawRequest, request)
	default:
		err = fmt.Errorf("invalid content-type: %s, only application/x-protobuf and application/json is supported", contentType)
	}
	return request, err
}

func unmarshalMetricsRequest(rawRequest []byte, contentType string) (*collectormetricpb.ExportMetricsServiceRequest, error) {
	request := &collectormetricpb.ExportMetricsServiceRequest{}
	var err error
	switch contentType {
	case pbContentType:
		err = proto.Unmarshal(rawRequest, request)
	case jsonContentType:
		err = protojson.Unmarshal(rawRequest, request)
	default:
		err = fmt.Errorf("invalid content-type: %s, only application/x-protobuf and application/json is supported", contentType)
	}
	return request, err
}

func readRequest(r *http.Request) ([]byte, error) {
	if r.Header.Get("Content-Encoding") == "gzip" {
		return readGzipBody(r.Body)
	}
	return ioutil.ReadAll(r.Body)
}

func readGzipBody(body io.Reader) ([]byte, error) {
	rawRequest := bytes.Buffer{}
	gunzipper, err := gzip.NewReader(body)
	if err != nil {
		return nil, err
	}
	defer gunzipper.Close()                  //nolint:errcheck
	_, err = io.Copy(&rawRequest, gunzipper) //nolint:gosec
	if err != nil {
		return nil, err
	}
	return rawRequest.Bytes(), nil
}

func writeReply(w http.ResponseWriter, rawResponse []byte, status int, ct string, h map[string]string) {
	contentType := "application/x-protobuf"
	if ct != "" {
		contentType = ct
	}
	w.Header().Set("Content-Type", contentType)
	for k, v := range h {
		w.Header().Add(k, v)
	}
	w.WriteHeader(status)
	_, _ = w.Write(rawResponse)
}
