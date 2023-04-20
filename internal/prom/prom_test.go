// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package prom

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	tu "github.com/GuanceCloud/cliutils/testutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

const promURL = "http://127.0.0.1:9100/metrics"

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

func (t *transportMock) CancelRequest(_ *http.Request) {}

func newTransportMock(body string) http.RoundTripper {
	return &transportMock{statusCode: http.StatusOK, body: body}
}

func TestCollect(t *testing.T) {
	testcases := []struct {
		in     *Option
		name   string
		fail   bool
		expect []string
	}{
		{
			name: "nil option",
			fail: true,
		},
		{
			name: "empty option",
			in:   &Option{},
			fail: true,
		},
		{
			name: "ok",
			expect: []string{
				`gogo gc_duration_seconds_count=0,gc_duration_seconds_sum=0`,
				`gogo,quantile=0 gc_duration_seconds=0`,
				`gogo,quantile=0.25 gc_duration_seconds=0`,
				`gogo,quantile=0.5 gc_duration_seconds=0`,
			},
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
		},

		{
			name: "option-only-URL",
			in:   &Option{URL: promURL},
			expect: []string{
				`go gc_duration_seconds_count=0,gc_duration_seconds_sum=0`,
				`go,quantile=0 gc_duration_seconds=0`,
				`go,quantile=0.25 gc_duration_seconds=0`,
				`go,quantile=0.5 gc_duration_seconds=0`,
				"http,le=1.2,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=+Inf,method=GET,status_code=403 request_duration_seconds_bucket=1i",
				`http,le=0.003,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,le=0.03,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,le=0.1,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,le=0.3,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,le=1.5,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,le=10,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,method=GET,status_code=404 request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013`,
				"http,method=GET,status_code=403 request_duration_seconds_count=0,request_duration_seconds_sum=0",
				`promhttp metric_handler_requests_in_flight=1`,
				`promhttp,cause=encoding metric_handler_errors_total=0`,
				`promhttp,cause=gathering metric_handler_errors_total=0`,
				`promhttp,code=200 metric_handler_requests_total=15143`,
				`promhttp,code=500 metric_handler_requests_total=0`,
				`promhttp,code=503 metric_handler_requests_total=0`,
				`up up=1`,
			},
		},

		{
			name: "option-ignore-tag-kv",
			in: &Option{
				URL: promURL,
				IgnoreTagKV: IgnoreTagKeyValMatch{
					"le":          []*regexp.Regexp{regexp.MustCompile("0.*")},
					"status_code": []*regexp.Regexp{regexp.MustCompile("403")},
				},
			},
			expect: []string{
				`go gc_duration_seconds_count=0,gc_duration_seconds_sum=0`,
				`go,quantile=0 gc_duration_seconds=0`,
				`go,quantile=0.25 gc_duration_seconds=0`,
				`go,quantile=0.5 gc_duration_seconds=0`,
				`http,le=1.2,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,le=1.5,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,method=GET,status_code=404 request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013`,
				`promhttp metric_handler_requests_in_flight=1`,
				`promhttp,cause=encoding metric_handler_errors_total=0`,
				`promhttp,cause=gathering metric_handler_errors_total=0`,
				`promhttp,code=200 metric_handler_requests_total=15143`,
				`promhttp,code=500 metric_handler_requests_total=0`,
				`promhttp,code=503 metric_handler_requests_total=0`,
				`up up=1`,
			},
		},
	}

	mockBody := `
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
http_request_duration_seconds_bucket{le="1.2",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="+Inf",status_code="403",method="GET"} 1
http_request_duration_seconds_sum{status_code="404",method="GET"} 0.002451013
http_request_duration_seconds_count{status_code="404",method="GET"} 1
# HELP up 1 = up, 0 = not up
# TYPE up untyped
up 1
`

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewProm(tc.in)
			if tc.fail && assert.Error(t, err) {
				return
			} else {
				assert.NoError(t, err)
			}

			p.SetClient(&http.Client{Transport: newTransportMock(mockBody)})
			p.opt.DisableInstanceTag = true

			pts, err := p.CollectFromHTTP(p.opt.URL)
			if tc.fail && assert.Error(t, err) {
				return
			} else {
				assert.NoError(t, err)
			}

			var arr []string
			for _, pt := range pts {
				arr = append(arr, pt.String())
			}

			sort.Strings(arr)
			sort.Strings(tc.expect)

			for i := range arr {
				assert.Equal(t, strings.HasPrefix(arr[i], tc.expect[i]), true)
				t.Logf(">>>\n%s\n%s", arr[i], tc.expect[i])
			}
		})
	}
}

func Test_BearerToken(t *testing.T) {
	tmpDir, err := ioutil.TempDir("./", "__tmp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir) // nolint:errcheck
	f, err := ioutil.TempFile(tmpDir, "token")
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
		tmpDir, err := ioutil.TempDir("./", "__tmp")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir) // nolint:errcheck
		f, err := ioutil.TempFile(tmpDir, "ca.crt")
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
		Disable: true,
	}
	assert.True(t, o.IsDisable(), o.Disable)

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
	mockBody := `
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

	tmpDir, err := ioutil.TempDir("./", "__tmp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir) // nolint:errcheck
	f, err := ioutil.TempFile(tmpDir, "output")
	assert.NoError(t, err)
	outputFile, err := filepath.Abs(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	p, err := NewProm(&Option{
		URL:         promURL,
		Output:      outputFile,
		MaxFileSize: 100000,
	})

	assert.NoError(t, err)
	p.SetClient(&http.Client{Transport: newTransportMock(mockBody)})
	err = p.WriteMetricText2File(p.opt.URL)

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
			_, err = p.CollectFromHTTP(p.opt.URL)
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
			name: "rename-measurement",
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

	mockBody := `
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

	for idx, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewProm(tc.in)
			if err != nil {
				t.Errorf("[%d] failed to init prom: %s", idx, err)
			}
			p.SetClient(&http.Client{Transport: newTransportMock(mockBody)})
			p.opt.DisableInstanceTag = true
			var points []*point.Point
			for _, u := range p.opt.URLs {
				pts, err := p.CollectFromHTTP(u)
				if err != nil {
					break
				}
				points = append(points, pts...)
			}
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
			for _, p := range points {
				s := p.String()
				// remove timestamp
				s = s[:strings.LastIndex(s, " ")]
				got = append(got, s)
			}
			sort.Strings(got)
			tu.Equals(t, strings.Join(tc.expected, "\n"), strings.Join(got, "\n"))
			t.Logf("[%d] PASS", idx)
		})
	}
}

func TestCollectFromFile(t *testing.T) {
	mockBody := `
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

	f, err := os.CreateTemp("./", "test_collect_from_file_")
	if err != nil {
		t.Errorf(err.Error())
	}
	defer os.Remove(f.Name()) //nolint:errcheck,gosec

	if _, err := f.WriteString(mockBody); err != nil {
		t.Errorf("fail to write mock body to temporary file: %v", err)
	}
	if err := f.Sync(); err != nil {
		t.Errorf("fail to flush data to disk: %v", err)
	}
	option := Option{
		URLs: []string{f.Name()},
	}
	p, err := NewProm(&option)
	if err != nil {
		t.Errorf("failed to init prom: %s", err)
	}
	if _, err := p.CollectFromFile(f.Name()); err != nil {
		t.Errorf(err.Error())
	}
}

func TestGetTimestampS(t *testing.T) {
	var (
		ts        int64 = 1647959040488
		startTime       = time.Unix(1600000000, 0)
	)
	m1 := dto.Metric{
		TimestampMs: &ts,
	}
	m2 := dto.Metric{}
	tu.Equals(t, int64(1647959040000000000), getTimestampS(&m1, startTime).UnixNano())
	tu.Equals(t, int64(1600000000000000000), getTimestampS(&m2, startTime).UnixNano())
}

func TestRenameTag(t *testing.T) {
	tm := time.Now()
	cases := []struct {
		name     string
		opt      *Option
		promdata string
		expect   []*point.Point
	}{
		{
			name: "rename-tags",
			opt: &Option{
				URL: "http://not-set",
				RenameTags: &RenameTags{
					OverwriteExistTags: false,
					Mapping: map[string]string{
						"status_code":    "StatusCode",
						"tag-not-exists": "do-nothing",
					},
				},
			},
			expect: []*point.Point{
				func() *point.Point {
					pt, err := point.NewPoint("http",
						map[string]string{"le": "0.003", "StatusCode": "404", "method": "GET"},
						map[string]interface{}{"request_duration_seconds_bucket": 1.0},
						&point.PointOption{Category: datakit.Metric})
					if err != nil {
						t.Errorf("NewPoint: %s", err)
						return nil
					}
					return pt
				}(),
			},
			promdata: `
http_request_duration_seconds_bucket{le="0.003",status_code="404",method="GET"} 1
			`,
		},

		{
			name: "rename-overwrite-tags",
			opt: &Option{
				URL: "http://not-set",
				RenameTags: &RenameTags{
					OverwriteExistTags: true, // enable overwrite
					Mapping: map[string]string{
						"status_code": "StatusCode",
						"method":      "tag_exists", // rename `method` to a exists tag key
					},
				},
			},
			expect: []*point.Point{
				func() *point.Point {
					pt, err := point.NewPoint("http",
						// method key removed, it's value overwrite tag_exists's value
						map[string]string{"le": "0.003", "StatusCode": "404", "tag_exists": "GET"},
						map[string]interface{}{"request_duration_seconds_bucket": 1.0},
						&point.PointOption{Category: datakit.Metric})
					if err != nil {
						t.Errorf("NewPoint: %s", err)
						return nil
					}
					return pt
				}(),
			},
			promdata: `
http_request_duration_seconds_bucket{le="0.003",tag_exists="yes",status_code="404",method="GET"} 1
			`,
		},

		{
			name: "rename-tags-disable-overwrite",
			opt: &Option{
				URL: "http://not-set",
				RenameTags: &RenameTags{
					OverwriteExistTags: false, // enable overwrite
					Mapping: map[string]string{
						"status_code": "StatusCode",
						"method":      "tag_exists", // rename `method` to a exists tag key
					},
				},
			},
			expect: []*point.Point{
				func() *point.Point {
					pt, err := point.NewPoint("http",
						map[string]string{"le": "0.003", "tag_exists": "yes", "StatusCode": "404", "method": "GET"}, // overwrite not work on method
						map[string]interface{}{"request_duration_seconds_bucket": 1.0},
						&point.PointOption{Category: datakit.Metric})
					if err != nil {
						t.Errorf("NewPoint: %s", err)
						return nil
					}
					return pt
				}(),
			},
			promdata: `
http_request_duration_seconds_bucket{le="0.003",status_code="404",tag_exists="yes", method="GET"} 1
			`,
		},

		{
			name: "empty-tags",
			opt: &Option{
				URL: "http://not-set",
				RenameTags: &RenameTags{
					OverwriteExistTags: true, // enable overwrite
					Mapping: map[string]string{
						"status_code": "StatusCode",
						"method":      "tag_exists", // rename `method` to a exists tag key
					},
				},
			},
			expect: []*point.Point{
				func() *point.Point {
					pt, err := point.NewPoint("http",
						nil,
						map[string]interface{}{"request_duration_seconds_bucket": 1.0},
						&point.PointOption{Category: datakit.Metric, Time: tm})
					if err != nil {
						t.Errorf("NewPoint: %s", err)
						return nil
					}
					return pt
				}(),
			},
			promdata: `
http_request_duration_seconds_bucket 1
			`,
		},
	}

	for _, tc := range cases {
		p, err := NewProm(tc.opt)
		if err != nil {
			t.Error(err)
			return
		}

		t.Run(tc.name, func(t *testing.T) {
			pts, err := p.text2Metrics(bytes.NewBufferString(tc.promdata), "")
			if err != nil {
				t.Error(err)
				return
			}

			for idx, pt := range pts {
				tu.Equals(t, tc.expect[idx].PrecisionString("m"), pt.PrecisionString("m"))
				t.Log(tc.expect[idx].PrecisionString("m"))
			}
		})
	}
}

func TestSetHeaders(t *testing.T) {
	testcases := []struct {
		name string
		opt  *Option
	}{
		{
			name: "add custom http header",
			opt: &Option{
				URL: "dummy_url",
				HTTPHeaders: map[string]string{
					"Root":    "passwd",
					"Michael": "12345",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewProm(tc.opt)
			if err != nil {
				t.Error(err)
				return
			}
			req, err := p.GetReq(p.opt.URL)
			assert.NoError(t, err)
			for k, v := range p.opt.HTTPHeaders {
				assert.Equal(t, v, req.Header.Get(k))
			}
		})
	}
}

func TestRepetitiveCommentsInText(t *testing.T) {
	testcases := []struct {
		in       *Option
		name     string
		repeated bool
		expected []string
		text     string
		fail     bool
	}{
		{
			name:     "non-repeated-comment",
			in:       &Option{URL: promURL},
			repeated: false,
			expected: []string{
				`go gc_duration_seconds_count=0,gc_duration_seconds_sum=0`,
				`go,quantile=0 gc_duration_seconds=0`,
				`go,quantile=0.25 gc_duration_seconds=0`,
				`go,quantile=0.5 gc_duration_seconds=0`,
				"http,le=1.2,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=+Inf,method=GET,status_code=403 request_duration_seconds_bucket=1i",
				`http,le=0.003,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,le=0.03,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,le=0.1,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,le=0.3,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,le=1.5,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,le=10,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,method=GET,status_code=404 request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013`,
				"http,method=GET,status_code=403 request_duration_seconds_count=0,request_duration_seconds_sum=0",
				`promhttp metric_handler_requests_in_flight=1`,
				`promhttp,cause=encoding metric_handler_errors_total=0`,
				`promhttp,cause=gathering metric_handler_errors_total=0`,
				`promhttp,code=200 metric_handler_requests_total=15143`,
				`promhttp,code=500 metric_handler_requests_total=0`,
				`promhttp,code=503 metric_handler_requests_total=0`,
				`up up=1`,
			},
			text: `
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
http_request_duration_seconds_bucket{le="1.2",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="+Inf",status_code="403",method="GET"} 1
http_request_duration_seconds_sum{status_code="404",method="GET"} 0.002451013
http_request_duration_seconds_count{status_code="404",method="GET"} 1
# HELP up 1 = up, 0 = not up
# TYPE up untyped
up 1
`,
		},
		{
			name:     "missing-some-comment",
			in:       &Option{URL: promURL},
			repeated: false,
			text: `
promhttp_metric_handler_errors_total{cause="encoding"} 0
promhttp_metric_handler_errors_total{cause="gathering"} 0
promhttp_metric_handler_requests_in_flight 1
promhttp_metric_handler_requests_total{code="200"} 15143
go_gc_duration_seconds{quantile="0"} 0
go_gc_duration_seconds{quantile="0.25"} 0
go_gc_duration_seconds{quantile="0.5"} 0
# HELP http_request_duration_seconds duration histogram of http responses labeled with: status_code, method
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_sum{status_code="404",method="GET"} 0.002451013
http_request_duration_seconds_count{status_code="404",method="GET"} 1
# HELP up 1 = up, 0 = not up
# TYPE up untyped
up 1
`,
		},
		{
			name:     "ends-with-a-comment",
			in:       &Option{URL: promURL},
			repeated: false,
			text: `
promhttp_metric_handler_errors_total{cause="encoding"} 0
promhttp_metric_handler_errors_total{cause="gathering"} 0
promhttp_metric_handler_requests_in_flight 1
promhttp_metric_handler_requests_total{code="200"} 15143
go_gc_duration_seconds{quantile="0"} 0
go_gc_duration_seconds{quantile="0.25"} 0
go_gc_duration_seconds{quantile="0.5"} 0
# HELP http_request_duration_seconds duration histogram of http responses labeled with: status_code, method
# TYPE http_request_duration_seconds histogram
`,
		},
		{
			name:     "repeated-comment",
			in:       &Option{URL: promURL},
			repeated: true,
			expected: []string{
				"node,address=00:00:00:00:00:00,broadcast=00:00:00:00:00:00,device=lo,operstate=unknown network_info=1",
				"node,address=a1:b2:c3:d4:e5:f6,broadcast=ff:ff:ff:ff:ff:ff,device=br-lan,duplex=unknown,operstate=up network_info=1",
				"node,address=a1:b2:c3:d4:e5:f6,broadcast=ff:ff:ff:ff:ff:ff,device=eth0,duplex=full,operstate=up network_info=1",
				"node,address=a1:b2:c3:d4:e5:f6,broadcast=ff:ff:ff:ff:ff:ff,device=lan1,duplex=unknown,operstate=lowerlayerdown network_info=1",
				"node,address=a1:b2:c3:d4:e5:f6,broadcast=ff:ff:ff:ff:ff:ff,device=lan2,duplex=full,operstate=up network_info=1",
				"node,address=a1:b2:c3:d4:e5:f6,broadcast=ff:ff:ff:ff:ff:ff,device=lan3,duplex=full,operstate=up network_info=1",
				"node,address=a1:b2:c3:d4:e5:f6,broadcast=ff:ff:ff:ff:ff:ff,device=wan,duplex=full,operstate=up network_info=1",
				"node,address=a1:b2:c3:d4:e5:f6,broadcast=ff:ff:ff:ff:ff:ff,device=wlan0,operstate=up network_info=1",
				"node,address=a1:b2:c3:d4:e5:f6,broadcast=ff:ff:ff:ff:ff:ff,device=wlan1,operstate=up network_info=1",
			},
			text: `# TYPE node_network_info gauge
node_network_info{duplex="unknown",broadcast="ff:ff:ff:ff:ff:ff",ifalias="",address="a1:b2:c3:d4:e5:f6",device="lan1",operstate="lowerlayerdown"} 1
# TYPE node_network_info gauge
node_network_info{duplex="unknown",broadcast="ff:ff:ff:ff:ff:ff",ifalias="",address="a1:b2:c3:d4:e5:f6",device="br-lan",operstate="up"} 1
# TYPE node_network_info gauge
node_network_info{operstate="up",broadcast="ff:ff:ff:ff:ff:ff",ifalias="",address="a1:b2:c3:d4:e5:f6",device="wlan1"} 1
# TYPE node_network_info gauge
node_network_info{operstate="unknown",broadcast="00:00:00:00:00:00",ifalias="",address="00:00:00:00:00:00",device="lo"} 1
# TYPE node_network_info gauge
node_network_info{operstate="up",broadcast="ff:ff:ff:ff:ff:ff",ifalias="",address="a1:b2:c3:d4:e5:f6",device="wlan0"} 1
# TYPE node_network_info gauge
node_network_info{duplex="full",broadcast="ff:ff:ff:ff:ff:ff",ifalias="",address="a1:b2:c3:d4:e5:f6",device="wan",operstate="up"} 1
# TYPE node_network_info gauge
node_network_info{duplex="full",broadcast="ff:ff:ff:ff:ff:ff",ifalias="",address="a1:b2:c3:d4:e5:f6",device="lan3",operstate="up"} 1
# TYPE node_network_info gauge
node_network_info{duplex="full",broadcast="ff:ff:ff:ff:ff:ff",ifalias="",address="a1:b2:c3:d4:e5:f6",device="eth0",operstate="up"} 1
# TYPE node_network_info gauge
node_network_info{duplex="full",broadcast="ff:ff:ff:ff:ff:ff",ifalias="",address="a1:b2:c3:d4:e5:f6",device="lan2",operstate="up"} 1
`,
		},
		{
			name:     "#TYPE",
			in:       &Option{URL: promURL},
			repeated: true,
			expected: []string{
				"node,address=a1:b2:c3:d4:e5:f6,broadcast=ff:ff:ff:ff:ff:ff,device=lan1,duplex=unknown,operstate=lowerlayerdown network_info=1",
				"node,address=a1:b2:c3:d4:e5:f6,broadcast=ff:ff:ff:ff:ff:ff,device=lan1,duplex=unknown,operstate=lowerlayerdown network_info=1",
			},
			text: `#TYPE node_network_info gauge
node_network_info{duplex="unknown",broadcast="ff:ff:ff:ff:ff:ff",ifalias="",address="a1:b2:c3:d4:e5:f6",device="lan1",operstate="lowerlayerdown"} 1
#TYPE node_network_info gauge
node_network_info{duplex="unknown",broadcast="ff:ff:ff:ff:ff:ff",ifalias="",address="a1:b2:c3:d4:e5:f6",device="lan1",operstate="lowerlayerdown"} 1
`,
		},
		{
			name:     "not-enough-token-after-#TYPE",
			in:       &Option{URL: promURL},
			repeated: false,
			expected: []string{
				"node,address=a1:b2:c3:d4:e5:f6,broadcast=ff:ff:ff:ff:ff:ff,device=lan1,duplex=unknown,operstate=lowerlayerdown network_info=1",
				"node,address=a1:b2:c3:d4:e5:f6,broadcast=ff:ff:ff:ff:ff:ff,device=lan1,duplex=unknown,operstate=lowerlayerdown network_info=1",
			},
			text: `#TYPE node_network_info gauge
node_network_info{duplex="unknown",broadcast="ff:ff:ff:ff:ff:ff",ifalias="",address="a1:b2:c3:d4:e5:f6",device="lan1",operstate="lowerlayerdown"} 1
#TYPE
node_network_info{duplex="unknown",broadcast="ff:ff:ff:ff:ff:ff",ifalias="",address="a1:b2:c3:d4:e5:f6",device="lan1",operstate="lowerlayerdown"} 1
`,
		},
		{
			name:     "common-comments",
			in:       &Option{URL: promURL},
			repeated: false,
			expected: []string{
				"node,address=a1:b2:c3:d4:e5:f6,broadcast=ff:ff:ff:ff:ff:ff,device=lan1,duplex=unknown,operstate=lowerlayerdown network_info=1",
				"node,address=a1:b2:c3:d4:e5:f6,broadcast=ff:ff:ff:ff:ff:ff,device=lan1,duplex=unknown,operstate=lowerlayerdown network_info=1",
			},
			text: `# This is a common comment
node_network_info{duplex="unknown",broadcast="ff:ff:ff:ff:ff:ff",ifalias="",address="a1:b2:c3:d4:e5:f6",device="lan1",operstate="lowerlayerdown"} 1
#This is another comment, which should be ignored
node_network_info{duplex="unknown",broadcast="ff:ff:ff:ff:ff:ff",ifalias="",address="a1:b2:c3:d4:e5:f6",device="lan1",operstate="lowerlayerdown"} 1
`,
		},
		{
			name:     "#HELP-not-at-start-of-line",
			in:       &Option{URL: promURL},
			repeated: false,
			expected: []string{
				"node,address=a1:b2:c3:d4:e5:f6,broadcast=ff:ff:ff:ff:ff:ff,device=lan1,duplex=unknown,operstate=lowerlayerdown network_info=1",
				"node,address=a1:b2:c3:d4:e5:f6,broadcast=ff:ff:ff:ff:ff:ff,device=lan1,duplex=unknown,operstate=lowerlayerdown network_info=1",
			},
			text: `# This is a comment. #TYPE node_network_info gauge
#TYPE node_network_info gauge
node_network_info{duplex="unknown",broadcast="ff:ff:ff:ff:ff:ff",ifalias="",address="a1:b2:c3:d4:e5:f6",device="lan1",operstate="lowerlayerdown"} 1
node_network_info{duplex="unknown",broadcast="ff:ff:ff:ff:ff:ff",ifalias="",address="a1:b2:c3:d4:e5:f6",device="lan1",operstate="lowerlayerdown"} 1
`,
		},
		{
			name: "incomplete-comment-handled-by-prom-library",
			in:   &Option{URL: promURL},
			fail: true,
			text: `# TYPE node_network_info gauge
node_network_info{duplex="unknown",broadcast="ff:ff:ff:ff:ff:ff",ifalias="",address="a1:b2:c3:d4:e5:f6",device="lan1",operstate="lowerlayerdown"} 1
# TYPE`,
		},
		{
			name: "random text as input",
			in:   &Option{URL: promURL},
			fail: true,
			text: `hello, world!`,
		},
		{
			name: "html as input",
			in:   &Option{URL: promURL},
			fail: true,
			text: `<html>
<meta http-equiv="refresh" content="0;url=http://www.baidu.com/">
</html>`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var arr []string
			p, err := NewProm(tc.in)
			assert.NoError(t, err)
			p.SetClient(&http.Client{Transport: newTransportMock(tc.text)})
			if tc.fail {
				_, err := p.text2Metrics(bytes.NewReader([]byte(tc.text)), "")
				assert.Error(t, err)
				return
			}
			if !tc.repeated {
				// If there is no repetitive comment like # TYPE or # HELP,
				// doText2Metrics should produce exactly the same metrics as text2Metrics.
				var arr1, arr2 []string

				pts1, err := p.doText2Metrics(bytes.NewReader([]byte(tc.text)), "")
				assert.NoError(t, err)
				pts2, err := p.text2Metrics(bytes.NewReader([]byte(tc.text)), "")
				assert.NoError(t, err)

				assert.Equal(t, len(pts1), len(pts2))

				for _, pt := range pts1 {
					arr1 = append(arr1, trimTimeStamp(pt.String()))
				}
				for _, pt := range pts2 {
					arr2 = append(arr2, trimTimeStamp(pt.String()))
				}

				sort.Strings(arr1)
				sort.Strings(arr2)
				for i := range arr1 {
					assert.Equal(t, arr1[i], arr2[i])
				}
				if tc.expected == nil {
					// Solely do comparison.
					return
				}
				arr = arr1
			} else {
				// Repetitive comments should make doText2Metrics fail.
				_, err := p.doText2Metrics(bytes.NewReader([]byte(tc.text)), "")
				assert.Error(t, err)

				pts, err := p.text2Metrics(bytes.NewReader([]byte(tc.text)), "")
				assert.NoError(t, err)

				for _, pt := range pts {
					arr = append(arr, trimTimeStamp(pt.String()))
				}
				sort.Strings(arr)
			}

			sort.Strings(tc.expected)
			assert.Equal(t, len(arr), len(tc.expected))
			for i := range arr {
				assert.Equal(t, arr[i], tc.expected[i])
				t.Logf(">>>\n%s\n%s", arr[i], tc.expected[i])
			}
		})
	}
}

func trimTimeStamp(metricString string) string {
	return metricString[:strings.LastIndex(metricString, " ")]
}

func getProm() (*Prom, error) {
	return NewProm(&Option{URL: "dummyURL"})
}

func Benchmark_text2Metrics(b *testing.B) {
	tests := []struct {
		name     string
		filepath string
		flag     bool
		numPts   int
	}{
		{
			name:     "text2Metrics-istio.text",
			filepath: "./_test/istio.txt",
			flag:     true,
			numPts:   792,
		},
		{
			name:     "doText2Metrics-istio.text",
			filepath: "./_test/istio.txt",
			flag:     false,
			numPts:   792,
		},
		{
			name:     "text2Metrics-istiod.txt",
			filepath: "./_test/istiod.txt",
			flag:     true,
			numPts:   225,
		},
		{
			name:     "doText2Metrics-istiod.txt",
			filepath: "./_test/istiod.txt",
			flag:     false,
			numPts:   225,
		},
	}
	for _, t := range tests {
		b.Run(t.name, func(b *testing.B) {
			p, err := getProm()
			assert.NoError(b, err)

			if t.flag {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					f, err := os.Open(t.filepath)
					assert.NoError(b, err)
					pts, err := p.text2Metrics(f, "")
					assert.NoError(b, err)
					assert.Equal(b, t.numPts, len(pts))
					err = f.Close()
					assert.NoError(b, err)
				}
				b.StopTimer()
			} else {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					f, err := os.Open(t.filepath)
					assert.NoError(b, err)
					pts, err := p.doText2Metrics(f, "")
					assert.NoError(b, err)
					assert.Equal(b, t.numPts, len(pts))
					err = f.Close()
					assert.NoError(b, err)
				}
				b.StopTimer()
			}
		})
	}
}

func TestTokenIterator(t *testing.T) {
	testcases := []struct {
		name           string
		text           string
		start          int
		expectedTokens []string
	}{
		{
			name:           "normal-token-iterating-1",
			text:           "# TYPE http_request_duration_seconds histogram\n",
			start:          1,
			expectedTokens: []string{"TYPE", "http_request_duration_seconds", "histogram"},
		},
		{
			name:           "normal-token-iterating-1",
			text:           "# TYPE http_request_duration_seconds histogram\n",
			start:          2,
			expectedTokens: []string{"TYPE", "http_request_duration_seconds", "histogram"},
		},
		{
			name:           "no-more-token",
			text:           "# \n",
			start:          1,
			expectedTokens: []string{},
		},
		{
			name:           "follow-by-extra-space",
			text:           "#      TYPE       http_request_duration_seconds     histogram\n",
			start:          2,
			expectedTokens: []string{"TYPE", "http_request_duration_seconds", "histogram"},
		},
		{
			name:           "start-at-the-end",
			text:           "hello\n",
			start:          5,
			expectedTokens: []string{},
		},
		{
			name:           "starting-index-out-of-bound",
			text:           "# \n",
			start:          100,
			expectedTokens: []string{},
		},
		{
			name:           "not-ends-with-newline",
			text:           "# TYPE http_request_duration_seconds histogram",
			start:          1,
			expectedTokens: []string{"TYPE", "http_request_duration_seconds", "histogram"},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ti := newTokenIterator([]byte(tc.text), tc.start)
			var actual []string
			tkn := ti.readNextToken()
			for tkn != "" {
				actual = append(actual, tkn)
				tkn = ti.readNextToken()
			}
			assert.Equal(t, len(tc.expectedTokens), len(actual))
			for i := 0; i < len(actual); i++ {
				assert.Equal(t, tc.expectedTokens[i], actual[i])
			}
		})
	}
}

func TestInfoTag(t *testing.T) {
	testcases := []struct {
		in       *Option
		name     string
		fail     bool
		expected []string
	}{
		{
			name: "type info",
			in:   &Option{URL: promURL},
			expected: []string{"process,otel_scope_name=io.opentelemetry.runtime-metrics,otel_scope_version=1.24.0-alpha-SNAPSHOT,pool=direct runtime_jvm_buffer_count=6",
				"process,otel_scope_name=io.opentelemetry.runtime-metrics,otel_scope_version=1.24.0-alpha-SNAPSHOT,pool=mapped runtime_jvm_buffer_count=0",
				"process,otel_scope_name=io.opentelemetry.runtime-metrics,otel_scope_version=1.24.0-alpha-SNAPSHOT,pool=mapped\\ -\\ 'non-volatile\\ memory' runtime_jvm_buffer_count=0"},
		},
	}
	mockBody := `# TYPE otel_scope_info info
# HELP otel_scope_info Scope metadata
otel_scope_info{otel_scope_name="io.opentelemetry.runtime-metrics",otel_scope_version="1.24.0-alpha-SNAPSHOT"} 1
# TYPE process_runtime_jvm_buffer_count gauge
# HELP process_runtime_jvm_buffer_count The number of buffers in the pool
process_runtime_jvm_buffer_count{pool="mapped - 'non-volatile memory'"} 0.0 1680231835149
process_runtime_jvm_buffer_count{pool="direct"} 6.0 1680231835149
process_runtime_jvm_buffer_count{pool="mapped"} 0.0 1680231835149
`

	for idx, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewProm(tc.in)
			if tc.fail && assert.Error(t, err) {
				return
			} else {
				assert.NoError(t, err)
			}

			p.SetClient(&http.Client{Transport: newTransportMock(mockBody)})
			p.opt.DisableInstanceTag = true

			points, err := p.CollectFromHTTP(p.opt.URL)
			if tc.fail && assert.Error(t, err) {
				return
			} else {
				assert.NoError(t, err)
			}

			var got []string
			for _, p := range points {
				s := p.String()
				// remove timestamp
				got = append(got, s[:strings.LastIndex(s, " ")])
			}
			sort.Strings(got)
			tu.Equals(t, strings.Join(tc.expected, "\n"), strings.Join(got, "\n"))
			t.Logf("[%d] PASS", idx)
		})
	}
}
