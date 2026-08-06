// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"google.golang.org/protobuf/proto"
)

func TestHandleReplayAssetsUpload(t *testing.T) {
	cacheDir := t.TempDir()

	ipt := defaultInput()
	ipt.SessionReplayCfg.CachePath = cacheDir
	ipt.SessionReplayCfg.CacheCapacity = 64
	ipt.SessionReplayCfg.ClearCacheOnStart = true
	ipt.SessionReplayCfg.UploadWorkers = 0

	dw := dataway.NewDefaultDataway()
	dw.URLs = []string{"https://testing-openway.dataflux.cn?token=xxxxxxxxxxxxxxx"}
	assert.NoError(t, dw.Init())
	useTestDataway(t, dw)
	assert.NoError(t, ipt.initReplayDiskQueue())
	defer ipt.replayDiskQueue.Close()

	router := gin.New()
	wrapper := &httpapi.HandlerWrapper{WrappedResponse: true}
	router.POST(datakit.SessionReplayAssetUpload, wrapper.RawHTTPWrapper(nil, ipt.handleReplayAssetsUpload))

	t.Run("missing-appid", func(t *testing.T) {
		body := &bytes.Buffer{}
		w := multipart.NewWriter(body)
		part, err := w.CreateFormFile("files", "1.png")
		assert.NoError(t, err)
		_, err = part.Write([]byte("img"))
		assert.NoError(t, err)
		assert.NoError(t, w.Close())

		req := httptest.NewRequest(http.MethodPost, datakit.SessionReplayAssetUpload, body)
		req.Header.Set("Content-Type", w.FormDataContentType())

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("enqueue-success", func(t *testing.T) {
		body := &bytes.Buffer{}
		w := multipart.NewWriter(body)
		assert.NoError(t, w.WriteField("appid", "web_abcdefg123456789"))
		assert.NoError(t, w.WriteField("tags", `{"wgtid":"linked-rum"}`))
		part1, err := w.CreateFormFile("files", "1.png")
		assert.NoError(t, err)
		_, err = part1.Write([]byte("img-1"))
		assert.NoError(t, err)
		part2, err := w.CreateFormFile("files", "2.png")
		assert.NoError(t, err)
		_, err = part2.Write([]byte("img-2"))
		assert.NoError(t, err)
		assert.NoError(t, w.Close())

		req := httptest.NewRequest(http.MethodPost, datakit.SessionReplayAssetUpload, body)
		req.Header.Set("Content-Type", w.FormDataContentType())

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Contains(t, resp.Body.String(), `"1.png"`)
		assert.Contains(t, resp.Body.String(), `"2.png"`)

		var cached RequestPB
		assert.NoError(t, ipt.replayDiskQueue.Rotate())
		if !assert.NoError(t, ipt.replayDiskQueue.Get(func(data []byte) error {
			return proto.Unmarshal(data, &cached)
		})) {
			return
		}
		assert.Equal(t, datakit.SessionReplayAssetUpload, cached.GetAPIPath())
		assert.Equal(t, "web_abcdefg123456789", cached.FormValues["app_id"].Values[0])
		assert.Equal(t, "linked-rum", cached.FormValues["wgtid"].Values[0])
	})

	t.Run("invalid-tags", func(t *testing.T) {
		body := &bytes.Buffer{}
		w := multipart.NewWriter(body)
		assert.NoError(t, w.WriteField("appid", "web_abcdefg123456789"))
		assert.NoError(t, w.WriteField("tags", "not-json"))
		part, err := w.CreateFormFile("files", "1.png")
		assert.NoError(t, err)
		_, err = part.Write([]byte("img"))
		assert.NoError(t, err)
		assert.NoError(t, w.Close())

		req := httptest.NewRequest(http.MethodPost, datakit.SessionReplayAssetUpload, body)
		req.Header.Set("Content-Type", w.FormDataContentType())
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("rejects oversized chunked body", func(t *testing.T) {
		body := &countingZeroReader{remaining: ReplayBodyMaxSize + MiB}
		req := httptest.NewRequest(http.MethodPost, datakit.SessionReplayAssetUpload, body)
		assert.Equal(t, int64(-1), req.ContentLength)
		req.Header.Set("Content-Type", "multipart/form-data")

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, int64(ReplayBodyMaxSize+1), body.read)
	})
}

func TestHandleReplayAssetsCheck(t *testing.T) {
	expectedBody := `{"content":{"1.png":true,"abc.txt":false}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, datakit.SessionReplayAssetCheck, r.URL.Path)
		assert.Contains(t, r.Header.Get(dataway.HeaderXGlobalTags), "app_id=web_abcdefg123456789")
		assert.Contains(t, r.Header.Get(dataway.HeaderXGlobalTags), "category=session_replay")
		assert.Contains(t, r.Header.Get(dataway.HeaderXGlobalTags), "service=replay_assets")
		assert.Contains(t, r.Header.Get(dataway.HeaderXGlobalTags), "wgtid=linked-rum")
		var payload replayAssetCheckReq
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, "linked-rum", payload.Tags["wgtid"])
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(expectedBody))
		assert.NoError(t, err)
	}))
	defer server.Close()

	dw := dataway.NewDefaultDataway()
	dw.EnableSinker = true
	dw.GlobalCustomerKeys = []string{"app_id", "category", "service", "wgtid"}
	dw.URLs = []string{server.URL + "?token=xxxxx"}
	assert.NoError(t, dw.Init())
	useTestDataway(t, dw)

	ipt := defaultInput()
	router := gin.New()
	wrapper := &httpapi.HandlerWrapper{WrappedResponse: true}
	router.POST(datakit.SessionReplayAssetCheck, wrapper.RawHTTPWrapper(nil, ipt.handleReplayAssetsCheck))

	t.Run("bad-request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, datakit.SessionReplayAssetCheck, bytes.NewBufferString(`{"appid":""}`))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("ok", func(t *testing.T) {
		body, err := json.Marshal(map[string]any{
			"appid": "web_abcdefg123456789",
			"files": []string{"1.png", "abc.txt"},
			"tags":  map[string]string{"wgtid": "linked-rum"},
		})
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, datakit.SessionReplayAssetCheck, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.JSONEq(t, expectedBody, resp.Body.String())
	})

	t.Run("rejects oversized chunked body", func(t *testing.T) {
		body := &countingZeroReader{remaining: replayAssetCheckBodyMaxSize + MiB}
		req := httptest.NewRequest(http.MethodPost, datakit.SessionReplayAssetCheck, body)
		assert.Equal(t, int64(-1), req.ContentLength)
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, int64(replayAssetCheckBodyMaxSize+1), body.read)
	})
}

func TestHandleReplayAssetsCheckUsesSinkerHeaderV2(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get(dataway.HeaderXGlobalTags))
		assert.Contains(t, r.Header.Get(dataway.HeaderXGlobalTagsV2), "wgtid=linked+rum")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"content":{}}`))
		assert.NoError(t, err)
	}))
	defer server.Close()

	dw := dataway.NewDefaultDataway()
	dw.EnableSinker = true
	dw.SinkerHeaderVersion = "v2"
	dw.GlobalCustomerKeys = []string{"wgtid"}
	dw.URLs = []string{server.URL + "?token=xxxxx"}
	assert.NoError(t, dw.Init())
	useTestDataway(t, dw)

	ipt := defaultInput()
	router := gin.New()
	wrapper := &httpapi.HandlerWrapper{WrappedResponse: true}
	router.POST(datakit.SessionReplayAssetCheck, wrapper.RawHTTPWrapper(nil, ipt.handleReplayAssetsCheck))

	req := httptest.NewRequest(http.MethodPost, datakit.SessionReplayAssetCheck,
		bytes.NewBufferString(`{"appid":"web_abcdefg123456789","files":["1.png"],"tags":{"wgtid":"linked rum"}}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestHandleReplayAssetsGet(t *testing.T) {
	const body = "binary-image"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, datakit.SessionReplayAssetGet, r.URL.Path)
		assert.Equal(t, "wksp_xxx", r.URL.Query().Get("workspace_uuid"))
		assert.Equal(t, "web_abcdefg123456789", r.URL.Query().Get("app_id"))
		assert.Equal(t, "1.png", r.URL.Query().Get("file"))
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, body)
		assert.NoError(t, err)
	}))
	defer server.Close()

	dw := dataway.NewDefaultDataway()
	dw.URLs = []string{server.URL + "?token=xxxxx"}
	assert.NoError(t, dw.Init())
	useTestDataway(t, dw)

	ipt := defaultInput()
	router := gin.New()
	router.GET(datakit.SessionReplayAssetGet, func(c *gin.Context) {
		ipt.handleReplayAssetsGet(c.Writer, c.Request)
	})

	t.Run("bad-request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, datakit.SessionReplayAssetGet, nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("ok", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet,
			datakit.SessionReplayAssetGet+"?workspace_uuid=wksp_xxx&app_id=web_abcdefg123456789&file=1.png",
			nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "image/png", resp.Header().Get("Content-Type"))
		assert.Equal(t, body, resp.Body.String())
	})
}

type countingZeroReader struct {
	remaining int64
	read      int64
}

func (r *countingZeroReader) Read(p []byte) (int, error) {
	if r.remaining == 0 {
		return 0, io.EOF
	}

	n := int64(len(p))
	if n > r.remaining {
		n = r.remaining
	}
	r.remaining -= n
	r.read += n

	return int(n), nil
}
