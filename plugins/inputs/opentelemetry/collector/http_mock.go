package collector

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

type MockCollector struct {
	endpoint string
	server   *http.Server

	clientTLSConfig *tls.Config
	expectedHeaders map[string]string
}

func (c *MockCollector) Stop() error {
	return c.server.Shutdown(context.Background())
}

func (c *MockCollector) MustStop(t *testing.T) {
	t.Helper()
	assert.NoError(t, c.server.Shutdown(context.Background()))
}

func (c *MockCollector) Endpoint() string {
	return c.endpoint
}

func (c *MockCollector) ClientTLSConfig() *tls.Config {
	return c.clientTLSConfig
}

func (c *MockCollector) ExpectedHeaders() map[string]string {
	return c.expectedHeaders
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

type MockCollectorConfig struct {
	URLPath         string
	Port            int
	ExpectedHeaders map[string]string
}

func RunMockCollector(t *testing.T, cfg MockCollectorConfig, h http.HandlerFunc) *MockCollector {
	t.Helper()
	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", cfg.Port))
	require.NoError(t, err)
	_, portStr, err := net.SplitHostPort(ln.Addr().String())
	require.NoError(t, err)
	m := &MockCollector{
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

// NewHTTPExporter http client.
func NewHTTPExporter(t *testing.T, ctx context.Context, path string, endpoint string) *otlptrace.Exporter {
	t.Helper()
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
