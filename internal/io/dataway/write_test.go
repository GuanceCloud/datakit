// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	T "testing"
	"time"

	"github.com/GuanceCloud/cliutils/diskcache"
	lp "github.com/GuanceCloud/cliutils/lineproto"
	"github.com/GuanceCloud/cliutils/metrics"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func TestIsGZip(t *T.T) {
	t.Run("is-gzip", func(t *T.T) {
		data := []byte("hello world")

		gz, err := datakit.GZip(data)
		assert.NoError(t, err)

		assert.True(t, isGzip(gz))
	})
}

func TestFailCache(t *T.T) {
	t.Run(`test-failcache-data`, func(t *T.T) {
		// server to accept not-sinked points(2 pts)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Logf("category: %s", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError) // mocked dataway fail
		}))

		t.Cleanup(func() {
			ts.Close()
		})

		p := t.TempDir()
		fc, err := diskcache.Open(diskcache.WithPath(p))
		assert.NoError(t, err)

		dw := &Dataway{
			URLs: []string{fmt.Sprintf("%s?token=tkn_11111111111111111111", ts.URL)},
			GZip: true,
		}

		assert.NoError(t, dw.Init())

		pts := point.RandPoints(100)

		// write logging
		assert.NoError(t, dw.Write(WithCategory(point.Logging),
			WithFailCache(fc),
			WithPoints(pts)))

		assert.NoError(t, fc.Rotate()) // force rotate

		assert.NoError(t, fc.Get(func(x []byte) error {
			if len(x) == 0 {
				return nil
			}

			pd, err := loadCache(x)
			assert.NoError(t, err)

			// check cached data
			assert.True(t, isGzip(pd.Payload))
			assert.Equal(t, point.Logging, point.Category(pd.Category))
			assert.Equal(t, point.LineProtocol, point.Encoding(pd.PayloadType))

			// unmarshal payload
			r, err := gzip.NewReader(bytes.NewBuffer(pd.Payload))
			assert.NoError(t, err)

			buf := bytes.NewBuffer(nil)
			_, err = io.Copy(buf, r)
			assert.NoError(t, err)

			dec := point.GetDecoder(point.WithDecEncoding(point.LineProtocol))
			defer point.PutDecoder(dec)

			got, err := dec.Decode(buf.Bytes())
			assert.NoError(t, err)

			assert.Len(t, got, len(pts))

			return nil
		}))
	})
}

func TestWriteWithCache(t *T.T) {
	t.Run(`write-with-failcache-on-dynamic-category`, func(t *T.T) {
		// server to accept not-sinked points(2 pts)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable) // mocked dataway timeout(eof)
		}))
		defer ts.Close()

		time.Sleep(time.Second)

		reg := prometheus.NewRegistry()

		reg.MustRegister(diskcache.Metrics()...)
		reg.MustRegister(Metrics()...)

		p := t.TempDir()
		fc, err := diskcache.Open(diskcache.WithPath(p))
		assert.NoError(t, err)

		dw := &Dataway{
			URLs:            []string{fmt.Sprintf("%s?token=tkn_11111111111111111111", ts.URL)},
			EnableHTTPTrace: true,
			HTTPTimeout:     10 * time.Millisecond, // easy timeout
			GZip:            true,
		}
		assert.NoError(t, dw.Init())

		pts := point.RandPoints(100)

		// write dialtesting on category logging
		assert.NoError(t, dw.Write(
			WithCategory(point.DynamicDWCategory),
			WithFailCache(fc),
			WithPoints(pts), WithDynamicURL(fmt.Sprintf("%s/v1/write/logging?token=tkn_for_dialtesting", ts.URL))))

		// write metric
		assert.NoError(t, dw.Write(WithCategory(point.Metric), WithPoints(pts)))

		// check cache content
		assert.NoError(t, fc.Rotate()) // force rotate

		mfs, err := reg.Gather()
		assert.NoError(t, err)

		m := metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_point_total`,
			point.Metric.String(),
			http.StatusText(http.StatusServiceUnavailable))
		assert.NotNil(t, m)
		assert.Equalf(t, float64(100), m.GetCounter().GetValue(), metrics.MetricFamily2Text(mfs))

		m = metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_point_total`,
			point.DynamicDWCategory.String(),
			http.StatusText(http.StatusServiceUnavailable))
		assert.NotNil(t, m)
		assert.Equal(t, float64(100), m.GetCounter().GetValue(), metrics.MetricFamily2Text(mfs))

		t.Cleanup(func() {
			assert.NoError(t, fc.Close())
			metricsReset()
			diskcache.ResetMetrics()
		})
	})

	t.Run(`test-metric-on-write-with-failcache`, func(t *T.T) {
		// server to accept not-sinked points(2 pts)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Logf("category: %s", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError) // mocked dataway fail
		}))

		reg := prometheus.NewRegistry()
		reg.MustRegister(diskcache.Metrics()...)
		reg.MustRegister(Metrics()...)

		defer ts.Close()

		p := t.TempDir()
		fc, err := diskcache.Open(diskcache.WithPath(p))
		assert.NoError(t, err)

		dw := &Dataway{
			URLs:            []string{fmt.Sprintf("%s?token=tkn_11111111111111111111", ts.URL)},
			EnableHTTPTrace: true,
			GZip:            true,
		}
		assert.NoError(t, dw.Init())

		pts := point.RandPoints(100)

		// write logging
		assert.NoError(t, dw.Write(WithCategory(point.Logging),
			WithFailCache(fc),
			WithPoints(pts)))

		// check cache content
		assert.NoError(t, fc.Rotate())

		// try clean cache, but API still failed, and again put to cache
		assert.NoError(t, dw.Write(WithCategory(point.Logging),
			WithFailCache(fc),
			WithCacheClean(true)))

		mfs, err := reg.Gather()
		assert.NoError(t, err)

		t.Logf("metrics: %s", metrics.MetricFamily2Text(mfs))

		m := metrics.GetMetricOnLabels(mfs, "diskcache_get_total", p)
		// only 1 get(in dw.Write with-cache-clean)
		assert.Equal(t, 1.0, m.GetCounter().GetValue())

		// 1 put(dw.Write with-cache-clean failed do not add another Put)
		m = metrics.GetMetricOnLabels(mfs, "diskcache_put_total", p)
		assert.Equal(t, 1.0, m.GetCounter().GetValue())

		// put-bytes same as get-bytes: 2 puts only trigger 1 cache,the 2nd do nothing
		mput := metrics.GetMetricOnLabels(mfs, "diskcache_put_bytes_total", p).GetCounter().GetValue()
		mget := metrics.GetMetricOnLabels(mfs, "diskcache_get_bytes_total", p).GetCounter().GetValue()
		assert.Equal(t, 1.0, mput/mget)

		t.Cleanup(func() {
			assert.NoError(t, fc.Close())
			metricsReset()
			diskcache.ResetMetrics()
		})
	})
}

func TestX(t *T.T) {
	t.Run("write-100pts-with-group", func(t *T.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, datakit.Logging, r.URL.Path)

			body, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			assert.NoError(t, err)
			t.Logf("body: %d", len(body))

			var x []byte

			assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

			x, err = uhttp.Unzip(body)
			assert.NoError(t, err)

			pts, err := lp.ParsePoints(x, nil)
			assert.NoError(t, err)
			assert.Len(t, pts, 100)

			for k := range r.Header {
				t.Logf("%s: %s", k, r.Header.Get(k))
			}

			assert.Equal(t, "tag1=value1,tag2=value2", r.Header.Get(HeaderXGlobalTags))

			time.Sleep(time.Second) // intended
			w.WriteHeader(200)
		}))

		reg := prometheus.NewRegistry()
		reg.MustRegister(diskcache.Metrics()...)
		reg.MustRegister(Metrics()...)

		pts := point.RandPoints(100)

		// add extra tags to match group tag key/value
		for _, pt := range pts {
			pt.MustAddTag("tag1", "value1")
			pt.MustAddTag("tag2", "value2")
		}

		dw := &Dataway{
			URLs:         []string{fmt.Sprintf("%s?token=tkn_11111111111111111111", ts.URL)},
			EnableSinker: true,
			GZip:         true,
			// GlobalCustomerKeys: []string{"tag1", "tag2"},
		}
		assert.NoError(t, dw.Init(
			WithGlobalTags(map[string]string{ // add global tag as match group tag key/value
				"tag1": "value1",
				"tag2": "value2",
				"tag3": "value3", // not used
			})))

		assert.NoError(t, dw.Write(WithCategory(point.Logging), WithPoints(pts)))

		t.Cleanup(func() {
			ts.Close()
			metricsReset()
			diskcache.ResetMetrics()
		})
	})
}

func TestWritePoints(t *T.T) {
	t.Run("write-100pts", func(t *T.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, datakit.Logging, r.URL.Path)

			body, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			assert.NoError(t, err)
			t.Logf("body: %d", len(body))

			var x []byte

			assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

			x, err = uhttp.Unzip(body)
			assert.NoError(t, err)

			pts, err := lp.ParsePoints(x, nil)
			assert.NoError(t, err)
			assert.Len(t, pts, 100)

			time.Sleep(time.Second) // intended
			w.WriteHeader(200)
		}))

		reg := prometheus.NewRegistry()
		reg.MustRegister(diskcache.Metrics()...)
		reg.MustRegister(Metrics()...)

		pts := point.RandPoints(100)

		dw := &Dataway{
			URLs: []string{fmt.Sprintf("%s?token=tkn_11111111111111111111", ts.URL)},
			GZip: true,
		}
		assert.NoError(t, dw.Init())

		assert.NoError(t, dw.Write(WithCategory(point.Logging), WithPoints(pts)))

		t.Cleanup(func() {
			ts.Close()
			metricsReset()
			diskcache.ResetMetrics()
		})
	})

	t.Run("write.with.pb", func(t *T.T) {
		r := point.NewRander()
		origin := r.Rand(10)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, point.Protobuf.HTTPContentType(), r.Header.Get("Content-Type"))
			assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			x, err := uhttp.Unzip(body)
			assert.NoError(t, err)

			dec := point.GetDecoder(point.WithDecEncoding(point.Protobuf))
			defer point.PutDecoder(dec)

			pts, err := dec.Decode(x)
			assert.NoError(t, err)

			assert.Len(t, pts, len(origin))
			for idx, pt := range pts {
				assert.True(t, pt.Equal(origin[idx]))
			}

			t.Logf("body size: %d/%d, pts: %d", len(body), len(x), len(pts))
		}))

		dw := &Dataway{
			URLs:            []string{fmt.Sprintf("%s?token=tkn_some", ts.URL)},
			ContentEncoding: "protobuf",
			GZip:            true,
		}
		assert.NoError(t, dw.Init())

		assert.NoError(t, dw.Write(WithCategory(point.Logging), WithPoints(origin)))
		t.Cleanup(func() {
			ts.Close()
			metricsReset()
			diskcache.ResetMetrics()
		})
	})

	t.Run("write.with.large.pb", func(t *T.T) {
		var (
			r      = point.NewRander(point.WithRandText(3))
			origin = r.Rand(1000)
			get    []*point.Point
		)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, point.Protobuf.HTTPContentType(), r.Header.Get("Content-Type"))
			assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			x, err := uhttp.Unzip(body)
			assert.NoError(t, err)

			dec := point.GetDecoder(point.WithDecEncoding(point.Protobuf))
			defer point.PutDecoder(dec)

			pts, err := dec.Decode(x)
			assert.NoError(t, err)
			get = append(get, pts...)

			t.Logf("body size: %d/%d, pts: %d", len(body), len(x), len(pts))
		}))

		dw := &Dataway{
			URLs:            []string{fmt.Sprintf("%s?token=tkn_some", ts.URL)},
			ContentEncoding: "protobuf",
			MaxRawBodySize:  512 * 1024,
			GZip:            true,
		}

		assert.NoError(t, dw.Init())
		assert.NoError(t, dw.Write(WithCategory(point.Logging), WithPoints(origin)))

		assert.Len(t, get, len(origin))
		for idx, pt := range get {
			assert.True(t, pt.Equal(origin[idx]))
		}

		t.Cleanup(func() {
			ts.Close()
			metricsReset()
			diskcache.ResetMetrics()
		})
	})
}
