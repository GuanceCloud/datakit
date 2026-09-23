// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package client

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return fn(req) }

func TestClassifyAPIServerRequest(t *testing.T) {
	tests := []struct {
		method, path, verb, resource string
	}{
		{http.MethodGet, "/api/v1/namespaces/default/pods/datakit", "get", "pods"},
		{http.MethodGet, "/api/v1/nodes", "list", "nodes"},
		{http.MethodGet, "/apis/apps/v1/namespaces/default/deployments?watch=true", "watch", "deployments"},
		{http.MethodPatch, "/apis/apps/v1/namespaces/default/deployments/web", "patch", "deployments"},
		{http.MethodGet, "/apis/example.io/v1/widgets", "list", "other"},
		{http.MethodGet, "/apis/apps/v1", "get", "non_resource"},
	}

	for _, test := range tests {
		req, err := http.NewRequest(test.method, "https://kubernetes.default.svc"+test.path, nil)
		if err != nil {
			t.Fatalf("new request: %v", err)
		}
		verb, resource := classifyAPIServerRequest(req)
		if verb != test.verb || resource != test.resource {
			t.Errorf("classify %s %s = %s/%s, want %s/%s",
				test.method, test.path, verb, resource, test.verb, test.resource)
		}
	}
}

func TestProfileIsValidAPIServerMetricsComponent(t *testing.T) {
	if !validComponent(ComponentProfile) {
		t.Fatal("profile should be a valid Kubernetes API client component")
	}
}

func TestAPIServerMetricsRoundTripper(t *testing.T) {
	transportErr := errors.New("connection reset")
	tests := []struct {
		name     string
		response *http.Response
		err      error
		code     string
	}{
		{
			name: "HTTP status",
			response: &http.Response{
				StatusCode: http.StatusForbidden,
				Body:       io.NopCloser(strings.NewReader("{}")),
			},
			code: "403",
		},
		{name: "transport error", err: transportErr, code: "error"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requests := newTestAPIServerRequestCounter()
			transport := newAPIServerMetricsRoundTripper(
				ComponentContainerRuntime,
				requests,
				roundTripFunc(func(*http.Request) (*http.Response, error) {
					return test.response, test.err
				}),
			)
			req, err := http.NewRequest(http.MethodGet,
				"https://kubernetes.default.svc/api/v1/pods?watch=true", nil)
			if err != nil {
				t.Fatalf("new request: %v", err)
			}

			resp, gotErr := transport.RoundTrip(req)
			if !errors.Is(gotErr, test.err) {
				t.Fatalf("round trip error = %v, want %v", gotErr, test.err)
			}
			if resp != test.response {
				t.Fatalf("round trip response was not preserved")
			}
			if resp != nil && resp.Body != nil {
				if err := resp.Body.Close(); err != nil {
					t.Fatalf("close response body: %v", err)
				}
			}
			if got := counterValue(t, requests,
				"container_runtime", "watch", "pods", test.code); got != 1 {
				t.Fatalf("request counter = %v, want 1", got)
			}
		})
	}
}

func newTestAPIServerRequestCounter() *prometheus.CounterVec {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_kubernetes_apiserver_requests_total",
			Help: "Total test requests sent to a Kubernetes API server.",
		},
		[]string{"component", "verb", "resource", "code"},
	)
}

func counterValue(t *testing.T, counter *prometheus.CounterVec, labels ...string) float64 {
	t.Helper()
	metric := &dto.Metric{}
	if err := counter.WithLabelValues(labels...).Write(metric); err != nil {
		t.Fatalf("write counter metric: %v", err)
	}
	return metric.GetCounter().GetValue()
}
