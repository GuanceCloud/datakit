// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promtail

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	tu "github.com/GuanceCloud/cliutils/testutil"
	"github.com/stretchr/testify/assert"
)

// GZip source string and return compressed string
func gzipString(source string) string {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write([]byte(source)); err != nil {
		log.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		log.Fatal(err)
	}
	return buf.String()
}

// Deflate source string and return compressed string
func deflateString(source string) string {
	var buf bytes.Buffer
	zw, _ := flate.NewWriter(&buf, 6)
	if _, err := zw.Write([]byte(source)); err != nil {
		log.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		log.Fatal(err)
	}
	return buf.String()
}

func TestParseRequest(t *testing.T) {
	cases := []struct {
		name            string
		path            string
		body            string
		contentType     string
		contentEncoding string
		valid           bool
		legacy          bool
	}{
		{
			name:        "empty body",
			path:        `/v1/write/promtail`,
			body:        ``,
			contentType: `application/json`,
			valid:       false,
		},
		{
			name:        "empty content type",
			path:        `/v1/write/promtail`,
			body:        `{"streams": [{ "stream": { "foo": "bar2" }, "values": [ [ "1570818238000000000", "fizzbuzz" ] ] }]}`,
			contentType: ``,
			valid:       false,
		},
		{
			name:        "json",
			path:        `/v1/write/promtail`,
			body:        `{"streams": [{ "stream": { "foo": "bar2" }, "values": [ [ "1570818238000000000", "fizzbuzz" ] ] }]}`,
			contentType: `application/json`,
			valid:       true,
		},
		{
			name:            "empty content encoding",
			path:            `/v1/write/promtail`,
			body:            `{"streams": [{ "stream": { "foo": "bar2" }, "values": [ [ "1570818238000000000", "fizzbuzz" ] ] }]}`,
			contentType:     `application/json`,
			contentEncoding: ``,
			valid:           true,
		},
		{
			name:            "gzip",
			path:            `/v1/write/promtail`,
			body:            gzipString(`{"streams": [{ "stream": { "foo": "bar2" }, "values": [ [ "1570818238000000000", "fizzbuzz" ] ] }]}`),
			contentType:     `application/json`,
			contentEncoding: `gzip`,
			valid:           true,
		},
		{
			name:            "defalte",
			path:            `/v1/write/promtail`,
			body:            deflateString(`{"streams": [{ "stream": { "foo": "bar2" }, "values": [ [ "1570818238000000000", "fizzbuzz" ] ] }]}`),
			contentType:     `application/json`,
			contentEncoding: `deflate`,
			valid:           true,
		},
		{
			name:            "expect snappy but gzip",
			path:            `/v1/write/promtail`,
			body:            gzipString(`{"streams": [{ "stream": { "foo": "bar2" }, "values": [ [ "1570818238000000000", "fizzbuzz" ] ] }]}`),
			contentType:     `application/json`,
			contentEncoding: `snappy`,
			valid:           false,
		},
		{
			name:            "gzip",
			path:            `/v1/write/promtail`,
			body:            gzipString(`{"streams": [{ "stream": { "foo": "bar2" }, "values": [ [ "1570818238000000000", "fizzbuzz" ] ] }]}`),
			contentType:     `application/json; charset=utf-8`,
			contentEncoding: `gzip`,
			valid:           true,
		},
		{
			name:            "deflate",
			path:            `/v1/write/promtail`,
			body:            deflateString(`{"streams": [{ "stream": { "foo": "bar2" }, "values": [ [ "1570818238000000000", "fizzbuzz" ] ] }]}`),
			contentType:     `application/json; charset=utf-8`,
			contentEncoding: `deflate`,
			valid:           true,
		},
		{
			name:            "incorrect content type with gzip",
			path:            `/v1/write/promtail`,
			body:            gzipString(`{"streams": [{ "stream": { "foo": "bar2" }, "values": [ [ "1570818238000000000", "fizzbuzz" ] ] }]}`),
			contentType:     `application/jsonn; charset=utf-8`,
			contentEncoding: `gzip`,
			valid:           false,
		},
		{
			name:            "incorrect content type with deflate",
			path:            `/v1/write/promtail`,
			body:            deflateString(`{"streams": [{ "stream": { "foo": "bar2" }, "values": [ [ "1570818238000000000", "fizzbuzz" ] ] }]}`),
			contentType:     `application/jsonn; charset=utf-8`,
			contentEncoding: `deflate`,
			valid:           false,
		},
		{
			name:            "incorrect charset with gzip",
			path:            `/v1/write/promtail`,
			body:            gzipString(`{"streams": [{ "stream": { "foo4": "bar2" }, "values": [ [ "1570818238000000000", "fizzbuzz" ] ] }]}`),
			contentType:     `application/json; charsetutf-8`,
			contentEncoding: `gzip`,
			valid:           false,
		},
		{
			name:            "incorrect charset with deflate",
			path:            `/v1/write/promtail`,
			body:            deflateString(`{"streams": [{ "stream": { "foo4": "bar2" }, "values": [ [ "1570818238000000000", "fizzbuzz" ] ] }]}`),
			contentType:     `application/json; charsetutf-8`,
			contentEncoding: `deflate`,
			valid:           false,
		},
		{
			name:            "incorrect charset with deflate",
			path:            `/v1/write/promtail`,
			body:            deflateString(`{"streams": [{ "stream": { "foo4": "bar2" }, "values": [ [ "1570818238000000000", "fizzbuzz" ] ] }]}`),
			contentType:     `application/json; charsetutf-8`,
			contentEncoding: `deflate`,
			valid:           false,
		},
		{
			name:            "legacy unmarshal",
			path:            `/v1/write/promtail`,
			body:            deflateString(`{"streams":[{"labels":"{foo: \"bar\"}","entries":[{"ts":"0001-01-01T00:00:00Z","line":"fizzbuzz"}]}]}`),
			contentType:     `application/json; charset=utf-8`,
			contentEncoding: `deflate`,
			valid:           true,
			legacy:          true,
		},
		{
			name:            "unsupported content-encoding",
			path:            `/v1/write/promtail`,
			body:            deflateString(`{"streams": [{ "stream": { "foo4": "bar2" }, "values": [ [ "1570818238000000000", "fizzbuzz" ] ] }]}`),
			contentType:     `application/json; charsetutf-8`,
			contentEncoding: `unsupported content-encoding`,
			valid:           false,
		},
	}

	for index, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest("POST", test.path, strings.NewReader(test.body))
			if len(test.contentType) > 0 {
				request.Header.Add("Content-Type", test.contentType)
			}
			if len(test.contentEncoding) > 0 {
				request.Header.Add("Content-Encoding", test.contentEncoding)
			}
			i := Input{Legacy: test.legacy}
			pr, err := i.parseRequest(request)
			if test.valid {
				t.Log(pr)
				assert.Nil(t, err, "Should not give error for %d", index)
				assert.NotNil(t, pr, "Should give data for %d", index)
				assert.NotNil(t, pr.Streams[0].Entries, "Parse result should have at least one entry")
			} else {
				assert.NotNil(t, err, "Should give error for %d", index)
				assert.Nil(t, pr, "Should not give data for %d", index)
			}
		})
	}
}

func TestGetSource(t *testing.T) {
	cases := []struct {
		name           string
		req            *http.Request
		expectedSource string
	}{
		{
			name:           "source default",
			req:            httptest.NewRequest("POST", "/v1/write/promtail", nil),
			expectedSource: "default",
		},
		{
			name:           "source test",
			req:            httptest.NewRequest("POST", "/v1/write/promtail?source=test", nil),
			expectedSource: "test",
		},
		{
			name:           "url with other args",
			req:            httptest.NewRequest("POST", "/v1/write/promtail?source=test&a=b", nil),
			expectedSource: "test",
		},
		{
			name:           "url with other args",
			req:            httptest.NewRequest("POST", "/v1/write/promtail?a=c&source=b", nil),
			expectedSource: "b",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedSource, getSource(tc.req))
		})
	}
}

func TestGetPipelinePath(t *testing.T) {
	cases := []struct {
		name     string
		req      *http.Request
		expected string
	}{
		{
			name:     "no pipeline",
			req:      httptest.NewRequest("POST", "/v1/write/promtail", nil),
			expected: "",
		},
		{
			name:     "test.p",
			req:      httptest.NewRequest("POST", "/v1/write/promtail?pipeline=test.p", nil),
			expected: "test.p",
		},
		{
			name:     "multiple args",
			req:      httptest.NewRequest("POST", "/v1/write/promtail?source=a&pipeline=apache.p&a=b", nil),
			expected: "apache.p",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, getPipelinePath(tc.req))
		})
	}
}

func TestGetCustomTags(t *testing.T) {
	cases := []struct {
		name     string
		req      *http.Request
		expected map[string]string
	}{
		{
			name: "two extra tags",
			req:  httptest.NewRequest("POST", "/v1/write/promtail?tags=key1=value1,key2=value2", nil),
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name:     "no extra tags",
			req:      httptest.NewRequest("POST", "/v1/write/promtail", nil),
			expected: map[string]string{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tu.Equals(t, tc.expected, getCustomTags(tc.req))
		})
	}
}
