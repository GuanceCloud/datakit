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
	"github.com/GuanceCloud/cliutils/point"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/promremote/prompb"
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

		feeder := NewBenchmarkMockedFeeder()
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
			buf := getBuffer()
			buf, ok = ipt.collectBody(res, req, buf)
			defer putBuffer(buf)
			bytes = buf.Bytes()
		}
		if !ok {
			return
		}

		promReq := reqPool.Get().(*prompb.WriteRequest)
		if err := promReq.Unmarshal(bytes); err != nil {
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

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = ipt.Parse(promReq.Timeseries, ipt, additionalTags)
		}
	})
}

func TestParseRejectsInvalidPrometheusLabels(t *testing.T) {
	tests := []struct {
		name          string
		parser        Parser
		labels        []prompb.Label
		wantErr       string
		wantPointName string
	}{
		{
			name:   "invalid metric name rejected before measurement name override",
			parser: Parser{MeasurementName: "fixed"},
			labels: []prompb.Label{
				{Name: []byte(model.MetricNameLabel), Value: []byte("go\xff_gc_duration_seconds")},
				{Name: []byte("job"), Value: []byte("prometheus")},
			},
			wantErr: "invalid prometheus metric name",
		},
		{
			name: "invalid label name",
			labels: []prompb.Label{
				{Name: []byte(model.MetricNameLabel), Value: []byte("go_gc_duration_seconds")},
				{Name: []byte("bad-label"), Value: []byte("prometheus")},
			},
			wantErr: "invalid prometheus label name",
		},
		{
			name:   "invalid label value rejected before job as measurement",
			parser: Parser{JobAsMeasurement: true},
			labels: []prompb.Label{
				{Name: []byte(model.MetricNameLabel), Value: []byte("go_gc_duration_seconds")},
				{Name: []byte("job"), Value: []byte("prom\xffetheus")},
			},
			wantErr: "invalid prometheus label value",
		},
		{
			name: "duplicate label name",
			labels: []prompb.Label{
				{Name: []byte(model.MetricNameLabel), Value: []byte("go_gc_duration_seconds")},
				{Name: []byte("job"), Value: []byte("prometheus")},
				{Name: []byte("job"), Value: []byte("other")},
			},
			wantErr: "duplicate prometheus label name",
		},
		{
			name: "out of order label name",
			labels: []prompb.Label{
				{Name: []byte(model.MetricNameLabel), Value: []byte("go_gc_duration_seconds")},
				{Name: []byte("job"), Value: []byte("prometheus")},
				{Name: []byte("instance"), Value: []byte("localhost:9090")},
			},
			wantErr: "prometheus label name \"instance\" is out of order",
		},
		{
			name: "empty label value",
			labels: []prompb.Label{
				{Name: []byte(model.MetricNameLabel), Value: []byte("go_gc_duration_seconds")},
				{Name: []byte("job"), Value: []byte{}},
			},
			wantErr: "empty prometheus label value",
		},
		{
			name:   "valid labels with measurement name",
			parser: Parser{MeasurementName: "fixed"},
			labels: []prompb.Label{
				{Name: []byte(model.MetricNameLabel), Value: []byte("go_gc_duration_seconds")},
				{Name: []byte("job"), Value: []byte("prometheus")},
			},
			wantPointName: "fixed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pts, err := tt.parser.Parse([]prompb.TimeSeries{{
				Labels:  tt.labels,
				Samples: []prompb.Sample{{Value: 1}},
			}}, defaultInput(), map[string]string{})

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				require.Nil(t, pts)
				return
			}
			require.NoError(t, err)
			require.Len(t, pts, 1)
			require.Equal(t, tt.wantPointName, pts[0].Name())
		})
	}
}

// ------ benchmark mock feeder ------

type BenchmarkMockedFeeder struct {
	PTs []*point.Point
}

func NewBenchmarkMockedFeeder() *BenchmarkMockedFeeder {
	return &BenchmarkMockedFeeder{}
}

func (m *BenchmarkMockedFeeder) Feed(category point.Category, pts []*point.Point, opts ...dkio.FeedOption) error {
	return nil
}

func (m *BenchmarkMockedFeeder) FeedLastError(err string, opts ...metrics.LastErrorOption) {}
