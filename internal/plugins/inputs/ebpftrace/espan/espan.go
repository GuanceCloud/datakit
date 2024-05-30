// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package espan define espan related struct
package espan

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"math"
	"strconv"
	"sync"

	"github.com/GuanceCloud/cliutils/point"
	expRand "golang.org/x/exp/rand"
	// expRand "math/rand".
)

type ESpans []*EBPFSpan

type EBPFSpan struct {
	Pre, Next *EBPFSpan
	Childs    []*EBPFSpan

	TraceID  ID128
	ParentID ID64

	Meta *SpanMeta

	Used bool
}

type RandID struct {
	// 使用 ulid 是期望其作为 id 时能尽可能的按时间索引
	rand *expRand.Rand
	sync.Mutex
}

func NewRandID() (*RandID, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	v := binary.LittleEndian.Uint64(b)
	r := expRand.New(expRand.NewSource(v))
	r.Seed(v)

	return &RandID{
		rand: r,
	}, nil
}

func (dkid *RandID) ID() (*ID128, bool) {
	dkid.Lock()
	defer dkid.Unlock()
	if dkid.rand != nil {
		return &ID128{
			Low:  expRand.Uint64(),
			High: expRand.Uint64(),
		}, true
	}
	return nil, false
}

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

// constants used for the Knuth hashing, same as dd-agent.
const knuthFactor = uint64(1111111111111111111)

func (id ID128) Sampled(rate float64) bool {
	if rate < 1 {
		return (id.Low+id.High)*knuthFactor < uint64(math.MaxUint64*rate)
	}

	return true
}

func PtSetMeta(pt *point.Point, sp *EBPFSpan) {
	if pt == nil || sp == nil || sp.Meta == nil {
		return
	}

	meta := sp.Meta
	// span id is always hex

	pt.MustAdd(SpanID, ID64(meta.SpanID).StringHex())

	// ebpf: traceid, parenid
	pt.MustAdd(EBPFTraceID, ID128{Low: meta.ETraceIDLow, High: meta.ETraceIDHigh}.StringHex())

	pt.MustAdd(EBPFParentID, ID64(meta.EParentID).StringHex())

	// set spanid, traceid and parentid
	if meta.Encode == Encode_EncDec { // only for datadog trace id
		// traceid, parenid
		pt.MustAdd(TraceID, sp.TraceID.StringDec())

		if sp.ParentID != ID64(meta.EParentID) { // parent not ebpf span
			pt.MustAdd(ParentID, sp.ParentID.StringDec())
		} else {
			pt.MustAdd(ParentID, sp.ParentID.StringHex())
		}
	} else {
		pt.MustAdd(TraceID, sp.TraceID.StringHex())
		pt.MustAdd(ParentID, sp.ParentID.StringHex())
	}

	// always hex

	pt.Del(AppTraceIDL)
	pt.Del(AppTraceIDH)
	pt.Del(AppParentIDL)
}

func SpanMetaData(pt *point.Point) (*SpanMeta, bool) {
	if pt == nil {
		return nil, false
	}

	spanid, ok := getID64(pt, SpanID)
	if !ok {
		return nil, false
	}

	var spanKind Kind
	espanType, ok := getPtValStr(pt, EBPFSpanType)
	if !ok {
		return nil, false
	}
	if espanType == SpanTypEntry {
		spanKind = Kind_Server
	} else {
		spanKind = Kind_Client
	}

	netraceid, ok := getID128(pt, ReqSeq, RespSeq)
	if !ok {
		return nil, false
	}

	thrTraceid, ok := getID64(pt, ThrTraceID)
	if !ok {
		return nil, false
	}

	directionStr, ok := getPtValStr(pt, DirectionStr)
	if !ok {
		return nil, false
	}

	var direction Direction
	if directionStr == INCO {
		direction = Direction_DIN
	} else if directionStr == OUTG {
		direction = Direction_DOUT
	}

	meta := &SpanMeta{
		SpanID: uint64(spanid),

		NetTraceIDLow:  netraceid.Low,
		NetTraceIDHigh: netraceid.High,
		ThreadTraceID:  uint64(thrTraceid),

		Direction: direction,
		Kind:      spanKind,
	}

	if appTraceid, ok := getID128(pt, AppTraceIDL, AppTraceIDH); ok {
		meta.AppTraceIDLow = appTraceid.Low
		meta.AppTraceIDHigh = appTraceid.High

		id64, _ := getID64(pt, AppParentIDL)
		meta.AppParentID = uint64(id64)

		if aSpanSampled, ok := getPtValInt(pt, AppSpanSampled); ok {
			switch {
			case aSpanSampled > 0:
				meta.AppSampled = AppSampled_SampleKept
			case aSpanSampled < 0:
				meta.AppSampled = AppSampled_SampleRejected
			}
		}
		if aEncode, ok := getPtValStr(pt, AppTraceEncode); ok {
			switch aEncode {
			case ENCHEX:
				meta.Encode = Encode_EncHex
			case ENCDEC:
				meta.Encode = Encode_EncDec
			}
		}
	}

	return meta, true
}

func getPtValStr(pt *point.Point, name string) (string, bool) {
	if pt == nil {
		return "", false
	}

	if val := pt.Get(name); val != nil {
		if val, ok := val.(string); ok {
			return val, ok
		}
	}

	return "", false
}

func getID128(pt *point.Point, low, high string) (ID128, bool) {
	if pt == nil {
		return ID128{}, false
	}

	id := ID128{}
	if val := pt.Get(low); val != nil {
		if val, ok := val.(int64); ok {
			id.Low = uint64(val)
		} else {
			return ID128{}, false
		}
	} else {
		return ID128{}, false
	}

	if val := pt.Get(high); val != nil {
		if val, ok := val.(int64); ok {
			id.High = uint64(val)
		} else {
			return ID128{}, false
		}
	} else {
		return ID128{}, false
	}

	return id, true
}

func getID64(pt *point.Point, name string) (ID64, bool) {
	if pt == nil {
		return 0, false
	}

	if val := pt.Get(name); val != nil {
		if val, ok := val.(int64); ok {
			return ID64(val), true
		}
	}
	return 0, false
}

func getPtValInt(pt *point.Point, name string) (int64, bool) {
	if pt == nil {
		return 0, false
	}

	if val := pt.Get(name); val != nil {
		if val, ok := val.(int64); ok {
			return val, ok
		}
	}

	return 0, false
}
