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
	"github.com/GuanceCloud/cliutils/metrics"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func TestIsGZip(t *T.T) {
	t.Run("is-gzip", func(t *T.T) {
		data := []byte("hello world")

		gz, err := datakit.GZip(data)
		assert.NoError(t, err)

		assert.Equal(t, gzipFlag(1), isGzip(gz))
	})
}

func TestFailCache(t *T.T) {
	t.Run(`test-failcache-data`, func(t *T.T) {
		// server to accept not-sinked points(2 pts)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Logf("%s category: %s", time.Now(), r.URL.Path)
			for k, v := range r.Header {
				t.Logf("%s: %s", k, v)
			}
			w.WriteHeader(http.StatusInternalServerError) // mocked dataway fail
		}))

		defer ts.Close()
		time.Sleep(time.Second)

		reg := prometheus.NewRegistry()

		reg.MustRegister(diskcache.Metrics()...)
		reg.MustRegister(Metrics()...)

		t.Cleanup(func() {
			metricsReset()
			diskcache.ResetMetrics()
		})

		cat := point.Logging

		dw := NewDefaultDataway()
		dw.WAL.Path = t.TempDir()
		dw.MaxRetryCount = 1

		require.NoError(t, dw.Init(WithURLs(ts.URL)))
		require.NoError(t, dw.setupWAL())

		pts := point.RandPoints(100)

		// write logging
		w := getWriter(WithPoints(pts),
			WithCategory(cat),
			WithBodyCallback(func(w *writer, b *body) error {
				return dw.doFlush(w, b)
			}),
			WithHTTPEncoding(dw.contentEncoding))
		defer putWriter(w)

		// POST body to @ts
		assert.NoError(t, w.buildPointsBody())

		// POST fail and the request dumpped to diskcache
		dc := dw.walFail.disk.(*diskcache.DiskCache)
		assert.NoError(t, dc.Rotate()) // force rotate
		require.True(t, dc.Size() > 0)

		// check if BufGet ok
		assert.Equal(t, "always error", dc.BufGet(nil, func(x []byte) error { // not EOF
			t.Logf("get %d bytes", len(x))
			return fmt.Errorf("always error")
		}).Error())

		assert.Equal(t, "error again", dc.BufGet(nil, func(x []byte) error { // auto fallback and not EOF
			t.Logf("get %d bytes", len(x))
			return fmt.Errorf("error again")
		}).Error())

		t.Logf("diskcache:\n%s", dc.Pretty())

		f := dw.newFlusher(cat)

		assert.Error(t, f.cleanFailCache()) // clean cache retry will fail: @ts still return 5XX

		// we can also get the cache from fail-cache: the diskcache will rollback the failed Get().
		assert.NoError(t, dw.walFail.DiskGet(func(b *body) error {
			defer putBody(b)

			if len(b.buf()) == 0 {
				return nil
			}

			// check cached data
			assert.Equal(t, gzipFlag(1), isGzip(b.buf()))
			assert.Equal(t, cat, b.cat())
			assert.Equal(t, point.Protobuf, b.enc())

			// unmarshal payload
			x, err := uhttp.Unzip(b.buf())
			assert.NoError(t, err)

			dec := point.GetDecoder(point.WithDecEncoding(dw.contentEncoding))
			defer point.PutDecoder(dec)

			got, err := dec.Decode(x)
			assert.NoError(t, err)

			assert.Len(t, got, len(pts))

			return nil
		}, withReusableBuffer(f.sendBuf, f.marshalBuf)))

		assert.Equal(t, diskcache.ErrNoData, dc.BufGet(nil, func([]byte) error {
			return nil // make sure no data available in fail-cache
		}))

		t.Logf("diskcache: %s", dc.Pretty())

		mfs, err := reg.Gather()
		assert.NoError(t, err)
		t.Logf("metrics:\n%s", metrics.MetricFamily2Text(mfs))
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

		dw := NewDefaultDataway()
		dw.EnableHTTPTrace = true
		dw.HTTPTimeout = 10 * time.Millisecond // easy timeout
		dw.GZip = true
		dw.WAL.Path = t.TempDir()
		dw.MaxRetryCount = 1

		assert.NoError(t, dw.Init(WithURLs(fmt.Sprintf("%s?token=tkn_11111111111111111111", ts.URL))))
		require.NoError(t, dw.setupWAL())

		pts := point.RandPoints(100)
		cat := point.DynamicDWCategory

		dynURL := fmt.Sprintf("%s/v1/write/logging?token=tkn_for_dialtesting", ts.URL)

		// write dial-testing
		w := getWriter(WithPoints(pts),
			WithCategory(cat),
			WithDynamicURL(dynURL),
			WithBodyCallback(func(w *writer, b *body) error {
				t.Logf("doFlush to %s", w.dynamicURL)
				return dw.doFlush(w, b, WithDynamicURL(dynURL))
			}),
			WithHTTPEncoding(dw.contentEncoding))
		defer putWriter(w)

		// POST dial-testing body to @ts
		assert.NoError(t, w.buildPointsBody())

		WithCategory(point.Metric)(w)                     // try send metric, we need to reset w
		WithBodyCallback(func(w *writer, b *body) error { // change callback
			return dw.doFlush(w, b)
		})(w)
		// POST metric body to @ts
		assert.NoError(t, w.buildPointsBody())

		// check cache content
		dc := dw.walFail.disk.(*diskcache.DiskCache)
		assert.NoError(t, dc.Rotate()) // force rotate

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
		defer ts.Close()

		time.Sleep(time.Second)

		reg := prometheus.NewRegistry()
		reg.MustRegister(diskcache.Metrics()...)
		reg.MustRegister(Metrics()...)

		defer ts.Close()

		dw := NewDefaultDataway()
		dw.EnableHTTPTrace = true
		dw.GZip = true
		dw.WAL.Path = t.TempDir()
		dw.MaxRetryCount = 1

		assert.NoError(t, dw.Init(WithURLs(fmt.Sprintf("%s?token=tkn_11111111111111111111", ts.URL))))
		require.NoError(t, dw.setupWAL())

		cat := point.Logging
		pts := point.RandPoints(100)

		w := getWriter(WithPoints(pts),
			WithCategory(cat),
			WithBodyCallback(func(w *writer, b *body) error {
				return dw.doFlush(w, b)
			}),
			WithHTTPEncoding(dw.contentEncoding))
		defer putWriter(w)

		assert.NoError(t, w.buildPointsBody())

		// check cache content
		dc := dw.walFail.disk.(*diskcache.DiskCache)
		dc.Size()
		assert.NoError(t, dc.Rotate()) // force rotate

		f := dw.newFlusher(cat)

		// try clean cache, but API still failed, and again put to cache
		assert.Contains(t, f.cleanFailCache().Error(), "internal error")

		mfs, err := reg.Gather()
		assert.NoError(t, err)

		t.Logf("metrics: %s", metrics.MetricFamily2Text(mfs))

		m := metrics.GetMetricOnLabels(mfs, "diskcache_get_bytes", dc.Path())
		// only 1 get(in dw.Write with-cache-clean)
		assert.Equal(t, uint64(1), m.GetSummary().GetSampleCount())

		// 1 put(dw.Write with-cache-clean failed do not add another Put)
		m = metrics.GetMetricOnLabels(mfs, "diskcache_put_bytes", dc.Path())
		assert.Equal(t, uint64(1), m.GetSummary().GetSampleCount())

		// put-bytes same as get-bytes: 2 puts only trigger 1 cache,the 2nd do nothing
		mput := metrics.GetMetricOnLabels(mfs, "diskcache_put_bytes", dc.Path()).GetSummary().GetSampleSum()
		mget := metrics.GetMetricOnLabels(mfs, "diskcache_get_bytes", dc.Path()).GetSummary().GetSampleSum()
		assert.Equal(t, 1.0, mput/mget)

		t.Cleanup(func() {
			metricsReset()
			diskcache.ResetMetrics()
		})
	})
}

func TestX(t *T.T) {
	t.Run("write-100pts-with-group", func(t *T.T) {
		cat := point.Logging
		npts := 100

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, cat.URL(), r.URL.Path)

			body, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			assert.NoError(t, err)
			t.Logf("body: %d", len(body))

			var x []byte

			assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

			x, err = uhttp.Unzip(body)
			assert.NoError(t, err)

			dec := point.GetDecoder(point.WithDecEncoding(point.Protobuf))
			defer point.PutDecoder(dec)

			pts, err := dec.Decode(x)
			assert.NoError(t, err)

			assert.Len(t, pts, npts)

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

		pts := point.RandPoints(npts)

		// add extra tags to match group tag key/value
		for _, pt := range pts {
			pt.MustAddTag("tag1", "value1")
			pt.MustAddTag("tag2", "value2")
			// NOTE: tag3 not added
		}

		dw := NewDefaultDataway()
		dw.EnableSinker = true
		dw.GZip = true
		dw.WAL.Path = t.TempDir()
		dw.MaxRetryCount = 1
		dw.GlobalCustomerKeys = nil // no customer keys, all sinker group based on global-tags

		assert.NoError(t, dw.Init(
			WithURLs(fmt.Sprintf("%s?token=tkn_11111111111111111111", ts.URL)),
			WithGlobalTags(map[string]string{ // add global tag as match group tag key/value
				"tag1": "value1",
				"tag2": "value2",
				"tag3": "value3", // not used in random point, so we should not sink on this tag.
			})))
		require.NoError(t, dw.setupWAL())

		assert.NoError(t, dw.Write(WithCategory(point.Logging),
			WithPoints(pts),
			WithHTTPEncoding(dw.contentEncoding),
			WithBodyCallback(func(w *writer, b *body) error {
				t.Logf("body: %s", b)
				return dw.doFlush(w, b)
			}),
		))

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

			dec := point.GetDecoder(point.WithDecEncoding(point.Protobuf))
			defer point.PutDecoder(dec)

			pts, err := dec.Decode(x)
			assert.NoError(t, err)
			assert.Len(t, pts, 100)

			time.Sleep(time.Second) // intended
			w.WriteHeader(200)
		}))

		reg := prometheus.NewRegistry()
		reg.MustRegister(diskcache.Metrics()...)
		reg.MustRegister(Metrics()...)

		pts := point.RandPoints(100)
		dw := NewDefaultDataway()
		dw.GZip = true
		dw.WAL.Path = t.TempDir()
		dw.MaxRetryCount = 1

		assert.NoError(t, dw.Init(WithURLs(fmt.Sprintf("%s?token=tkn_11111111111111111111", ts.URL))))
		require.NoError(t, dw.setupWAL())

		assert.NoError(t, dw.Write(
			WithBodyCallback(func(w *writer, b *body) error {
				return dw.doFlush(w, b)
			}),
			WithHTTPEncoding(dw.contentEncoding),
			WithCategory(point.Logging),
			WithPoints(pts)))

		t.Cleanup(func() {
			ts.Close()
			metricsReset()
			diskcache.ResetMetrics()
		})
	})

	t.Run("write-with-pb", func(t *T.T) {
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

		dw := NewDefaultDataway()
		dw.ContentEncoding = "protobuf"
		dw.GZip = true
		dw.WAL.Path = t.TempDir()
		dw.MaxRetryCount = 1

		assert.NoError(t, dw.Init(WithURLs(fmt.Sprintf("%s?token=tkn_some", ts.URL))))
		require.NoError(t, dw.setupWAL())

		assert.NoError(t, dw.Write(
			WithBodyCallback(func(w *writer, b *body) error {
				return dw.doFlush(w, b)
			}),
			WithHTTPEncoding(dw.contentEncoding),
			WithCategory(point.Logging),
			WithPoints(origin)))
		t.Cleanup(func() {
			ts.Close()
			metricsReset()
			diskcache.ResetMetrics()
		})
	})

	t.Run("write-with-large-pb", func(t *T.T) {
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

		dw := NewDefaultDataway()
		dw.ContentEncoding = "protobuf"
		dw.MaxRawBodySize = 512 * 1024
		dw.GZip = true
		dw.WAL.Path = t.TempDir()
		dw.MaxRetryCount = 1

		assert.NoError(t, dw.Init(WithURLs(fmt.Sprintf("%s?token=tkn_some", ts.URL))))
		require.NoError(t, dw.setupWAL())
		assert.NoError(t, dw.Write(
			WithBodyCallback(func(w *writer, b *body) error {
				return dw.doFlush(w, b)
			}),
			WithHTTPEncoding(dw.contentEncoding),
			WithCategory(point.Logging),
			WithPoints(origin)))

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

	t.Run("write-with-sink", func(t *T.T) {
		pts := []*point.Point{
			point.NewPointV2("m1", point.NewKVs(nil).AddV2("tag1", "val1", true, point.WithKVTagSet(true)).AddV2("f1", 1.23, true)),
			point.NewPointV2("m2", point.NewKVs(nil).AddV2("tag2", "val2", true, point.WithKVTagSet(true)).AddV2("f2", 3.14, true)),
		}

		requests := 0

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests++
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

			t.Logf("body size: %d/%d, pts: %d", len(body), len(x), len(pts))
		}))

		dw := NewDefaultDataway()
		dw.ContentEncoding = "protobuf"
		dw.EnableSinker = true
		dw.GlobalCustomerKeys = []string{"tag1", "tag2"}
		dw.GZip = true
		dw.WAL.Path = t.TempDir()
		dw.MaxRetryCount = 1

		assert.NoError(t, dw.Init(WithURLs(fmt.Sprintf("%s?token=tkn_some", ts.URL))))
		require.NoError(t, dw.setupWAL())

		assert.NoError(t, dw.Write(
			WithBodyCallback(func(w *writer, b *body) error {
				return dw.doFlush(w, b)
			}),
			WithHTTPEncoding(dw.contentEncoding),
			WithPoints(pts),
			WithCategory(point.Logging)))

		// len(pts) == 2, sinked into 2 requests according to the tags.
		assert.Equal(t, 2, requests)

		t.Cleanup(func() {
			ts.Close()
			metricsReset()
		})
	})
}
