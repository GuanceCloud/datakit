package prom

import (
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

const promURL = "http://127.0.0.1:9100/metrics"

const mockBody = `
# HELP promhttp_metric_handler_errors_total Total number of internal errors encountered by the promhttp metric handler.
# TYPE promhttp_metric_handler_errors_total counter
promhttp_metric_handler_errors_total{cause="encoding"} 0
promhttp_metric_handler_errors_total{cause="gathering"} 0
# HELP promhttp_metric_handler_requests_in_flight Current number of scrapes being served.
# TYPE promhttp_metric_handler_requests_in_flight gauge
promhttp_metric_handler_requests_in_flight 1
# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_handler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 15143
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0
go_gc_duration_seconds{quantile="0.25"} 0
go_gc_duration_seconds{quantile="0.5"} 0
# HELP http_request_duration_seconds duration histogram of http responses labeled with: status_code, method
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.003",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="0.03",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="0.1",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="0.3",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="1.5",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="10",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="+Inf",status_code="404",method="GET"} 1
http_request_duration_seconds_sum{status_code="404",method="GET"} 0.002451013
http_request_duration_seconds_count{status_code="404",method="GET"} 1
# HELP up 1 = up, 0 = not up
# TYPE up untyped
up 1
`

const caContent = `-----BEGIN CERTIFICATE-----
MIICqDCCAZACCQC27UZHg8A/CjANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQDDAt0
b255YmFpLmNvbTAeFw0yMTExMjUwMTU3MzBaFw0zNTA4MDQwMTU3MzBaMBYxFDAS
BgNVBAMMC3RvbnliYWkuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKC
AQEAozWNMKEeVVKRg5QuOPv9bmuGOShRWaMxmyLnfzvV5tS/Odg63jEecE3K/HHa
OUrTwHKl2NSfwfUZPfCf1gVYHBzozX66XXXYR+qV2aeg+GsMg+o8foH8mmBL9cW+
fvbpNNv9k9G4W0zX9YdWmXt8KHKr5KThSUq46KN8qUCUPqIBnPMKfDJuEjLMPuxi
hliehoeHY32YcglLKSSAYMos2SWUA/D81wyfZKeH8KQu7lPKdCEXKLJBS4+HxUFx
gwDV84m+H8v9bf8PIeKrUGmzSuCYUCxrQyoiaIawB6iY7BJAQeaqEr+W6bS9BU2f
p6KHG8yEHDfz3gFuNR3vCLzGDQIDAQABMA0GCSqGSIb3DQEBCwUAA4IBAQA8DiJy
o6BHVTumAbBv9+Q0FsXKQRH1YVwR7spHLdbzbqJadghTGPrXGzwYBaGiLTHHaLXX
Ksbdc8T/C7pRIVXc9Jbx1EzCQFlaDBk89okAG/cWcbr0P5sMDJ96UrapBo2PYKNq
QvSQhSjvKTVB19wwSoD7zbOqITXWQKcv1d10yd3X5Q2PComjMMuhWKAOtJvIvEru
/3WiDpYgGGz/XN1YRFnNvRsXEVa6T0Q7lOi/7Lfv+N96R643Zv5fcyAFMGIiQ0na
vfqe/FB05Gl89x+Bb7xti8bzAlsFy1byeIfFKU3Gmvb8INRJyH5wRWVu29poXl1N
g/pAjggcs8zy5GxR
-----END CERTIFICATE-----`

type transportMock struct {
	statusCode int
	body       string
}

func (t *transportMock) RoundTrip(r *http.Request) (*http.Response, error) {
	res := &http.Response{
		Header:     make(http.Header),
		Request:    r,
		StatusCode: t.statusCode,
	}
	res.Body = ioutil.NopCloser(strings.NewReader(t.body))
	return res, nil
}

func (t *transportMock) CancelRequest(_ *http.Request) {
}

func newTransportMock(body string) http.RoundTripper {
	return &transportMock{
		statusCode: http.StatusOK,
		body:       body,
	}
}

func TestCollect(t *testing.T) {
	testcases := []struct {
		in   *Option
		fail bool
	}{
		{
			fail: true,
		},
		{
			in:   &Option{},
			fail: true,
		},
		{
			in: &Option{
				URL:         promURL,
				MetricTypes: []string{},
				Measurements: []Rule{
					{
						Pattern: `^go_.*`,
						Name:    "gogo",
						Prefix:  "go_",
					},
				},
				MetricNameFilter: []string{"go"},
			},
			fail: false,
		},
		{
			in:   &Option{URL: promURL},
			fail: false,
		},
	}

	for _, tc := range testcases {
		p, err := NewProm(tc.in)
		if tc.fail && assert.Error(t, err) {
			continue
		} else {
			assert.NoError(t, err)
		}

		p.SetClient(&http.Client{Transport: newTransportMock(mockBody)})

		pts, err := p.Collect()
		if tc.fail && assert.Error(t, err) {
			continue
		} else {
			assert.NoError(t, err)
		}

		for _, pt := range pts {
			t.Log(pt.String())
		}
	}
}

func TestProm_DebugCollect(t *testing.T) {
	testcases := []struct {
		in   *Option
		fail bool
	}{
		{
			in:   &Option{},
			fail: true,
		},

		{
			in:   &Option{URL: "http://127.0.0.1:9100/metrics"},
			fail: false,
		},
	}

	for _, tc := range testcases {
		p, err := NewProm(tc.in)
		if tc.fail && assert.Error(t, err) {
			continue
		} else {
			assert.NoError(t, err)
		}

		p.SetClient(&http.Client{Transport: newTransportMock(mockBody)})

		pts, err := p.CollectFromFile()
		if tc.fail && assert.Error(t, err) {
			continue
		} else {
			assert.NoError(t, err)
		}

		for _, pt := range pts {
			t.Log(pt.String())
		}
	}
}

func Test_BearerToken(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "token")
	assert.NoError(t, err)
	token_file := f.Name()
	defer os.Remove(token_file) // nolint:errcheck

	testcases := []struct {
		auth    map[string]string
		url     string
		isError bool
	}{
		{
			auth:    map[string]string{},
			url:     "http://localhost",
			isError: true,
		},
		{
			auth:    map[string]string{"token": "xxxxxxxxxx"},
			url:     "http://localhost",
			isError: false,
		},
		{
			auth:    map[string]string{"token_file": "invalid_file"},
			url:     "http://localhost",
			isError: true,
		},
		{
			auth:    map[string]string{"token_file": token_file},
			url:     "http://localhost",
			isError: false,
		},
	}

	for _, tc := range testcases {
		r, err := BearerToken(tc.auth, tc.url)

		if tc.isError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			authHeader, ok := r.Header["Authorization"]
			assert.True(t, ok)
			assert.Equal(t, len(authHeader), 1)
			assert.Contains(t, authHeader[0], "Bearer")
		}
	}
}

func Test_Tls(t *testing.T) {
	t.Run("enable tls", func(t *testing.T) {
		p, err := NewProm(&Option{
			URL:     "http://127.0.0.1:9100",
			TLSOpen: true,
		})
		assert.NoError(t, err)
		transport, ok := p.client.Transport.(*http.Transport)
		assert.True(t, ok)
		assert.Equal(t, transport.TLSClientConfig.InsecureSkipVerify, true)
	})

	t.Run("tls with ca", func(t *testing.T) {
		f, err := ioutil.TempFile(os.TempDir(), "ca.crt")
		assert.NoError(t, err)
		_, err = f.WriteString(caContent)
		assert.NoError(t, err)
		caFile := f.Name()
		defer os.Remove(caFile) // nolint:errcheck
		p, err := NewProm(&Option{
			URL:        "http://127.0.0.1:9100",
			TLSOpen:    true,
			CacertFile: caFile,
		})

		assert.NoError(t, err)
		transport, ok := p.client.Transport.(*http.Transport)
		assert.True(t, ok)
		assert.Equal(t, transport.TLSClientConfig.InsecureSkipVerify, false)
	})
}

func Test_Auth(t *testing.T) {
	p, err := NewProm(&Option{
		URL: promURL,
		Auth: map[string]string{
			"type":  "bearer_token",
			"token": ".....",
		},
	})
	assert.NoError(t, err)
	r, err := p.GetReq(p.opt.URL)
	assert.NoError(t, err)
	authHeader, ok := r.Header["Authorization"]
	assert.True(t, ok)

	assert.Equal(t, len(authHeader), 1)
	assert.Contains(t, authHeader[0], "Bearer ")
}

func Test_Option(t *testing.T) {
	o := Option{
		Disabel: true,
	}
	assert.True(t, o.IsDisable(), o.Disabel)

	// GetSource
	assert.Equal(t, o.GetSource("p"), "p")
	assert.Equal(t, o.GetSource(), "prom")
	o.Source = "p1"
	assert.Equal(t, o.GetSource("p"), "p1")

	// GetIntervalDuration
	assert.Equal(t, o.GetIntervalDuration(), defaultInterval)
	o.interval = 1 * time.Second
	assert.Equal(t, o.GetIntervalDuration(), 1*time.Second)
	o.interval = 0
	o.Interval = "10s"
	assert.Equal(t, o.GetIntervalDuration(), 10*time.Second)
	assert.Equal(t, o.interval, 10*time.Second)
}

func Test_WriteFile(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "output")
	assert.NoError(t, err)
	outputFile := f.Name()
	defer os.Remove(outputFile) // nolint:errcheck

	p, err := NewProm(&Option{
		URL:         promURL,
		Output:      outputFile,
		MaxFileSize: 100000,
	})

	assert.NoError(t, err)
	p.SetClient(&http.Client{Transport: newTransportMock(mockBody)})
	err = p.WriteFile()

	assert.NoError(t, err)

	fileContent, err := ioutil.ReadFile(outputFile)

	assert.NoError(t, err)

	assert.Equal(t, string(fileContent), mockBody)
}

func TestIgnoreReqErr(t *testing.T) {
	testCases := []struct {
		name string
		in   *Option
		fail bool
	}{
		{
			name: "ignore url request error",
			in:   &Option{IgnoreReqErr: true, URL: "127.0.0.1:999999"},
			fail: false,
		},
		{
			name: "do not ignore url request error",
			in:   &Option{IgnoreReqErr: false, URL: "127.0.0.1:999999"},
			fail: true,
		},
	}
	for idx, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewProm(tc.in)
			if err != nil {
				t.Errorf("[%d] failed to init prom: %s", idx, err)
			}
			_, err = p.Collect()
			if err != nil {
				if tc.fail {
					t.Logf("[%d] returned an error as expected: %s", idx, err)
				} else {
					t.Errorf("[%d] failed: %s", idx, err)
				}
				return
			}
			// Expect to fail but it didn't.
			if tc.fail {
				t.Errorf("[%d] expected to fail but it didn't", idx)
			}
			t.Logf("[%d] PASS", idx)
		})
	}
}

func TestProm(t *testing.T) {
	testCases := []struct {
		name     string
		in       *Option
		fail     bool
		expected []string
	}{
		{
			name: "counter metric type only",
			in: &Option{
				URL:         promURL,
				MetricTypes: []string{"counter"},
			},
			fail: false,
			expected: []string{
				"promhttp,cause=encoding metric_handler_errors_total=0",
				"promhttp,cause=gathering metric_handler_errors_total=0",
				"promhttp,code=200 metric_handler_requests_total=15143",
				"promhttp,code=500 metric_handler_requests_total=0",
				"promhttp,code=503 metric_handler_requests_total=0",
			},
		},

		{
			name: "histogram metric type only",
			in: &Option{
				URL:         promURL,
				MetricTypes: []string{"histogram"},
			},
			fail: false,
			expected: []string{
				"http,le=+Inf,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.003,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.03,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.1,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.3,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=1.5,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=10,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,method=GET,status_code=404 request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013",
			},
		},

		{
			name: "default metric types",
			in: &Option{
				URL:         promURL,
				MetricTypes: []string{},
			},
			fail: false,
			expected: []string{
				"go gc_duration_seconds_count=0,gc_duration_seconds_sum=0",
				"go,quantile=0 gc_duration_seconds=0",
				"go,quantile=0.25 gc_duration_seconds=0",
				"go,quantile=0.5 gc_duration_seconds=0",
				"http,le=+Inf,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.003,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.03,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.1,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.3,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=1.5,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=10,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,method=GET,status_code=404 request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013",
				"promhttp metric_handler_requests_in_flight=1",
				"promhttp,cause=encoding metric_handler_errors_total=0",
				"promhttp,cause=gathering metric_handler_errors_total=0",
				"promhttp,code=200 metric_handler_requests_total=15143",
				"promhttp,code=500 metric_handler_requests_total=0",
				"promhttp,code=503 metric_handler_requests_total=0",
				"up up=1",
			},
		},

		{
			name: "all metric types",
			in: &Option{
				URL:         promURL,
				MetricTypes: []string{"histogram", "gauge", "counter", "summary", "untyped"},
			},
			fail: false,
			expected: []string{
				"go gc_duration_seconds_count=0,gc_duration_seconds_sum=0",
				"go,quantile=0 gc_duration_seconds=0",
				"go,quantile=0.25 gc_duration_seconds=0",
				"go,quantile=0.5 gc_duration_seconds=0",
				"http,le=+Inf,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.003,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.03,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.1,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.3,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=1.5,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=10,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,method=GET,status_code=404 request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013",
				"promhttp metric_handler_requests_in_flight=1",
				"promhttp,cause=encoding metric_handler_errors_total=0",
				"promhttp,cause=gathering metric_handler_errors_total=0",
				"promhttp,code=200 metric_handler_requests_total=15143",
				"promhttp,code=500 metric_handler_requests_total=0",
				"promhttp,code=503 metric_handler_requests_total=0",
				"up up=1",
			},
		},

		{
			name: "metric name filtering",
			in: &Option{
				URL:              promURL,
				MetricNameFilter: []string{"http"},
			},
			fail: false,
			expected: []string{
				"http,le=+Inf,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.003,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.03,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.1,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.3,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=1.5,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=10,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,method=GET,status_code=404 request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013",
				"promhttp metric_handler_requests_in_flight=1",
				"promhttp,cause=encoding metric_handler_errors_total=0",
				"promhttp,cause=gathering metric_handler_errors_total=0",
				"promhttp,code=200 metric_handler_requests_total=15143",
				"promhttp,code=500 metric_handler_requests_total=0",
				"promhttp,code=503 metric_handler_requests_total=0",
			},
		},

		{
			name: "regex metric name filtering",
			in: &Option{
				URL:              promURL,
				MetricNameFilter: []string{"promht+p_metric_han[a-z]ler_req[^abcd]ests_total?"},
			},
			fail: false,
			expected: []string{
				"promhttp,code=200 metric_handler_requests_total=15143",
				"promhttp,code=500 metric_handler_requests_total=0",
				"promhttp,code=503 metric_handler_requests_total=0",
			},
		},

		{
			name: "measurement name prefix",
			in: &Option{
				URL:               promURL,
				MeasurementPrefix: "prefix_",
			},
			fail: false,
			expected: []string{
				"prefix_go gc_duration_seconds_count=0,gc_duration_seconds_sum=0",
				"prefix_go,quantile=0 gc_duration_seconds=0",
				"prefix_go,quantile=0.25 gc_duration_seconds=0",
				"prefix_go,quantile=0.5 gc_duration_seconds=0",
				"prefix_http,le=+Inf,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"prefix_http,le=0.003,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"prefix_http,le=0.03,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"prefix_http,le=0.1,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"prefix_http,le=0.3,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"prefix_http,le=1.5,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"prefix_http,le=10,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"prefix_http,method=GET,status_code=404 request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013",
				"prefix_promhttp metric_handler_requests_in_flight=1",
				"prefix_promhttp,cause=encoding metric_handler_errors_total=0",
				"prefix_promhttp,cause=gathering metric_handler_errors_total=0",
				"prefix_promhttp,code=200 metric_handler_requests_total=15143",
				"prefix_promhttp,code=500 metric_handler_requests_total=0",
				"prefix_promhttp,code=503 metric_handler_requests_total=0",
				"prefix_up up=1",
			},
		},

		{
			name: "measurement name",
			in: &Option{
				URL:             promURL,
				MeasurementName: "measurement_name",
			},
			fail: false,
			expected: []string{
				"measurement_name go_gc_duration_seconds_count=0,go_gc_duration_seconds_sum=0",
				"measurement_name promhttp_metric_handler_requests_in_flight=1",
				"measurement_name up=1",
				"measurement_name,cause=encoding promhttp_metric_handler_errors_total=0",
				"measurement_name,cause=gathering promhttp_metric_handler_errors_total=0",
				"measurement_name,code=200 promhttp_metric_handler_requests_total=15143",
				"measurement_name,code=500 promhttp_metric_handler_requests_total=0",
				"measurement_name,code=503 promhttp_metric_handler_requests_total=0",
				"measurement_name,le=+Inf,method=GET,status_code=404 http_request_duration_seconds_bucket=1i",
				"measurement_name,le=0.003,method=GET,status_code=404 http_request_duration_seconds_bucket=1i",
				"measurement_name,le=0.03,method=GET,status_code=404 http_request_duration_seconds_bucket=1i",
				"measurement_name,le=0.1,method=GET,status_code=404 http_request_duration_seconds_bucket=1i",
				"measurement_name,le=0.3,method=GET,status_code=404 http_request_duration_seconds_bucket=1i",
				"measurement_name,le=1.5,method=GET,status_code=404 http_request_duration_seconds_bucket=1i",
				"measurement_name,le=10,method=GET,status_code=404 http_request_duration_seconds_bucket=1i",
				"measurement_name,method=GET,status_code=404 http_request_duration_seconds_count=1,http_request_duration_seconds_sum=0.002451013",
				"measurement_name,quantile=0 go_gc_duration_seconds=0",
				"measurement_name,quantile=0.25 go_gc_duration_seconds=0",
				"measurement_name,quantile=0.5 go_gc_duration_seconds=0",
			},
		},

		{
			name: "tags filtering",
			in: &Option{
				URL:        promURL,
				TagsIgnore: []string{"status_code", "method"},
			},
			fail: false,
			expected: []string{
				"go gc_duration_seconds_count=0,gc_duration_seconds_sum=0",
				"go,quantile=0 gc_duration_seconds=0",
				"go,quantile=0.25 gc_duration_seconds=0",
				"go,quantile=0.5 gc_duration_seconds=0",
				"http request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013",
				"http,le=+Inf request_duration_seconds_bucket=1i",
				"http,le=0.003 request_duration_seconds_bucket=1i",
				"http,le=0.03 request_duration_seconds_bucket=1i",
				"http,le=0.1 request_duration_seconds_bucket=1i",
				"http,le=0.3 request_duration_seconds_bucket=1i",
				"http,le=1.5 request_duration_seconds_bucket=1i",
				"http,le=10 request_duration_seconds_bucket=1i",
				"promhttp metric_handler_requests_in_flight=1",
				"promhttp,cause=encoding metric_handler_errors_total=0",
				"promhttp,cause=gathering metric_handler_errors_total=0",
				"promhttp,code=200 metric_handler_requests_total=15143",
				"promhttp,code=500 metric_handler_requests_total=0",
				"promhttp,code=503 metric_handler_requests_total=0",
				"up up=1",
			},
		},

		{
			name: "tags filtering",
			in: &Option{
				URL: promURL,
				Measurements: []Rule{
					{
						Prefix: "go_",
						Name:   "with_prefix_go",
					},
					{
						Prefix: "request_",
						Name:   "with_prefix_request",
					},
				},
			},
			fail: false,
			expected: []string{
				"http,le=+Inf,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.003,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.03,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.1,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.3,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=1.5,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=10,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,method=GET,status_code=404 request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013",
				"promhttp metric_handler_requests_in_flight=1",
				"promhttp,cause=encoding metric_handler_errors_total=0",
				"promhttp,cause=gathering metric_handler_errors_total=0",
				"promhttp,code=200 metric_handler_requests_total=15143",
				"promhttp,code=500 metric_handler_requests_total=0",
				"promhttp,code=503 metric_handler_requests_total=0",
				"up up=1",
				"with_prefix_go gc_duration_seconds_count=0,gc_duration_seconds_sum=0",
				"with_prefix_go,quantile=0 gc_duration_seconds=0",
				"with_prefix_go,quantile=0.25 gc_duration_seconds=0",
				"with_prefix_go,quantile=0.5 gc_duration_seconds=0",
			},
		},

		{
			name: "custom tags",
			in: &Option{
				URL:  promURL,
				Tags: map[string]string{"some_tag": "some_value", "more_tag": "some_other_value"},
			},
			fail: false,
			expected: []string{
				"go,more_tag=some_other_value,quantile=0,some_tag=some_value gc_duration_seconds=0",
				"go,more_tag=some_other_value,quantile=0.25,some_tag=some_value gc_duration_seconds=0",
				"go,more_tag=some_other_value,quantile=0.5,some_tag=some_value gc_duration_seconds=0",
				"go,more_tag=some_other_value,some_tag=some_value gc_duration_seconds_count=0,gc_duration_seconds_sum=0",
				"http,le=+Inf,method=GET,more_tag=some_other_value,some_tag=some_value,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.003,method=GET,more_tag=some_other_value,some_tag=some_value,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.03,method=GET,more_tag=some_other_value,some_tag=some_value,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.1,method=GET,more_tag=some_other_value,some_tag=some_value,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.3,method=GET,more_tag=some_other_value,some_tag=some_value,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=1.5,method=GET,more_tag=some_other_value,some_tag=some_value,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=10,method=GET,more_tag=some_other_value,some_tag=some_value,status_code=404 request_duration_seconds_bucket=1i",
				"http,method=GET,more_tag=some_other_value,some_tag=some_value,status_code=404 request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013",
				"promhttp,cause=encoding,more_tag=some_other_value,some_tag=some_value metric_handler_errors_total=0",
				"promhttp,cause=gathering,more_tag=some_other_value,some_tag=some_value metric_handler_errors_total=0",
				"promhttp,code=200,more_tag=some_other_value,some_tag=some_value metric_handler_requests_total=15143",
				"promhttp,code=500,more_tag=some_other_value,some_tag=some_value metric_handler_requests_total=0",
				"promhttp,code=503,more_tag=some_other_value,some_tag=some_value metric_handler_requests_total=0",
				"promhttp,more_tag=some_other_value,some_tag=some_value metric_handler_requests_in_flight=1",
				"up,more_tag=some_other_value,some_tag=some_value up=1",
			},
		},

		{
			name: "multiple urls",
			in: &Option{
				URLs:         []string{"localhost:1234", "localhost:5678"},
				IgnoreReqErr: true,
			},
			fail: false,
			expected: []string{
				"go gc_duration_seconds_count=0,gc_duration_seconds_sum=0",
				"go gc_duration_seconds_count=0,gc_duration_seconds_sum=0",
				"go,quantile=0 gc_duration_seconds=0",
				"go,quantile=0 gc_duration_seconds=0",
				"go,quantile=0.25 gc_duration_seconds=0",
				"go,quantile=0.25 gc_duration_seconds=0",
				"go,quantile=0.5 gc_duration_seconds=0",
				"go,quantile=0.5 gc_duration_seconds=0",
				"http,le=+Inf,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=+Inf,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.003,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.003,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.03,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.03,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.1,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.1,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.3,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.3,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=1.5,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=1.5,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=10,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=10,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,method=GET,status_code=404 request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013",
				"http,method=GET,status_code=404 request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013",
				"promhttp metric_handler_requests_in_flight=1",
				"promhttp metric_handler_requests_in_flight=1",
				"promhttp,cause=encoding metric_handler_errors_total=0",
				"promhttp,cause=encoding metric_handler_errors_total=0",
				"promhttp,cause=gathering metric_handler_errors_total=0",
				"promhttp,cause=gathering metric_handler_errors_total=0",
				"promhttp,code=200 metric_handler_requests_total=15143",
				"promhttp,code=200 metric_handler_requests_total=15143",
				"promhttp,code=500 metric_handler_requests_total=0",
				"promhttp,code=500 metric_handler_requests_total=0",
				"promhttp,code=503 metric_handler_requests_total=0",
				"promhttp,code=503 metric_handler_requests_total=0",
				"up up=1",
				"up up=1",
			},
		},
	}
	for idx, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewProm(tc.in)
			if err != nil {
				t.Errorf("[%d] failed to init prom: %s", idx, err)
			}
			p.SetClient(&http.Client{Transport: newTransportMock(mockBody)})
			pts, err := p.Collect()
			if err != nil {
				if tc.fail {
					t.Logf("[%d] returned an error as expected: %s", idx, err)
				} else {
					t.Errorf("[%d] failed: %s", idx, err)
				}
				return
			}
			// Expect to fail but it didn't.
			if tc.fail {
				t.Errorf("[%d] expected to fail but it didn't", idx)
			}
			var got []string
			for _, p := range pts {
				s := p.String()
				// remove timestamp
				s = s[:strings.LastIndex(s, " ")]
				got = append(got, s)
			}
			sort.Strings(got)
			tu.Equals(t, tc.expected, got)
			t.Logf("[%d] PASS", idx)
		})
	}
}
