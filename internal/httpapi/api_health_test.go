// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	componenthealth "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/health"
)

func TestHealthHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	registry := componenthealth.NewRegistry()
	reporter := registry.Register("container-runtime", "containerd.sock", componenthealth.Options{
		FailureTimeout: time.Nanosecond,
	})

	assertHealthStatusCode(t, registry, http.StatusOK)
	reporter.Failure()
	time.Sleep(time.Millisecond)
	assertHealthStatusCode(t, registry, http.StatusServiceUnavailable)
}

func assertHealthStatusCode(t *testing.T, registry *componenthealth.Registry, want int) {
	t.Helper()
	router := gin.New()
	router.GET("/v1/health", healthHandler(registry))
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	router.ServeHTTP(recorder, req)
	if recorder.Code != want {
		t.Fatalf("status code: got %d, want %d; body: %s", recorder.Code, want, recorder.Body.String())
	}
	var body map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body) != 3 || body["live"] == nil || body["checked_at"] == nil || body["failed_components"] == nil {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}
