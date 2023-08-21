// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	T "testing"

	"github.com/stretchr/testify/assert"
)

func TestDWAPIs(t *T.T) {
	t.Run("apis-with-global-tags", func(t *T.T) {
		var dw *Dataway

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equalf(t, dw.globalTagsHTTPHeaderValue, r.Header.Get(HeaderXGlobalTags), "failed on request %s", r.URL.Path)

			body, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			assert.NoError(t, err)
			t.Logf("%s %s => body: %d", r.Method, r.URL.Path, len(body))
			w.WriteHeader(200)
		}))

		dw = &Dataway{
			URLs: []string{fmt.Sprintf("%s?token=tkn_11111111111111111111", ts.URL)},
		}

		assert.NoError(t, dw.Init(WithGlobalTags(map[string]string{
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
