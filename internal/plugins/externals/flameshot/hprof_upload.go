// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
)

const (
	hprofUploadProviderOSS = "oss"
	hprofUploadProviderS3  = "s3"

	defaultHProfUploadPathTemplate = "{service}/{pod_name}/{timestamp}/{filename}"
	defaultHProfUploadTimeout      = 5 * time.Minute
)

type hprofUploadRequest struct {
	FilePath string
	Context  hprofTemplateContext
}

type hprofUploadResult struct {
	Provider    string
	Bucket      string
	ObjectKey   string
	DownloadURL string
}

type hprofTemplateContext struct {
	Service       string
	PodName       string
	PodNamespace  string
	Host          string
	PID           int32
	Timestamp     string
	Filename      string
	ProfilingPath string
	Bucket        string
	Endpoint      string
	ObjectKey     string
}

type hprofUploader interface {
	Upload(ctx context.Context, req *hprofUploadRequest) (*hprofUploadResult, error)
}

type ossHProfUploader struct {
	cfg     *Config
	timeout time.Duration
}

type s3HProfUploader struct {
	cfg     *Config
	timeout time.Duration
	client  *http.Client
}

func (c *Config) profilingEnabled() bool {
	if c == nil || c.ProfilingEnabled == nil {
		return true
	}
	return *c.ProfilingEnabled
}

func (c *Config) hprofUploadEnabled() bool {
	return c != nil && c.HProfUploadEnabled
}

func (c *Config) hprofUploadTimeout() time.Duration {
	if c == nil || c.HProfUploadTimeout == "" {
		return defaultHProfUploadTimeout
	}
	d, err := time.ParseDuration(c.HProfUploadTimeout)
	if err != nil || d <= 0 {
		return defaultHProfUploadTimeout
	}
	return d
}

func (c *Config) hprofUploadS3PathStyle() bool {
	if c == nil || c.HProfUploadS3PathStyle == nil {
		return true
	}
	return *c.HProfUploadS3PathStyle
}

func newHProfUploader(cfg *Config) (hprofUploader, error) {
	if cfg == nil || !cfg.hprofUploadEnabled() {
		return nil, nil
	}
	if cfg.HProfUploadEndpoint == "" {
		return nil, fmt.Errorf("hprof upload endpoint is required")
	}
	if cfg.HProfUploadBucket == "" {
		return nil, fmt.Errorf("hprof upload bucket is required")
	}
	if cfg.HProfUploadAccessKeyID == "" || cfg.HProfUploadAccessKeySecret == "" {
		return nil, fmt.Errorf("hprof upload access key id/secret is required")
	}

	timeout := cfg.hprofUploadTimeout()
	switch strings.ToLower(strings.TrimSpace(cfg.HProfUploadProvider)) {
	case hprofUploadProviderOSS:
		return &ossHProfUploader{cfg: cfg, timeout: timeout}, nil
	case hprofUploadProviderS3:
		return &s3HProfUploader{
			cfg:     cfg,
			timeout: timeout,
			client:  &http.Client{Timeout: timeout},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported hprof upload provider %q", cfg.HProfUploadProvider)
	}
}

func (u *ossHProfUploader) Upload(ctx context.Context, req *hprofUploadRequest) (*hprofUploadResult, error) {
	if u == nil || u.cfg == nil || req == nil {
		return nil, fmt.Errorf("oss uploader is not initialized")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	key := renderHProfObjectKey(u.cfg, req)
	connectTimeout := int64(30)
	readWriteTimeout := int64(u.timeout.Seconds())
	if readWriteTimeout <= 0 {
		readWriteTimeout = int64(defaultHProfUploadTimeout.Seconds())
	}
	client, err := oss.New(
		normalizeEndpoint(u.cfg.HProfUploadEndpoint),
		u.cfg.HProfUploadAccessKeyID,
		u.cfg.HProfUploadAccessKeySecret,
		oss.Timeout(connectTimeout, readWriteTimeout),
	)
	if err != nil {
		return nil, err
	}
	bucket, err := client.Bucket(u.cfg.HProfUploadBucket)
	if err != nil {
		return nil, err
	}
	if err := bucket.PutObjectFromFile(key, req.FilePath); err != nil {
		return nil, err
	}

	ctxCopy := req.Context
	ctxCopy.Bucket = u.cfg.HProfUploadBucket
	ctxCopy.Endpoint = normalizeEndpoint(u.cfg.HProfUploadEndpoint)
	ctxCopy.ObjectKey = key
	return &hprofUploadResult{
		Provider:    hprofUploadProviderOSS,
		Bucket:      u.cfg.HProfUploadBucket,
		ObjectKey:   key,
		DownloadURL: renderHProfDownloadURL(u.cfg, hprofUploadProviderOSS, ctxCopy),
	}, nil
}

func (u *s3HProfUploader) Upload(ctx context.Context, req *hprofUploadRequest) (*hprofUploadResult, error) {
	if u == nil || u.cfg == nil || req == nil {
		return nil, fmt.Errorf("s3 uploader is not initialized")
	}
	key := renderHProfObjectKey(u.cfg, req)
	file, err := os.Open(req.FilePath) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer file.Close() //nolint:errcheck,gosec

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	putURL, err := buildS3ObjectURL(u.cfg, key)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, putURL, file)
	if err != nil {
		return nil, err
	}
	httpReq.ContentLength = stat.Size()
	httpReq.Header.Set("Content-Type", "application/octet-stream")
	payloadHash, err := fileSHA256Hex(file)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("X-Amz-Content-Sha256", payloadHash)
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	region := u.cfg.HProfUploadRegion
	if region == "" {
		region = "us-east-1"
	}
	signer := v4.NewSigner(
		credentials.NewStaticCredentials(u.cfg.HProfUploadAccessKeyID, u.cfg.HProfUploadAccessKeySecret, ""),
		func(s *v4.Signer) {
			s.DisableURIPathEscaping = true
		},
	)
	if _, err := signer.Sign(httpReq, file, "s3", region, time.Now()); err != nil {
		return nil, err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	resp, err := u.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("s3 put object failed, status=%d", resp.StatusCode)
	}

	ctxCopy := req.Context
	ctxCopy.Bucket = u.cfg.HProfUploadBucket
	ctxCopy.Endpoint = normalizeEndpoint(u.cfg.HProfUploadEndpoint)
	ctxCopy.ObjectKey = key
	return &hprofUploadResult{
		Provider:    hprofUploadProviderS3,
		Bucket:      u.cfg.HProfUploadBucket,
		ObjectKey:   key,
		DownloadURL: renderHProfDownloadURL(u.cfg, hprofUploadProviderS3, ctxCopy),
	}, nil
}

func renderHProfObjectKey(cfg *Config, req *hprofUploadRequest) string {
	template := defaultHProfUploadPathTemplate
	if cfg != nil && cfg.HProfUploadPathTemplate != "" {
		template = cfg.HProfUploadPathTemplate
	}
	ctx := req.Context
	if ctx.Filename == "" {
		ctx.Filename = filepath.Base(req.FilePath)
	}
	if ctx.Timestamp == "" {
		ctx.Timestamp = time.Now().UTC().Format("20060102T150405Z")
	}
	if cfg != nil {
		ctx.Bucket = cfg.HProfUploadBucket
		ctx.Endpoint = normalizeEndpoint(cfg.HProfUploadEndpoint)
	}

	key := renderHProfTemplate(template, ctx)
	key = strings.TrimPrefix(path.Clean("/"+key), "/")
	if key == "." {
		return ctx.Filename
	}
	return key
}

func renderHProfDownloadURL(cfg *Config, provider string, ctx hprofTemplateContext) string {
	if cfg != nil && cfg.HProfDownloadURLTemplate != "" {
		return renderHProfTemplate(cfg.HProfDownloadURLTemplate, ctx)
	}

	endpoint := normalizeEndpoint(ctx.Endpoint)
	if endpoint == "" || ctx.Bucket == "" || ctx.ObjectKey == "" {
		return ""
	}
	escapedKey := escapeObjectKey(ctx.ObjectKey)
	if provider == hprofUploadProviderOSS {
		return strings.TrimRight(endpoint, "/") + "/" + ctx.Bucket + "/" + escapedKey
	}
	if cfg != nil && !cfg.hprofUploadS3PathStyle() {
		u, err := url.Parse(endpoint)
		if err == nil && u.Host != "" {
			u.Host = ctx.Bucket + "." + u.Host
			u.Path = "/" + ctx.ObjectKey
			return u.String()
		}
	}
	return strings.TrimRight(endpoint, "/") + "/" + ctx.Bucket + "/" + escapedKey
}

func renderHProfTemplate(template string, ctx hprofTemplateContext) string {
	replacer := strings.NewReplacer(
		"{service}", ctx.Service,
		"{pod_name}", ctx.PodName,
		"{pod_namespace}", ctx.PodNamespace,
		"{host}", ctx.Host,
		"{pid}", strconv.Itoa(int(ctx.PID)),
		"{timestamp}", ctx.Timestamp,
		"{filename}", ctx.Filename,
		"{profiling_path}", ctx.ProfilingPath,
		"{bucket}", ctx.Bucket,
		"{endpoint}", ctx.Endpoint,
		"{object_key}", ctx.ObjectKey,
	)
	return replacer.Replace(template)
}

func buildHProfTemplateContext(service string, pid int32, filePath string, profilingPath string, tags []string, detectedAt time.Time) hprofTemplateContext {
	tagMap := tagsToMap(tags)
	return hprofTemplateContext{
		Service:       firstNonEmpty(service, tagMap["service"]),
		PodName:       tagMap["pod_name"],
		PodNamespace:  tagMap["pod_namespace"],
		Host:          tagMap["host"],
		PID:           pid,
		Timestamp:     detectedAt.UTC().Format("20060102T150405Z"),
		Filename:      filepath.Base(filePath),
		ProfilingPath: profilingPath,
	}
}

func tagsToMap(tags []string) map[string]string {
	m := make(map[string]string, len(tags))
	for _, tag := range tags {
		k, v, ok := strings.Cut(tag, ":")
		if ok && k != "" {
			m[k] = v
		}
	}
	return m
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func normalizeEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return ""
	}
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return strings.TrimRight(endpoint, "/")
	}
	return "https://" + strings.TrimRight(endpoint, "/")
}

func buildS3ObjectURL(cfg *Config, objectKey string) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("config is nil")
	}
	endpoint := normalizeEndpoint(cfg.HProfUploadEndpoint)
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid s3 endpoint %q", cfg.HProfUploadEndpoint)
	}
	if cfg.hprofUploadS3PathStyle() {
		u.Path = path.Join(u.Path, cfg.HProfUploadBucket, objectKey)
		return u.String(), nil
	}
	u.Host = cfg.HProfUploadBucket + "." + u.Host
	u.Path = path.Join(u.Path, objectKey)
	return u.String(), nil
}

func escapeObjectKey(key string) string {
	parts := strings.Split(key, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func fileSHA256Hex(file *os.File) (string, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func uploadHProfForSummary(ctx context.Context, cfg *Config, summary *hprofSummary, tags []string) {
	if cfg == nil || summary == nil || !cfg.hprofUploadEnabled() {
		return
	}

	uploader, err := newHProfUploader(cfg)
	if err != nil {
		summary.HProfUploadStatus = "failed"
		summary.HProfUploadError = err.Error()
		return
	}

	uploadCtx, cancel := context.WithTimeout(ctx, cfg.hprofUploadTimeout())
	defer cancel()

	result, err := uploader.Upload(uploadCtx, &hprofUploadRequest{
		FilePath: summary.HProfPath,
		Context:  buildHProfTemplateContext(summary.Service, summary.PID, summary.HProfPath, cfg.ProfilingPath, tags, summary.DetectedAt),
	})
	if err != nil {
		summary.HProfUploadProvider = strings.ToLower(strings.TrimSpace(cfg.HProfUploadProvider))
		summary.HProfUploadStatus = "failed"
		summary.HProfUploadError = err.Error()
		log.Errorf("hprof upload failed, provider=%s bucket=%s file=%s err=%v",
			summary.HProfUploadProvider, cfg.HProfUploadBucket, summary.HProfPath, err)
		return
	}

	summary.HProfUploadProvider = result.Provider
	summary.HProfObjectKey = result.ObjectKey
	summary.HProfDownloadURL = result.DownloadURL
	summary.HProfUploadStatus = "ok"
	log.Infof("hprof upload ok, provider=%s bucket=%s object_key=%s file=%s download_url=%s",
		result.Provider, result.Bucket, result.ObjectKey, summary.HProfPath, result.DownloadURL)
}
