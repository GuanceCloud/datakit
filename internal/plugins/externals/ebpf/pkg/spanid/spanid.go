// Package spanid define spanid
package spanid

import (
	"encoding/binary"
	"encoding/hex"
	"strconv"
)

const (
	// 原始数据中必须存在.
	Direction = "direction"

	// NetTraceID     = "net_trace_id".

	AppTraceIDL    = "app_trace_id_l"
	AppTraceIDH    = "app_trace_id_h"
	AppParentIDL   = "app_parent_id_l"
	AppSpanSampled = "app_span_sampled"
	AppTraceEncode = "app_trace_encode"

	ThrTraceID = "thread_trace_id"
	ReqSeq     = "req_seq"
	RespSeq    = "resp_seq"

	EBPFSpanType = "ebpf_span_type"
)

type ID64 uint64

type ID128 struct {
	Low  uint64
	High uint64
}

func (id ID64) StringHex() string {
	if id == 0 {
		return "0"
	}

	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(id))
	return hex.EncodeToString(b)
}

func (id ID64) StringDec() string {
	return strconv.FormatUint(uint64(id), 10)
}

func (id ID64) Byte() []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(id))
	return b
}

func (id ID64) Zero() bool {
	return id == 0
}

func (id ID128) StringHex() string {
	if id.Low == 0 && id.High == 0 {
		return "0"
	}
	b := make([]byte, 16)
	binary.BigEndian.PutUint64(b[8:], id.Low)
	binary.BigEndian.PutUint64(b[:8], id.High)
	return hex.EncodeToString(b)
}

func (id ID128) StringDec() string {
	return strconv.FormatUint(id.Low, 10)
}

func (id ID128) Byte() []byte {
	b := make([]byte, 16)
	binary.LittleEndian.PutUint64(b, id.Low)
	binary.LittleEndian.PutUint64(b[8:], id.High)
	return b
}

func (id ID128) Zero() bool {
	return id.Low == 0 && id.High == 0
}
