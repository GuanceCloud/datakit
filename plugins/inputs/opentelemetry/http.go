package opentelemetry

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"

	dkHTTP "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
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
	Enable   bool `toml:"enable"`
	spanLock sync.Mutex
	// spansStorage otlptracetest.SpansStorage

	injectHTTPStatus     []int
	injectResponseHeader []map[string]string
	injectContentType    string
	delay                <-chan struct{} // todo  要不要单线程 或者延迟接收 time.tick

	// clientTLSConfig *tls.Config
	ExpectedHeaders map[string]string `toml:"expectedHeaders"` // 用于检测是否包含特定的 header
}

// apiOtlpCollector :trace
func (o *otlpHTTPCollector) apiOtlpTrace(w http.ResponseWriter, r *http.Request) {
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
	writeReply(w, rawResponse, 0, "", nil) // 先将信息返回到客户端 然后再处理spans

	dt := mkDKTrace(request.GetResourceSpans())
	l.Infof("dt len=%d", len(dt))
	after := itrace.NewAfterGather()
	// todo add filter : 增加tag
	after.Run(inputName, dt, false)
}

// apiOtlpCollector : todo metric
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
	writeReply(w, rawResponse, 0, "", nil) // 先将信息返回到客户端 然后再处理spans
	fmt.Println(request)
	rss := request.GetResourceMetrics()
	for _, spans := range rss {
		ls := spans.GetInstrumentationLibraryMetrics()
		for _, librarySpans := range ls {
			spans := librarySpans.Metrics
			for _, span := range spans {
				fmt.Println(span.Name) // todo: 组装 metric
			}
		}
	}
}
func unmarshalTraceRequest(rawRequest []byte, contentType string) (*collectortracepb.ExportTraceServiceRequest, error) {
	request := &collectortracepb.ExportTraceServiceRequest{}
	if contentType != "application/x-protobuf" {
		return request, fmt.Errorf("invalid content-type: %s, only application/x-protobuf is supported", contentType)
	}
	err := proto.Unmarshal(rawRequest, request)
	return request, err
}

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
	defer gunzipper.Close()
	_, err = io.Copy(&rawRequest, gunzipper)
	if err != nil {
		return nil, err
	}
	return rawRequest.Bytes(), nil
}

func writeReply(w http.ResponseWriter, rawResponse []byte, s int, ct string, h map[string]string) {
	status := http.StatusOK
	if s != 0 {
		status = s
	}
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

func (o *otlpHTTPCollector) RunHttp() {
	// 注册到http模块
	dkHTTP.RegHTTPHandler("POST", "/otel/v11/trace", o.apiOtlpTrace)
	dkHTTP.RegHTTPHandler("POST", "/otel/v11/metric", o.apiOtlpMetric)
}

// todo del
func (o *otlpHTTPCollector) stop() {
	// empty func
}
