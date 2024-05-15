//go:build linux
// +build linux

package l4log

import (
	"bufio"
	"bytes"
	"net/http"
	"strings"
)

// var _ L7ProtoEventAndMetric = (*HTTPLog)(nil)

// const maxTCPHDRCountLimit = 320

type HTTPLog struct {
	elems []*HTTPLogElem
}

func (h *HTTPLog) Handle(v any, txrx int8, cnt []byte,
	cntSize int64, ln *PktTCPHdr, k *PMeta, pktState int8, chunkid int64,
) error {
	reqResp := httpReqOrResp(cnt)
	var elem *HTTPLogElem
	if len(h.elems) == 0 {
		elem = &HTTPLogElem{}
		h.elems = append(h.elems, elem)
	} else {
		elem = h.elems[len(h.elems)-1]
		if elem == nil {
			elem = &HTTPLogElem{}
			h.elems[len(h.elems)-1] = elem
		}

		// HTTP 长连接处理逻辑
		// 如果是重传报文，不切换到新的请求
		// 注意：在finish 函数执行后，且当前是 req，则将切换到新的请求
		if elem.reqSeq != ln.Seq && elem.finished(reqResp == 1) {
			// 尝试更新上一个请求的时间信息
			// such as `psh(resp)` -> `ack-psh(new req)`
			elem.recReqRespTS(txrx, 0, ln)

			elem = &HTTPLogElem{}
			h.elems = append(h.elems, elem)
		}
	}

	if elem.ChunkRange[0] == 0 || elem.ChunkRange[0] > chunkid {
		elem.ChunkRange[0] = chunkid
	}
	if elem.ChunkRange[1] < chunkid {
		elem.ChunkRange[1] = chunkid
	}

	elem.recReqRespTS(txrx, cntSize, ln)

	return elem.handle(v, txrx, cnt, cntSize, ln, k, reqResp, pktState)
}

const (
	DIncoming = "incoming"
	DOutging  = "outgoing"
)

type HTTPLogElem struct {
	Direction string `json:"direction"`
	// tcp seq

	ChunkRange [2]int64 `json:"pkt_chunk_range"`

	reqSeq  uint32
	respSeq uint32

	// fist packet arrive time
	// ReqTS  int64 `json:"req_first_arrive_ts"`
	// RespTS int64 `json:"resp_first_arrive_ts"`

	txFirstByteTS int64
	rxFirstByteTS int64

	txNxtAckSeq uint32
	txAcked     bool

	rxNxtAckSeq uint32
	rxAcked     bool
	// 如果是 tx 则通过 rx 的 ack 或 tx 含 payload 的包的时间计算
	txLastByteTS int64

	rxLastByteTS int64

	// tcp packets
	txPkts  int64
	rxPkts  int64
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

func httpReqOrResp(cnt []byte) int8 { // 1: req, 2: resp
	if s := bytes.Index(cnt, []byte{'\r', '\n'}); s > 0 {
		buf := cnt[:s]
		hinfo := bytes.Split(buf, []byte{' '})
		if len(hinfo) < 2 {
			return 0
		}

		switch string(hinfo[0]) {
		case "GET":
		case "HEAD":
		case "POST":
		case "PUT":
		case "DELETE":
		case "CONNECT":
		case "OPTIONS":
		case "PATCH":
		case "TRACE":
		default:
			if len(hinfo[0]) > 5 &&
				string(hinfo[0][:5]) == "HTTP/" {
				return 2
			}
			return 0
		}

		return 1
	}

	return 0
}

func (h *HTTPLog) DetectProto(cnt []byte) bool {
	return httpReqOrResp(cnt) > 0
}

func (h *HTTPLogElem) recReqRespTS(txrx int8, cntSize int64, ln *PktTCPHdr) {
	switch txrx {
	case directionTX:
		if cntSize > 0 {
			if h.txFirstByteTS == 0 {
				h.txFirstByteTS = ln.TS
			}
			h.txLastByteTS = ln.TS
			h.txNxtAckSeq = ln.Seq + uint32(cntSize)
			h.txAcked = false
		} else if !h.rxAcked && h.rxNxtAckSeq == ln.AckSeq {
			h.rxLastByteTS = ln.TS
			h.rxAcked = true
		}
	case directionRX:
		if cntSize > 0 {
			if h.rxFirstByteTS == 0 {
				h.rxFirstByteTS = ln.TS
			}
			h.rxNxtAckSeq = ln.Seq + uint32(cntSize)
			h.rxLastByteTS = ln.TS
			h.rxAcked = false
		} else if !h.txAcked && h.txNxtAckSeq == ln.AckSeq {
			h.txLastByteTS = ln.TS
			h.txAcked = true
		}
	}
}

func (h *HTTPLogElem) handle(v any, txrx int8, cnt []byte, cntSize int64,
	ln *PktTCPHdr, k *PMeta, reqResp int8, pktState int8,
) error {
	if pktState == 1 {
		switch txrx {
		case directionRX:
			h.rxRetransmits++
		case directionTX:
			h.txRetransmits++
		}
	}

	switch reqResp {
	case 1: // req
		reader := bufio.NewReader(bytes.NewReader(cnt))
		if req, err := http.ReadRequest(reader); err == nil {
			switch txrx {
			case directionRX:
				h.Direction = DIncoming

				h.rxPkts++
				h.rxBytes += cntSize
			case directionTX:
				h.Direction = DOutging

				h.txPkts++
				h.txBytes += cntSize
			}

			h.reqSeq = ln.Seq

			h.hState = 1
			// h.ReqTS = ln.TS

			h.Method = req.Method
			h.Param = req.URL.RawQuery
			h.Path = req.URL.Path
			// h.ReqHeaders = req.Header
			for k, v := range req.Header {
				if strings.ToLower(k) == "traceparent" {
					if len(v) > 0 {
						v := strings.Split(v[0], "-")
						if len(v) == 4 {
							h.TraceID = v[1]
							h.ParentID = v[2]
						}
					}
				}
			}

			_ = req.Body.Close()
			return nil
		} else {
			log.Error("parse http req", err)
			h.StatusCode = -1
		}
	case 2: // resp
		reader := bufio.NewReader(bytes.NewReader(cnt))
		if resp, err := http.ReadResponse(reader, nil); err == nil {
			switch txrx {
			case directionRX:
				h.Direction = DOutging

				h.rxPkts++
				h.rxBytes += cntSize
			case directionTX:
				h.Direction = DIncoming

				h.txPkts++
				h.txBytes += cntSize
			}

			h.respSeq = ln.Seq

			h.hState = 2
			// h.RespTS = ln.TS
			h.StatusCode = resp.StatusCode
			// h.RespHeaders = resp.Header

			_ = resp.Body.Close()
			return nil
		} else {
			log.Error("parse http resp", err)
			h.StatusCode = -2
		}

	default:
		if h.hState > 0 && !h.hFinished {
			switch txrx {
			case directionRX:
				h.rxPkts++
				h.rxBytes += cntSize
			case directionTX:
				h.txPkts++
				h.txBytes += cntSize
			}
		}
	}

	return nil
}

func (h *HTTPLogElem) finished(haveNxtReq bool) bool {
	if haveNxtReq && h.hState != 0 {
		h.hFinished = true
		// todo: do some thing here
	}

	return h.hFinished
}
