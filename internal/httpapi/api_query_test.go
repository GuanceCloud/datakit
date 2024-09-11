// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	T "testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type apiRawQueryMock struct{}

var _ IAPIQueryRaw = (*apiRawQueryMock)(nil)

func (m *apiRawQueryMock) GetTokens() []string {
	return []string{"tkn_mocked"}
}

var expectedBody = `
{
  "content": [
    {
      "series": [
        {
          "name": "disk",
          "columns": [ "time", "inodes_used", "inodes_used_mb", "inodes_total_mb", "used_percent", "inodes_free", "inodes_total", "total", "inodes_free_mb", "free", "inodes_used_percent", "used", "global_tag1", "global_tag2", "mount_point", "host", "device", "fstype" ],
          "values": [
            [ 1710810931302, 333, 0, 0, 15.992896326429, 523955, 524288, 1063256064, 0, 893210624, 0.06351470947265, 170045440, "v1", "v2", "/boot", "tan-vm", "/dev/sda1", "xfs" ],
            [ 1710810931302, 373717, 0, 17, 76.4681684582, 17297899, 17671616, 37558423552, 17, 8838184960, 2.114786785769, 28720238592, "v1", "v2", "/", "tan-vm", "/dev/dm-0", "xfs" ]
          ]
        }
      ],
      "points": null,
      "cost": "36.041516ms",
      "is_running": false,
      "async_id": "",
      "query_parse": {
        "namespace": "metric",
        "sources": {
          "disk": "exact"
        },
        "fields": {},
        "funcs": {}
      },
      "index_name": "",
      "index_store_type": "",
      "query_type": "guancedb",
      "complete": false,
      "index_names": "",
      "scan_completed": false,
      "scan_index": "",
      "next_cursor_time": -1,
      "sample": 1,
      "interval": 0,
      "window": 0
    },
    {
      "series": [
        {
          "columns": [ "time", "time_us", "__docid", "__source", "message", "__namespace", "source", "host", "status", "filename", "filepath", "create_time", "df_metering_size", "message_length", "date_ns", "global_tag1", "global_tag2", "index", "log_read_lines", "log_read_offset", "log_read_time", "service" ],
          "values": [
            [ 1710810611818, 1710810611818376, "L_1710810611818_cnsebtudn41qqv5hgn5g", "default", "Mar 19 09:10:01 tan-vm systemd: Started Session 49467 of user root.", "logging", "default", "tan-vm", "unknown", "messages", "/var/log/messages", 1710810615288, 1, 67, 376670, "v1", "v2", "default", 4624, 27406, 1710810601810919700, "default" ],
            [ 1710810071429, 1710810071429957, "L_1710810071429_cnse7mvm8ulok8f960tg", "default", "Mar 19 09:01:01 tan-vm systemd: Started Session 49466 of user root.", "logging", "default", "tan-vm", "unknown", "messages", "/var/log/messages", 1710810075303, 1, 67, 957894, "v1", "v2", "default", 4623, 27338, 1710810061423048200, "default" ]
          ]
        }
      ],
      "points": null,
      "cost": "60.248712ms",
      "total_hits": 4,
      "is_running": false,
      "async_id": "",
      "query_parse": {
        "namespace": "logging",
        "sources": {
          "default": "exact"
        },
        "fields": {},
        "funcs": {}
      },
      "index_name": "59",
      "index_store_type": "doris",
      "query_type": "",
      "complete": false,
      "index_names": "",
      "scan_completed": false,
      "scan_index": "",
      "next_cursor_time": 1710810071429,
      "sample": 1,
      "interval": 0,
      "window": 0
    }
  ]
}`

func (m *apiRawQueryMock) DQLQuery([]byte) (*http.Response, error) {
	r := httptest.NewRecorder()

	if _, err := r.Write([]byte(expectedBody)); err != nil {
		return nil, err
	}

	return r.Result(), nil
}

func TestAPIRawQuery(t *T.T) {
	router := gin.New()

	router.POST("/ok", RawHTTPWrapper(nil, apiQueryRaw, &apiRawQueryMock{}))
	router.POST("/invalid/handler", RawHTTPWrapper(nil, apiQueryRaw, apiRawQueryMock{}))

	var nilMock *apiRawQueryMock
	router.POST("/nil/handler", RawHTTPWrapper(nil, apiQueryRaw, nilMock))

	ts := httptest.NewServer(router)
	defer ts.Close()

	time.Sleep(time.Second) // wait server ok

	t.Run("basic", func(t *T.T) {
		j := `{
	"token": "tkn_some_token",
	"queries": [
		{
			"query": "M::disk LIMIT 2"
		},
		{
			"query": "L::default LIMIT 2"
		}
	],
	"echo_explain": true
}`

		resp, err := http.Post(fmt.Sprintf("%s%s", ts.URL, "/ok"),
			"application/json",
			bytes.NewBufferString(j))
		assert.NoError(t, err)

		respBody, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, expectedBody, string(respBody))

		for k := range resp.Header {
			t.Logf("%s: %v", k, resp.Header.Get(k))
		}
	})

	t.Run("invalid-API-handler", func(t *T.T) {
		resp, err := http.Post(fmt.Sprintf("%s%s", ts.URL, "/invalid/handler"),
			"application/json", nil)
		assert.NoError(t, err)

		assert.Equal(t, 5, resp.StatusCode/100)

		respBody, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		t.Logf("body: %s", string(respBody))
		assert.Contains(t, string(respBody), ErrInvalidAPIHandler.ErrCode)
	})

	t.Run("nil-API-handler", func(t *T.T) {
		resp, err := http.Post(fmt.Sprintf("%s%s", ts.URL, "/nil/handler"),
			"application/json", nil)
		assert.NoError(t, err)

		assert.Equal(t, 5, resp.StatusCode/100)

		respBody, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		t.Logf("body: %s", string(respBody))
		assert.Contains(t, string(respBody), ErrInvalidAPIHandler.ErrCode)
	})

	t.Run("invalid-body-json", func(t *T.T) {
		resp, err := http.Post(fmt.Sprintf("%s%s", ts.URL, "/ok"),
			"application/json", nil)
		assert.NoError(t, err)

		assert.Equal(t, 4, resp.StatusCode/100)

		respBody, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		t.Logf("body: %s", string(respBody))
		assert.Contains(t, string(respBody), ErrInvalidJSON.ErrCode)
	})
}
