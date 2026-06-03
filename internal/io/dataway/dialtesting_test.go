// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	T "testing"
	"time"

	"github.com/GuanceCloud/cliutils/diskcache"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
)

func TestDTSender(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		var get []*point.Point

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, point.LineProtocol.HTTPContentType(), r.Header.Get("Content-Type"))
			assert.Equal(t, "", r.Header.Get("Content-Encoding")) // default not gzip
			assert.Equal(t, "dialtesting", r.Header.Get("X-Sub-Category"))

			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)

			dec := point.GetDecoder(point.WithDecEncoding(point.LineProtocol))
			defer point.PutDecoder(dec)

			pts, err := dec.Decode(body)
			assert.NoError(t, err)
			get = append(get, pts...)

			t.Logf("body size: %d, pts: %d", len(body), len(pts))
		}))

		time.Sleep(time.Second)

		ds := &DialtestingSender{}
		assert.NoError(t, ds.Init(nil))

		r := point.NewRander()
		pts := r.Rand(10)

		assert.NoError(t, ds.WriteData(fmt.Sprintf("%s?token=tkn_some", ts.URL), pts))

		assert.Len(t, get, len(pts))

		t.Cleanup(func() {
			ts.Close()
			metricsReset()
			diskcache.ResetMetrics()
		})
	})

	t.Run(`browser screenshot upload`, func(t *T.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/write/dialtesting_browser_screenshot", r.URL.Path)
			assert.Equal(t, "tkn_some", r.URL.Query().Get("token"))
			assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")

			assert.NoError(t, r.ParseMultipartForm(1024*1024))
			assert.Equal(t, "task-1", r.FormValue("task_id"))
			assert.Equal(t, "run-1", r.FormValue("run_id"))
			assert.Equal(t, "2", r.FormValue("step_seq"))

			file, header, err := r.FormFile("file")
			assert.NoError(t, err)
			defer file.Close() //nolint:errcheck

			body, err := io.ReadAll(file)
			assert.NoError(t, err)
			assert.Equal(t, "step-image", string(body))
			assert.Equal(t, "step.png", header.Filename)

			_ = json.NewEncoder(w).Encode(BrowserScreenshotUploadResult{
				ScreenshotID:   "run-1-step-2",
				ScreenshotDate: "20260528",
				FileName:       "run-1-step-2.png",
				FileSize:       int64(len(body)),
				SHA256:         "sha",
			})
		}))
		defer ts.Close()

		ds := &DialtestingSender{}
		assert.NoError(t, ds.Init(nil))

		got, err := ds.UploadBrowserScreenshot(
			fmt.Sprintf("%s/v1/write/dialtesting_browser_screenshot?token=tkn_some", ts.URL),
			&BrowserScreenshotUpload{
				FileName:    "step.png",
				ContentType: "image/png",
				Data:        []byte("step-image"),
				TaskID:      "task-1",
				RunID:       "run-1",
				StepSeq:     "2",
			},
		)
		assert.NoError(t, err)
		if assert.NotNil(t, got) {
			assert.Equal(t, "run-1-step-2", got.ScreenshotID)
			assert.Equal(t, "20260528", got.ScreenshotDate)
			assert.Equal(t, "run-1-step-2.png", got.FileName)
			assert.Equal(t, int64(len("step-image")), got.FileSize)
			assert.Equal(t, "sha", got.SHA256)
		}
	})

	t.Run(`browser screenshot upload wrapped response`, func(t *T.T) {
		result, err := decodeBrowserScreenshotUploadResult([]byte(`{
			"content": {
				"screenshot_id": "run-1-step-2",
				"screenshot_date": "20260528",
				"file_name": "run-1-step-2.png",
				"file_size": 10,
				"sha256": "sha"
			}
		}`))
		assert.NoError(t, err)
		if assert.NotNil(t, result) {
			assert.Equal(t, "run-1-step-2", result.ScreenshotID)
			assert.Equal(t, "20260528", result.ScreenshotDate)
			assert.Equal(t, "run-1-step-2.png", result.FileName)
			assert.Equal(t, int64(10), result.FileSize)
			assert.Equal(t, "sha", result.SHA256)
		}
	})
}
