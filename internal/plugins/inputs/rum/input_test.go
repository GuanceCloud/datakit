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
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
