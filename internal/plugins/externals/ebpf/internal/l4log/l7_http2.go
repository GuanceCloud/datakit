//go:build linux
// +build linux

package l4log

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/exporter"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

type HTTP2LogElem struct {
	streamid uint32

	StreamID uint32 `json:"stream_id"`

	Direction string `json:"direction"`
	// tcp seq

	ChunkRange [2]int64 `json:"pkt_chunk_range"`

	reqSeq  uint32
	respSeq uint32

	grpcStatus  int
	grpcMessage string

	// fist packet arrive time
	// ReqTS  int64 `json:"req_first_arrive_ts"`
	// RespTS int64 `json:"resp_first_arrive_ts"`

	txFirstByteTS int64
	rxFirstByteTS int64

	txLastByteTS int64

	rxLastByteTS int64

	// tcp packets
	// txPkts int64
	// rxPkts int64

	txBytes int64
	rxBytes int64

	txRetransmits int
	rxRetransmits int

	// req/resp content size (tcp payload <http>)
	// Send int64 `json:"send_bytes"`
	// Recv int64 `json:"recv_bytes"`

	TraceID  string `json:"trace_id"`
	ParentID string `json:"parent_id"`

	TraceProvider string `json:"trace_provider,omitempty"`

	// HTTPVersion string
	ReqHeaders  map[string]string `json:"req_headers,omitempty"`
	RespHeaders map[string]string `json:"resp_headers,omitempty"`
	HeaderCount int               `json:"header_count,omitempty"`
	HeaderBytes int               `json:"header_bytes,omitempty"`

	UserAgent         string `json:"user_agent,omitempty"`
	ReqContentType    string `json:"req_content_type,omitempty"`
	RespContentType   string `json:"resp_content_type,omitempty"`
	ReqContentLength  *int64 `json:"req_content_length,omitempty"`
	RespContentLength *int64 `json:"resp_content_length,omitempty"`
	GRPCStatus        string `json:"grpc_status,omitempty"`
	GRPCMessage       string `json:"grpc_message,omitempty"`

	// URL
	Path  string `json:"path"`
	Host  string `json:"host,omitempty"`
	Param string `json:"param"`

	Method string `json:"method"`

	// response
	StatusCode int `json:"status_code"`

	hState    int8 // 1: req, 2: resp
	hFinished bool

	messageCache string
	messageDirty bool
}

type HTTP2Log struct {
	elems   []*HTTP2LogElem
	isGRPC  bool
	isHTTP2 bool
	h2dec   *HTTP2Decoder

	flow        *tcpFlowTracker
	txPreface   http2PrefaceState
	streamLimit int

	probePackets   uint8
	probeBytes     int
	probeExhausted bool
}

func (h2log *HTTP2Log) GetElem(streamid uint32) *HTTP2LogElem {
	for _, v := range h2log.elems {
		if v.streamid == streamid {
			return v
		}
	}
	if h2log.streamLimit <= 0 {
		h2log.streamLimit = http2StreamLimit()
	}
	if h2log.streamLimit > 0 && len(h2log.elems) >= h2log.streamLimit {
		h2log.trimFinishedElems()
		if len(h2log.elems) >= h2log.streamLimit {
			exporter.IncBPFEventDrop("l4log", "http2_stream", "limit")
			return nil
		}
	}
	elem := &HTTP2LogElem{
		streamid:     streamid,
		StreamID:     streamid,
		messageDirty: true,
	}

	h2log.elems = append(h2log.elems, elem)
	return elem
}

var ProtoAllowGRPC = true

const (
	http2ProbePacketBudget = 8
	http2ProbeByteBudget   = 4 * 1024

	http2StreamLimitEnv     = "DK_EBPF_L4LOG_HTTP2_STREAM_LIMIT"
	defaultHTTP2StreamLimit = 1024
	maxHTTP2StreamLimit     = 65536
)

func NewH2Log() *HTTP2Log {
	return &HTTP2Log{
		h2dec:       NewH2Dec(),
		flow:        newTCPFlowTracker(8, 64*1024),
		streamLimit: http2StreamLimit(),
	}
}

func http2StreamLimit() int {
	raw := strings.TrimSpace(os.Getenv(http2StreamLimitEnv))
	if raw == "" {
		return defaultHTTP2StreamLimit
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit <= 0 {
		log.Warnf("invalid %s=%q, use default %d", http2StreamLimitEnv, raw, defaultHTTP2StreamLimit)
		return defaultHTTP2StreamLimit
	}
	if limit > maxHTTP2StreamLimit {
		return maxHTTP2StreamLimit
	}
	return limit
}

func (h2log *HTTP2Log) trimFinishedElems() {
	if h2log == nil || len(h2log.elems) == 0 {
		return
	}
	oldLen := len(h2log.elems)
	keep := h2log.elems[:0]
	for _, elem := range h2log.elems {
		if elem != nil && !elem.hFinished {
			keep = append(keep, elem)
		}
	}
	for i := len(keep); i < oldLen; i++ {
		h2log.elems[i] = nil
	}
	h2log.elems = keep
}

func (h2log *HTTP2Log) ShouldHandle(txrx int8, cnt []byte) bool {
	if len(cnt) == 0 {
		return false
	}

	if h2log.isHTTP2 || len(h2log.txPreface.buf) > 0 || len(h2log.elems) > 0 {
		return true
	}

	if h2log.probeExhausted || txrx != directionTX {
		return false
	}

	if maybeHTTP2PrefaceStart(cnt) {
		return true
	}

	h2log.probePackets++
	h2log.probeBytes += len(cnt)
	if h2log.probePackets >= http2ProbePacketBudget || h2log.probeBytes >= http2ProbeByteBudget {
		h2log.probeExhausted = true
	}
	return false
}

func (h2log *HTTP2Log) Handle(txrx int8, cnt []byte,
	cntSize int64, ln *PktTCPHdr, k *PMeta, pktState int8, chunkid int64,
) {
	if h2log.isGRPC && !ProtoAllowGRPC {
		return
	}

	if h2log.h2dec == nil {
		h2log.h2dec = NewH2Dec()
	}

	if h2log.flow == nil {
		h2log.flow = newTCPFlowTracker(8, 64*1024)
	}

	res := h2log.flow.Push(txrx, ln.Seq, ln.Flags, cnt, ln.TS)
	if len(res.Deliveries) == 0 {
		return
	}

	for _, delivery := range res.Deliveries {
		payload := delivery.Payload

		if !h2log.isHTTP2 {
			if txrx != directionTX {
				continue
			}
			payload = h2log.txPreface.Feed(payload)
			if payload == nil {
				continue
			}
			h2log.isHTTP2 = true
		}

		if len(payload) == 0 {
			continue
		}

		frames, err := h2log.h2dec.Decode(txrx == directionTX, payload)
		if err != nil {
			log.Debug(err)
			continue
		}

		seqOffset := delivery.Seq
		for _, fr := range frames.Fr {
			frHdr := fr.Header()
			frLen := frHdr.Length + 9

			curSeqOffset := seqOffset
			seqOffset += frLen

			streamID := frHdr.StreamID
			if streamID == 0 {
				continue
			}

			elem := h2log.GetElem(streamID)
			if elem == nil {
				continue
			}
			elem.markMessageDirty()

			if pktState == 1 || res.Retransmit {
				switch txrx {
				case directionRX:
					elem.rxRetransmits++
				case directionTX:
					elem.txRetransmits++
				}
			}

			switch fr := fr.(type) {
			case *H2HeaderFrame:
				var traceHeaders traceHeaderCarrier
				var headers map[string]string
				var headerKind int8
				for _, hdr := range fr.Headers {
					traceHeaders.addString(hdr.Name, hdr.Value)
					if enableNetlog {
						headers = recordL7LogHeaderString(headers, hdr.Name, hdr.Value)
					}

					switch hdr.Name {
					case H2HdrMethod:
						headerKind = 1
						elem.Method = hdr.Value
						elem.reqSeq = curSeqOffset
						elem.hState = 1

						switch txrx {
						case directionRX:
							elem.Direction = DIncoming
							if elem.rxFirstByteTS == 0 {
								elem.rxFirstByteTS = delivery.TS
							}
						case directionTX:
							elem.Direction = DOutging
							if elem.txFirstByteTS == 0 {
								elem.txFirstByteTS = delivery.TS
							}
						}

					case H2HdrPath:
						elem.Path = hdr.Value
					case H2HdrScheme:
					case H2HdrHost:
						elem.Host = normalizeHTTPHostString(hdr.Value)
					case H2HdrStatus:
						headerKind = 2
						v, _ := strconv.ParseInt(hdr.Value, 10, 32)
						elem.StatusCode = int(v)
						elem.respSeq = curSeqOffset
						elem.hState = 2
						switch txrx {
						case directionRX:
							if elem.rxFirstByteTS == 0 {
								elem.rxFirstByteTS = delivery.TS
							}
						case directionTX:
							if elem.txFirstByteTS == 0 {
								elem.txFirstByteTS = delivery.TS
							}
						}
					case "content-type":
						if hdr.Value == "application/grpc" {
							h2log.isGRPC = true
							if !ProtoAllowGRPC {
								h2log.elems = nil
								return
							}
						}
					case "grpc-status":
						headerKind = 2
						st, _ := strconv.ParseInt(hdr.Value, 10, 32)
						elem.grpcStatus = int(st)
						elem.GRPCStatus = hdr.Value
					case "grpc-message":
						headerKind = 2
						elem.grpcMessage = hdr.Value
						elem.GRPCMessage = hdr.Value
					default:
						// pass
					}
				}
				if traceID, parentID, provider := traceHeaders.traceIDsWithProvider(); traceID != "" {
					elem.TraceID = traceID
					elem.ParentID = parentID
					elem.TraceProvider = provider
				}
				switch headerKind {
				case 1:
					elem.setReqHeaders(mergeL7LogHeaders(elem.ReqHeaders, headers))
				case 2:
					elem.setRespHeaders(mergeL7LogHeaders(elem.RespHeaders, headers))
				}

				switch txrx {
				case directionRX:
					elem.rxBytes += int64(frLen)
					elem.rxLastByteTS = delivery.TS
				case directionTX:
					elem.txBytes += int64(frLen)
					elem.txLastByteTS = delivery.TS
				}

			case *H2DataFrame:
				switch txrx {
				case directionRX:
					elem.rxBytes += int64(frLen)
					elem.rxLastByteTS = delivery.TS
				case directionTX:
					elem.txBytes += int64(frLen)
					elem.txLastByteTS = delivery.TS
				}
			}

			if elem.hState == 2 && frHdr.Flags.Has(http2.FlagDataEndStream) {
				elem.hFinished = true
			}
		}
	}
}

func (h *HTTP2LogElem) markMessageDirty() {
	h.messageDirty = true
}

func (h *HTTP2LogElem) setReqHeaders(headers map[string]string) {
	h.ReqHeaders = headers
	h.UserAgent = l7HeaderValue(headers, "user-agent")
	h.ReqContentType = l7HeaderValue(headers, "content-type")
	h.ReqContentLength = l7ContentLength(headers)
	h.updateHeaderSummary()
}

func (h *HTTP2LogElem) setRespHeaders(headers map[string]string) {
	h.RespHeaders = headers
	h.RespContentType = l7HeaderValue(headers, "content-type")
	h.RespContentLength = l7ContentLength(headers)
	if status := l7HeaderValue(headers, "grpc-status"); status != "" {
		h.GRPCStatus = status
	}
	if msg := l7HeaderValue(headers, "grpc-message"); msg != "" {
		h.GRPCMessage = msg
	}
	h.updateHeaderSummary()
}

func (h *HTTP2LogElem) updateHeaderSummary() {
	h.HeaderCount, h.HeaderBytes = l7LogHeadersSummary(h.ReqHeaders, h.RespHeaders)
}

func maybeHTTP2PrefaceStart(payload []byte) bool {
	if len(payload) == 0 {
		return false
	}
	if len(payload) >= len(_http2Magic) {
		return bytes.HasPrefix(payload, _http2Magic)
	}
	return bytes.Equal(payload, _http2Magic[:len(payload)])
}

type http2PrefaceState struct {
	buf []byte
}

func (p *http2PrefaceState) Feed(payload []byte) []byte {
	if len(p.buf) == 0 && HasHTTP2Magic(payload) > 0 {
		return payload[len(_http2Magic):]
	}

	p.buf = append(p.buf, payload...)
	if len(p.buf) < len(_http2Magic) {
		if bytes.Equal(p.buf, _http2Magic[:len(p.buf)]) {
			return nil
		}
		p.buf = p.buf[:0]
		return nil
	}

	if !bytes.Equal(p.buf[:len(_http2Magic)], _http2Magic) {
		p.buf = p.buf[:0]
		return nil
	}

	out := append([]byte(nil), p.buf[len(_http2Magic):]...)
	p.buf = p.buf[:0]
	return out
}

type HTTP2Decoder struct {
	enableH2BodyData bool

	h2Framer   *http2.Framer
	hpackTxDec *hpack.Decoder
	hpackRxDec *hpack.Decoder

	Reader bytes.Reader
}

func NewH2Dec() *HTTP2Decoder {
	h2dec := HTTP2Decoder{}

	h2dec.h2Framer = http2.NewFramer(nil, &h2dec.Reader)
	h2dec.hpackTxDec = hpack.NewDecoder(4096, nil)
	h2dec.hpackRxDec = hpack.NewDecoder(4096, nil)
	return &h2dec
}

type Frame interface {
	Header() http2.FrameHeader
}

type H2Frames struct {
	Fr []Frame
}

var (
	_ Frame = (*H2HeaderFrame)(nil)
	_ Frame = (*H2DataFrame)(nil)
	_ Frame = (*H2RSTStreamFrame)(nil)
	_ Frame = (*H2OtherFrame)(nil)
)

type H2HeaderFrame struct {
	http2.FrameHeader

	Headers []hpack.HeaderField
}

type H2OtherFrame struct {
	http2.FrameHeader
}

func (f *H2HeaderFrame) HeadersEnded() bool {
	return f.FrameHeader.Flags.Has(
		http2.FlagHeadersEndHeaders)
}

func (f *H2HeaderFrame) StreamEnded() bool {
	return f.FrameHeader.Flags.Has(
		http2.FlagDataEndStream)
}

type H2DataFrame struct {
	http2.FrameHeader

	buf []byte
}

func (f *H2DataFrame) StreamEnded() bool {
	return f.FrameHeader.Flags.Has(
		http2.FlagDataEndStream)
}

type H2RSTStreamFrame struct {
	http2.FrameHeader
	ERRCode http2.ErrCode
}

const (
	H2HdrMethod = ":method"
	H2HdrPath   = ":path"
	H2HdrScheme = ":scheme"
	// https://www.rfc-editor.org/rfc/rfc3986.html#section-3.2
	// authority   = [ userinfo "@" ] host [ ":" port ].
	H2HdrHost   = ":authority"
	H2HdrStatus = ":status"
)

var _http2Magic = []byte("\x50\x52\x49\x20\x2a\x20\x48\x54\x54\x50\x2f\x32\x2e\x30\x0d\x0a" +
	"\x0d\x0a\x53\x4d\x0d\x0a\x0d\x0a")

const H2MagicStr = "\x50\x52\x49\x20\x2a\x20\x48\x54\x54\x50\x2f\x32\x2e\x30\x0d\x0a" +
	"\x0d\x0a\x53\x4d\x0d\x0a\x0d\x0a"

func HasHTTP2Magic(buf []byte) int {
	lenM := len(_http2Magic)
	if len(buf) < lenM {
		return 0
	}

	if bytes.HasPrefix(buf, _http2Magic) {
		return lenM
	}

	return 0
}

func (h2dec *HTTP2Decoder) Decode(tx bool, buf []byte) (*H2Frames, error) {
	h2dec.Reader.Reset(buf)

	h2Rslt := H2Frames{}

	for {
		fr, err := h2dec.h2Framer.ReadFrame()
		if err != nil {
			switch {
			case errors.Is(err, io.EOF):
			default:
				// errors.Is(err, io.ErrUnexpectedEOF):
				// errors.Is(err, http2.ErrFrameTooLarge):
				// ...
				return nil, err
			}
			break
		}

		switch fr := fr.(type) {
		case *http2.HeadersFrame:
			hdrFrg := fr.HeaderBlockFragment()

			var hdec *hpack.Decoder
			if tx {
				hdec = h2dec.hpackTxDec
			} else {
				hdec = h2dec.hpackRxDec
			}
			headers, err := hdec.DecodeFull(hdrFrg)
			if err != nil {
				return nil, err
			}

			h2hdr := &H2HeaderFrame{}
			h2hdr.FrameHeader = fr.FrameHeader
			h2hdr.Headers = headers
			h2Rslt.Fr = append(h2Rslt.Fr, h2hdr)

		case *http2.DataFrame:
			h2body := &H2DataFrame{
				FrameHeader: fr.FrameHeader,
			}

			if h2dec.enableH2BodyData {
				// copy data
				h2body.buf = fr.Data()
			}

			fr.StreamEnded()
			h2Rslt.Fr = append(h2Rslt.Fr, h2body)

		case *http2.RSTStreamFrame:
			h2rst := &H2RSTStreamFrame{
				FrameHeader: fr.FrameHeader,
				ERRCode:     fr.ErrCode,
			}
			h2Rslt.Fr = append(h2Rslt.Fr, h2rst)
		default:
			h2Rslt.Fr = append(h2Rslt.Fr, &H2OtherFrame{
				FrameHeader: fr.Header(),
			})
		}
	}

	return &h2Rslt, nil
}
