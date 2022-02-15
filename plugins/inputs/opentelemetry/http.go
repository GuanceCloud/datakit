// Package opentelemetry http method

package opentelemetry

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	dkHTTP "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	collectormetricpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/protobuf/proto"
)

/*
	API 接收从client端发送的trace数据
		数据格式为 protobuf
*/

// handler collector
type otlpHTTPCollector struct {
	Enable          bool              `toml:"enable"`
	HTTPStatusOK    int               `toml:"http_status_ok"`
	ExpectedHeaders map[string]string `toml:"expectedHeaders"` // 用于检测是否包含特定的 header
}

// apiOtlpCollector :trace
func (o *otlpHTTPCollector) apiOtlpTrace(w http.ResponseWriter, r *http.Request) {
	if !o.checkHeaders(r) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := collectortracepb.ExportTraceServiceResponse{}
	rawResponse, err := proto.Marshal(&response)
	if err != nil {
		l.Infof("proto marshal error=%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rawRequest, err := readRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	request, err := unmarshalTraceRequest(rawRequest, r.Header.Get("content-type"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	writeReply(w, rawResponse, o.HTTPStatusOK, nil) // 先将信息返回到客户端 然后再处理spans
	storage.AddSpans(request.ResourceSpans)
}

// apiOtlpCollector : todo metric
// nolint:unused
func (o *otlpHTTPCollector) apiOtlpMetric(w http.ResponseWriter, r *http.Request) {
	response := collectormetricpb.ExportMetricsServiceResponse{}
	rawResponse, err := proto.Marshal(&response)
	if err != nil {
		l.Infof("proto marshal error=%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rawRequest, err := readRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	request, err := unmarshalMetricsRequest(rawRequest, r.Header.Get("content-type"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	writeReply(w, rawResponse, 200, nil) // 先将信息返回到客户端 然后再处理spans
	orms := make([]*otelResourceMetric, 0)
	if rss := request.ResourceMetrics; len(rss) > 0 {
		for _, resourceMetrics := range rss {
			tags := toDatakitTags(resourceMetrics.Resource.Attributes)
			LibraryMetrics := resourceMetrics.GetInstrumentationLibraryMetrics()
			for _, libraryMetric := range LibraryMetrics {
				metrices := libraryMetric.GetMetrics()
				for _, metrice := range metrices {
					l.Debugf(metrice.Name)
					bts, err := json.MarshalIndent(metrice, "\t", "")
					if err == nil {
						l.Info(string(bts))
					}
					l.Infof("metric string=%s", metrice.String())
					ps := getData(metrice)
					for _, p := range ps {
						orm := &otelResourceMetric{
							name: metrice.Name, attributes: tags,
							description: metrice.Description, dataType: p.typeName, startTime: p.startTime,
							unitTime: p.unitTime, data: p.val,
						}
						orms = append(orms, orm)
						// todo 将 orms 转换成 行协议格式 并发送到IO
					}
				}
			}
		}
	}
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
	if contentType != "application/x-protobuf" {
		return request, fmt.Errorf("invalid content-type: %s, only application/x-protobuf is supported", contentType)
	}
	err := proto.Unmarshal(rawRequest, request)
	return request, err
}

// nolint:unused
func unmarshalMetricsRequest(rawRequest []byte, contentType string) (*collectormetricpb.ExportMetricsServiceRequest, error) {
	request := &collectormetricpb.ExportMetricsServiceRequest{}
	if contentType != "application/x-protobuf" {
		return request, fmt.Errorf("invalid content-type: %s, only application/x-protobuf is supported", contentType)
	}
	err := proto.Unmarshal(rawRequest, request)
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

func writeReply(w http.ResponseWriter, rawResponse []byte, status int, h map[string]string) {
	contentType := "application/x-protobuf"
	w.Header().Set("Content-Type", contentType)
	for k, v := range h {
		w.Header().Add(k, v)
	}
	w.WriteHeader(status)
	_, _ = w.Write(rawResponse)
}

func (o *otlpHTTPCollector) RunHTTP() {
	// 注册到http模块
	dkHTTP.RegHTTPHandler("POST", "/otel/v11/trace", o.apiOtlpTrace)
	dkHTTP.RegHTTPHandler("GET", "/otel/v11/trace", o.apiOtlpTrace)
	// metrice 暂时不开
	// dkHTTP.RegHTTPHandler("POST", "/otel/v11/metric", o.apiOtlpMetric)
}

/*
// todo del
func (o *otlpHTTPCollector) stop() {
	// empty func
}
*/
