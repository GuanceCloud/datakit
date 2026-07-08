// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderHProfObjectKeyAndDownloadURL(t *testing.T) {
	cfg := &Config{
		HProfUploadProvider:      hprofUploadProviderOSS,
		HProfUploadEndpoint:      "oss-cn-hangzhou.aliyuncs.com",
		HProfUploadBucket:        "dump-bucket",
		HProfUploadPathTemplate:  "{service}/{pod_name}/{timestamp}/{filename}",
		HProfDownloadURLTemplate: "{endpoint}/{bucket}/{object_key}",
	}
	req := &hprofUploadRequest{
		FilePath: "/flameshot-data/app.hprof",
		Context: hprofTemplateContext{
			Service:   "svc-a",
			PodName:   "pod-a",
			Timestamp: "20260625T071122Z",
			Filename:  "app.hprof",
		},
	}

	key := renderHProfObjectKey(cfg, req)
	assert.Equal(t, "svc-a/pod-a/20260625T071122Z/app.hprof", key)

	ctx := req.Context
	ctx.Bucket = cfg.HProfUploadBucket
	ctx.Endpoint = normalizeEndpoint(cfg.HProfUploadEndpoint)
	ctx.ObjectKey = key
	assert.Equal(t, "https://oss-cn-hangzhou.aliyuncs.com/dump-bucket/svc-a/pod-a/20260625T071122Z/app.hprof",
		renderHProfDownloadURL(cfg, hprofUploadProviderOSS, ctx))
}

func TestS3HProfUploaderPutObjectSignsRequest(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "app.hprof")
	require.NoError(t, os.WriteFile(filePath, []byte("dump"), 0o644))

	var gotAuth string
	var gotPath string
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &Config{
		HProfUploadEnabled:         true,
		HProfUploadProvider:        hprofUploadProviderS3,
		HProfUploadEndpoint:        server.URL,
		HProfUploadRegion:          "us-east-1",
		HProfUploadBucket:          "dump-bucket",
		HProfUploadAccessKeyID:     "ak",
		HProfUploadAccessKeySecret: "sk",
		HProfUploadPathTemplate:    "{service}/{pod_name}/{timestamp}/{filename}",
		HProfUploadS3PathStyle:     boolPtr(true),
	}
	uploader, err := newHProfUploader(cfg)
	require.NoError(t, err)

	result, err := uploader.Upload(context.Background(), &hprofUploadRequest{
		FilePath: filePath,
		Context: hprofTemplateContext{
			Service:   "svc-a",
			PodName:   "pod-a",
			Timestamp: "20260625T071122Z",
			Filename:  "app.hprof",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "/dump-bucket/svc-a/pod-a/20260625T071122Z/app.hprof", gotPath)
	assert.Contains(t, gotAuth, "AWS4-HMAC-SHA256")
	assert.Equal(t, "dump", gotBody)
	assert.Equal(t, hprofUploadProviderS3, result.Provider)
	assert.Equal(t, "svc-a/pod-a/20260625T071122Z/app.hprof", result.ObjectKey)
	assert.Contains(t, result.DownloadURL, "/dump-bucket/svc-a/pod-a/20260625T071122Z/app.hprof")
}

func TestBuildHProfTemplateContextFromTags(t *testing.T) {
	ctx := buildHProfTemplateContext(
		"svc-a",
		1234,
		"/data/app.hprof",
		"/data",
		[]string{"pod_name:pod-a", "pod_namespace:prod", "host:node-a"},
		time.Date(2026, 6, 25, 7, 11, 22, 0, time.UTC),
	)

	assert.Equal(t, "svc-a", ctx.Service)
	assert.Equal(t, "pod-a", ctx.PodName)
	assert.Equal(t, "prod", ctx.PodNamespace)
	assert.Equal(t, "node-a", ctx.Host)
	assert.Equal(t, int32(1234), ctx.PID)
	assert.Equal(t, "20260625T071122Z", ctx.Timestamp)
	assert.Equal(t, "app.hprof", ctx.Filename)
	assert.Equal(t, "/data", ctx.ProfilingPath)
}
