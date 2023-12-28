// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promremote

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/gogo/protobuf/proto"
	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/require"
)

// $ go test -benchmem -bench Benchmark_Parse -run=^$ -cpu=1  -cpuprofile=cpu.pprof -memprofile=mem.pprof
func Benchmark_Parse(b *testing.B) {
	b.Run("old way", func(b *testing.B) {
		conf := `
		[tags]
		  tag1 = "some_value"
		`
		ipt := defaultInput()
		_, err := toml.Decode(conf, ipt)
		require.NoError(b, err)

		feeder := &blankFeeder{}
		ipt.feeder = feeder
		ipt.tagger = &mockTagger{}
		ipt.Run()
		req := &http.Request{
			Method: http.MethodPut,
			URL: &url.URL{
				Path:     "/prom_remote_write",
				RawQuery: "foo=bar&remoteip=1.2.3.4",
			},
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     make(http.Header),
			Host:       "1.1.1.1",
			Body:       io.NopCloser(bytes.NewReader(mock90pts)),
		}
		res := httpResponseWriter{}

		var bytes []byte
		var ok bool
		switch strings.ToLower(ipt.DataSource) {
		case query:
			bytes, ok = ipt.collectQuery(res, req)
		default:
			bytes, ok = ipt.collectBody(res, req)
		}
		if !ok {
			return
		}

		promReq := reqPool.Get().(*prompb.WriteRequest)
		if err := proto.Unmarshal(bytes, promReq); err != nil {
			l.Errorf("unable to unmarshal request body: %w", err)
		}
		defer func() {
			promReq.Reset()
			reqPool.Put(promReq)
		}()

		additionalTags := map[string]string{}

		for k, v := range ipt.mergedTags {
			additionalTags[k] = v
		}

		// Add query tags.
		for k, v := range req.URL.Query() {
			if len(v) > 0 {
				additionalTags[k] = v[0]
			}
		}

		for i := 0; i < b.N; i++ {
			_, _ = ipt.Parse(promReq.Timeseries, ipt, additionalTags)
		}
	})
}
