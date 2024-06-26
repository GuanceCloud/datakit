//go:build linux
// +build linux

package protodec

import (
	"strconv"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l4log"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/tracing"
	"golang.org/x/net/http2"
)

type h2Info struct {
	streamid uint32

	meta ProtoMeta

	method string
	path   string

	statusCode int

	reqBytes  int
	respBytes int

	ktime [4]uint64
	dur   [2]uint64

	grpcStatusCode int
	grpcMessage    string
	reqResp        int8
	hFinished      bool
	ts             int64
}

type h2DecPipe struct {
	direction  comm.Direcion
	elems      []*h2Info
	dec        *l4log.HTTP2Decoder
	proto      L7Protocol
	isGRPC     bool
	connClosed bool
}

func H2ProtoDetect(data []byte, actSize int) (L7Protocol, ProtoDecPipe, bool) {
	if v := l4log.HasHTTP2Magic(data); v > 0 {
		return ProtoHTTP2, newH2DecPipe(ProtoHTTP2), true
	}
	return ProtoUnknown, nil, false
}

func (dec *h2DecPipe) GetElem(streamid uint32) *h2Info {
	for _, v := range dec.elems {
		if v.streamid == streamid {
			return v
		}
	}
	elem := &h2Info{
		streamid: streamid,
	}

	dec.elems = append(dec.elems, elem)
	return elem
}

func (dec *h2DecPipe) Decode(txRx comm.NICDirection, data *comm.NetwrkData,
	ts int64, thrTr threadTrace,
) {
	if len(data.Payload) == 0 {
		return
	}

	frames, err := dec.dec.Decode(txRx == comm.NICDEgress, data.Payload)
	if err != nil {
		log.Debug(err)
		return
	}

	seqOffset := data.TCPSeq

	var elem *h2Info
	var currentStreamID uint32
	var tsSetted bool
	for _, fr := range frames.Fr {
		frHdr := fr.Header()
		frLen := frHdr.Length + 9
		frBytesCount := 0

		curSeqOffset := seqOffset
		seqOffset += frLen

		if frHdr.StreamID == 0 {
			continue
		}

		if elem == nil || currentStreamID != frHdr.StreamID {
			currentStreamID = frHdr.StreamID
			elem = dec.GetElem(currentStreamID)
			tsSetted = false
		}

		switch fr := fr.(type) {
		case *l4log.H2HeaderFrame:
			for _, hdr := range fr.Headers {
				switch hdr.Name {
				case l4log.H2HdrMethod:
					elem.method = hdr.Value
					elem.meta.ReqTCPSeq = curSeqOffset
					elem.reqResp = 1
					elem.ktime[0] = data.TSTail
					elem.dur[0] = data.TS
					elem.ts = ts
					elem.meta.Threads[0] = data.Thread

					if dec.direction == comm.DUnknown {
						switch txRx { //nolint:exhaustive
						case comm.NICDIngress:
							dec.direction = comm.DIn
						case comm.NICDEgress:
							dec.direction = comm.DOut
						}
					}

					if dec.direction == comm.DIn {
						elem.meta.InnerID = thrTr.Insert(dec.direction, data.Thread, data.TSTail)
					}

				case l4log.H2HdrPath:
					elem.path = hdr.Value
				case l4log.H2HdrScheme:
				case l4log.H2HdrHost:
				case l4log.H2HdrStatus:
					v, _ := strconv.ParseInt(hdr.Value, 10, 32)
					elem.statusCode = int(v)
					elem.meta.RespTCPSeq = curSeqOffset
					elem.reqResp = 2
					elem.ktime[2] = data.TSTail
					elem.dur[1] = data.TS
					elem.meta.Threads[1] = data.Thread

					if dec.direction == comm.DUnknown {
						switch txRx { //nolint:exhaustive
						case comm.NICDIngress:
							dec.direction = comm.DOut
						case comm.NICDEgress:
							dec.direction = comm.DIn
						}
					}
				case "content-type":
					if hdr.Value == "application/grpc" {
						dec.isGRPC = true
					}
				case "grpc-status":
					st, _ := strconv.ParseInt(hdr.Value, 10, 32)
					elem.grpcStatusCode = int(st)
				case "grpc-message":
					elem.grpcMessage = hdr.Value
				case "x-datadog-trace-id":
					elem.meta.TraceID.Low = uint64(tracing.DecTraceOrSpanid2ID64(hdr.Value))
					elem.meta.SpanHexEnc = false
				case "x-datadog-span-id":
					elem.meta.ParentSpanID = tracing.DecTraceOrSpanid2ID64(hdr.Value)
				case "x-datadog-sampling-priority":
					elem.meta.SampledSpan = tracing.SampledDataDog(hdr.Value)
				case "traceparent":
					if elem.meta.TraceID.Zero() && elem.meta.ParentSpanID.Zero() {
						elem.meta.SpanHexEnc = true
						traceParent := strings.Split(hdr.Value, "-")
						if len(traceParent) == 4 {
							elem.meta.SampledSpan = tracing.SampledW3C(traceParent[3])
							elem.meta.TraceID = tracing.HexTraceid2ID128(traceParent[1])
							elem.meta.ParentSpanID = tracing.HexSpanid2ID64(traceParent[2])
							elem.meta.SpanHexEnc = true
						}
					}
				default:
					// pass
				}
			}

			frBytesCount = int(frLen)
		case *l4log.H2DataFrame:
			frBytesCount = int(frLen)
		}

		if !tsSetted {
			switch elem.reqResp {
			case 1:
				elem.ktime[1] = data.TSTail
			case 2:
				elem.ktime[3] = data.TSTail
			}
		}

		switch dec.direction { //nolint:exhaustive
		case comm.DIn:
			switch txRx { //nolint:exhaustive
			case comm.NICDIngress:
				elem.reqBytes += frBytesCount
			case comm.NICDEgress:
				elem.respBytes += frBytesCount
			}
		case comm.DOut:
			switch txRx { //nolint:exhaustive
			case comm.NICDIngress:
				elem.respBytes += frBytesCount
			case comm.NICDEgress:
				elem.reqBytes += frBytesCount
			}
		}

		// FlagDataEndStream == FlagHeadersEndStream
		if elem.reqResp == 2 && frHdr.Flags.Has(http2.FlagDataEndStream) {
			elem.hFinished = true
		}
	}
}

func (dec *h2DecPipe) Export(force bool) []*ProtoData {
	var rst []*ProtoData
	var keep []*h2Info
	for _, inf := range dec.elems {
		if inf == nil {
			continue
		}
		if inf.hFinished {
			kvs := make(point.KVs, 0, 20)

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
			kvs = kvs.Add(comm.FieldHTTPStatusCode, strconv.Itoa(inf.statusCode), false, true)
			kvs = kvs.Add(comm.FieldStatus, httpCode2Status(inf.statusCode), false, true)
			var proto L7Protocol
			if dec.isGRPC {
				proto = ProtoGRPC
				if inf.grpcMessage != "" {
					kvs = kvs.Add(comm.FieldGRPCMessage, inf.grpcMessage, false, true)
				}
				kvs = kvs.Add(comm.FieldGRPCStatusCode, strconv.Itoa(inf.grpcStatusCode), false, true)
			} else {
				proto = ProtoHTTP2
			}

			// 页面 span 上显示的是 `<opperation> <resource>`
			kvs = kvs.Add(comm.FieldOperation, proto.String(), false, true)
			kvs = kvs.Add(comm.FieldResource, inf.method+" "+inf.path, false, true)

			dur := int64(inf.ktime[3] - inf.ktime[0])
			cost := int64(inf.ktime[2] - inf.ktime[1])
			rst = append(rst, &ProtoData{
				Meta:      inf.meta,
				Time:      inf.ts,
				KVs:       kvs,
				Cost:      cost,
				Duration:  dur,
				Direction: dec.direction,
				L7Proto:   proto,
				KTime:     inf.ktime[1],
			})
		} else {
			keep = append(keep, inf)
		}
	}
	dec.elems = keep
	return rst
}

func (dec *h2DecPipe) Proto() L7Protocol {
	return dec.proto
}

func (dec *h2DecPipe) ConnClose() {
	dec.connClosed = true
}

func newH2DecPipe(L7Protocol) ProtoDecPipe {
	return &h2DecPipe{
		dec: l4log.NewH2Dec(),
	}
}
