// Package tracing parse http header
package tracing

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/spanid"
)

type TraceInfo struct {
	Host string

	Method  string
	Path    string
	Version string

	// param string

	ThrTraceid spanid.ID64

	ESpanType string

	TraceID      spanid.ID128
	ParentSpanID spanid.ID64
	HexEncode    bool
	HaveTracID   bool

	PidTid uint64

	ASpanSampled bool

	// TraceProvider string

	Headers map[string]string

	Param       string
	TaskComm    string
	ProcessName string
	Service     string
	AllowTrace  bool

	TS int64
}

func GetHTTPHeader(payload []byte) map[string]string {
	if payload[0] == '0' {
		return nil
	}

	idx := bytes.LastIndex(payload, []byte("\r\n\r\n"))
	if idx > 0 {
		payload = payload[:idx]
	} else if idx == 0 {
		return nil
	}

	// line 1
	idx = bytes.Index(payload, []byte("\r\n"))
	if idx < 0 {
		return nil
	}
	ln := payload[0:idx]

	req := strings.Split(string(ln), " ")
	if len(req) != 3 {
		return nil
	}
	uriAndParam := strings.Split(req[1], "?")

	uri := uriAndParam[0]
	switch {
	case len(uri) > 8 && (uri[:8] == "https://"):
		off := strings.Index(uri[8:], "/")
		if off == -1 {
			return nil
		}
	case len(uri) > 7 && (uri[:7] == "http://"):
		off := strings.Index(uri[7:], "/")
		if off == -1 {
			return nil
		}
	case (len(uri) > 0) && (uri[:1] == "/"):
	default:
		return nil
	}

	headers := map[string]string{}
	payload = payload[idx+2:]
	hdr := strings.Split(string(payload), "\r\n")
	for _, v := range hdr {
		kv := strings.SplitN(v, ":", 2)
		if len(kv) != 2 {
			break
		}
		if _, ok := headers[kv[0]]; !ok {
			headers[kv[0]] = strings.TrimSpace(kv[1])
		}
	}

	return headers
}

func GetTraceInfo(headers map[string]string) (sampled bool, hexEnc bool,
	traceID spanid.ID128, parentID spanid.ID64,
) {
	if tid, ok := headers["x-datadog-trace-id"]; ok {
		traceID.Low = uint64(DecTraceOrSpanid2ID64(tid))
		if psid, ok := headers["x-datadog-parent-id"]; ok {
			parentID = DecTraceOrSpanid2ID64(psid)
		}
		if v, ok := headers["x-datadog-sampling-priority"]; ok {
			sampled = SampledDataDog(v)
		}
		hexEnc = false
	} else if v, ok := headers["traceparent"]; ok {
		traceParent := strings.Split(v, "-")
		if len(traceParent) == 4 {
			sampled = SampledW3C(traceParent[3])
			traceID = HexTraceid2ID128(traceParent[1])
			parentID = HexSpanid2ID64(traceParent[2])
			hexEnc = true
		}
	}

	return
}

func FormatSpanID(i uint64, base16 bool) string {
	if base16 {
		_id := make([]byte, 8)
		binary.BigEndian.PutUint64(_id, i)
		return hex.EncodeToString(_id)
	} else {
		return strconv.FormatUint(i, 10)
	}
}

func HexTraceid2ID128(s string) spanid.ID128 {
	if b, _ := hex.DecodeString(s); len(b) == 16 {
		return spanid.ID128{
			Low:  binary.BigEndian.Uint64(b[8:]),
			High: binary.BigEndian.Uint64(b[:8]),
		}
	}
	return spanid.ID128{}
}

func DecTraceOrSpanid2ID64(s string) spanid.ID64 {
	if strings.HasPrefix(s, "-") {
		id, _ := strconv.ParseInt(s, 10, 64)
		return spanid.ID64(id)
	} else {
		id, _ := strconv.ParseUint(s, 10, 64)
		return spanid.ID64(id)
	}
}

func HexSpanid2ID64(s string) spanid.ID64 {
	if b, _ := hex.DecodeString(s); len(b) == 8 {
		return spanid.ID64(binary.BigEndian.Uint64(b))
	} else {
		return 0
	}
}
