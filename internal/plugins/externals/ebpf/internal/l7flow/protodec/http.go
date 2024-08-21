//go:build linux
// +build linux

package protodec

import (
	"bytes"
	"regexp"
	"strconv"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/tracing"
)

var _ ProtoDecPipe = (*httpDecPipe)(nil)

var (
	// skip http method: CONNECT, TRACE.
	RequestPrefixHTTP  = regexp.MustCompile(`^(?:GET|PUT|POST|HEAD|PATCH|DELETE|OPTIONS) `)
	ResponsePrefixHTTP = regexp.MustCompile(`^HTTP/\d.\d \d{3}`)
)

var log = logger.SLogger("decoder")

const (
	HTTPMethodUnknown = 0x00 + iota
	HTTPMethodGET
	HTTPMethodPOST
	HTTPMethodPUT
	HTTPMethodDELETE
	HTTPMethodHEAD
	HTTPMethodOPTIONS
	HTTPMethodPATCH

	// TODO parse such HTTP data.
	HTTPMethodCONNECT
	HTTPMethodTRACE
)

func HTTPMethodInt(method int) string {
	switch method {
	case HTTPMethodGET:
		return "GET"
	case HTTPMethodPOST:
		return "POST"
	case HTTPMethodPUT:
		return "PUT"
	case HTTPMethodDELETE:
		return "DELETE"
	case HTTPMethodHEAD:
		return "HEAD"
	case HTTPMethodOPTIONS:
		return "OPTIONS"
	case HTTPMethodPATCH:
		return "PATCH"
	default:
		return ""
	}
}

// func HTTPMethodString(method string) int {
// 	switch method {
// 	case "GET":
// 		return HTTPMethodGET
// 	case "POST":
// 		return HTTPMethodPOST
// 	case "PUT":
// 		return HTTPMethodPUT
// 	case "DELETE":
// 		return HTTPMethodDELETE
// 	case "HEAD":
// 		return HTTPMethodHEAD
// 	case "OPTIONS":
// 		return HTTPMethodOPTIONS
// 	case "PATCH":
// 		return HTTPMethodPATCH
// 	default:
// 		return HTTPMethodUnknown
// 	}
// }

func FindHTTPURI(payload []byte) string {
	idx := bytes.Index(payload, []byte(" "))

	if idx == -1 || idx+1 >= len(payload) {
		return ""
	}

	uri := payload[idx+1:]

	idx2 := bytes.Index(uri, []byte(" "))
	if idx2 != -1 {
		uri = uri[:idx2]
	}

	startOffset := -1
	switch {
	case len(uri) > 8 && (string(uri[:8]) == "https://"):
		off := bytes.Index(uri[8:], []byte("/"))
		if off == -1 {
			return "/"
		}
		startOffset = 8 + off
	case len(uri) > 7 && (string(uri[:7]) == "http://"):
		off := bytes.Index(uri[7:], []byte("/"))
		if off == -1 {
			return "/"
		}
		startOffset = 7 + off
	case (len(uri) > 0) && (string(uri[:1]) == "/"):
		startOffset = 0
	}

	if startOffset == -1 {
		return ""
	}

	endOffset := bytes.Index(uri, []byte("?"))
	if endOffset > 0 && startOffset < endOffset {
		return string(uri[startOffset:endOffset])
	}
	return string(uri[startOffset:])
}

func HTTPProtoDetect(data []byte, actSize int) (L7Protocol, ProtoDecPipe, bool) {
	if RequestPrefixHTTP.Match(data) {
		return ProtoHTTP, newHTTPDecPipe(ProtoHTTP), true
	}

	return ProtoUnknown, nil, false
}

type httpInfo struct {
	meta ProtoMeta

	method string
	path   string

	httpVersion string

	statusCode int

	reqBytes  int
	respBytes int

	ktime [4]uint64
	ts    int64
	dur   [2]uint64
}

type httpDecPipe struct {
	direction  comm.Direcion
	infCache   []*httpInfo
	inf        *httpInfo
	reqResp    int // 0, 1, 2
	connClosed bool
}

func (dec *httpDecPipe) Decode(txRx comm.NICDirection, data *comm.NetwrkData,
	ts int64, thrTr threadTrace,
) {
	var (
		req, resp, ok bool
		v             string
		statusCode    int
	)
	if v, ok = httpMethod(data.Payload); ok {
		req = true
	} else if v, statusCode, ok = httpProtoVersion(data.Payload); ok {
		resp = true
	}

	if dec.inf == nil {
		dec.inf = &httpInfo{}
		dec.reqResp = 0
	}
	inf := dec.inf

	switch {
	case req:
		url := FindHTTPURI(data.Payload)
		if url == "" {
			break
		}

		if dec.reqResp == 2 { // switch to next req-resp
			dec.infCache = append(dec.infCache, inf)
			inf = &httpInfo{}
			dec.inf = inf
		}
		dec.reqResp = 1
		inf.method = v
		inf.path = url
		inf.meta.Threads[0] = data.Thread
		inf.ts = ts

		if dec.direction == comm.DUnknown {
			switch txRx { //nolint:exhaustive
			case comm.NICDIngress:
				dec.direction = comm.DIn
			case comm.NICDEgress:
				dec.direction = comm.DOut
			}
		}

		if dec.direction == comm.DIn {
			inf.meta.InnerID = thrTr.Insert(dec.direction, int32(data.Conn.Pid), data.Thread, data.TSTail)
		}

		inf.meta.ReqTCPSeq = data.TCPSeq
		inf.dur[0] = data.TS
		inf.ktime[0] = data.TSTail

		headers := tracing.GetHTTPHeader(data.Payload)
		inf.meta.SampledSpan,
			inf.meta.SpanHexEnc,
			inf.meta.TraceID,
			inf.meta.ParentSpanID = tracing.GetTraceInfo(headers)

	case resp:
		dec.reqResp = 2
		inf.meta.RespTCPSeq = data.TCPSeq
		inf.meta.Threads[1] = data.Thread

		inf.statusCode = statusCode
		inf.httpVersion = v

		if dec.direction == comm.DUnknown {
			switch txRx { //nolint:exhaustive
			case comm.NICDIngress:
				dec.direction = comm.DOut
			case comm.NICDEgress:
				dec.direction = comm.DIn
			}
		}
		inf.dur[1] = data.TS
		inf.ktime[2] = data.TSTail
	}

	switch dec.direction { //nolint:exhaustive
	case comm.DIn:
		switch txRx { //nolint:exhaustive
		case comm.NICDIngress:
			inf.reqBytes += data.CaptureSize
		case comm.NICDEgress:
			inf.respBytes += data.CaptureSize
		}
	case comm.DOut:
		switch txRx { //nolint:exhaustive
		case comm.NICDIngress:
			inf.respBytes += data.CaptureSize
		case comm.NICDEgress:
			inf.reqBytes += data.CaptureSize
		}
	}

	switch dec.reqResp {
	case 1:
		inf.ktime[1] = data.TSTail
	case 2:
		inf.ktime[3] = data.TSTail
	}
}

func (dec *httpDecPipe) Proto() L7Protocol {
	return ProtoHTTP
}

func (dec *httpDecPipe) Export(force bool) []*ProtoData {
	if force {
		dec.infCache = append(dec.infCache, dec.inf)
		dec.inf = nil
	}

	var result []*ProtoData

	for _, inf := range dec.infCache {
		if inf == nil {
			continue
		}

		kvs := make(point.KVs, 0, 20)

		// 这几个字段需要与聚合函数的字段相同
		switch dec.direction { //nolint:exhaustive
		case comm.DIn:
			kvs = kvs.Add(comm.FieldBytesRead, int64(inf.reqBytes), false, true)
			kvs = kvs.Add(comm.FieldBytesWritten, int64(inf.respBytes), false, true)
		default:
			kvs = kvs.Add(comm.FieldBytesRead, int64(inf.respBytes), false, true)
			kvs = kvs.Add(comm.FieldBytesWritten, int64(inf.reqBytes), false, true)
		}
		kvs = kvs.Add(comm.FieldHTTPMethod, inf.method, false, true)
		kvs = kvs.Add(comm.FieldHTTPRoute, inf.path, false, true)
		kvs = kvs.Add(comm.FieldHTTPVersion, inf.httpVersion, false, true)
		kvs = kvs.Add(comm.FieldHTTPStatusCode, strconv.Itoa(inf.statusCode), false, true)
		kvs = kvs.Add(comm.FieldStatus, httpCode2Status(inf.statusCode), false, true)

		// 页面 span 上显示的是 `<opperation> <resource>`
		kvs = kvs.Add(comm.FieldOperation, ProtoHTTP.String(), false, true)
		kvs = kvs.Add(comm.FieldResource, inf.method+" "+inf.path, false, true)

		dur := int64(inf.ktime[3] - inf.ktime[0])
		cost := int64(inf.ktime[2] - inf.ktime[1])
		result = append(result, &ProtoData{
			Meta:      inf.meta,
			Time:      inf.ts,
			KVs:       kvs,
			Cost:      cost,
			Duration:  dur,
			Direction: dec.direction,
			L7Proto:   ProtoHTTP,
			KTime:     inf.ktime[1],
		})
	}

	dec.infCache = dec.infCache[:0]

	return result
}

func (dec *httpDecPipe) ConnClose() {
	dec.connClosed = true
}

func newHTTPDecPipe(L7Protocol) ProtoDecPipe {
	return &httpDecPipe{}
}

func httpMethod(d []byte) (string, bool) {
	result := RequestPrefixHTTP.Find(d)
	if len(result) == 0 {
		return "", false
	}
	return string(result[0 : len(result)-1]), true
}

func httpProtoVersion(d []byte) (string, int, bool) {
	result := ResponsePrefixHTTP.Find(d)
	if len(result) != 12 {
		return "", 0, false
	}
	i, err := strconv.ParseInt(string(result[9:12]), 10, 32)
	if err != nil {
		return "", 0, false
	}
	return string(result[5:8]), int(i), true
}

func httpCode2Status(code int) string {
	switch {
	case code < 400:
		return "ok"
	case code >= 400 && code < 500:
		return "warning"
	case code >= 500:
		return "error"
	default:
		return ""
	}
}
