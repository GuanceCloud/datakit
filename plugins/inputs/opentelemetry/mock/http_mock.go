// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mock

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
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
	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_opentelemetry_mock"})
	g.Go(func(ctx context.Context) error {
		_ = server.Serve(ln)
		return nil
	})
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
