// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package profile

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/rum"

	bstoml "github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
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

	ipt := defaultInput()
	assert.Equal(t, defaultDiskCachePath(), ipt.IOConfig.CachePath)
	assert.Equal(t, defaultDiskCacheSize, ipt.IOConfig.CacheCapacityMB)
	assert.Equal(t, false, ipt.IOConfig.ClearCacheOnStart)
	assert.Equal(t, defaultConsumeWorkersCount, ipt.IOConfig.UploadWorkers)
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
	assert.Equal(t, int64(64<<20), ipt.getBodySizeLimit())
	assert.Equal(t, "/usr/local/datakit/cache/profiling_inputs", ipt.IOConfig.CachePath)
	assert.Equal(t, 20480, ipt.IOConfig.CacheCapacityMB)
	assert.Equal(t, int64(20480<<20), ipt.getDiskCacheCapacity())
	assert.Equal(t, true, ipt.IOConfig.ClearCacheOnStart)
	assert.Equal(t, 16, ipt.IOConfig.UploadWorkers)
	assert.Equal(t, time.Second*105, ipt.IOConfig.SendTimeout)
	assert.Equal(t, 5, ipt.IOConfig.SendRetryCount)
}

func TestMinHeap(t *testing.T) {
	heap := newMinHeap(16)

	fmt.Println(heap.getTop())

	tm1, _ := time.Parse("2006-01-02 15:04:05Z07:00", "2022-06-08 15:04:06Z")
	tm2, _ := time.Parse("2006-01-02 15:04:05Z07:00", "2022-06-06 15:04:06Z")
	// tm3, _ := time.Parse("2006-01-02 15:04:05Z07:00", "2022-06-05 15:04:06Z")
	tm4, _ := time.Parse("2006-01-02 15:04:05Z07:00", "2022-06-07 15:04:06Z")
	tm5, _ := time.Parse("2006-01-02 15:04:05Z07:00", "2022-06-08 15:04:06Z")

	pb1 := &profileBase{
		profileID: "1111111111111111",
		birth:     tm1,
		point:     nil,
	}
	heap.push(pb1)

	pb2 := &profileBase{
		profileID: "222222222222222",
		birth:     tm2,
		point:     nil,
	}

	heap.push(pb2)

	pb3 := &profileBase{
		profileID: "3333333333333333",
		birth:     tm5,
		point:     nil,
	}

	heap.push(pb3)

	fmt.Println("top: ", heap.getTop())
	fmt.Println("heap.Len: ", heap.Len())
	fmt.Println(heap.indexes)

	pb4 := &profileBase{
		profileID: "44444444444444",
		birth:     tm4,
		point:     nil,
	}

	heap.push(pb4)

	fmt.Println("top: ", heap.getTop())
	fmt.Println("heap.Len: ", heap.Len())
	fmt.Println(heap.indexes)

	pb := heap.pop()

	fmt.Println(pb == pb2)

	fmt.Println("top: ", heap.getTop())
	fmt.Println("heap.Len: ", heap.Len())
	fmt.Println(heap.indexes)

	heap.remove(pb3)

	fmt.Println("top: ", heap.getTop())
	fmt.Println("heap.Len: ", heap.Len())
	fmt.Println(heap.indexes)

	heap.remove(pb1)

	fmt.Println("top: ", heap.getTop())
	fmt.Println("heap.Len: ", heap.Len())
	fmt.Println(heap.indexes)

	heap.push(pb2)

	fmt.Println("top: ", heap.getTop())
	fmt.Println("heap.Len: ", heap.Len())
	fmt.Println(heap.indexes)
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
	ipt := defaultInput()

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

	testCases := []testCase{
		{
			Name:    "Background",
			url:     "https://www.google.com/profiling/v1/input",
			httpCli: &http.Client{Timeout: time.Millisecond * 200},
			ctx:     context.Background(),
			expect:  fmt.Sprintf("%d", ipt.IOConfig.SendRetryCount),
		},
		{
			Name:    "CtxCanceled",
			url:     "https://www.google.com/profiling/v1/input",
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
			url:     "https://www.google.com/profiling/v1/input",
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
