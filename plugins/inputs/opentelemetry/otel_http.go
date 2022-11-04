// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package opentelemetry http method

package opentelemetry

import (
	"bytes"
	"fmt"
	"net/http"

	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
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

func httpStatusRespFunc(resp http.ResponseWriter, req *http.Request, _ error) {
	response := collectortracepb.ExportTraceServiceResponse{}
	rawResponse, err := proto.Marshal(&response)
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusInternalServerError)

		return
	}

	media, _, _, err := itrace.ParseTracerRequest(req)
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	writeReply(resp, rawResponse, http.StatusOK, media, nil)
}

// handler collector.
type otlpHTTPCollector struct {
	spanStorage     *collector.SpansStorage
	Enable          bool              `toml:"enable"`
	HTTPStatusOK    int               `toml:"http_status_ok"`
	ExpectedHeaders map[string]string // 用于检测是否包含特定的 header
}

// apiOtlpCollector :trace.
func (o *otlpHTTPCollector) apiOtlpTrace(resp http.ResponseWriter, req *http.Request) {
	if o.spanStorage == nil {
		log.Error("storage is nil")
		resp.WriteHeader(http.StatusInternalServerError)

		return
	}

	if !o.checkHeaders(req) {
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	response := collectortracepb.ExportTraceServiceResponse{}
	rawResponse, err := proto.Marshal(&response)
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusInternalServerError)

		return
	}

	media, _, buf, err := itrace.ParseTracerRequest(req)
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	param := &itrace.TraceParameters{
		URLPath: req.URL.Path,
		Media:   media,
		Body:    bytes.NewBuffer(buf),
	}
	if err = o.parseOtelTrace(param); err != nil {
		log.Errorf("### parse otel trace failed: %s", err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	writeReply(resp, rawResponse, o.HTTPStatusOK, param.Media, nil)
}

func (o *otlpHTTPCollector) parseOtelTrace(param *itrace.TraceParameters) error {
	request, err := unmarshalTraceRequest(param.Body.Bytes(), param.Media)
	if err != nil {
		return err
	}

	if len(request.ResourceSpans) != 0 {
		o.spanStorage.AddSpans(request.ResourceSpans)
	}

	return nil
}

func (o *otlpHTTPCollector) apiOtlpMetric(resp http.ResponseWriter, req *http.Request) {
	if o.spanStorage == nil {
		log.Error("storage is nil")
		resp.WriteHeader(http.StatusInternalServerError)

		return
	}

	response := collectormetricpb.ExportMetricsServiceResponse{}
	rawResponse, err := proto.Marshal(&response)
	if err != nil {
		log.Errorf("proto marshal error=%v", err)
		resp.WriteHeader(http.StatusInternalServerError)

		return
	}

	media, _, buf, err := itrace.ParseTracerRequest(req)
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	request, err := unmarshalMetricsRequest(buf, media)
	if err != nil {
		log.Errorf("unmarshalMetricsRequest err=%v", err)
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	writeReply(resp, rawResponse, o.HTTPStatusOK, media, nil)

	orms := o.spanStorage.ToDatakitMetric(request.ResourceMetrics)
	o.spanStorage.AddMetric(orms)
}

func (o *otlpHTTPCollector) checkHeaders(req *http.Request) bool {
	for k, v := range o.ExpectedHeaders {
		got := req.Header.Get(k)
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

func writeReply(resp http.ResponseWriter, rawResponse []byte, status int, ct string, h map[string]string) {
	contentType := "application/x-protobuf"
	if ct != "" {
		contentType = ct
	}
	resp.Header().Set("Content-Type", contentType)
	for k, v := range h {
		resp.Header().Add(k, v)
	}
	resp.WriteHeader(status)
	_, _ = resp.Write(rawResponse)
}
