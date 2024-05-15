//go:build linux
// +build linux

package l4log

import (
	"bytes"
	"errors"
	"io"
	"strconv"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

type HTTP2LogElem struct {
	streamid uint32

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

	// HTTPVersion string
	// ReqHeaders map[string][]string `json:"req_headers"`
	// RespHeaders map[string][]string `json:"resp_headers"`

	// URL
	Path  string `json:"path"`
	Param string `json:"param"`

	Method string `json:"method"`

	// response
	StatusCode int `json:"status_code"`

	hState    int8 // 1: req, 2: resp
	hFinished bool
}

type HTTP2Log struct {
	elems   []*HTTP2LogElem
	isGRPC  bool
	isHTTP2 bool
	h2dec   *HTTP2Decoder
}

func (h2log *HTTP2Log) GetElem(streamid uint32) *HTTP2LogElem {
	for _, v := range h2log.elems {
		if v.streamid == streamid {
			return v
		}
	}
	elem := &HTTP2LogElem{
		streamid: streamid,
	}

	h2log.elems = append(h2log.elems, elem)
	return elem
}

var ProtoAllowGRPC = true

func NewH2Log() *HTTP2Log {
	return &HTTP2Log{
		h2dec: NewH2Dec(),
	}
}

func (h2log *HTTP2Log) Handle(txrx int8, cnt []byte,
	cntSize int64, ln *PktTCPHdr, k *PMeta, pktState int8, chunkid int64,
) {
	if h2log.isGRPC && !ProtoAllowGRPC {
		return
	}

	if h2log.h2dec == nil {
		return
	}

	if !h2log.isHTTP2 {
		if v := HasHTTP2Magic(cnt); v > 0 {
			cnt = cnt[v:]
			h2log.isHTTP2 = true
		} else {
			return
		}
	}

	if len(cnt) == 0 {
		return
	}

	frames, err := h2log.h2dec.Decode(txrx == directionTX, cnt)
	if err != nil {
		log.Debug(err)
		return
	}

	seqOffset := ln.Seq
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

		if pktState == 1 {
			switch txrx {
			case directionRX:
				elem.rxRetransmits++
			case directionTX:
				elem.txRetransmits++
			}
		}

		switch fr := fr.(type) {
		case *H2HeaderFrame:
			for _, hdr := range fr.Headers {
				switch hdr.Name {
				case H2HdrMethod:
					elem.Method = hdr.Value
					elem.reqSeq = curSeqOffset
					elem.hState = 1

					switch txrx {
					case directionRX:
						elem.Direction = DIncoming
						if elem.rxFirstByteTS == 0 {
							elem.rxFirstByteTS = ln.TS
						}
					case directionTX:
						elem.Direction = DOutging
						if elem.txFirstByteTS == 0 {
							elem.txFirstByteTS = ln.TS
						}
					}

				case H2HdrPath:
					elem.Path = hdr.Value
				case H2HdrScheme:
				case H2HdrHost:
				case H2HdrStatus:
					v, _ := strconv.ParseInt(hdr.Value, 10, 32)
					elem.StatusCode = int(v)
					elem.respSeq = curSeqOffset
					elem.hState = 2
					switch txrx {
					case directionRX:
						if elem.rxFirstByteTS == 0 {
							elem.rxFirstByteTS = ln.TS
						}
					case directionTX:
						if elem.txFirstByteTS == 0 {
							elem.txFirstByteTS = ln.TS
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
					st, _ := strconv.ParseInt(hdr.Value, 10, 32)
					elem.grpcStatus = int(st)
				case "grpc-message":
					elem.grpcMessage = hdr.Value
				default:
					// pass
				}
			}

			switch txrx {
			case directionRX:
				elem.rxBytes += int64(frLen)
				elem.rxLastByteTS = ln.TS
			case directionTX:
				elem.txBytes += int64(frLen)
				elem.txLastByteTS = ln.TS
			}

		case *H2DataFrame:
			switch txrx {
			case directionRX:
				elem.rxBytes += int64(frLen)
				elem.rxLastByteTS = ln.TS
			case directionTX:
				elem.txBytes += int64(frLen)
				elem.txLastByteTS = ln.TS
			}
		}

		if elem.hState == 2 && frHdr.Flags.Has(http2.FlagDataEndStream) {
			elem.hFinished = true
		}
	}
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
