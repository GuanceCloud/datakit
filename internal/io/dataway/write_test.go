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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
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
			HTTPTimeout:     10 * time.Millisecond, // easy timeout
		}
		assert.NoError(t, dw.Init())

		pts := dkpt.RandPoints(100)

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
		assert.NoError(t, dw.Write(WithCategory(point.Logging),
			WithFailCache(fc),
			WithPoints(pts)))

		// check cache content
		assert.NoError(t, fc.Rotate())

		// try clean cache, but the cache re-put to cache
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

		// put-bytes twice as get-bytes
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

func TestWritePoints(t *T.T) {
	t.Run("write-100pts-with-group", func(t *T.T) {
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

		pts := dkpt.RandPoints(100)

		dw := &Dataway{
			URLs:         []string{fmt.Sprintf("%s?token=tkn_11111111111111111111", ts.URL)},
			EnableSinker: true,
		}
		assert.NoError(t, dw.Init(
			WithGlobalTags(map[string]string{
				"tag1": "value1",
				"tag2": "value2",
			})))

		assert.NoError(t, dw.Write(WithCategory(point.Logging), WithPoints(pts)))

		t.Cleanup(func() {
			ts.Close()
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

		assert.NoError(t, dw.Write(WithCategory(point.Logging), WithPoints(pts)))

		t.Cleanup(func() {
			ts.Close()
			metricsReset()
			diskcache.ResetMetrics()
		})
	})
}

func TestGroupPoint(t *T.T) {
	t.Run("basic", func(t *T.T) {
		dw := &Dataway{
			URLs: []string{
				"https://fake-dataway.com?token=tkn_xxxxxxxxxx",
			},
			GlobalCustomerKeys: []string{"namespace", "app"},
			EnableSinker:       true,
		}

		assert.NoError(t, dw.Init(WithGlobalTags(map[string]string{
			"tag1": "value1",
			"tag2": "value2",
		})))

		pts := []*dkpt.Point{
			dkpt.MustNewPoint("some", map[string]string{"t1": "v1", "tag1": "new-value"}, map[string]any{"f1": false}, nil),
			dkpt.MustNewPoint("some", map[string]string{"t1": "v1"}, map[string]any{"f1": false}, nil),
			dkpt.MustNewPoint("some", map[string]string{"t1": "v1", "tag2": "new-value"}, map[string]any{"f1": false}, nil),
			dkpt.MustNewPoint("some", nil /* no tags */, map[string]any{"f1": false}, nil),
		}

		res := dw.groupPoints(point.Metric, pts)

		assert.Len(t, res["tag1=new-value,tag2=value2"], 1)
		assert.Len(t, res["tag1=value1,tag2=new-value"], 1)
		assert.Len(t, res["tag1=value1,tag2=value2"], 2)
	})

	t.Run("no-global-tags", func(t *T.T) {
		dw := &Dataway{
			URLs: []string{
				"https://fake-dataway.com?token=tkn_xxxxxxxxxx",
			},

			EnableSinker:       true,
			GlobalCustomerKeys: []string{"namespace", "app"},
		}

		assert.NoError(t, dw.Init())

		pts := []*dkpt.Point{
			dkpt.MustNewPoint("some", map[string]string{"t1": "v1", "tag1": "new-value"}, map[string]any{"f1": false}, nil),
			dkpt.MustNewPoint("some", map[string]string{"t1": "v1"}, map[string]any{"f1": false}, nil),
			dkpt.MustNewPoint("some", map[string]string{"t1": "v1", "tag2": "new-value"}, map[string]any{"f1": false}, nil),
			dkpt.MustNewPoint("some", map[string]string{"namespace": "ns1", "tag2": "new-value"}, map[string]any{"f1": false}, nil),
			dkpt.MustNewPoint("some", nil /* no tags */, map[string]any{"f1": false}, nil),
		}

		res := dw.groupPoints(point.Logging, pts)
		assert.Len(t, res["namespace=ns1"], 1)
		assert.Len(t, res[""], 4)
	})

	t.Run("no-global-tags-on-object", func(t *T.T) {
		dw := &Dataway{
			URLs: []string{
				"https://fake-dataway.com?token=tkn_xxxxxxxxxx",
			},
			GlobalCustomerKeys: []string{"class"},
			EnableSinker:       true,
		}

		assert.NoError(t, dw.Init())

		pts := []*dkpt.Point{
			dkpt.MustNewPoint("some", map[string]string{"t1": "v1", "tag1": "new-value"}, map[string]any{"f1": false}, nil),
			dkpt.MustNewPoint("some", map[string]string{"t1": "v1"}, map[string]any{"f1": false}, nil),
			dkpt.MustNewPoint("some", map[string]string{"t1": "v1", "tag2": "new-value"}, map[string]any{"f1": false}, nil),
			dkpt.MustNewPoint("some", map[string]string{"namespace": "ns1", "tag2": "new-value"}, map[string]any{"f1": false}, nil),
			dkpt.MustNewPoint("some", nil /* no tags */, map[string]any{"f1": false}, nil),
		}

		res := dw.groupPoints(point.Object, pts)

		for k := range res {
			t.Logf("key: %s", k)
		}

		assert.Len(t, res["class=some"], 5)
	})

	t.Run("no-global-tags-no-customer-tag-keys", func(t *T.T) {
		dw := &Dataway{
			URLs: []string{
				"https://fake-dataway.com?token=tkn_xxxxxxxxxx",
			},
			EnableSinker: true,
		}

		assert.NoError(t, dw.Init())

		pts := []*dkpt.Point{
			dkpt.MustNewPoint("some", map[string]string{"t1": "v1", "tag1": "new-value"}, map[string]any{"f1": false}, nil),
			dkpt.MustNewPoint("some", map[string]string{"t1": "v1"}, map[string]any{"f1": false}, nil),
			dkpt.MustNewPoint("some", map[string]string{"t1": "v1", "tag2": "new-value"}, map[string]any{"f1": false}, nil),
			dkpt.MustNewPoint("some", map[string]string{"namespace": "ns1", "tag2": "new-value"}, map[string]any{"f1": false}, nil),
			dkpt.MustNewPoint("some", nil /* no tags */, map[string]any{"f1": false}, nil),
		}

		res := dw.groupPoints(point.Object, pts)
		assert.Len(t, res[""], 5)
	})
}
