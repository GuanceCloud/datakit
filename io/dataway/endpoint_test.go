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
	"net/http/httputil"
	"net/url"
	"strconv"
	T "testing"
	"time"

	"github.com/GuanceCloud/cliutils/metrics"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/avast/retry-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

func TestEndpointRetry(t *T.T) {
	t.Skip() // for feature test

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	urlstr := fmt.Sprintf("%s?token=abc", ts.URL)
	ep, err := newEndpoint(urlstr, withHTTPTrace(true))
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", urlstr, nil)
	assert.NoError(t, err)

	retry.Do(
		func() error {
			resp, err := ep.sendReq(req)
			if err != nil {
				return err
			}

			if resp.StatusCode/100 == 5 { // server-side error
				t.Logf("status: %s", resp.Status)
				return fmt.Errorf("5XX: %s", resp.Status)
			}

			time.Sleep(time.Second)
			return nil
		},
		retry.Attempts(4),
		retry.Delay(time.Second*1),
	)
}

func TestEndpointFailedRequest(t *T.T) {
	t.Skip() // only for feature test

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	urlstr := fmt.Sprintf("%s?token=abc", ts.URL)
	ep, err := newEndpoint(urlstr, withHTTPTrace(true))
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", urlstr, nil)
	assert.NoError(t, err)
	for {
		resp, err := ep.sendReq(req)
		t.Logf("sendReq: %v, resp: %+#v", err, resp)
		time.Sleep(time.Second)
	}
}

func TestEndpoint(t *T.T) {
	t.Run("new", func(t *T.T) {
		ep, err := newEndpoint("https://openway.guance.com?token=tkn_for_testing", nil)

		assert.NoError(t, err)

		assert.Equal(t, "https", ep.scheme)
		assert.Equal(t, "tkn_for_testing", ep.token)
		assert.Equal(t, "openway.guance.com", ep.host)
		assert.Equal(t, 0, len(ep.categoryURL))
	})

	t.Run("write-points-4xx", func(t *T.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)

			defer r.Body.Close()

			npts, err := strconv.ParseInt(r.Header.Get("X-Points"), 10, 64)
			assert.NoError(t, err)
			assert.True(t, npts > 0)

			for k := range r.Header {
				t.Logf("%s: %s", k, r.Header.Get(k))
			}

			x, err := uhttp.Unzip(body)
			assert.NoError(t, err)

			assert.Equal(t, []byte(`test-1 f1=1i,f2=false 123
test-2 f1=1i,f2=false 123`), x)

			t.Logf("body: %q", x)

			time.Sleep(time.Second) // intended

			w.WriteHeader(http.StatusBadRequest)
		}))

		reg := prometheus.NewRegistry()
		reg.MustRegister(Metrics()...)

		urlstr := fmt.Sprintf("%s?token=abc", ts.URL)
		ep, err := newEndpoint(urlstr, withAPIs([]string{datakit.Metric}))
		assert.NoError(t, err)

		s := &writer{
			category: datakit.Metric,
			pts: []*dkpt.Point{
				dkpt.MustNewPoint("test-1", nil, map[string]any{"f1": 1, "f2": false},
					&dkpt.PointOption{Category: datakit.Metric, Time: time.Unix(0, 123)}),

				dkpt.MustNewPoint("test-2", nil, map[string]any{"f1": 1, "f2": false},
					&dkpt.PointOption{Category: datakit.Metric, Time: time.Unix(0, 123)}),
			},
		}

		assert.NoError(t, ep.writePoints(s))

		mfs, err := reg.Gather()
		require.NoError(t, err)
		t.Logf("get metrics: %s", metrics.MetricFamily2Text(mfs))

		require.Len(t, mfs, 4, "get %d metrics", len(mfs))

		m := metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_api_request_total`,
			point.Metric.URL(),
			http.StatusText(http.StatusBadRequest))
		assert.Equal(t, float64(1), m.GetCounter().GetValue())

		m = metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_api_latency`,
			point.Metric.URL(),
			http.StatusText(http.StatusBadRequest))
		assert.Equal(t, uint64(1), m.GetSummary().GetSampleCount())
		assert.True(t, m.GetSummary().GetSampleSum() > 0.0)

		m = metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_point_total`,
			point.Metric.String(),
			http.StatusText(http.StatusBadRequest))
		assert.Equal(t, float64(2), m.GetCounter().GetValue())

		m = metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_point_bytes_total`,
			point.Metric.String(),
			http.StatusText(http.StatusBadRequest))
		assert.True(t, m.GetCounter().GetValue() > 0)

		t.Cleanup(func() {
			ts.Close()
			metricsReset()
		})
	})

	t.Run("write-n-points-ok", func(t *T.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			assert.NoError(t, err)

			for k := range r.Header {
				t.Logf("%s: %s", k, r.Header.Get(k))
			}

			x, err := uhttp.Unzip(body)
			assert.NoError(t, err)

			assert.Equal(t, []byte(`test-1 f1=1i,f2=false 123
test-2 f1=1i,f2=false 123`), x)

			t.Logf("body: %q", x)

			time.Sleep(time.Second) // intended

			w.WriteHeader(200)
		}))

		urlstr := fmt.Sprintf("%s?token=abc", ts.URL)
		ep, err := newEndpoint(urlstr, withAPIs([]string{datakit.Metric}))
		assert.NoError(t, err)

		reg := prometheus.NewRegistry()
		reg.MustRegister(Metrics()...)

		w := &writer{
			category: datakit.Metric,
			pts: []*dkpt.Point{
				dkpt.MustNewPoint("test-1", nil, map[string]any{"f1": 1, "f2": false},
					&dkpt.PointOption{Category: datakit.Metric, Time: time.Unix(0, 123)}),

				dkpt.MustNewPoint("test-2", nil, map[string]any{"f1": 1, "f2": false},
					&dkpt.PointOption{Category: datakit.Metric, Time: time.Unix(0, 123)}),
			},
		}

		assert.NoError(t, ep.writePoints(w))

		mfs, err := reg.Gather()
		require.NoError(t, err)

		t.Logf("metric: %s", metrics.MetricFamily2Text(mfs))

		assert.Equal(t, float64(1),
			metrics.GetMetricOnLabels(mfs,
				`datakit_io_dataway_api_request_total`,
				point.Metric.URL(),
				http.StatusText(http.StatusOK)).GetCounter().GetValue())

		m := metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_api_latency`,
			point.Metric.URL(),
			http.StatusText(http.StatusOK))
		assert.Equal(t, uint64(1), m.GetSummary().GetSampleCount())
		assert.True(t, m.GetSummary().GetSampleSum() > 0.0)

		m = metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_point_total`,
			point.Metric.String(),
			http.StatusText(http.StatusOK))
		assert.Equal(t, float64(2), m.GetCounter().GetValue())

		m = metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_point_bytes_total`,
			point.Metric.String(),
			http.StatusText(http.StatusOK))
		assert.True(t, m.GetCounter().GetValue() > 0)

		t.Cleanup(func() {
			ts.Close()
			metricsReset()
		})
	})

	t.Run("with-proxy", func(t *T.T) {
		backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)

			t.Logf("proxed request headers: %+#v", r.Header)
			assert.NotEmptyf(t, r.Header.Get(`X-Forwarded-For`), "got nothing on header X-Forwarded-For")

			defer r.Body.Close()

			x, err := uhttp.Unzip(body)
			assert.NoError(t, err)

			assert.Equal(t, []byte(`test-1 f1=1i,f2=false 123
test-2 f1=1i,f2=false 123`), x)

			t.Logf("body: %q", x)

			time.Sleep(time.Second) // intended

			w.WriteHeader(200)
		}))
		defer backend.Close()

		backendURL, err := url.Parse(backend.URL)
		assert.NoError(t, err)
		proxyHandler := httputil.NewSingleHostReverseProxy(backendURL)
		frontend := httptest.NewServer(proxyHandler)
		defer frontend.Close()

		urlstr := fmt.Sprintf("%s?token=abc", backend.URL)
		ep, err := newEndpoint(urlstr, withAPIs([]string{datakit.Metric}), withProxy(frontend.URL))
		assert.NoError(t, err)

		w := &writer{
			category: datakit.Metric,
			pts: []*dkpt.Point{
				dkpt.MustNewPoint("test-1", nil, map[string]any{"f1": 1, "f2": false},
					&dkpt.PointOption{Category: datakit.Metric, Time: time.Unix(0, 123)}),

				dkpt.MustNewPoint("test-2", nil, map[string]any{"f1": 1, "f2": false},
					&dkpt.PointOption{Category: datakit.Metric, Time: time.Unix(0, 123)}),
			},
		}

		reg := prometheus.NewRegistry()
		reg.MustRegister(Metrics()...)

		assert.NoError(t, ep.writePoints(w))

		mfs, err := reg.Gather()
		assert.NoError(t, err)

		t.Logf("metric: %s", metrics.MetricFamily2Text(mfs))

		m := metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_api_request_total`,
			point.Metric.URL(),
			http.StatusText(http.StatusOK))
		assert.Equal(t, float64(1), m.GetCounter().GetValue())

		m = metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_api_latency`,
			point.Metric.URL(),
			http.StatusText(http.StatusOK))
		assert.Equal(t, uint64(1), m.GetSummary().GetSampleCount())
		assert.True(t, m.GetSummary().GetSampleSum() > 0.0)

		m = metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_point_total`,
			point.Metric.String(),
			http.StatusText(http.StatusOK))
		assert.Equal(t, float64(2), m.GetCounter().GetValue())

		m = metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_point_bytes_total`,
			point.Metric.String(),
			http.StatusText(http.StatusOK))
		assert.True(t, m.GetCounter().GetValue() > 0)

		t.Cleanup(func() {
			metricsReset()
		})
	})
}
