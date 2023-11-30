// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package proxy

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	T "testing"
	"time"

	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

type slowReader struct {
	data []byte
	slow time.Duration
}

func (r *slowReader) Read(p []byte) (n int, err error) {
	if r.slow > time.Duration(0) {
		time.Sleep(r.slow)
	}

	if r.data == nil {
		return 0, nil
	}

	return copy(p, r.data), nil
}

func TestProxyHTTPSNoMITM(t *T.T) {
	t.Run(`https`, func(t *T.T) {
		resetMetrics()

		ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		ts.StartTLS()
		defer ts.Close()
		time.Sleep(time.Millisecond * 10) // wait ok

		t.Logf("real host: %s", ts.URL)

		// proxy server
		randPort := testutils.RandPort("tcp")
		p := Input{Bind: "0.0.0.0", Port: randPort, Verbose: true, MITM: false} // no MITM
		p.doInitProxy()

		proxyURLString := fmt.Sprintf("http://0.0.0.0:%d", randPort)
		t.Logf("proxy URL: %q", proxyURLString)

		go func() {
			if err := p.proxyServer.ListenAndServe(); err != nil {
				t.Errorf("ListenAndServe: %s", err.Error())
			}
		}()

		time.Sleep(time.Millisecond * 10) // wait ok

		// client
		cli := http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(proxyURLString)
				},
			},
		}

		method := http.MethodPost
		req, err := http.NewRequest(method, fmt.Sprintf("%s/some/url", ts.URL), nil)

		require.NoError(t, err)

		resp, err := cli.Do(req)
		assert.NoError(t, err)
		_ = resp

		reg := prometheus.NewRegistry()
		reg.MustRegister(allMetrics()...)
		mfs, err := reg.Gather()
		assert.NoError(t, err)

		m := metrics.GetMetricOnLabels(mfs,
			`datakit_input_proxy_api_total`,
			"/some/url",
			method,
		)

		assert.Equalf(t,
			float64(0),
			m.GetCounter().GetValue(),
			"get metrics:\n%s", metrics.MetricFamily2Text(mfs))

		m = metrics.GetMetricOnLabels(mfs,
			`datakit_input_proxy_connect_total`,
			"127.0.0.1",
		)

		assert.Equalf(t,
			float64(1),
			m.GetCounter().GetValue(),
			"get metrics:\n%s", metrics.MetricFamily2Text(mfs))

		m = metrics.GetMetricOnLabels(mfs,
			`datakit_input_proxy_api_latency_seconds`,
			"/some/url",
			method,
			http.StatusText(resp.StatusCode),
		)

		assert.Equalf(t,
			uint64(0),
			m.GetSummary().GetSampleCount(),
			"get metrics:\n%s", metrics.MetricFamily2Text(mfs))

		t.Logf("get metrics:\n%s", metrics.MetricFamily2Text(mfs))
	})
}

func TestProxyHTTPS(t *T.T) {
	t.Run(`https`, func(t *T.T) {
		resetMetrics()

		ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Logf("request: %q", r.URL.Path)

			t.Logf("try read body...")
			n, err := io.ReadAll(r.Body)
			if err != nil {
				t.Logf("ReadAll: %s", err.Error())
				w.WriteHeader(http.StatusRequestTimeout)
			} else {
				t.Logf("read %d bytes", len(n))
				if _, err := w.Write(make([]byte, 8<<16)); err != nil {
					t.Logf("Write: %s", err.Error())
				}
			}
		}))

		ts.StartTLS()
		defer ts.Close()
		time.Sleep(time.Second) // wait ok

		t.Logf("real host: %s", ts.URL)

		// proxy server
		randPort := testutils.RandPort("tcp")
		p := Input{Bind: "0.0.0.0", Port: randPort, Verbose: true, MITM: true}
		p.doInitProxy()

		proxyURLString := fmt.Sprintf("http://0.0.0.0:%d", randPort)
		t.Logf("proxy URL: %q", proxyURLString)

		go func() {
			if err := p.proxyServer.ListenAndServe(); err != nil {
				t.Errorf("ListenAndServe: %s", err.Error())
			}
		}()

		time.Sleep(time.Millisecond * 10) // wait ok

		// client
		cli := http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(proxyURLString)
				},
			},
		}

		method := http.MethodPost
		req, err := http.NewRequest(method, fmt.Sprintf("%s/some/url", ts.URL), nil)

		require.NoError(t, err)

		resp, err := cli.Do(req)
		assert.NoError(t, err)
		_ = resp

		reg := prometheus.NewRegistry()
		reg.MustRegister(allMetrics()...)
		mfs, err := reg.Gather()
		assert.NoError(t, err)

		m := metrics.GetMetricOnLabels(mfs,
			`datakit_input_proxy_api_total`,
			"/some/url",
			method)

		assert.Equalf(t,
			float64(1),
			m.GetCounter().GetValue(),
			"get metrics:\n%s", metrics.MetricFamily2Text(mfs))

		m = metrics.GetMetricOnLabels(mfs,
			`datakit_input_proxy_connect_total`,
			"127.0.0.1")

		assert.Equalf(t,
			float64(1),
			m.GetCounter().GetValue(),
			"get metrics:\n%s", metrics.MetricFamily2Text(mfs))

		m = metrics.GetMetricOnLabels(mfs,
			`datakit_input_proxy_api_latency_seconds`,
			"/some/url",
			method,
			http.StatusText(resp.StatusCode),
		)

		assert.Equalf(t,
			uint64(1),
			m.GetSummary().GetSampleCount(),
			"get metrics:\n%s", metrics.MetricFamily2Text(mfs))

		t.Logf("get metrics:\n%s", metrics.MetricFamily2Text(mfs))
	})
}

func TestProxyWarns(t *T.T) {
	t.Run(`slow-client`, func(t *T.T) {
		resetMetrics()

		// real http server
		ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Logf("request: %q", r.URL.Path)

			t.Logf("try read body...")
			n, err := io.ReadAll(r.Body)
			if err != nil {
				t.Logf("ReadAll: %s", err.Error())
				w.WriteHeader(http.StatusRequestTimeout)
			} else {
				t.Logf("read %d bytes", len(n))
				if _, err := w.Write(make([]byte, 8<<16)); err != nil {
					t.Logf("Write: %s", err.Error())
				}
			}
		}))

		ts.Config.ReadTimeout = 10 * time.Millisecond       // set easy read-timeout
		ts.Config.ReadHeaderTimeout = 10 * time.Millisecond // set easy read-timeout

		ts.Start()
		defer ts.Close()

		time.Sleep(time.Millisecond * 10) // wait ok

		// proxy server
		randPort := testutils.RandPort("tcp")
		p := Input{Bind: "0.0.0.0", Port: randPort, Verbose: true, MITM: true}
		p.doInitProxy()

		proxyURLString := fmt.Sprintf("http://0.0.0.0:%d", randPort)
		t.Logf("proxy URL: %q", proxyURLString)

		go func() {
			if err := p.proxyServer.ListenAndServe(); err != nil {
				t.Errorf("ListenAndServe: %s", err.Error())
			}
		}()

		time.Sleep(time.Millisecond * 10) // wait ok

		cli := http.Client{
			Timeout: time.Second * 1,
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(proxyURLString)
				},
			},
		}

		method := http.MethodPost
		req, err := http.NewRequest(method,
			fmt.Sprintf("%s/some/url", ts.URL),
			&slowReader{data: make([]byte, 5<<10), slow: time.Second},
		)

		require.NoError(t, err)

		cli.Do(req) // nolint: errcheck

		reg := prometheus.NewRegistry()
		reg.MustRegister(allMetrics()...)
		mfs, err := reg.Gather()
		assert.NoError(t, err)

		t.Logf("get metrics:\n%s", metrics.MetricFamily2Text(mfs))
	})
}

func TestProxy(t *T.T) {
	t.Run(`test-client-timeout`, func(t *T.T) {
		defer resetMetrics()

		// http server
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(time.Second * 3) // long response time
			t.Logf("request: %q", r.URL.Path)
			w.WriteHeader(200)
		}))

		defer ts.Close()

		time.Sleep(time.Second)

		// proxy server
		randPort := testutils.RandPort("tcp")
		p := Input{Bind: "0.0.0.0", Port: randPort, Verbose: false, MITM: true}
		p.doInitProxy()

		proxyURLString := fmt.Sprintf("http://0.0.0.0:%d", randPort)

		go func() {
			if err := p.proxyServer.ListenAndServe(); err != nil {
				t.Errorf("ListenAndServe: %s", err.Error())
			}
		}()

		time.Sleep(time.Second)

		cli := http.Client{
			Timeout: time.Second, // short timeout
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(proxyURLString)
				},
			},
		}

		method := http.MethodPost
		req, err := http.NewRequest(method, fmt.Sprintf("%s/other/url", ts.URL), nil)
		require.NoError(t, err)

		resp, err := cli.Do(req)
		require.Error(t, err)
		require.Nil(t, resp)

		reg := prometheus.NewRegistry()
		reg.MustRegister(allMetrics()...)
		mfs, err := reg.Gather()
		assert.NoError(t, err)

		t.Logf("get metrics:\n%s", metrics.MetricFamily2Text(mfs))
	})

	t.Run(`basic`, func(t *T.T) {
		t.Skip()
		defer resetMetrics()

		// http server
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(time.Millisecond * 100)
			t.Logf("request: %q", r.URL.Path)
			w.WriteHeader(200)
		}))

		time.Sleep(time.Second)

		// proxy server
		randPort := testutils.RandPort("tcp")
		p := Input{Bind: "0.0.0.0", Port: randPort, Verbose: false, MITM: true}
		p.doInitProxy()

		proxyURLString := fmt.Sprintf("http://0.0.0.0:%d", randPort)

		go func() {
			if err := p.proxyServer.ListenAndServe(); err != nil {
				t.Errorf("ListenAndServe: %s", err.Error())
			}
		}()

		time.Sleep(time.Second)

		cli := http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(proxyURLString)
				},
			},
		}
		method := http.MethodPost
		req, err := http.NewRequest(method, fmt.Sprintf("%s/some/url", ts.URL), nil)
		require.NoError(t, err)

		resp, err := cli.Do(req)
		require.NoError(t, err)
		_ = resp

		reg := prometheus.NewRegistry()
		reg.MustRegister(allMetrics()...)

		mfs, err := reg.Gather()
		assert.NoError(t, err)

		m1 := metrics.GetMetricOnLabels(mfs,
			`datakit_input_proxy_api_total`,
			"/some/url",
			method)

		assert.Equalf(t,
			float64(1),
			m1.GetCounter().GetValue(),
			"get metrics:\n%s", metrics.MetricFamily2Text(mfs))

		m2 := metrics.GetMetricOnLabels(mfs,
			`datakit_input_proxy_api_latency_seconds`,
			"/some/url",
			method,
			http.StatusText(http.StatusOK))

		require.NotNilf(t, m2, "get metrics:\n%s", metrics.MetricFamily2Text(mfs))

		assert.Equalf(t,
			uint64(1),
			m2.GetSummary().GetSampleCount(),
			"get metrics:\n%s", metrics.MetricFamily2Text(mfs))

		t.Logf("get metrics:\n%s", metrics.MetricFamily2Text(mfs))
	})
}

func BenchmarkProxy(b *T.B) {
	cases := []struct {
		name string
		r    *bytes.Buffer
		mitm bool
	}{
		{
			name: "server-with-small-body(no-mitm)",
			r:    bytes.NewBuffer(make([]byte, 1<<10)),
			mitm: false,
		},

		{
			name: "server-with-large-body(no-mitm)",
			r:    bytes.NewBuffer(make([]byte, 4<<20)),
			mitm: false,
		},

		{
			name: "server-with-no-body(no-mitm)",
			r:    nil,
			mitm: false,
		},

		{
			name: "server-with-small-body(mitm)",
			r:    bytes.NewBuffer(make([]byte, 1<<10)),
			mitm: true,
		},

		{
			name: "server-with-large-body(mitm)",
			r:    bytes.NewBuffer(make([]byte, 4<<20)),
			mitm: true,
		},

		{
			name: "server-with-no-body(mitm)",
			r:    nil,
			mitm: true,
		},
	}

	// http server
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))

	ts.StartTLS()
	defer ts.Close()
	time.Sleep(time.Millisecond * 10) // wait ok

	b.Logf("real host: %q", ts.URL)

	b.Cleanup(func() {
		reg := prometheus.NewRegistry()
		reg.MustRegister(allMetrics()...)
		mfs, err := reg.Gather()
		assert.NoError(b, err)

		b.Logf("get metrics:\n%s", metrics.MetricFamily2Text(mfs))
		resetMetrics()
	})

	for _, bc := range cases {
		b.Run(bc.name, func(b *T.B) {
			// proxy server
			randPort := testutils.RandPort("tcp")
			p := Input{Bind: "0.0.0.0", Port: randPort, Verbose: false, MITM: bc.mitm}
			p.doInitProxy()

			proxyURLString := fmt.Sprintf("http://0.0.0.0:%d", randPort)

			go func() {
				if err := p.proxyServer.ListenAndServe(); err != nil {
					b.Logf("ListenAndServe: %s", err.Error())
				}
			}()

			time.Sleep(time.Millisecond * 100) // wait proxy ok

			cli := http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
					Proxy: func(req *http.Request) (*url.URL, error) {
						return url.Parse(proxyURLString)
					},
				},
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if bc.r == nil {
					req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/other/url", ts.URL), nil)
					assert.NoError(b, err)

					resp, err := cli.Do(req)
					assert.NoError(b, err)

					if resp != nil {
						resp.Body.Close()
					}
				} else {
					req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/other/url", ts.URL), bc.r)
					assert.NoError(b, err)

					resp, err := cli.Do(req)
					assert.NoError(b, err)

					if resp != nil {
						resp.Body.Close()
					}
				}
			}
		})
	}
}
