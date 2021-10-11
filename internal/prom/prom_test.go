package prom

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
`

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

func TestProm(t *testing.T) {
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

		pts, err := p.DebugCollect()
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
