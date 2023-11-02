// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/diskcache"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
)

func TestLimitReaderClose(t *testing.T) {
	r := io.NopCloser(strings.NewReader("hello world!!!!!"))

	lr := newLimitReader(r, 10)

	c, err := io.ReadAll(lr)

	assert.Error(t, err)
	assert.Len(t, c, 10)
	assert.ErrorIs(t, err, errLimitReader)

	t.Log(err, len(c), string(c), errors.Is(err, errLimitReader))
}

func buildSessionReplayRequest() (string, []byte) {
	var (
		appID      = "web_abcdefghijklmn"
		viewID     = "8a5fcfb3-9f18-4658-8898-5adcc9684abd"
		appENV     = "production"
		appVersion = "1.0.0"
		sessionID  = "f4b0ba4f-6176-462d-9569-2db109d7b9g9"
	)

	buf := bytes.NewBuffer(nil)
	w := multipart.NewWriter(buf)

	fields := map[string]string{
		"records_count":     "429",
		"index_in_view":     "0",
		"source":            "browser",
		"sdk_version":       "v3.1.0",
		"start":             strconv.FormatInt(time.Now().Add(time.Hour*-1).UnixMilli(), 10),
		"end":               strconv.FormatInt(time.Now().UnixMilli(), 10),
		"app_id":            appID,
		"view_id":           viewID,
		"creation_reason":   "init",
		"session_id":        sessionID,
		"env":               appENV,
		"service":           "session-replay",
		"version":           appVersion,
		"raw_segment_size":  "51897",
		"has_full_snapshot": "true",
		"sdk_name":          "df_web_rum_sdk",
	}

	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			log.Fatal(err)
		}
	}

	wp, err := w.CreateFormFile("segment", "segment")
	if err != nil {
		log.Fatal(err)
	}

	fileBytes, err := os.ReadFile("./testdata/session_replay.dat")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := wp.Write(fileBytes); err != nil {
		log.Fatal(err)
	}

	if err := w.Close(); err != nil {
		log.Fatal(err)
	}

	return w.FormDataContentType(), buf.Bytes()
}

func TestInitDiskQueue(t *testing.T) {
	ipt := defaultInput()
	cacheDir := "./session_replay"
	ipt.SessionReplayCfg.CachePath = cacheDir

	defer os.RemoveAll(cacheDir)

	err := ipt.initDiskQueue()
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
	assert.NoError(t, ipt.initDiskQueue())

	err = ipt.replayDiskQueue.Get(func(_ []byte) error {
		return nil
	})

	assert.ErrorIs(t, err, diskcache.ErrEOF)

	assert.NoError(t, ipt.replayDiskQueue.Close())
}

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

	config.Cfg.Dataway.URLs = []string{"https://testing-openway.dataflux.cn?token=xxxxxxxxxxxxxxx"}
	err := config.Cfg.Dataway.Init()
	assert.NoError(t, err)

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

	assert.ErrorIs(t, err, diskcache.ErrEOF)

	assert.NoError(t, ipt.replayDiskQueue.Close())
}
