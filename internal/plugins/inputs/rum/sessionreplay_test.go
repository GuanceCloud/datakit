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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
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

func TestReplayDiskQueue(t *testing.T) {
	ipt := defaultInput()
	cacheDir := "./session_replay"
	ipt.SessionReplayCfg.CachePath = cacheDir

	defer os.RemoveAll(cacheDir)

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

	assert.ErrorIs(t, err, diskcache.ErrEOF)

	assert.NoError(t, ipt.replayDiskQueue.Close())
}
