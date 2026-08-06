// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package profile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	bstoml "github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/profile/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/rum"
)

func TestRetryError(t *testing.T) {
	err1 := error(nil)
	err2 := newRetryError(fs.ErrClosed)
	err3 := fmt.Errorf("unable to read: %w", err2)
	err4 := fmt.Errorf("unable to open file: %w", fs.ErrNotExist)

	var rePtr *retryError
	assert.False(t, errors.As(err1, &rePtr))
	assert.True(t, errors.As(err2, &rePtr))
	assert.True(t, errors.As(err3, &rePtr))
	assert.False(t, errors.As(err4, &rePtr))
}

func TestIsPythonPProfMetadata(t *testing.T) {
	assert.True(t, isPythonPProfMetadata(map[string]string{"format": "pprof"}))
	assert.True(t, isPythonPProfMetadata(map[string]string{}))
	assert.False(t, isPythonPProfMetadata(map[string]string{"format": "collapse", "profiler": "pyspy"}))
	assert.False(t, isPythonPProfMetadata(map[string]string{"format": "rawflamegraph", "profiler": "pyspy"}))
}

func TestIOConfig(t *testing.T) {
	testCfg := `
[[inputs.profile]]
  ## profile Agent endpoints register by version respectively.
  ## Endpoints can be skipped listen by remove them from the list.
  ## Default value set as below. DO NOT MODIFY THESE ENDPOINTS if not necessary.
  endpoints = ["/profiling/v1/input"]

  ## set true to enable election, pull mode only
  election = true

  ## the max allowed size of http request body (of MB), 32MB by default.
  body_size_limit_mb = 64 # MB

  ## io_config is used to control profiling uploading behavior.
  ## cache_path set the disk directory where temporarily cache profiling data.
  ## cache_capacity_mb specify the max storage space (in MiB) that profiling cache can use.
  ## clear_cache_on_start set whether we should clear all previous profiling cache on restarting Datakit.
  ## upload_workers set the count of profiling uploading workers.
  ## send_timeout specify the http timeout when uploading profiling data to dataway.
  ## send_retry_count set the max retry count when sending every profiling request.
  [inputs.profile.io_config]
     cache_path = "/usr/local/datakit/cache/profiling_inputs"  # C:\Program Files\datakit\cache\profile_inputs by default on Windows
     cache_capacity_mb = 20480  # 10240MB
     clear_cache_on_start = true
     upload_workers = 16
     send_timeout = "105s"
     send_retry_count = 5
`

	ipt := DefaultInput()
	assert.Equal(t, defaultDiskCachePath(), ipt.IOConfig.CachePath)
	assert.Equal(t, defaultDiskCacheSize, ipt.IOConfig.CacheCapacityMB)
	assert.Equal(t, false, ipt.IOConfig.ClearCacheOnStart)
	assert.Equal(t, defaultHTTPClientTimeout, ipt.IOConfig.SendTimeout)
	assert.Equal(t, defaultHTTPRetryCount, ipt.IOConfig.SendRetryCount)

	type inputs struct {
		Profile []*Input `toml:"profile"`
	}

	type mainCfg struct {
		Inputs inputs `toml:"inputs"`
	}

	mainConf := mainCfg{
		Inputs: inputs{
			Profile: []*Input{ipt},
		},
	}

	_, err := bstoml.Decode(testCfg, &mainConf)
	assert.NoError(t, err)

	assert.Equal(t, 64, ipt.BodySizeLimitMB)
	assert.Equal(t, int64(64<<20), ipt.GetBodySizeLimit())
	assert.Equal(t, "/usr/local/datakit/cache/profiling_inputs", ipt.IOConfig.CachePath)
	assert.Equal(t, 20480, ipt.IOConfig.CacheCapacityMB)
	assert.Equal(t, int64(20480<<20), ipt.getDiskCacheCapacity())
	assert.Equal(t, true, ipt.IOConfig.ClearCacheOnStart)
	assert.Equal(t, 16, ipt.IOConfig.UploadWorkers)
	assert.Equal(t, time.Second*105, ipt.IOConfig.SendTimeout)
	assert.Equal(t, 5, ipt.IOConfig.SendRetryCount)
}

func TestDoSendAddsGlobalTagsHeaderWhenSinkerEnabled(t *testing.T) {
	origDW := config.Cfg.Dataway
	t.Cleanup(func() { config.Cfg.Dataway = origDW })

	dw := dataway.NewDefaultDataway(dataway.WithGlobalTags(map[string]string{
		"env": "prod",
	}))
	dw.EnableSinker = true
	require.NoError(t, dw.Init(dataway.WithURLs("http://127.0.0.1?token=tkn_profile")))
	config.Cfg.Dataway = dw

	var gotHeader string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get(dataway.HeaderXGlobalTags)
		assert.Empty(t, r.Header.Get(dataway.HeaderXGlobalTagsV2))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ipt := DefaultInput()
	ipt.IOConfig.SendRetryCount = 1
	ipt.httpClient = ts.Client()
	profileURL, err := url.Parse(ts.URL + datakit.ProfilingUpload)
	require.NoError(t, err)
	ipt.profileSendingAPI = profileURL

	buf := &bytes.Buffer{}
	mw := multipart.NewWriter(buf)
	eventFile, err := mw.CreateFormFile("event", "event.json")
	require.NoError(t, err)
	_, err = eventFile.Write([]byte(`{
		"attachments": ["main.pprof"],
		"tags_profiler": "service:svc-profile,env:testing,language:go",
		"start": "2022-06-17T09:20:07.002305Z",
		"end": "2022-06-17T09:21:08.261768Z",
		"family": "go"
	}`))
	require.NoError(t, err)
	profileFile, err := mw.CreateFormFile("auto", "main.pprof")
	require.NoError(t, err)
	_, err = profileFile.Write([]byte{0x01, 0x02, 0x03})
	require.NoError(t, err)
	require.NoError(t, mw.Close())

	reqPB := &rum.RequestPB{
		Header: map[string]string{"Content-Type": mw.FormDataContentType()},
		Body:   buf.Bytes(),
	}
	pbBytes, err := proto.Marshal(reqPB)
	require.NoError(t, err)

	require.NoError(t, ipt.sendRequestToDW(context.Background(), pbBytes))
	assert.Equal(t, "env=testing", gotHeader)
}

func TestSendRequestToDWPreservesCollapsedMetadataWhenAddingTags(t *testing.T) {
	var forwardedBody []byte
	var forwardedContentType string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwardedContentType = r.Header.Get("Content-Type")
		var err error
		forwardedBody, err = io.ReadAll(r.Body)
		assert.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ipt := DefaultInput()
	ipt.GenerateMetrics = false
	ipt.Tags = map[string]string{"project": "testing"}
	ipt.IOConfig.SendRetryCount = 1
	ipt.httpClient = ts.Client()
	profileURL, err := url.Parse(ts.URL + datakit.ProfilingUpload)
	require.NoError(t, err)
	ipt.profileSendingAPI = profileURL

	buf := &bytes.Buffer{}
	mw := multipart.NewWriter(buf)
	eventFile, err := mw.CreateFormFile(metrics.EventFile, metrics.EventJSONFile)
	require.NoError(t, err)
	_, err = eventFile.Write([]byte(`{
		"attachments": ["prof"],
		"tags_profiler": "service:front-backend,env:testing,language:python",
		"start": "2026-08-03T07:45:37.479503441Z",
		"end": "2026-08-03T07:46:09.107494696Z",
		"family": "python",
		"format": "collapse",
		"profiler": "pyspy"
	}`))
	require.NoError(t, err)
	profileFile, err := mw.CreateFormFile("prof", "prof")
	require.NoError(t, err)
	_, err = profileFile.Write([]byte("process 7:\"gunicorn: master\";run (app.py:1) 1\n"))
	require.NoError(t, err)
	require.NoError(t, mw.Close())

	reqPB := &rum.RequestPB{
		Header: map[string]string{"Content-Type": mw.FormDataContentType()},
		Body:   buf.Bytes(),
	}
	pbBytes, err := proto.Marshal(reqPB)
	require.NoError(t, err)
	require.NoError(t, ipt.sendRequestToDW(context.Background(), pbBytes))

	forwardedReq, err := http.NewRequest(http.MethodPost, "/", bytes.NewReader(forwardedBody))
	require.NoError(t, err)
	forwardedReq.Header.Set("Content-Type", forwardedContentType)
	require.NoError(t, forwardedReq.ParseMultipartForm(1<<20))

	eventFiles := forwardedReq.MultipartForm.File[metrics.EventFile]
	require.Len(t, eventFiles, 1)
	f, err := eventFiles[0].Open()
	require.NoError(t, err)
	defer f.Close()

	var got metrics.Metadata
	require.NoError(t, json.NewDecoder(f).Decode(&got))
	assert.Equal(t, metrics.Collapsed, got.Format)
	assert.Equal(t, metrics.Profiler("pyspy"), got.Profiler)
	assert.Equal(t, []string{"prof"}, got.Attachments)
	assert.Contains(t, got.TagsProfiler, "project:testing")
}

// go test -v -timeout 30s -run ^Test_originAddTagsSafe$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/profile
func Test_originAddTagsSafe(t *testing.T) {
	cases := []struct {
		name         string
		inOriginTags map[string]string
		inNewKey     string
		inNewVal     string
		expect       map[string]string
	}{
		{
			name:         "add",
			inOriginTags: map[string]string{"a1": "a11", "b1": "b11"},
			inNewKey:     "c1",
			inNewVal:     "c11",
			expect:       map[string]string{"a1": "a11", "b1": "b11", "c1": "c11"},
		},
		{
			name:         "new",
			inOriginTags: map[string]string{},
			inNewKey:     "c1",
			inNewVal:     "c11",
			expect:       map[string]string{"c1": "c11"},
		},
		{
			name:         "empty_key",
			inOriginTags: map[string]string{},
			inNewKey:     "",
			inNewVal:     "c11",
			expect:       map[string]string{},
		},
		{
			name:         "empty_value",
			inOriginTags: map[string]string{},
			inNewKey:     "c1",
			inNewVal:     "",
			expect:       map[string]string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			originAddTagsSafe(tc.inOriginTags, tc.inNewKey, tc.inNewVal)
			assert.Equal(t, tc.expect, tc.inOriginTags)
		})
	}
}

// go test -v -timeout 30s -run ^Test_getPyroscopeTagFromLabels$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/profile
func Test_getPyroscopeTagFromLabels(t *testing.T) {
	cases := []struct {
		name     string
		inLabels map[string]string
		expect   map[string]string
	}{
		{
			name:     "empty",
			inLabels: map[string]string{},
			expect:   map[string]string{},
		},
		{
			name:     "name",
			inLabels: map[string]string{"__name__": "server", "a1": "a11", "a2": "a22"},
			expect:   map[string]string{"a1": "a11", "a2": "a22"},
		},
		{
			name:     "no_name",
			inLabels: map[string]string{"a1": "a11", "a2": "a22"},
			expect:   map[string]string{"a1": "a11", "a2": "a22"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := getPyroscopeTagFromLabels(tc.inLabels)
			assert.Equal(t, tc.expect, out)
		})
	}
}

func TestInput_sendRequestToDW(t *testing.T) {
	originExit := datakit.Exit
	defer func() {
		datakit.Exit = originExit
	}()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	eventJSON := `{
    "attachments": [
        "main.jfr",
        "metrics.json"
    ],
    "tags_profiler": "process_id:31145,service:zy-profiling-test,profiler_version:0.102.0~b67f6e3380,host:zydeMacBook-Air.local,runtime-id:06dddda1-957b-4619-97cb-1a78fc7e3f07,language:jvm,env:test,version:v1.2",
    "start": "2022-06-17T09:20:07.002305Z",
    "end": "2022-06-17T09:21:08.261768Z",
    "family": "java",
    "version": "4",
	"numbers": [1, 3, 5],
	"stable": false
}`

	ipt := DefaultInput()

	buf := &bytes.Buffer{}
	mw := multipart.NewWriter(buf)

	f, err := mw.CreateFormFile("event", "event.json")
	assert.NoError(t, err)
	_, err = f.Write([]byte(eventJSON))
	assert.NoError(t, err)

	f, err = mw.CreateFormFile("auto", "auto.pprof")
	assert.NoError(t, err)
	_, err = f.Write([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06})
	assert.NoError(t, err)
	err = mw.Close()
	assert.NoError(t, err)

	pbOBJ := &rum.RequestPB{
		Header: map[string]string{"Content-Type": mw.FormDataContentType()},
		Body:   buf.Bytes(),
	}

	pbBytes, err := proto.Marshal(pbOBJ)
	assert.NoError(t, err)

	type testCase struct {
		Name     string
		url      string
		httpCli  *http.Client
		ctx      context.Context
		callback func()
		expect   string
	}

	timeCtx, cancelFunc := context.WithCancel(context.Background())
	profileURL := srv.URL + "/profiling/v1/input"

	testCases := []testCase{
		{
			Name:    "Background",
			url:     profileURL,
			httpCli: &http.Client{Timeout: time.Millisecond * 200},
			ctx:     context.Background(),
			expect:  fmt.Sprintf("%d", ipt.IOConfig.SendRetryCount),
		},
		{
			Name:    "CtxCanceled",
			url:     profileURL,
			httpCli: http.DefaultClient,
			ctx:     timeCtx,
			callback: func() {
				go func() {
					time.Sleep(time.Millisecond * 500)
					cancelFunc()
				}()
			},
			expect: ErrRequestCtxCanceled.Error(),
		},
		{
			Name:    "DKExit",
			url:     profileURL,
			httpCli: &http.Client{Timeout: time.Millisecond * 300},
			ctx:     context.TODO(),
			callback: func() {
				go func() {
					time.Sleep(time.Millisecond * 500)
					datakit.Exit.Close()
				}()
			},
			expect: ErrDatakitExiting.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Name == "DKExit" {
				datakit.Exit = cliutils.NewSem()
			}

			ipt.httpClient = tc.httpCli
			ipt.profileSendingAPI, _ = url.Parse(tc.url)
			if tc.callback != nil {
				tc.callback()
			}
			err = ipt.sendRequestToDW(tc.ctx, pbBytes)
			t.Log(err)
			assert.True(t, strings.Contains(err.Error(), tc.expect))
		})
	}
}
