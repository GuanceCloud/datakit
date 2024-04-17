// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
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
}
