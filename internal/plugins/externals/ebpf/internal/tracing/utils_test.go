package tracing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/spanid"
)

const (
	data = "GET /rolldice1?asdasd=1 HTTP/1.1\r\n" +
		"Host: 10.200.7.127:23306\r\n" +
		"User-Agent: python-requests/2.31.0\r\n" +
		"Accept-Encoding: gzip, deflate\r\n" +
		"Accept: */*\r\n" +
		"Connection: keep-alive\r\n" +
		"traceparent: 00-464d5ff3a7c6ba32626b625cc201d9ee-fb94ce3aa8b7363b-01\r\n" +
		"\r\n"

	data1 = "GET /rolldice1?asdasd=1 HTTP/1.1\r\n" +
		"Host: 10.200.7.127:23306\r\n" +
		"User-Agent: python-requests/2.31.0\r\n" +
		"Accept-Encoding: gzip, deflate\r\n" +
		"Accept: */*\r\n" +
		"Connection: keep-alive\r\n" +
		"traceparent: 00-464d5ff3a7c6ba32626b625cc201d9ee-28cb6ae746469c5b-01\r\n" +
		"\r\n" +
		"\r\n" +
		"fsdfsf: ff"

	data2 = "GET /rolldice2?asdasd=1 HTTP/1.1\r\n" +
		"Host: 10.200.7.127:23307\r\n" +
		"User-Agent: python-requests/2.31.0\r\n" +
		"Accept-Encoding: gzip, deflate\r\n" +
		"Accept: */*\r\n" +
		"Connection: keep-alive\r\n" +
		"x-datadog-trace-id: 7091870188756392430\r\n" +
		"x-datadog-parent-id: 2939560723338402907\r\n" +
		"x-datadog-sampling-priority: 1\r\n" +
		"x-datadog-tags: _dd.p.tid=464d5ff3a7c6ba32\r\n" +
		"traceparent: 00-464d5ff3a7c6ba32626b625cc201d9ee-28cb6ae746469c5b-01\r\n" +
		"tracestate: dd=s:1;t.tid:464d5ff3a7c6ba32\r\n"

	data3 = "GET /rolldice HTTP/1.1\r\n" +
		"Host: 10.200.7.127:23305\r\n" +
		"User-Agent: curl/7.81.0\r\n" +
		"Accept: */*\r\n"
)

func TestParseHTTPHeader(t *testing.T) {
	if v, ok := ParseHTTP1xHeader([]byte(data), 0, false); !ok {
		t.Error("!ok")
	} else {
		assert.Equal(t, TraceInfo{
			Method:       "GET",
			Path:         "/rolldice1",
			Version:      "HTTP/1.1",
			TraceID:      HexTraceid2ID128("464d5ff3a7c6ba32626b625cc201d9ee"),
			ParentSpanID: HexSpanid2ID64("fb94ce3aa8b7363b"),
			ASpanSampled: true,
			HaveTracID:   true,
			HexEncode:    true,
			Headers: map[string]string{
				"Host":            "10.200.7.127:23306",
				"User-Agent":      "python-requests/2.31.0",
				"Accept-Encoding": "gzip, deflate",
				"Accept":          "*/*",
				"Connection":      "keep-alive",
				"traceparent":     "00-464d5ff3a7c6ba32626b625cc201d9ee-fb94ce3aa8b7363b-01",
			},
			Param: "asdasd=1",
		}, *v)
	}

	if v, ok := ParseHTTP1xHeader([]byte(data1), 0, true); !ok {
		t.Error("!ok")
	} else {
		assert.Equal(t, TraceInfo{
			Method:  "GET",
			Path:    "/rolldice1",
			Version: "HTTP/1.1",
			TraceID: spanid.ID128{
				Low: uint64(DecTraceOrSpanid2ID64("7091870188756392430")),
			},
			ParentSpanID: DecTraceOrSpanid2ID64("2939560723338402907"),
			ASpanSampled: true,
			HaveTracID:   true,
			Headers: map[string]string{
				"Host":            "10.200.7.127:23306",
				"User-Agent":      "python-requests/2.31.0",
				"Accept-Encoding": "gzip, deflate",
				"Accept":          "*/*",
				"Connection":      "keep-alive",
				"traceparent":     "00-464d5ff3a7c6ba32626b625cc201d9ee-28cb6ae746469c5b-01",
			},
			Param: "asdasd=1",
		}, *v)
	}

	if v, ok := ParseHTTP1xHeader([]byte(data2), 0, true); !ok {
		t.Error("!ok")
	} else {
		assert.Equal(t, TraceInfo{
			Method:  "GET",
			Path:    "/rolldice2",
			Version: "HTTP/1.1",
			TraceID: spanid.ID128{
				Low: uint64(DecTraceOrSpanid2ID64("7091870188756392430")),
			},
			ParentSpanID: DecTraceOrSpanid2ID64("2939560723338402907"),
			ASpanSampled: true,
			Headers: map[string]string{
				"Host":                        "10.200.7.127:23307",
				"User-Agent":                  "python-requests/2.31.0",
				"Accept-Encoding":             "gzip, deflate",
				"Accept":                      "*/*",
				"Connection":                  "keep-alive",
				"x-datadog-trace-id":          "7091870188756392430",
				"x-datadog-parent-id":         "2939560723338402907",
				"x-datadog-sampling-priority": "1",
				"x-datadog-tags":              "_dd.p.tid=464d5ff3a7c6ba32",
				"traceparent":                 "00-464d5ff3a7c6ba32626b625cc201d9ee-28cb6ae746469c5b-01",
				"tracestate":                  "dd=s:1;t.tid:464d5ff3a7c6ba32",
			},
			HaveTracID: true,
			Param:      "asdasd=1",
		}, *v)
	}

	if v, ok := ParseHTTP1xHeader([]byte(data3), 0, true); !ok {
		t.Error("!ok")
	} else {
		assert.Equal(t, TraceInfo{
			Method:       "GET",
			Path:         "/rolldice",
			Version:      "HTTP/1.1",
			ASpanSampled: true,
			HaveTracID:   false,
			Headers: map[string]string{
				"Host":       "10.200.7.127:23305",
				"User-Agent": "curl/7.81.0",
				"Accept":     "*/*",
			},
		}, *v)
	}
}
