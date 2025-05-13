// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	T "testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
)

func TestLimitReaderClose(t *T.T) {
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

func TestZipInject(t *T.T) {
	t.Run(`clean-path`, func(t *T.T) {
		cases := []string{
			"../a/../b",
			"./a//../b",
			"./a//./b",
			"./a/../..",
			"a/../../../../../b/..",
		}

		for _, c := range cases {
			t.Logf("%s -> %s", c, filepath.Clean(c))
		}
	})

	t.Run(`join-path`, func(t *T.T) {
		t.Logf("%s", filepath.Join("/a/b/c", "/d/e"))
		t.Logf("%s", filepath.Join("a/b/c", "/d/e"))
	})
}

func TestEnvVariableHandler(t *testing.T) {
	expectedVariableBody := `{content:{"a":"b"}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedVariableBody))
	}))
	defer server.Close()

	config.Cfg.Dataway.URLs = []string{server.URL + "?token=xxxxx"}
	err := config.Cfg.Dataway.Init()
	assert.NoError(t, err)

	ipt := defaultInput()
	router := gin.New()
	wrapper1 := &httpapi.HandlerWrapper{WrappedResponse: true}

	router.GET("/v1/env_variable", wrapper1.RawHTTPWrapper(nil, ipt.handleEnvVariable))

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(fmt.Sprintf("%s/v1/env_variable", ts.URL))
	assert.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	assert.Equal(t, expectedVariableBody, string(respBody))
}
