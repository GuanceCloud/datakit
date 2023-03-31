// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"fmt"
	"io/ioutil"
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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

func TestIsGZip(t *T.T) {
	t.Run("is-gzip", func(t *T.T) {
		data := []byte("hello world")

		gz, err := datakit.GZip(data)
		assert.NoError(t, err)

		assert.True(t, isGzip(gz))
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
			HTTPTimeout:     "10ms", // easy timeout
		}
		assert.NoError(t, dw.Init())

		pts := dkpt.RandPoints(100)

		// write dialtesting on category logging
		assert.NoError(t, dw.Write(
			WithCategory(datakit.DynamicDatawayCategory),
			WithFailCache(fc),
			WithPoints(pts), WithDynamicURL(fmt.Sprintf("%s/v1/write/logging?token=tkn_for_dialtesting", ts.URL))))

		// write metric
		assert.NoError(t, dw.Write(WithCategory(datakit.Metric), WithPoints(pts)))

		// check cache content
		assert.NoError(t, fc.Rotate()) // force rotate

		mfs, err := reg.Gather()
		assert.NoError(t, err)

		m := metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_point_total`,
			point.Metric.String(),
			http.StatusText(http.StatusInternalServerError))
		assert.NotNil(t, m)
		assert.Equal(t, float64(100), m.GetCounter().GetValue())

		m = metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_point_total`,
			point.DynamicDWCategory.String(),
			http.StatusText(http.StatusInternalServerError))
		assert.NotNil(t, m)
		assert.Equal(t, float64(100), m.GetCounter().GetValue())

		t.Cleanup(func() {
			assert.NoError(t, fc.Close())
			metricsReset()
			diskcache.ResetMetrics()
		})
	})

	t.Run(`write-with-failcache`, func(t *T.T) {
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
		}
		assert.NoError(t, dw.Init())

		pts := dkpt.RandPoints(100)

		// write logging
		assert.NoError(t, dw.Write(WithCategory(datakit.Logging),
			WithFailCache(fc),
			WithPoints(pts)))

		// check cache content
		assert.NoError(t, fc.Rotate())

		// try clean cache, but the cache re-put to cache
		assert.NoError(t, dw.Write(WithCategory(datakit.Logging),
			WithFailCache(fc),
			WithCacheClean(true)))

		mfs, err := reg.Gather()
		assert.NoError(t, err)

		t.Logf("metrics: %s", metrics.MetricFamily2Text(mfs))

		m := metrics.GetMetricOnLabels(mfs, "diskcache_get_total", fc.Labels()...)
		// only 1 get(in dw.Write with-cache-clean)
		assert.Equal(t, 1.0, m.GetCounter().GetValue())

		// 1 put(dw.Write with-cache-clean failed do not add another Put)
		m = metrics.GetMetricOnLabels(mfs, "diskcache_put_total", fc.Labels()...)
		assert.Equal(t, 1.0, m.GetCounter().GetValue())

		// put-bytes twice as get-bytes
		mput := metrics.GetMetricOnLabels(mfs, "diskcache_put_bytes_total", fc.Labels()...).GetCounter().GetValue()
		mget := metrics.GetMetricOnLabels(mfs, "diskcache_get_bytes_total", fc.Labels()...).GetCounter().GetValue()
		assert.Equal(t, 1.0, mput/mget)

		t.Cleanup(func() {
			assert.NoError(t, fc.Close())
			metricsReset()
			diskcache.ResetMetrics()
		})
	})
}

func TestWritePoints(t *T.T) {
	t.Run("write-100pts-with-sinker", func(t *T.T) {
		// server to accept not-sinked points(2 pts)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, datakit.Logging, r.URL.Path)

			body, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			assert.NoError(t, err)
			t.Logf("body: %d", len(body))

			assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

			x, err := uhttp.Unzip(body)
			assert.NoError(t, err)

			pts, err := lp.ParsePoints(x, nil)
			assert.NoError(t, err)
			assert.Len(t, pts, 100)

			w.WriteHeader(200)
		}))

		// server to accept sinked points(random 100 pts)
		sinkerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, datakit.Logging, r.URL.Path)

			body, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			assert.NoError(t, err)
			t.Logf("body: %d", len(body))

			assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

			x, err := uhttp.Unzip(body)
			assert.NoError(t, err)

			pts, err := lp.ParsePoints(x, nil)
			assert.NoError(t, err)
			assert.Len(t, pts, 2)

			w.WriteHeader(200)
		}))

		reg := prometheus.NewRegistry()
		reg.MustRegister(diskcache.Metrics()...)
		reg.MustRegister(Metrics()...)

		pts := dkpt.RandPoints(100)

		// add 2 sinked-point
		pts = append(pts,
			dkpt.MustNewPoint("sinked-1", nil, map[string]any{"f1": 123}, &dkpt.PointOption{Category: datakit.Logging}),
			dkpt.MustNewPoint("sinked-2", nil, map[string]any{"f2": 123}, &dkpt.PointOption{Category: datakit.Logging}))

		// setup sinker
		sinker := &Sinker{
			Categories: []string{"L"},
			Filters: []string{
				`{source='sinked-1'}`,
				`{source='sinked-2'}`,
			}, // all random points send to default dataway
			URL: sinkerServer.URL,
		}
		assert.NoError(t, sinker.Setup())

		// setup dataway
		dw := &Dataway{
			URLs:    []string{fmt.Sprintf("%s?token=tkn_for_test", ts.URL)},
			Sinkers: []*Sinker{sinker},
		}
		assert.NoError(t, dw.Init())

		// send points via dataway
		assert.NoError(t, dw.Write(WithCategory(datakit.Logging), WithPoints(pts)))

		// check metrics on sinker
		mfs, err := reg.Gather()
		assert.NoError(t, err)

		assert.Equal(t, 1.0, // one sink
			metrics.GetMetricOnLabels(mfs,
				"datakit_io_dataway_sink_total",
				point.Logging.String()).GetCounter().GetValue())

		assert.Equal(t, 2.0, // with 2 points sinked
			metrics.GetMetricOnLabels(mfs,
				"datakit_io_dataway_sink_point_total",
				point.Logging.String(), http.StatusText(http.StatusOK)).GetCounter().GetValue())

		assert.Equal(t, 2.0, // 2 dataway request: each sink got a API request
			metrics.GetMetricOnLabels(mfs,
				"datakit_io_dataway_api_request_total",
				point.Logging.URL(), http.StatusText(http.StatusOK)).GetCounter().GetValue())

		assert.Equal(t, 102.0, // 102 points
			metrics.GetMetricOnLabels(mfs,
				"datakit_io_dataway_point_total",
				point.Logging.String(), http.StatusText(http.StatusOK)).GetCounter().GetValue())

		t.Cleanup(func() {
			ts.Close()
			sinkerServer.Close()

			metricsReset()
			diskcache.ResetMetrics()
		})
	})

	t.Run("write-100pts", func(t *T.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, datakit.Logging, r.URL.Path)

			body, err := ioutil.ReadAll(r.Body)
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

		pts := dkpt.RandPoints(100)

		dw := &Dataway{
			URLs: []string{fmt.Sprintf("%s?token=tkn_11111111111111111111", ts.URL)},
		}
		assert.NoError(t, dw.Init())

		assert.NoError(t, dw.Write(WithCategory(datakit.Logging), WithPoints(pts)))

		t.Cleanup(func() {
			ts.Close()
			metricsReset()
			diskcache.ResetMetrics()
		})
	})
}
