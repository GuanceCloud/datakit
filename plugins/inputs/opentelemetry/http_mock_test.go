package opentelemetry

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

type mockCollector struct {
	endpoint string
	server   *http.Server

	spanLock sync.Mutex

	injectHTTPStatus  []int
	injectContentType string
	delay             <-chan struct{}

	clientTLSConfig *tls.Config
	expectedHeaders map[string]string
}

func (c *mockCollector) Stop() error {
	return c.server.Shutdown(context.Background())
}

func (c *mockCollector) MustStop(t *testing.T) {
	assert.NoError(t, c.server.Shutdown(context.Background()))
}

func (c *mockCollector) Endpoint() string {
	return c.endpoint
}

func (c *mockCollector) ClientTLSConfig() *tls.Config {
	return c.clientTLSConfig
}

/*
func (c *mockCollector) serveMetrics(w http.ResponseWriter, r *http.Request) {
	if c.delay != nil {
		select {
		case <-c.delay:
		case <-r.Context().Done():
			return
		}
	}

	if !c.checkHeaders(r) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	response := collectormetricpb.ExportMetricsServiceResponse{}
	rawResponse, err := proto.Marshal(&response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if injectedStatus := c.getInjectHTTPStatus(); injectedStatus != 0 {
		writeReply(w, rawResponse, injectedStatus, c.injectContentType)
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
	writeReply(w, rawResponse, 0, c.injectContentType)
	c.spanLock.Lock()
	defer c.spanLock.Unlock()
	c.metricsStorage.AddMetrics(request)
}
*/

type mockCollectorConfig struct {
	URLPath         string
	Port            int
	ExpectedHeaders map[string]string
}

func runMockCollector(t *testing.T, cfg mockCollectorConfig, h http.HandlerFunc) *mockCollector {
	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", cfg.Port))
	require.NoError(t, err)
	_, portStr, err := net.SplitHostPort(ln.Addr().String())
	require.NoError(t, err)
	m := &mockCollector{
		endpoint: fmt.Sprintf("localhost:%s", portStr),
	}
	mux := http.NewServeMux()
	mux.Handle(cfg.URLPath, h)
	server := &http.Server{
		Handler: mux,
	}

	go func() {
		_ = server.Serve(ln)
	}()
	m.server = server
	return m
}

// newHTTPExporter http client
func newHTTPExporter(t *testing.T, ctx context.Context, path string, endpoint string) *otlptrace.Exporter {
	httpClent, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithURLPath(path),
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithHeaders(map[string]string{"header1": "1"}))
	if err != nil {
		t.Fatalf("failed to create a new collector exporter: %v", err)
	}
	return httpClent
}
