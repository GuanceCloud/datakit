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
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	tu "github.com/GuanceCloud/cliutils/testutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
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

type optionMock struct {
	source   string
	interval string
	timeout  datakit.Duration

	// URL               string
	// urls              []string
	ignoreReqErr           bool
	metricTypes            []string
	metricNameFilter       []string
	metricNameFilterIgnore []string
	measurementPrefix      string
	measurementName        string
	measurements           []Rule
	output                 string
	maxFileSize            int64

	tlsOpen    bool
	udsPath    string
	cacertFile string
	certFile   string
	keyFile    string

	tagsIgnore  []string // do not keep these tags in scraped prom data
	tagsRename  *RenameTags
	asLogging   *AsLogging
	ignoreTagKV map[string][]string // drop scraped prom data if tag key's value matched
	httpHeaders map[string]string

	tags               map[string]string
	disableHostTag     bool
	disableInstanceTag bool
	disableInfoTag     bool

	auth map[string]string
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

func createOpts(in *optionMock) []PromOption {
	opts := make([]PromOption, 0)
	opts = append(
		opts,
		// WithLogger(in.l),
		WithSource(in.source),
		WithInterval(in.interval),
		WithTimeout(in.timeout),
		WithIgnoreReqErr(in.ignoreReqErr),
		WithMetricTypes(in.metricTypes),
		WithMetricNameFilter(in.metricNameFilter),
		WithMetricNameFilterIgnore(in.metricNameFilterIgnore),
		WithMeasurementPrefix(in.measurementPrefix),
		WithMeasurementName(in.measurementName),
		WithMeasurements(in.measurements),
		WithOutput(in.output),
		WithMaxFileSize(in.maxFileSize),
		WithTLSOpen(in.tlsOpen),
		WithUDSPath(in.udsPath),
		WithCacertFile(in.cacertFile),
		WithCertFile(in.certFile),
		WithKeyFile(in.keyFile),
		WithTagsIgnore(in.tagsIgnore),
		WithTagsRename(in.tagsRename),
		WithAsLogging(in.asLogging),
		WithIgnoreTagKV(in.ignoreTagKV),
		WithHTTPHeaders(in.httpHeaders),
		WithTags(in.tags),
		WithDisableHostTag(in.disableHostTag),
		WithDisableInstanceTag(in.disableInstanceTag),
		WithDisableInfoTag(in.disableInfoTag),
		WithAuth(in.auth),
	)
	return opts
}

func newTestPoint(name string, tags map[string]string, fields map[string]interface{}) *point.Point {
	opts := point.DefaultMetricOptions()
	return point.NewPointV2([]byte(name),
		append(point.NewTags(tags), point.NewKVs(fields)...),
		opts...)
}

func TestCollect(t *testing.T) {
	testcases := []struct {
		in     *optionMock
		u      string
		name   string
		fail   bool
		expect []string
	}{
		{
			name: "ok",
			expect: []string{
				`gogo gc_duration_seconds_count=0,gc_duration_seconds_sum=0`,
				`gogo,quantile=0 gc_duration_seconds=0`,
				`gogo,quantile=0.25 gc_duration_seconds=0`,
				`gogo,quantile=0.5 gc_duration_seconds=0`,
			},
			u: promURL,
			in: &optionMock{
				metricTypes: []string{},
				measurements: []Rule{
					{
						Pattern: `^go_.*`,
						Name:    "gogo",
						Prefix:  "go_",
					},
				},
				metricNameFilter: []string{"go"},
			},
		},

		{
			name: "option-only-URL",
			u:    promURL,
			in:   &optionMock{},
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
			u:    promURL,
			in: &optionMock{
				ignoreTagKV: map[string][]string{
					"le":          ([]string{"0.*"}),
					"status_code": ([]string{"403"}),
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
			opts := createOpts(tc.in)
			p, err := NewProm(opts...)
			if tc.fail && assert.Error(t, err) {
				return
			} else {
				assert.NoError(t, err)
			}

			p.SetClient(&http.Client{Transport: newTransportMock(mockBody)})
			p.opt.disableInstanceTag = true

			pts, err := p.CollectFromHTTPV2(tc.u)
			if tc.fail && assert.Error(t, err) {
				return
			} else {
				assert.NoError(t, err)
			}

			var arr []string
			for _, pt := range pts {
				arr = append(arr, pt.LineProto())
			}

			sort.Strings(arr)
			sort.Strings(tc.expect)

			if len(tc.expect) != len(arr) {
				t.Errorf("the length of want != got.")
			}
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
		in := &optionMock{
			tlsOpen: true,
		}
		opts := createOpts(in)
		p, err := NewProm(opts...)

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

		in := &optionMock{
			tlsOpen:    true,
			cacertFile: caFile,
		}
		opts := createOpts(in)
		p, err := NewProm(opts...)

		assert.NoError(t, err)
		transport, ok := p.client.Transport.(*http.Transport)
		assert.True(t, ok)
		assert.Equal(t, transport.TLSClientConfig.InsecureSkipVerify, false)
	})
}

func Test_Auth(t *testing.T) {
	in := &optionMock{
		auth: map[string]string{
			"type":  "bearer_token",
			"token": ".....",
		},
	}
	u := promURL
	opts := createOpts(in)
	p, err := NewProm(opts...)

	assert.NoError(t, err)
	r, err := p.GetReq(u)
	assert.NoError(t, err)
	authHeader, ok := r.Header["Authorization"]
	assert.True(t, ok)

	assert.Equal(t, len(authHeader), 1)
	assert.Contains(t, authHeader[0], "Bearer ")
}

func Test_Option(t *testing.T) {
	o := option{}

	// GetSource
	assert.Equal(t, o.GetSource("p"), "p")
	assert.Equal(t, o.GetSource(), "prom")
	o.source = "p1"
	assert.Equal(t, o.GetSource("p"), "p1")

	// GetIntervalDuration
	assert.Equal(t, o.GetIntervalDuration(), defaultInterval)
	o.interval = datakit.Duration{Duration: 1 * time.Second}
	assert.Equal(t, o.GetIntervalDuration(), 1*time.Second)
	o.interval = datakit.Duration{Duration: 10 * time.Second}
	assert.Equal(t, o.GetIntervalDuration(), 10*time.Second)
	assert.Equal(t, o.interval.Duration, 10*time.Second)
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

	in := &optionMock{
		output:      outputFile,
		maxFileSize: 100000,
	}
	u := promURL
	opts := createOpts(in)
	p, err := NewProm(opts...)

	assert.NoError(t, err)
	p.SetClient(&http.Client{Transport: newTransportMock(mockBody)})
	err = p.WriteMetricText2File(u)

	assert.NoError(t, err)

	fileContent, err := ioutil.ReadFile(outputFile)

	assert.NoError(t, err)

	assert.Equal(t, string(fileContent), mockBody)
}

func TestIgnoreReqErr(t *testing.T) {
	testCases := []struct {
		name string
		in   *optionMock
		u    string
		fail bool
	}{
		{
			name: "ignore url request error",
			in:   &optionMock{ignoreReqErr: true},
			u:    "127.0.0.1:999999",
			fail: false,
		},
		{
			name: "do not ignore url request error",
			in:   &optionMock{ignoreReqErr: false},
			u:    "127.0.0.1:999999",
			fail: true,
		},
	}
	for idx, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := createOpts(tc.in)
			p, err := NewProm(opts...)
			if err != nil {
				t.Errorf("[%d] failed to init prom: %s", idx, err)
			}
			_, err = p.CollectFromHTTPV2(tc.u)
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
		in       *optionMock
		u        string
		fail     bool
		expected []string
	}{
		{
			name: "counter metric type only",
			in:   &optionMock{metricTypes: []string{"counter"}},
			u:    promURL,
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
			in:   &optionMock{metricTypes: []string{"histogram"}},
			u:    promURL,
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
			in:   &optionMock{metricTypes: []string{}},
			u:    promURL,
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
			in:   &optionMock{metricTypes: []string{"histogram", "gauge", "counter", "summary", "untyped"}},
			u:    promURL,
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
			in:   &optionMock{metricNameFilter: []string{"http"}},
			u:    promURL,
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
			in:   &optionMock{metricNameFilter: []string{"promht+p_metric_han[a-z]ler_req[^abcd]ests_total?"}},
			u:    promURL,
			fail: false,
			expected: []string{
				"promhttp,code=200 metric_handler_requests_total=15143",
				"promhttp,code=500 metric_handler_requests_total=0",
				"promhttp,code=503 metric_handler_requests_total=0",
			},
		},

		{
			name: "measurement name prefix",
			in:   &optionMock{measurementPrefix: "prefix_"},
			u:    promURL,
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
			in:   &optionMock{measurementName: "measurement_name"},
			u:    promURL,
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
			in:   &optionMock{tagsIgnore: []string{"status_code", "method"}},
			u:    promURL,
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
			in: &optionMock{
				measurements: []Rule{
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
			u:    promURL,
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
			in:   &optionMock{tags: map[string]string{"some_tag": "some_value", "more_tag": "some_other_value"}},
			u:    promURL,
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
			opts := createOpts(tc.in)
			p, err := NewProm(opts...)
			if err != nil {
				t.Errorf("[%d] failed to init prom: %s", idx, err)
			}
			p.SetClient(&http.Client{Transport: newTransportMock(mockBody)})
			p.opt.disableInstanceTag = true
			points, err := p.CollectFromHTTPV2(tc.u)
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
				s := p.LineProto()
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

	in := &optionMock{}
	u := f.Name()
	opts := createOpts(in)
	p, err := NewProm(opts...)
	if err != nil {
		t.Errorf("failed to init prom: %s", err)
	}
	if _, err := p.CollectFromFileV2(u); err != nil {
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
	cases := []struct {
		name     string
		opt      *optionMock
		u        string
		promdata string
		expect   []*point.Point
	}{
		{
			name: "rename-tags",
			u:    "http://not-set",
			opt: &optionMock{
				tagsRename: &RenameTags{
					OverwriteExistTags: false,
					Mapping: map[string]string{
						"status_code":    "StatusCode",
						"tag-not-exists": "do-nothing",
					},
				},
			},
			expect: []*point.Point{
				func() *point.Point {
					pt := newTestPoint(
						"http",
						map[string]string{"le": "0.003", "StatusCode": "404", "method": "GET"},
						map[string]interface{}{"request_duration_seconds_bucket": 1.0},
					)
					return pt
				}(),
			},
			promdata: `
http_request_duration_seconds_bucket{le="0.003",status_code="404",method="GET"} 1
			`,
		},

		{
			name: "rename-overwrite-tags",
			u:    "http://not-set",
			opt: &optionMock{
				tagsRename: &RenameTags{
					OverwriteExistTags: true, // enable overwrite
					Mapping: map[string]string{
						"status_code": "StatusCode",
						"method":      "tag_exists", // rename `method` to a exists tag key
					},
				},
			},
			expect: []*point.Point{
				func() *point.Point {
					pt := newTestPoint(
						"http",
						map[string]string{"le": "0.003", "StatusCode": "404", "tag_exists": "GET"},
						map[string]interface{}{"request_duration_seconds_bucket": 1.0},
					)
					return pt
				}(),
			},
			promdata: `
		http_request_duration_seconds_bucket{le="0.003",tag_exists="yes",status_code="404",method="GET"} 1
					`,
		},

		{
			name: "rename-tags-disable-overwrite",
			u:    "http://not-set",
			opt: &optionMock{
				tagsRename: &RenameTags{
					OverwriteExistTags: false, // enable overwrite
					Mapping: map[string]string{
						"status_code": "StatusCode",
						"method":      "tag_exists", // rename `method` to a exists tag key
					},
				},
			},
			expect: []*point.Point{
				func() *point.Point {
					pt := newTestPoint(
						"http",
						map[string]string{"le": "0.003", "tag_exists": "yes", "StatusCode": "404", "method": "GET"}, // overwrite not work on method
						map[string]interface{}{"request_duration_seconds_bucket": 1.0},
					)
					return pt
				}(),
			},
			promdata: `
		http_request_duration_seconds_bucket{le="0.003",status_code="404",tag_exists="yes", method="GET"} 1
					`,
		},

		{
			name: "empty-tags",
			u:    "http://not-set",
			opt: &optionMock{
				tagsRename: &RenameTags{
					OverwriteExistTags: true, // enable overwrite
					Mapping: map[string]string{
						"status_code": "StatusCode",
						"method":      "tag_exists", // rename `method` to a exists tag key
					},
				},
			},
			expect: []*point.Point{
				func() *point.Point {
					pt := newTestPoint(
						"http",
						nil,
						map[string]interface{}{"request_duration_seconds_bucket": 1.0},
					)
					return pt
				}(),
			},
			promdata: `
		http_request_duration_seconds_bucket 1
					`,
		},
	}

	for _, tc := range cases {
		opts := createOpts(tc.opt)
		p, err := NewProm(opts...)
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
				left := tc.expect[idx].LineProto()
				left = left[:strings.LastIndex(left, " ")]
				right := pt.LineProto()
				right = right[:strings.LastIndex(right, " ")]
				tu.Equals(t, left, right)
				t.Log(tc.expect[idx].LineProto())
			}
		})
	}
}

func TestSetHeaders(t *testing.T) {
	testcases := []struct {
		name string
		u    string
		opt  *optionMock
	}{
		{
			name: "add custom http header",
			u:    "dummy_url",
			opt: &optionMock{
				httpHeaders: map[string]string{
					"Root":    "passwd",
					"Michael": "12345",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := createOpts(tc.opt)
			p, err := NewProm(opts...)
			if err != nil {
				t.Error(err)
				return
			}
			req, err := p.GetReq(tc.u)
			assert.NoError(t, err)
			for k, v := range p.opt.httpHeaders {
				assert.Equal(t, v, req.Header.Get(k))
			}
		})
	}
}

func TestInfoTag(t *testing.T) {
	testcases := []struct {
		u        string
		in       *optionMock
		name     string
		fail     bool
		expected []string
	}{
		{
			name: "type-info",
			u:    promURL,
			in:   &optionMock{},
			expected: []string{
				"process,host_arch=amd64,host_name=DESKTOP-3JJLRI8,os_description=Windows\\ 11\\ 10.0,os_type=windows,otel_scope_name=otlp-server,pool=mapped\\ -\\ 'non-volatile\\ memory',process_command_line=D:\\software_installer\\java\\jdk-17\\bin\\java.exe\\ -javaagent:D:/code_zy/opentelemetry-java-instrumentation/javaagent/build/libs/opentelemetry-javaagent-1.24.0-SNAPSHOT.jar\\ -Dotel.traces.exporter\\=otlp\\ -Dotel.exporter.otlp.endpoint\\=http://localhost:4317\\ -Dotel.resource.attributes\\=service.name\\=server\\,username\\=liu\\ -Dotel.metrics.exporter\\=otlp\\ -Dotel.propagators\\=b3\\ -Dotel.metrics.exporter\\=prometheus\\ -Dotel.exporter.prometheus.port\\=10086\\ -Dotel.exporter.prometheus.resource_to_telemetry_conversion.enabled\\=true\\ -XX:TieredStopAtLevel\\=1\\ -Xverify:none\\ -Dspring.output.ansi.enabled\\=always\\ -Dcom.sun.management.jmxremote\\ -Dspring.jmx.enabled\\=true\\ -Dspring.liveBeansView.mbeanDomain\\ -Dspring.application.admin.enabled\\=true\\ -javaagent:D:\\software_installer\\JetBrains\\IntelliJ\\ IDEA\\ 2022.1.4\\lib\\idea_rt.jar\\=55275:D:\\software_installer\\JetBrains\\IntelliJ\\ IDEA\\ 2022.1.4\\bin\\ -Dfile.encoding\\=UTF-8,process_executable_path=D:\\software_installer\\java\\jdk-17\\bin\\java.exe,process_pid=23592,process_runtime_description=Oracle\\ Corporation\\ Java\\ HotSpot(TM)\\ 64-Bit\\ Server\\ VM\\ 17.0.6+9-LTS-190,process_runtime_name=Java(TM)\\ SE\\ Runtime\\ Environment,process_runtime_version=17.0.6+9-LTS-190,service_name=server,telemetry_auto_version=1.24.0-SNAPSHOT,telemetry_sdk_language=java,telemetry_sdk_name=opentelemetry,telemetry_sdk_version=1.23.1,username=liu runtime_jvm_buffer_count=0",
			},
		},
	}
	mockBody := `# TYPE target info
# HELP target Target metadata
target_info{host_arch="amd64",host_name="DESKTOP-3JJLRI8",os_description="Windows 11 10.0",os_type="windows",process_command_line="D:\\software_installer\\java\\jdk-17\\bin\\java.exe -javaagent:D:/code_zy/opentelemetry-java-instrumentation/javaagent/build/libs/opentelemetry-javaagent-1.24.0-SNAPSHOT.jar -Dotel.traces.exporter=otlp -Dotel.exporter.otlp.endpoint=http://localhost:4317 -Dotel.resource.attributes=service.name=server,username=liu -Dotel.metrics.exporter=otlp -Dotel.propagators=b3 -Dotel.metrics.exporter=prometheus -Dotel.exporter.prometheus.port=10086 -Dotel.exporter.prometheus.resource_to_telemetry_conversion.enabled=true -XX:TieredStopAtLevel=1 -Xverify:none -Dspring.output.ansi.enabled=always -Dcom.sun.management.jmxremote -Dspring.jmx.enabled=true -Dspring.liveBeansView.mbeanDomain -Dspring.application.admin.enabled=true -javaagent:D:\\software_installer\\JetBrains\\IntelliJ IDEA 2022.1.4\\lib\\idea_rt.jar=55275:D:\\software_installer\\JetBrains\\IntelliJ IDEA 2022.1.4\\bin -Dfile.encoding=UTF-8",process_executable_path="D:\\software_installer\\java\\jdk-17\\bin\\java.exe",process_pid="23592",process_runtime_description="Oracle Corporation Java HotSpot(TM) 64-Bit Server VM 17.0.6+9-LTS-190",process_runtime_name="Java(TM) SE Runtime Environment",process_runtime_version="17.0.6+9-LTS-190",service_name="server",telemetry_auto_version="1.24.0-SNAPSHOT",telemetry_sdk_language="java",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="1.23.1",username="liu"} 1
# TYPE process_runtime_jvm_buffer_count gauge
# HELP process_runtime_jvm_buffer_count The number of buffers in the pool
process_runtime_jvm_buffer_count{pool="mapped - 'non-volatile memory'"} 0.0 1680231835149
# TYPE otel_scope_info info
# HELP otel_scope_info Scope metadata
otel_scope_info{otel_scope_name="otlp-server"} 1
`

	for idx, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := createOpts(tc.in)
			p, err := NewProm(opts...)
			if tc.fail && assert.Error(t, err) {
				return
			} else {
				assert.NoError(t, err)
			}

			p.SetClient(&http.Client{Transport: newTransportMock(mockBody)})
			p.opt.disableInstanceTag = true

			points, err := p.CollectFromHTTPV2(tc.u)
			if tc.fail && assert.Error(t, err) {
				return
			} else {
				assert.NoError(t, err)
			}

			var got []string
			for _, p := range points {
				s := p.LineProto()
				// remove timestamp
				got = append(got, s[:strings.LastIndex(s, " ")])
			}
			sort.Strings(got)
			tu.Equals(t, strings.Join(tc.expected, "\n"), strings.Join(got, "\n"))
			t.Logf("[%d] PASS", idx)
		})
	}
}

func TestDuplicateComment(t *testing.T) {
	testcases := []struct {
		u        string
		in       *optionMock
		name     string
		fail     bool
		expected []string
	}{
		{
			name: "duplicate-comment",
			u:    promURL,
			in:   &optionMock{},
			expected: []string{
				"go gc_duration_seconds_count=0,gc_duration_seconds_sum=0",
				"go,quantile=0 gc_duration_seconds=0",
				"go,quantile=0 gc_duration_seconds=1",
				"go,quantile=0.25 gc_duration_seconds=0",
				"go,quantile=0.25 gc_duration_seconds=1",
				"go,quantile=0.5 gc_duration_seconds=0",
				"go,quantile=0.5 gc_duration_seconds=1",
				"http,le=0.003,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.003,method=GET,status_code=404 request_duration_seconds_bucket=2i",
				"http,method=GET,status_code=404 request_duration_seconds_count=0,request_duration_seconds_sum=0",
				"promhttp metric_handler_requests_in_flight=1",
				"promhttp metric_handler_requests_in_flight=2",
				"promhttp,cause=encoding metric_handler_errors_total=0",
				"promhttp,cause=encoding metric_handler_errors_total=1",
				"promhttp,cause=gathering metric_handler_errors_total=0",
				"promhttp,cause=gathering metric_handler_errors_total=1",
				"up up=0",
				"up up=1",
			},
		},
	}
	mockBody := `# HELP promhttp_metric_handler_errors_total Total number of internal errors encountered by the promhttp metric handler.
# TYPE promhttp_metric_handler_errors_total counter
promhttp_metric_handler_errors_total{cause="encoding"} 0
promhttp_metric_handler_errors_total{cause="gathering"} 0
# HELP promhttp_metric_handler_errors_total Total number of internal errors encountered by the promhttp metric handler.
# TYPE promhttp_metric_handler_errors_total counter
promhttp_metric_handler_errors_total{cause="encoding"} 1
promhttp_metric_handler_errors_total{cause="gathering"} 1
# HELP promhttp_metric_handler_requests_in_flight Current number of scrapes being served.
# TYPE promhttp_metric_handler_requests_in_flight gauge
promhttp_metric_handler_requests_in_flight 1
# HELP promhttp_metric_handler_requests_in_flight Current number of scrapes being served.
# TYPE promhttp_metric_handler_requests_in_flight gauge
promhttp_metric_handler_requests_in_flight 2
# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0
go_gc_duration_seconds{quantile="0.25"} 0
go_gc_duration_seconds{quantile="0.5"} 0
# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 1
go_gc_duration_seconds{quantile="0.25"} 1
go_gc_duration_seconds{quantile="0.5"} 1
# HELP http_request_duration_seconds duration histogram of http responses labeled with: status_code, method
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.003",status_code="404",method="GET"} 1
# HELP http_request_duration_seconds duration histogram of http responses labeled with: status_code, method
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.003",status_code="404",method="GET"} 2
# HELP up 1 = up, 0 = not up
# TYPE up untyped
up 1
# HELP up 1 = up, 0 = not up
# TYPE up untyped
up 0
`

	for idx, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := createOpts(tc.in)
			p, err := NewProm(opts...)
			if tc.fail && assert.Error(t, err) {
				return
			} else {
				assert.NoError(t, err)
			}

			p.SetClient(&http.Client{Transport: newTransportMock(mockBody)})
			p.opt.disableInstanceTag = true

			points, err := p.CollectFromHTTPV2(tc.u)
			if tc.fail && assert.Error(t, err) {
				return
			} else {
				assert.NoError(t, err)
			}

			var got []string
			for _, p := range points {
				s := p.LineProto()
				// remove timestamp
				got = append(got, s[:strings.LastIndex(s, " ")])
			}
			sort.Strings(got)
			tu.Equals(t, strings.Join(tc.expected, "\n"), strings.Join(got, "\n"))
			t.Logf("[%d] PASS", idx)
		})
	}
}

func TestInfoType(t *testing.T) {
	testcases := []struct {
		in     *optionMock
		u      string
		name   string
		fail   bool
		expect []string
	}{
		{
			name: "no-info-type",
			u:    promURL,
			in: &optionMock{
				metricTypes: []string{"counter", "gauge", "histogram", "summary", "untyped"},
			},
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
			name: "with-info-type",
			u:    promURL,
			in:   &optionMock{},
			expect: []string{
				`go,entity=replica,name=prettier\ name,quantile=0,version=8.1.9 gc_duration_seconds=0`,
				`go,entity=replica,name=prettier\ name,quantile=0.25,version=8.1.9 gc_duration_seconds=0`,
				`go,entity=replica,name=prettier\ name,quantile=0.5,version=8.1.9 gc_duration_seconds=0`,
				`go,entity=replica,name=prettier\ name,version=8.1.9 gc_duration_seconds_count=0,gc_duration_seconds_sum=0`,
				`http,entity=replica,method=GET,name=prettier\ name,status_code=403,version=8.1.9 request_duration_seconds_count=0,request_duration_seconds_sum=0`,
				`http,entity=replica,method=GET,name=prettier\ name,status_code=404,version=8.1.9 request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013`,
				`http,le=+Inf,method=GET,status_code=403 request_duration_seconds_bucket=1i`,
				`http,le=0.003,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,le=0.03,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,le=0.1,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,le=0.3,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,le=1.2,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,le=1.5,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,le=10,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`promhttp,cause=encoding,entity=replica,name=prettier\ name,version=8.1.9 metric_handler_errors_total=0`,
				`promhttp,cause=gathering,entity=replica,name=prettier\ name,version=8.1.9 metric_handler_errors_total=0`,
				`promhttp,code=200,entity=replica,name=prettier\ name,version=8.1.9 metric_handler_requests_total=15143`,
				`promhttp,code=500,entity=replica,name=prettier\ name,version=8.1.9 metric_handler_requests_total=0`,
				`promhttp,code=503,entity=replica,name=prettier\ name,version=8.1.9 metric_handler_requests_total=0`,
				`promhttp,entity=replica,name=prettier\ name,version=8.1.9 metric_handler_requests_in_flight=1`,
				`up,entity=replica,name=prettier\ name,version=8.1.9 up=1`,
			},
		},
	}

	mockBody := `

# TYPE foo info
foo_info{name="pretty name",version="8.2.7"} 1

# TYPE foo info
foo_info{entity="controller",name="pretty name",version="8.2.7"} 1.0
foo_info{entity="replica",name="prettier name",version="8.1.9"} 1.0

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
			opts := createOpts(tc.in)
			p, err := NewProm(opts...)
			if tc.fail && assert.Error(t, err) {
				return
			} else {
				assert.NoError(t, err)
			}

			p.SetClient(&http.Client{Transport: newTransportMock(mockBody)})
			p.opt.disableInstanceTag = true

			pts, err := p.CollectFromHTTPV2(tc.u)
			if tc.fail && assert.Error(t, err) {
				return
			} else {
				assert.NoError(t, err)
			}

			var arr []string
			for _, pt := range pts {
				arr = append(arr, pt.LineProto())
			}

			sort.Strings(arr)
			sort.Strings(tc.expect)

			if len(tc.expect) != len(arr) {
				t.Errorf("the length of want != got.")
			}
			for i := range arr {
				assert.Equal(t, strings.HasPrefix(arr[i], tc.expect[i]), true)
				t.Logf(">>>\n%s\n%s", arr[i], tc.expect[i])
			}
		})
	}
}

func TestHistogramType(t *testing.T) {
	testcases := []struct {
		in     *optionMock
		u      string
		name   string
		fail   bool
		expect []string
	}{
		{
			name: "without-histogram-type",
			u:    promURL,
			in: &optionMock{
				metricTypes: []string{"counter", "gauge", "summary", "untyped"},
			},
			expect: []string{
				`etcd debugging_auth_revision=1`,
				`etcd,cluster_version=3.4 cluster_version=1`,
			},
		},

		{
			name: "with-histogram-type",
			u:    promURL,
			in:   &optionMock{},
			expect: []string{
				`etcd debugging_auth_revision=1`,
				`etcd debugging_disk_backend_commit_rebalance_duration_seconds_count=24920,debugging_disk_backend_commit_rebalance_duration_seconds_sum=0.0007340149999999998`,
				`etcd,cluster_version=3.4 cluster_version=1`,
				`etcd,le=+Inf debugging_disk_backend_commit_rebalance_duration_seconds_bucket=24920i`,
				`etcd,le=0.001 debugging_disk_backend_commit_rebalance_duration_seconds_bucket=24920i`,
				`etcd,le=0.002 debugging_disk_backend_commit_rebalance_duration_seconds_bucket=24920i`,
				`etcd,le=0.004 debugging_disk_backend_commit_rebalance_duration_seconds_bucket=24920i`,
				`etcd,le=0.008 debugging_disk_backend_commit_rebalance_duration_seconds_bucket=24920i`,
				`etcd,le=0.016 debugging_disk_backend_commit_rebalance_duration_seconds_bucket=24920i`,
				`etcd,le=0.032 debugging_disk_backend_commit_rebalance_duration_seconds_bucket=24920i`,
				`etcd,le=0.064 debugging_disk_backend_commit_rebalance_duration_seconds_bucket=24920i`,
				`etcd,le=0.128 debugging_disk_backend_commit_rebalance_duration_seconds_bucket=24920i`,
				`etcd,le=0.256 debugging_disk_backend_commit_rebalance_duration_seconds_bucket=24920i`,
				`etcd,le=0.512 debugging_disk_backend_commit_rebalance_duration_seconds_bucket=24920i`,
				`etcd,le=1.024 debugging_disk_backend_commit_rebalance_duration_seconds_bucket=24920i`,
				`etcd,le=2.048 debugging_disk_backend_commit_rebalance_duration_seconds_bucket=24920i`,
				`etcd,le=4.096 debugging_disk_backend_commit_rebalance_duration_seconds_bucket=24920i`,
				`etcd,le=8.192 debugging_disk_backend_commit_rebalance_duration_seconds_bucket=24920i`,
			},
		},
	}

	mockBody := `
etcd_cluster_version{cluster_version="3.4"} 1
# HELP etcd_debugging_auth_revision The current revision of auth store.
# TYPE etcd_debugging_auth_revision gauge
etcd_debugging_auth_revision 1
# HELP etcd_debugging_disk_backend_commit_rebalance_duration_seconds The latency distributions of commit.rebalance called by bboltdb backend.
# TYPE etcd_debugging_disk_backend_commit_rebalance_duration_seconds histogram
etcd_debugging_disk_backend_commit_rebalance_duration_seconds_bucket{le="0.001"} 24920
etcd_debugging_disk_backend_commit_rebalance_duration_seconds_bucket{le="0.002"} 24920
etcd_debugging_disk_backend_commit_rebalance_duration_seconds_bucket{le="0.004"} 24920
etcd_debugging_disk_backend_commit_rebalance_duration_seconds_bucket{le="0.008"} 24920
etcd_debugging_disk_backend_commit_rebalance_duration_seconds_bucket{le="0.016"} 24920
etcd_debugging_disk_backend_commit_rebalance_duration_seconds_bucket{le="0.032"} 24920
etcd_debugging_disk_backend_commit_rebalance_duration_seconds_bucket{le="0.064"} 24920
etcd_debugging_disk_backend_commit_rebalance_duration_seconds_bucket{le="0.128"} 24920
etcd_debugging_disk_backend_commit_rebalance_duration_seconds_bucket{le="0.256"} 24920
etcd_debugging_disk_backend_commit_rebalance_duration_seconds_bucket{le="0.512"} 24920
etcd_debugging_disk_backend_commit_rebalance_duration_seconds_bucket{le="1.024"} 24920
etcd_debugging_disk_backend_commit_rebalance_duration_seconds_bucket{le="2.048"} 24920
etcd_debugging_disk_backend_commit_rebalance_duration_seconds_bucket{le="4.096"} 24920
etcd_debugging_disk_backend_commit_rebalance_duration_seconds_bucket{le="8.192"} 24920
etcd_debugging_disk_backend_commit_rebalance_duration_seconds_bucket{le="+Inf"} 24920
etcd_debugging_disk_backend_commit_rebalance_duration_seconds_sum 0.0007340149999999998
etcd_debugging_disk_backend_commit_rebalance_duration_seconds_count 24920
`

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := createOpts(tc.in)
			p, err := NewProm(opts...)
			if tc.fail && assert.Error(t, err) {
				return
			} else {
				assert.NoError(t, err)
			}

			p.SetClient(&http.Client{Transport: newTransportMock(mockBody)})
			p.opt.disableInstanceTag = true

			pts, err := p.CollectFromHTTPV2(tc.u)
			if tc.fail && assert.Error(t, err) {
				return
			} else {
				assert.NoError(t, err)
			}

			var arr []string
			for _, pt := range pts {
				arr = append(arr, pt.LineProto())
			}

			sort.Strings(arr)
			sort.Strings(tc.expect)

			if len(tc.expect) != len(arr) {
				t.Errorf("the length of want != got.")
			}
			for i := range arr {
				assert.Equal(t, strings.HasPrefix(arr[i], tc.expect[i]), true)
				t.Logf(">>>\n%s\n%s", arr[i], tc.expect[i])
			}
		})
	}
}

func TestMetricNameFilterIgnore(t *testing.T) {
	testcases := []struct {
		in     *optionMock
		u      string
		name   string
		fail   bool
		expect []string
	}{
		{
			name: "metric-name-filter-ignore-complete-equal",
			u:    promURL,
			in: &optionMock{
				metricNameFilterIgnore: []string{
					"^http_", // this match "http" only, not "promhttp"
					"abcd",   // this not useful
				},
			},
			expect: []string{
				`go gc_duration_seconds_count=0,gc_duration_seconds_sum=0`,
				`go,quantile=0 gc_duration_seconds=0`,
				`go,quantile=0.25 gc_duration_seconds=0`,
				`go,quantile=0.5 gc_duration_seconds=0`,
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
			name: "metric-name-filter-ignore",
			u:    promURL,
			in: &optionMock{
				metricNameFilterIgnore: []string{
					"http", // this match "http" and "promhttp"
					"abcd", // this not useful
				},
			},
			expect: []string{
				`go gc_duration_seconds_count=0,gc_duration_seconds_sum=0`,
				`go,quantile=0 gc_duration_seconds=0`,
				`go,quantile=0.25 gc_duration_seconds=0`,
				`go,quantile=0.5 gc_duration_seconds=0`,
				`up up=1`,
			},
		},
		{
			name: "metric-name-filter-black-and-white",
			u:    promURL,
			in: &optionMock{
				metricNameFilterIgnore: []string{
					"promhttp", // this match "promhttp"
					"abcd",     // this not useful
				},
				metricNameFilter: []string{
					"http", // this match "http" and "promhttp", but blacklist will cancel "promhttp"
					"go",   // this match "go"
					"xyz",  // this not useful
				},
			},
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
			opts := createOpts(tc.in)
			p, err := NewProm(opts...)
			if tc.fail && assert.Error(t, err) {
				return
			} else {
				assert.NoError(t, err)
			}

			p.SetClient(&http.Client{Transport: newTransportMock(mockBody)})
			p.opt.disableInstanceTag = true

			pts, err := p.CollectFromHTTPV2(tc.u)
			if tc.fail && assert.Error(t, err) {
				return
			} else {
				assert.NoError(t, err)
			}

			var arr []string
			for _, pt := range pts {
				arr = append(arr, pt.LineProto())
			}

			sort.Strings(arr)
			sort.Strings(tc.expect)

			if len(tc.expect) != len(arr) {
				t.Errorf("the length of want != got.")
			}
			for i := range arr {
				assert.Equal(t, strings.HasPrefix(arr[i], tc.expect[i]), true)
				t.Logf(">>>\n%s\n%s", arr[i], tc.expect[i])
			}
		})
	}
}
