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

func TestOSSHProfUploaderPutObjectUsesSecurityToken(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "app.hprof")
	require.NoError(t, os.WriteFile(filePath, []byte("dump"), 0o644))

	var gotSecurityToken string
	var gotPath string
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSecurityToken = r.Header.Get("X-Oss-Security-Token")
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &Config{
		HProfUploadEnabled:         true,
		HProfUploadProvider:        hprofUploadProviderOSS,
		HProfUploadEndpoint:        server.URL,
		HProfUploadBucket:          "dump-bucket",
		HProfUploadAccessKeyID:     "sts-ak",
		HProfUploadAccessKeySecret: "sts-sk",
		HProfUploadSecurityToken:   "sts-token",
		HProfUploadPathTemplate:    "{service}/{filename}",
	}
	uploader, err := newHProfUploader(cfg)
	require.NoError(t, err)

	result, err := uploader.Upload(context.Background(), &hprofUploadRequest{
		FilePath: filePath,
		Context: hprofTemplateContext{
			Service:   "svc-a",
			Timestamp: "20260720T060000Z",
			Filename:  "app.hprof",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "/dump-bucket/svc-a/app.hprof", gotPath)
	assert.Equal(t, "sts-token", gotSecurityToken)
	assert.Equal(t, "dump", gotBody)
	assert.Equal(t, hprofUploadProviderOSS, result.Provider)
	assert.Equal(t, "svc-a/app.hprof", result.ObjectKey)
}

func TestOSSHProfUploaderAssumeRoleUsesReturnedTemporaryCredentials(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "app.hprof")
	require.NoError(t, os.WriteFile(filePath, []byte("dump"), 0o644))

	var gotSecurityToken string
	var gotPath string
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSecurityToken = r.Header.Get("X-Oss-Security-Token")
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	oldFactory := assumeRoleCredentialsProviderFactory
	assumeRoleCredentialsProviderFactory = func(cfg *Config) (hprofUploadCredentialsProvider, error) {
		assert.Equal(t, "acs:ram::1234567890123456:role/flameshot-uploader", cfg.HProfUploadAssumeRoleARN)
		assert.Equal(t, "source-ak", cfg.HProfUploadAssumeRoleSourceAccessKeyID)
		assert.Equal(t, "source-sk", cfg.HProfUploadAssumeRoleSourceAccessKeySecret)
		return staticTestHProfUploadCredentialsProvider{
			creds: hprofUploadCredentials{
				AccessKeyID:     "assumed-ak",
				AccessKeySecret: "assumed-sk",
				SecurityToken:   "assumed-token",
			},
		}, nil
	}
	defer func() {
		assumeRoleCredentialsProviderFactory = oldFactory
	}()

	cfg := &Config{
		HProfUploadEnabled:                         true,
		HProfUploadProvider:                        hprofUploadProviderOSS,
		HProfUploadAuthType:                        hprofUploadAuthTypeAssumeRole,
		HProfUploadEndpoint:                        server.URL,
		HProfUploadBucket:                          "dump-bucket",
		HProfUploadAssumeRoleARN:                   "acs:ram::1234567890123456:role/flameshot-uploader",
		HProfUploadAssumeRoleSourceAccessKeyID:     "source-ak",
		HProfUploadAssumeRoleSourceAccessKeySecret: "source-sk",
		HProfUploadPathTemplate:                    "{service}/{filename}",
	}
	uploader, err := newHProfUploader(cfg)
	require.NoError(t, err)

	result, err := uploader.Upload(context.Background(), &hprofUploadRequest{
		FilePath: filePath,
		Context: hprofTemplateContext{
			Service:   "svc-a",
			Timestamp: "20260722T060000Z",
			Filename:  "app.hprof",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "/dump-bucket/svc-a/app.hprof", gotPath)
	assert.Equal(t, "assumed-token", gotSecurityToken)
	assert.Equal(t, "dump", gotBody)
	assert.Equal(t, hprofUploadProviderOSS, result.Provider)
	assert.Equal(t, "svc-a/app.hprof", result.ObjectKey)
}

func TestNewOSSHProfUploaderAssumeRoleRequiresRoleARN(t *testing.T) {
	cfg := &Config{
		HProfUploadEnabled:                         true,
		HProfUploadProvider:                        hprofUploadProviderOSS,
		HProfUploadAuthType:                        hprofUploadAuthTypeAssumeRole,
		HProfUploadEndpoint:                        "oss-cn-hangzhou.aliyuncs.com",
		HProfUploadBucket:                          "dump-bucket",
		HProfUploadAssumeRoleSourceAccessKeyID:     "source-ak",
		HProfUploadAssumeRoleSourceAccessKeySecret: "source-sk",
	}

	_, err := newHProfUploader(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "FLAMESHOT_HPROF_UPLOAD_ASSUME_ROLE_ARN")
}

func TestNewS3HProfUploaderRejectsAssumeRole(t *testing.T) {
	cfg := &Config{
		HProfUploadEnabled:         true,
		HProfUploadProvider:        hprofUploadProviderS3,
		HProfUploadAuthType:        hprofUploadAuthTypeAssumeRole,
		HProfUploadEndpoint:        "https://s3.example.com",
		HProfUploadBucket:          "dump-bucket",
		HProfUploadAccessKeyID:     "ak",
		HProfUploadAccessKeySecret: "sk",
	}

	_, err := newHProfUploader(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported s3 hprof upload auth type")
}

type staticTestHProfUploadCredentialsProvider struct {
	creds hprofUploadCredentials
}

func (p staticTestHProfUploadCredentialsProvider) Credentials(context.Context) (hprofUploadCredentials, error) {
	return p.creds, nil
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
