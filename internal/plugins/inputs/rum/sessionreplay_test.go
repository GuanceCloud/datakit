// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/diskcache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"google.golang.org/protobuf/proto"
)

func TestSessionReplayHandler(t *testing.T) {
	ipt := defaultInput()
	cacheDir := "./session_replay"
	ipt.SessionReplayCfg.CachePath = cacheDir
	ipt.SessionReplayCfg.CacheCapacity = 4096
	ipt.SessionReplayCfg.ClearCacheOnStart = true
	ipt.SessionReplayCfg.UploadWorkers = 0
	workersCount := 10
	dataCount := 66
	testCnt := workersCount * dataCount

	defer os.RemoveAll(cacheDir)

	dw := dataway.NewDefaultDataway()
	dw.URLs = []string{"https://testing-openway.dataflux.cn?token=xxxxxxxxxxxxxxx"}
	err := dw.Init()
	assert.NoError(t, err)
	useTestDataway(t, dw)

	handle, err := ipt.sessionReplayHandler()
	assert.NoError(t, err)

	serv := httptest.NewServer(handle)
	defer serv.Close()

	contentType, body := buildSessionReplayRequest()
	req, err := http.NewRequest(http.MethodPost, serv.URL, bytes.NewReader(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", contentType)

	for i := 0; i < testCnt; i++ {
		func() {
			req.Body = io.NopCloser(bytes.NewReader(body))
			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)
			defer func(resp *http.Response) {
				err := resp.Body.Close()
				assert.NoError(t, err)
			}(resp)

			assert.Equal(t, 200, resp.StatusCode)
		}()
	}

	assert.NoError(t, ipt.replayDiskQueue.Rotate())

	wg := &sync.WaitGroup{}
	wg.Add(workersCount)

	for i := 0; i < workersCount; i++ {
		go func() {
			for j := 0; j < dataCount; j++ {
				err := ipt.replayDiskQueue.Get(func(dat []byte) error {
					var pb RequestPB
					err := proto.Unmarshal(dat, &pb)
					assert.NoError(t, err)

					assert.Equal(t, req.Header.Get("Content-Type"), pb.Header["Content-Type"])

					assert.Equal(t, body, pb.Body)

					return nil
				})
				assert.NoError(t, err)
			}
			wg.Done()
		}()
	}

	wait := func() <-chan struct{} {
		ch := make(chan struct{})
		go func() {
			wg.Wait()
			close(ch)
		}()
		return ch
	}

	select {
	case <-time.After(time.Minute * 5):
		t.Fatal("timeout")
	case <-wait():
	}

	err = ipt.replayDiskQueue.Get(func(_ []byte) error {
		return nil
	})

	assert.ErrorIs(t, err, diskcache.ErrNoData)

	assert.NoError(t, ipt.replayDiskQueue.Close())
}

func TestUploadSessionReplayAddsGlobalTagsHeader(t *testing.T) {
	dw := dataway.NewDefaultDataway(dataway.WithGlobalTags(map[string]string{
		"env": "prod",
	}))
	dw.EnableSinker = true
	require.NoError(t, dw.Init(dataway.WithURLs("http://127.0.0.1?token=tkn_replay")))
	useTestDataway(t, dw)

	var gotHeader string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, datakit.SessionReplayUpload, r.URL.Path)
		gotHeader = r.Header.Get(dataway.HeaderXGlobalTags)
		assert.Empty(t, r.Header.Get(dataway.HeaderXGlobalTagsV2))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ipt := defaultInput()
	ipt.SessionReplayCfg.SendRetryCount = 1
	ipt.replayHTTPClient = ts.Client()
	ipt.replayUploadAPI = ts.URL + datakit.SessionReplayUpload

	reqPB := &RequestPB{
		Header: map[string]string{
			"Content-Type": "application/octet-stream",
		},
		Body: []byte("session replay payload"),
		FormValues: map[string]*ValuesSlice{
			"env":     {Values: []string{"testing"}},
			"app_id":  {Values: []string{"app-1"}},
			"service": {Values: []string{"svc-rum"}},
		},
	}
	msg, err := proto.Marshal(reqPB)
	require.NoError(t, err)

	require.NoError(t, ipt.uploadSessionReplay(msg))
	assert.Equal(t, "env=testing", gotHeader)
}

func TestReplayDiskQueue(t *testing.T) {
	ipt := defaultInput()
	cacheDir := "./session_replay"
	ipt.SessionReplayCfg.CachePath = cacheDir

	t.Cleanup(func() {
		os.RemoveAll(cacheDir)
	})

	err := ipt.initReplayDiskQueue()
	assert.NoError(t, err)

	for i := 0; i < 10; i++ {
		err = ipt.replayDiskQueue.Put([]byte("hello world"))
		assert.NoError(t, err)
	}

	err = ipt.replayDiskQueue.Close()
	assert.NoError(t, err)

	entries, err := os.ReadDir(cacheDir)
	assert.NoError(t, err)
	assert.True(t, len(entries) > 0)

	ipt.SessionReplayCfg.ClearCacheOnStart = true
	assert.NoError(t, ipt.initReplayDiskQueue())

	err = ipt.replayDiskQueue.Get(func(_ []byte) error {
		return nil
	})

	assert.ErrorIs(t, err, diskcache.ErrNoData)

	assert.NoError(t, ipt.replayDiskQueue.Close())
}

func TestUploadSessionReplayUsesReplayAssetAPI(t *testing.T) {
	contentType, body := buildSessionReplayRequest()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, datakit.SessionReplayAssetUpload, r.URL.Path)
		assert.Equal(t, contentType, r.Header.Get("Content-Type"))
		assert.Empty(t, r.Header.Get(dataway.HeaderXGlobalTags))
		assert.Contains(t, r.Header.Get(dataway.HeaderXGlobalTagsV2), "wgtid=linked+rum")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	dw := dataway.NewDefaultDataway()
	dw.EnableSinker = true
	dw.SinkerHeaderVersion = "v2"
	dw.GlobalCustomerKeys = []string{"wgtid"}
	dw.URLs = []string{ts.URL + "?token=xxxxx"}
	assert.NoError(t, dw.Init())
	useTestDataway(t, dw)

	ipt := defaultInput()
	assert.NoError(t, ipt.initReplayHTTPClient())

	reqPB := &RequestPB{
		Header: map[string]string{
			"Content-Type": contentType,
		},
		Body:    body,
		APIPath: datakit.SessionReplayAssetUpload,
		FormValues: map[string]*ValuesSlice{
			"app_id": {Values: []string{"web_abcdefghijklmn"}},
			"wgtid":  {Values: []string{"linked rum"}},
		},
	}

	pbData, err := proto.Marshal(reqPB)
	assert.NoError(t, err)
	assert.NoError(t, ipt.uploadSessionReplay(pbData))
}

func TestCopyReplayRequestHeadersDropsLegacyRouteMarker(t *testing.T) {
	headers := copyReplayRequestHeaders(http.Header{
		"Content-Type":            []string{"application/octet-stream"},
		legacyReplayAPIPathHeader: []string{datakit.SessionReplayAssetUpload},
		"X-Forwarded-For":         []string{"127.0.0.1"},
	})

	assert.Equal(t, "application/octet-stream", headers["Content-Type"])
	assert.Equal(t, "127.0.0.1", headers["X-Forwarded-For"])
	assert.NotContains(t, headers, legacyReplayAPIPathHeader)
}
