// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	T "testing"
	"time"

	"github.com/GuanceCloud/cliutils/metrics"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func TestEndpointRetry(t *T.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	urlstr := fmt.Sprintf("%s?token=abc", ts.URL)
	ep, err := newEndpoint(urlstr, withHTTPTrace(true))
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", urlstr, nil)
	assert.NoError(t, err)

	resp, err := ep.sendReq(req)
	assert.Error(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	t.Logf("resp: %+#v\nerr: %s", resp, err)
}

func TestEndpointMetrics(t *T.T) {
	t.Run("5xx-request", func(t *T.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))

		reg := prometheus.NewRegistry()
		reg.MustRegister(Metrics()...)

		t.Cleanup(func() {
			ts.Close()
			metricsReset()
		})

		api := "/some/path"
		urlstr := fmt.Sprintf("%s%s?token=abc", ts.URL, api)
		ep, err := newEndpoint(urlstr, withHTTPTrace(true))
		assert.NoError(t, err)

		req, err := http.NewRequest("POST", urlstr, nil)
		assert.NoError(t, err)

		_, err = ep.sendReq(req)
		if err != nil {
			t.Logf("%s", err)
		}

		mfs, err := reg.Gather()
		assert.NoError(t, err)

		m := metrics.GetMetricOnLabels(mfs,
			`datakit_io_http_retry_total`,
			api,
			http.StatusText(http.StatusInternalServerError))

		t.Logf("get metrics:\n%s", metrics.MetricFamily2Text(mfs))

		assert.Equal(t, float64(DefaultRetryCount), m.GetCounter().GetValue())
	})

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
			body, err := io.ReadAll(r.Body)
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
			category: point.Metric,
			points: []*point.Point{
				point.NewPointV2("test-1", point.NewKVs(map[string]any{"f1": 1, "f2": false}), point.WithTime(time.Unix(0, 123))),
				point.NewPointV2("test-2", point.NewKVs(map[string]any{"f1": 1, "f2": false}), point.WithTime(time.Unix(0, 123))),
			},
		}

		assert.NoError(t, ep.writePoints(s))

		mfs, err := reg.Gather()
		require.NoError(t, err)
		t.Logf("get metrics:\n%s", metrics.MetricFamily2Text(mfs))

		m := metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_api_latency_seconds`,
			point.Metric.URL(),
			http.StatusText(http.StatusBadRequest))
		assert.Equal(t, uint64(1), m.GetSummary().GetSampleCount())
		assert.True(t, m.GetSummary().GetSampleSum() > 0.0)

		m = metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_point_total`,
			point.Metric.String(),
			http.StatusText(http.StatusBadRequest))
		assert.Equal(t, 2.0, m.GetCounter().GetValue())

		m = metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_point_bytes_total`,
			point.Metric.String(),
			"gzip",
			http.StatusText(http.StatusBadRequest))
		assert.True(t, m.GetCounter().GetValue() > 0)

		t.Cleanup(func() {
			ts.Close()
			metricsReset()
		})
	})

	t.Run("write-n-points-ok", func(t *T.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
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
			category: point.Metric,
			points: []*point.Point{
				point.NewPointV2("test-1", point.NewKVs(map[string]any{"f1": 1, "f2": false}), point.WithTime(time.Unix(0, 123))),
				point.NewPointV2("test-2", point.NewKVs(map[string]any{"f1": 1, "f2": false}), point.WithTime(time.Unix(0, 123))),
			},
		}

		assert.NoError(t, ep.writePoints(w))

		mfs, err := reg.Gather()
		require.NoError(t, err)

		t.Logf("metric: %s", metrics.MetricFamily2Text(mfs))

		m := metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_api_latency_seconds`,
			point.Metric.URL(),
			http.StatusText(http.StatusOK))
		assert.Equal(t, uint64(1), m.GetSummary().GetSampleCount())
		assert.True(t, m.GetSummary().GetSampleSum() > 0.0)

		m = metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_point_total`,
			point.Metric.String(),
			http.StatusText(http.StatusOK))
		assert.Equal(t, 2.0, m.GetCounter().GetValue())

		m = metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_point_bytes_total`,
			point.Metric.String(),
			"gzip",
			http.StatusText(http.StatusOK))
		assert.True(t, m.GetCounter().GetValue() > 0)

		t.Cleanup(func() {
			ts.Close()
			metricsReset()
		})
	})

	t.Run("with-proxy", func(t *T.T) {
		backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
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
			category: point.Metric,
			points: []*point.Point{
				point.NewPointV2("test-1", point.NewKVs(map[string]any{"f1": 1, "f2": false}), point.WithTime(time.Unix(0, 123))),
				point.NewPointV2("test-2", point.NewKVs(map[string]any{"f1": 1, "f2": false}), point.WithTime(time.Unix(0, 123))),
			},
		}

		reg := prometheus.NewRegistry()
		reg.MustRegister(Metrics()...)

		assert.NoError(t, ep.writePoints(w))

		mfs, err := reg.Gather()
		assert.NoError(t, err)

		t.Logf("metric: %s", metrics.MetricFamily2Text(mfs))

		m := metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_api_latency_seconds`,
			point.Metric.URL(),
			http.StatusText(http.StatusOK))
		assert.Equal(t, uint64(1), m.GetSummary().GetSampleCount())
		assert.True(t, m.GetSummary().GetSampleSum() > 0.0)

		m = metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_point_total`,
			point.Metric.String(),
			http.StatusText(http.StatusOK))
		assert.Equal(t, 2.0, m.GetCounter().GetValue())

		m = metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_point_bytes_total`,
			point.Metric.String(),
			"gzip",
			http.StatusText(http.StatusOK))
		assert.True(t, m.GetCounter().GetValue() > 0)

		t.Cleanup(func() {
			metricsReset()
		})
	})
}

func TestRetryGetBodyNil(t *T.T) {
	bodyText := `观测云提供快速实现系统可观测的解决方案，满足云、云原生、应用和业务上的监测需求。
通过自定义监测方案，实现实时可交互仪表板、高效观测基础设施、全链路应用性能可观测等功能，保障系统稳定性`

	reqCnt := 5

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		b, err := io.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Equal(t, bodyText, string(b))

		t.Logf("body: %s, retry: %d", string(b), reqCnt)

		reqCnt--
		if reqCnt > 0 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			_, err := io.WriteString(w, "It works")
			assert.NoError(t, err)
		}
	}))

	time.Sleep(time.Second)

	dw := fmt.Sprintf("%s?token=abc", ts.URL)
	ep, err := newEndpoint(dw, withHTTPTrace(true),
		withMaxRetryCount(reqCnt),
		withRetryDelay(time.Millisecond*100),
	)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, dw, bufio.NewReader(strings.NewReader(bodyText)))
	assert.NoError(t, err)

	// Because the body is not any of *bytes.Buffer, *bytes.Reader and *strings.Reader, but a *bufio.Reader,
	// so req.GetBody is nil
	assert.Nil(t, req.GetBody, "expect req.GetBody nil")

	resp, err := ep.sendReq(req)
	assert.NoErrorf(t, err, "endpoint: %+#v", ep)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
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
