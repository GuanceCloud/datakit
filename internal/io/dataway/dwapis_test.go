// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	T "testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNTP(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.URL.Query().Get("token"))

			n := ntpResp{
				TimestampSec: time.Now().Unix(),
			}
			j, err := json.Marshal(n)
			assert.NoError(t, err)

			w.WriteHeader(200)
			w.Write(j)
		}))

		dw := NewDefaultDataway()
		dw.NTP = &ntp{
			Interval:   time.Second,
			SyncOnDiff: time.Second,
		}

		assert.NoError(t, dw.Init(WithURLs(fmt.Sprintf("%s?token=tkn_xxxxxxxx", ts.URL))))

		diff, err := dw.doTimeDiff()

		assert.NoErrorf(t, err, "dataway: %+#v", dw)

		assert.Equal(t, int64(0), diff)
	})
}

func TestDWAPIs(t *T.T) {
	t.Run("apis-with-global-tags", func(t *T.T) {
		dw := NewDefaultDataway()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equalf(t, dw.globalTagsHTTPHeaderValue, r.Header.Get(HeaderXGlobalTags), "failed on request %s", r.URL.Path)

			body, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			assert.NoError(t, err)
			t.Logf("%s %s => body: %d", r.Method, r.URL.Path, len(body))
			w.WriteHeader(200)
		}))

		assert.NoError(t, dw.Init(
			WithURLs(fmt.Sprintf("%s?token=tkn_11111111111111111111", ts.URL)),
			WithGlobalTags(map[string]string{
				"tag1": "value1",
				"tag2": "value2",
			})))

		_, err := dw.Pull("some-args")
		assert.NoError(t, err)

		buf := bytes.NewBuffer([]byte(`some log`))
		_, err = dw.UploadLog(buf, "some-host")
		assert.NoError(t, err)

		_, err = dw.DeleteObjectLabels("", nil)
		assert.NoError(t, err)

		_, err = dw.UpsertObjectLabels("", nil)
		assert.NoError(t, err)

		dw.DatawayList()

		buf = bytes.NewBuffer([]byte("some txt"))
		_, err = dw.ElectionHeartbeat("", "", buf)
		assert.NoError(t, err)

		buf = bytes.NewBuffer([]byte("some txt"))
		_, err = dw.Election("", "", buf)
		assert.NoError(t, err)

		_, err = dw.DQLQuery(nil)
		assert.NoError(t, err)

		_, err = dw.WorkspaceQuery(nil)
		assert.NoError(t, err)

		t.Cleanup(func() {
			ts.Close()
		})
	})
}
