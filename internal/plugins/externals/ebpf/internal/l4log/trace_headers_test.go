//go:build linux
// +build linux

package l4log

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

func TestTraceHeaderCarrierTraceIDs(t *testing.T) {
	sw8 := makeTestSW8("sky-trace-id", "sky-segment-id", "7")

	cases := []struct {
		name     string
		headers  [][2]string
		traceID  string
		parentID string
	}{
		{
			name: "datadog",
			headers: [][2]string{
				{"x-datadog-trace-id", "7091870188756392430"},
				{"x-datadog-parent-id", "2939560723338402907"},
			},
			traceID:  "7091870188756392430",
			parentID: "2939560723338402907",
		},
		{
			name: "datadog-preferred-over-traceparent",
			headers: [][2]string{
				{"traceparent", "00-464d5ff3a7c6ba32626b625cc201d9ee-fb94ce3aa8b7363b-01"},
				{"x-datadog-trace-id", "7091870188756392430"},
				{"x-datadog-parent-id", "2939560723338402907"},
			},
			traceID:  "7091870188756392430",
			parentID: "2939560723338402907",
		},
		{
			name: "datadog-span-id-fallback",
			headers: [][2]string{
				{"x-datadog-trace-id", "7091870188756392430"},
				{"x-datadog-span-id", "2939560723338402907"},
			},
			traceID:  "7091870188756392430",
			parentID: "2939560723338402907",
		},
		{
			name: "invalid-datadog-falls-back-to-traceparent",
			headers: [][2]string{
				{"x-datadog-trace-id", "not-a-number"},
				{"x-datadog-parent-id", "2939560723338402907"},
				{"traceparent", "00-464d5ff3a7c6ba32626b625cc201d9ee-fb94ce3aa8b7363b-01"},
			},
			traceID:  "464d5ff3a7c6ba32626b625cc201d9ee",
			parentID: "fb94ce3aa8b7363b",
		},
		{
			name: "oversized-datadog-falls-back-to-traceparent",
			headers: [][2]string{
				{"x-datadog-trace-id", strings.Repeat("9", maxDatadogDecimalIDLen+1)},
				{"x-datadog-parent-id", "2939560723338402907"},
				{"traceparent", "00-464d5ff3a7c6ba32626b625cc201d9ee-fb94ce3aa8b7363b-01"},
			},
			traceID:  "464d5ff3a7c6ba32626b625cc201d9ee",
			parentID: "fb94ce3aa8b7363b",
		},
		{
			name: "traceparent",
			headers: [][2]string{
				{"traceparent", "00-464d5ff3a7c6ba32626b625cc201d9ee-fb94ce3aa8b7363b-01"},
			},
			traceID:  "464d5ff3a7c6ba32626b625cc201d9ee",
			parentID: "fb94ce3aa8b7363b",
		},
		{
			name: "b3-multi",
			headers: [][2]string{
				{"X-B3-TraceId", "463ac35c9f6413ad48485a3953bb6124"},
				{"X-B3-SpanId", "a2fb4a1d1a96d312"},
			},
			traceID:  "463ac35c9f6413ad48485a3953bb6124",
			parentID: "a2fb4a1d1a96d312",
		},
		{
			name: "b3-single",
			headers: [][2]string{
				{"b3", "463ac35c9f6413ad-a2fb4a1d1a96d312-1"},
			},
			traceID:  "463ac35c9f6413ad",
			parentID: "a2fb4a1d1a96d312",
		},
		{
			name: "sw8",
			headers: [][2]string{
				{"sw8", sw8},
			},
			traceID:  "sky-trace-id",
			parentID: "sky-segment-id7",
		},
		{
			name: "sw8-oversized-parent-dropped",
			headers: [][2]string{
				{"sw8", makeTestSW8("sky-trace-id",
					strings.Repeat("s", maxSW8SegmentIDLen+1),
					strings.Repeat("1", maxSW8SpanIDLen+1))},
			},
			traceID: "sky-trace-id",
		},
		{
			name: "jaeger",
			headers: [][2]string{
				{"uber-trace-id", "463ac35c9f6413ad48485a3953bb6124:a2fb4a1d1a96d312:0:1"},
			},
			traceID:  "463ac35c9f6413ad48485a3953bb6124",
			parentID: "a2fb4a1d1a96d312",
		},
		{
			name: "opentracing",
			headers: [][2]string{
				{"ot-tracer-traceid", "463ac35c9f6413ad48485a3953bb6124"},
				{"ot-tracer-spanid", "a2fb4a1d1a96d312"},
			},
			traceID:  "463ac35c9f6413ad48485a3953bb6124",
			parentID: "a2fb4a1d1a96d312",
		},
		{
			name: "xray",
			headers: [][2]string{
				{"x-amzn-trace-id", "Root=1-5759e988-bd862e3fe1be46a994272793;Parent=53995c3f42cd8ad8;Sampled=1"},
			},
			traceID:  "5759e988bd862e3fe1be46a994272793",
			parentID: "53995c3f42cd8ad8",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var carrier traceHeaderCarrier
			for _, header := range tc.headers {
				carrier.addString(header[0], header[1])
			}

			traceID, parentID := carrier.traceIDs()
			if traceID != tc.traceID || parentID != tc.parentID {
				t.Fatalf("unexpected trace IDs: got (%q, %q), want (%q, %q)",
					traceID, parentID, tc.traceID, tc.parentID)
			}
		})
	}
}

func TestTraceHeaderCarrierInvalidTraceIDs(t *testing.T) {
	cases := []struct {
		name    string
		headers [][2]string
	}{
		{
			name: "zero-datadog",
			headers: [][2]string{
				{"x-datadog-trace-id", "0"},
			},
		},
		{
			name: "oversized-datadog",
			headers: [][2]string{
				{"x-datadog-trace-id", strings.Repeat("9", maxDatadogDecimalIDLen+1)},
				{"x-datadog-parent-id", "2939560723338402907"},
			},
		},
		{
			name: "overflow-datadog",
			headers: [][2]string{
				{"x-datadog-trace-id", "18446744073709551616"},
				{"x-datadog-parent-id", "2939560723338402907"},
			},
		},
		{
			name: "zero-traceparent",
			headers: [][2]string{
				{"traceparent", "00-00000000000000000000000000000000-fb94ce3aa8b7363b-01"},
			},
		},
		{
			name: "traceparent-zero-parent",
			headers: [][2]string{
				{"traceparent", "00-464d5ff3a7c6ba32626b625cc201d9ee-0000000000000000-01"},
			},
		},
		{
			name: "b3-sampled-only",
			headers: [][2]string{
				{"b3", "1"},
			},
		},
		{
			name: "invalid-sw8",
			headers: [][2]string{
				{"sw8", "1-not-base64-seg-1"},
			},
		},
		{
			name: "oversized-sw8-trace",
			headers: [][2]string{
				{"sw8", makeTestSW8(strings.Repeat("t", maxSW8TraceIDLen+1), "sky-segment-id", "7")},
			},
		},
		{
			name: "zero-xray",
			headers: [][2]string{
				{"x-amzn-trace-id", "Root=1-00000000-000000000000000000000000;Parent=53995c3f42cd8ad8"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var carrier traceHeaderCarrier
			for _, header := range tc.headers {
				carrier.addString(header[0], header[1])
			}

			traceID, parentID := carrier.traceIDs()
			if traceID != "" || parentID != "" {
				t.Fatalf("expected empty trace IDs, got (%q, %q)", traceID, parentID)
			}
		})
	}
}

func TestParseHTTPRequestMetaTraceHeaders(t *testing.T) {
	cases := []struct {
		name     string
		headers  string
		traceID  string
		parentID string
	}{
		{
			name: "datadog",
			headers: "X-Datadog-Trace-Id: 7091870188756392430\r\n" +
				"X-Datadog-Parent-Id: 2939560723338402907\r\n",
			traceID:  "7091870188756392430",
			parentID: "2939560723338402907",
		},
		{
			name: "b3",
			headers: "X-B3-TraceId: 463ac35c9f6413ad48485a3953bb6124\r\n" +
				"X-B3-SpanId: a2fb4a1d1a96d312\r\n",
			traceID:  "463ac35c9f6413ad48485a3953bb6124",
			parentID: "a2fb4a1d1a96d312",
		},
		{
			name:     "sw8",
			headers:  "sw8: " + makeTestSW8("sky-trace-id", "sky-segment-id", "7") + "\r\n",
			traceID:  "sky-trace-id",
			parentID: "sky-segment-id7",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := "GET /api/v1?q=1 HTTP/1.1\r\n" +
				"Host: api.example\r\n" +
				tc.headers +
				"\r\n"

			method, path, host, traceID, parentID, ok := parseHTTPRequestMeta([]byte(req))
			if !ok {
				t.Fatal("expected HTTP request metadata")
			}
			if method != "GET" || path != "/api/v1" || host != "api.example" {
				t.Fatalf("unexpected HTTP metadata: method=%q path=%q host=%q", method, path, host)
			}
			if traceID != tc.traceID || parentID != tc.parentID {
				t.Fatalf("unexpected trace IDs: got (%q, %q), want (%q, %q)",
					traceID, parentID, tc.traceID, tc.parentID)
			}
		})
	}
}

func TestL7LogHeadersRecordsRecommendedHeaders(t *testing.T) {
	req := "POST /api/v1?q=1 HTTP/1.1\r\n" +
		"Host: Api.Example:443\r\n" +
		"User-Agent: curl/8.0\r\n" +
		"X-Request-ID: req-123\r\n" +
		"X-Forwarded-For: 10.0.0.1\r\n" +
		"Content-Type: application/json\r\n" +
		"Authorization: Bearer secret\r\n" +
		"X-Internal-Noise: skip-me\r\n" +
		"\r\n"

	method, path, host, traceID, parentID, provider, headers, ok := parseHTTPRequestMetaWithHeaders([]byte(req), true)
	if !ok {
		t.Fatal("expected HTTP request metadata")
	}
	if method != "POST" || path != "/api/v1" || host != "api.example" {
		t.Fatalf("unexpected HTTP metadata: method=%q path=%q host=%q", method, path, host)
	}
	if traceID != "" || parentID != "" || provider != "" {
		t.Fatalf("unexpected trace metadata: trace_id=%q parent_id=%q provider=%q", traceID, parentID, provider)
	}

	assertHeaderValue(t, headers, "host", "Api.Example:443")
	assertHeaderValue(t, headers, "user-agent", "curl/8.0")
	assertHeaderValue(t, headers, "x-request-id", "req-123")
	assertHeaderValue(t, headers, "x-forwarded-for", "10.0.0.1")
	assertHeaderValue(t, headers, "content-type", "application/json")
	assertHeaderAbsent(t, headers, "authorization")
	assertHeaderAbsent(t, headers, "x-internal-noise")
}

func TestL7LogHeadersRecordsResponseHeaders(t *testing.T) {
	resp := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: application/json\r\n" +
		"X-Request-ID: req-123\r\n" +
		"Set-Cookie: token=secret\r\n" +
		"\r\n"

	status, headers, ok := parseHTTPResponseMeta([]byte(resp), true)
	if !ok {
		t.Fatal("expected HTTP response metadata")
	}
	if status != 200 {
		t.Fatalf("unexpected status: %d", status)
	}

	assertHeaderValue(t, headers, "content-type", "application/json")
	assertHeaderValue(t, headers, "x-request-id", "req-123")
	assertHeaderAbsent(t, headers, "set-cookie")
}

func TestHTTPLogRecordsSuggestedMetadataToPoint(t *testing.T) {
	reqLength := int64(15)
	respLength := int64(20)
	elem := &HTTPLogElem{
		Method:        "POST",
		Path:          "/api",
		Direction:     DOutging,
		TraceID:       "7091870188756392430",
		ParentID:      "2939560723338402907",
		TraceProvider: "datadog",
		txFirstByteTS: 10,
		messageDirty:  true,
	}
	elem.setReqHeaders(map[string]string{
		"user-agent":     "curl/8.0",
		"content-type":   "application/json",
		"content-length": "15",
		"x-request-id":   "req-http",
	})
	elem.setRespHeaders(map[string]string{
		"content-type":   "application/json",
		"content-length": "20",
	})

	kvs, _, ok, err := buildHTTPLog(&PMeta{}, &PValue{}, elem, point.NewTags(nil), nil, "ns-http", nil)
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	if !ok {
		t.Fatal("expected HTTP log point")
	}

	assertKVString(t, kvs, "trace_provider", "datadog")
	assertKVString(t, kvs, "http_user_agent", "curl/8.0")
	assertKVString(t, kvs, "http_req_content_type", "application/json")
	assertKVString(t, kvs, "http_resp_content_type", "application/json")
	assertKVInt(t, kvs, "http_req_content_length", reqLength)
	assertKVInt(t, kvs, "http_resp_content_length", respLength)
	assertKVInt(t, kvs, "http_header_count", 6)
	if elem.HeaderBytes == 0 {
		t.Fatal("expected header byte summary")
	}
}

func TestL7LogHeaderValueLimit(t *testing.T) {
	headers := recordL7LogHeaderString(nil, "user-agent", strings.Repeat("a", maxL7HeaderValueLen+20))
	if got := len(headers["user-agent"]); got != maxL7HeaderValueLen {
		t.Fatalf("unexpected header value length: %d", got)
	}
}

func TestHTTP2LogParsesTraceHeaders(t *testing.T) {
	cases := []struct {
		name     string
		headers  []hpack.HeaderField
		traceID  string
		parentID string
	}{
		{
			name: "datadog",
			headers: []hpack.HeaderField{
				{Name: "x-datadog-trace-id", Value: "7091870188756392430"},
				{Name: "x-datadog-parent-id", Value: "2939560723338402907"},
			},
			traceID:  "7091870188756392430",
			parentID: "2939560723338402907",
		},
		{
			name: "b3",
			headers: []hpack.HeaderField{
				{Name: "x-b3-traceid", Value: "463ac35c9f6413ad48485a3953bb6124"},
				{Name: "x-b3-spanid", Value: "a2fb4a1d1a96d312"},
			},
			traceID:  "463ac35c9f6413ad48485a3953bb6124",
			parentID: "a2fb4a1d1a96d312",
		},
		{
			name: "sw8",
			headers: []hpack.HeaderField{
				{Name: "sw8", Value: makeTestSW8("sky-trace-id", "sky-segment-id", "7")},
			},
			traceID:  "sky-trace-id",
			parentID: "sky-segment-id7",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			headers := []hpack.HeaderField{
				{Name: H2HdrMethod, Value: "GET"},
				{Name: H2HdrPath, Value: "/api/v1"},
				{Name: H2HdrHost, Value: "api.example"},
			}
			headers = append(headers, tc.headers...)

			payload := buildTestH2HeadersPayload(t, headers)
			h2log := NewH2Log()
			h2log.Handle(directionTX, payload, int64(len(payload)), &PktTCPHdr{Seq: 100, TS: 200}, nil, 0, 1)

			if len(h2log.elems) != 1 {
				t.Fatalf("expected one HTTP2 log element, got %d", len(h2log.elems))
			}
			elem := h2log.elems[0]
			if elem.TraceID != tc.traceID || elem.ParentID != tc.parentID {
				t.Fatalf("unexpected trace IDs: got (%q, %q), want (%q, %q)",
					elem.TraceID, elem.ParentID, tc.traceID, tc.parentID)
			}
		})
	}
}

func TestHTTP2LogRecordsRecommendedHeaders(t *testing.T) {
	oldEnableNetlog := enableNetlog
	enableNetlog = true
	defer func() {
		enableNetlog = oldEnableNetlog
	}()

	reqHeaders := []hpack.HeaderField{
		{Name: H2HdrMethod, Value: "POST"},
		{Name: H2HdrPath, Value: "/grpc.Service/Call"},
		{Name: H2HdrHost, Value: "api.example"},
		{Name: "user-agent", Value: "grpc-go/1.60"},
		{Name: "content-type", Value: "application/grpc"},
		{Name: "x-request-id", Value: "req-h2"},
		{Name: "authorization", Value: "Bearer secret"},
	}
	respHeaders := []hpack.HeaderField{
		{Name: H2HdrStatus, Value: "200"},
		{Name: "grpc-status", Value: "0"},
		{Name: "grpc-message", Value: "ok"},
		{Name: "set-cookie", Value: "token=secret"},
	}

	h2log := NewH2Log()
	reqPayload := buildTestH2HeadersPayload(t, reqHeaders)
	h2log.Handle(directionTX, reqPayload, int64(len(reqPayload)), &PktTCPHdr{Seq: 100, TS: 200}, nil, 0, 1)
	respPayload := buildTestH2HeadersFramePayload(t, respHeaders, false)
	h2log.Handle(directionRX, respPayload, int64(len(respPayload)), &PktTCPHdr{Seq: 300, TS: 400}, nil, 0, 2)

	if len(h2log.elems) != 1 {
		t.Fatalf("expected one HTTP2 log element, got %d", len(h2log.elems))
	}
	elem := h2log.elems[0]
	assertHeaderValue(t, elem.ReqHeaders, "host", "api.example")
	assertHeaderValue(t, elem.ReqHeaders, "user-agent", "grpc-go/1.60")
	assertHeaderValue(t, elem.ReqHeaders, "content-type", "application/grpc")
	assertHeaderValue(t, elem.ReqHeaders, "x-request-id", "req-h2")
	assertHeaderAbsent(t, elem.ReqHeaders, "authorization")
	assertHeaderValue(t, elem.RespHeaders, "grpc-status", "0")
	assertHeaderValue(t, elem.RespHeaders, "grpc-message", "ok")
	assertHeaderAbsent(t, elem.RespHeaders, "set-cookie")
}

func TestHTTPLogMessageJSONIncludesHeaders(t *testing.T) {
	elem := &HTTPLogElem{
		Method:       "GET",
		Path:         "/api",
		ReqHeaders:   map[string]string{"x-request-id": "req-json"},
		RespHeaders:  map[string]string{"content-type": "application/json"},
		messageDirty: true,
	}
	msg, err := httpLogMessageJSON(elem)
	if err != nil {
		t.Fatal(err)
	}

	var decoded struct {
		HTTP struct {
			ReqHeaders  map[string]string `json:"req_headers"`
			RespHeaders map[string]string `json:"resp_headers"`
		} `json:"http"`
	}
	if err := json.Unmarshal([]byte(msg), &decoded); err != nil {
		t.Fatal(err)
	}
	assertHeaderValue(t, decoded.HTTP.ReqHeaders, "x-request-id", "req-json")
	assertHeaderValue(t, decoded.HTTP.RespHeaders, "content-type", "application/json")
}

func TestHTTP2LogFlushesTraceHeadersToPoint(t *testing.T) {
	oldEnableNetlog := enableNetlog
	enableNetlog = true
	defer func() {
		enableNetlog = oldEnableNetlog
	}()

	headers := []hpack.HeaderField{
		{Name: H2HdrMethod, Value: "GET"},
		{Name: H2HdrPath, Value: "/api/v1"},
		{Name: H2HdrHost, Value: "api.example"},
		{Name: "user-agent", Value: "grpc-go/1.60"},
		{Name: "content-type", Value: "application/grpc"},
		{Name: "content-length", Value: "0"},
		{Name: "x-datadog-trace-id", Value: "7091870188756392430"},
		{Name: "x-datadog-parent-id", Value: "2939560723338402907"},
	}
	payload := buildTestH2HeadersPayload(t, headers)
	value := &PValue{
		sMACEQ: true,
		tcpInfo: TCPLog{
			direction: directionOutgoing,
		},
	}

	if !value.http2Info.ShouldHandle(directionTX, payload) {
		t.Fatal("expected HTTP2 log to handle client preface")
	}
	value.http2Info.Handle(directionTX, payload, int64(len(payload)), &PktTCPHdr{Seq: 100, TS: 200}, nil, 0, 1)
	respHeaders := []hpack.HeaderField{
		{Name: H2HdrStatus, Value: "200"},
		{Name: "content-type", Value: "application/grpc"},
		{Name: "grpc-status", Value: "0"},
		{Name: "grpc-message", Value: "ok"},
	}
	respPayload := buildTestH2HeadersFramePayload(t, respHeaders, false)
	value.http2Info.Handle(directionRX, respPayload, int64(len(respPayload)), &PktTCPHdr{Seq: 300, TS: 400}, nil, 0, 2)

	conns := &TCPConns{
		nsUID:        "ns-h2",
		ifaceNameMAC: [2]string{"eth0", "aa:bb:cc:dd:ee:ff"},
	}
	pts, err := conns.netlogConv2Point(&PMeta{
		SrcIP: "10.0.0.1", DstIP: "10.0.0.2", SrcPort: 12345, DstPort: 8080,
	}, value, point.CommonLoggingOptions(), true, nil)
	if err != nil {
		t.Fatalf("unexpected conversion error: %v", err)
	}
	if len(pts) != 1 {
		t.Fatalf("expected one log point, got %d", len(pts))
	}
	if got := pts[0].GetTag("trace_id"); got != "7091870188756392430" {
		t.Fatalf("unexpected trace_id tag: %q", got)
	}
	if got := pts[0].GetTag("parent_id"); got != "2939560723338402907" {
		t.Fatalf("unexpected parent_id tag: %q", got)
	}
	if got := pts[0].GetTag("l7_proto"); got != "http2" {
		t.Fatalf("unexpected l7_proto tag: %q", got)
	}
	kvs := pts[0].KVs()
	assertKVString(t, kvs, "trace_provider", "datadog")
	assertKVString(t, kvs, "http_user_agent", "grpc-go/1.60")
	assertKVString(t, kvs, "http_req_content_type", "application/grpc")
	assertKVString(t, kvs, "http_resp_content_type", "application/grpc")
	assertKVInt(t, kvs, "http_req_content_length", 0)
	assertKVString(t, kvs, "grpc_status", "0")
	assertKVString(t, kvs, "grpc_message", "ok")
	if got := kvs.Get("http2_stream_id"); got == nil || got.GetU() != 1 {
		t.Fatalf("unexpected http2_stream_id field: %+v", got)
	}
}

func TestHTTP2LogShouldHandlePrefaceOnly(t *testing.T) {
	var h2log HTTP2Log
	if h2log.ShouldHandle(directionRX, _http2Magic) {
		t.Fatal("server-to-client preface must not start HTTP2 detection")
	}
	if h2log.ShouldHandle(directionTX, []byte("GET / HTTP/1.1\r\n\r\n")) {
		t.Fatal("HTTP/1 payload must not start HTTP2 detection")
	}

	var split HTTP2Log
	splitPayload := buildTestH2HeadersPayload(t, []hpack.HeaderField{
		{Name: H2HdrMethod, Value: "GET"},
		{Name: H2HdrPath, Value: "/split"},
		{Name: "traceparent", Value: "00-464d5ff3a7c6ba32626b625cc201d9ee-fb94ce3aa8b7363b-01"},
	})
	first := splitPayload[:4]
	second := splitPayload[4:]
	if !split.ShouldHandle(directionTX, first) {
		t.Fatal("expected split client preface to start HTTP2 detection")
	}
	split.Handle(directionTX, first, int64(len(first)), &PktTCPHdr{Seq: 100, TS: 1}, nil, 0, 1)
	if !split.ShouldHandle(directionTX, second) {
		t.Fatal("expected buffered preface state to keep handling")
	}
	split.Handle(directionTX, second, int64(len(second)), &PktTCPHdr{Seq: 104, TS: 2}, nil, 0, 1)
	if len(split.elems) != 1 {
		t.Fatalf("expected split HTTP2 request to be parsed, got %d elements", len(split.elems))
	}
	if got := split.elems[0].TraceID; got != "464d5ff3a7c6ba32626b625cc201d9ee" {
		t.Fatalf("unexpected split trace_id: %q", got)
	}
}

func TestConfigFuncHTTPEnablesHTTP2Log(t *testing.T) {
	oldEnableNetlog := enableNetlog
	oldEnableMetric := enabledNetMetric
	oldEnableHTTP := enableL7HTTP
	oldEnableHTTP2 := enableL7HTTP2
	oldHeaders := l7RecordedHeaderAllowlist
	defer func() {
		enableNetlog = oldEnableNetlog
		enabledNetMetric = oldEnableMetric
		enableL7HTTP = oldEnableHTTP
		enableL7HTTP2 = oldEnableHTTP2
		l7RecordedHeaderAllowlist = oldHeaders
	}()

	ConfigFunc(false, false, []string{"http"})
	if !enableL7HTTP || !enableL7HTTP2 {
		t.Fatalf("http protocol should enable HTTP/1 and HTTP/2, got http=%v http2=%v",
			enableL7HTTP, enableL7HTTP2)
	}

	ConfigFunc(false, false, []string{"http1"})
	if !enableL7HTTP || enableL7HTTP2 {
		t.Fatalf("http1 protocol should only enable HTTP/1, got http=%v http2=%v",
			enableL7HTTP, enableL7HTTP2)
	}
}

func TestConfigFuncL7LogHeaders(t *testing.T) {
	oldEnableNetlog := enableNetlog
	oldEnableMetric := enabledNetMetric
	oldEnableHTTP := enableL7HTTP
	oldEnableHTTP2 := enableL7HTTP2
	oldHeaders := l7RecordedHeaderAllowlist
	defer func() {
		enableNetlog = oldEnableNetlog
		enabledNetMetric = oldEnableMetric
		enableL7HTTP = oldEnableHTTP
		enableL7HTTP2 = oldEnableHTTP2
		l7RecordedHeaderAllowlist = oldHeaders
	}()

	ConfigFunc(true, false, []string{"http"}, []string{"x-tenant-id", "authorization", ":authority"})
	headers := recordL7LogHeaderString(nil, "X-Tenant-ID", "tenant-a")
	headers = recordL7LogHeaderString(headers, ":authority", "api.example")
	headers = recordL7LogHeaderString(headers, "Authorization", "Bearer secret")
	headers = recordL7LogHeaderString(headers, "x-request-id", "req-skipped")

	assertHeaderValue(t, headers, "x-tenant-id", "tenant-a")
	assertHeaderValue(t, headers, "host", "api.example")
	assertHeaderAbsent(t, headers, "authorization")
	assertHeaderAbsent(t, headers, "x-request-id")

	ConfigFunc(true, false, []string{"http"}, []string{"none"})
	headers = recordL7LogHeaderString(nil, "host", "api.example")
	if len(headers) != 0 {
		t.Fatalf("expected L7 log header recording to be disabled, got %+v", headers)
	}

	ConfigFunc(true, false, []string{"http"})
	headers = recordL7LogHeaderString(nil, "x-request-id", "req-default")
	assertHeaderValue(t, headers, "x-request-id", "req-default")
}

func makeTestSW8(traceID, parentTraceSegmentID, parentSpanID string) string {
	return "1-" +
		base64.StdEncoding.EncodeToString([]byte(traceID)) + "-" +
		base64.StdEncoding.EncodeToString([]byte(parentTraceSegmentID)) + "-" +
		parentSpanID + "-c2VydmljZQ==-aW5zdGFuY2U=-ZW5kcG9pbnQ=-YWRkcmVzcw=="
}

func buildTestH2HeadersPayload(t *testing.T, headers []hpack.HeaderField) []byte {
	t.Helper()

	return buildTestH2HeadersFramePayload(t, headers, true)
}

func buildTestH2HeadersFramePayload(t *testing.T, headers []hpack.HeaderField, preface bool) []byte {
	t.Helper()

	var block bytes.Buffer
	enc := hpack.NewEncoder(&block)
	for _, header := range headers {
		if err := enc.WriteField(header); err != nil {
			t.Fatal(err)
		}
	}

	var payload bytes.Buffer
	if preface {
		payload.Write(_http2Magic)
	}

	framer := http2.NewFramer(&payload, nil)
	if err := framer.WriteHeaders(http2.HeadersFrameParam{
		StreamID:      1,
		BlockFragment: block.Bytes(),
		EndHeaders:    true,
	}); err != nil {
		t.Fatal(err)
	}
	return payload.Bytes()
}

func assertHeaderValue(t *testing.T, headers map[string]string, key, want string) {
	t.Helper()

	if got := headers[key]; got != want {
		t.Fatalf("unexpected header %q: got %q, want %q", key, got, want)
	}
}

func assertHeaderAbsent(t *testing.T, headers map[string]string, key string) {
	t.Helper()

	if _, ok := headers[key]; ok {
		t.Fatalf("expected header %q to be absent", key)
	}
}

func assertKVString(t *testing.T, kvs point.KVs, key, want string) {
	t.Helper()

	if got := kvs.Get(key); got == nil || got.GetS() != want {
		t.Fatalf("unexpected %s field: got %+v, want %q", key, got, want)
	}
}

func assertKVInt(t *testing.T, kvs point.KVs, key string, want int64) {
	t.Helper()

	if got := kvs.Get(key); got == nil || got.GetI() != want {
		t.Fatalf("unexpected %s field: got %+v, want %d", key, got, want)
	}
}
